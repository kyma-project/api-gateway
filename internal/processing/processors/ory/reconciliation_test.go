package ory_test

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	rulev1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
	"golang.org/x/exp/slices"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	. "github.com/kyma-project/api-gateway/internal/processing/processing_test"
	"github.com/kyma-project/api-gateway/internal/processing/processors/ory"
)

var _ = Describe("Reconciliation", func() {
	When("multiple handlers in addition to ory jwt", func() {
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

		It("with ory noop should provide only ory rules", func() {
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
			faceClient := GetFakeClient()

			// when
			var createdObjects []client.Object
			reconciliation := ory.NewOryReconciliation(apiRule, GetTestConfig(), &testLogger)
			for _, processor := range reconciliation.GetProcessors() {
				results, err := processor.EvaluateReconciliation(context.Background(), faceClient)
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
			Expect(createdObjects).To(HaveLen(3))

			vsCreated, oryRuleCreated := false, false
			for _, createdObj := range createdObjects {
				vs, vsOk := createdObj.(*networkingv1beta1.VirtualService)
				ar, arOk := createdObj.(*rulev1alpha1.Rule)

				if vsOk {
					vsCreated = true
					Expect(vs).NotTo(BeNil())
					Expect(vs.Spec.Http[1].Route[0].Destination.Host).To(Equal(OathkeeperSvc))
				} else if arOk {
					oryRuleCreated = true
					Expect(ar).NotTo(BeNil())
					expectedHandlers := []string{"jwt", "noop"}
					Expect(slices.Contains(expectedHandlers, ar.Spec.Authenticators[0].Handler.Name)).To(BeTrue())
					Expect(ar.Spec.Upstream.URL).To(Equal(fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", ServiceName, ApiNamespace, ServicePort)))
				}
			}

			Expect(vsCreated && oryRuleCreated).To(BeTrue())
		})

		It("with ory oauth2 should provide only ory rules", func() {
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
			faceClient := GetFakeClient()

			// when
			var createdObjects []client.Object
			reconciliation := ory.NewOryReconciliation(apiRule, GetTestConfig(), &testLogger)
			for _, processor := range reconciliation.GetProcessors() {
				results, err := processor.EvaluateReconciliation(context.Background(), faceClient)
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
			Expect(createdObjects).To(HaveLen(3))

			vsCreated, oryRuleCreated := false, false
			for _, createdObj := range createdObjects {
				vs, vsOk := createdObj.(*networkingv1beta1.VirtualService)
				ar, arOk := createdObj.(*rulev1alpha1.Rule)

				if vsOk {
					vsCreated = true
					Expect(vs).NotTo(BeNil())
					Expect(vs.Spec.Http[1].Route[0].Destination.Host).To(Equal(OathkeeperSvc))
				} else if arOk {
					oryRuleCreated = true
					Expect(ar).NotTo(BeNil())
					expectedHandlers := []string{"jwt", "oauth2_introspection"}
					Expect(slices.Contains(expectedHandlers, ar.Spec.Authenticators[0].Handler.Name)).To(BeTrue())
					Expect(ar.Spec.Upstream.URL).To(Equal(fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", ServiceName, ApiNamespace, ServicePort)))
				}
			}

			Expect(vsCreated && oryRuleCreated).To(BeTrue())
		})
	})
})
