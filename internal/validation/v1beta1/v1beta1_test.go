package v1beta1

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"github.com/kyma-project/api-gateway/internal/types/ory"
	rulev1alpha1 "github.com/kyma-project/api-gateway/internal/types/ory/oathkeeper-maester/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

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

var _ = Describe("Validate function", func() {
	It("Should fail for empty rules", func() {
		//given
		testAllowList := []string{"foo.bar", "bar.foo", "kyma.local"}
		apiRule := &v1beta1.APIRule{
			Spec: v1beta1.APIRuleSpec{
				Rules:   nil,
				Service: getApiRuleService(sampleServiceName, uint32(8080)),
				Host:    getHost(sampleValidHost),
			},
		}

		service := getService(sampleServiceName)
		fakeClient := buildFakeClient(service)

		//when
		problems := (&APIRuleValidator{
			ApiRule:                   apiRule,
			HandlerValidator:          handlerValidatorMock,
			AccessStrategiesValidator: asValidatorMock,
			DomainAllowList:           testAllowList,
		}).Validate(context.Background(), fakeClient, networkingv1beta1.VirtualServiceList{}, networkingv1beta1.GatewayList{})

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
		apiRule := &v1beta1.APIRule{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
			},
			Spec: v1beta1.APIRuleSpec{
				Service: getApiRuleService(sampleBlocklistedService, uint32(443)),
				Host:    getHost(validHost),
				Rules: []v1beta1.Rule{
					{
						Path: "/abc",
						AccessStrategies: []*v1beta1.Authenticator{
							toAuthenticator("jwt", simpleJWTConfig()),
						},
					},
				},
			}}

		service := getService(sampleBlocklistedService)
		fakeClient := buildFakeClient(service)

		//when
		problems := (&APIRuleValidator{
			ApiRule:                   apiRule,
			HandlerValidator:          handlerValidatorMock,
			AccessStrategiesValidator: asValidatorMock,
			ServiceBlockList:          testBlockList,
			DomainAllowList:           testDomainAllowlist,
		}).Validate(context.Background(), fakeClient, networkingv1beta1.VirtualServiceList{}, networkingv1beta1.GatewayList{})

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
		apiRule := &v1beta1.APIRule{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
			},
			Spec: v1beta1.APIRuleSpec{
				Service: getApiRuleService(sampleBlocklistedService, uint32(443), &sampleBlocklistedNamespace),
				Host:    getHost(validHost),
				Rules: []v1beta1.Rule{
					{
						Path: "/abc",
						AccessStrategies: []*v1beta1.Authenticator{
							toAuthenticator("jwt", simpleJWTConfig()),
						},
					},
				},
			}}

		service := getService(sampleBlocklistedService, sampleBlocklistedNamespace)
		fakeClient := buildFakeClient(service)

		//when
		problems := (&APIRuleValidator{
			ApiRule:                   apiRule,
			HandlerValidator:          handlerValidatorMock,
			AccessStrategiesValidator: asValidatorMock,
			ServiceBlockList:          testBlockList,
			DomainAllowList:           testDomainAllowlist,
		}).Validate(context.Background(), fakeClient, networkingv1beta1.VirtualServiceList{}, networkingv1beta1.GatewayList{})

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
		apiRule := &v1beta1.APIRule{
			Spec: v1beta1.APIRuleSpec{
				Service: getApiRuleService(sampleServiceName, uint32(8080)),
				Host:    getHost(invalidHost),
				Rules: []v1beta1.Rule{
					{
						Path: "/abc",
						AccessStrategies: []*v1beta1.Authenticator{
							toAuthenticator("jwt", simpleJWTConfig()),
						},
					},
				},
			}}

		service := getService(sampleServiceName)
		fakeClient := buildFakeClient(service)

		//when
		problems := (&APIRuleValidator{
			ApiRule:                   apiRule,
			HandlerValidator:          handlerValidatorMock,
			AccessStrategiesValidator: asValidatorMock,
			ServiceBlockList:          testBlockList,
			DomainAllowList:           testDomainAllowlist,
		}).Validate(context.Background(), fakeClient, networkingv1beta1.VirtualServiceList{}, networkingv1beta1.GatewayList{})

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
		apiRule := &v1beta1.APIRule{
			Spec: v1beta1.APIRuleSpec{
				Service: getApiRuleService(sampleServiceName, uint32(8080)),
				Host:    getHost(blockedhost),
				Rules: []v1beta1.Rule{
					{
						Path: "/abc",
						AccessStrategies: []*v1beta1.Authenticator{
							toAuthenticator("jwt", simpleJWTConfig()),
						},
					},
				},
			}}

		service := getService(sampleServiceName)
		fakeClient := buildFakeClient(service)

		//when
		problems := (&APIRuleValidator{
			ApiRule:                   apiRule,
			HandlerValidator:          handlerValidatorMock,
			AccessStrategiesValidator: asValidatorMock,
			ServiceBlockList:          testBlockList,
			DomainAllowList:           testDomainAllowlist,
			HostBlockList:             testHostBlockList,
			DefaultDomainName:         testDefaultDomain,
		}).Validate(context.Background(), fakeClient, networkingv1beta1.VirtualServiceList{}, networkingv1beta1.GatewayList{})

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
		apiRule := &v1beta1.APIRule{
			Spec: v1beta1.APIRuleSpec{
				Service: getApiRuleService(sampleServiceName, uint32(8080)),
				Host:    getHost(customHost),
				Rules: []v1beta1.Rule{
					{
						Path: "/abc",
						AccessStrategies: []*v1beta1.Authenticator{
							toAuthenticator("jwt", simpleJWTConfig()),
						},
					},
				},
			}}

		service := getService(sampleServiceName)
		fakeClient := buildFakeClient(service)

		//when
		problems := (&APIRuleValidator{
			ApiRule:                   apiRule,
			HandlerValidator:          handlerValidatorMock,
			AccessStrategiesValidator: asValidatorMock,
			ServiceBlockList:          testBlockList,
			DomainAllowList:           testDomainAllowlist,
			HostBlockList:             testHostBlockList,
			DefaultDomainName:         testDefaultDomain,
		}).Validate(context.Background(), fakeClient, networkingv1beta1.VirtualServiceList{}, networkingv1beta1.GatewayList{})

		//then
		Expect(problems).To(HaveLen(0))
	})

	It("Should NOT fail for empty allowlisted domain", func() {
		//given
		validHost := sampleServiceName + "." + notAllowlistedDomain
		testBlockList := map[string][]string{
			"default": {"kubernetes", "kube-dns"},
			"example": {"service"}}
		apiRule := &v1beta1.APIRule{
			Spec: v1beta1.APIRuleSpec{
				Service: getApiRuleService(sampleServiceName, uint32(8080)),
				Host:    getHost(validHost),
				Rules: []v1beta1.Rule{
					{
						Path: "/abc",
						AccessStrategies: []*v1beta1.Authenticator{
							toAuthenticator("jwt", simpleJWTConfig()),
						},
					},
				},
			}}

		service := getService(sampleServiceName)
		fakeClient := buildFakeClient(service)

		//when
		problems := (&APIRuleValidator{
			ApiRule:                   apiRule,
			HandlerValidator:          handlerValidatorMock,
			AccessStrategiesValidator: asValidatorMock,
			ServiceBlockList:          testBlockList,
			DomainAllowList:           []string{},
		}).Validate(context.Background(), fakeClient, networkingv1beta1.VirtualServiceList{}, networkingv1beta1.GatewayList{})

		//then
		Expect(problems).To(HaveLen(0))
	})

	It("Should fail for not allowlisted domain containing allowlisted domain", func() {
		//given
		invalidHost := sampleServiceName + "." + allowlistedDomain + "." + notAllowlistedDomain
		testBlockList := map[string][]string{
			"default": {"kubernetes", "kube-dns"},
			"example": {"service"}}
		apiRule := &v1beta1.APIRule{
			Spec: v1beta1.APIRuleSpec{
				Service: getApiRuleService(sampleServiceName, uint32(8080)),
				Host:    getHost(invalidHost),
				Rules: []v1beta1.Rule{
					{
						Path: "/abc",
						AccessStrategies: []*v1beta1.Authenticator{
							toAuthenticator("jwt", simpleJWTConfig()),
						},
					},
				},
			}}

		service := getService(sampleServiceName)
		fakeClient := buildFakeClient(service)

		//when
		problems := (&APIRuleValidator{
			ApiRule:                   apiRule,
			HandlerValidator:          handlerValidatorMock,
			AccessStrategiesValidator: asValidatorMock,
			ServiceBlockList:          testBlockList,
			DomainAllowList:           testDomainAllowlist,
		}).Validate(context.Background(), fakeClient, networkingv1beta1.VirtualServiceList{}, networkingv1beta1.GatewayList{})

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
		apiRule := &v1beta1.APIRule{
			Spec: v1beta1.APIRuleSpec{
				Service: getApiRuleService(sampleServiceName, uint32(8080)),
				Host:    getHost(hostWithoutDomain),
				Rules: []v1beta1.Rule{
					{
						Path: "/abc",
						AccessStrategies: []*v1beta1.Authenticator{
							toAuthenticator("jwt", simpleJWTConfig()),
						},
					},
				},
			}}

		service := getService(sampleServiceName)
		fakeClient := buildFakeClient(service)

		//when
		problems := (&APIRuleValidator{
			ApiRule:                   apiRule,
			HandlerValidator:          handlerValidatorMock,
			AccessStrategiesValidator: asValidatorMock,
			ServiceBlockList:          testBlockList,
			DomainAllowList:           testDomainAllowlist,
		}).Validate(context.Background(), fakeClient, networkingv1beta1.VirtualServiceList{}, networkingv1beta1.GatewayList{})

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
		apiRule := &v1beta1.APIRule{
			Spec: v1beta1.APIRuleSpec{
				Service: getApiRuleService(sampleServiceName, uint32(8080)),
				Host:    getHost(hostWithoutDomain),
				Rules: []v1beta1.Rule{
					{
						Path: "/abc",
						AccessStrategies: []*v1beta1.Authenticator{
							toAuthenticator("jwt", simpleJWTConfig()),
						},
					},
				},
			}}

		service := getService(sampleServiceName)
		fakeClient := buildFakeClient(service)

		//when
		problems := (&APIRuleValidator{
			ApiRule:                   apiRule,
			HandlerValidator:          handlerValidatorMock,
			AccessStrategiesValidator: asValidatorMock,
			ServiceBlockList:          testBlockList,
			DomainAllowList:           testDomainAllowlist,
			DefaultDomainName:         testDefaultDomain,
		}).Validate(context.Background(), fakeClient, networkingv1beta1.VirtualServiceList{}, networkingv1beta1.GatewayList{})

		//then
		Expect(problems).To(HaveLen(0))
	})

	It("Should fail for serviceHost containing duplicated allowlisted domain", func() {
		//given
		invalidHost := sampleServiceName + "." + allowlistedDomain + "." + allowlistedDomain
		testBlockList := map[string][]string{
			"default": {"kubernetes", "kube-dns"},
			"example": {"service"}}
		apiRule := &v1beta1.APIRule{
			Spec: v1beta1.APIRuleSpec{
				Service: getApiRuleService(sampleServiceName, uint32(8080)),
				Host:    getHost(invalidHost),
				Rules: []v1beta1.Rule{
					{
						Path: "/abc",
						AccessStrategies: []*v1beta1.Authenticator{
							toAuthenticator("jwt", simpleJWTConfig()),
						},
					},
				},
			}}

		service := getService(sampleServiceName)
		fakeClient := buildFakeClient(service)

		//when
		problems := (&APIRuleValidator{
			ApiRule:                   apiRule,
			HandlerValidator:          handlerValidatorMock,
			AccessStrategiesValidator: asValidatorMock,
			ServiceBlockList:          testBlockList,
			DomainAllowList:           testDomainAllowlist,
		}).Validate(context.Background(), fakeClient, networkingv1beta1.VirtualServiceList{}, networkingv1beta1.GatewayList{})

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal(".spec.host"))
		Expect(problems[0].Message).To(Equal("Host is not allowlisted"))
	})

	It("Should fail for a host that is occupied by a VS exposed by another resource", func() {
		//given
		occupiedHost := "occupied-host" + allowlistedDomain
		existingVS := networkingv1beta1.VirtualService{}
		existingVS.Spec.Hosts = []string{occupiedHost}

		apiRule := &v1beta1.APIRule{
			ObjectMeta: metav1.ObjectMeta{
				UID: "67890",
			},
			Spec: v1beta1.APIRuleSpec{
				Service: getApiRuleService(sampleServiceName, uint32(8080)),
				Host:    getHost(occupiedHost),
				Rules: []v1beta1.Rule{
					{
						Path: "/abc",
						AccessStrategies: []*v1beta1.Authenticator{
							toAuthenticator("jwt", simpleJWTConfig()),
						},
					},
				},
			},
		}

		service := getService(sampleServiceName)
		fakeClient := buildFakeClient(service)

		//when
		problems := (&APIRuleValidator{
			ApiRule:                   apiRule,
			HandlerValidator:          handlerValidatorMock,
			AccessStrategiesValidator: asValidatorMock,
			DomainAllowList:           testDomainAllowlist,
		}).Validate(context.Background(), fakeClient, networkingv1beta1.VirtualServiceList{Items: []*networkingv1beta1.VirtualService{&existingVS}}, networkingv1beta1.GatewayList{})

		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal(".spec.host"))
		Expect(problems[0].Message).To(Equal("This host is occupied by another Virtual Service"))
	})

	It("Should NOT fail for a host that is occupied by a VS exposed by this resource", func() {
		//given
		occupiedHost := "occupied-host" + allowlistedDomain

		apiRule := &v1beta1.APIRule{
			ObjectMeta: metav1.ObjectMeta{
				UID: "12345",
			},
			Spec: v1beta1.APIRuleSpec{
				Service: getApiRuleService(sampleServiceName, uint32(8080)),
				Host:    getHost(occupiedHost),
				Rules: []v1beta1.Rule{
					{
						Path: "/abc",
						AccessStrategies: []*v1beta1.Authenticator{
							toAuthenticator("jwt", simpleJWTConfig()),
						},
					},
				},
			},
		}

		service := getService(sampleServiceName)
		fakeClient := buildFakeClient(service)

		existingVS := networkingv1beta1.VirtualService{}
		existingVS.Labels = getOwnerLabels(apiRule)
		existingVS.Spec.Hosts = []string{occupiedHost}

		//when
		problems := (&APIRuleValidator{
			ApiRule:                   apiRule,
			HandlerValidator:          handlerValidatorMock,
			AccessStrategiesValidator: asValidatorMock,
			DomainAllowList:           testDomainAllowlist,
		}).Validate(context.Background(), fakeClient, networkingv1beta1.VirtualServiceList{Items: []*networkingv1beta1.VirtualService{&existingVS}}, networkingv1beta1.GatewayList{})

		Expect(problems).To(HaveLen(0))
	})

	It("Should return an error when no service is defined for rule with no service on spec level", func() {
		//given
		apiRule := &v1beta1.APIRule{
			Spec: v1beta1.APIRuleSpec{
				Host: getHost(sampleValidHost),
				Rules: []v1beta1.Rule{
					{
						Path: "/abc",
						AccessStrategies: []*v1beta1.Authenticator{
							toAuthenticator("noop", emptyConfig()),
						},
					},
				},
			},
		}

		service := getService(sampleServiceName, "default")
		fakeClient := buildFakeClient(service)

		//when
		problems := (&APIRuleValidator{
			ApiRule:                   apiRule,
			AccessStrategiesValidator: asValidatorMock,
			DomainAllowList:           testDomainAllowlist,
		}).Validate(context.Background(), fakeClient, networkingv1beta1.VirtualServiceList{}, networkingv1beta1.GatewayList{})

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

		apiRule := &v1beta1.APIRule{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
			},
			Spec: v1beta1.APIRuleSpec{
				Host: getHost(validHost),
				Rules: []v1beta1.Rule{
					{
						Path: "/abc",
						AccessStrategies: []*v1beta1.Authenticator{
							toAuthenticator("noop", emptyConfig()),
						},
						Service: getApiRuleService(sampleServiceName, uint32(8080)),
					},
					{
						Path: "/abcd",
						AccessStrategies: []*v1beta1.Authenticator{
							toAuthenticator("noop", emptyConfig()),
						},
						Service: getApiRuleService(sampleBlocklistedService, uint32(8080)),
					},
				},
			},
		}

		service1 := getService(sampleServiceName)
		service2 := getService(sampleBlocklistedService)
		fakeClient := buildFakeClient(service1, service2)

		//when
		problems := (&APIRuleValidator{
			ApiRule:                   apiRule,
			AccessStrategiesValidator: asValidatorMock,
			ServiceBlockList:          testBlockList,
			DomainAllowList:           testDomainAllowlist,
		}).Validate(context.Background(), fakeClient, networkingv1beta1.VirtualServiceList{}, networkingv1beta1.GatewayList{})

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

		apiRule := &v1beta1.APIRule{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
			},
			Spec: v1beta1.APIRuleSpec{
				Host: getHost(validHost),
				Rules: []v1beta1.Rule{
					{
						Path: "/abc",
						AccessStrategies: []*v1beta1.Authenticator{
							toAuthenticator("noop", emptyConfig()),
						},
						Service: getApiRuleService(sampleServiceName, uint32(8080)),
					},
					{
						Path: "/abcd",
						AccessStrategies: []*v1beta1.Authenticator{
							toAuthenticator("noop", emptyConfig()),
						},
						Service: getApiRuleService(sampleBlocklistedService, uint32(8080), &sampleBlocklistedNamespace),
					},
				},
			},
		}

		service1 := getService(sampleServiceName)
		service2 := getService(sampleBlocklistedService, sampleBlocklistedNamespace)
		fakeClient := buildFakeClient(service1, service2)

		//when
		problems := (&APIRuleValidator{
			ApiRule:                   apiRule,
			AccessStrategiesValidator: asValidatorMock,
			ServiceBlockList:          testBlockList,
			DomainAllowList:           testDomainAllowlist,
		}).Validate(context.Background(), fakeClient, networkingv1beta1.VirtualServiceList{}, networkingv1beta1.GatewayList{})

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal(".spec.rules[1].service.name"))
		Expect(problems[0].Message).To(Equal(fmt.Sprintf("Service %s in namespace %s is blocklisted", sampleBlocklistedService, sampleBlocklistedNamespace)))
	})

	It("Should detect several problems", func() {
		//given
		apiRule := &v1beta1.APIRule{
			Spec: v1beta1.APIRuleSpec{
				Service: getApiRuleService(sampleServiceName, uint32(8080)),
				Host:    getHost(sampleValidHost),
				Rules: []v1beta1.Rule{
					{
						Path: "/abc",
						AccessStrategies: []*v1beta1.Authenticator{
							toAuthenticator("noop", simpleJWTConfig()),
						},
					},
					{
						Path: "/abc",
						AccessStrategies: []*v1beta1.Authenticator{
							toAuthenticator("anonymous", simpleJWTConfig()),
						},
					},
					{
						Path: "/def",
						AccessStrategies: []*v1beta1.Authenticator{
							toAuthenticator("non-existing", nil),
						},
					},
					{
						Path:             "/ghi",
						AccessStrategies: []*v1beta1.Authenticator{},
					},
				},
			},
		}

		service := getService(sampleServiceName)
		fakeClient := buildFakeClient(service)

		//when
		problems := (&APIRuleValidator{
			ApiRule:                   apiRule,
			AccessStrategiesValidator: asValidatorMock,
			DomainAllowList:           testDomainAllowlist,
		}).Validate(context.Background(), fakeClient, networkingv1beta1.VirtualServiceList{}, networkingv1beta1.GatewayList{})

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
		apiRule := &v1beta1.APIRule{
			Spec: v1beta1.APIRuleSpec{
				Service: getApiRuleService(sampleServiceName, uint32(8080)),
				Host:    getHost(sampleValidHost),
				Rules: []v1beta1.Rule{
					{
						Path: "/abc",
						AccessStrategies: []*v1beta1.Authenticator{
							toAuthenticator("noop", emptyConfig()),
						},
						Methods: []v1beta1.HttpMethod{http.MethodGet},
					},
					{
						Path: "/abc",
						AccessStrategies: []*v1beta1.Authenticator{
							toAuthenticator("anonymous", emptyConfig()),
						},
						Methods: []v1beta1.HttpMethod{http.MethodGet, http.MethodPost},
					},
				},
			},
		}

		service := getService(sampleServiceName)
		fakeClient := buildFakeClient(service)

		//when
		problems := (&APIRuleValidator{
			ApiRule:                   apiRule,
			AccessStrategiesValidator: asValidatorMock,
			DomainAllowList:           testDomainAllowlist,
		}).Validate(context.Background(), fakeClient, networkingv1beta1.VirtualServiceList{}, networkingv1beta1.GatewayList{})

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
		existingVS.Spec.Hosts = []string{occupiedHost}

		apiRule := &v1beta1.APIRule{
			ObjectMeta: metav1.ObjectMeta{
				UID: "67890",
			},
			Spec: v1beta1.APIRuleSpec{
				Service: getApiRuleService(sampleServiceName, uint32(8080)),
				Host:    getHost(notOccupiedHost),
				Rules: []v1beta1.Rule{
					{
						Path: "/abc",
						AccessStrategies: []*v1beta1.Authenticator{
							toAuthenticator("jwt", simpleJWTConfig()),
						},
						Methods: []v1beta1.HttpMethod{http.MethodPost},
					},
					{
						Path: "/abc",
						AccessStrategies: []*v1beta1.Authenticator{
							toAuthenticator("jwt", simpleJWTConfig()),
						},
						Methods: []v1beta1.HttpMethod{http.MethodGet},
					},
					{
						Path: "/bcd",
						AccessStrategies: []*v1beta1.Authenticator{
							toAuthenticator("anonymous", emptyConfig()),
						},
					},
					{
						Path: "/def",
						AccessStrategies: []*v1beta1.Authenticator{
							toAuthenticator("allow", nil),
						},
					},
				},
			},
		}

		service := getService(sampleServiceName)
		fakeClient := buildFakeClient(service)

		//when
		problems := (&APIRuleValidator{
			ApiRule:                   apiRule,
			HandlerValidator:          handlerValidatorMock,
			AccessStrategiesValidator: asValidatorMock,
			DomainAllowList:           testDomainAllowlist,
		}).Validate(context.Background(), fakeClient, networkingv1beta1.VirtualServiceList{Items: []*networkingv1beta1.VirtualService{&existingVS}}, networkingv1beta1.GatewayList{})

		//then
		Expect(problems).To(HaveLen(0))
	})

	It("Should succeed for the same path but different methods", func() {
		//given
		occupiedHost := "occupied-host" + allowlistedDomain
		notOccupiedHost := "not-occupied-host" + allowlistedDomain
		existingVS := networkingv1beta1.VirtualService{}
		existingVS.Spec.Hosts = []string{occupiedHost}

		apiRule := &v1beta1.APIRule{
			ObjectMeta: metav1.ObjectMeta{
				UID: "67890",
			},
			Spec: v1beta1.APIRuleSpec{
				Service: getApiRuleService(sampleServiceName, uint32(8080)),
				Host:    getHost(notOccupiedHost),
				Rules: []v1beta1.Rule{
					{
						Path: "/abc",
						AccessStrategies: []*v1beta1.Authenticator{
							toAuthenticator("jwt", simpleJWTConfig()),
						},
						Methods: []v1beta1.HttpMethod{http.MethodPost},
					},
					{
						Path: "/abc",
						AccessStrategies: []*v1beta1.Authenticator{
							toAuthenticator("jwt", simpleJWTConfig()),
						},
						Methods: []v1beta1.HttpMethod{http.MethodGet},
					},
				},
			},
		}

		service := getService(sampleServiceName)
		fakeClient := buildFakeClient(service)

		//when
		problems := (&APIRuleValidator{
			ApiRule:                   apiRule,
			HandlerValidator:          handlerValidatorMock,
			AccessStrategiesValidator: asValidatorMock,
			DomainAllowList:           testDomainAllowlist,
		}).Validate(context.Background(), fakeClient, networkingv1beta1.VirtualServiceList{Items: []*networkingv1beta1.VirtualService{&existingVS}}, networkingv1beta1.GatewayList{})

		//then
		Expect(problems).To(HaveLen(0))
	})

	It("Should fail for noop and jwt access strategy on the same path", func() {
		//given
		apiRule := &v1beta1.APIRule{
			Spec: v1beta1.APIRuleSpec{
				Service: getApiRuleService(sampleServiceName, uint32(8080)),
				Host:    getHost(sampleValidHost),
				Rules: []v1beta1.Rule{
					{
						Path: "/abc",
						AccessStrategies: []*v1beta1.Authenticator{
							toAuthenticator("noop", emptyConfig()),
							toAuthenticator("jwt", simpleJWTConfig()),
						},
					},
				},
			},
		}

		service := getService(sampleServiceName)
		service.Spec.Selector = map[string]string{}
		fakeClient := buildFakeClient(service)

		//when
		problems := (&APIRuleValidator{
			ApiRule:                   apiRule,
			HandlerValidator:          handlerValidatorMock,
			AccessStrategiesValidator: asValidatorMock,
			DomainAllowList:           testDomainAllowlist,
		}).Validate(context.Background(), fakeClient, networkingv1beta1.VirtualServiceList{}, networkingv1beta1.GatewayList{})

		// then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal(".spec.rules[0].accessStrategies"))
		Expect(problems[0].Message).To(Equal("Secure access strategies cannot be used in combination with unsecure access strategies"))
	})

	It("Should fail for allow and jwt access strategy on the same path", func() {
		//given
		apiRule := &v1beta1.APIRule{
			Spec: v1beta1.APIRuleSpec{
				Service: getApiRuleService(sampleServiceName, uint32(8080)),
				Host:    getHost(sampleValidHost),
				Rules: []v1beta1.Rule{
					{
						Path: "/abc",
						AccessStrategies: []*v1beta1.Authenticator{
							toAuthenticator("allow", emptyConfig()),
							toAuthenticator("jwt", simpleJWTConfig()),
						},
					},
				},
			},
		}

		service := getService(sampleServiceName)
		service.Spec.Selector = map[string]string{}
		fakeClient := buildFakeClient(service)

		//when
		problems := (&APIRuleValidator{
			ApiRule:                   apiRule,
			HandlerValidator:          handlerValidatorMock,
			AccessStrategiesValidator: asValidatorMock,
			DomainAllowList:           testDomainAllowlist,
		}).Validate(context.Background(), fakeClient, networkingv1beta1.VirtualServiceList{}, networkingv1beta1.GatewayList{})

		// then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal(".spec.rules[0].accessStrategies"))
		Expect(problems[0].Message).To(Equal("Secure access strategies cannot be used in combination with unsecure access strategies"))
	})

	It("Should not fail with service without labels selector by default", func() {
		//given
		apiRule := &v1beta1.APIRule{
			Spec: v1beta1.APIRuleSpec{
				Rules: []v1beta1.Rule{
					{
						Path: "/abc",
						AccessStrategies: []*v1beta1.Authenticator{
							toAuthenticator("noop", emptyConfig()),
						},
						Methods: []v1beta1.HttpMethod{http.MethodPost},
					},
				},
				Service: getApiRuleService(sampleServiceName, uint32(8080)),
				Host:    getHost(sampleValidHost),
			},
		}

		service := getService(sampleServiceName)
		service.Spec.Selector = map[string]string{}
		fakeClient := buildFakeClient(service)

		//when
		problems := (&APIRuleValidator{
			ApiRule:                   apiRule,
			HandlerValidator:          handlerValidatorMock,
			AccessStrategiesValidator: asValidatorMock,
			DomainAllowList:           testDomainAllowlist,
		}).Validate(context.Background(), fakeClient, networkingv1beta1.VirtualServiceList{}, networkingv1beta1.GatewayList{})

		//then
		Expect(problems).To(HaveLen(0))
	})

	It("Should not fail with service on path level by default", func() {
		//given
		apiRule := &v1beta1.APIRule{
			Spec: v1beta1.APIRuleSpec{
				Rules: []v1beta1.Rule{
					{
						Path:    "/abc",
						Service: getApiRuleService(sampleServiceName, uint32(8080)),
						AccessStrategies: []*v1beta1.Authenticator{
							toAuthenticator("noop", emptyConfig()),
						},
						Methods: []v1beta1.HttpMethod{http.MethodPost},
					},
				},
				Host: getHost(sampleValidHost),
			},
		}

		service := getService(sampleServiceName)
		service.Spec.Selector = map[string]string{}
		fakeClient := buildFakeClient(service)

		//when
		problems := (&APIRuleValidator{
			ApiRule:                   apiRule,
			HandlerValidator:          handlerValidatorMock,
			AccessStrategiesValidator: asValidatorMock,
			DomainAllowList:           testDomainAllowlist,
		}).Validate(context.Background(), fakeClient, networkingv1beta1.VirtualServiceList{}, networkingv1beta1.GatewayList{})

		//then
		Expect(problems).To(HaveLen(0))
	})

	It("Should succeed with service without namespace", func() {
		//given
		apiRule := &v1beta1.APIRule{
			Spec: v1beta1.APIRuleSpec{
				Rules: []v1beta1.Rule{
					{
						Path: "/abc",
						AccessStrategies: []*v1beta1.Authenticator{
							toAuthenticator("noop", emptyConfig()),
						},
						Methods: []v1beta1.HttpMethod{http.MethodPost},
					},
				},
				Service: getApiRuleService(sampleServiceName, uint32(8080)),
				Host:    getHost(sampleValidHost),
			},
		}

		service := getService(sampleServiceName)
		fakeClient := buildFakeClient(service)

		//when
		problems := (&APIRuleValidator{
			ApiRule:                   apiRule,
			HandlerValidator:          handlerValidatorMock,
			AccessStrategiesValidator: asValidatorMock,
			DomainAllowList:           testDomainAllowlist,
		}).Validate(context.Background(), fakeClient, networkingv1beta1.VirtualServiceList{}, networkingv1beta1.GatewayList{})

		//then
		Expect(problems).To(HaveLen(0))
	})

	It("Should succeed with service on path level without namespace", func() {
		//given
		apiRule := &v1beta1.APIRule{
			Spec: v1beta1.APIRuleSpec{
				Rules: []v1beta1.Rule{
					{
						Path:    "/abc",
						Service: getApiRuleService(sampleServiceName, uint32(8080)),
						AccessStrategies: []*v1beta1.Authenticator{
							toAuthenticator("noop", emptyConfig()),
						},
						Methods: []v1beta1.HttpMethod{http.MethodPost},
					},
				},
				Host: getHost(sampleValidHost),
			},
		}

		service := getService(sampleServiceName)
		fakeClient := buildFakeClient(service)

		//when
		problems := (&APIRuleValidator{
			ApiRule:                   apiRule,
			HandlerValidator:          handlerValidatorMock,
			AccessStrategiesValidator: asValidatorMock,
			DomainAllowList:           testDomainAllowlist,
		}).Validate(context.Background(), fakeClient, networkingv1beta1.VirtualServiceList{}, networkingv1beta1.GatewayList{})

		//then
		Expect(problems).To(HaveLen(0))
	})
})

var _ = Describe("Validator for", func() {
	Describe("NoConfig access strategy", func() {
		It("Should fail with non-empty config", func() {
			//given
			handler := &v1beta1.Handler{Name: "noop", Config: simpleJWTConfig("http://atgo.org")}

			//when
			problems := (&noConfigAccStrValidator{}).Validate("some.attribute", handler)

			//then
			Expect(problems).NotTo(BeNil())
			Expect(problems).To(HaveLen(1))
			Expect(problems[0].AttributePath).To(Equal("some.attribute.config"))
			Expect(problems[0].Message).To(Equal("strategy: noop does not support configuration"))
		})

		It("Should succeed with empty config: {}", func() {
			//given
			handler := &v1beta1.Handler{Name: "noop", Config: emptyConfig()}

			//when
			problems := (&noConfigAccStrValidator{}).Validate("some.attribute", handler)

			//then
			Expect(problems).To(HaveLen(0))
		})

		It("Should succeed with null config", func() {
			//given
			handler := &v1beta1.Handler{Name: "noop", Config: nil}

			//when
			problems := (&noConfigAccStrValidator{}).Validate("some.attribute", handler)

			//then
			Expect(problems).To(HaveLen(0))
		})
	})

})

func buildFakeClient(objs ...client.Object) client.Client {
	scheme := runtime.NewScheme()
	err := networkingv1beta1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())
	err = rulev1alpha1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())
	err = securityv1beta1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())
	err = v1beta1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())
	err = corev1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	return fake.NewClientBuilder().WithScheme(scheme).WithObjects(objs...).Build()
}

func getService(name string, namespace ...string) *corev1.Service {
	svc := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": name,
			},
		},
	}
	if len(namespace) > 0 {
		svc.Namespace = namespace[0]
	}
	if svc.Namespace == "" {
		svc.Namespace = "default"
	}
	return &svc
}

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

func toAuthenticator(name string, config *runtime.RawExtension) *v1beta1.Authenticator {
	return &v1beta1.Authenticator{
		Handler: &v1beta1.Handler{
			Name:   name,
			Config: config,
		},
	}
}

func getApiRuleService(serviceName string, servicePort uint32, namespace ...*string) *v1beta1.Service {
	svc := v1beta1.Service{
		Name: &serviceName,
		Port: &servicePort,
	}
	if len(namespace) > 0 {
		svc.Namespace = namespace[0]
	}
	return &svc
}

func getHost(host string) *string {
	return &host
}
