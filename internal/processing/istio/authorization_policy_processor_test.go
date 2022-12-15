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

var _ = Describe("Authorization Policy Processor", func() {
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

	It("should produce two APs for a rule with one issuer and two paths", func() {
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
		processor := istio.NewAuthorizationPolicyProcessor(GetTestConfig())

		// when
		result, err := processor.EvaluateReconciliation(context.TODO(), client, apiRule)

		ap1 := result[0].Obj.(*securityv1beta1.AuthorizationPolicy)
		ap2 := result[1].Obj.(*securityv1beta1.AuthorizationPolicy)

		// then
		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(2))

		Expect(ap1).NotTo(BeNil())
		Expect(ap1.ObjectMeta.Name).To(BeEmpty())
		Expect(ap1.ObjectMeta.GenerateName).To(Equal(ApiName + "-"))
		Expect(ap1.ObjectMeta.Namespace).To(Equal(ApiNamespace))
		Expect(ap1.ObjectMeta.Labels[TestLabelKey]).To(Equal(TestLabelValue))

		Expect(ap1.Spec.Selector.MatchLabels[TestSelectorKey]).NotTo(BeNil())
		Expect(ap1.Spec.Selector.MatchLabels[TestSelectorKey]).To(Equal(ServiceName))
		Expect(len(ap1.Spec.Rules)).To(Equal(1))
		Expect(len(ap1.Spec.Rules[0].From)).To(Equal(1))
		Expect(len(ap1.Spec.Rules[0].From[0].Source.RequestPrincipals)).To(Equal(1))
		Expect(ap1.Spec.Rules[0].From[0].Source.RequestPrincipals[0]).To(Equal("*"))
		Expect(len(ap1.Spec.Rules[0].To)).To(Equal(1))
		Expect(len(ap1.Spec.Rules[0].To[0].Operation.Methods)).To(Equal(1))
		Expect(ap1.Spec.Rules[0].To[0].Operation.Methods).To(ContainElements(ApiMethods))
		Expect(len(ap1.Spec.Rules[0].To[0].Operation.Paths)).To(Equal(1))

		Expect(len(ap1.OwnerReferences)).To(Equal(1))
		Expect(ap1.OwnerReferences[0].APIVersion).To(Equal(ApiAPIVersion))
		Expect(ap1.OwnerReferences[0].Kind).To(Equal(ApiKind))
		Expect(ap1.OwnerReferences[0].Name).To(Equal(ApiName))
		Expect(ap1.OwnerReferences[0].UID).To(Equal(ApiUID))

		Expect(ap2).NotTo(BeNil())
		Expect(ap2.ObjectMeta.Name).To(BeEmpty())
		Expect(ap2.ObjectMeta.GenerateName).To(Equal(ApiName + "-"))
		Expect(ap2.ObjectMeta.Namespace).To(Equal(ApiNamespace))
		Expect(ap2.ObjectMeta.Labels[TestLabelKey]).To(Equal(TestLabelValue))

		Expect(ap2.Spec.Selector.MatchLabels[TestSelectorKey]).NotTo(BeNil())
		Expect(ap2.Spec.Selector.MatchLabels[TestSelectorKey]).To(Equal(ServiceName))
		Expect(len(ap2.Spec.Rules)).To(Equal(1))
		Expect(len(ap2.Spec.Rules[0].From)).To(Equal(1))
		Expect(len(ap2.Spec.Rules[0].From[0].Source.RequestPrincipals)).To(Equal(1))
		Expect(ap2.Spec.Rules[0].From[0].Source.RequestPrincipals[0]).To(Equal("*"))
		Expect(len(ap2.Spec.Rules[0].To)).To(Equal(1))
		Expect(len(ap2.Spec.Rules[0].To[0].Operation.Methods)).To(Equal(1))
		Expect(ap2.Spec.Rules[0].To[0].Operation.Methods).To(ContainElements(ApiMethods))
		Expect(len(ap2.Spec.Rules[0].To[0].Operation.Paths)).To(Equal(1))

		Expect(ap2.OwnerReferences[0].APIVersion).To(Equal(ApiAPIVersion))
		Expect(ap2.OwnerReferences[0].Kind).To(Equal(ApiKind))
		Expect(ap2.OwnerReferences[0].Name).To(Equal(ApiName))
		Expect(ap2.OwnerReferences[0].UID).To(Equal(ApiUID))
	})

	It("should produce one AP for a Rule without service, but service definition on ApiRule level", func() {
		// given
		jwt := createIstioJwtAccessStrategy()
		client := GetEmptyFakeClient()
		ruleJwt := GetRuleFor(HeadersApiPath, ApiMethods, []*gatewayv1beta1.Mutator{}, []*gatewayv1beta1.Authenticator{jwt})
		apiRule := GetAPIRuleFor([]gatewayv1beta1.Rule{ruleJwt})
		processor := istio.NewAuthorizationPolicyProcessor(GetTestConfig())

		// when
		result, err := processor.EvaluateReconciliation(context.TODO(), client, apiRule)

		// then
		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(1))

		ap := result[0].Obj.(*securityv1beta1.AuthorizationPolicy)
		Expect(ap).NotTo(BeNil())
		Expect(ap.Spec.Selector.MatchLabels[TestSelectorKey]).To(Equal(ServiceName))
	})

	It("should produce AP with service from Rule, when service is configured on Rule and ApiRule level", func() {
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
		processor := istio.NewAuthorizationPolicyProcessor(GetTestConfig())

		// when
		result, err := processor.EvaluateReconciliation(context.TODO(), client, apiRule)

		// then
		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(1))

		ap := result[0].Obj.(*securityv1beta1.AuthorizationPolicy)
		Expect(ap).NotTo(BeNil())
		Expect(ap.Spec.Selector.MatchLabels[TestSelectorKey]).To(Equal(ruleServiceName))
	})

	It("should produce AP from a rule with two issuers and one path", func() {
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
		processor := istio.NewAuthorizationPolicyProcessor(GetTestConfig())

		// when
		result, err := processor.EvaluateReconciliation(context.TODO(), client, apiRule)

		// then
		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(1))
		ap := result[0].Obj.(*securityv1beta1.AuthorizationPolicy)

		Expect(ap).NotTo(BeNil())
		Expect(ap.ObjectMeta.Name).To(BeEmpty())
		Expect(ap.ObjectMeta.GenerateName).To(Equal(ApiName + "-"))
		Expect(ap.ObjectMeta.Namespace).To(Equal(ApiNamespace))
		Expect(ap.ObjectMeta.Labels[TestLabelKey]).To(Equal(TestLabelValue))

		Expect(ap.Spec.Selector.MatchLabels[TestSelectorKey]).NotTo(BeNil())
		Expect(ap.Spec.Selector.MatchLabels[TestSelectorKey]).To(Equal(ServiceName))
		Expect(len(ap.Spec.Rules)).To(Equal(1))
		Expect(len(ap.Spec.Rules[0].From)).To(Equal(1))
		Expect(len(ap.Spec.Rules[0].From[0].Source.RequestPrincipals)).To(Equal(1))
		Expect(ap.Spec.Rules[0].From[0].Source.RequestPrincipals[0]).To(Equal("*"))
		Expect(len(ap.Spec.Rules[0].To)).To(Equal(1))
		Expect(len(ap.Spec.Rules[0].To[0].Operation.Methods)).To(Equal(1))
		Expect(ap.Spec.Rules[0].To[0].Operation.Methods).To(ContainElements(ApiMethods))
		Expect(len(ap.Spec.Rules[0].To[0].Operation.Paths)).To(Equal(1))
		Expect(ap.Spec.Rules[0].To[0].Operation.Paths).To(ContainElements(HeadersApiPath))

		Expect(ap.OwnerReferences[0].APIVersion).To(Equal(ApiAPIVersion))
		Expect(ap.OwnerReferences[0].Kind).To(Equal(ApiKind))
		Expect(ap.OwnerReferences[0].Name).To(Equal(ApiName))
		Expect(ap.OwnerReferences[0].UID).To(Equal(ApiUID))
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
		processor := istio.NewAuthorizationPolicyProcessor(GetTestConfig())

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
		processor := istio.NewAuthorizationPolicyProcessor(GetTestConfig())

		// when
		result, err := processor.EvaluateReconciliation(context.TODO(), client, apiRule)

		// then
		Expect(err).To(BeNil())
		Expect(result).To(BeEmpty())
	})
})
