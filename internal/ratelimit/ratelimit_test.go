package ratelimit

import (
	"os"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/util/json"
	"sigs.k8s.io/yaml"

	"github.com/kyma-project/api-gateway/internal/builders/envoyfilter"
)

var _ = Describe("RateLimit", func() {
	Context("NewLocalRateLimit", func() {
		It("should return a new RateLimit For descriptors d0, d1", func() {
			d0 := Descriptor{
				Entries: DescriptorEntries{
					{Key: "x-api-version", Val: "v1"},
					{Key: "path", Val: "/headers"},
				},
				Bucket: Bucket{
					MaxTokens:     2,
					TokensPerFill: 2,
					FillInterval:  30 * time.Second,
				},
			}
			d1 := Descriptor{
				Entries: DescriptorEntries{
					{Key: "path", Val: "/ip"},
				},
				Bucket: Bucket{
					MaxTokens:     20,
					TokensPerFill: 10,
					FillInterval:  30 * time.Second,
				},
			}

			rl := NewLocalRateLimit().
				For(d0).
				For(d1)
			Expect(rl.descriptors).Should(HaveLen(2))
			Expect(rl.actions).Should(HaveLen(2))
		})
		It("should have :path action when descriptor contains pseudo header 'path'", func() {
			d := Descriptor{
				Entries: DescriptorEntries{
					{Key: "path", Val: "/headers"},
				},
				Bucket: Bucket{
					MaxTokens:     2,
					TokensPerFill: 2,
					FillInterval:  30 * time.Second,
				},
			}
			rl := NewLocalRateLimit().For(d)
			Expect(rl.descriptors).Should(HaveLen(1))
			Expect(rl.actions).Should(HaveLen(1))
			Expect(rl.actions[0].RequestHeaders.Name).Should(Equal(":path"))
			Expect(rl.actions[0].RequestHeaders.DescriptorKey).Should(Equal("path"))
		})
		It("should return a new RateLimit For defaultBucket b", func() {
			b := Bucket{
				MaxTokens:     10,
				TokensPerFill: 5,
				FillInterval:  30 * time.Second,
			}
			rl := NewLocalRateLimit().WithDefaultBucket(b)
			Expect(rl.defaultBucket).Should(Equal(b))
		})
	})
	Context("AddToEnvoyFilter", func() {
		b := Bucket{
			MaxTokens:     10,
			TokensPerFill: 5,
			FillInterval:  30 * time.Second,
		}
		d0 := Descriptor{
			Entries: DescriptorEntries{
				{Key: "x-api-version", Val: "v1"},
				{Key: "path", Val: "/headers"},
			},
			Bucket: Bucket{
				MaxTokens:     2,
				TokensPerFill: 2,
				FillInterval:  30 * time.Second,
			},
		}
		d1 := Descriptor{
			Entries: DescriptorEntries{
				{Key: "path", Val: "/ip"},
			},
			Bucket: Bucket{
				MaxTokens:     20,
				TokensPerFill: 10,
				FillInterval:  30 * time.Second,
			},
		}
		It("adds 2 ConfigPatches to the EnvoyFilterBuilder", func() {
			ef := envoyfilter.NewEnvoyFilterBuilder().
				WithName("httpbin-local-rate-limit").
				WithNamespace("default").
				WithWorkloadSelector("app", "httpbin").Build()
			rl := NewLocalRateLimit().
				For(d1).
				For(d0).
				WithDefaultBucket(b)
			rl.SetConfigPatches(ef)
			Expect(ef.Spec.ConfigPatches).To(HaveLen(2))
		})
	})
	Context("RateLimitConfigPatch", func() {
		d0 := Descriptor{
			Entries: DescriptorEntries{
				{Key: "x-api-version", Val: "v1"},
				{Key: "path", Val: "/headers"},
			},
			Bucket: Bucket{
				MaxTokens:     2,
				TokensPerFill: 2,
				FillInterval:  30 * time.Second,
			},
		}
		defaultBucket := Bucket{
			MaxTokens:     20,
			TokensPerFill: 10,
			FillInterval:  30 * time.Second,
		}
		rl := NewLocalRateLimit().For(d0).WithDefaultBucket(defaultBucket)
		config := rl.RateLimitConfigPatch(patchContextSidecar)
		It("returns ConfigPatch with 2 route actions", func() {
			vals := config.Patch.Value.GetFields()
			rateLimits := vals["route"].GetStructValue().GetFields()["rate_limits"].GetListValue().GetValues()
			Expect(rateLimits).Should(HaveLen(1))
			actions := rateLimits[0].GetStructValue().GetFields()["actions"].GetListValue().GetValues()
			Expect(actions).Should(HaveLen(2))

			exp1 := Action{RequestHeaders: RequestHeader{Name: "x-api-version", DescriptorKey: "x-api-version"}}
			Expect(actions).Should(ContainElement(exp1.Value()))
			exp2 := Action{RequestHeaders: RequestHeader{Name: ":path", DescriptorKey: "path"}}
			Expect(actions).Should(ContainElement(exp2.Value()))
		})
		It("returns ConfigPatch With Default bucket", func() {
			vals := config.Patch.Value.GetFields()
			config := vals["typed_per_filter_config"].GetStructValue().GetFields()[LocalRateLimitFilterName]
			Expect(config).ShouldNot(BeNil())
			gotBucket := config.GetStructValue().GetFields()["value"].GetStructValue().GetFields()["token_bucket"]
			expBucket := Bucket{
				MaxTokens:     20,
				TokensPerFill: 10,
				FillInterval:  30 * time.Second,
			}
			Expect(gotBucket).Should(Equal(expBucket.Value()))
		})
		It("returns ConfigPatch with 1 descriptor", func() {
			vals := config.Patch.Value.GetFields()
			config := vals["typed_per_filter_config"].GetStructValue().GetFields()[LocalRateLimitFilterName]
			Expect(config).ShouldNot(BeNil())
			gotDesc := config.GetStructValue().GetFields()["value"].GetStructValue().GetFields()["descriptors"].GetListValue().GetValues()
			Expect(gotDesc).Should(HaveLen(1))

			expDesc := Descriptor{
				Entries: DescriptorEntries{
					{Key: "x-api-version", Val: "v1"},
					{Key: "path", Val: "/headers"},
				},
				Bucket: Bucket{
					MaxTokens:     2,
					TokensPerFill: 2,
					FillInterval:  30 * time.Second,
				},
			}

			Expect(gotDesc).Should(ContainElement(expDesc.Value()))
		})
	})
	Context("RateLimit to EnvoyFilter conversion", func() {
		d0 := Descriptor{
			Entries: DescriptorEntries{
				{Key: "x-api-version", Val: "v1"},
				{Key: "path", Val: "/headers"},
			},
			Bucket: Bucket{
				MaxTokens:     2,
				TokensPerFill: 2,
				FillInterval:  30 * time.Second,
			},
		}
		d1 := Descriptor{
			Entries: DescriptorEntries{
				{Key: "path", Val: "/ip"},
			},
			Bucket: Bucket{
				MaxTokens:     20,
				TokensPerFill: 10,
				FillInterval:  1 * time.Hour,
			},
		}
		bucket := Bucket{
			MaxTokens:     10,
			TokensPerFill: 5,
			FillInterval:  50 * time.Millisecond,
		}
		rl := NewLocalRateLimit().
			For(d1).
			For(d0).
			WithDefaultBucket(bucket).
			Enforce(true).
			EnableResponseHeaders(true)
		ef := envoyfilter.NewEnvoyFilterBuilder().
			WithName("httpbin-local-rate-limit").
			WithNamespace("default").
			WithWorkloadSelector("app", "httpbin").
			WithConfigPatch(&envoyfilter.ConfigPatch{}).
			WithConfigPatch(&envoyfilter.ConfigPatch{}).
			WithConfigPatch(&envoyfilter.ConfigPatch{}).
			Build()
		rl.SetConfigPatches(ef)
		It("builds EnvoyFilter with exactly 2 ConfigPatches", func() {
			Expect(ef.Spec.ConfigPatches).To(HaveLen(2))
		})
		It("builds EnvoyFilter with expected configuration", func() {
			f, err := os.ReadFile("testdata/envoy_patches.yaml")
			Expect(err).Should(Succeed())
			var fi []envoyfilter.ConfigPatch
			exp, err := yaml.YAMLToJSON(f)
			Expect(err).To(Succeed())
			Expect(json.Unmarshal(exp, &fi)).Should(Succeed())
			got, err := json.Marshal(ef.Spec.ConfigPatches)
			Expect(err).Should(Succeed())
			// This may fail if struct gets marshalled in not expected order.
			// However, I don't have much idea how to cover if marshaled structure is compatible with Envoy's
			// So just make sure the testdata is ordered.
			// Probably EnvTest would be better.
			Expect(got).Should(Equal(exp))
		})
	})
	Context("RateLimit to EnvoyFilter conversion for ingressgateway", func() {
		d0 := Descriptor{
			Entries: DescriptorEntries{
				{Key: "x-api-version", Val: "v1"},
				{Key: "path", Val: "/headers"},
			},
			Bucket: Bucket{
				MaxTokens:     2,
				TokensPerFill: 2,
				FillInterval:  30 * time.Second,
			},
		}
		d1 := Descriptor{
			Entries: DescriptorEntries{
				{Key: "path", Val: "/ip"},
			},
			Bucket: Bucket{
				MaxTokens:     20,
				TokensPerFill: 10,
				FillInterval:  1 * time.Hour,
			},
		}
		bucket := Bucket{
			MaxTokens:     10,
			TokensPerFill: 5,
			FillInterval:  50 * time.Millisecond,
		}
		rl := NewLocalRateLimit().
			For(d1).
			For(d0).
			WithDefaultBucket(bucket).
			Enforce(true).
			EnableResponseHeaders(true)
		ef := envoyfilter.NewEnvoyFilterBuilder().
			WithName("ingressgateway-local-rate-limit").
			WithNamespace("default").
			WithWorkloadSelector("app", "istio-ingressgateway").
			WithConfigPatch(&envoyfilter.ConfigPatch{}).
			WithConfigPatch(&envoyfilter.ConfigPatch{}).
			WithConfigPatch(&envoyfilter.ConfigPatch{}).
			Build()
		ef.Spec.WorkloadSelector.Labels = map[string]string{
			"app": "istio-ingressgateway",
		}
		rl.SetConfigPatches(ef)
		It("builds EnvoyFilter with exactly 2 ConfigPatches", func() {
			Expect(ef.Spec.ConfigPatches).To(HaveLen(2))
		})
		It("builds EnvoyFilter targeting ingressgateway with context set as gateway", func() {
			Expect(ef.Spec.ConfigPatches[0].Match.Context).To(Equal(patchContextGateway))
			Expect(ef.Spec.ConfigPatches[1].Match.Context).To(Equal(patchContextGateway))
		})
	})
})

func TestRateLimitSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "RateLimit Suite")
}
