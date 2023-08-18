package ory_test

import (
	"context"
	"fmt"
	"time"

	gatewayv1beta1 "github.com/kyma-project/api-gateway/api/v1beta1"
	"github.com/kyma-project/api-gateway/internal/processing"
	. "github.com/kyma-project/api-gateway/internal/processing/internal/test"
	"github.com/kyma-project/api-gateway/internal/processing/ory"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	rulev1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("Virtual Service Processor", func() {
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

			allowRule := GetRuleFor(ApiPath, ApiMethods, []*gatewayv1beta1.Mutator{}, strategies)
			rules := []gatewayv1beta1.Rule{allowRule}

			apiRule := GetAPIRuleFor(rules)
			client := GetFakeClient()
			processor := ory.NewVirtualServiceProcessor(GetTestConfig())

			// when
			result, err := processor.EvaluateReconciliation(context.TODO(), client, apiRule)

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
			Expect(vs.ObjectMeta.Labels[TestLabelKey]).To(Equal(TestLabelValue))
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

			allowRule := GetRuleFor(ApiPath, ApiMethods, []*gatewayv1beta1.Mutator{}, strategies)
			rules := []gatewayv1beta1.Rule{allowRule}

			apiRule := GetAPIRuleFor(rules)

			overrideServiceName := "testName"
			overrideServiceNamespace := "testName-namespace"
			overrideServicePort := uint32(8080)

			apiRule.Spec.Service = &gatewayv1beta1.Service{
				Name:      &overrideServiceName,
				Namespace: &overrideServiceNamespace,
				Port:      &overrideServicePort,
			}
			client := GetFakeClient()
			processor := ory.NewVirtualServiceProcessor(GetTestConfig())

			// when
			result, err := processor.EvaluateReconciliation(context.TODO(), client, apiRule)

			// then
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

			allowRule := GetRuleWithServiceFor(ApiPath, ApiMethods, []*gatewayv1beta1.Mutator{}, strategies, service)
			rules := []gatewayv1beta1.Rule{allowRule}

			apiRule := GetAPIRuleFor(rules)
			client := GetFakeClient()
			processor := ory.NewVirtualServiceProcessor(GetTestConfig())

			// when
			result, err := processor.EvaluateReconciliation(context.TODO(), client, apiRule)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(1))

			vs := result[0].Obj.(*networkingv1beta1.VirtualService)

			//verify VS has rule level destination host
			Expect(len(vs.Spec.Http[0].Route)).To(Equal(1))
			Expect(vs.Spec.Http[0].Route[0].Destination.Host).To(Equal(overrideServiceName + "." + overrideServiceNamespace + ".svc.cluster.local"))

		})
		It("should return VS with default domain name when the hostname does not contain domain name", func() {
			strategies := []*gatewayv1beta1.Authenticator{
				{
					Handler: &gatewayv1beta1.Handler{
						Name: "allow",
					},
				},
			}

			allowRule := GetRuleFor(ApiPath, ApiMethods, []*gatewayv1beta1.Mutator{}, strategies)
			rules := []gatewayv1beta1.Rule{allowRule}

			apiRule := GetAPIRuleFor(rules)
			apiRule.Spec.Host = &ServiceHostWithNoDomain
			client := GetFakeClient()
			processor := ory.NewVirtualServiceProcessor(GetTestConfig())

			// when
			result, err := processor.EvaluateReconciliation(context.TODO(), client, apiRule)

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

			allowRule := GetRuleWithServiceFor(ApiPath, ApiMethods, []*gatewayv1beta1.Mutator{}, strategies, service)
			rules := []gatewayv1beta1.Rule{allowRule}

			apiRule := GetAPIRuleFor(rules)
			client := GetFakeClient()
			processor := ory.NewVirtualServiceProcessor(GetTestConfig())

			// when
			result, err := processor.EvaluateReconciliation(context.TODO(), client, apiRule)

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
				noop := []*gatewayv1beta1.Authenticator{
					{
						Handler: &gatewayv1beta1.Handler{
							Name: "noop",
						},
					},
				}

				noopRule := GetRuleFor(ApiPath, ApiMethods, []*gatewayv1beta1.Mutator{}, noop)
				rules := []gatewayv1beta1.Rule{noopRule}

				apiRule := GetAPIRuleFor(rules)

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
				processor := ory.NewVirtualServiceProcessor(GetTestConfig())

				// when
				result, err := processor.EvaluateReconciliation(context.TODO(), client, apiRule)

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
					}`, JwtIssuer, ToCSVList(ApiScopes))

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

			noopRule := GetRuleFor(ApiPath, ApiMethods, []*gatewayv1beta1.Mutator{}, noop)
			jwtRule := GetRuleFor(HeadersApiPath, ApiMethods, testMutators, jwt)
			rules := []gatewayv1beta1.Rule{noopRule, jwtRule}

			apiRule := GetAPIRuleFor(rules)
			client := GetFakeClient()
			processor := ory.NewVirtualServiceProcessor(GetTestConfig())

			// when
			result, err := processor.EvaluateReconciliation(context.TODO(), client, apiRule)

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
			Expect(vs.Spec.Http[1].Route[0].Destination.Host).To(Equal(OathkeeperSvc))
			Expect(vs.Spec.Http[1].Route[0].Destination.Port.Number).To(Equal(OathkeeperSvcPort))
			Expect(len(vs.Spec.Http[1].Match)).To(Equal(1))
			Expect(vs.Spec.Http[1].Match[0].Uri.GetRegex()).To(Equal(apiRule.Spec.Rules[1].Path))

			Expect(vs.Spec.Http[1].CorsPolicy.AllowOrigins).To(Equal(TestCors.AllowOrigins))
			Expect(vs.Spec.Http[1].CorsPolicy.AllowMethods).To(Equal(TestCors.AllowMethods))
			Expect(vs.Spec.Http[1].CorsPolicy.AllowHeaders).To(Equal(TestCors.AllowHeaders))

			Expect(vs.ObjectMeta.Name).To(BeEmpty())
			Expect(vs.ObjectMeta.GenerateName).To(Equal(ApiName + "-"))
			Expect(vs.ObjectMeta.Namespace).To(Equal(ApiNamespace))
			Expect(vs.ObjectMeta.Labels[TestLabelKey]).To(Equal(TestLabelValue))
		})

		It("should return service for two same paths and different methods", func() {
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
					}`, JwtIssuer, ToCSVList(ApiScopes))

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
			noopRule := GetRuleFor(ApiPath, getMethod, []*gatewayv1beta1.Mutator{}, noop)
			jwtRule := GetRuleFor(ApiPath, postMethod, testMutators, jwt)
			rules := []gatewayv1beta1.Rule{noopRule, jwtRule}

			apiRule := GetAPIRuleFor(rules)
			client := GetFakeClient()
			processor := ory.NewVirtualServiceProcessor(GetTestConfig())

			// when
			result, err := processor.EvaluateReconciliation(context.TODO(), client, apiRule)

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
			Expect(vs.ObjectMeta.Labels[TestLabelKey]).To(Equal(TestLabelValue))
		})

		It("should return service for two same paths and one different", func() {
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
					}`, JwtIssuer, ToCSVList(ApiScopes))

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
			noopGetRule := GetRuleFor(ApiPath, getMethod, []*gatewayv1beta1.Mutator{}, noop)
			noopPostRule := GetRuleFor(ApiPath, postMethod, []*gatewayv1beta1.Mutator{}, noop)
			jwtRule := GetRuleFor(HeadersApiPath, ApiMethods, testMutators, jwt)
			rules := []gatewayv1beta1.Rule{noopGetRule, noopPostRule, jwtRule}

			apiRule := GetAPIRuleFor(rules)
			client := GetFakeClient()
			processor := ory.NewVirtualServiceProcessor(GetTestConfig())

			// when
			result, err := processor.EvaluateReconciliation(context.TODO(), client, apiRule)

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
			Expect(vs.Spec.Http[1].Route[0].Destination.Host).To(Equal(OathkeeperSvc))
			Expect(vs.Spec.Http[1].Route[0].Destination.Port.Number).To(Equal(OathkeeperSvcPort))
			Expect(len(vs.Spec.Http[1].Match)).To(Equal(1))
			Expect(vs.Spec.Http[1].Match[0].Uri.GetRegex()).To(Equal(apiRule.Spec.Rules[2].Path))

			Expect(vs.Spec.Http[1].CorsPolicy.AllowOrigins).To(Equal(TestCors.AllowOrigins))
			Expect(vs.Spec.Http[1].CorsPolicy.AllowMethods).To(Equal(TestCors.AllowMethods))
			Expect(vs.Spec.Http[1].CorsPolicy.AllowHeaders).To(Equal(TestCors.AllowHeaders))

			Expect(vs.ObjectMeta.Name).To(BeEmpty())
			Expect(vs.ObjectMeta.GenerateName).To(Equal(ApiName + "-"))
			Expect(vs.ObjectMeta.Namespace).To(Equal(ApiNamespace))
			Expect(vs.ObjectMeta.Labels[TestLabelKey]).To(Equal(TestLabelValue))
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

			allowRule := GetRuleFor(ApiPath, ApiMethods, []*gatewayv1beta1.Mutator{}, strategies)
			rules := []gatewayv1beta1.Rule{allowRule}

			apiRule := GetAPIRuleFor(rules)
			client := GetFakeClient()
			processor := ory.NewVirtualServiceProcessor(GetTestConfig())

			// when
			result, err := processor.EvaluateReconciliation(context.TODO(), client, apiRule)

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
			Expect(vs.ObjectMeta.Labels[TestLabelKey]).To(Equal(TestLabelValue))
		})
	})

	Context("timeout", func() {

		var (
			timeout10s gatewayv1beta1.Timeout = 10
			timeout20s gatewayv1beta1.Timeout = 20
		)

		It("should set default timeout when timeout is not configured", func() {
			// given
			strategies := []*gatewayv1beta1.Authenticator{
				{
					Handler: &gatewayv1beta1.Handler{
						Name: "allow",
					},
				},
			}

			rule := GetRuleFor(ApiPath, ApiMethods, []*gatewayv1beta1.Mutator{}, strategies)
			rules := []gatewayv1beta1.Rule{rule}

			apiRule := GetAPIRuleFor(rules)
			client := GetFakeClient()
			processor := ory.NewVirtualServiceProcessor(GetTestConfig())

			// when
			result, err := processor.EvaluateReconciliation(context.TODO(), client, apiRule)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(1))

			vs := result[0].Obj.(*networkingv1beta1.VirtualService)
			Expect(len(vs.Spec.Http)).To(Equal(1))

			Expect(vs.Spec.Http[0].Timeout.AsDuration()).To(Equal(180 * time.Second))
		})

		It("should set timeout from APIRule spec level when no timeout is configured for rule", func() {
			// given
			strategies := []*gatewayv1beta1.Authenticator{
				{
					Handler: &gatewayv1beta1.Handler{
						Name: "allow",
					},
				},
			}

			rule := GetRuleFor(ApiPath, ApiMethods, []*gatewayv1beta1.Mutator{}, strategies)
			rules := []gatewayv1beta1.Rule{rule}

			apiRule := GetAPIRuleFor(rules)
			apiRule.Spec.Timeout = &timeout10s
			client := GetFakeClient()
			processor := ory.NewVirtualServiceProcessor(GetTestConfig())

			// when
			result, err := processor.EvaluateReconciliation(context.TODO(), client, apiRule)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(1))

			vs := result[0].Obj.(*networkingv1beta1.VirtualService)
			Expect(len(vs.Spec.Http)).To(Equal(1))

			Expect(vs.Spec.Http[0].Timeout.AsDuration()).To(Equal(10 * time.Second))
		})

		It("should set timeout from rule level when timeout is configured for APIRule spec and rule", func() {
			// given
			strategies := []*gatewayv1beta1.Authenticator{
				{
					Handler: &gatewayv1beta1.Handler{
						Name: "allow",
					},
				},
			}

			rule := GetRuleFor(ApiPath, ApiMethods, []*gatewayv1beta1.Mutator{}, strategies)
			rule.Timeout = &timeout20s
			rules := []gatewayv1beta1.Rule{rule}

			apiRule := GetAPIRuleFor(rules)
			apiRule.Spec.Timeout = &timeout10s
			client := GetFakeClient()
			processor := ory.NewVirtualServiceProcessor(GetTestConfig())

			// when
			result, err := processor.EvaluateReconciliation(context.TODO(), client, apiRule)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(1))

			vs := result[0].Obj.(*networkingv1beta1.VirtualService)
			Expect(len(vs.Spec.Http)).To(Equal(1))

			Expect(vs.Spec.Http[0].Timeout.AsDuration()).To(Equal(20 * time.Second))
		})

		It("should set timeout on rule with explicit timeout configuration and on rule that doesn't have timeout when there are multiple rules and timeout on api rule spec is configured", func() {
			// given
			strategies := []*gatewayv1beta1.Authenticator{
				{
					Handler: &gatewayv1beta1.Handler{
						Name: "allow",
					},
				},
			}
			ruleWithoutTimeout := GetRuleFor("/api-rule-spec-timeout", ApiMethods, []*gatewayv1beta1.Mutator{}, strategies)
			ruleWithTimeout := GetRuleFor("/rule-timeout", ApiMethods, []*gatewayv1beta1.Mutator{}, strategies)
			ruleWithTimeout.Timeout = &timeout20s
			rules := []gatewayv1beta1.Rule{ruleWithoutTimeout, ruleWithTimeout}

			apiRule := GetAPIRuleFor(rules)
			apiRule.Spec.Timeout = &timeout10s
			client := GetFakeClient()
			processor := ory.NewVirtualServiceProcessor(GetTestConfig())

			// when
			result, err := processor.EvaluateReconciliation(context.TODO(), client, apiRule)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(1))

			vs := result[0].Obj.(*networkingv1beta1.VirtualService)
			Expect(len(vs.Spec.Http)).To(Equal(2))

			Expect(getTimeoutByPath(vs, "/api-rule-spec-timeout")).To(Equal(10 * time.Second))
			Expect(getTimeoutByPath(vs, "/rule-timeout")).To(Equal(20 * time.Second))
		})

		It("should set timeout on rule with explicit timeout configuration and default timeout on rule that doesn't have a timeout when there are multiple rules", func() {
			// given
			strategies := []*gatewayv1beta1.Authenticator{
				{
					Handler: &gatewayv1beta1.Handler{
						Name: "allow",
					},
				},
			}
			ruleWithoutTimeout := GetRuleFor("/default-timeout", ApiMethods, []*gatewayv1beta1.Mutator{}, strategies)
			ruleWithTimeout := GetRuleFor("/rule-timeout", ApiMethods, []*gatewayv1beta1.Mutator{}, strategies)
			ruleWithTimeout.Timeout = &timeout20s
			rules := []gatewayv1beta1.Rule{ruleWithoutTimeout, ruleWithTimeout}

			apiRule := GetAPIRuleFor(rules)
			client := GetFakeClient()
			processor := ory.NewVirtualServiceProcessor(GetTestConfig())

			// when
			result, err := processor.EvaluateReconciliation(context.TODO(), client, apiRule)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(1))

			vs := result[0].Obj.(*networkingv1beta1.VirtualService)
			Expect(len(vs.Spec.Http)).To(Equal(2))

			Expect(getTimeoutByPath(vs, "/default-timeout")).To(Equal(180 * time.Second))
			Expect(getTimeoutByPath(vs, "/rule-timeout")).To(Equal(20 * time.Second))
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
