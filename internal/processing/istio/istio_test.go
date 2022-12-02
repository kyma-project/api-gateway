package istio

import (
	"fmt"

	"istio.io/api/networking/v1beta1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	"github.com/kyma-incubator/api-gateway/internal/processing"
	rulev1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	apiName                     = "test-apirule"
	apiUID            types.UID = "eab0f1c8-c417-11e9-bf11-4ac644044351"
	apiNamespace                = "some-namespace"
	apiAPIVersion               = "gateway.kyma-project.io/v1alpha1"
	apiKind                     = "ApiRule"
	apiPath                     = "/.*"
	headersAPIPath              = "/headers"
	oauthAPIPath                = "/img"
	jwtIssuer                   = "https://oauth2.example.com/"
	oathkeeperSvc               = "fake.oathkeeper"
	oathkeeperSvcPort uint32    = 1234
	testLabelKey                = "key"
	testLabelValue              = "value"
	defaultDomain               = "myDomain.com"
)

var (
	apiMethods                     = []string{"GET"}
	apiScopes                      = []string{"write", "read"}
	servicePort             uint32 = 8080
	apiGateway                     = "some-gateway"
	serviceName                    = "example-service"
	serviceHostWithNoDomain        = "myService"
	serviceHost                    = serviceHostWithNoDomain + "." + defaultDomain

	testAllowOrigin  = []*v1beta1.StringMatch{{MatchType: &v1beta1.StringMatch_Regex{Regex: ".*"}}}
	testAllowMethods = []string{"GET", "POST", "PUT", "DELETE"}
	testAllowHeaders = []string{"header1", "header2"}

	testCors = &CorsConfig{
		AllowOrigins: testAllowOrigin,
		AllowMethods: testAllowMethods,
		AllowHeaders: testAllowHeaders,
	}

	testAdditionalLabels = map[string]string{testLabelKey: testLabelValue}
)

var _ = Describe("Access Rule Processor", func() {
	Context("when the jwt handler is istio", func() {

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

		It("should produce VS, AP and RA for a rule with one issuer and two paths", func() {

			jwt := createIstioJwtAccessStrategy()
			service := &gatewayv1beta1.Service{
				Name: &serviceName,
				Port: &servicePort,
			}

			ruleJwt := getRuleWithServiceFor(headersAPIPath, apiMethods, []*gatewayv1beta1.Mutator{}, []*gatewayv1beta1.Authenticator{jwt}, service)
			ruleJwt2 := getRuleWithServiceFor(oauthAPIPath, apiMethods, []*gatewayv1beta1.Mutator{}, []*gatewayv1beta1.Authenticator{jwt}, service)
			apiRule := getAPIRuleFor([]gatewayv1beta1.Rule{ruleJwt, ruleJwt2})

			f := NewFactory(nil, ctrl.Log.WithName("test"), oathkeeperSvc, oathkeeperSvcPort, testCors, testAdditionalLabels, defaultDomain)

			desiredState := f.CalculateRequiredState(apiRule, &configIstioJWT)
			vs := desiredState.virtualService
			ap := desiredState.authorizationPolicies
			ra := desiredState.requestAuthentications

			//verify VS
			Expect(vs).NotTo(BeNil())
			Expect(len(vs.Spec.Gateways)).To(Equal(1))
			Expect(len(vs.Spec.Hosts)).To(Equal(1))
			Expect(vs.Spec.Hosts[0]).To(Equal(serviceHost))
			Expect(len(vs.Spec.Http)).To(Equal(2))

			Expect(len(vs.Spec.Http[0].Route)).To(Equal(1))
			Expect(vs.Spec.Http[0].Route[0].Destination.Host).To(Equal(serviceName + "." + apiNamespace + ".svc.cluster.local"))
			Expect(vs.Spec.Http[0].Route[0].Destination.Port.Number).To(Equal(servicePort))
			Expect(len(vs.Spec.Http[0].Match)).To(Equal(1))
			Expect(vs.Spec.Http[0].Match[0].Uri.GetRegex()).To(Equal(apiRule.Spec.Rules[0].Path))

			Expect(vs.Spec.Http[0].CorsPolicy.AllowOrigins).To(Equal(testCors.AllowOrigins))
			Expect(vs.Spec.Http[0].CorsPolicy.AllowMethods).To(Equal(testCors.AllowMethods))
			Expect(vs.Spec.Http[0].CorsPolicy.AllowHeaders).To(Equal(testCors.AllowHeaders))

			Expect(len(vs.Spec.Http[1].Route)).To(Equal(1))
			Expect(vs.Spec.Http[1].Route[0].Destination.Host).To(Equal(serviceName + "." + apiNamespace + ".svc.cluster.local"))
			Expect(vs.Spec.Http[1].Route[0].Destination.Port.Number).To(Equal(servicePort))
			Expect(len(vs.Spec.Http[1].Match)).To(Equal(1))
			Expect(vs.Spec.Http[1].Match[0].Uri.GetRegex()).To(Equal(apiRule.Spec.Rules[1].Path))

			Expect(vs.Spec.Http[1].CorsPolicy.AllowOrigins).To(Equal(testCors.AllowOrigins))
			Expect(vs.Spec.Http[1].CorsPolicy.AllowMethods).To(Equal(testCors.AllowMethods))
			Expect(vs.Spec.Http[1].CorsPolicy.AllowHeaders).To(Equal(testCors.AllowHeaders))

			Expect(vs.ObjectMeta.Name).To(BeEmpty())
			Expect(vs.ObjectMeta.GenerateName).To(Equal(apiName + "-"))
			Expect(vs.ObjectMeta.Namespace).To(Equal(apiNamespace))
			Expect(vs.ObjectMeta.Labels[testLabelKey]).To(Equal(testLabelValue))

			Expect(vs.ObjectMeta.OwnerReferences[0].APIVersion).To(Equal(apiAPIVersion))
			Expect(vs.ObjectMeta.OwnerReferences[0].Kind).To(Equal(apiKind))
			Expect(vs.ObjectMeta.OwnerReferences[0].Name).To(Equal(apiName))
			Expect(vs.ObjectMeta.OwnerReferences[0].UID).To(Equal(apiUID))

			// Verify AP and RA
			Expect(len(ap)).To(Equal(2))
			Expect(len(ra)).To(Equal(1))

			Expect(ap[headersAPIPath]).NotTo(BeNil())
			Expect(ap[headersAPIPath].ObjectMeta.Name).To(BeEmpty())
			Expect(ap[headersAPIPath].ObjectMeta.GenerateName).To(Equal(apiName + "-"))
			Expect(ap[headersAPIPath].ObjectMeta.Namespace).To(Equal(apiNamespace))
			Expect(ap[headersAPIPath].ObjectMeta.Labels[testLabelKey]).To(Equal(testLabelValue))

			Expect(ap[headersAPIPath].Spec.Selector.MatchLabels[testSelectorKey]).NotTo(BeNil())
			Expect(ap[headersAPIPath].Spec.Selector.MatchLabels[testSelectorKey]).To(Equal(serviceName))
			Expect(len(ap[headersAPIPath].Spec.Rules)).To(Equal(1))
			Expect(len(ap[headersAPIPath].Spec.Rules[0].From)).To(Equal(1))
			Expect(len(ap[headersAPIPath].Spec.Rules[0].From[0].Source.RequestPrincipals)).To(Equal(1))
			Expect(ap[headersAPIPath].Spec.Rules[0].From[0].Source.RequestPrincipals[0]).To(Equal("*"))
			Expect(len(ap[headersAPIPath].Spec.Rules[0].To)).To(Equal(1))
			Expect(len(ap[headersAPIPath].Spec.Rules[0].To[0].Operation.Methods)).To(Equal(1))
			Expect(ap[headersAPIPath].Spec.Rules[0].To[0].Operation.Methods).To(ContainElements(apiMethods))
			Expect(len(ap[headersAPIPath].Spec.Rules[0].To[0].Operation.Paths)).To(Equal(1))
			Expect(ap[headersAPIPath].Spec.Rules[0].To[0].Operation.Paths).To(ContainElements(headersAPIPath))

			Expect(ap[headersAPIPath].OwnerReferences[0].APIVersion).To(Equal(apiAPIVersion))
			Expect(ap[headersAPIPath].OwnerReferences[0].Kind).To(Equal(apiKind))
			Expect(ap[headersAPIPath].OwnerReferences[0].Name).To(Equal(apiName))
			Expect(ap[headersAPIPath].OwnerReferences[0].UID).To(Equal(apiUID))

			Expect(ap[oauthAPIPath]).NotTo(BeNil())
			Expect(ap[oauthAPIPath].ObjectMeta.Name).To(BeEmpty())
			Expect(ap[oauthAPIPath].ObjectMeta.GenerateName).To(Equal(apiName + "-"))
			Expect(ap[oauthAPIPath].ObjectMeta.Namespace).To(Equal(apiNamespace))
			Expect(ap[oauthAPIPath].ObjectMeta.Labels[testLabelKey]).To(Equal(testLabelValue))

			Expect(ap[oauthAPIPath].Spec.Selector.MatchLabels[testSelectorKey]).NotTo(BeNil())
			Expect(ap[oauthAPIPath].Spec.Selector.MatchLabels[testSelectorKey]).To(Equal(serviceName))
			Expect(len(ap[oauthAPIPath].Spec.Rules)).To(Equal(1))
			Expect(len(ap[oauthAPIPath].Spec.Rules[0].From)).To(Equal(1))
			Expect(len(ap[oauthAPIPath].Spec.Rules[0].From[0].Source.RequestPrincipals)).To(Equal(1))
			Expect(ap[oauthAPIPath].Spec.Rules[0].From[0].Source.RequestPrincipals[0]).To(Equal("*"))
			Expect(len(ap[oauthAPIPath].Spec.Rules[0].To)).To(Equal(1))
			Expect(len(ap[oauthAPIPath].Spec.Rules[0].To[0].Operation.Methods)).To(Equal(1))
			Expect(ap[oauthAPIPath].Spec.Rules[0].To[0].Operation.Methods).To(ContainElements(apiMethods))
			Expect(len(ap[oauthAPIPath].Spec.Rules[0].To[0].Operation.Paths)).To(Equal(1))
			Expect(ap[oauthAPIPath].Spec.Rules[0].To[0].Operation.Paths).To(ContainElements(oauthAPIPath))

			Expect(len(ap[oauthAPIPath].OwnerReferences)).To(Equal(1))
			Expect(ap[oauthAPIPath].OwnerReferences[0].APIVersion).To(Equal(apiAPIVersion))
			Expect(ap[oauthAPIPath].OwnerReferences[0].Kind).To(Equal(apiKind))
			Expect(ap[oauthAPIPath].OwnerReferences[0].Name).To(Equal(apiName))
			Expect(ap[oauthAPIPath].OwnerReferences[0].UID).To(Equal(apiUID))

			raKey := fmt.Sprintf("%s:%s", jwtIssuer, jwksUri)
			Expect(ra[raKey]).NotTo(BeNil())
			Expect(ra[raKey].ObjectMeta.Name).To(BeEmpty())
			Expect(ra[raKey].ObjectMeta.GenerateName).To(Equal(apiName + "-"))
			Expect(ra[raKey].ObjectMeta.Namespace).To(Equal(apiNamespace))
			Expect(ra[raKey].ObjectMeta.Labels[testLabelKey]).To(Equal(testLabelValue))

			Expect(len(ra[raKey].OwnerReferences)).To(Equal(1))
			Expect(ra[raKey].OwnerReferences[0].APIVersion).To(Equal(apiAPIVersion))
			Expect(ra[raKey].OwnerReferences[0].Kind).To(Equal(apiKind))
			Expect(ra[raKey].OwnerReferences[0].Name).To(Equal(apiName))
			Expect(ra[raKey].OwnerReferences[0].UID).To(Equal(apiUID))

			Expect(ra[raKey].Spec.Selector.MatchLabels[testSelectorKey]).NotTo(BeNil())
			Expect(ra[raKey].Spec.Selector.MatchLabels[testSelectorKey]).To(Equal(serviceName))
			Expect(len(ra[raKey].Spec.JwtRules)).To(Equal(1))
			Expect(ra[raKey].Spec.JwtRules[0].Issuer).To(Equal(jwtIssuer))
			Expect(ra[raKey].Spec.JwtRules[0].JwksUri).To(Equal(jwksUri))
		})

		It("should produce AP and RA for a Rule without service, but service definition on ApiRule level", func() {
			// given
			jwt := createIstioJwtAccessStrategy()

			ruleJwt := getRuleFor(headersAPIPath, apiMethods, []*gatewayv1beta1.Mutator{}, []*gatewayv1beta1.Authenticator{jwt})
			apiRule := getAPIRuleFor([]gatewayv1beta1.Rule{ruleJwt})

			f := NewFactory(nil, ctrl.Log.WithName("test"), oathkeeperSvc, oathkeeperSvcPort, testCors, testAdditionalLabels, defaultDomain)

			// when
			desiredState := f.CalculateRequiredState(apiRule, &configIstioJWT)

			// then
			ap := desiredState.authorizationPolicies
			Expect(ap[headersAPIPath]).NotTo(BeNil())
			Expect(ap[headersAPIPath].Spec.Selector.MatchLabels[testSelectorKey]).To(Equal(serviceName))

			ra := desiredState.requestAuthentications
			raKey := fmt.Sprintf("%s:%s", jwtIssuer, jwksUri)
			Expect(ra[raKey]).NotTo(BeNil())
			Expect(ra[raKey].Spec.Selector.MatchLabels[testSelectorKey]).To(Equal(serviceName))
		})

		It("should produce AP and RA with service from Rule, when service is configured on Rule and ApiRule level", func() {
			// given
			jwt := createIstioJwtAccessStrategy()

			ruleServiceName := "rule-scope-example-service"
			service := &gatewayv1beta1.Service{
				Name: &ruleServiceName,
				Port: &servicePort,
			}

			ruleJwt := getRuleWithServiceFor(headersAPIPath, apiMethods, []*gatewayv1beta1.Mutator{}, []*gatewayv1beta1.Authenticator{jwt}, service)
			apiRule := getAPIRuleFor([]gatewayv1beta1.Rule{ruleJwt})

			f := NewFactory(nil, ctrl.Log.WithName("test"), oathkeeperSvc, oathkeeperSvcPort, testCors, testAdditionalLabels, defaultDomain)

			// when
			desiredState := f.CalculateRequiredState(apiRule, &configIstioJWT)

			// then
			ap := desiredState.authorizationPolicies
			Expect(ap[headersAPIPath]).NotTo(BeNil())
			Expect(ap[headersAPIPath].Spec.Selector.MatchLabels[testSelectorKey]).To(Equal(ruleServiceName))

			ra := desiredState.requestAuthentications
			raKey := fmt.Sprintf("%s:%s", jwtIssuer, jwksUri)
			Expect(ra[raKey]).NotTo(BeNil())
			Expect(ra[raKey].Spec.Selector.MatchLabels[testSelectorKey]).To(Equal(ruleServiceName))
		})

		It("should produce VS, AP and RA from a rule with two issuers and one path", func() {
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

			service := &gatewayv1beta1.Service{
				Name: &serviceName,
				Port: &servicePort,
			}

			ruleJwt := getRuleWithServiceFor(headersAPIPath, apiMethods, []*gatewayv1beta1.Mutator{}, []*gatewayv1beta1.Authenticator{jwt}, service)
			apiRule := getAPIRuleFor([]gatewayv1beta1.Rule{ruleJwt})

			f := NewFactory(nil, ctrl.Log.WithName("test"), oathkeeperSvc, oathkeeperSvcPort, testCors, testAdditionalLabels, defaultDomain)

			desiredState := f.CalculateRequiredState(apiRule, &configIstioJWT)
			vs := desiredState.virtualService
			ap := desiredState.authorizationPolicies
			ra := desiredState.requestAuthentications

			//verify VS
			Expect(vs).NotTo(BeNil())
			Expect(len(vs.Spec.Gateways)).To(Equal(1))
			Expect(len(vs.Spec.Hosts)).To(Equal(1))
			Expect(vs.Spec.Hosts[0]).To(Equal(serviceHost))
			Expect(len(vs.Spec.Http)).To(Equal(1))

			Expect(len(vs.Spec.Http[0].Route)).To(Equal(1))
			Expect(vs.Spec.Http[0].Route[0].Destination.Host).To(Equal(serviceName + "." + apiNamespace + ".svc.cluster.local"))
			Expect(vs.Spec.Http[0].Route[0].Destination.Port.Number).To(Equal(servicePort))
			Expect(len(vs.Spec.Http[0].Match)).To(Equal(1))
			Expect(vs.Spec.Http[0].Match[0].Uri.GetRegex()).To(Equal(apiRule.Spec.Rules[0].Path))

			Expect(vs.Spec.Http[0].CorsPolicy.AllowOrigins).To(Equal(testCors.AllowOrigins))
			Expect(vs.Spec.Http[0].CorsPolicy.AllowMethods).To(Equal(testCors.AllowMethods))
			Expect(vs.Spec.Http[0].CorsPolicy.AllowHeaders).To(Equal(testCors.AllowHeaders))

			Expect(vs.ObjectMeta.Name).To(BeEmpty())
			Expect(vs.ObjectMeta.GenerateName).To(Equal(apiName + "-"))
			Expect(vs.ObjectMeta.Namespace).To(Equal(apiNamespace))
			Expect(vs.ObjectMeta.Labels[testLabelKey]).To(Equal(testLabelValue))

			Expect(vs.ObjectMeta.OwnerReferences[0].APIVersion).To(Equal(apiAPIVersion))
			Expect(vs.ObjectMeta.OwnerReferences[0].Kind).To(Equal(apiKind))
			Expect(vs.ObjectMeta.OwnerReferences[0].Name).To(Equal(apiName))
			Expect(vs.ObjectMeta.OwnerReferences[0].UID).To(Equal(apiUID))

			// Verify AP and RA
			Expect(len(ap)).To(Equal(1))
			Expect(len(ra)).To(Equal(1))

			Expect(ap[headersAPIPath]).NotTo(BeNil())
			Expect(ap[headersAPIPath].ObjectMeta.Name).To(BeEmpty())
			Expect(ap[headersAPIPath].ObjectMeta.GenerateName).To(Equal(apiName + "-"))
			Expect(ap[headersAPIPath].ObjectMeta.Namespace).To(Equal(apiNamespace))
			Expect(ap[headersAPIPath].ObjectMeta.Labels[testLabelKey]).To(Equal(testLabelValue))

			Expect(ap[headersAPIPath].Spec.Selector.MatchLabels[testSelectorKey]).NotTo(BeNil())
			Expect(ap[headersAPIPath].Spec.Selector.MatchLabels[testSelectorKey]).To(Equal(serviceName))
			Expect(len(ap[headersAPIPath].Spec.Rules)).To(Equal(1))
			Expect(len(ap[headersAPIPath].Spec.Rules[0].From)).To(Equal(1))
			Expect(len(ap[headersAPIPath].Spec.Rules[0].From[0].Source.RequestPrincipals)).To(Equal(1))
			Expect(ap[headersAPIPath].Spec.Rules[0].From[0].Source.RequestPrincipals[0]).To(Equal("*"))
			Expect(len(ap[headersAPIPath].Spec.Rules[0].To)).To(Equal(1))
			Expect(len(ap[headersAPIPath].Spec.Rules[0].To[0].Operation.Methods)).To(Equal(1))
			Expect(ap[headersAPIPath].Spec.Rules[0].To[0].Operation.Methods).To(ContainElements(apiMethods))
			Expect(len(ap[headersAPIPath].Spec.Rules[0].To[0].Operation.Paths)).To(Equal(1))
			Expect(ap[headersAPIPath].Spec.Rules[0].To[0].Operation.Paths).To(ContainElements(headersAPIPath))

			Expect(ap[headersAPIPath].OwnerReferences[0].APIVersion).To(Equal(apiAPIVersion))
			Expect(ap[headersAPIPath].OwnerReferences[0].Kind).To(Equal(apiKind))
			Expect(ap[headersAPIPath].OwnerReferences[0].Name).To(Equal(apiName))
			Expect(ap[headersAPIPath].OwnerReferences[0].UID).To(Equal(apiUID))

			raKey := fmt.Sprintf("%s:%s%s:%s", jwtIssuer, jwksUri, jwtIssuer2, jwksUri2)
			Expect(ra[raKey]).NotTo(BeNil())
			Expect(ra[raKey].ObjectMeta.Name).To(BeEmpty())
			Expect(ra[raKey].ObjectMeta.GenerateName).To(Equal(apiName + "-"))
			Expect(ra[raKey].ObjectMeta.Namespace).To(Equal(apiNamespace))
			Expect(ra[raKey].ObjectMeta.Labels[testLabelKey]).To(Equal(testLabelValue))

			Expect(len(ra[raKey].OwnerReferences)).To(Equal(1))
			Expect(ra[raKey].OwnerReferences[0].APIVersion).To(Equal(apiAPIVersion))
			Expect(ra[raKey].OwnerReferences[0].Kind).To(Equal(apiKind))
			Expect(ra[raKey].OwnerReferences[0].Name).To(Equal(apiName))
			Expect(ra[raKey].OwnerReferences[0].UID).To(Equal(apiUID))

			Expect(ra[raKey].Spec.Selector.MatchLabels[testSelectorKey]).NotTo(BeNil())
			Expect(ra[raKey].Spec.Selector.MatchLabels[testSelectorKey]).To(Equal(serviceName))
			Expect(len(ra[raKey].Spec.JwtRules)).To(Equal(2))
			Expect(ra[raKey].Spec.JwtRules[0].Issuer).To(Equal(jwtIssuer))
			Expect(ra[raKey].Spec.JwtRules[0].JwksUri).To(Equal(jwksUri))
			Expect(ra[raKey].Spec.JwtRules[1].Issuer).To(Equal(jwtIssuer2))
			Expect(ra[raKey].Spec.JwtRules[1].JwksUri).To(Equal(jwksUri2))
		})
	})
})

type mockCreator struct {
	createMock func() map[string]*rulev1alpha1.Rule
}

func (r mockCreator) Create(_ *gatewayv1beta1.APIRule) map[string]*rulev1alpha1.Rule {
	return r.createMock()
}

var actionToString = func(a processing.Action) string { return a.String() }

func getRuleFor(path string, methods []string, mutators []*gatewayv1beta1.Mutator, accessStrategies []*gatewayv1beta1.Authenticator) gatewayv1beta1.Rule {
	return gatewayv1beta1.Rule{
		Path:             path,
		Methods:          methods,
		Mutators:         mutators,
		AccessStrategies: accessStrategies,
	}
}

func getRuleWithServiceFor(path string, methods []string, mutators []*gatewayv1beta1.Mutator, accessStrategies []*gatewayv1beta1.Authenticator, service *gatewayv1beta1.Service) gatewayv1beta1.Rule {
	return gatewayv1beta1.Rule{
		Path:             path,
		Methods:          methods,
		Mutators:         mutators,
		AccessStrategies: accessStrategies,
		Service:          service,
	}
}

func getAPIRuleFor(rules []gatewayv1beta1.Rule) *gatewayv1beta1.APIRule {
	return &gatewayv1beta1.APIRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      apiName,
			UID:       apiUID,
			Namespace: apiNamespace,
		},
		TypeMeta: metav1.TypeMeta{
			APIVersion: apiAPIVersion,
			Kind:       apiKind,
		},
		Spec: gatewayv1beta1.APIRuleSpec{
			Gateway: &apiGateway,
			Service: &gatewayv1beta1.Service{
				Name: &serviceName,
				Port: &servicePort,
			},
			Host:  &serviceHost,
			Rules: rules,
		},
	}
}

func toCSVList(input []string) string {
	if len(input) == 0 {
		return ""
	}

	res := `"` + input[0] + `"`

	for i := 1; i < len(input); i++ {
		res = res + "," + `"` + input[i] + `"`
	}

	return res
}
