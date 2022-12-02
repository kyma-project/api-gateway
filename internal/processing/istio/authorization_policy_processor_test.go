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
	"k8s.io/apimachinery/pkg/types"
)

const (
	apiName                   = "test-apirule"
	apiUID          types.UID = "eab0f1c8-c417-11e9-bf11-4ac644044351"
	apiNamespace              = "some-namespace"
	apiAPIVersion             = "gateway.kyma-project.io/v1alpha1"
	apiKind                   = "ApiRule"
	headersAPIPath            = "/headers"
	oauthAPIPath              = "/img"
	jwtIssuer                 = "https://oauth2.example.com/"
	jwksUri                   = "https://oauth2.example.com/.well-known/jwks.json"
	jwtIssuer2                = "https://oauth2.another.example.com/"
	jwksUri2                  = "https://oauth2.another.example.com/.well-known/jwks.json"
	testLabelKey              = "key"
	testLabelValue            = "value"
	testSelectorKey           = "app"
)

var (
	apiMethods         = []string{"GET"}
	servicePort uint32 = 8080
	serviceName        = "example-service"
)

var _ = Describe("Authorization Policy Processor", func() {

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
	It("should produce two APs for a rule with one issuer and two paths", func() {
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
		Expect(ap1.ObjectMeta.GenerateName).To(Equal(apiName + "-"))
		Expect(ap1.ObjectMeta.Namespace).To(Equal(apiNamespace))
		Expect(ap1.ObjectMeta.Labels[testLabelKey]).To(Equal(testLabelValue))

		Expect(ap1.Spec.Selector.MatchLabels[testSelectorKey]).NotTo(BeNil())
		Expect(ap1.Spec.Selector.MatchLabels[testSelectorKey]).To(Equal(serviceName))
		Expect(len(ap1.Spec.Rules)).To(Equal(1))
		Expect(len(ap1.Spec.Rules[0].From)).To(Equal(1))
		Expect(len(ap1.Spec.Rules[0].From[0].Source.RequestPrincipals)).To(Equal(1))
		Expect(ap1.Spec.Rules[0].From[0].Source.RequestPrincipals[0]).To(Equal("*"))
		Expect(len(ap1.Spec.Rules[0].To)).To(Equal(1))
		Expect(len(ap1.Spec.Rules[0].To[0].Operation.Methods)).To(Equal(1))
		Expect(ap1.Spec.Rules[0].To[0].Operation.Methods).To(ContainElements(apiMethods))
		Expect(len(ap1.Spec.Rules[0].To[0].Operation.Paths)).To(Equal(1))

		Expect(len(ap1.OwnerReferences)).To(Equal(1))
		Expect(ap1.OwnerReferences[0].APIVersion).To(Equal(apiAPIVersion))
		Expect(ap1.OwnerReferences[0].Kind).To(Equal(apiKind))
		Expect(ap1.OwnerReferences[0].Name).To(Equal(apiName))
		Expect(ap1.OwnerReferences[0].UID).To(Equal(apiUID))

		Expect(ap2).NotTo(BeNil())
		Expect(ap2.ObjectMeta.Name).To(BeEmpty())
		Expect(ap2.ObjectMeta.GenerateName).To(Equal(apiName + "-"))
		Expect(ap2.ObjectMeta.Namespace).To(Equal(apiNamespace))
		Expect(ap2.ObjectMeta.Labels[testLabelKey]).To(Equal(testLabelValue))

		Expect(ap2.Spec.Selector.MatchLabels[testSelectorKey]).NotTo(BeNil())
		Expect(ap2.Spec.Selector.MatchLabels[testSelectorKey]).To(Equal(serviceName))
		Expect(len(ap2.Spec.Rules)).To(Equal(1))
		Expect(len(ap2.Spec.Rules[0].From)).To(Equal(1))
		Expect(len(ap2.Spec.Rules[0].From[0].Source.RequestPrincipals)).To(Equal(1))
		Expect(ap2.Spec.Rules[0].From[0].Source.RequestPrincipals[0]).To(Equal("*"))
		Expect(len(ap2.Spec.Rules[0].To)).To(Equal(1))
		Expect(len(ap2.Spec.Rules[0].To[0].Operation.Methods)).To(Equal(1))
		Expect(ap2.Spec.Rules[0].To[0].Operation.Methods).To(ContainElements(apiMethods))
		Expect(len(ap2.Spec.Rules[0].To[0].Operation.Paths)).To(Equal(1))

		Expect(ap2.OwnerReferences[0].APIVersion).To(Equal(apiAPIVersion))
		Expect(ap2.OwnerReferences[0].Kind).To(Equal(apiKind))
		Expect(ap2.OwnerReferences[0].Name).To(Equal(apiName))
		Expect(ap2.OwnerReferences[0].UID).To(Equal(apiUID))
	})

	It("should produce one AP for a Rule without service, but service definition on ApiRule level", func() {
		// given
		jwt := createIstioJwtAccessStrategy()
		client := GetEmptyFakeClient()
		ruleJwt := GetRuleFor(headersAPIPath, apiMethods, []*gatewayv1beta1.Mutator{}, []*gatewayv1beta1.Authenticator{jwt})
		apiRule := GetAPIRuleFor([]gatewayv1beta1.Rule{ruleJwt})
		processor := istio.NewAuthorizationPolicyProcessor(GetTestConfig())

		// when
		result, err := processor.EvaluateReconciliation(context.TODO(), client, apiRule)

		// then
		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(1))

		ap := result[0].Obj.(*securityv1beta1.AuthorizationPolicy)
		Expect(ap).NotTo(BeNil())
		Expect(ap.Spec.Selector.MatchLabels[testSelectorKey]).To(Equal(serviceName))
	})

	It("should produce AP with service from Rule, when service is configured on Rule and ApiRule level", func() {
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

		processor := istio.NewAuthorizationPolicyProcessor(GetTestConfig())

		// when
		result, err := processor.EvaluateReconciliation(context.TODO(), client, apiRule)

		// then
		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(1))

		ap := result[0].Obj.(*securityv1beta1.AuthorizationPolicy)
		Expect(ap).NotTo(BeNil())
		Expect(ap.Spec.Selector.MatchLabels[testSelectorKey]).To(Equal(ruleServiceName))
	})

	It("should produce AP from a rule with two issuers and one path", func() {
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
		processor := istio.NewAuthorizationPolicyProcessor(GetTestConfig())

		// when
		result, err := processor.EvaluateReconciliation(context.TODO(), client, apiRule)

		// then
		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(1))
		ap := result[0].Obj.(*securityv1beta1.AuthorizationPolicy)

		Expect(ap).NotTo(BeNil())
		Expect(ap.ObjectMeta.Name).To(BeEmpty())
		Expect(ap.ObjectMeta.GenerateName).To(Equal(apiName + "-"))
		Expect(ap.ObjectMeta.Namespace).To(Equal(apiNamespace))
		Expect(ap.ObjectMeta.Labels[testLabelKey]).To(Equal(testLabelValue))

		Expect(ap.Spec.Selector.MatchLabels[testSelectorKey]).NotTo(BeNil())
		Expect(ap.Spec.Selector.MatchLabels[testSelectorKey]).To(Equal(serviceName))
		Expect(len(ap.Spec.Rules)).To(Equal(1))
		Expect(len(ap.Spec.Rules[0].From)).To(Equal(1))
		Expect(len(ap.Spec.Rules[0].From[0].Source.RequestPrincipals)).To(Equal(1))
		Expect(ap.Spec.Rules[0].From[0].Source.RequestPrincipals[0]).To(Equal("*"))
		Expect(len(ap.Spec.Rules[0].To)).To(Equal(1))
		Expect(len(ap.Spec.Rules[0].To[0].Operation.Methods)).To(Equal(1))
		Expect(ap.Spec.Rules[0].To[0].Operation.Methods).To(ContainElements(apiMethods))
		Expect(len(ap.Spec.Rules[0].To[0].Operation.Paths)).To(Equal(1))
		Expect(ap.Spec.Rules[0].To[0].Operation.Paths).To(ContainElements(headersAPIPath))

		Expect(ap.OwnerReferences[0].APIVersion).To(Equal(apiAPIVersion))
		Expect(ap.OwnerReferences[0].Kind).To(Equal(apiKind))
		Expect(ap.OwnerReferences[0].Name).To(Equal(apiName))
		Expect(ap.OwnerReferences[0].UID).To(Equal(apiUID))
	})
})
