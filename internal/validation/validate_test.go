package validation

import (
	"encoding/json"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"testing"

	gatewayv1alpha1 "github.com/kyma-incubator/api-gateway/api/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	rulev1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestValidators(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Validators Suite")
}

const (
	sampleServiceName    = "some-service"
	whitelistedDomain    = "foo.bar"
	notWhitelistedDomain = "myDomain.xyz"
	testDefaultDomain    = whitelistedDomain
	sampleValidHost      = sampleServiceName + "." + whitelistedDomain
)

var (
	testDomainWhitelist = []string{"foo.bar", "bar.foo", "kyma.local"}
)

var _ = Describe("Validate function", func() {

	It("Should fail for empty rules", func() {

		//given
		testWhiteList := []string{"foo.bar", "bar.foo", "kyma.local"}
		input := &gatewayv1alpha1.APIRule{
			Spec: gatewayv1alpha1.APIRuleSpec{
				Rules:   nil,
				Service: getService(sampleServiceName, uint32(8080), sampleValidHost),
			},
		}

		//when
		problems := (&APIRule{
			DomainWhiteList: testWhiteList,
		}).Validate(input, networkingv1beta1.VirtualServiceList{})

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal(".spec.rules"))
		Expect(problems[0].Message).To(Equal("No rules defined"))
	})

	It("Should fail for blacklisted service", func() {
		//given
		sampleBlacklistedService := "kubernetes"
		validHost := sampleBlacklistedService + "." + whitelistedDomain
		testBlackList := map[string][]string{
			"default": []string{sampleBlacklistedService, "kube-dns"},
			"example": []string{"service"}}
		input := &gatewayv1alpha1.APIRule{
			ObjectMeta: v1.ObjectMeta{
				Namespace: "default",
			},
			Spec: gatewayv1alpha1.APIRuleSpec{
				Service: getService(sampleBlacklistedService, uint32(443), validHost),
				Rules: []gatewayv1alpha1.Rule{
					{
						Path: "/abc",
						AccessStrategies: []*rulev1alpha1.Authenticator{
							toAuthenticator("jwt", simpleJWTConfig()),
							toAuthenticator("noop", emptyConfig()),
						},
					},
				},
			}}

		//when
		problems := (&APIRule{
			ServiceBlackList: testBlackList,
			DomainWhiteList:  testDomainWhitelist,
		}).Validate(input, networkingv1beta1.VirtualServiceList{})

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal(".spec.service.name"))
		Expect(problems[0].Message).To(Equal("Service kubernetes in namespace default is blacklisted"))
	})

	It("Should fail for not whitelisted domain", func() {
		//given
		invalidHost := sampleServiceName + "." + notWhitelistedDomain
		testBlackList := map[string][]string{
			"default": []string{"kubernetes", "kube-dns"},
			"example": []string{"service"}}
		input := &gatewayv1alpha1.APIRule{
			Spec: gatewayv1alpha1.APIRuleSpec{
				Service: getService(sampleServiceName, uint32(8080), invalidHost),
				Rules: []gatewayv1alpha1.Rule{
					{
						Path: "/abc",
						AccessStrategies: []*rulev1alpha1.Authenticator{
							toAuthenticator("jwt", simpleJWTConfig()),
							toAuthenticator("noop", emptyConfig()),
						},
					},
				},
			}}

		//when
		problems := (&APIRule{
			ServiceBlackList: testBlackList,
			DomainWhiteList:  testDomainWhitelist,
		}).Validate(input, networkingv1beta1.VirtualServiceList{})

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal(".spec.service.host"))
		Expect(problems[0].Message).To(Equal("Host is not whitelisted"))
	})

	It("Should fail for not whitelisted domain containing whitelisted domain", func() {
		//given
		invalidHost := sampleServiceName + "." + whitelistedDomain + "." + notWhitelistedDomain
		testBlackList := map[string][]string{
			"default": []string{"kubernetes", "kube-dns"},
			"example": []string{"service"}}
		input := &gatewayv1alpha1.APIRule{
			Spec: gatewayv1alpha1.APIRuleSpec{
				Service: getService(sampleServiceName, uint32(8080), invalidHost),
				Rules: []gatewayv1alpha1.Rule{
					{
						Path: "/abc",
						AccessStrategies: []*rulev1alpha1.Authenticator{
							toAuthenticator("jwt", simpleJWTConfig()),
							toAuthenticator("noop", emptyConfig()),
						},
					},
				},
			}}

		//when
		problems := (&APIRule{
			ServiceBlackList: testBlackList,
			DomainWhiteList:  testDomainWhitelist,
		}).Validate(input, networkingv1beta1.VirtualServiceList{})

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal(".spec.service.host"))
		Expect(problems[0].Message).To(Equal("Host is not whitelisted"))
	})

	It("Should fail for no domain when default domain is not configured", func() {
		//given
		hostWithoutDomain := sampleServiceName
		testBlackList := map[string][]string{
			"default": []string{"kubernetes", "kube-dns"},
			"example": []string{"service"}}
		input := &gatewayv1alpha1.APIRule{
			Spec: gatewayv1alpha1.APIRuleSpec{
				Service: getService(sampleServiceName, uint32(8080), hostWithoutDomain),
				Rules: []gatewayv1alpha1.Rule{
					{
						Path: "/abc",
						AccessStrategies: []*rulev1alpha1.Authenticator{
							toAuthenticator("jwt", simpleJWTConfig()),
							toAuthenticator("noop", emptyConfig()),
						},
					},
				},
			}}

		//when
		problems := (&APIRule{
			ServiceBlackList: testBlackList,
			DomainWhiteList:  testDomainWhitelist,
		}).Validate(input, networkingv1beta1.VirtualServiceList{})

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal(".spec.service.host"))
		Expect(problems[0].Message).To(Equal("Host does not contain a domain name and no default domain name is configured"))
	})

	It("Should NOT fail for no domain when default domain is configured", func() {
		//given
		hostWithoutDomain := sampleServiceName
		testBlackList := map[string][]string{
			"default": []string{"kubernetes", "kube-dns"},
			"example": []string{"service"}}
		input := &gatewayv1alpha1.APIRule{
			Spec: gatewayv1alpha1.APIRuleSpec{
				Service: getService(sampleServiceName, uint32(8080), hostWithoutDomain),
				Rules: []gatewayv1alpha1.Rule{
					{
						Path: "/abc",
						AccessStrategies: []*rulev1alpha1.Authenticator{
							toAuthenticator("jwt", simpleJWTConfig()),
							toAuthenticator("noop", emptyConfig()),
						},
					},
				},
			}}

		//when
		problems := (&APIRule{
			ServiceBlackList:  testBlackList,
			DomainWhiteList:   testDomainWhitelist,
			DefaultDomainName: testDefaultDomain,
		}).Validate(input, networkingv1beta1.VirtualServiceList{})

		//then
		Expect(problems).To(HaveLen(0))
	})

	It("Should fail for serviceHost containing duplicated whitelisted domain", func() {
		//given
		invalidHost := sampleServiceName + "." + whitelistedDomain + "." + whitelistedDomain
		testBlackList := map[string][]string{
			"default": []string{"kubernetes", "kube-dns"},
			"example": []string{"service"}}
		input := &gatewayv1alpha1.APIRule{
			Spec: gatewayv1alpha1.APIRuleSpec{
				Service: getService(sampleServiceName, uint32(8080), invalidHost),
				Rules: []gatewayv1alpha1.Rule{
					{
						Path: "/abc",
						AccessStrategies: []*rulev1alpha1.Authenticator{
							toAuthenticator("jwt", simpleJWTConfig()),
							toAuthenticator("noop", emptyConfig()),
						},
					},
				},
			}}

		//when
		problems := (&APIRule{
			ServiceBlackList: testBlackList,
			DomainWhiteList:  testDomainWhitelist,
		}).Validate(input, networkingv1beta1.VirtualServiceList{})

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal(".spec.service.host"))
		Expect(problems[0].Message).To(Equal("Host is not whitelisted"))
	})

	It("Should fail for a host that is occupied by a VS exposed by another resource", func() {
		//given
		occupiedHost := "occupied-host" + whitelistedDomain
		existingVS := networkingv1beta1.VirtualService{}
		existingVS.OwnerReferences = []v1.OwnerReference{{UID: "12345"}}
		existingVS.Spec.Hosts = []string{occupiedHost}

		input := &gatewayv1alpha1.APIRule{
			ObjectMeta: v1.ObjectMeta{
				UID: "67890",
			},
			Spec: gatewayv1alpha1.APIRuleSpec{
				Service: getService(sampleServiceName, uint32(8080), occupiedHost),
				Rules: []gatewayv1alpha1.Rule{
					{
						Path: "/abc",
						AccessStrategies: []*rulev1alpha1.Authenticator{
							toAuthenticator("jwt", simpleJWTConfig()),
							toAuthenticator("noop", emptyConfig()),
						},
					},
				},
			},
		}

		//when
		problems := (&APIRule{
			DomainWhiteList: testDomainWhitelist,
		}).Validate(input, networkingv1beta1.VirtualServiceList{Items: []networkingv1beta1.VirtualService{existingVS}})

		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal(".spec.service.host"))
		Expect(problems[0].Message).To(Equal("This host is occupied by another Virtual Service"))
	})

	It("Should NOT fail for a host that is occupied by a VS exposed by this resource", func() {
		//given
		occupiedHost := "occupied-host" + whitelistedDomain
		existingVS := networkingv1beta1.VirtualService{}
		existingVS.OwnerReferences = []v1.OwnerReference{{UID: "12345"}}
		existingVS.Spec.Hosts = []string{occupiedHost}

		input := &gatewayv1alpha1.APIRule{
			ObjectMeta: v1.ObjectMeta{
				UID: "12345",
			},
			Spec: gatewayv1alpha1.APIRuleSpec{
				Service: getService(sampleServiceName, uint32(8080), occupiedHost),
				Rules: []gatewayv1alpha1.Rule{
					{
						Path: "/abc",
						AccessStrategies: []*rulev1alpha1.Authenticator{
							toAuthenticator("jwt", simpleJWTConfig()),
							toAuthenticator("noop", emptyConfig()),
						},
					},
				},
			},
		}

		//when
		problems := (&APIRule{
			DomainWhiteList: testDomainWhitelist,
		}).Validate(input, networkingv1beta1.VirtualServiceList{Items: []networkingv1beta1.VirtualService{existingVS}})

		Expect(problems).To(HaveLen(0))
	})

	It("Should detect several problems", func() {
		//given
		input := &gatewayv1alpha1.APIRule{
			Spec: gatewayv1alpha1.APIRuleSpec{
				Service: getService(sampleServiceName, uint32(8080), sampleValidHost),
				Rules: []gatewayv1alpha1.Rule{
					{
						Path: "/abc",
						AccessStrategies: []*rulev1alpha1.Authenticator{
							toAuthenticator("noop", simpleJWTConfig()),
							toAuthenticator("jwt", emptyConfig()),
						},
					},
					{
						Path: "/abc",
						AccessStrategies: []*rulev1alpha1.Authenticator{
							toAuthenticator("anonymous", simpleJWTConfig()),
						},
					},
					{
						Path: "/def",
						AccessStrategies: []*rulev1alpha1.Authenticator{
							toAuthenticator("non-existing", nil),
						},
					},
					{
						Path:             "/ghi",
						AccessStrategies: []*rulev1alpha1.Authenticator{},
					},
				},
			},
		}
		//when
		problems := (&APIRule{
			DomainWhiteList: testDomainWhitelist,
		}).Validate(input, networkingv1beta1.VirtualServiceList{})

		//then
		Expect(problems).To(HaveLen(6))
		Expect(problems[0].AttributePath).To(Equal(".spec.rules"))
		Expect(problems[0].Message).To(Equal("multiple rules defined for the same path"))

		Expect(problems[1].AttributePath).To(Equal(".spec.rules[0].accessStrategies[0].config"))
		Expect(problems[1].Message).To(Equal("strategy: noop does not support configuration"))

		Expect(problems[2].AttributePath).To(Equal(".spec.rules[0].accessStrategies[1].config"))
		Expect(problems[2].Message).To(Equal("supplied config cannot be empty"))

		Expect(problems[3].AttributePath).To(Equal(".spec.rules[1].accessStrategies[0].config"))
		Expect(problems[3].Message).To(Equal("strategy: anonymous does not support configuration"))

		Expect(problems[4].AttributePath).To(Equal(".spec.rules[2].accessStrategies[0].handler"))
		Expect(problems[4].Message).To(Equal("Unsupported accessStrategy: non-existing"))

		Expect(problems[5].AttributePath).To(Equal(".spec.rules[3].accessStrategies"))
		Expect(problems[5].Message).To(Equal("No accessStrategies defined"))
	})

	It("Should succeed for valid input", func() {
		//given
		occupiedHost := "occupied-host" + whitelistedDomain
		notOccupiedHost := "not-occupied-host" + whitelistedDomain
		existingVS := networkingv1beta1.VirtualService{}
		existingVS.OwnerReferences = []v1.OwnerReference{{UID: "12345"}}
		existingVS.Spec.Hosts = []string{occupiedHost}

		input := &gatewayv1alpha1.APIRule{
			ObjectMeta: v1.ObjectMeta{
				UID: "67890",
			},
			Spec: gatewayv1alpha1.APIRuleSpec{
				Service: getService(sampleServiceName, uint32(8080), notOccupiedHost),
				Rules: []gatewayv1alpha1.Rule{
					{
						Path: "/abc",
						AccessStrategies: []*rulev1alpha1.Authenticator{
							toAuthenticator("jwt", simpleJWTConfig()),
							toAuthenticator("noop", emptyConfig()),
						},
					},
					{
						Path: "/bcd",
						AccessStrategies: []*rulev1alpha1.Authenticator{
							toAuthenticator("anonymous", emptyConfig()),
						},
					},
					{
						Path: "/def",
						AccessStrategies: []*rulev1alpha1.Authenticator{
							toAuthenticator("allow", nil),
						},
					},
				},
			},
		}
		//when
		problems := (&APIRule{
			DomainWhiteList: testDomainWhitelist,
		}).Validate(input, networkingv1beta1.VirtualServiceList{Items: []networkingv1beta1.VirtualService{existingVS}})

		//then
		Expect(problems).To(HaveLen(0))
	})
})

var _ = Describe("Validator for", func() {

	Describe("NoConfig access strategy", func() {

		It("Should fail with non-empty config", func() {
			//given
			handler := &rulev1alpha1.Handler{Name: "noop", Config: simpleJWTConfig("http://atgo.org")}

			//when
			problems := (&noConfigAccStrValidator{}).Validate("some.attribute", handler)

			//then
			Expect(problems).To(HaveLen(1))
			Expect(problems[0].AttributePath).To(Equal("some.attribute.config"))
			Expect(problems[0].Message).To(Equal("strategy: noop does not support configuration"))
		})

		It("Should succeed with empty config: {}", func() {
			//given
			handler := &rulev1alpha1.Handler{Name: "noop", Config: emptyConfig()}

			//when
			problems := (&noConfigAccStrValidator{}).Validate("some.attribute", handler)

			//then
			Expect(problems).To(HaveLen(0))
		})

		It("Should succeed with null config", func() {
			//given
			handler := &rulev1alpha1.Handler{Name: "noop", Config: nil}

			//when
			problems := (&noConfigAccStrValidator{}).Validate("some.attribute", handler)

			//then
			Expect(problems).To(HaveLen(0))
		})
	})

	Describe("JWT access strategy", func() {

		It("Should fail with empty config", func() {
			//given
			handler := &rulev1alpha1.Handler{Name: "jwt", Config: emptyConfig()}

			//when
			problems := (&jwtAccStrValidator{}).Validate("some.attribute", handler)

			//then
			Expect(problems).To(HaveLen(1))
			Expect(problems[0].AttributePath).To(Equal("some.attribute.config"))
			Expect(problems[0].Message).To(Equal("supplied config cannot be empty"))
		})

		It("Should fail for config with invalid trustedIssuers", func() {
			//given
			handler := &rulev1alpha1.Handler{Name: "jwt", Config: simpleJWTConfig("a t g o")}

			//when
			problems := (&jwtAccStrValidator{}).Validate("some.attribute", handler)

			//then
			Expect(problems).To(HaveLen(1))
			Expect(problems[0].AttributePath).To(Equal("some.attribute.config.trusted_issuers[0]"))
			Expect(problems[0].Message).To(Equal("value is empty or not a valid url"))
		})

		It("Should fail for invalid JSON", func() {
			//given
			handler := &rulev1alpha1.Handler{Name: "jwt", Config: &runtime.RawExtension{Raw: []byte("/abc]")}}

			//when
			problems := (&jwtAccStrValidator{}).Validate("some.attribute", handler)

			//then
			Expect(problems).To(HaveLen(1))
			Expect(problems[0].AttributePath).To(Equal("some.attribute.config"))
			Expect(problems[0].Message).To(Equal("Can't read json: invalid character '/' looking for beginning of value"))
		})

		It("Should succeed with valid config", func() {
			//given
			handler := &rulev1alpha1.Handler{Name: "jwt", Config: simpleJWTConfig()}

			//when
			problems := (&jwtAccStrValidator{}).Validate("some.attribute", handler)

			//then
			Expect(problems).To(HaveLen(0))
		})
	})
})

func emptyConfig() *runtime.RawExtension {
	return getRawConfig(
		&gatewayv1alpha1.JWTAccStrConfig{})
}

func simpleJWTConfig(trustedIssuers ...string) *runtime.RawExtension {
	return getRawConfig(
		&gatewayv1alpha1.JWTAccStrConfig{
			TrustedIssuers: trustedIssuers,
			RequiredScopes: []string{"atgo"},
		})
}

func getRawConfig(config *gatewayv1alpha1.JWTAccStrConfig) *runtime.RawExtension {
	bytes, err := json.Marshal(config)
	Expect(err).To(BeNil())
	return &runtime.RawExtension{
		Raw: bytes,
	}
}

func toAuthenticator(name string, config *runtime.RawExtension) *rulev1alpha1.Authenticator {
	return &rulev1alpha1.Authenticator{
		Handler: &rulev1alpha1.Handler{
			Name:   name,
			Config: config,
		},
	}
}

func getService(serviceName string, servicePort uint32, serviceHost string) *gatewayv1alpha1.Service {
	return &gatewayv1alpha1.Service{
		Name: &serviceName,
		Port: &servicePort,
		Host: &serviceHost,
	}
}
