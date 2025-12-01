package istio_test

import (
	"context"
	"fmt"
	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"

	. "github.com/kyma-project/api-gateway/internal/processing/processing_test"
	"github.com/kyma-project/api-gateway/internal/processing/processors/istio"
	rulev1alpha1 "github.com/kyma-project/api-gateway/internal/types/ory/oathkeeper-maester/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Reconciliation", func() {
	When("multiple handlers in addition to Istio JWT", func() {
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

		DescribeTable("should provide Istio VS, RA and 2 APs with handler", func(handler string) {
			// given
			allow := []*gatewayv1beta1.Authenticator{
				{
					Handler: &gatewayv1beta1.Handler{
						Name: handler,
					},
				},
			}

			allowRule := GetRuleFor(ApiPath, ApiMethods, []*gatewayv1beta1.Mutator{}, allow)
			jwtRule := GetRuleFor(HeadersApiPath, ApiMethods, []*gatewayv1beta1.Mutator{}, jwt)
			rules := []gatewayv1beta1.Rule{allowRule, jwtRule}

			apiRule := GetAPIRuleFor(rules)
			service := GetService(ServiceName)
			fakeClient := GetFakeClient(service)

			// when
			var createdObjects []client.Object
			reconciliation := istio.NewIstioReconciliation(apiRule, GetTestConfig(), &testLogger, fakeClient)
			for _, processor := range reconciliation.GetProcessors() {
				results, err := processor.EvaluateReconciliation(context.Background(), fakeClient)
				Expect(err).To(BeNil())
				for _, result := range results {
					createdObjects = append(createdObjects, result.Obj)
				}
			}

			// then
			Expect(createdObjects).To(HaveLen(4))

			vsCreated, raCreated := false, false
			apCreated := 0
			for _, createdObj := range createdObjects {
				vs, vsOk := createdObj.(*networkingv1beta1.VirtualService)
				ra, raOk := createdObj.(*securityv1beta1.RequestAuthentication)
				ap, apOk := createdObj.(*securityv1beta1.AuthorizationPolicy)

				if vsOk {
					vsCreated = true
					Expect(vs).NotTo(BeNil())
					Expect(len(vs.Spec.Http)).To(Equal(2))
					Expect(len(vs.Spec.Http[1].Route)).To(Equal(1))
					Expect(vs.Spec.Http[1].Route[0].Destination.Host).To(Equal(ServiceName + "." + ApiNamespace + ".svc.cluster.local"))
				} else if raOk {
					raCreated = true
					Expect(ra).NotTo(BeNil())
					Expect(len(ra.Spec.JwtRules)).To(Equal(1))
					Expect(ra.Spec.JwtRules[0].Issuer).To(Equal(JwtIssuer))
					Expect(ra.Spec.JwtRules[0].JwksUri).To(Equal(JwksUri))
				} else if apOk {
					apCreated++
					Expect(ap).NotTo(BeNil())
					Expect(len(ap.Spec.Rules)).To(Equal(1))
				}
			}

			Expect(vsCreated && raCreated && apCreated == 2).To(BeTrue())
		},
			Entry("only no_auth handler", gatewayv1beta1.AccessStrategyNoAuth),
			Entry("only allow handler", gatewayv1beta1.AccessStrategyAllow),
		)

		It("with Ory oauth2 should provide Istio VS, AP, RA and Ory rule", func() {
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
			service := GetService(ServiceName)
			fakeClient := GetFakeClient(service)

			// when
			var createdObjects []client.Object
			reconciliation := istio.NewIstioReconciliation(apiRule, GetTestConfig(), &testLogger, fakeClient)
			for _, processor := range reconciliation.GetProcessors() {
				results, err := processor.EvaluateReconciliation(context.Background(), fakeClient)
				Expect(err).To(BeNil())
				for _, result := range results {
					createdObjects = append(createdObjects, result.Obj)
				}
			}

			// then
			Expect(createdObjects).To(HaveLen(5))

			vsCreated, oryRuleCreated, raCreated := false, false, false
			apCreated := 0
			for _, createdObj := range createdObjects {
				vs, vsOk := createdObj.(*networkingv1beta1.VirtualService)
				ar, arOk := createdObj.(*rulev1alpha1.Rule)
				ra, raOk := createdObj.(*securityv1beta1.RequestAuthentication)
				ap, apOk := createdObj.(*securityv1beta1.AuthorizationPolicy)

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
				} else if raOk {
					raCreated = true
					Expect(ra).NotTo(BeNil())
					Expect(len(ra.Spec.JwtRules)).To(Equal(1))
					Expect(ra.Spec.JwtRules[0].Issuer).To(Equal(JwtIssuer))
					Expect(ra.Spec.JwtRules[0].JwksUri).To(Equal(JwksUri))
				} else if apOk {
					apCreated++
					Expect(ap).NotTo(BeNil())
					Expect(len(ap.Spec.Rules)).To(Equal(1))
				}
			}

			Expect(vsCreated && oryRuleCreated && raCreated && apCreated == 2).To(BeTrue())
		})

		It("with Ory noop should provide Istio VS, AP, RA and Ory rule", func() {
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
			service := GetService(ServiceName)
			fakeClient := GetFakeClient(service)

			// when
			var createdObjects []client.Object
			reconciliation := istio.NewIstioReconciliation(apiRule, GetTestConfig(), &testLogger, fakeClient)
			for _, processor := range reconciliation.GetProcessors() {
				results, err := processor.EvaluateReconciliation(context.Background(), fakeClient)
				Expect(err).To(BeNil())
				for _, result := range results {
					createdObjects = append(createdObjects, result.Obj)
				}
			}

			// then
			Expect(createdObjects).To(HaveLen(5))

			vsCreated, oryRuleCreated, raCreated := false, false, false
			apCreated := 0
			for _, createdObj := range createdObjects {
				vs, vsOk := createdObj.(*networkingv1beta1.VirtualService)
				ar, arOk := createdObj.(*rulev1alpha1.Rule)
				ra, raOk := createdObj.(*securityv1beta1.RequestAuthentication)
				ap, apOk := createdObj.(*securityv1beta1.AuthorizationPolicy)

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
				} else if raOk {
					raCreated = true
					Expect(ra).NotTo(BeNil())
					Expect(len(ra.Spec.JwtRules)).To(Equal(1))
					Expect(ra.Spec.JwtRules[0].Issuer).To(Equal(JwtIssuer))
					Expect(ra.Spec.JwtRules[0].JwksUri).To(Equal(JwksUri))
				} else if apOk {
					apCreated++
					Expect(ap).NotTo(BeNil())
					Expect(len(ap.Spec.Rules)).To(Equal(1))
				}
			}

			Expect(vsCreated && oryRuleCreated && raCreated && apCreated == 2).To(BeTrue())
		})
	})
})
