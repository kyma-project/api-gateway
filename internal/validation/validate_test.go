package validation

import (
	"encoding/json"
	"fmt"

	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"testing"

	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestValidators(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Validators Suite")
}

const (
	sampleServiceName    = "some-service"
	allowlistedDomain    = "foo.bar"
	notAllowlistedDomain = "myDomain.xyz"
	testDefaultDomain    = allowlistedDomain
	sampleValidHost      = sampleServiceName + "." + allowlistedDomain
)

var (
	testDomainAllowlist = []string{"foo.bar", "bar.foo", "kyma.local"}
)

var _ = Describe("Validate function", func() {

	It("Should fail for empty rules", func() {

		//given
		testAllowList := []string{"foo.bar", "bar.foo", "kyma.local"}
		input := &gatewayv1beta1.APIRule{
			Spec: gatewayv1beta1.APIRuleSpec{
				Rules:   nil,
				Service: getService(sampleServiceName, uint32(8080)),
				Host:    getHost(sampleValidHost),
			},
		}

		//when
		problems := (&APIRule{
			DomainAllowList: testAllowList,
		}).Validate(input, networkingv1beta1.VirtualServiceList{})

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal(".spec.rules"))
		Expect(problems[0].Message).To(Equal("No rules defined"))
	})

	It("Should fail for blocklisted service", func() {
		//given
		sampleBlocklistedService := "kubernetes"
		validHost := sampleBlocklistedService + "." + allowlistedDomain
		testBlockList := map[string][]string{
			"default": {sampleBlocklistedService, "kube-dns"},
			"example": {"service"}}
		input := &gatewayv1beta1.APIRule{
			ObjectMeta: v1.ObjectMeta{
				Namespace: "default",
			},
			Spec: gatewayv1beta1.APIRuleSpec{
				Service: getService(sampleBlocklistedService, uint32(443)),
				Host:    getHost(validHost),
				Rules: []gatewayv1beta1.Rule{
					{
						Path: "/abc",
						AccessStrategies: []*gatewayv1beta1.Authenticator{
							toAuthenticator("jwt", simpleJWTConfig()),
							toAuthenticator("noop", emptyConfig()),
						},
					},
				},
			}}

		//when
		problems := (&APIRule{
			ServiceBlockList: testBlockList,
			DomainAllowList:  testDomainAllowlist,
		}).Validate(input, networkingv1beta1.VirtualServiceList{})

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal(".spec.service.name"))
		Expect(problems[0].Message).To(Equal("Service kubernetes in namespace default is blocklisted"))
	})

	It("Should fail for not allowlisted domain", func() {
		//given
		invalidHost := sampleServiceName + "." + notAllowlistedDomain
		testBlockList := map[string][]string{
			"default": {"kubernetes", "kube-dns"},
			"example": {"service"}}
		input := &gatewayv1beta1.APIRule{
			Spec: gatewayv1beta1.APIRuleSpec{
				Service: getService(sampleServiceName, uint32(8080)),
				Host:    getHost(invalidHost),
				Rules: []gatewayv1beta1.Rule{
					{
						Path: "/abc",
						AccessStrategies: []*gatewayv1beta1.Authenticator{
							toAuthenticator("jwt", simpleJWTConfig()),
							toAuthenticator("noop", emptyConfig()),
						},
					},
				},
			}}

		//when
		problems := (&APIRule{
			ServiceBlockList: testBlockList,
			DomainAllowList:  testDomainAllowlist,
		}).Validate(input, networkingv1beta1.VirtualServiceList{})

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal(".spec.host"))
		Expect(problems[0].Message).To(Equal("Host is not allowlisted"))
	})

	It("Should fail for blocklisted subdomain with default domainName (FQDN)", func() {
		//given
		blocklistedSubdomain := "api"
		blockedhost := blocklistedSubdomain + "." + testDefaultDomain
		testBlockList := map[string][]string{
			"default": {"kubernetes", "kube-dns"},
			"example": {"service"}}
		testHostBlockList := []string{blockedhost}
		input := &gatewayv1beta1.APIRule{
			Spec: gatewayv1beta1.APIRuleSpec{
				Service: getService(sampleServiceName, uint32(8080)),
				Host:    getHost(blockedhost),
				Rules: []gatewayv1beta1.Rule{
					{
						Path: "/abc",
						AccessStrategies: []*gatewayv1beta1.Authenticator{
							toAuthenticator("jwt", simpleJWTConfig()),
							toAuthenticator("noop", emptyConfig()),
						},
					},
				},
			}}

		//when
		problems := (&APIRule{
			ServiceBlockList:  testBlockList,
			DomainAllowList:   testDomainAllowlist,
			HostBlockList:     testHostBlockList,
			DefaultDomainName: testDefaultDomain,
		}).Validate(input, networkingv1beta1.VirtualServiceList{})

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal(".spec.host"))
		Expect(problems[0].Message).To(Equal(fmt.Sprintf("The subdomain %s is blocklisted for %s domain", blocklistedSubdomain, testDefaultDomain)))
	})

	It("Should NOT fail for blocklisted subdomain with custom domainName", func() {
		//given
		blocklistedSubdomain := "api"
		customDomainName := "bar.foo"
		blockedhost := blocklistedSubdomain + "." + testDefaultDomain
		customHost := blocklistedSubdomain + "." + customDomainName
		testBlockList := map[string][]string{
			"default": {"kubernetes", "kube-dns"},
			"example": {"service"}}
		testHostBlockList := []string{blockedhost}
		input := &gatewayv1beta1.APIRule{
			Spec: gatewayv1beta1.APIRuleSpec{
				Service: getService(sampleServiceName, uint32(8080)),
				Host:    getHost(customHost),
				Rules: []gatewayv1beta1.Rule{
					{
						Path: "/abc",
						AccessStrategies: []*gatewayv1beta1.Authenticator{
							toAuthenticator("jwt", simpleJWTConfig()),
							toAuthenticator("noop", emptyConfig()),
						},
					},
				},
			}}

		//when
		problems := (&APIRule{
			ServiceBlockList:  testBlockList,
			DomainAllowList:   testDomainAllowlist,
			HostBlockList:     testHostBlockList,
			DefaultDomainName: testDefaultDomain,
		}).Validate(input, networkingv1beta1.VirtualServiceList{})

		//then
		Expect(problems).To(HaveLen(0))
	})

	It("Should NOT fail for empty allowlisted domain", func() {
		//given
		validHost := sampleServiceName + "." + notAllowlistedDomain
		testBlockList := map[string][]string{
			"default": {"kubernetes", "kube-dns"},
			"example": {"service"}}
		input := &gatewayv1beta1.APIRule{
			Spec: gatewayv1beta1.APIRuleSpec{
				Service: getService(sampleServiceName, uint32(8080)),
				Host:    getHost(validHost),
				Rules: []gatewayv1beta1.Rule{
					{
						Path: "/abc",
						AccessStrategies: []*gatewayv1beta1.Authenticator{
							toAuthenticator("jwt", simpleJWTConfig()),
							toAuthenticator("noop", emptyConfig()),
						},
					},
				},
			}}

		//when
		problems := (&APIRule{
			ServiceBlockList: testBlockList,
			DomainAllowList:  []string{},
		}).Validate(input, networkingv1beta1.VirtualServiceList{})

		//then
		Expect(problems).To(HaveLen(0))
	})

	It("Should fail for not allowlisted domain containing allowlisted domain", func() {
		//given
		invalidHost := sampleServiceName + "." + allowlistedDomain + "." + notAllowlistedDomain
		testBlockList := map[string][]string{
			"default": {"kubernetes", "kube-dns"},
			"example": {"service"}}
		input := &gatewayv1beta1.APIRule{
			Spec: gatewayv1beta1.APIRuleSpec{
				Service: getService(sampleServiceName, uint32(8080)),
				Host:    getHost(invalidHost),
				Rules: []gatewayv1beta1.Rule{
					{
						Path: "/abc",
						AccessStrategies: []*gatewayv1beta1.Authenticator{
							toAuthenticator("jwt", simpleJWTConfig()),
							toAuthenticator("noop", emptyConfig()),
						},
					},
				},
			}}

		//when
		problems := (&APIRule{
			ServiceBlockList: testBlockList,
			DomainAllowList:  testDomainAllowlist,
		}).Validate(input, networkingv1beta1.VirtualServiceList{})

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal(".spec.host"))
		Expect(problems[0].Message).To(Equal("Host is not allowlisted"))
	})

	It("Should fail for no domain when default domain is not configured", func() {
		//given
		hostWithoutDomain := sampleServiceName
		testBlockList := map[string][]string{
			"default": {"kubernetes", "kube-dns"},
			"example": {"service"}}
		input := &gatewayv1beta1.APIRule{
			Spec: gatewayv1beta1.APIRuleSpec{
				Service: getService(sampleServiceName, uint32(8080)),
				Host:    getHost(hostWithoutDomain),
				Rules: []gatewayv1beta1.Rule{
					{
						Path: "/abc",
						AccessStrategies: []*gatewayv1beta1.Authenticator{
							toAuthenticator("jwt", simpleJWTConfig()),
							toAuthenticator("noop", emptyConfig()),
						},
					},
				},
			}}

		//when
		problems := (&APIRule{
			ServiceBlockList: testBlockList,
			DomainAllowList:  testDomainAllowlist,
		}).Validate(input, networkingv1beta1.VirtualServiceList{})

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal(".spec.host"))
		Expect(problems[0].Message).To(Equal("Host does not contain a domain name and no default domain name is configured"))
	})

	It("Should NOT fail for no domain when default domain is configured", func() {
		//given
		hostWithoutDomain := sampleServiceName
		testBlockList := map[string][]string{
			"default": {"kubernetes", "kube-dns"},
			"example": {"service"}}
		input := &gatewayv1beta1.APIRule{
			Spec: gatewayv1beta1.APIRuleSpec{
				Service: getService(sampleServiceName, uint32(8080)),
				Host:    getHost(hostWithoutDomain),
				Rules: []gatewayv1beta1.Rule{
					{
						Path: "/abc",
						AccessStrategies: []*gatewayv1beta1.Authenticator{
							toAuthenticator("jwt", simpleJWTConfig()),
							toAuthenticator("noop", emptyConfig()),
						},
					},
				},
			}}

		//when
		problems := (&APIRule{
			ServiceBlockList:  testBlockList,
			DomainAllowList:   testDomainAllowlist,
			DefaultDomainName: testDefaultDomain,
		}).Validate(input, networkingv1beta1.VirtualServiceList{})

		//then
		Expect(problems).To(HaveLen(0))
	})

	It("Should fail for serviceHost containing duplicated allowlisted domain", func() {
		//given
		invalidHost := sampleServiceName + "." + allowlistedDomain + "." + allowlistedDomain
		testBlockList := map[string][]string{
			"default": {"kubernetes", "kube-dns"},
			"example": {"service"}}
		input := &gatewayv1beta1.APIRule{
			Spec: gatewayv1beta1.APIRuleSpec{
				Service: getService(sampleServiceName, uint32(8080)),
				Host:    getHost(invalidHost),
				Rules: []gatewayv1beta1.Rule{
					{
						Path: "/abc",
						AccessStrategies: []*gatewayv1beta1.Authenticator{
							toAuthenticator("jwt", simpleJWTConfig()),
							toAuthenticator("noop", emptyConfig()),
						},
					},
				},
			}}

		//when
		problems := (&APIRule{
			ServiceBlockList: testBlockList,
			DomainAllowList:  testDomainAllowlist,
		}).Validate(input, networkingv1beta1.VirtualServiceList{})

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal(".spec.host"))
		Expect(problems[0].Message).To(Equal("Host is not allowlisted"))
	})

	It("Should fail for a host that is occupied by a VS exposed by another resource", func() {
		//given
		occupiedHost := "occupied-host" + allowlistedDomain
		existingVS := networkingv1beta1.VirtualService{}
		existingVS.OwnerReferences = []v1.OwnerReference{{UID: "12345"}}
		existingVS.Spec.Hosts = []string{occupiedHost}

		input := &gatewayv1beta1.APIRule{
			ObjectMeta: v1.ObjectMeta{
				UID: "67890",
			},
			Spec: gatewayv1beta1.APIRuleSpec{
				Service: getService(sampleServiceName, uint32(8080)),
				Host:    getHost(occupiedHost),
				Rules: []gatewayv1beta1.Rule{
					{
						Path: "/abc",
						AccessStrategies: []*gatewayv1beta1.Authenticator{
							toAuthenticator("jwt", simpleJWTConfig()),
							toAuthenticator("noop", emptyConfig()),
						},
					},
				},
			},
		}

		//when
		problems := (&APIRule{
			DomainAllowList: testDomainAllowlist,
		}).Validate(input, networkingv1beta1.VirtualServiceList{Items: []*networkingv1beta1.VirtualService{&existingVS}})

		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal(".spec.host"))
		Expect(problems[0].Message).To(Equal("This host is occupied by another Virtual Service"))
	})

	It("Should NOT fail for a host that is occupied by a VS exposed by this resource", func() {
		//given
		occupiedHost := "occupied-host" + allowlistedDomain
		existingVS := networkingv1beta1.VirtualService{}
		existingVS.OwnerReferences = []v1.OwnerReference{{UID: "12345"}}
		existingVS.Spec.Hosts = []string{occupiedHost}

		input := &gatewayv1beta1.APIRule{
			ObjectMeta: v1.ObjectMeta{
				UID: "12345",
			},
			Spec: gatewayv1beta1.APIRuleSpec{
				Service: getService(sampleServiceName, uint32(8080)),
				Host:    getHost(occupiedHost),
				Rules: []gatewayv1beta1.Rule{
					{
						Path: "/abc",
						AccessStrategies: []*gatewayv1beta1.Authenticator{
							toAuthenticator("jwt", simpleJWTConfig()),
							toAuthenticator("noop", emptyConfig()),
						},
					},
				},
			},
		}

		//when
		problems := (&APIRule{
			DomainAllowList: testDomainAllowlist,
		}).Validate(input, networkingv1beta1.VirtualServiceList{Items: []*networkingv1beta1.VirtualService{&existingVS}})

		Expect(problems).To(HaveLen(0))
	})

	It("Should return an error when no service is defined for rule with no service on spec level", func() {
		//given
		input := &gatewayv1beta1.APIRule{
			Spec: gatewayv1beta1.APIRuleSpec{
				Host: getHost(sampleValidHost),
				Rules: []gatewayv1beta1.Rule{
					{
						Path: "/abc",
						AccessStrategies: []*gatewayv1beta1.Authenticator{
							toAuthenticator("noop", emptyConfig()),
						},
					},
				},
			},
		}
		//when
		problems := (&APIRule{
			DomainAllowList: testDomainAllowlist,
		}).Validate(input, networkingv1beta1.VirtualServiceList{})

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal(".spec.rules[0].service"))
		Expect(problems[0].Message).To(Equal("No service defined with no main service on spec level"))
	})

	It("Should return an error when rule is defined with blocklisted service", func() {
		//given
		sampleBlocklistedService := "kubernetes"
		validHost := sampleBlocklistedService + "." + allowlistedDomain
		testBlockList := map[string][]string{
			"default": {sampleBlocklistedService, "kube-dns"},
			"example": {"service"}}

		input := &gatewayv1beta1.APIRule{
			ObjectMeta: v1.ObjectMeta{
				Namespace: "default",
			},
			Spec: gatewayv1beta1.APIRuleSpec{
				Host: getHost(validHost),
				Rules: []gatewayv1beta1.Rule{
					{
						Path: "/abc",
						AccessStrategies: []*gatewayv1beta1.Authenticator{
							toAuthenticator("noop", emptyConfig()),
						},
						Service: getService(sampleServiceName, uint32(8080)),
					},
					{
						Path: "/abcd",
						AccessStrategies: []*gatewayv1beta1.Authenticator{
							toAuthenticator("noop", emptyConfig()),
						},
						Service: getService(sampleBlocklistedService, uint32(8080)),
					},
				},
			},
		}
		//when
		problems := (&APIRule{
			ServiceBlockList: testBlockList,
			DomainAllowList:  testDomainAllowlist,
		}).Validate(input, networkingv1beta1.VirtualServiceList{})

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal(".spec.rules[1].service.name"))
		Expect(problems[0].Message).To(Equal(fmt.Sprintf("Service %s in namespace default is blocklisted", sampleBlocklistedService)))
	})

	It("Should detect several problems", func() {
		//given
		input := &gatewayv1beta1.APIRule{
			Spec: gatewayv1beta1.APIRuleSpec{
				Service: getService(sampleServiceName, uint32(8080)),
				Host:    getHost(sampleValidHost),
				Rules: []gatewayv1beta1.Rule{
					{
						Path: "/abc",
						AccessStrategies: []*gatewayv1beta1.Authenticator{
							toAuthenticator("noop", simpleJWTConfig()),
							toAuthenticator("jwt", emptyConfig()),
						},
					},
					{
						Path: "/abc",
						AccessStrategies: []*gatewayv1beta1.Authenticator{
							toAuthenticator("anonymous", simpleJWTConfig()),
						},
					},
					{
						Path: "/def",
						AccessStrategies: []*gatewayv1beta1.Authenticator{
							toAuthenticator("non-existing", nil),
						},
					},
					{
						Path:             "/ghi",
						AccessStrategies: []*gatewayv1beta1.Authenticator{},
					},
				},
			},
		}
		//when
		problems := (&APIRule{
			DomainAllowList: testDomainAllowlist,
		}).Validate(input, networkingv1beta1.VirtualServiceList{})

		//then
		Expect(problems).To(HaveLen(6))
		Expect(problems[0].AttributePath).To(Equal(".spec.rules"))
		Expect(problems[0].Message).To(Equal("Multiple rules defined for the same path"))

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
		occupiedHost := "occupied-host" + allowlistedDomain
		notOccupiedHost := "not-occupied-host" + allowlistedDomain
		existingVS := networkingv1beta1.VirtualService{}
		existingVS.OwnerReferences = []v1.OwnerReference{{UID: "12345"}}
		existingVS.Spec.Hosts = []string{occupiedHost}

		input := &gatewayv1beta1.APIRule{
			ObjectMeta: v1.ObjectMeta{
				UID: "67890",
			},
			Spec: gatewayv1beta1.APIRuleSpec{
				Service: getService(sampleServiceName, uint32(8080)),
				Host:    getHost(notOccupiedHost),
				Rules: []gatewayv1beta1.Rule{
					{
						Path: "/abc",
						AccessStrategies: []*gatewayv1beta1.Authenticator{
							toAuthenticator("jwt", simpleJWTConfig()),
							toAuthenticator("noop", emptyConfig()),
						},
					},
					{
						Path: "/bcd",
						AccessStrategies: []*gatewayv1beta1.Authenticator{
							toAuthenticator("anonymous", emptyConfig()),
						},
					},
					{
						Path: "/def",
						AccessStrategies: []*gatewayv1beta1.Authenticator{
							toAuthenticator("allow", nil),
						},
					},
				},
			},
		}
		//when
		problems := (&APIRule{
			DomainAllowList: testDomainAllowlist,
		}).Validate(input, networkingv1beta1.VirtualServiceList{Items: []*networkingv1beta1.VirtualService{&existingVS}})

		//then
		Expect(problems).To(HaveLen(0))
	})
})

var _ = Describe("Validator for", func() {

	Describe("NoConfig access strategy", func() {

		It("Should fail with non-empty config", func() {
			//given
			handler := &gatewayv1beta1.Handler{Name: "noop", Config: simpleJWTConfig("http://atgo.org")}

			//when
			problems := (&noConfigAccStrValidator{}).Validate("some.attribute", handler)

			//then
			Expect(problems).To(HaveLen(1))
			Expect(problems[0].AttributePath).To(Equal("some.attribute.config"))
			Expect(problems[0].Message).To(Equal("strategy: noop does not support configuration"))
		})

		It("Should succeed with empty config: {}", func() {
			//given
			handler := &gatewayv1beta1.Handler{Name: "noop", Config: emptyConfig()}

			//when
			problems := (&noConfigAccStrValidator{}).Validate("some.attribute", handler)

			//then
			Expect(problems).To(HaveLen(0))
		})

		It("Should succeed with null config", func() {
			//given
			handler := &gatewayv1beta1.Handler{Name: "noop", Config: nil}

			//when
			problems := (&noConfigAccStrValidator{}).Validate("some.attribute", handler)

			//then
			Expect(problems).To(HaveLen(0))
		})
	})

	Describe("JWT access strategy", func() {

		It("Should fail with empty config", func() {
			//given
			handler := &gatewayv1beta1.Handler{Name: "jwt", Config: emptyConfig()}

			//when
			problems := (&jwtAccStrValidator{}).Validate("some.attribute", handler)

			//then
			Expect(problems).To(HaveLen(1))
			Expect(problems[0].AttributePath).To(Equal("some.attribute.config"))
			Expect(problems[0].Message).To(Equal("supplied config cannot be empty"))
		})

		It("Should fail for config with invalid trustedIssuers and JWKSUrls", func() {
			//given
			handler := &gatewayv1beta1.Handler{Name: "jwt", Config: simpleJWTConfig("a t g o")}

			//when
			problems := (&jwtAccStrValidator{}).Validate("some.attribute", handler)

			//then
			Expect(problems).To(HaveLen(2))
			Expect(problems[0].AttributePath).To(Equal("some.attribute.config.trusted_issuers[0]"))
			Expect(problems[0].Message).To(ContainSubstring("value is empty or not a valid url"))
			Expect(problems[1].AttributePath).To(Equal("some.attribute.config.jwks_urls[0]"))
			Expect(problems[1].Message).To(ContainSubstring("value is empty or not a valid url"))
		})

		It("Should fail for config with plain HTTP JWKSUrls and trustedIssuers", func() {
			//given
			handler := &gatewayv1beta1.Handler{Name: "jwt", Config: testURLJWTConfig("http://issuer.test/.well-known/jwks.json", "http://issuer.test/")}

			//when
			problems := (&jwtAccStrValidator{}).Validate("some.attribute", handler)

			//then
			Expect(problems).To(HaveLen(2))
			Expect(problems[0].AttributePath).To(Equal("some.attribute.config.trusted_issuers[0]"))
			Expect(problems[0].Message).To(ContainSubstring("value is not a secured url"))
			Expect(problems[1].AttributePath).To(Equal("some.attribute.config.jwks_urls[0]"))
			Expect(problems[1].Message).To(ContainSubstring("value is not a secured url"))
		})

		It("Should succeed for config with file JWKSUrls and HTTPS trustedIssuers", func() {
			//given
			handler := &gatewayv1beta1.Handler{Name: "jwt", Config: testURLJWTConfig("file://.well-known/jwks.json", "https://issuer.test/")}

			//when
			problems := (&jwtAccStrValidator{}).Validate("some.attribute", handler)

			//then
			Expect(problems).To(HaveLen(0))
		})

		It("Should succeed for config with HTTPS JWKSUrls and trustedIssuers", func() {
			//given
			handler := &gatewayv1beta1.Handler{Name: "jwt", Config: testURLJWTConfig("https://issuer.test/.well-known/jwks.json", "https://issuer.test/")}

			//when
			problems := (&jwtAccStrValidator{}).Validate("some.attribute", handler)

			//then
			Expect(problems).To(HaveLen(0))
		})

		It("Should fail for invalid JSON", func() {
			//given
			handler := &gatewayv1beta1.Handler{Name: "jwt", Config: &runtime.RawExtension{Raw: []byte("/abc]")}}

			//when
			problems := (&jwtAccStrValidator{}).Validate("some.attribute", handler)

			//then
			Expect(problems).To(HaveLen(1))
			Expect(problems[0].AttributePath).To(Equal("some.attribute.config"))
			Expect(problems[0].Message).To(Equal("Can't read json: invalid character '/' looking for beginning of value"))
		})

		It("Should succeed with valid config", func() {
			//given
			handler := &gatewayv1beta1.Handler{Name: "jwt", Config: simpleJWTConfig()}

			//when
			problems := (&jwtAccStrValidator{}).Validate("some.attribute", handler)

			//then
			Expect(problems).To(HaveLen(0))
		})
	})
})

func emptyConfig() *runtime.RawExtension {
	return getRawConfig(
		&gatewayv1beta1.JWTAccStrConfig{})
}

func simpleJWTConfig(trustedIssuers ...string) *runtime.RawExtension {
	return getRawConfig(
		&gatewayv1beta1.JWTAccStrConfig{
			JWKSUrls:       trustedIssuers,
			TrustedIssuers: trustedIssuers,
			RequiredScopes: []string{"atgo"},
		})
}

func testURLJWTConfig(JWKSUrls string, trustedIssuers string) *runtime.RawExtension {
	return getRawConfig(
		&gatewayv1beta1.JWTAccStrConfig{
			JWKSUrls:       []string{JWKSUrls},
			TrustedIssuers: []string{trustedIssuers},
			RequiredScopes: []string{"atgo"},
		})
}

func getRawConfig(config *gatewayv1beta1.JWTAccStrConfig) *runtime.RawExtension {
	bytes, err := json.Marshal(config)
	Expect(err).To(BeNil())
	return &runtime.RawExtension{
		Raw: bytes,
	}
}

func toAuthenticator(name string, config *runtime.RawExtension) *gatewayv1beta1.Authenticator {
	return &gatewayv1beta1.Authenticator{
		Handler: &gatewayv1beta1.Handler{
			Name:   name,
			Config: config,
		},
	}
}

func getService(serviceName string, servicePort uint32) *gatewayv1beta1.Service {
	return &gatewayv1beta1.Service{
		Name: &serviceName,
		Port: &servicePort,
	}
}

func getHost(host string) *string {
	return &host
}
