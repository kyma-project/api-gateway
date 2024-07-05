package v2alpha1

import (
	"context"
	"fmt"
	"github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	rulev1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
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
	testDomainAllowlist = []string{"foo.bar", "bar.foo", "kyma.local"}
)

var _ = Describe("Validate function", func() {
	It("Should fail for empty rules", func() {
		//given
		testAllowList := []string{"foo.bar", "bar.foo", "kyma.local"}
		apiRule := &v2alpha1.APIRule{
			Spec: v2alpha1.APIRuleSpec{
				Rules:   nil,
				Service: getApiRuleService(sampleServiceName, uint32(8080)),
				Hosts:   getHosts(sampleValidHost),
			},
		}

		service := getService(sampleServiceName)
		fakeClient := buildFakeClient(service)

		//when
		problems := (&APIRuleValidator{
			Api:             apiRule,
			DomainAllowList: testAllowList,
		}).Validate(context.TODO(), fakeClient, networkingv1beta1.VirtualServiceList{})

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal(".spec.rules"))
		Expect(problems[0].Message).To(Equal("No rules defined"))
	})

	It("Should fail for blocklisted service", func() {
		//given
		sampleBlocklistedService := "kubernetes"
		validHost := v2alpha1.Host(sampleBlocklistedService + "." + allowlistedDomain)
		testBlockList := map[string][]string{
			"default": {sampleBlocklistedService, "kube-dns"},
			"example": {"service"}}
		apiRule := &v2alpha1.APIRule{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
			},
			Spec: v2alpha1.APIRuleSpec{
				Service: getApiRuleService(sampleBlocklistedService, uint32(443)),
				Hosts:   getHosts(validHost),
				Rules: []v2alpha1.Rule{
					{
						Path: "/abc",
						Jwt: &v2alpha1.JwtConfig{
							Authentications: []*v2alpha1.JwtAuthentication{
								{
									Issuer:  "",
									JwksUri: "",
								},
							},
						},
					},
				},
			}}

		service := getService(sampleBlocklistedService)
		fakeClient := buildFakeClient(service)

		//when
		problems := (&APIRuleValidator{
			Api:              apiRule,
			ServiceBlockList: testBlockList,
			DomainAllowList:  testDomainAllowlist,
		}).Validate(context.TODO(), fakeClient, networkingv1beta1.VirtualServiceList{})

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
		apiRule := &v2alpha1.APIRule{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
			},
			Spec: v2alpha1.APIRuleSpec{
				Service: getApiRuleService(sampleBlocklistedService, uint32(443), &sampleBlocklistedNamespace),
				Hosts:   getHosts(v2alpha1.Host(validHost)),
				Rules: []v2alpha1.Rule{
					{
						Path: "/abc",
						Jwt: &v2alpha1.JwtConfig{
							Authentications: []*v2alpha1.JwtAuthentication{
								{
									Issuer:  "",
									JwksUri: "",
								},
							},
						},
					},
				},
			}}

		service := getService(sampleBlocklistedService, sampleBlocklistedNamespace)
		fakeClient := buildFakeClient(service)

		//when
		problems := (&APIRuleValidator{
			Api:              apiRule,
			ServiceBlockList: testBlockList,
			DomainAllowList:  testDomainAllowlist,
		}).Validate(context.TODO(), fakeClient, networkingv1beta1.VirtualServiceList{})

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
		apiRule := &v2alpha1.APIRule{
			Spec: v2alpha1.APIRuleSpec{
				Service: getApiRuleService(sampleServiceName, uint32(8080)),
				Hosts:   getHosts(v2alpha1.Host(invalidHost)),
				Rules: []v2alpha1.Rule{
					{
						Path: "/abc",
						Jwt: &v2alpha1.JwtConfig{
							Authentications: []*v2alpha1.JwtAuthentication{
								{
									Issuer:  "",
									JwksUri: "",
								},
							},
						},
					},
				},
			}}

		service := getService(sampleServiceName)
		fakeClient := buildFakeClient(service)

		//when
		problems := (&APIRuleValidator{
			Api:              apiRule,
			ServiceBlockList: testBlockList,
			DomainAllowList:  testDomainAllowlist,
		}).Validate(context.TODO(), fakeClient, networkingv1beta1.VirtualServiceList{})

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal(".spec.host"))
		Expect(problems[0].Message).To(Equal("Host is not allowlisted"))
	})

	It("Should fail for blocklisted subdomain with default domainName (FQDN)", func() {
		//given
		blocklistedSubdomain := "api"
		blockedHost := blocklistedSubdomain + "." + testDefaultDomain
		testBlockList := map[string][]string{
			"default": {"kubernetes", "kube-dns"},
			"example": {"service"}}
		testHostBlockList := []string{blockedHost}
		apiRule := &v2alpha1.APIRule{
			Spec: v2alpha1.APIRuleSpec{
				Service: getApiRuleService(sampleServiceName, uint32(8080)),
				Hosts:   getHosts(v2alpha1.Host(blockedHost)),
				Rules: []v2alpha1.Rule{
					{
						Path: "/abc",
						Jwt: &v2alpha1.JwtConfig{
							Authentications: []*v2alpha1.JwtAuthentication{
								{
									Issuer:  "",
									JwksUri: "",
								},
							},
						},
					},
				},
			}}

		service := getService(sampleServiceName)
		fakeClient := buildFakeClient(service)

		//when
		problems := (&APIRuleValidator{
			Api:               apiRule,
			ServiceBlockList:  testBlockList,
			DomainAllowList:   testDomainAllowlist,
			HostBlockList:     testHostBlockList,
			DefaultDomainName: testDefaultDomain,
		}).Validate(context.TODO(), fakeClient, networkingv1beta1.VirtualServiceList{})

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal(".spec.host"))
		Expect(problems[0].Message).To(Equal(fmt.Sprintf("The subdomain %s is blocklisted for %s domain", blocklistedSubdomain, testDefaultDomain)))
	})

	It("Should NOT fail for blocklisted subdomain with custom domainName", func() {
		//given
		blocklistedSubdomain := "api"
		customDomainName := "bar.foo"
		blockedHost := blocklistedSubdomain + "." + testDefaultDomain
		customHost := blocklistedSubdomain + "." + customDomainName
		testBlockList := map[string][]string{
			"default": {"kubernetes", "kube-dns"},
			"example": {"service"}}
		testHostBlockList := []string{blockedHost}
		apiRule := &v2alpha1.APIRule{
			Spec: v2alpha1.APIRuleSpec{
				Service: getApiRuleService(sampleServiceName, uint32(8080)),
				Hosts:   getHosts(v2alpha1.Host(customHost)),
				Rules: []v2alpha1.Rule{
					{
						Path: "/abc",
						Jwt: &v2alpha1.JwtConfig{
							Authentications: []*v2alpha1.JwtAuthentication{
								{
									Issuer:  "",
									JwksUri: "",
								},
							},
						},
					},
				},
			}}

		service := getService(sampleServiceName)
		fakeClient := buildFakeClient(service)

		//when
		problems := (&APIRuleValidator{
			Api:               apiRule,
			ServiceBlockList:  testBlockList,
			DomainAllowList:   testDomainAllowlist,
			HostBlockList:     testHostBlockList,
			DefaultDomainName: testDefaultDomain,
		}).Validate(context.TODO(), fakeClient, networkingv1beta1.VirtualServiceList{})

		//then
		Expect(problems).To(HaveLen(0))
	})

	It("Should NOT fail for empty allowlisted domain", func() {
		//given
		validHost := sampleServiceName + "." + notAllowlistedDomain
		testBlockList := map[string][]string{
			"default": {"kubernetes", "kube-dns"},
			"example": {"service"}}
		apiRule := &v2alpha1.APIRule{
			Spec: v2alpha1.APIRuleSpec{
				Service: getApiRuleService(sampleServiceName, uint32(8080)),
				Hosts:   getHosts(v2alpha1.Host(validHost)),
				Rules: []v2alpha1.Rule{
					{
						Path: "/abc",
						Jwt: &v2alpha1.JwtConfig{
							Authentications: []*v2alpha1.JwtAuthentication{
								{
									Issuer:  "",
									JwksUri: "",
								},
							},
						},
					},
				},
			}}

		service := getService(sampleServiceName)
		fakeClient := buildFakeClient(service)

		//when
		problems := (&APIRuleValidator{
			Api:              apiRule,
			ServiceBlockList: testBlockList,
			DomainAllowList:  []string{},
		}).Validate(context.TODO(), fakeClient, networkingv1beta1.VirtualServiceList{})

		//then
		Expect(problems).To(HaveLen(0))
	})

	It("Should fail for not allowlisted domain containing allowlisted domain", func() {
		//given
		invalidHost := sampleServiceName + "." + allowlistedDomain + "." + notAllowlistedDomain
		testBlockList := map[string][]string{
			"default": {"kubernetes", "kube-dns"},
			"example": {"service"}}
		apiRule := &v2alpha1.APIRule{
			Spec: v2alpha1.APIRuleSpec{
				Service: getApiRuleService(sampleServiceName, uint32(8080)),
				Hosts:   getHosts(v2alpha1.Host(invalidHost)),
				Rules: []v2alpha1.Rule{
					{
						Path: "/abc",
						Jwt: &v2alpha1.JwtConfig{
							Authentications: []*v2alpha1.JwtAuthentication{
								{
									Issuer:  "",
									JwksUri: "",
								},
							},
						},
					},
				},
			}}

		service := getService(sampleServiceName)
		fakeClient := buildFakeClient(service)

		//when
		problems := (&APIRuleValidator{
			Api:              apiRule,
			ServiceBlockList: testBlockList,
			DomainAllowList:  testDomainAllowlist,
		}).Validate(context.TODO(), fakeClient, networkingv1beta1.VirtualServiceList{})

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
		apiRule := &v2alpha1.APIRule{
			Spec: v2alpha1.APIRuleSpec{
				Service: getApiRuleService(sampleServiceName, uint32(8080)),
				Hosts:   getHosts(v2alpha1.Host(hostWithoutDomain)),
				Rules: []v2alpha1.Rule{
					{
						Path: "/abc",
						Jwt: &v2alpha1.JwtConfig{
							Authentications: []*v2alpha1.JwtAuthentication{
								{
									Issuer:  "",
									JwksUri: "",
								},
							},
						},
					},
				},
			}}

		service := getService(sampleServiceName)
		fakeClient := buildFakeClient(service)

		//when
		problems := (&APIRuleValidator{
			Api:              apiRule,
			ServiceBlockList: testBlockList,
			DomainAllowList:  testDomainAllowlist,
		}).Validate(context.TODO(), fakeClient, networkingv1beta1.VirtualServiceList{})

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
		apiRule := &v2alpha1.APIRule{
			Spec: v2alpha1.APIRuleSpec{
				Service: getApiRuleService(sampleServiceName, uint32(8080)),
				Hosts:   getHosts(v2alpha1.Host(hostWithoutDomain)),
				Rules: []v2alpha1.Rule{
					{
						Path: "/abc",
						Jwt: &v2alpha1.JwtConfig{
							Authentications: []*v2alpha1.JwtAuthentication{
								{
									Issuer:  "",
									JwksUri: "",
								},
							},
						},
					},
				},
			}}

		service := getService(sampleServiceName)
		fakeClient := buildFakeClient(service)

		//when
		problems := (&APIRuleValidator{
			Api:               apiRule,
			ServiceBlockList:  testBlockList,
			DomainAllowList:   testDomainAllowlist,
			DefaultDomainName: testDefaultDomain,
		}).Validate(context.TODO(), fakeClient, networkingv1beta1.VirtualServiceList{})

		//then
		Expect(problems).To(HaveLen(0))
	})

	It("Should fail for serviceHost containing duplicated allowlisted domain", func() {
		//given
		invalidHost := sampleServiceName + "." + allowlistedDomain + "." + allowlistedDomain
		testBlockList := map[string][]string{
			"default": {"kubernetes", "kube-dns"},
			"example": {"service"}}
		apiRule := &v2alpha1.APIRule{
			Spec: v2alpha1.APIRuleSpec{
				Service: getApiRuleService(sampleServiceName, uint32(8080)),
				Hosts:   getHosts(v2alpha1.Host(invalidHost)),
				Rules: []v2alpha1.Rule{
					{
						Path: "/abc",
						Jwt: &v2alpha1.JwtConfig{
							Authentications: []*v2alpha1.JwtAuthentication{
								{
									Issuer:  "",
									JwksUri: "",
								},
							},
						},
					},
				},
			}}

		service := getService(sampleServiceName)
		fakeClient := buildFakeClient(service)

		//when
		problems := (&APIRuleValidator{
			Api:              apiRule,
			ServiceBlockList: testBlockList,
			DomainAllowList:  testDomainAllowlist,
		}).Validate(context.TODO(), fakeClient, networkingv1beta1.VirtualServiceList{})

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

		apiRule := &v2alpha1.APIRule{
			ObjectMeta: metav1.ObjectMeta{
				UID: "67890",
			},
			Spec: v2alpha1.APIRuleSpec{
				Service: getApiRuleService(sampleServiceName, uint32(8080)),
				Hosts:   getHosts(v2alpha1.Host(occupiedHost)),
				Rules: []v2alpha1.Rule{
					{
						Path: "/abc",
						Jwt: &v2alpha1.JwtConfig{
							Authentications: []*v2alpha1.JwtAuthentication{
								{
									Issuer:  "",
									JwksUri: "",
								},
							},
						},
					},
				},
			},
		}

		service := getService(sampleServiceName)
		fakeClient := buildFakeClient(service)

		//when
		problems := (&APIRuleValidator{
			Api:             apiRule,
			DomainAllowList: testDomainAllowlist,
		}).Validate(context.TODO(), fakeClient, networkingv1beta1.VirtualServiceList{Items: []*networkingv1beta1.VirtualService{&existingVS}})

		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal(".spec.host"))
		Expect(problems[0].Message).To(Equal("This host is occupied by another Virtual Service"))
	})

	It("Should NOT fail for a host that is occupied by a VS exposed by this resource", func() {
		//given
		occupiedHost := "occupied-host" + allowlistedDomain

		apiRule := &v2alpha1.APIRule{
			ObjectMeta: metav1.ObjectMeta{
				UID: "12345",
			},
			Spec: v2alpha1.APIRuleSpec{
				Service: getApiRuleService(sampleServiceName, uint32(8080)),
				Hosts:   getHosts(v2alpha1.Host(occupiedHost)),
				Rules: []v2alpha1.Rule{
					{
						Path: "/abc",
						Jwt: &v2alpha1.JwtConfig{
							Authentications: []*v2alpha1.JwtAuthentication{
								{
									Issuer:  "",
									JwksUri: "",
								},
							},
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
			Api:             apiRule,
			DomainAllowList: testDomainAllowlist,
		}).Validate(context.TODO(), fakeClient, networkingv1beta1.VirtualServiceList{Items: []*networkingv1beta1.VirtualService{&existingVS}})

		Expect(problems).To(HaveLen(0))
	})

	It("Should return an error when no service is defined for rule with no service on spec level", func() {
		//given
		apiRule := &v2alpha1.APIRule{
			Spec: v2alpha1.APIRuleSpec{
				Hosts: getHosts(sampleValidHost),
				Rules: []v2alpha1.Rule{
					{
						Path: "/abc",
						Jwt: &v2alpha1.JwtConfig{
							Authentications: []*v2alpha1.JwtAuthentication{
								{
									Issuer:  "",
									JwksUri: "",
								},
							},
						},
					},
				},
			},
		}

		service := getService(sampleServiceName, "default")
		fakeClient := buildFakeClient(service)

		//when
		problems := (&APIRuleValidator{
			Api:             apiRule,
			DomainAllowList: testDomainAllowlist,
		}).Validate(context.TODO(), fakeClient, networkingv1beta1.VirtualServiceList{})

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
		noAuth := true

		apiRule := &v2alpha1.APIRule{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
			},
			Spec: v2alpha1.APIRuleSpec{
				Hosts: getHosts(v2alpha1.Host(validHost)),
				Rules: []v2alpha1.Rule{
					{
						Path:    "/abc",
						NoAuth:  &noAuth,
						Service: getApiRuleService(sampleServiceName, uint32(8080)),
					},
					{
						Path:    "/abcd",
						NoAuth:  &noAuth,
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
			Api:              apiRule,
			ServiceBlockList: testBlockList,
			DomainAllowList:  testDomainAllowlist,
		}).Validate(context.TODO(), fakeClient, networkingv1beta1.VirtualServiceList{})

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
		noAuth := true

		apiRule := &v2alpha1.APIRule{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
			},
			Spec: v2alpha1.APIRuleSpec{
				Hosts: getHosts(v2alpha1.Host(validHost)),
				Rules: []v2alpha1.Rule{
					{
						Path:    "/abc",
						NoAuth:  &noAuth,
						Service: getApiRuleService(sampleServiceName, uint32(8080)),
					},
					{
						Path:    "/abcd",
						NoAuth:  &noAuth,
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
			Api:              apiRule,
			ServiceBlockList: testBlockList,
			DomainAllowList:  testDomainAllowlist,
		}).Validate(context.TODO(), fakeClient, networkingv1beta1.VirtualServiceList{})

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal(".spec.rules[1].service.name"))
		Expect(problems[0].Message).To(Equal(fmt.Sprintf("Service %s in namespace %s is blocklisted", sampleBlocklistedService, sampleBlocklistedNamespace)))
	})

	It("Should detect several problems", func() {
		//given
		noAuth := true
		apiRule := &v2alpha1.APIRule{
			Spec: v2alpha1.APIRuleSpec{
				Service: getApiRuleService(sampleServiceName, uint32(8080)),
				Hosts:   getHosts(sampleValidHost),
				Rules: []v2alpha1.Rule{
					{
						Path:   "/abc",
						NoAuth: &noAuth,
					},
					{
						Path: "/abc",
						Jwt: &v2alpha1.JwtConfig{
							Authentications: []*v2alpha1.JwtAuthentication{
								{
									Issuer:  "",
									JwksUri: "",
								},
							},
						},
					},
				},
			},
		}

		service := getService(sampleServiceName)
		fakeClient := buildFakeClient(service)

		//when
		problems := (&APIRuleValidator{
			Api:             apiRule,
			DomainAllowList: testDomainAllowlist,
		}).Validate(context.TODO(), fakeClient, networkingv1beta1.VirtualServiceList{})

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
		noAuth := true

		apiRule := &v2alpha1.APIRule{
			ObjectMeta: metav1.ObjectMeta{
				UID: "67890",
			},
			Spec: v2alpha1.APIRuleSpec{
				Service: getApiRuleService(sampleServiceName, uint32(8080)),
				Hosts:   getHosts(v2alpha1.Host(notOccupiedHost)),
				Rules: []v2alpha1.Rule{
					{
						Path: "/abc",
						Jwt: &v2alpha1.JwtConfig{
							Authentications: []*v2alpha1.JwtAuthentication{
								{
									Issuer:  "",
									JwksUri: "",
								},
							},
						},
						Methods: []v2alpha1.HttpMethod{http.MethodPost},
					},
					{
						Path: "/abc",
						Jwt: &v2alpha1.JwtConfig{
							Authentications: []*v2alpha1.JwtAuthentication{
								{
									Issuer:  "",
									JwksUri: "",
								},
							},
						},
						Methods: []v2alpha1.HttpMethod{http.MethodGet},
					},
					{
						Path:   "/bcd",
						NoAuth: &noAuth,
					},
					{
						Path:   "/def",
						NoAuth: &noAuth,
					},
				},
			},
		}

		service := getService(sampleServiceName)
		fakeClient := buildFakeClient(service)

		//when
		problems := (&APIRuleValidator{
			Api:             apiRule,
			DomainAllowList: testDomainAllowlist,
		}).Validate(context.TODO(), fakeClient, networkingv1beta1.VirtualServiceList{Items: []*networkingv1beta1.VirtualService{&existingVS}})

		//then
		Expect(problems).To(HaveLen(0))
	})

	It("Should succeed for the same path but different methods", func() {
		//given
		occupiedHost := "occupied-host" + allowlistedDomain
		notOccupiedHost := "not-occupied-host" + allowlistedDomain
		existingVS := networkingv1beta1.VirtualService{}
		existingVS.Spec.Hosts = []string{occupiedHost}

		apiRule := &v2alpha1.APIRule{
			ObjectMeta: metav1.ObjectMeta{
				UID: "67890",
			},
			Spec: v2alpha1.APIRuleSpec{
				Service: getApiRuleService(sampleServiceName, uint32(8080)),
				Hosts:   getHosts(v2alpha1.Host(notOccupiedHost)),
				Rules: []v2alpha1.Rule{
					{
						Path: "/abc",
						Jwt: &v2alpha1.JwtConfig{
							Authentications: []*v2alpha1.JwtAuthentication{
								{
									Issuer:  "",
									JwksUri: "",
								},
							},
						},
						Methods: []v2alpha1.HttpMethod{http.MethodPost},
					},
					{
						Path: "/abc",
						Jwt: &v2alpha1.JwtConfig{
							Authentications: []*v2alpha1.JwtAuthentication{
								{
									Issuer:  "",
									JwksUri: "",
								},
							},
						},
						Methods: []v2alpha1.HttpMethod{http.MethodGet},
					},
				},
			},
		}

		service := getService(sampleServiceName)
		fakeClient := buildFakeClient(service)

		//when
		problems := (&APIRuleValidator{
			Api:             apiRule,
			DomainAllowList: testDomainAllowlist,
		}).Validate(context.TODO(), fakeClient, networkingv1beta1.VirtualServiceList{Items: []*networkingv1beta1.VirtualService{&existingVS}})

		//then
		Expect(problems).To(HaveLen(0))
	})

	It("Should fail for noAuth and jwt access strategy on the same path", func() {
		//given
		noAuth := true
		apiRule := &v2alpha1.APIRule{
			Spec: v2alpha1.APIRuleSpec{
				Service: getApiRuleService(sampleServiceName, uint32(8080)),
				Hosts:   getHosts(sampleValidHost),
				Rules: []v2alpha1.Rule{
					{
						Path:   "/abc",
						NoAuth: &noAuth,
						Jwt: &v2alpha1.JwtConfig{
							Authentications: []*v2alpha1.JwtAuthentication{
								{
									Issuer:  "",
									JwksUri: "",
								},
							},
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
			Api:             apiRule,
			DomainAllowList: testDomainAllowlist,
		}).Validate(context.TODO(), fakeClient, networkingv1beta1.VirtualServiceList{})

		// then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal(".spec.rules[0].accessStrategies"))
		Expect(problems[0].Message).To(Equal("Secure access strategies cannot be used in combination with unsecure access strategies"))
	})

	It("Should not fail with service without labels selector by default", func() {
		//given
		noAuth := true
		apiRule := &v2alpha1.APIRule{
			Spec: v2alpha1.APIRuleSpec{
				Rules: []v2alpha1.Rule{
					{
						Path:    "/abc",
						NoAuth:  &noAuth,
						Methods: []v2alpha1.HttpMethod{http.MethodPost},
					},
				},
				Service: getApiRuleService(sampleServiceName, uint32(8080)),
				Hosts:   getHosts(sampleValidHost),
			},
		}

		service := getService(sampleServiceName)
		service.Spec.Selector = map[string]string{}
		fakeClient := buildFakeClient(service)

		//when
		problems := (&APIRuleValidator{
			Api:             apiRule,
			DomainAllowList: testDomainAllowlist,
		}).Validate(context.TODO(), fakeClient, networkingv1beta1.VirtualServiceList{})

		//then
		Expect(problems).To(HaveLen(0))
	})

	It("Should not fail with service on path level by default", func() {
		//given
		noAuth := true
		apiRule := &v2alpha1.APIRule{
			Spec: v2alpha1.APIRuleSpec{
				Rules: []v2alpha1.Rule{
					{
						Path:    "/abc",
						Service: getApiRuleService(sampleServiceName, uint32(8080)),
						NoAuth:  &noAuth,
						Methods: []v2alpha1.HttpMethod{http.MethodPost},
					},
				},
				Hosts: getHosts(sampleValidHost),
			},
		}

		service := getService(sampleServiceName)
		service.Spec.Selector = map[string]string{}
		fakeClient := buildFakeClient(service)

		//when
		problems := (&APIRuleValidator{
			Api:             apiRule,
			DomainAllowList: testDomainAllowlist,
		}).Validate(context.TODO(), fakeClient, networkingv1beta1.VirtualServiceList{})

		//then
		Expect(problems).To(HaveLen(0))
	})

	It("Should succeed with service without namespace", func() {
		//given
		noAuth := true
		apiRule := &v2alpha1.APIRule{
			Spec: v2alpha1.APIRuleSpec{
				Rules: []v2alpha1.Rule{
					{
						Path:    "/abc",
						NoAuth:  &noAuth,
						Methods: []v2alpha1.HttpMethod{http.MethodPost},
					},
				},
				Service: getApiRuleService(sampleServiceName, uint32(8080)),
				Hosts:   getHosts(sampleValidHost),
			},
		}

		service := getService(sampleServiceName)
		fakeClient := buildFakeClient(service)

		//when
		problems := (&APIRuleValidator{
			Api:             apiRule,
			DomainAllowList: testDomainAllowlist,
		}).Validate(context.TODO(), fakeClient, networkingv1beta1.VirtualServiceList{})

		//then
		Expect(problems).To(HaveLen(0))
	})

	It("Should succeed with service on path level without namespace", func() {
		//given
		noAuth := true
		apiRule := &v2alpha1.APIRule{
			Spec: v2alpha1.APIRuleSpec{
				Rules: []v2alpha1.Rule{
					{
						Path:    "/abc",
						Service: getApiRuleService(sampleServiceName, uint32(8080)),
						NoAuth:  &noAuth,
						Methods: []v2alpha1.HttpMethod{http.MethodPost},
					},
				},
				Hosts: getHosts(sampleValidHost),
			},
		}

		service := getService(sampleServiceName)
		fakeClient := buildFakeClient(service)

		//when
		problems := (&APIRuleValidator{
			Api:             apiRule,
			DomainAllowList: testDomainAllowlist,
		}).Validate(context.TODO(), fakeClient, networkingv1beta1.VirtualServiceList{})

		//then
		Expect(problems).To(HaveLen(0))
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
	err = v2alpha1.AddToScheme(scheme)
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
		svc.ObjectMeta.Namespace = namespace[0]
	}
	if svc.ObjectMeta.Namespace == "" {
		svc.ObjectMeta.Namespace = "default"
	}
	return &svc
}

func getApiRuleService(serviceName string, servicePort uint32, namespace ...*string) *v2alpha1.Service {
	svc := v2alpha1.Service{
		Name: &serviceName,
		Port: &servicePort,
	}
	if len(namespace) > 0 {
		svc.Namespace = namespace[0]
	}
	return &svc
}

func getHosts(host v2alpha1.Host) []*v2alpha1.Host {
	return []*v2alpha1.Host{&host}
}
