package istio_test

import (
	"context"
	"fmt"

	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	"github.com/kyma-incubator/api-gateway/internal/builders"
	. "github.com/kyma-incubator/api-gateway/internal/processing/internal/test"
	"github.com/kyma-incubator/api-gateway/internal/processing/istio"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	TestAudience1 string = "kyma-goats"
	TestAudience2 string = "kyma-goats2"
)

var _ = Describe("JwtAuthorization Policy Processor", func() {
	createIstioJwtAccessStrategy := func() *gatewayv1beta1.Authenticator {
		jwtConfigJSON := fmt.Sprintf(`{
			"authentications": [{"issuer": "%s", "jwksUri": "%s"}],
			"authorizations": [{"requiredScopes": ["%s", "%s"], "audiences": ["%s", "%s"]}]}`, JwtIssuer, JwksUri, RequiredScopeA, RequiredScopeB, TestAudience1, TestAudience2)
		return &gatewayv1beta1.Authenticator{
			Handler: &gatewayv1beta1.Handler{
				Name: "jwt",
				Config: &runtime.RawExtension{
					Raw: []byte(jwtConfigJSON),
				},
			},
		}
	}

	createIstioJwtAccessStrategyWithAudiencesOnly := func() *gatewayv1beta1.Authenticator {
		jwtConfigJSON := fmt.Sprintf(`{
			"authentications": [{"issuer": "%s", "jwksUri": "%s"}],
			"authorizations": [{"audiences": ["%s", "%s"]}]}`, JwtIssuer, JwksUri, TestAudience1, TestAudience2)
		return &gatewayv1beta1.Authenticator{
			Handler: &gatewayv1beta1.Handler{
				Name: "jwt",
				Config: &runtime.RawExtension{
					Raw: []byte(jwtConfigJSON),
				},
			},
		}
	}

	It("should produce one AP for a rule with two audiences", func() {
		// given
		jwt := createIstioJwtAccessStrategyWithAudiencesOnly()
		service := &gatewayv1beta1.Service{
			Name: &ServiceName,
			Port: &ServicePort,
		}

		ruleJwt := GetRuleWithServiceFor(HeadersApiPath, ApiMethods, []*gatewayv1beta1.Mutator{}, []*gatewayv1beta1.Authenticator{jwt}, service)
		apiRule := GetAPIRuleFor([]gatewayv1beta1.Rule{ruleJwt})
		client := GetFakeClient()
		processor := istio.NewAuthorizationPolicyProcessor(GetTestConfig())

		// when
		result, err := processor.EvaluateReconciliation(context.TODO(), client, apiRule)

		// then
		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(1))

		ap1 := result[0].Obj.(*securityv1beta1.AuthorizationPolicy)

		Expect(ap1).NotTo(BeNil())
		Expect(ap1.Spec.Rules).To(HaveLen(1))
		Expect(ap1.Spec.Rules[0].When).To(HaveLen(2))
		Expect(ap1.Spec.Rules[0].When).To(ContainElement(builders.NewConditionBuilder().WithKey("request.auth.audiences").WithValues([]string{TestAudience1}).Get()))
		Expect(ap1.Spec.Rules[0].When).To(ContainElement(builders.NewConditionBuilder().WithKey("request.auth.audiences").WithValues([]string{TestAudience2}).Get()))
	})

	It("should produce one AP for a rule with two scopes and two audiences", func() {
		// given
		jwt := createIstioJwtAccessStrategy()
		service := &gatewayv1beta1.Service{
			Name: &ServiceName,
			Port: &ServicePort,
		}

		ruleJwt := GetRuleWithServiceFor(HeadersApiPath, ApiMethods, []*gatewayv1beta1.Mutator{}, []*gatewayv1beta1.Authenticator{jwt}, service)
		apiRule := GetAPIRuleFor([]gatewayv1beta1.Rule{ruleJwt})
		client := GetFakeClient()
		processor := istio.NewAuthorizationPolicyProcessor(GetTestConfig())

		// when
		result, err := processor.EvaluateReconciliation(context.TODO(), client, apiRule)

		// then
		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(1))

		ap1 := result[0].Obj.(*securityv1beta1.AuthorizationPolicy)

		Expect(ap1).NotTo(BeNil())
		Expect(ap1.Spec.Rules).To(HaveLen(3))
		Expect(ap1.Spec.Rules[0].When).To(HaveLen(4))
		Expect(ap1.Spec.Rules[0].When).To(ContainElement(builders.NewConditionBuilder().WithKey("request.auth.audiences").WithValues([]string{TestAudience1}).Get()))
		Expect(ap1.Spec.Rules[0].When).To(ContainElement(builders.NewConditionBuilder().WithKey("request.auth.audiences").WithValues([]string{TestAudience2}).Get()))
	})
})
