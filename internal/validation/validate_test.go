package validation

import (
	"encoding/json"
	"fmt"

	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"testing"

	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	"github.com/kyma-incubator/api-gateway/internal/helpers"
	"github.com/kyma-incubator/api-gateway/internal/types/ory"
	. "github.com/onsi/ginkgo/v2"
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
	testDomainAllowlist  = []string{"foo.bar", "bar.foo", "kyma.local"}
	handlerValidatorMock = &dummyHandlerValidator{}
	asValidatorMock      = &dummyAccessStrategiesValidator{}
)

var _ = Describe("ValidateConfig function", func() {
	It("Should fail for missing config", func() {
		//when
		problems := (&APIRule{}).ValidateConfig(nil)

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].Message).To(Equal("Configuration is missing"))
	})

	It("Should fail for wrong config", func() {
		//given
		input := &helpers.Config{JWTHandler: "foo"}

		//when
		problems := (&APIRule{}).ValidateConfig(input)

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].Message).To(Equal("Unsupported JWT Handler: foo"))
	})
})

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
			HandlerValidator:          handlerValidatorMock,
			AccessStrategiesValidator: asValidatorMock,
			DomainAllowList:           testAllowList,
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
			HandlerValidator:          handlerValidatorMock,
			AccessStrategiesValidator: asValidatorMock,
			ServiceBlockList:          testBlockList,
			DomainAllowList:           testDomainAllowlist,
		}).Validate(input, networkingv1beta1.VirtualServiceList{})

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal(".spec.service.name"))
		Expect(problems[0].Message).To(Equal("Service kubernetes in namespace default is blocklisted"))
	})

	It("Should fail for blocklisted service for specific namespace", func() {
		//given
		sampleBlocklistedService := "service"
		sampleBlocklistedNamespace := "service-namespace"
		validHost := sampleBlocklistedService + "." + allowlistedDomain
		testBlockList := map[string][]string{
			"default":                  {"kube-dns"},
			sampleBlocklistedNamespace: {sampleBlocklistedService}}
		input := &gatewayv1beta1.APIRule{
			ObjectMeta: v1.ObjectMeta{
				Namespace: "default",
			},
			Spec: gatewayv1beta1.APIRuleSpec{
				Service: getService(sampleBlocklistedService, uint32(443), &sampleBlocklistedNamespace),
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
			HandlerValidator:          handlerValidatorMock,
			AccessStrategiesValidator: asValidatorMock,
			ServiceBlockList:          testBlockList,
			DomainAllowList:           testDomainAllowlist,
		}).Validate(input, networkingv1beta1.VirtualServiceList{})

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal(".spec.service.name"))
		Expect(problems[0].Message).To(Equal(fmt.Sprintf("Service %s in namespace %s is blocklisted", sampleBlocklistedService, sampleBlocklistedNamespace)))
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
			HandlerValidator:          handlerValidatorMock,
			AccessStrategiesValidator: asValidatorMock,
			ServiceBlockList:          testBlockList,
			DomainAllowList:           testDomainAllowlist,
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
			HandlerValidator:          handlerValidatorMock,
			AccessStrategiesValidator: asValidatorMock,
			ServiceBlockList:          testBlockList,
			DomainAllowList:           testDomainAllowlist,
			HostBlockList:             testHostBlockList,
			DefaultDomainName:         testDefaultDomain,
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
			HandlerValidator:          handlerValidatorMock,
			AccessStrategiesValidator: asValidatorMock,
			ServiceBlockList:          testBlockList,
			DomainAllowList:           testDomainAllowlist,
			HostBlockList:             testHostBlockList,
			DefaultDomainName:         testDefaultDomain,
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
			HandlerValidator:          handlerValidatorMock,
			AccessStrategiesValidator: asValidatorMock,
			ServiceBlockList:          testBlockList,
			DomainAllowList:           []string{},
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
			HandlerValidator:          handlerValidatorMock,
			AccessStrategiesValidator: asValidatorMock,
			ServiceBlockList:          testBlockList,
			DomainAllowList:           testDomainAllowlist,
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
			HandlerValidator:          handlerValidatorMock,
			AccessStrategiesValidator: asValidatorMock,
			ServiceBlockList:          testBlockList,
			DomainAllowList:           testDomainAllowlist,
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
			HandlerValidator:          handlerValidatorMock,
			AccessStrategiesValidator: asValidatorMock,
			ServiceBlockList:          testBlockList,
			DomainAllowList:           testDomainAllowlist,
			DefaultDomainName:         testDefaultDomain,
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
			HandlerValidator:          handlerValidatorMock,
			AccessStrategiesValidator: asValidatorMock,
			ServiceBlockList:          testBlockList,
			DomainAllowList:           testDomainAllowlist,
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
			HandlerValidator:          handlerValidatorMock,
			AccessStrategiesValidator: asValidatorMock,
			DomainAllowList:           testDomainAllowlist,
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
			HandlerValidator:          handlerValidatorMock,
			AccessStrategiesValidator: asValidatorMock,
			DomainAllowList:           testDomainAllowlist,
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
			AccessStrategiesValidator: asValidatorMock,
			DomainAllowList:           testDomainAllowlist,
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
			AccessStrategiesValidator: asValidatorMock,
			ServiceBlockList:          testBlockList,
			DomainAllowList:           testDomainAllowlist,
		}).Validate(input, networkingv1beta1.VirtualServiceList{})

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal(".spec.rules[1].service.name"))
		Expect(problems[0].Message).To(Equal(fmt.Sprintf("Service %s in namespace default is blocklisted", sampleBlocklistedService)))
	})

	It("Should return an error when rule is defined with blocklisted service in specific namespace", func() {
		//given
		sampleBlocklistedService := "service"
		sampleBlocklistedNamespace := "service-namespace"
		validHost := sampleBlocklistedService + "." + allowlistedDomain
		testBlockList := map[string][]string{
			"default":                  {"kube-dns"},
			sampleBlocklistedNamespace: {sampleBlocklistedService}}

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
						Service: getService(sampleBlocklistedService, uint32(8080), &sampleBlocklistedNamespace),
					},
				},
			},
		}
		//when
		problems := (&APIRule{
			AccessStrategiesValidator: asValidatorMock,
			ServiceBlockList:          testBlockList,
			DomainAllowList:           testDomainAllowlist,
		}).Validate(input, networkingv1beta1.VirtualServiceList{})

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal(".spec.rules[1].service.name"))
		Expect(problems[0].Message).To(Equal(fmt.Sprintf("Service %s in namespace %s is blocklisted", sampleBlocklistedService, sampleBlocklistedNamespace)))
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
			AccessStrategiesValidator: asValidatorMock,
			DomainAllowList:           testDomainAllowlist,
		}).Validate(input, networkingv1beta1.VirtualServiceList{})

		//then
		Expect(problems).To(HaveLen(5))
		Expect(problems[0].AttributePath).To(Equal(".spec.rules"))
		Expect(problems[0].Message).To(Equal("multiple rules defined for the same path and method"))

		Expect(problems[1].AttributePath).To(Equal(".spec.rules[0].accessStrategies[0].config"))
		Expect(problems[1].Message).To(Equal("strategy: noop does not support configuration"))

		Expect(problems[2].AttributePath).To(Equal(".spec.rules[1].accessStrategies[0].config"))
		Expect(problems[2].Message).To(Equal("strategy: anonymous does not support configuration"))

		Expect(problems[3].AttributePath).To(Equal(".spec.rules[2].accessStrategies[0].handler"))
		Expect(problems[3].Message).To(Equal("Unsupported accessStrategy: non-existing"))

		Expect(problems[4].AttributePath).To(Equal(".spec.rules[3].accessStrategies"))
		Expect(problems[4].Message).To(Equal("No accessStrategies defined"))

	})

	It("Should fail for the same path and method", func() {
		//given
		input := &gatewayv1beta1.APIRule{
			Spec: gatewayv1beta1.APIRuleSpec{
				Service: getService(sampleServiceName, uint32(8080)),
				Host:    getHost(sampleValidHost),
				Rules: []gatewayv1beta1.Rule{
					{
						Path: "/abc",
						AccessStrategies: []*gatewayv1beta1.Authenticator{
							toAuthenticator("noop", emptyConfig()),
						},
						Methods: []string{"GET"},
					},
					{
						Path: "/abc",
						AccessStrategies: []*gatewayv1beta1.Authenticator{
							toAuthenticator("anonymous", emptyConfig()),
						},
						Methods: []string{"GET", "POST"},
					},
				},
			},
		}
		//when
		problems := (&APIRule{
			AccessStrategiesValidator: asValidatorMock,
			DomainAllowList:           testDomainAllowlist,
		}).Validate(input, networkingv1beta1.VirtualServiceList{})

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal(".spec.rules"))
		Expect(problems[0].Message).To(Equal("multiple rules defined for the same path and method"))
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
						Methods: []string{"POST"},
					},
					{
						Path: "/abc",
						AccessStrategies: []*gatewayv1beta1.Authenticator{
							toAuthenticator("jwt", simpleJWTConfig()),
							toAuthenticator("noop", emptyConfig()),
						},
						Methods: []string{"GET"},
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
			HandlerValidator:          handlerValidatorMock,
			AccessStrategiesValidator: asValidatorMock,
			DomainAllowList:           testDomainAllowlist,
		}).Validate(input, networkingv1beta1.VirtualServiceList{Items: []*networkingv1beta1.VirtualService{&existingVS}})

		//then
		Expect(problems).To(HaveLen(0))
	})

	It("Should succeed for the same path but different methods", func() {
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
						Methods: []string{"POST"},
					},
					{
						Path: "/abc",
						AccessStrategies: []*gatewayv1beta1.Authenticator{
							toAuthenticator("jwt", simpleJWTConfig()),
							toAuthenticator("noop", emptyConfig()),
						},
						Methods: []string{"GET"},
					},
				},
			},
		}
		//when
		problems := (&APIRule{
			HandlerValidator:          handlerValidatorMock,
			AccessStrategiesValidator: asValidatorMock,
			DomainAllowList:           testDomainAllowlist,
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

})

func emptyConfig() *runtime.RawExtension {
	return getRawConfig(
		&ory.JWTAccStrConfig{})
}

func simpleJWTConfig(trustedIssuers ...string) *runtime.RawExtension {
	return getRawConfig(
		&ory.JWTAccStrConfig{
			JWKSUrls:       trustedIssuers,
			TrustedIssuers: trustedIssuers,
			RequiredScopes: []string{"atgo"},
		})
}

func getRawConfig(config *ory.JWTAccStrConfig) *runtime.RawExtension {
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

func getService(serviceName string, servicePort uint32, namespace ...*string) *gatewayv1beta1.Service {
	var serviceNamespace *string
	if len(namespace) > 0 {
		serviceNamespace = namespace[0]
	}
	return &gatewayv1beta1.Service{
		Name:      &serviceName,
		Namespace: serviceNamespace,
		Port:      &servicePort,
	}
}

func getHost(host string) *string {
	return &host
}
