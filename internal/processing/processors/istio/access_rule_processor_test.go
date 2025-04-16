package istio_test

import (
	"context"
	"fmt"
	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"net/http"

	"github.com/kyma-project/api-gateway/internal/processing"
	. "github.com/kyma-project/api-gateway/internal/processing/processing_test"
	"github.com/kyma-project/api-gateway/internal/processing/processors/istio"
	"github.com/kyma-project/api-gateway/internal/processing/processors/ory"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	rulev1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("Access Rule Processor", func() {

	DescribeTable("should not create access rules when handler is", func(handler string) {
		// given
		strategies := []*gatewayv1beta1.Authenticator{
			{
				Handler: &gatewayv1beta1.Handler{
					Name: handler,
				},
			},
		}

		rule := GetRuleFor(ApiPath, ApiMethods, []*gatewayv1beta1.Mutator{}, strategies)
		rules := []gatewayv1beta1.Rule{rule}

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
		processor := istio.Newv1beta1AccessRuleProcessor(GetTestConfig(), apiRule)

		// when
		result, err := processor.EvaluateReconciliation(context.Background(), client)

		// then
		Expect(err).To(BeNil())
		Expect(result).To(BeEmpty())
	},
		Entry(nil, gatewayv1beta1.AccessStrategyNoAuth),
		Entry(nil, gatewayv1beta1.AccessStrategyAllow),
		Entry(nil, gatewayv1beta1.AccessStrategyJwt),
	)

	When("handler is noop", func() {
		It("should override rule with meta data", func() {
			// given
			strategies := []*gatewayv1beta1.Authenticator{
				{
					Handler: &gatewayv1beta1.Handler{
						Name: "noop",
					},
				},
			}

			allowRule := GetRuleWithServiceFor(ApiPath, ApiMethods, []*gatewayv1beta1.Mutator{}, strategies, nil)
			rules := []gatewayv1beta1.Rule{allowRule}

			apiRule := GetAPIRuleFor(rules)
			client := GetFakeClient()
			processor := istio.Newv1beta1AccessRuleProcessor(GetTestConfig(), apiRule)

			// when
			result, err := processor.EvaluateReconciliation(context.Background(), client)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(1))

			accessRule := result[0].Obj.(*rulev1alpha1.Rule)

			Expect(accessRule.ObjectMeta.Name).To(BeEmpty())
			Expect(accessRule.ObjectMeta.GenerateName).To(Equal(ApiName + "-"))
			Expect(accessRule.ObjectMeta.Namespace).To(Equal(ApiNamespace))

		})

		It("should override rule upstream with rule level service", func() {
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
			processor := istio.Newv1beta1AccessRuleProcessor(GetTestConfig(), apiRule)

			// when
			result, err := processor.EvaluateReconciliation(context.Background(), client)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(1))

			accessRule := result[0].Obj.(*rulev1alpha1.Rule)
			expectedRuleUpstreamURL := fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", overrideServiceName, ApiNamespace, overrideServicePort)
			Expect(accessRule.Spec.Upstream.URL).To(Equal(expectedRuleUpstreamURL))
		})

		It("should override rule upstream with rule level service for specified namespace", func() {
			// given
			strategies := []*gatewayv1beta1.Authenticator{
				{
					Handler: &gatewayv1beta1.Handler{
						Name: "noop",
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
			processor := istio.Newv1beta1AccessRuleProcessor(GetTestConfig(), apiRule)

			// when
			result, err := processor.EvaluateReconciliation(context.Background(), client)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(1))

			accessRule := result[0].Obj.(*rulev1alpha1.Rule)
			expectedRuleUpstreamURL := fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", overrideServiceName, overrideServiceNamespace, overrideServicePort)
			Expect(accessRule.Spec.Upstream.URL).To(Equal(expectedRuleUpstreamURL))
		})

		It("should return rule with default domain name when the hostname does not contain domain name", func() {
			strategies := []*gatewayv1beta1.Authenticator{
				{
					Handler: &gatewayv1beta1.Handler{
						Name: "noop",
					},
				},
			}

			allowRule := GetRuleFor(ApiPath, ApiMethods, []*gatewayv1beta1.Mutator{}, strategies)
			rules := []gatewayv1beta1.Rule{allowRule}

			apiRule := GetAPIRuleFor(rules)
			apiRule.Spec.Host = &ServiceHostWithNoDomain
			client := GetFakeClient()
			processor := istio.Newv1beta1AccessRuleProcessor(GetTestConfig(), apiRule)

			// when
			result, err := processor.EvaluateReconciliation(context.Background(), client)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(1))

			expectedRuleMatchURL := fmt.Sprintf("<http|https>://%s<%s>", ServiceHost, ApiPath)

			accessRule := result[0].Obj.(*rulev1alpha1.Rule)
			Expect(accessRule.Spec.Match.URL).To(Equal(expectedRuleMatchURL))
		})

		Context("when existing rule has owner v1beta1 owner label", func() {
			It("should get and update match methods of rule", func() {
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
							processing.OwnerLabel: fmt.Sprintf("%s.%s", apiRule.Name, apiRule.Namespace),
						},
					},
					Spec: rulev1alpha1.RuleSpec{
						Match: &rulev1alpha1.Match{
							URL:     fmt.Sprintf("<http|https>://%s<%s>", ServiceHost, ApiPath),
							Methods: []string{"DELETE"},
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
				err = gatewayv1beta1.AddToScheme(scheme)
				Expect(err).NotTo(HaveOccurred())

				client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(&rule, &vs).Build()
				processor := istio.Newv1beta1AccessRuleProcessor(GetTestConfig(), apiRule)

				// when
				result, err := processor.EvaluateReconciliation(context.Background(), client)

				// then
				Expect(err).To(BeNil())
				Expect(result).To(HaveLen(1))
				Expect(result[0].Action.String()).To(Equal("update"))

				accessRule := result[0].Obj.(*rulev1alpha1.Rule)
				Expect(accessRule.Spec.Match.Methods).To(Equal([]string{"GET"}))
			})
		})
	})

	When("handler is oauth2", func() {
		It("should return rule for oauth authenticators for given path", func() {
			// given
			oauthConfigJSON := fmt.Sprintf(`{"required_scope": [%s]}`, ToCSVList(ApiScopes))
			oauth := &gatewayv1beta1.Authenticator{
				Handler: &gatewayv1beta1.Handler{
					Name: "oauth2_introspection",
					Config: &runtime.RawExtension{
						Raw: []byte(oauthConfigJSON),
					},
				},
			}

			strategies := []*gatewayv1beta1.Authenticator{oauth}

			allowRule := GetRuleFor(ApiPath, ApiMethods, []*gatewayv1beta1.Mutator{}, strategies)
			rules := []gatewayv1beta1.Rule{allowRule}

			apiRule := GetAPIRuleFor(rules)
			client := GetFakeClient()
			processor := ory.NewAccessRuleProcessor(GetTestConfig(), apiRule)

			// when
			result, err := processor.EvaluateReconciliation(context.Background(), client)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(1))

			rule := result[0].Obj.(*rulev1alpha1.Rule)

			Expect(len(rule.Spec.Authenticators)).To(Equal(1))

			Expect(rule.Spec.Authorizer.Name).To(Equal("allow"))
			Expect(rule.Spec.Authorizer.Config).To(BeNil())

			Expect(rule.Spec.Authenticators[0].Handler.Name).To(Equal("oauth2_introspection"))
			Expect(rule.Spec.Authenticators[0].Handler.Config).NotTo(BeNil())
			Expect(string(rule.Spec.Authenticators[0].Handler.Config.Raw)).To(Equal(oauthConfigJSON))

			expectedRuleMatchURL := fmt.Sprintf("<http|https>://%s<%s>", ServiceHost, ApiPath)
			expectedRuleUpstreamURL := fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", ServiceName, ApiNamespace, ServicePort)

			Expect(len(rule.Spec.Match.Methods)).To(Equal(len(ApiMethods)))
			Expect(rule.Spec.Match.Methods).To(Equal([]string{http.MethodGet}))
			Expect(rule.Spec.Match.URL).To(Equal(expectedRuleMatchURL))

			Expect(rule.Spec.Upstream.URL).To(Equal(expectedRuleUpstreamURL))
		})
	})
})
