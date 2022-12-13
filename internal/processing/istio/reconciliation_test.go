package istio_test

import (
	"context"
	"fmt"

	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	. "github.com/kyma-incubator/api-gateway/internal/processing/internal/test"
	"github.com/kyma-incubator/api-gateway/internal/processing/istio"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	rulev1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Reconciliation", func() {
	When("multiple handlers in addtion to istio jwt", func() {
		jwtConfigJSON := fmt.Sprintf(`{"authentications": [{"issuer": "%s", "jwksUri": "%s"}]}`, JwtIssuer, JwksUri)
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

		It("with ory noop should provide istio vs and ory rule", func() {
			// given
			noop := []*gatewayv1beta1.Authenticator{
				{
					Handler: &gatewayv1beta1.Handler{
						Name: "noop",
					},
				},
			}

			noopRule := GetRuleFor(ApiPath, ApiMethods, []*gatewayv1beta1.Mutator{}, noop)
			jwtRule := GetRuleFor(HeadersApiPath, ApiMethods, []*gatewayv1beta1.Mutator{}, jwt)
			rules := []gatewayv1beta1.Rule{noopRule, jwtRule}

			apiRule := GetAPIRuleFor(rules)
			faceClient := GetEmptyFakeClient()

			// when
			var createdObjects []client.Object
			reconciliation := istio.NewIstioReconciliation(GetTestConfig())
			for _, processor := range reconciliation.GetProcessors() {
				results, err := processor.EvaluateReconciliation(context.TODO(), faceClient, apiRule)
				Expect(err).To(BeNil())

				for _, result := range results {
					_, vsOk := result.Obj.(*networkingv1beta1.VirtualService)
					_, arOk := result.Obj.(*rulev1alpha1.Rule)

					if vsOk || arOk {
						createdObjects = append(createdObjects, result.Obj)
					}
				}
			}

			// then
			Expect(createdObjects).To(HaveLen(2))

			vsCreated, oryRuleCreated := false, false
			for _, createdObj := range createdObjects {
				vs, vsOk := createdObj.(*networkingv1beta1.VirtualService)
				ar, arOk := createdObj.(*rulev1alpha1.Rule)

				if vsOk {
					vsCreated = true

					Expect(vs).NotTo(BeNil())
					Expect(len(vs.Spec.Http)).To(Equal(2))
					Expect(len(vs.Spec.Http[1].Route)).To(Equal(1))
					Expect(vs.Spec.Http[1].Route[0].Destination.Host).To(Equal(ServiceName + "." + ApiNamespace + ".svc.cluster.local"))
				} else if arOk {
					oryRuleCreated = true

					Expect(ar).NotTo(BeNil())
					Expect(ar.Spec.Authenticators[0].Handler.Name).To(Equal("noop"))
					Expect(ar.Spec.Upstream.URL).To(Equal(fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", ServiceName, ApiNamespace, ServicePort)))
				}
			}

			Expect(vsCreated && oryRuleCreated).To(BeTrue())
		})

		It("with ory oauth2 should provide istio vs and ory rule", func() {
			// given
			oauthConfigJSON := fmt.Sprintf(`{"required_scope": [%s]}`, ToCSVList(ApiScopes))
			oauth := []*gatewayv1beta1.Authenticator{
				{
					Handler: &gatewayv1beta1.Handler{
						Name: "oauth2_introspection",
						Config: &runtime.RawExtension{
							Raw: []byte(oauthConfigJSON),
						},
					},
				},
			}

			oauthRule := GetRuleFor(ApiPath, ApiMethods, []*gatewayv1beta1.Mutator{}, oauth)
			jwtRule := GetRuleFor(HeadersApiPath, ApiMethods, []*gatewayv1beta1.Mutator{}, jwt)
			rules := []gatewayv1beta1.Rule{oauthRule, jwtRule}

			apiRule := GetAPIRuleFor(rules)
			faceClient := GetEmptyFakeClient()

			// when
			var createdObjects []client.Object
			reconciliation := istio.NewIstioReconciliation(GetTestConfig())
			for _, processor := range reconciliation.GetProcessors() {
				results, err := processor.EvaluateReconciliation(context.TODO(), faceClient, apiRule)
				Expect(err).To(BeNil())

				for _, result := range results {
					_, vsOk := result.Obj.(*networkingv1beta1.VirtualService)
					_, arOk := result.Obj.(*rulev1alpha1.Rule)

					if vsOk || arOk {
						createdObjects = append(createdObjects, result.Obj)
					}
				}
			}

			// then
			Expect(createdObjects).To(HaveLen(2))

			vsCreated, oryRuleCreated := false, false
			for _, createdObj := range createdObjects {
				vs, vsOk := createdObj.(*networkingv1beta1.VirtualService)
				ar, arOk := createdObj.(*rulev1alpha1.Rule)

				if vsOk {
					vsCreated = true

					Expect(vs).NotTo(BeNil())
					Expect(len(vs.Spec.Http)).To(Equal(2))
					Expect(len(vs.Spec.Http[1].Route)).To(Equal(1))
					Expect(vs.Spec.Http[1].Route[0].Destination.Host).To(Equal(ServiceName + "." + ApiNamespace + ".svc.cluster.local"))
				} else if arOk {
					oryRuleCreated = true

					Expect(ar).NotTo(BeNil())
					Expect(ar.Spec.Authenticators[0].Handler.Name).To(Equal("oauth2_introspection"))
					Expect(ar.Spec.Upstream.URL).To(Equal(fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", ServiceName, ApiNamespace, ServicePort)))
				}
			}

			Expect(vsCreated && oryRuleCreated).To(BeTrue())
		})
	})
})
