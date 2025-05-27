package istio_test

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	rulev1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
	v1beta12 "istio.io/api/networking/v1beta1"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"github.com/kyma-project/api-gateway/internal/builders"
	"github.com/kyma-project/api-gateway/internal/processing"
	. "github.com/kyma-project/api-gateway/internal/processing/processing_test"
	"github.com/kyma-project/api-gateway/internal/processing/processors/istio"
)

var _ = Describe("Virtual Service Processor", func() {

	allowHandlers := []string{v1beta1.AccessStrategyAllow, v1beta1.AccessStrategyNoAuth}
	for _, handler := range allowHandlers {
		When("handler is "+handler, func() {
			It("should create", func() {
				// given
				strategies := []*v1beta1.Authenticator{
					{
						Handler: &v1beta1.Handler{
							Name: handler,
						},
					},
				}

				rule := GetRuleFor(ApiPath, ApiMethods, []*v1beta1.Mutator{}, strategies)
				rules := []v1beta1.Rule{rule}

				apiRule := GetAPIRuleFor(rules)
				client := GetFakeClient()
				processor := istio.Newv1beta1VirtualServiceProcessor(GetTestConfig(), apiRule)

				// when
				result, err := processor.EvaluateReconciliation(context.Background(), client)

				// then
				Expect(err).To(BeNil())
				Expect(result).To(HaveLen(1))
				Expect(result[0].Action.String()).To(Equal("create"))

				vs := result[0].Obj.(*networkingv1beta1.VirtualService)

				Expect(vs).NotTo(BeNil())
				Expect(len(vs.Spec.Gateways)).To(Equal(1))
				Expect(len(vs.Spec.Hosts)).To(Equal(1))
				Expect(vs.Spec.Hosts[0]).To(Equal(ServiceHost))
				Expect(len(vs.Spec.Http)).To(Equal(1))

				Expect(len(vs.Spec.Http[0].Route)).To(Equal(1))
				Expect(vs.Spec.Http[0].Route[0].Destination.Host).To(Equal(ServiceName + "." + ApiNamespace + ".svc.cluster.local"))
				Expect(vs.Spec.Http[0].Route[0].Destination.Port.Number).To(Equal(ServicePort))

				Expect(len(vs.Spec.Http[0].Match)).To(Equal(1))
				Expect(vs.Spec.Http[0].Match[0].Uri.GetRegex()).To(Equal(apiRule.Spec.Rules[0].Path))

				Expect(vs.Spec.Http[0].CorsPolicy.AllowOrigins).To(Equal(TestCors.AllowOrigins))
				Expect(vs.Spec.Http[0].CorsPolicy.AllowMethods).To(Equal(TestCors.AllowMethods))
				Expect(vs.Spec.Http[0].CorsPolicy.AllowHeaders).To(Equal(TestCors.AllowHeaders))

				Expect(vs.ObjectMeta.Name).To(BeEmpty())
				Expect(vs.ObjectMeta.GenerateName).To(Equal(ApiName + "-"))
				Expect(vs.ObjectMeta.Namespace).To(Equal(ApiNamespace))
			})

			It("should override destination host for specified spec level service namespace", func() {
				// given
				strategies := []*v1beta1.Authenticator{
					{
						Handler: &v1beta1.Handler{
							Name: handler,
						},
					},
				}

				rule := GetRuleFor(ApiPath, ApiMethods, []*v1beta1.Mutator{}, strategies)
				rules := []v1beta1.Rule{rule}

				apiRule := GetAPIRuleFor(rules)

				overrideServiceName := "testName"
				overrideServiceNamespace := "testName-namespace"
				overrideServicePort := uint32(8080)

				apiRule.Spec.Service = &v1beta1.Service{
					Name:      &overrideServiceName,
					Namespace: &overrideServiceNamespace,
					Port:      &overrideServicePort,
				}
				client := GetFakeClient()
				processor := istio.Newv1beta1VirtualServiceProcessor(GetTestConfig(), apiRule)

				// when
				result, err := processor.EvaluateReconciliation(context.Background(), client)

				// then
				Expect(err).To(BeNil())
				Expect(result).To(HaveLen(1))

				vs := result[0].Obj.(*networkingv1beta1.VirtualService)

				Expect(len(vs.Spec.Http[0].Route)).To(Equal(1))
				Expect(vs.Spec.Http[0].Route[0].Destination.Host).To(Equal(overrideServiceName + "." + overrideServiceNamespace + ".svc.cluster.local"))
			})

			It("should override destination host with rule level service namespace", func() {
				// given
				strategies := []*v1beta1.Authenticator{
					{
						Handler: &v1beta1.Handler{
							Name: handler,
						},
					},
				}

				overrideServiceName := "testName"
				overrideServiceNamespace := "testName-namespace"
				overrideServicePort := uint32(8080)

				service := &v1beta1.Service{
					Name:      &overrideServiceName,
					Namespace: &overrideServiceNamespace,
					Port:      &overrideServicePort,
				}

				rule := GetRuleWithServiceFor(ApiPath, ApiMethods, []*v1beta1.Mutator{}, strategies, service)
				rules := []v1beta1.Rule{rule}

				apiRule := GetAPIRuleFor(rules)
				client := GetFakeClient()
				processor := istio.Newv1beta1VirtualServiceProcessor(GetTestConfig(), apiRule)

				// when
				result, err := processor.EvaluateReconciliation(context.Background(), client)

				// then
				Expect(err).To(BeNil())
				Expect(result).To(HaveLen(1))

				vs := result[0].Obj.(*networkingv1beta1.VirtualService)

				//verify VS has rule level destination host
				Expect(len(vs.Spec.Http[0].Route)).To(Equal(1))
				Expect(vs.Spec.Http[0].Route[0].Destination.Host).To(Equal(overrideServiceName + "." + overrideServiceNamespace + ".svc.cluster.local"))

			})

			It("should return VS with default domain name when the hostname does not contain domain name", func() {
				strategies := []*v1beta1.Authenticator{
					{
						Handler: &v1beta1.Handler{
							Name: handler,
						},
					},
				}

				rule := GetRuleFor(ApiPath, ApiMethods, []*v1beta1.Mutator{}, strategies)
				rules := []v1beta1.Rule{rule}

				apiRule := GetAPIRuleFor(rules)
				apiRule.Spec.Host = &ServiceHostWithNoDomain
				client := GetFakeClient()
				processor := istio.Newv1beta1VirtualServiceProcessor(GetTestConfig(), apiRule)

				// when
				result, err := processor.EvaluateReconciliation(context.Background(), client)

				// then
				Expect(err).To(BeNil())
				Expect(result).To(HaveLen(1))

				vs := result[0].Obj.(*networkingv1beta1.VirtualService)

				//verify VS
				Expect(vs).NotTo(BeNil())
				Expect(len(vs.Spec.Hosts)).To(Equal(1))
				Expect(vs.Spec.Hosts[0]).To(Equal(ServiceHost))

			})
		})

	}

	When("handler is noop", func() {
		It("should not override Oathkeeper service destination host with spec level service", func() {
			// given
			strategies := []*v1beta1.Authenticator{
				{
					Handler: &v1beta1.Handler{
						Name: "noop",
					},
				},
			}

			overrideServiceName := "testName"
			overrideServicePort := uint32(8080)

			service := &v1beta1.Service{
				Name: &overrideServiceName,
				Port: &overrideServicePort,
			}

			allowRule := GetRuleWithServiceFor(ApiPath, ApiMethods, []*v1beta1.Mutator{}, strategies, service)
			rules := []v1beta1.Rule{allowRule}

			apiRule := GetAPIRuleFor(rules)
			client := GetFakeClient()
			processor := istio.Newv1beta1VirtualServiceProcessor(GetTestConfig(), apiRule)

			// when
			result, err := processor.EvaluateReconciliation(context.Background(), client)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(1))

			vs := result[0].Obj.(*networkingv1beta1.VirtualService)

			Expect(len(vs.Spec.Http[0].Route)).To(Equal(1))
			Expect(vs.Spec.Http[0].Route[0].Destination.Host).To(Equal(OathkeeperSvc))
		})

		When("existing virtual service has owner v1alpha1 owner label", func() {
			It("should get and update", func() {
				// given
				noop := []*v1beta1.Authenticator{
					{
						Handler: &v1beta1.Handler{
							Name: "noop",
						},
					},
				}

				noopRule := GetRuleFor(ApiPath, ApiMethods, []*v1beta1.Mutator{}, noop)
				rules := []v1beta1.Rule{noopRule}

				apiRule := GetAPIRuleFor(rules)

				rule := rulev1alpha1.Rule{
					ObjectMeta: metav1.ObjectMeta{},
					Spec: rulev1alpha1.RuleSpec{
						Match: &rulev1alpha1.Match{
							URL: "some url",
						},
					},
				}

				vs := networkingv1beta1.VirtualService{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							processing.OwnerLabel: fmt.Sprintf("%s.%s", apiRule.Name, apiRule.Namespace),
						},
					},
				}

				scheme := runtime.NewScheme()
				err := rulev1alpha1.AddToScheme(scheme)
				Expect(err).NotTo(HaveOccurred())
				err = networkingv1beta1.AddToScheme(scheme)
				Expect(err).NotTo(HaveOccurred())
				err = v1beta1.AddToScheme(scheme)
				Expect(err).NotTo(HaveOccurred())

				client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(&rule, &vs).Build()
				processor := istio.Newv1beta1VirtualServiceProcessor(GetTestConfig(), apiRule)

				// when
				result, err := processor.EvaluateReconciliation(context.Background(), client)

				// then
				Expect(err).To(BeNil())
				Expect(result).To(HaveLen(1))
				Expect(result[0].Action.String()).To(Equal("update"))

				resultVs := result[0].Obj.(*networkingv1beta1.VirtualService)

				Expect(resultVs).NotTo(BeNil())
				Expect(resultVs).NotTo(BeNil())
				Expect(len(resultVs.Spec.Gateways)).To(Equal(1))
				Expect(len(resultVs.Spec.Hosts)).To(Equal(1))
				Expect(resultVs.Spec.Hosts[0]).To(Equal(ServiceHost))
				Expect(len(resultVs.Spec.Http)).To(Equal(1))

				Expect(len(resultVs.Spec.Http[0].Route)).To(Equal(1))
				Expect(resultVs.Spec.Http[0].Route[0].Destination.Host).To(Equal(OathkeeperSvc))
				Expect(resultVs.Spec.Http[0].Route[0].Destination.Port.Number).To(Equal(OathkeeperSvcPort))

				Expect(len(resultVs.Spec.Http[0].Match)).To(Equal(1))
				Expect(resultVs.Spec.Http[0].Match[0].Uri.GetRegex()).To(Equal(apiRule.Spec.Rules[0].Path))

				Expect(resultVs.Spec.Http[0].CorsPolicy.AllowOrigins).To(Equal(TestCors.AllowOrigins))
				Expect(resultVs.Spec.Http[0].CorsPolicy.AllowMethods).To(Equal(TestCors.AllowMethods))
				Expect(resultVs.Spec.Http[0].CorsPolicy.AllowHeaders).To(Equal(TestCors.AllowHeaders))
			})
		})
	})

	When("multiple handler", func() {
		It("should return service for given paths", func() {
			// given
			noop := []*v1beta1.Authenticator{
				{
					Handler: &v1beta1.Handler{
						Name: "noop",
					},
				},
			}

			jwtConfigJSON := fmt.Sprintf(`
						{
							"trusted_issuers": ["%s"],
							"jwks": [],
							"required_scope": [%s]
					}`, JwtIssuer, ToCSVList(ApiScopes))

			jwt := []*v1beta1.Authenticator{
				{
					Handler: &v1beta1.Handler{
						Name: "jwt",
						Config: &runtime.RawExtension{
							Raw: []byte(jwtConfigJSON),
						},
					},
				},
			}

			testMutators := []*v1beta1.Mutator{
				{
					Handler: &v1beta1.Handler{
						Name: "noop",
					},
				},
				{
					Handler: &v1beta1.Handler{
						Name: "idtoken",
					},
				},
			}

			noopRule := GetRuleFor(ApiPath, ApiMethods, []*v1beta1.Mutator{}, noop)
			jwtRule := GetRuleFor(HeadersApiPath, ApiMethods, testMutators, jwt)
			rules := []v1beta1.Rule{noopRule, jwtRule}

			apiRule := GetAPIRuleFor(rules)
			client := GetFakeClient()
			processor := istio.Newv1beta1VirtualServiceProcessor(GetTestConfig(), apiRule)

			// when
			result, err := processor.EvaluateReconciliation(context.Background(), client)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(1))

			vs := result[0].Obj.(*networkingv1beta1.VirtualService)

			Expect(vs).NotTo(BeNil())
			Expect(len(vs.Spec.Gateways)).To(Equal(1))
			Expect(len(vs.Spec.Hosts)).To(Equal(1))
			Expect(vs.Spec.Hosts[0]).To(Equal(ServiceHost))
			Expect(len(vs.Spec.Http)).To(Equal(2))

			Expect(len(vs.Spec.Http[0].Route)).To(Equal(1))
			Expect(vs.Spec.Http[0].Route[0].Destination.Host).To(Equal(OathkeeperSvc))
			Expect(vs.Spec.Http[0].Route[0].Destination.Port.Number).To(Equal(OathkeeperSvcPort))
			Expect(len(vs.Spec.Http[0].Match)).To(Equal(1))
			Expect(vs.Spec.Http[0].Match[0].Uri.GetRegex()).To(Equal(apiRule.Spec.Rules[0].Path))

			Expect(vs.Spec.Http[0].CorsPolicy.AllowOrigins).To(Equal(TestCors.AllowOrigins))
			Expect(vs.Spec.Http[0].CorsPolicy.AllowMethods).To(Equal(TestCors.AllowMethods))
			Expect(vs.Spec.Http[0].CorsPolicy.AllowHeaders).To(Equal(TestCors.AllowHeaders))

			Expect(len(vs.Spec.Http[1].Route)).To(Equal(1))
			Expect(vs.Spec.Http[1].Route[0].Destination.Host).To(Equal(ServiceName + "." + ApiNamespace + ".svc.cluster.local"))
			Expect(vs.Spec.Http[1].Route[0].Destination.Port.Number).To(Equal(ServicePort))
			Expect(len(vs.Spec.Http[1].Match)).To(Equal(1))
			Expect(vs.Spec.Http[1].Match[0].Uri.GetRegex()).To(Equal(apiRule.Spec.Rules[1].Path))

			Expect(vs.Spec.Http[1].CorsPolicy.AllowOrigins).To(Equal(TestCors.AllowOrigins))
			Expect(vs.Spec.Http[1].CorsPolicy.AllowMethods).To(Equal(TestCors.AllowMethods))
			Expect(vs.Spec.Http[1].CorsPolicy.AllowHeaders).To(Equal(TestCors.AllowHeaders))

			Expect(vs.ObjectMeta.Name).To(BeEmpty())
			Expect(vs.ObjectMeta.GenerateName).To(Equal(ApiName + "-"))
			Expect(vs.ObjectMeta.Namespace).To(Equal(ApiNamespace))
		})

		It("should return service for two same paths and different methods", func() {
			// given
			noop := []*v1beta1.Authenticator{
				{
					Handler: &v1beta1.Handler{
						Name: "noop",
					},
				},
			}

			jwtConfigJSON := fmt.Sprintf(`
						{
							"trusted_issuers": ["%s"],
							"jwks": [],
							"required_scope": [%s]
					}`, JwtIssuer, ToCSVList(ApiScopes))

			jwt := []*v1beta1.Authenticator{
				{
					Handler: &v1beta1.Handler{
						Name: "jwt",
						Config: &runtime.RawExtension{
							Raw: []byte(jwtConfigJSON),
						},
					},
				},
			}

			testMutators := []*v1beta1.Mutator{
				{
					Handler: &v1beta1.Handler{
						Name: "noop",
					},
				},
				{
					Handler: &v1beta1.Handler{
						Name: "idtoken",
					},
				},
			}
			noopRule := GetRuleFor(ApiPath, methodsGet, []*v1beta1.Mutator{}, noop)
			jwtRule := GetRuleFor(ApiPath, methodsPost, testMutators, jwt)
			rules := []v1beta1.Rule{noopRule, jwtRule}

			apiRule := GetAPIRuleFor(rules)
			client := GetFakeClient()
			processor := istio.Newv1beta1VirtualServiceProcessor(GetTestConfig(), apiRule)

			// when
			result, err := processor.EvaluateReconciliation(context.Background(), client)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(1))

			vs := result[0].Obj.(*networkingv1beta1.VirtualService)

			Expect(vs).NotTo(BeNil())
			Expect(len(vs.Spec.Gateways)).To(Equal(1))
			Expect(len(vs.Spec.Hosts)).To(Equal(1))
			Expect(vs.Spec.Hosts[0]).To(Equal(ServiceHost))
			Expect(len(vs.Spec.Http)).To(Equal(1))

			Expect(len(vs.Spec.Http[0].Route)).To(Equal(1))
			Expect(vs.Spec.Http[0].Route[0].Destination.Host).To(Equal(OathkeeperSvc))
			Expect(vs.Spec.Http[0].Route[0].Destination.Port.Number).To(Equal(OathkeeperSvcPort))
			Expect(len(vs.Spec.Http[0].Match)).To(Equal(1))
			Expect(vs.Spec.Http[0].Match[0].Uri.GetRegex()).To(Equal(apiRule.Spec.Rules[0].Path))

			Expect(vs.Spec.Http[0].CorsPolicy.AllowOrigins).To(Equal(TestCors.AllowOrigins))
			Expect(vs.Spec.Http[0].CorsPolicy.AllowMethods).To(Equal(TestCors.AllowMethods))
			Expect(vs.Spec.Http[0].CorsPolicy.AllowHeaders).To(Equal(TestCors.AllowHeaders))

			Expect(vs.ObjectMeta.Name).To(BeEmpty())
			Expect(vs.ObjectMeta.GenerateName).To(Equal(ApiName + "-"))
			Expect(vs.ObjectMeta.Namespace).To(Equal(ApiNamespace))
		})

		It("should return service for two same paths and one different", func() {
			// given
			noop := []*v1beta1.Authenticator{
				{
					Handler: &v1beta1.Handler{
						Name: "noop",
					},
				},
			}

			jwtConfigJSON := fmt.Sprintf(`
						{
							"trusted_issuers": ["%s"],
							"jwks": [],
							"required_scope": [%s]
					}`, JwtIssuer, ToCSVList(ApiScopes))

			jwt := []*v1beta1.Authenticator{
				{
					Handler: &v1beta1.Handler{
						Name: "jwt",
						Config: &runtime.RawExtension{
							Raw: []byte(jwtConfigJSON),
						},
					},
				},
			}

			testMutators := []*v1beta1.Mutator{
				{
					Handler: &v1beta1.Handler{
						Name: "noop",
					},
				},
				{
					Handler: &v1beta1.Handler{
						Name: "idtoken",
					},
				},
			}

			noopGetRule := GetRuleFor(ApiPath, methodsGet, []*v1beta1.Mutator{}, noop)
			noopPostRule := GetRuleFor(ApiPath, methodsPost, []*v1beta1.Mutator{}, noop)
			jwtRule := GetRuleFor(HeadersApiPath, ApiMethods, testMutators, jwt)
			rules := []v1beta1.Rule{noopGetRule, noopPostRule, jwtRule}

			apiRule := GetAPIRuleFor(rules)
			client := GetFakeClient()
			processor := istio.Newv1beta1VirtualServiceProcessor(GetTestConfig(), apiRule)

			// when
			result, err := processor.EvaluateReconciliation(context.Background(), client)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(1))

			vs := result[0].Obj.(*networkingv1beta1.VirtualService)

			Expect(vs).NotTo(BeNil())
			Expect(len(vs.Spec.Gateways)).To(Equal(1))
			Expect(len(vs.Spec.Hosts)).To(Equal(1))
			Expect(vs.Spec.Hosts[0]).To(Equal(ServiceHost))
			Expect(len(vs.Spec.Http)).To(Equal(2))

			Expect(len(vs.Spec.Http[0].Route)).To(Equal(1))
			Expect(vs.Spec.Http[0].Route[0].Destination.Host).To(Equal(OathkeeperSvc))
			Expect(vs.Spec.Http[0].Route[0].Destination.Port.Number).To(Equal(OathkeeperSvcPort))
			Expect(len(vs.Spec.Http[0].Match)).To(Equal(1))
			Expect(vs.Spec.Http[0].Match[0].Uri.GetRegex()).To(Equal(apiRule.Spec.Rules[0].Path))

			Expect(vs.Spec.Http[0].CorsPolicy.AllowOrigins).To(Equal(TestCors.AllowOrigins))
			Expect(vs.Spec.Http[0].CorsPolicy.AllowMethods).To(Equal(TestCors.AllowMethods))
			Expect(vs.Spec.Http[0].CorsPolicy.AllowHeaders).To(Equal(TestCors.AllowHeaders))

			Expect(len(vs.Spec.Http[1].Route)).To(Equal(1))
			Expect(vs.Spec.Http[1].Route[0].Destination.Host).To(Equal(ServiceName + "." + ApiNamespace + ".svc.cluster.local"))
			Expect(vs.Spec.Http[1].Route[0].Destination.Port.Number).To(Equal(ServicePort))
			Expect(len(vs.Spec.Http[1].Match)).To(Equal(1))
			Expect(vs.Spec.Http[1].Match[0].Uri.GetRegex()).To(Equal(apiRule.Spec.Rules[2].Path))

			Expect(vs.Spec.Http[1].CorsPolicy.AllowOrigins).To(Equal(TestCors.AllowOrigins))
			Expect(vs.Spec.Http[1].CorsPolicy.AllowMethods).To(Equal(TestCors.AllowMethods))
			Expect(vs.Spec.Http[1].CorsPolicy.AllowHeaders).To(Equal(TestCors.AllowHeaders))

			Expect(vs.ObjectMeta.Name).To(BeEmpty())
			Expect(vs.ObjectMeta.GenerateName).To(Equal(ApiName + "-"))
			Expect(vs.ObjectMeta.Namespace).To(Equal(ApiNamespace))
		})

		It("should return service for jwt & oauth authenticators for given path", func() {
			// given
			oauthConfigJSON := fmt.Sprintf(`{"required_scope": [%s]}`, ToCSVList(ApiScopes))

			jwtConfigJSON := fmt.Sprintf(`
						{
							"trusted_issuers": ["%s"],
							"jwks": [],
							"required_scope": [%s]
					}`, JwtIssuer, ToCSVList(ApiScopes))

			jwt := &v1beta1.Authenticator{
				Handler: &v1beta1.Handler{
					Name: "jwt",
					Config: &runtime.RawExtension{
						Raw: []byte(jwtConfigJSON),
					},
				},
			}
			oauth := &v1beta1.Authenticator{
				Handler: &v1beta1.Handler{
					Name: "oauth2_introspection",
					Config: &runtime.RawExtension{
						Raw: []byte(oauthConfigJSON),
					},
				},
			}

			strategies := []*v1beta1.Authenticator{jwt, oauth}

			allowRule := GetRuleFor(ApiPath, ApiMethods, []*v1beta1.Mutator{}, strategies)
			rules := []v1beta1.Rule{allowRule}

			apiRule := GetAPIRuleFor(rules)
			client := GetFakeClient()
			processor := istio.Newv1beta1VirtualServiceProcessor(GetTestConfig(), apiRule)

			// when
			result, err := processor.EvaluateReconciliation(context.Background(), client)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(1))

			vs := result[0].Obj.(*networkingv1beta1.VirtualService)

			Expect(vs).NotTo(BeNil())
			Expect(len(vs.Spec.Gateways)).To(Equal(1))
			Expect(len(vs.Spec.Hosts)).To(Equal(1))
			Expect(vs.Spec.Hosts[0]).To(Equal(ServiceHost))
			Expect(len(vs.Spec.Http)).To(Equal(1))

			Expect(len(vs.Spec.Http[0].Route)).To(Equal(1))
			Expect(vs.Spec.Http[0].Route[0].Destination.Host).To(Equal(ServiceName + "." + ApiNamespace + ".svc.cluster.local"))
			Expect(vs.Spec.Http[0].Route[0].Destination.Port.Number).To(Equal(ServicePort))

			Expect(len(vs.Spec.Http[0].Match)).To(Equal(1))
			Expect(vs.Spec.Http[0].Match[0].Uri.GetRegex()).To(Equal(apiRule.Spec.Rules[0].Path))

			Expect(vs.Spec.Http[0].CorsPolicy.AllowOrigins).To(Equal(TestCors.AllowOrigins))
			Expect(vs.Spec.Http[0].CorsPolicy.AllowMethods).To(Equal(TestCors.AllowMethods))
			Expect(vs.Spec.Http[0].CorsPolicy.AllowHeaders).To(Equal(TestCors.AllowHeaders))

			Expect(vs.ObjectMeta.Name).To(BeEmpty())
			Expect(vs.ObjectMeta.GenerateName).To(Equal(ApiName + "-"))
			Expect(vs.ObjectMeta.Namespace).To(Equal(ApiNamespace))
		})
	})

	When("the path is `/*`", func() {
		It("should set the match to prefix `/`", func() {
			strategies := []*v1beta1.Authenticator{
				{
					Handler: &v1beta1.Handler{
						Name: v1beta1.AccessStrategyNoAuth,
					},
				},
			}

			allowRule := GetRuleFor("/*", ApiMethods, []*v1beta1.Mutator{}, strategies)
			rules := []v1beta1.Rule{allowRule}

			apiRule := GetAPIRuleFor(rules)
			client := GetFakeClient()
			processor := istio.Newv1beta1VirtualServiceProcessor(GetTestConfig(), apiRule)

			// when
			result, err := processor.EvaluateReconciliation(context.Background(), client)

			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(HaveLen(1))

			resultVs := result[0].Obj.(*networkingv1beta1.VirtualService)

			Expect(resultVs.Spec.Http).To(HaveLen(1))
			Expect(resultVs.Spec.Http[0].Match).To(HaveLen(1))
			Expect(resultVs.Spec.Http[0].Match[0].Uri.GetPrefix()).To(Equal("/"))
		})
	})
	Context("mutators are defined", func() {
		When("access strategy is JWT", func() {
			It("should return VS cookie and header configuration set", func() {

				jwtConfigJSON := fmt.Sprintf(`{"trusted_issuers": ["%s"],"jwks": [],}`, JwtIssuer)

				strategies := []*v1beta1.Authenticator{
					{
						Handler: &v1beta1.Handler{
							Name: "jwt",
							Config: &runtime.RawExtension{
								Raw: []byte(jwtConfigJSON),
							},
						},
					},
				}

				mutators := []*v1beta1.Mutator{
					{
						Handler: &v1beta1.Handler{
							Name: "cookie",
							Config: GetRawConfig(
								v1beta1.CookieMutatorConfig{
									Cookies: map[string]string{
										"x-test-cookie-1": "cookie-value1",
										"x-test-cookie-2": "cookie-value2",
									},
								},
							),
						},
					},
					{
						Handler: &v1beta1.Handler{
							Name: "header",
							Config: GetRawConfig(
								v1beta1.HeaderMutatorConfig{
									Headers: map[string]string{
										"x-test-header-1": "header-value1",
										"x-test-header-2": "header-value2",
									},
								},
							),
						},
					},
				}

				allowRule := GetRuleFor(ApiPath, ApiMethods, mutators, strategies)
				rules := []v1beta1.Rule{allowRule}

				apiRule := GetAPIRuleFor(rules)
				apiRule.Spec.Host = &ServiceHostWithNoDomain
				client := GetFakeClient()
				processor := istio.Newv1beta1VirtualServiceProcessor(GetTestConfig(), apiRule)

				// when
				result, err := processor.EvaluateReconciliation(context.Background(), client)

				// then
				Expect(err).To(BeNil())
				Expect(result).To(HaveLen(1))

				vs := result[0].Obj.(*networkingv1beta1.VirtualService)

				//verify VS
				Expect(vs).NotTo(BeNil())
				Expect(vs.Spec.Http).To(HaveLen(1))
				Expect(vs.Spec.Http[0].Headers.Request.Set).To(HaveKey("Cookie"))
				Expect(vs.Spec.Http[0].Headers.Request.Set["Cookie"]).To(ContainSubstring("x-test-cookie-1=cookie-value1"))
				Expect(vs.Spec.Http[0].Headers.Request.Set["Cookie"]).To(ContainSubstring("x-test-cookie-2=cookie-value2"))
				Expect(vs.Spec.Http[0].Headers.Request.Set).To(HaveKeyWithValue("x-test-header-1", "header-value1"))
				Expect(vs.Spec.Http[0].Headers.Request.Set).To(HaveKeyWithValue("x-test-header-2", "header-value2"))
			})

			It("should not override x-forwarded-for header", func() {
				jwtConfigJSON := fmt.Sprintf(`{"trusted_issuers": ["%s"],"jwks": [],}`, JwtIssuer)

				strategies := []*v1beta1.Authenticator{
					{
						Handler: &v1beta1.Handler{
							Name: "jwt",
							Config: &runtime.RawExtension{
								Raw: []byte(jwtConfigJSON),
							},
						},
					},
				}

				mutators := []*v1beta1.Mutator{
					{
						Handler: &v1beta1.Handler{
							Name: "header",
							Config: GetRawConfig(
								v1beta1.HeaderMutatorConfig{
									Headers: map[string]string{
										"x-test-header-1": "header-value1",
									},
								},
							),
						},
					},
				}

				allowRule := GetRuleFor(ApiPath, ApiMethods, mutators, strategies)
				rules := []v1beta1.Rule{allowRule}

				apiRule := GetAPIRuleFor(rules)
				apiRule.Spec.Host = &ServiceHostWithNoDomain
				client := GetFakeClient()
				processor := istio.Newv1beta1VirtualServiceProcessor(GetTestConfig(), apiRule)

				// when
				result, err := processor.EvaluateReconciliation(context.Background(), client)

				// then
				Expect(err).To(BeNil())
				Expect(result).To(HaveLen(1))

				vs := result[0].Obj.(*networkingv1beta1.VirtualService)

				//verify VS
				Expect(vs).NotTo(BeNil())
				Expect(vs.Spec.Http).To(HaveLen(1))
				Expect(vs.Spec.Http[0].Headers.Request.Set).To(HaveKeyWithValue("x-forwarded-host", "myService.myDomain.com"))
				Expect(vs.Spec.Http[0].Headers.Request.Set).To(HaveKeyWithValue("x-test-header-1", "header-value1"))
			})

		})

		DescribeTable("should not add mutator config to VS when access strategy is ", func(handler string) {
			strategies := []*v1beta1.Authenticator{
				{
					Handler: &v1beta1.Handler{
						Name: handler,
					},
				},
			}

			mutators := []*v1beta1.Mutator{
				{
					Handler: &v1beta1.Handler{
						Name: "header",
						Config: GetRawConfig(
							v1beta1.HeaderMutatorConfig{
								Headers: map[string]string{
									"x-test-header-1": "header-value1",
								},
							},
						),
					},
				},
			}

			allowRule := GetRuleFor(ApiPath, ApiMethods, mutators, strategies)
			rules := []v1beta1.Rule{allowRule}

			apiRule := GetAPIRuleFor(rules)
			apiRule.Spec.Host = &ServiceHostWithNoDomain
			client := GetFakeClient()
			processor := istio.Newv1beta1VirtualServiceProcessor(GetTestConfig(), apiRule)

			// when
			result, err := processor.EvaluateReconciliation(context.Background(), client)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(1))

			vs := result[0].Obj.(*networkingv1beta1.VirtualService)

			//verify VS
			Expect(vs).NotTo(BeNil())
			Expect(vs.Spec.Http).To(HaveLen(1))
			Expect(vs.Spec.Http[0].Headers.Request.Set).ToNot(HaveKey("Cookie"))
			Expect(vs.Spec.Http[0].Headers.Request.Set).ToNot(HaveKeyWithValue("x-test-header-1", "header-value1"))
		},
			Entry(nil, v1beta1.AccessStrategyNoAuth),
			Entry(nil, v1beta1.AccessStrategyAllow),
		)

		It("should not add mutator config to VS when access strategy is noop", func() {

			strategies := []*v1beta1.Authenticator{
				{
					Handler: &v1beta1.Handler{
						Name: "noop",
					},
				},
			}

			mutators := []*v1beta1.Mutator{
				{
					Handler: &v1beta1.Handler{
						Name: "header",
						Config: GetRawConfig(
							v1beta1.HeaderMutatorConfig{
								Headers: map[string]string{
									"x-test-header-1": "header-value1",
								},
							},
						),
					},
				},
			}

			allowRule := GetRuleFor(ApiPath, ApiMethods, mutators, strategies)
			rules := []v1beta1.Rule{allowRule}

			apiRule := GetAPIRuleFor(rules)
			apiRule.Spec.Host = &ServiceHostWithNoDomain
			client := GetFakeClient()
			processor := istio.Newv1beta1VirtualServiceProcessor(GetTestConfig(), apiRule)

			// when
			result, err := processor.EvaluateReconciliation(context.Background(), client)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(1))

			vs := result[0].Obj.(*networkingv1beta1.VirtualService)

			//verify VS
			Expect(vs).NotTo(BeNil())
			Expect(vs.Spec.Http).To(HaveLen(1))
			Expect(vs.Spec.Http[0].Headers.Request.Set).ToNot(HaveKey("Cookie"))
			Expect(vs.Spec.Http[0].Headers.Request.Set).ToNot(HaveKeyWithValue("x-test-header-1", "header-value1"))
		})

		It("should not add mutator config to VS when access strategy is oauth2_introspection", func() {

			strategies := []*v1beta1.Authenticator{
				{
					Handler: &v1beta1.Handler{
						Name: "oauth2_introspection",
					},
				},
			}

			mutators := []*v1beta1.Mutator{
				{
					Handler: &v1beta1.Handler{
						Name: "header",
						Config: GetRawConfig(
							v1beta1.HeaderMutatorConfig{
								Headers: map[string]string{
									"x-test-header-1": "header-value1",
								},
							},
						),
					},
				},
			}

			allowRule := GetRuleFor(ApiPath, ApiMethods, mutators, strategies)
			rules := []v1beta1.Rule{allowRule}

			apiRule := GetAPIRuleFor(rules)
			apiRule.Spec.Host = &ServiceHostWithNoDomain
			client := GetFakeClient()
			processor := istio.Newv1beta1VirtualServiceProcessor(GetTestConfig(), apiRule)

			// when
			result, err := processor.EvaluateReconciliation(context.Background(), client)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(1))

			vs := result[0].Obj.(*networkingv1beta1.VirtualService)

			//verify VS
			Expect(vs).NotTo(BeNil())
			Expect(vs.Spec.Http).To(HaveLen(1))
			Expect(vs.Spec.Http[0].Headers.Request.Set).ToNot(HaveKey("Cookie"))
			Expect(vs.Spec.Http[0].Headers.Request.Set).ToNot(HaveKeyWithValue("x-test-header-1", "header-value1"))
		})
	})

	Context("CORS", func() {
		It("should set default values in CORSPolicy when it is not configured in APIRule", func() {
			// given
			strategies := []*v1beta1.Authenticator{
				{
					Handler: &v1beta1.Handler{
						Name: v1beta1.AccessStrategyNoAuth,
					},
				},
			}

			allowRule := GetRuleFor(ApiPath, ApiMethods, nil, strategies)
			rules := []v1beta1.Rule{allowRule}

			apiRule := GetAPIRuleFor(rules)
			apiRule.Spec.Host = &ServiceHostWithNoDomain
			client := GetFakeClient()
			processor := istio.Newv1beta1VirtualServiceProcessor(GetTestConfig(), apiRule)

			// when
			result, err := processor.EvaluateReconciliation(context.Background(), client)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(1))

			vs := result[0].Obj.(*networkingv1beta1.VirtualService)

			//verify VS
			Expect(vs).NotTo(BeNil())
			Expect(vs.Spec.Http[0].CorsPolicy).NotTo(BeNil())
			Expect(vs.Spec.Http[0].CorsPolicy.AllowMethods).To(ContainElements(TestCors.AllowMethods))
			Expect(vs.Spec.Http[0].CorsPolicy.AllowOrigins).To(ContainElements(TestCors.AllowOrigins))
			Expect(vs.Spec.Http[0].CorsPolicy.AllowHeaders).To(ContainElements(TestCors.AllowHeaders))
		})

		It("should not set default values in CORSPolicy when it is configured in APIRule, and set headers", func() {
			// given
			strategies := []*v1beta1.Authenticator{
				{
					Handler: &v1beta1.Handler{
						Name: v1beta1.AccessStrategyNoAuth,
					},
				},
			}

			corsPolicy := v1beta1.CorsPolicy{
				AllowMethods:     []string{"GET", "POST"},
				AllowOrigins:     v1beta1.StringMatch{{"exact": "localhost"}},
				AllowCredentials: ptr.To(true),
			}

			allowRule := GetRuleFor(ApiPath, ApiMethods, nil, strategies)
			rules := []v1beta1.Rule{allowRule}

			apiRule := GetAPIRuleFor(rules)
			apiRule.Spec.Host = &ServiceHostWithNoDomain
			apiRule.Spec.CorsPolicy = ptr.To(corsPolicy)

			client := GetFakeClient()
			processor := istio.Newv1beta1VirtualServiceProcessor(GetTestConfig(), apiRule)

			// when
			result, err := processor.EvaluateReconciliation(context.Background(), client)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(1))

			vs := result[0].Obj.(*networkingv1beta1.VirtualService)

			//verify VS
			Expect(vs).NotTo(BeNil())
			Expect(vs.Spec.Http[0].CorsPolicy.AllowOrigins).To(ConsistOf(&v1beta12.StringMatch{MatchType: &v1beta12.StringMatch_Exact{Exact: "localhost"}}))

			Expect(vs.Spec.Http[0].Headers.Response.Remove).To(ConsistOf([]string{
				builders.ExposeHeadersName,
				builders.MaxAgeName,
				builders.AllowHeadersName,
				builders.AllowCredentialsName,
				builders.AllowMethodsName,
				builders.AllowOriginName,
			}))

			Expect(vs.Spec.Http[0].CorsPolicy.AllowMethods).To(ConsistOf("GET", "POST"))
			Expect(vs.Spec.Http[0].CorsPolicy.AllowCredentials).To(Not(BeNil()))
			Expect(vs.Spec.Http[0].CorsPolicy.AllowCredentials.Value).To(BeTrue())
		})

		It("should remove all headers when CORSPolicy is empty", func() {
			// given
			strategies := []*v1beta1.Authenticator{
				{
					Handler: &v1beta1.Handler{
						Name: v1beta1.AccessStrategyNoAuth,
					},
				},
			}

			corsPolicy := v1beta1.CorsPolicy{}

			allowRule := GetRuleFor(ApiPath, ApiMethods, nil, strategies)
			rules := []v1beta1.Rule{allowRule}

			apiRule := GetAPIRuleFor(rules)
			apiRule.Spec.Host = &ServiceHostWithNoDomain
			apiRule.Spec.CorsPolicy = ptr.To(corsPolicy)

			client := GetFakeClient()
			processor := istio.Newv1beta1VirtualServiceProcessor(GetTestConfig(), apiRule)

			// when
			result, err := processor.EvaluateReconciliation(context.Background(), client)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(1))

			vs := result[0].Obj.(*networkingv1beta1.VirtualService)

			//verify VS
			Expect(vs).NotTo(BeNil())
			Expect(vs.Spec.Http[0].CorsPolicy).To(BeNil())

			Expect(vs.Spec.Http[0].Headers.Response.Remove).To(ConsistOf([]string{
				builders.ExposeHeadersName,
				builders.MaxAgeName,
				builders.AllowHeadersName,
				builders.AllowCredentialsName,
				builders.AllowMethodsName,
				builders.AllowOriginName,
			}))
		})

		It("should apply all CORSPolicy headers correctly", func() {
			// given
			strategies := []*v1beta1.Authenticator{
				{
					Handler: &v1beta1.Handler{
						Name: v1beta1.AccessStrategyNoAuth,
					},
				},
			}

			corsPolicy := v1beta1.CorsPolicy{
				AllowMethods:     []string{"GET", "POST"},
				AllowOrigins:     v1beta1.StringMatch{{"exact": "localhost"}},
				AllowCredentials: ptr.To(true),
				AllowHeaders:     []string{"Allowed-Header"},
				ExposeHeaders:    []string{"Exposed-Header"},
				MaxAge:           &metav1.Duration{Duration: 10 * time.Second},
			}

			allowRule := GetRuleFor(ApiPath, ApiMethods, nil, strategies)
			rules := []v1beta1.Rule{allowRule}

			apiRule := GetAPIRuleFor(rules)
			apiRule.Spec.Host = &ServiceHostWithNoDomain
			apiRule.Spec.CorsPolicy = ptr.To(corsPolicy)

			client := GetFakeClient()
			processor := istio.Newv1beta1VirtualServiceProcessor(GetTestConfig(), apiRule)

			// when
			result, err := processor.EvaluateReconciliation(context.Background(), client)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(1))

			vs := result[0].Obj.(*networkingv1beta1.VirtualService)

			//verify VS
			Expect(vs).NotTo(BeNil())
			Expect(vs.Spec.Http[0].CorsPolicy.AllowOrigins).To(ContainElements(&v1beta12.StringMatch{MatchType: &v1beta12.StringMatch_Exact{Exact: "localhost"}}))

			Expect(vs.Spec.Http[0].Headers.Response.Remove).To(ContainElements([]string{
				builders.ExposeHeadersName,
				builders.MaxAgeName,
				builders.AllowHeadersName,
				builders.AllowCredentialsName,
				builders.AllowMethodsName,
				builders.AllowOriginName,
			}))
		})
	})

	Context("timeout", func() {

		var (
			timeout10s v1beta1.Timeout = 10
			timeout20s v1beta1.Timeout = 20
		)

		It("should set default timeout when timeout is not configured", func() {
			// given
			strategies := []*v1beta1.Authenticator{
				{
					Handler: &v1beta1.Handler{
						Name: v1beta1.AccessStrategyNoAuth,
					},
				},
			}

			rule := GetRuleFor(ApiPath, ApiMethods, []*v1beta1.Mutator{}, strategies)
			rules := []v1beta1.Rule{rule}

			apiRule := GetAPIRuleFor(rules)
			client := GetFakeClient()
			processor := istio.Newv1beta1VirtualServiceProcessor(GetTestConfig(), apiRule)

			// when
			result, err := processor.EvaluateReconciliation(context.Background(), client)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(1))

			vs := result[0].Obj.(*networkingv1beta1.VirtualService)
			Expect(len(vs.Spec.Http)).To(Equal(1))

			Expect(vs.Spec.Http[0].Timeout.AsDuration()).To(Equal(180 * time.Second))
		})

		It("should set timeout from APIRule spec level when no timeout is configured for rule", func() {
			// given
			strategies := []*v1beta1.Authenticator{
				{
					Handler: &v1beta1.Handler{
						Name: v1beta1.AccessStrategyNoAuth,
					},
				},
			}

			rule := GetRuleFor(ApiPath, ApiMethods, []*v1beta1.Mutator{}, strategies)
			rules := []v1beta1.Rule{rule}

			apiRule := GetAPIRuleFor(rules)
			apiRule.Spec.Timeout = &timeout10s
			client := GetFakeClient()
			processor := istio.Newv1beta1VirtualServiceProcessor(GetTestConfig(), apiRule)

			// when
			result, err := processor.EvaluateReconciliation(context.Background(), client)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(1))

			vs := result[0].Obj.(*networkingv1beta1.VirtualService)
			Expect(len(vs.Spec.Http)).To(Equal(1))

			Expect(vs.Spec.Http[0].Timeout.AsDuration()).To(Equal(10 * time.Second))
		})

		It("should set timeout from rule level when timeout is configured for APIRule spec and rule", func() {
			// given
			strategies := []*v1beta1.Authenticator{
				{
					Handler: &v1beta1.Handler{
						Name: v1beta1.AccessStrategyNoAuth,
					},
				},
			}

			rule := GetRuleFor(ApiPath, ApiMethods, []*v1beta1.Mutator{}, strategies)
			rule.Timeout = &timeout20s
			rules := []v1beta1.Rule{rule}

			apiRule := GetAPIRuleFor(rules)
			apiRule.Spec.Timeout = &timeout10s
			client := GetFakeClient()
			processor := istio.Newv1beta1VirtualServiceProcessor(GetTestConfig(), apiRule)

			// when
			result, err := processor.EvaluateReconciliation(context.Background(), client)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(1))

			vs := result[0].Obj.(*networkingv1beta1.VirtualService)
			Expect(len(vs.Spec.Http)).To(Equal(1))

			Expect(vs.Spec.Http[0].Timeout.AsDuration()).To(Equal(20 * time.Second))
		})

		It(
			"should set timeout on rule with explicit timeout configuration and on rule that doesn't have timeout when there are multiple rules and timeout on api rule spec is configured",
			func() {
				// given
				strategies := []*v1beta1.Authenticator{
					{
						Handler: &v1beta1.Handler{
							Name: v1beta1.AccessStrategyNoAuth,
						},
					},
				}
				ruleWithoutTimeout := GetRuleFor("/api-rule-spec-timeout", ApiMethods, []*v1beta1.Mutator{}, strategies)
				ruleWithTimeout := GetRuleFor("/rule-timeout", ApiMethods, []*v1beta1.Mutator{}, strategies)
				ruleWithTimeout.Timeout = &timeout20s
				rules := []v1beta1.Rule{ruleWithoutTimeout, ruleWithTimeout}

				apiRule := GetAPIRuleFor(rules)
				apiRule.Spec.Timeout = &timeout10s
				client := GetFakeClient()
				processor := istio.Newv1beta1VirtualServiceProcessor(GetTestConfig(), apiRule)

				// when
				result, err := processor.EvaluateReconciliation(context.Background(), client)

				// then
				Expect(err).To(BeNil())
				Expect(result).To(HaveLen(1))

				vs := result[0].Obj.(*networkingv1beta1.VirtualService)
				Expect(len(vs.Spec.Http)).To(Equal(2))

				Expect(getTimeoutByPath(vs, "/api-rule-spec-timeout")).To(Equal(10 * time.Second))
				Expect(getTimeoutByPath(vs, "/rule-timeout")).To(Equal(20 * time.Second))
			},
		)

		It("should set timeout on rule with explicit timeout configuration and default timeout on rule that doesn't have a timeout when there are multiple rules", func() {
			// given
			strategies := []*v1beta1.Authenticator{
				{
					Handler: &v1beta1.Handler{
						Name: v1beta1.AccessStrategyNoAuth,
					},
				},
			}
			ruleWithoutTimeout := GetRuleFor("/default-timeout", ApiMethods, []*v1beta1.Mutator{}, strategies)
			ruleWithTimeout := GetRuleFor("/rule-timeout", ApiMethods, []*v1beta1.Mutator{}, strategies)
			ruleWithTimeout.Timeout = &timeout20s
			rules := []v1beta1.Rule{ruleWithoutTimeout, ruleWithTimeout}

			apiRule := GetAPIRuleFor(rules)
			client := GetFakeClient()
			processor := istio.Newv1beta1VirtualServiceProcessor(GetTestConfig(), apiRule)

			// when
			result, err := processor.EvaluateReconciliation(context.Background(), client)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(1))

			vs := result[0].Obj.(*networkingv1beta1.VirtualService)
			Expect(len(vs.Spec.Http)).To(Equal(2))

			Expect(getTimeoutByPath(vs, "/default-timeout")).To(Equal(180 * time.Second))
			Expect(getTimeoutByPath(vs, "/rule-timeout")).To(Equal(20 * time.Second))
		})
	})

	Context("HTTP matching", func() {

		DescribeTable("should restrict access for the path and methods defined in APIRule", func(handler string) {
			// Given
			strategies := []*v1beta1.Authenticator{
				{
					Handler: &v1beta1.Handler{
						Name: handler,
					},
				},
			}

			rule := GetRuleFor("/", methodsGetPost, []*v1beta1.Mutator{}, strategies)
			rules := []v1beta1.Rule{rule}

			apiRule := GetAPIRuleFor(rules)
			client := GetFakeClient()
			processor := istio.Newv1beta1VirtualServiceProcessor(GetTestConfig(), apiRule)

			// when
			result, err := processor.EvaluateReconciliation(context.Background(), client)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(1))

			vs := result[0].Obj.(*networkingv1beta1.VirtualService)

			Expect(vs).NotTo(BeNil())

			Expect(len(vs.Spec.Http[0].Match)).To(Equal(1))
			Expect(vs.Spec.Http[0].Match[0].Uri.GetRegex()).To(Equal("/"))
			Expect(vs.Spec.Http[0].Match[0].Method.GetRegex()).To(Equal("^(GET|POST)$"))
		},
			Entry("When access strategy is no_auth", v1beta1.AccessStrategyNoAuth),
			Entry("When access strategy is noop", v1beta1.AccessStrategyNoop),
			Entry("When access strategy is jwt", v1beta1.AccessStrategyJwt),
			Entry("When access strategy is oauth2_introspection", v1beta1.AccessStrategyOauth2Introspection),
		)

		It("should not restrict methods available for the given path when access strategy allow is used", func() {
			// Given
			allowStrategy := []*v1beta1.Authenticator{
				{
					Handler: &v1beta1.Handler{
						Name: v1beta1.AccessStrategyAllow,
					},
				},
			}

			allowRule := GetRuleFor("/", methodsGetPost, []*v1beta1.Mutator{}, allowStrategy)

			rules := []v1beta1.Rule{allowRule}

			apiRule := GetAPIRuleFor(rules)
			client := GetFakeClient()
			processor := istio.Newv1beta1VirtualServiceProcessor(GetTestConfig(), apiRule)

			// when
			result, err := processor.EvaluateReconciliation(context.Background(), client)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(1))

			vs := result[0].Obj.(*networkingv1beta1.VirtualService)

			Expect(vs).NotTo(BeNil())

			Expect(len(vs.Spec.Http[0].Match)).To(Equal(1))
			Expect(vs.Spec.Http[0].Match[0].Uri.GetRegex()).To(Equal("/"))
			Expect(vs.Spec.Http[0].Match[0].Method).To(BeNil())
		})
	})
})

func getTimeoutByPath(vs *networkingv1beta1.VirtualService, path string) time.Duration {
	for _, route := range vs.Spec.Http {
		if route.Match[0].Uri.GetRegex() == path {
			return route.Timeout.AsDuration()
		}
	}

	Fail(fmt.Sprintf("Path '%s' not found on virtual service", path))
	return 0
}
