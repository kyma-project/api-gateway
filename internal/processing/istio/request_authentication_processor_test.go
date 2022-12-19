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
			"authentications": [{"issuer": "%s", "jwksUri": "%s"}]}`, JwtIssuer, JwksUri)
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
			Name: &ServiceName,
			Port: &ServicePort,
		}

		ruleJwt := GetRuleWithServiceFor(HeadersApiPath, ApiMethods, []*gatewayv1beta1.Mutator{}, []*gatewayv1beta1.Authenticator{jwt}, service)
		ruleJwt2 := GetRuleWithServiceFor(ImgApiPath, ApiMethods, []*gatewayv1beta1.Mutator{}, []*gatewayv1beta1.Authenticator{jwt}, service)
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
		Expect(ra.ObjectMeta.GenerateName).To(Equal(ApiName + "-"))
		Expect(ra.ObjectMeta.Namespace).To(Equal(ApiNamespace))
		Expect(ra.ObjectMeta.Labels[TestLabelKey]).To(Equal(TestLabelValue))

		Expect(len(ra.OwnerReferences)).To(Equal(1))
		Expect(ra.OwnerReferences[0].APIVersion).To(Equal(ApiAPIVersion))
		Expect(ra.OwnerReferences[0].Kind).To(Equal(ApiKind))
		Expect(ra.OwnerReferences[0].Name).To(Equal(ApiName))
		Expect(ra.OwnerReferences[0].UID).To(Equal(ApiUID))

		Expect(ra.Spec.Selector.MatchLabels[TestSelectorKey]).NotTo(BeNil())
		Expect(ra.Spec.Selector.MatchLabels[TestSelectorKey]).To(Equal(ServiceName))
		Expect(len(ra.Spec.JwtRules)).To(Equal(1))
		Expect(ra.Spec.JwtRules[0].Issuer).To(Equal(JwtIssuer))
		Expect(ra.Spec.JwtRules[0].JwksUri).To(Equal(JwksUri))
	})

	It("should produce RA for a Rule without service, but service definition on ApiRule level", func() {
		// given
		jwt := createIstioJwtAccessStrategy()
		client := GetEmptyFakeClient()
		ruleJwt := GetRuleFor(HeadersApiPath, ApiMethods, []*gatewayv1beta1.Mutator{}, []*gatewayv1beta1.Authenticator{jwt})
		apiRule := GetAPIRuleFor([]gatewayv1beta1.Rule{ruleJwt})
		processor := istio.NewRequestAuthenticationProcessor(GetTestConfig())

		// when
		result, err := processor.EvaluateReconciliation(context.TODO(), client, apiRule)

		// then
		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(1))

		ra := result[0].Obj.(*securityv1beta1.RequestAuthentication)
		Expect(ra).NotTo(BeNil())
		Expect(ra.Spec.Selector.MatchLabels[TestSelectorKey]).To(Equal(ServiceName))
	})

	It("should produce RA with service from Rule, when service is configured on Rule and ApiRule level", func() {
		// given
		jwt := createIstioJwtAccessStrategy()
		ruleServiceName := "rule-scope-example-service"
		service := &gatewayv1beta1.Service{
			Name: &ruleServiceName,
			Port: &ServicePort,
		}
		client := GetEmptyFakeClient()
		ruleJwt := GetRuleWithServiceFor(HeadersApiPath, ApiMethods, []*gatewayv1beta1.Mutator{}, []*gatewayv1beta1.Authenticator{jwt}, service)
		apiRule := GetAPIRuleFor([]gatewayv1beta1.Rule{ruleJwt})

		processor := istio.NewRequestAuthenticationProcessor(GetTestConfig())

		// when
		result, err := processor.EvaluateReconciliation(context.TODO(), client, apiRule)

		// then
		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(1))

		ra := result[0].Obj.(*securityv1beta1.RequestAuthentication)
		Expect(ra).NotTo(BeNil())
		Expect(ra.Spec.Selector.MatchLabels[TestSelectorKey]).To(Equal(ruleServiceName))
	})

	It("should produce RA from a rule with two issuers and one path", func() {
		jwtConfigJSON := fmt.Sprintf(`{
			"authentications": [{"issuer": "%s", "jwksUri": "%s"}, {"issuer": "%s", "jwksUri": "%s"}]
			}`, JwtIssuer, JwksUri, JwtIssuer2, JwksUri2)
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
			Name: &ServiceName,
			Port: &ServicePort,
		}
		ruleJwt := GetRuleWithServiceFor(HeadersApiPath, ApiMethods, []*gatewayv1beta1.Mutator{}, []*gatewayv1beta1.Authenticator{jwt}, service)
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
		Expect(ra.ObjectMeta.GenerateName).To(Equal(ApiName + "-"))
		Expect(ra.ObjectMeta.Namespace).To(Equal(ApiNamespace))
		Expect(ra.ObjectMeta.Labels[TestLabelKey]).To(Equal(TestLabelValue))

		Expect(len(ra.OwnerReferences)).To(Equal(1))
		Expect(ra.OwnerReferences[0].APIVersion).To(Equal(ApiAPIVersion))
		Expect(ra.OwnerReferences[0].Kind).To(Equal(ApiKind))
		Expect(ra.OwnerReferences[0].Name).To(Equal(ApiName))
		Expect(ra.OwnerReferences[0].UID).To(Equal(ApiUID))

		Expect(ra.Spec.Selector.MatchLabels[TestSelectorKey]).NotTo(BeNil())
		Expect(ra.Spec.Selector.MatchLabels[TestSelectorKey]).To(Equal(ServiceName))
		Expect(len(ra.Spec.JwtRules)).To(Equal(2))
		Expect(ra.Spec.JwtRules[0].Issuer).To(Equal(JwtIssuer))
		Expect(ra.Spec.JwtRules[0].JwksUri).To(Equal(JwksUri))
		Expect(ra.Spec.JwtRules[1].Issuer).To(Equal(JwtIssuer2))
		Expect(ra.Spec.JwtRules[1].JwksUri).To(Equal(JwksUri2))
	})

	It("should not create RA if handler is allow", func() {
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

		client := GetEmptyFakeClient()
		processor := istio.NewRequestAuthenticationProcessor(GetTestConfig())

		// when
		result, err := processor.EvaluateReconciliation(context.TODO(), client, apiRule)

		// then
		Expect(err).To(BeNil())
		Expect(result).To(BeEmpty())
	})

	It("should not create RA if handler is noop", func() {
		// given
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

		overrideServiceName := "testName"
		overrideServiceNamespace := "testName-namespace"
		overrideServicePort := uint32(8080)

		apiRule.Spec.Service = &gatewayv1beta1.Service{
			Name:      &overrideServiceName,
			Namespace: &overrideServiceNamespace,
			Port:      &overrideServicePort,
		}

		client := GetEmptyFakeClient()
		processor := istio.NewRequestAuthenticationProcessor(GetTestConfig())

		// when
		result, err := processor.EvaluateReconciliation(context.TODO(), client, apiRule)

		// then
		Expect(err).To(BeNil())
		Expect(result).To(BeEmpty())
	})
})
