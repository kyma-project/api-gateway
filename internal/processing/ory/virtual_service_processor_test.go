package ory_test

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	"github.com/kyma-incubator/api-gateway/internal/processing"
	"github.com/kyma-incubator/api-gateway/internal/processing/ory"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	rulev1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
	"istio.io/api/networking/v1beta1"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	apiName                     = "test-apirule"
	apiUID            types.UID = "eab0f1c8-c417-11e9-bf11-4ac644044351"
	apiNamespace                = "some-namespace"
	apiAPIVersion               = "gateway.kyma-project.io/v1alpha1"
	apiKind                     = "ApiRule"
	apiPath                     = "/.*"
	headersAPIPath              = "/headers"
	jwtIssuer                   = "https://oauth2.example.com/"
	oathkeeperSvc               = "fake.oathkeeper"
	oathkeeperSvcPort uint32    = 1234
	testLabelKey                = "key"
	testLabelValue              = "value"
	defaultDomain               = "myDomain.com"
)

var (
	apiMethods                     = []string{"GET"}
	apiScopes                      = []string{"write", "read"}
	servicePort             uint32 = 8080
	apiGateway                     = "some-gateway"
	serviceName                    = "example-service"
	serviceHostWithNoDomain        = "myService"
	serviceHost                    = serviceHostWithNoDomain + "." + defaultDomain

	testAllowOrigin  = []*v1beta1.StringMatch{{MatchType: &v1beta1.StringMatch_Regex{Regex: ".*"}}}
	testAllowMethods = []string{"GET", "POST", "PUT", "DELETE"}
	testAllowHeaders = []string{"header1", "header2"}

	testCors = &processing.CorsConfig{
		AllowOrigins: testAllowOrigin,
		AllowMethods: testAllowMethods,
		AllowHeaders: testAllowHeaders,
	}

	testAdditionalLabels = map[string]string{testLabelKey: testLabelValue}
)

var _ = Describe("Virtual Service Processor", func() {
	Context("create", func() {
		When("handler is allow", func() {
			It("should create for allow authenticator", func() {
				// given
				strategies := []*gatewayv1beta1.Authenticator{
					{
						Handler: &gatewayv1beta1.Handler{
							Name: "allow",
						},
					},
				}

				allowRule := getRuleFor(apiPath, apiMethods, []*gatewayv1beta1.Mutator{}, strategies)
				rules := []gatewayv1beta1.Rule{allowRule}

				apiRule := getAPIRuleFor(rules)

				vsProcessor := ory.NewVirtualServiceProcessor(getConfigWithEmptyFakeClient())

				// when
				result, err := vsProcessor.EvaluateReconciliation(apiRule)

				// then
				Expect(err).To(BeNil())
				Expect(result).To(HaveLen(1))
				Expect(result[0].Action).To(Equal("create"))

				vs := result[0].Obj.(*networkingv1beta1.VirtualService)

				Expect(vs).NotTo(BeNil())
				Expect(len(vs.Spec.Gateways)).To(Equal(1))
				Expect(len(vs.Spec.Hosts)).To(Equal(1))
				Expect(vs.Spec.Hosts[0]).To(Equal(serviceHost))
				Expect(len(vs.Spec.Http)).To(Equal(1))

				Expect(len(vs.Spec.Http[0].Route)).To(Equal(1))
				Expect(vs.Spec.Http[0].Route[0].Destination.Host).To(Equal(serviceName + "." + apiNamespace + ".svc.cluster.local"))
				Expect(vs.Spec.Http[0].Route[0].Destination.Port.Number).To(Equal(servicePort))

				Expect(len(vs.Spec.Http[0].Match)).To(Equal(1))
				Expect(vs.Spec.Http[0].Match[0].Uri.GetRegex()).To(Equal(apiRule.Spec.Rules[0].Path))

				Expect(vs.Spec.Http[0].CorsPolicy.AllowOrigins).To(Equal(testCors.AllowOrigins))
				Expect(vs.Spec.Http[0].CorsPolicy.AllowMethods).To(Equal(testCors.AllowMethods))
				Expect(vs.Spec.Http[0].CorsPolicy.AllowHeaders).To(Equal(testCors.AllowHeaders))

				Expect(vs.ObjectMeta.Name).To(BeEmpty())
				Expect(vs.ObjectMeta.GenerateName).To(Equal(apiName + "-"))
				Expect(vs.ObjectMeta.Namespace).To(Equal(apiNamespace))
				Expect(vs.ObjectMeta.Labels[testLabelKey]).To(Equal(testLabelValue))

				Expect(vs.ObjectMeta.OwnerReferences[0].APIVersion).To(Equal(apiAPIVersion))
				Expect(vs.ObjectMeta.OwnerReferences[0].Kind).To(Equal(apiKind))
				Expect(vs.ObjectMeta.OwnerReferences[0].Name).To(Equal(apiName))
				Expect(vs.ObjectMeta.OwnerReferences[0].UID).To(Equal(apiUID))
			})

			It("should override destination host for specified spec level service namespace", func() {
				// given
				strategies := []*gatewayv1beta1.Authenticator{
					{
						Handler: &gatewayv1beta1.Handler{
							Name: "allow",
						},
					},
				}

				allowRule := getRuleFor(apiPath, apiMethods, []*gatewayv1beta1.Mutator{}, strategies)
				rules := []gatewayv1beta1.Rule{allowRule}

				apiRule := getAPIRuleFor(rules)

				overrideServiceName := "testName"
				overrideServiceNamespace := "testName-namespace"
				overrideServicePort := uint32(8080)

				apiRule.Spec.Service = &gatewayv1beta1.Service{
					Name:      &overrideServiceName,
					Namespace: &overrideServiceNamespace,
					Port:      &overrideServicePort,
				}

				// when
				vsProcessor := ory.NewVirtualServiceProcessor(getConfigWithEmptyFakeClient())

				// then
				result, err := vsProcessor.EvaluateReconciliation(apiRule)

				Expect(err).To(BeNil())
				Expect(result).To(HaveLen(1))

				vs := result[0].Obj.(*networkingv1beta1.VirtualService)

				Expect(len(vs.Spec.Http[0].Route)).To(Equal(1))
				Expect(vs.Spec.Http[0].Route[0].Destination.Host).To(Equal(overrideServiceName + "." + overrideServiceNamespace + ".svc.cluster.local"))
			})

			It("should override destination host with rule level service namespace", func() {
				// given
				strategies := []*gatewayv1beta1.Authenticator{
					{
						Handler: &gatewayv1beta1.Handler{
							Name: "allow",
						},
					},
				}

				overrideServiceName := "testName"
				overrideServiceNamespace := "testName-namespace"
				overrideServicePort := uint32(8080)

				service := &gatewayv1beta1.Service{
					Name:      &overrideServiceName,
					Namespace: &overrideServiceNamespace,
					Port:      &overrideServicePort,
				}

				allowRule := getRuleWithServiceFor(apiPath, apiMethods, []*gatewayv1beta1.Mutator{}, strategies, service)
				rules := []gatewayv1beta1.Rule{allowRule}

				apiRule := getAPIRuleFor(rules)

				// when
				vsProcessor := ory.NewVirtualServiceProcessor(getConfigWithEmptyFakeClient())

				// then
				result, err := vsProcessor.EvaluateReconciliation(apiRule)

				Expect(err).To(BeNil())
				Expect(result).To(HaveLen(1))

				vs := result[0].Obj.(*networkingv1beta1.VirtualService)

				//verify VS has rule level destination host
				Expect(len(vs.Spec.Http[0].Route)).To(Equal(1))
				Expect(vs.Spec.Http[0].Route[0].Destination.Host).To(Equal(overrideServiceName + "." + overrideServiceNamespace + ".svc.cluster.local"))

			})
			It("should produce VS with default domain name when the hostname does not contain domain name", func() {
				strategies := []*gatewayv1beta1.Authenticator{
					{
						Handler: &gatewayv1beta1.Handler{
							Name: "allow",
						},
					},
				}

				allowRule := getRuleFor(apiPath, apiMethods, []*gatewayv1beta1.Mutator{}, strategies)
				rules := []gatewayv1beta1.Rule{allowRule}

				apiRule := getAPIRuleFor(rules)
				apiRule.Spec.Host = &serviceHostWithNoDomain

				vsProcessor := ory.NewVirtualServiceProcessor(getConfigWithEmptyFakeClient())

				// when
				result, err := vsProcessor.EvaluateReconciliation(apiRule)

				// then
				Expect(err).To(BeNil())
				Expect(result).To(HaveLen(1))

				vs := result[0].Obj.(*networkingv1beta1.VirtualService)

				//verify VS
				Expect(vs).NotTo(BeNil())
				Expect(len(vs.Spec.Hosts)).To(Equal(1))
				Expect(vs.Spec.Hosts[0]).To(Equal(serviceHost))

			})
		})

		When("handler is noop", func() {
			It("should not override Oathkeeper service destination host with spec level service", func() {
				// given
				strategies := []*gatewayv1beta1.Authenticator{
					{
						Handler: &gatewayv1beta1.Handler{
							Name: "noop",
						},
					},
				}

				overrideServiceName := "testName"
				overrideServicePort := uint32(8080)

				service := &gatewayv1beta1.Service{
					Name: &overrideServiceName,
					Port: &overrideServicePort,
				}

				allowRule := getRuleWithServiceFor(apiPath, apiMethods, []*gatewayv1beta1.Mutator{}, strategies, service)
				rules := []gatewayv1beta1.Rule{allowRule}

				apiRule := getAPIRuleFor(rules)

				vsProcessor := ory.NewVirtualServiceProcessor(getConfigWithEmptyFakeClient())

				// when
				result, err := vsProcessor.EvaluateReconciliation(apiRule)

				// then
				Expect(err).To(BeNil())
				Expect(result).To(HaveLen(1))

				vs := result[0].Obj.(*networkingv1beta1.VirtualService)

				Expect(len(vs.Spec.Http[0].Route)).To(Equal(1))
				Expect(vs.Spec.Http[0].Route[0].Destination.Host).To(Equal(oathkeeperSvc))
			})
		})

		When("multiple handler", func() {
			It("should produce VS for given paths", func() {
				// given
				noop := []*gatewayv1beta1.Authenticator{
					{
						Handler: &gatewayv1beta1.Handler{
							Name: "noop",
						},
					},
				}

				jwtConfigJSON := fmt.Sprintf(`
						{
							"trusted_issuers": ["%s"],
							"jwks": [],
							"required_scope": [%s]
					}`, jwtIssuer, toCSVList(apiScopes))

				jwt := []*gatewayv1beta1.Authenticator{
					{
						Handler: &gatewayv1beta1.Handler{
							Name: "jwt",
							Config: &runtime.RawExtension{
								Raw: []byte(jwtConfigJSON),
							},
						},
					},
				}

				testMutators := []*gatewayv1beta1.Mutator{
					{
						Handler: &gatewayv1beta1.Handler{
							Name: "noop",
						},
					},
					{
						Handler: &gatewayv1beta1.Handler{
							Name: "idtoken",
						},
					},
				}

				noopRule := getRuleFor(apiPath, apiMethods, []*gatewayv1beta1.Mutator{}, noop)
				jwtRule := getRuleFor(headersAPIPath, apiMethods, testMutators, jwt)
				rules := []gatewayv1beta1.Rule{noopRule, jwtRule}

				apiRule := getAPIRuleFor(rules)

				vsProcessor := ory.NewVirtualServiceProcessor(getConfigWithEmptyFakeClient())

				// when
				result, err := vsProcessor.EvaluateReconciliation(apiRule)

				// then
				Expect(err).To(BeNil())
				Expect(result).To(HaveLen(1))

				vs := result[0].Obj.(*networkingv1beta1.VirtualService)

				Expect(vs).NotTo(BeNil())
				Expect(len(vs.Spec.Gateways)).To(Equal(1))
				Expect(len(vs.Spec.Hosts)).To(Equal(1))
				Expect(vs.Spec.Hosts[0]).To(Equal(serviceHost))
				Expect(len(vs.Spec.Http)).To(Equal(2))

				Expect(len(vs.Spec.Http[0].Route)).To(Equal(1))
				Expect(vs.Spec.Http[0].Route[0].Destination.Host).To(Equal(oathkeeperSvc))
				Expect(vs.Spec.Http[0].Route[0].Destination.Port.Number).To(Equal(oathkeeperSvcPort))
				Expect(len(vs.Spec.Http[0].Match)).To(Equal(1))
				Expect(vs.Spec.Http[0].Match[0].Uri.GetRegex()).To(Equal(apiRule.Spec.Rules[0].Path))

				Expect(vs.Spec.Http[0].CorsPolicy.AllowOrigins).To(Equal(testCors.AllowOrigins))
				Expect(vs.Spec.Http[0].CorsPolicy.AllowMethods).To(Equal(testCors.AllowMethods))
				Expect(vs.Spec.Http[0].CorsPolicy.AllowHeaders).To(Equal(testCors.AllowHeaders))

				Expect(len(vs.Spec.Http[1].Route)).To(Equal(1))
				Expect(vs.Spec.Http[1].Route[0].Destination.Host).To(Equal(oathkeeperSvc))
				Expect(vs.Spec.Http[1].Route[0].Destination.Port.Number).To(Equal(oathkeeperSvcPort))
				Expect(len(vs.Spec.Http[1].Match)).To(Equal(1))
				Expect(vs.Spec.Http[1].Match[0].Uri.GetRegex()).To(Equal(apiRule.Spec.Rules[1].Path))

				Expect(vs.Spec.Http[1].CorsPolicy.AllowOrigins).To(Equal(testCors.AllowOrigins))
				Expect(vs.Spec.Http[1].CorsPolicy.AllowMethods).To(Equal(testCors.AllowMethods))
				Expect(vs.Spec.Http[1].CorsPolicy.AllowHeaders).To(Equal(testCors.AllowHeaders))

				Expect(vs.ObjectMeta.Name).To(BeEmpty())
				Expect(vs.ObjectMeta.GenerateName).To(Equal(apiName + "-"))
				Expect(vs.ObjectMeta.Namespace).To(Equal(apiNamespace))
				Expect(vs.ObjectMeta.Labels[testLabelKey]).To(Equal(testLabelValue))

				Expect(vs.ObjectMeta.OwnerReferences[0].APIVersion).To(Equal(apiAPIVersion))
				Expect(vs.ObjectMeta.OwnerReferences[0].Kind).To(Equal(apiKind))
				Expect(vs.ObjectMeta.OwnerReferences[0].Name).To(Equal(apiName))
				Expect(vs.ObjectMeta.OwnerReferences[0].UID).To(Equal(apiUID))
			})

			It("should produce VS for two same paths and different methods", func() {
				// given
				noop := []*gatewayv1beta1.Authenticator{
					{
						Handler: &gatewayv1beta1.Handler{
							Name: "noop",
						},
					},
				}

				jwtConfigJSON := fmt.Sprintf(`
						{
							"trusted_issuers": ["%s"],
							"jwks": [],
							"required_scope": [%s]
					}`, jwtIssuer, toCSVList(apiScopes))

				jwt := []*gatewayv1beta1.Authenticator{
					{
						Handler: &gatewayv1beta1.Handler{
							Name: "jwt",
							Config: &runtime.RawExtension{
								Raw: []byte(jwtConfigJSON),
							},
						},
					},
				}

				testMutators := []*gatewayv1beta1.Mutator{
					{
						Handler: &gatewayv1beta1.Handler{
							Name: "noop",
						},
					},
					{
						Handler: &gatewayv1beta1.Handler{
							Name: "idtoken",
						},
					},
				}
				getMethod := []string{"GET"}
				postMethod := []string{"POST"}
				noopRule := getRuleFor(apiPath, getMethod, []*gatewayv1beta1.Mutator{}, noop)
				jwtRule := getRuleFor(apiPath, postMethod, testMutators, jwt)
				rules := []gatewayv1beta1.Rule{noopRule, jwtRule}

				apiRule := getAPIRuleFor(rules)

				vsProcessor := ory.NewVirtualServiceProcessor(getConfigWithEmptyFakeClient())

				// when
				result, err := vsProcessor.EvaluateReconciliation(apiRule)

				// then
				Expect(err).To(BeNil())
				Expect(result).To(HaveLen(1))

				vs := result[0].Obj.(*networkingv1beta1.VirtualService)

				Expect(vs).NotTo(BeNil())
				Expect(len(vs.Spec.Gateways)).To(Equal(1))
				Expect(len(vs.Spec.Hosts)).To(Equal(1))
				Expect(vs.Spec.Hosts[0]).To(Equal(serviceHost))
				Expect(len(vs.Spec.Http)).To(Equal(1))

				Expect(len(vs.Spec.Http[0].Route)).To(Equal(1))
				Expect(vs.Spec.Http[0].Route[0].Destination.Host).To(Equal(oathkeeperSvc))
				Expect(vs.Spec.Http[0].Route[0].Destination.Port.Number).To(Equal(oathkeeperSvcPort))
				Expect(len(vs.Spec.Http[0].Match)).To(Equal(1))
				Expect(vs.Spec.Http[0].Match[0].Uri.GetRegex()).To(Equal(apiRule.Spec.Rules[0].Path))

				Expect(vs.Spec.Http[0].CorsPolicy.AllowOrigins).To(Equal(testCors.AllowOrigins))
				Expect(vs.Spec.Http[0].CorsPolicy.AllowMethods).To(Equal(testCors.AllowMethods))
				Expect(vs.Spec.Http[0].CorsPolicy.AllowHeaders).To(Equal(testCors.AllowHeaders))

				Expect(vs.ObjectMeta.Name).To(BeEmpty())
				Expect(vs.ObjectMeta.GenerateName).To(Equal(apiName + "-"))
				Expect(vs.ObjectMeta.Namespace).To(Equal(apiNamespace))
				Expect(vs.ObjectMeta.Labels[testLabelKey]).To(Equal(testLabelValue))

				Expect(vs.ObjectMeta.OwnerReferences[0].APIVersion).To(Equal(apiAPIVersion))
				Expect(vs.ObjectMeta.OwnerReferences[0].Kind).To(Equal(apiKind))
				Expect(vs.ObjectMeta.OwnerReferences[0].Name).To(Equal(apiName))
				Expect(vs.ObjectMeta.OwnerReferences[0].UID).To(Equal(apiUID))
			})

			It("should produce VS for two same paths and one different", func() {
				// given
				noop := []*gatewayv1beta1.Authenticator{
					{
						Handler: &gatewayv1beta1.Handler{
							Name: "noop",
						},
					},
				}

				jwtConfigJSON := fmt.Sprintf(`
						{
							"trusted_issuers": ["%s"],
							"jwks": [],
							"required_scope": [%s]
					}`, jwtIssuer, toCSVList(apiScopes))

				jwt := []*gatewayv1beta1.Authenticator{
					{
						Handler: &gatewayv1beta1.Handler{
							Name: "jwt",
							Config: &runtime.RawExtension{
								Raw: []byte(jwtConfigJSON),
							},
						},
					},
				}

				testMutators := []*gatewayv1beta1.Mutator{
					{
						Handler: &gatewayv1beta1.Handler{
							Name: "noop",
						},
					},
					{
						Handler: &gatewayv1beta1.Handler{
							Name: "idtoken",
						},
					},
				}
				getMethod := []string{"GET"}
				postMethod := []string{"POST"}
				noopGetRule := getRuleFor(apiPath, getMethod, []*gatewayv1beta1.Mutator{}, noop)
				noopPostRule := getRuleFor(apiPath, postMethod, []*gatewayv1beta1.Mutator{}, noop)
				jwtRule := getRuleFor(headersAPIPath, apiMethods, testMutators, jwt)
				rules := []gatewayv1beta1.Rule{noopGetRule, noopPostRule, jwtRule}

				apiRule := getAPIRuleFor(rules)

				vsProcessor := ory.NewVirtualServiceProcessor(getConfigWithEmptyFakeClient())

				// when
				result, err := vsProcessor.EvaluateReconciliation(apiRule)

				// then
				Expect(err).To(BeNil())
				Expect(result).To(HaveLen(1))

				vs := result[0].Obj.(*networkingv1beta1.VirtualService)

				Expect(vs).NotTo(BeNil())
				Expect(len(vs.Spec.Gateways)).To(Equal(1))
				Expect(len(vs.Spec.Hosts)).To(Equal(1))
				Expect(vs.Spec.Hosts[0]).To(Equal(serviceHost))
				Expect(len(vs.Spec.Http)).To(Equal(2))

				Expect(len(vs.Spec.Http[0].Route)).To(Equal(1))
				Expect(vs.Spec.Http[0].Route[0].Destination.Host).To(Equal(oathkeeperSvc))
				Expect(vs.Spec.Http[0].Route[0].Destination.Port.Number).To(Equal(oathkeeperSvcPort))
				Expect(len(vs.Spec.Http[0].Match)).To(Equal(1))
				Expect(vs.Spec.Http[0].Match[0].Uri.GetRegex()).To(Equal(apiRule.Spec.Rules[0].Path))

				Expect(vs.Spec.Http[0].CorsPolicy.AllowOrigins).To(Equal(testCors.AllowOrigins))
				Expect(vs.Spec.Http[0].CorsPolicy.AllowMethods).To(Equal(testCors.AllowMethods))
				Expect(vs.Spec.Http[0].CorsPolicy.AllowHeaders).To(Equal(testCors.AllowHeaders))

				Expect(len(vs.Spec.Http[1].Route)).To(Equal(1))
				Expect(vs.Spec.Http[1].Route[0].Destination.Host).To(Equal(oathkeeperSvc))
				Expect(vs.Spec.Http[1].Route[0].Destination.Port.Number).To(Equal(oathkeeperSvcPort))
				Expect(len(vs.Spec.Http[1].Match)).To(Equal(1))
				Expect(vs.Spec.Http[1].Match[0].Uri.GetRegex()).To(Equal(apiRule.Spec.Rules[2].Path))

				Expect(vs.Spec.Http[1].CorsPolicy.AllowOrigins).To(Equal(testCors.AllowOrigins))
				Expect(vs.Spec.Http[1].CorsPolicy.AllowMethods).To(Equal(testCors.AllowMethods))
				Expect(vs.Spec.Http[1].CorsPolicy.AllowHeaders).To(Equal(testCors.AllowHeaders))

				Expect(vs.ObjectMeta.Name).To(BeEmpty())
				Expect(vs.ObjectMeta.GenerateName).To(Equal(apiName + "-"))
				Expect(vs.ObjectMeta.Namespace).To(Equal(apiNamespace))
				Expect(vs.ObjectMeta.Labels[testLabelKey]).To(Equal(testLabelValue))

				Expect(vs.ObjectMeta.OwnerReferences[0].APIVersion).To(Equal(apiAPIVersion))
				Expect(vs.ObjectMeta.OwnerReferences[0].Kind).To(Equal(apiKind))
				Expect(vs.ObjectMeta.OwnerReferences[0].Name).To(Equal(apiName))
				Expect(vs.ObjectMeta.OwnerReferences[0].UID).To(Equal(apiUID))
			})

			It("should produce VS for jwt & oauth authenticators for given path", func() {
				// given
				oauthConfigJSON := fmt.Sprintf(`{"required_scope": [%s]}`, toCSVList(apiScopes))

				jwtConfigJSON := fmt.Sprintf(`
						{
							"trusted_issuers": ["%s"],
							"jwks": [],
							"required_scope": [%s]
					}`, jwtIssuer, toCSVList(apiScopes))

				jwt := &gatewayv1beta1.Authenticator{
					Handler: &gatewayv1beta1.Handler{
						Name: "jwt",
						Config: &runtime.RawExtension{
							Raw: []byte(jwtConfigJSON),
						},
					},
				}
				oauth := &gatewayv1beta1.Authenticator{
					Handler: &gatewayv1beta1.Handler{
						Name: "oauth2_introspection",
						Config: &runtime.RawExtension{
							Raw: []byte(oauthConfigJSON),
						},
					},
				}

				strategies := []*gatewayv1beta1.Authenticator{jwt, oauth}

				allowRule := getRuleFor(apiPath, apiMethods, []*gatewayv1beta1.Mutator{}, strategies)
				rules := []gatewayv1beta1.Rule{allowRule}

				apiRule := getAPIRuleFor(rules)

				vsProcessor := ory.NewVirtualServiceProcessor(getConfigWithEmptyFakeClient())

				// when
				result, err := vsProcessor.EvaluateReconciliation(apiRule)

				// then
				Expect(err).To(BeNil())
				Expect(result).To(HaveLen(1))

				vs := result[0].Obj.(*networkingv1beta1.VirtualService)

				Expect(vs).NotTo(BeNil())
				Expect(len(vs.Spec.Gateways)).To(Equal(1))
				Expect(len(vs.Spec.Hosts)).To(Equal(1))
				Expect(vs.Spec.Hosts[0]).To(Equal(serviceHost))
				Expect(len(vs.Spec.Http)).To(Equal(1))

				Expect(len(vs.Spec.Http[0].Route)).To(Equal(1))
				Expect(vs.Spec.Http[0].Route[0].Destination.Host).To(Equal(oathkeeperSvc))
				Expect(vs.Spec.Http[0].Route[0].Destination.Port.Number).To(Equal(oathkeeperSvcPort))

				Expect(len(vs.Spec.Http[0].Match)).To(Equal(1))
				Expect(vs.Spec.Http[0].Match[0].Uri.GetRegex()).To(Equal(apiRule.Spec.Rules[0].Path))

				Expect(vs.Spec.Http[0].CorsPolicy.AllowOrigins).To(Equal(testCors.AllowOrigins))
				Expect(vs.Spec.Http[0].CorsPolicy.AllowMethods).To(Equal(testCors.AllowMethods))
				Expect(vs.Spec.Http[0].CorsPolicy.AllowHeaders).To(Equal(testCors.AllowHeaders))

				Expect(vs.ObjectMeta.Name).To(BeEmpty())
				Expect(vs.ObjectMeta.GenerateName).To(Equal(apiName + "-"))
				Expect(vs.ObjectMeta.Namespace).To(Equal(apiNamespace))
				Expect(vs.ObjectMeta.Labels[testLabelKey]).To(Equal(testLabelValue))

				Expect(vs.ObjectMeta.OwnerReferences[0].APIVersion).To(Equal(apiAPIVersion))
				Expect(vs.ObjectMeta.OwnerReferences[0].Kind).To(Equal(apiKind))
				Expect(vs.ObjectMeta.OwnerReferences[0].Name).To(Equal(apiName))
				Expect(vs.ObjectMeta.OwnerReferences[0].UID).To(Equal(apiUID))
			})
		})
	})

	Context("update", func() {
		Context("when virtual service has owner v1alpha1 owner label", func() {
			It("should get and update", func() {
				// given
				noop := []*gatewayv1beta1.Authenticator{
					{
						Handler: &gatewayv1beta1.Handler{
							Name: "noop",
						},
					},
				}

				noopRule := getRuleFor(apiPath, apiMethods, []*gatewayv1beta1.Mutator{}, noop)
				rules := []gatewayv1beta1.Rule{noopRule}

				apiRule := getAPIRuleFor(rules)

				rule := rulev1alpha1.Rule{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							processing.OwnerLabelv1alpha1: fmt.Sprintf("%s.%s", apiRule.ObjectMeta.Name, apiRule.ObjectMeta.Namespace),
						},
					},
					Spec: rulev1alpha1.RuleSpec{
						Match: &rulev1alpha1.Match{
							URL: "some url",
						},
					},
				}

				vs := networkingv1beta1.VirtualService{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							processing.OwnerLabelv1alpha1: fmt.Sprintf("%s.%s", apiRule.ObjectMeta.Name, apiRule.ObjectMeta.Namespace),
						},
					},
				}

				scheme := runtime.NewScheme()
				err := rulev1alpha1.AddToScheme(scheme)
				Expect(err).NotTo(HaveOccurred())
				err = networkingv1beta1.AddToScheme(scheme)
				Expect(err).NotTo(HaveOccurred())
				err = gatewayv1beta1.AddToScheme(scheme)
				Expect(err).NotTo(HaveOccurred())

				client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(&rule, &vs).Build()

				vsProcessor := ory.NewVirtualServiceProcessor(getConfigWithClient(client))

				// when
				result, err := vsProcessor.EvaluateReconciliation(apiRule)

				// then
				Expect(err).To(BeNil())
				Expect(result).To(HaveLen(1))
				Expect(result[0].Action).To(Equal("update"))

				resultVs := result[0].Obj.(*networkingv1beta1.VirtualService)

				Expect(resultVs).NotTo(BeNil())
				Expect(resultVs).NotTo(BeNil())
				Expect(len(resultVs.Spec.Gateways)).To(Equal(1))
				Expect(len(resultVs.Spec.Hosts)).To(Equal(1))
				Expect(resultVs.Spec.Hosts[0]).To(Equal(serviceHost))
				Expect(len(resultVs.Spec.Http)).To(Equal(1))

				Expect(len(resultVs.Spec.Http[0].Route)).To(Equal(1))
				Expect(resultVs.Spec.Http[0].Route[0].Destination.Host).To(Equal(oathkeeperSvc))
				Expect(resultVs.Spec.Http[0].Route[0].Destination.Port.Number).To(Equal(oathkeeperSvcPort))

				Expect(len(resultVs.Spec.Http[0].Match)).To(Equal(1))
				Expect(resultVs.Spec.Http[0].Match[0].Uri.GetRegex()).To(Equal(apiRule.Spec.Rules[0].Path))

				Expect(resultVs.Spec.Http[0].CorsPolicy.AllowOrigins).To(Equal(testCors.AllowOrigins))
				Expect(resultVs.Spec.Http[0].CorsPolicy.AllowMethods).To(Equal(testCors.AllowMethods))
				Expect(resultVs.Spec.Http[0].CorsPolicy.AllowHeaders).To(Equal(testCors.AllowHeaders))
			})
		})
	})
})

func getConfigWithEmptyFakeClient() processing.ReconciliationConfig {
	return getConfigWithClient(getEmptyFakeClient())
}

func getConfigWithClient(client client.Client) processing.ReconciliationConfig {
	return processing.ReconciliationConfig{
		Client:            client,
		Ctx:               context.TODO(),
		Logger:            logr.Logger{},
		OathkeeperSvc:     oathkeeperSvc,
		OathkeeperSvcPort: oathkeeperSvcPort,
		CorsConfig:        testCors,
		AdditionalLabels:  testAdditionalLabels,
		DefaultDomainName: defaultDomain,
	}
}

func getEmptyFakeClient() client.Client {
	scheme := runtime.NewScheme()
	err := networkingv1beta1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	return fake.NewClientBuilder().WithScheme(scheme).WithObjects().Build()
}

func getRuleFor(path string, methods []string, mutators []*gatewayv1beta1.Mutator, accessStrategies []*gatewayv1beta1.Authenticator) gatewayv1beta1.Rule {
	return gatewayv1beta1.Rule{
		Path:             path,
		Methods:          methods,
		Mutators:         mutators,
		AccessStrategies: accessStrategies,
	}
}

func getRuleWithServiceFor(path string, methods []string, mutators []*gatewayv1beta1.Mutator, accessStrategies []*gatewayv1beta1.Authenticator, service *gatewayv1beta1.Service) gatewayv1beta1.Rule {
	return gatewayv1beta1.Rule{
		Path:             path,
		Methods:          methods,
		Mutators:         mutators,
		AccessStrategies: accessStrategies,
		Service:          service,
	}
}

func getAPIRuleFor(rules []gatewayv1beta1.Rule) *gatewayv1beta1.APIRule {
	return &gatewayv1beta1.APIRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      apiName,
			UID:       apiUID,
			Namespace: apiNamespace,
		},
		TypeMeta: metav1.TypeMeta{
			APIVersion: apiAPIVersion,
			Kind:       apiKind,
		},
		Spec: gatewayv1beta1.APIRuleSpec{
			Gateway: &apiGateway,
			Service: &gatewayv1beta1.Service{
				Name: &serviceName,
				Port: &servicePort,
			},
			Host:  &serviceHost,
			Rules: rules,
		},
	}
}

func toCSVList(input []string) string {
	if len(input) == 0 {
		return ""
	}

	res := `"` + input[0] + `"`

	for i := 1; i < len(input); i++ {
		res = res + "," + `"` + input[i] + `"`
	}

	return res
}
