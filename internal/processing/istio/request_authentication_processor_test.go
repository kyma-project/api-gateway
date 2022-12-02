package istio_test

import (
	"context"
	"fmt"

	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	. "github.com/kyma-incubator/api-gateway/internal/processing/internal/test"
	"github.com/kyma-incubator/api-gateway/internal/processing/istio"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
)

var _ = Describe("Request Authentication Processor", func() {

	createIstioJwtAccessStrategy := func() *gatewayv1beta1.Authenticator {
		jwtConfigJSON := fmt.Sprintf(`{
			"authentications": [{"issuer": "%s", "jwksUri": "%s"}]}`, jwtIssuer, jwksUri)
		return &gatewayv1beta1.Authenticator{
			Handler: &gatewayv1beta1.Handler{
				Name: "jwt",
				Config: &runtime.RawExtension{
					Raw: []byte(jwtConfigJSON),
				},
			},
		}
	}

	It("should produce one RA for a rule with one issuer and two paths", func() {
		// given
		jwt := createIstioJwtAccessStrategy()
		service := &gatewayv1beta1.Service{
			Name: &serviceName,
			Port: &servicePort,
		}

		ruleJwt := GetRuleWithServiceFor(headersAPIPath, apiMethods, []*gatewayv1beta1.Mutator{}, []*gatewayv1beta1.Authenticator{jwt}, service)
		ruleJwt2 := GetRuleWithServiceFor(oauthAPIPath, apiMethods, []*gatewayv1beta1.Mutator{}, []*gatewayv1beta1.Authenticator{jwt}, service)
		apiRule := GetAPIRuleFor([]gatewayv1beta1.Rule{ruleJwt, ruleJwt2})
		client := GetEmptyFakeClient()
		processor := istio.NewRequestAuthenticationProcessor(GetTestConfig())

		// when
		result, err := processor.EvaluateReconciliation(context.TODO(), client, apiRule)

		// then
		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(1))
		ra := result[0].Obj.(*securityv1beta1.RequestAuthentication)
		Expect(ra).NotTo(BeNil())
		Expect(ra.ObjectMeta.Name).To(BeEmpty())
		Expect(ra.ObjectMeta.GenerateName).To(Equal(apiName + "-"))
		Expect(ra.ObjectMeta.Namespace).To(Equal(apiNamespace))
		Expect(ra.ObjectMeta.Labels[testLabelKey]).To(Equal(testLabelValue))

		Expect(len(ra.OwnerReferences)).To(Equal(1))
		Expect(ra.OwnerReferences[0].APIVersion).To(Equal(apiAPIVersion))
		Expect(ra.OwnerReferences[0].Kind).To(Equal(apiKind))
		Expect(ra.OwnerReferences[0].Name).To(Equal(apiName))
		Expect(ra.OwnerReferences[0].UID).To(Equal(apiUID))

		Expect(ra.Spec.Selector.MatchLabels[testSelectorKey]).NotTo(BeNil())
		Expect(ra.Spec.Selector.MatchLabels[testSelectorKey]).To(Equal(serviceName))
		Expect(len(ra.Spec.JwtRules)).To(Equal(1))
		Expect(ra.Spec.JwtRules[0].Issuer).To(Equal(jwtIssuer))
		Expect(ra.Spec.JwtRules[0].JwksUri).To(Equal(jwksUri))
	})

	It("should produce RA for a Rule without service, but service definition on ApiRule level", func() {
		// given
		jwt := createIstioJwtAccessStrategy()
		client := GetEmptyFakeClient()
		ruleJwt := GetRuleFor(headersAPIPath, apiMethods, []*gatewayv1beta1.Mutator{}, []*gatewayv1beta1.Authenticator{jwt})
		apiRule := GetAPIRuleFor([]gatewayv1beta1.Rule{ruleJwt})
		processor := istio.NewRequestAuthenticationProcessor(GetTestConfig())

		// when
		result, err := processor.EvaluateReconciliation(context.TODO(), client, apiRule)

		// then
		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(1))

		ra := result[0].Obj.(*securityv1beta1.RequestAuthentication)
		Expect(ra).NotTo(BeNil())
		Expect(ra.Spec.Selector.MatchLabels[testSelectorKey]).To(Equal(serviceName))
	})

	It("should produce RA with service from Rule, when service is configured on Rule and ApiRule level", func() {
		// given
		jwt := createIstioJwtAccessStrategy()
		ruleServiceName := "rule-scope-example-service"
		service := &gatewayv1beta1.Service{
			Name: &ruleServiceName,
			Port: &servicePort,
		}
		client := GetEmptyFakeClient()
		ruleJwt := GetRuleWithServiceFor(headersAPIPath, apiMethods, []*gatewayv1beta1.Mutator{}, []*gatewayv1beta1.Authenticator{jwt}, service)
		apiRule := GetAPIRuleFor([]gatewayv1beta1.Rule{ruleJwt})

		processor := istio.NewRequestAuthenticationProcessor(GetTestConfig())

		// when
		result, err := processor.EvaluateReconciliation(context.TODO(), client, apiRule)

		// then
		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(1))

		ra := result[0].Obj.(*securityv1beta1.RequestAuthentication)
		Expect(ra).NotTo(BeNil())
		Expect(ra.Spec.Selector.MatchLabels[testSelectorKey]).To(Equal(ruleServiceName))
	})
	It("should produce RA from a rule with two issuers and one path", func() {
		jwtConfigJSON := fmt.Sprintf(`{
			"authentications": [{"issuer": "%s", "jwksUri": "%s"}, {"issuer": "%s", "jwksUri": "%s"}]
			}`, jwtIssuer, jwksUri, jwtIssuer2, jwksUri2)
		jwt := &gatewayv1beta1.Authenticator{
			Handler: &gatewayv1beta1.Handler{
				Name: "jwt",
				Config: &runtime.RawExtension{
					Raw: []byte(jwtConfigJSON),
				},
			},
		}
		client := GetEmptyFakeClient()
		service := &gatewayv1beta1.Service{
			Name: &serviceName,
			Port: &servicePort,
		}
		ruleJwt := GetRuleWithServiceFor(headersAPIPath, apiMethods, []*gatewayv1beta1.Mutator{}, []*gatewayv1beta1.Authenticator{jwt}, service)
		apiRule := GetAPIRuleFor([]gatewayv1beta1.Rule{ruleJwt})
		processor := istio.NewRequestAuthenticationProcessor(GetTestConfig())

		// when
		result, err := processor.EvaluateReconciliation(context.TODO(), client, apiRule)

		// then
		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(1))
		ra := result[0].Obj.(*securityv1beta1.RequestAuthentication)

		Expect(ra).NotTo(BeNil())
		Expect(ra.ObjectMeta.Name).To(BeEmpty())
		Expect(ra.ObjectMeta.GenerateName).To(Equal(apiName + "-"))
		Expect(ra.ObjectMeta.Namespace).To(Equal(apiNamespace))
		Expect(ra.ObjectMeta.Labels[testLabelKey]).To(Equal(testLabelValue))

		Expect(len(ra.OwnerReferences)).To(Equal(1))
		Expect(ra.OwnerReferences[0].APIVersion).To(Equal(apiAPIVersion))
		Expect(ra.OwnerReferences[0].Kind).To(Equal(apiKind))
		Expect(ra.OwnerReferences[0].Name).To(Equal(apiName))
		Expect(ra.OwnerReferences[0].UID).To(Equal(apiUID))

		Expect(ra.Spec.Selector.MatchLabels[testSelectorKey]).NotTo(BeNil())
		Expect(ra.Spec.Selector.MatchLabels[testSelectorKey]).To(Equal(serviceName))
		Expect(len(ra.Spec.JwtRules)).To(Equal(2))
		Expect(ra.Spec.JwtRules[0].Issuer).To(Equal(jwtIssuer))
		Expect(ra.Spec.JwtRules[0].JwksUri).To(Equal(jwksUri))
		Expect(ra.Spec.JwtRules[1].Issuer).To(Equal(jwtIssuer2))
		Expect(ra.Spec.JwtRules[1].JwksUri).To(Equal(jwksUri2))
	})
})
