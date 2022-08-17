package processing

import (
	"fmt"
	"testing"

	"istio.io/api/networking/v1beta1"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
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
	expectedCorsPolicy = v1beta1.CorsPolicy{
		AllowOrigins: testAllowOrigin,
		AllowMethods: testAllowMethods,
		AllowHeaders: testAllowHeaders,
	}

	testAdditionalLabels = map[string]string{testLabelKey: testLabelValue}
)

func TestProcessing(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Processing Suite")
}

var _ = Describe("Factory", func() {
	Describe("CalculateRequiredState", func() {
		Context("APIRule", func() {
			It("should produce VS for allow authenticator", func() {
				strategies := []*gatewayv1beta1.Authenticator{
					{
						Handler: &gatewayv1beta1.Handler{
							Name: "allow",
						},
					},
				}

				allowRule := getRuleFor(apiPath, apiMethods, []*gatewayv1beta1.Mutator{}, strategies)
				rules := []gatewayv1beta1.Rule{allowRule}

				apiRule := getAPIRuleFor(rules)

				f := NewFactory(nil, ctrl.Log.WithName("test"), oathkeeperSvc, oathkeeperSvcPort, "https://example.com/.well-known/jwks.json", testCors, testAdditionalLabels, defaultDomain)

				desiredState := f.CalculateRequiredState(apiRule)
				vs := desiredState.virtualService
				accessRules := desiredState.accessRules

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

				//Verify AR
				Expect(len(accessRules)).To(Equal(0))
			})

			It("noop: should override access rule upstream with rule level service", func() {
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

				allowRule := getRuleWithServiceFor(apiPath, apiMethods, []*gatewayv1beta1.Mutator{}, strategies, service)
				rules := []gatewayv1beta1.Rule{allowRule}

				apiRule := getAPIRuleFor(rules)

				f := NewFactory(nil, ctrl.Log.WithName("test"), oathkeeperSvc, oathkeeperSvcPort, "https://example.com/.well-known/jwks.json", testCors, testAdditionalLabels, defaultDomain)

				desiredState := f.CalculateRequiredState(apiRule)
				vs := desiredState.virtualService
				accessRules := desiredState.accessRules

				expectedNoopRuleMatchURL := fmt.Sprintf("<http|https>://%s<%s>", serviceHost, apiPath)
				expectedRuleUpstreamURL := fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", overrideServiceName, apiNamespace, overrideServicePort)

				//Verify VS
				Expect(len(vs.Spec.Http[0].Route)).To(Equal(1))
				Expect(vs.Spec.Http[0].Route[0].Destination.Host).To(Equal(oathkeeperSvc))

				//Verify AR has rule level upstream
				Expect(len(accessRules)).To(Equal(1))
				Expect(accessRules[expectedNoopRuleMatchURL].Spec.Upstream.URL).To(Equal(expectedRuleUpstreamURL))
			})

			It("allow: should override VS destination host", func() {
				strategies := []*gatewayv1beta1.Authenticator{
					{
						Handler: &gatewayv1beta1.Handler{
							Name: "allow",
						},
					},
				}

				overrideServiceName := "testName"
				overrideServicePort := uint32(8080)

				service := &gatewayv1beta1.Service{
					Name: &overrideServiceName,
					Port: &overrideServicePort,
				}

				allowRule := getRuleWithServiceFor(apiPath, apiMethods, []*gatewayv1beta1.Mutator{}, strategies, service)
				rules := []gatewayv1beta1.Rule{allowRule}

				apiRule := getAPIRuleFor(rules)

				f := NewFactory(nil, ctrl.Log.WithName("test"), oathkeeperSvc, oathkeeperSvcPort, "https://example.com/.well-known/jwks.json", testCors, testAdditionalLabels, defaultDomain)

				desiredState := f.CalculateRequiredState(apiRule)
				vs := desiredState.virtualService
				accessRules := desiredState.accessRules

				//verify VS has rule level destination host
				Expect(len(vs.Spec.Http[0].Route)).To(Equal(1))
				Expect(vs.Spec.Http[0].Route[0].Destination.Host).To(Equal(overrideServiceName + "." + apiNamespace + ".svc.cluster.local"))

				//Verify AR has rule level upstream
				Expect(len(accessRules)).To(Equal(0))
			})

			It("should produce VS and ARs for given paths", func() {
				noop := []*gatewayv1beta1.Authenticator{
					{
						Handler: &gatewayv1beta1.Handler{
							Name: "noop",
						},
					},
				}

				jwtConfigJSON := fmt.Sprintf(`
					{
						"trusted_issuers": ["%s"],
						"jwks": [],
						"required_scope": [%s]
				}`, jwtIssuer, toCSVList(apiScopes))

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

				testMutators := []*gatewayv1beta1.Mutator{
					{
						Handler: &gatewayv1beta1.Handler{
							Name: "noop",
						},
					},
					{
						Handler: &gatewayv1beta1.Handler{
							Name: "idtoken",
						},
					},
				}

				noopRule := getRuleFor(apiPath, apiMethods, []*gatewayv1beta1.Mutator{}, noop)
				jwtRule := getRuleFor(headersAPIPath, apiMethods, testMutators, jwt)
				rules := []gatewayv1beta1.Rule{noopRule, jwtRule}

				expectedNoopRuleMatchURL := fmt.Sprintf("<http|https>://%s<%s>", serviceHost, apiPath)
				expectedJwtRuleMatchURL := fmt.Sprintf("<http|https>://%s<%s>", serviceHost, headersAPIPath)
				expectedRuleUpstreamURL := fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", serviceName, apiNamespace, servicePort)

				apiRule := getAPIRuleFor(rules)

				f := NewFactory(nil, ctrl.Log.WithName("test"), oathkeeperSvc, oathkeeperSvcPort, "https://example.com/.well-known/jwks.json", testCors, testAdditionalLabels, defaultDomain)

				desiredState := f.CalculateRequiredState(apiRule)
				vs := desiredState.virtualService
				accessRules := desiredState.accessRules

				//verify VS
				Expect(vs).NotTo(BeNil())
				Expect(len(vs.Spec.Gateways)).To(Equal(1))
				Expect(len(vs.Spec.Hosts)).To(Equal(1))
				Expect(vs.Spec.Hosts[0]).To(Equal(serviceHost))
				Expect(len(vs.Spec.Http)).To(Equal(2))

				Expect(len(vs.Spec.Http[0].Route)).To(Equal(1))
				Expect(vs.Spec.Http[0].Route[0].Destination.Host).To(Equal(oathkeeperSvc))
				Expect(vs.Spec.Http[0].Route[0].Destination.Port.Number).To(Equal(oathkeeperSvcPort))
				Expect(len(vs.Spec.Http[0].Match)).To(Equal(1))
				Expect(vs.Spec.Http[0].Match[0].Uri.GetRegex()).To(Equal(apiRule.Spec.Rules[0].Path))

				Expect(vs.Spec.Http[0].CorsPolicy.AllowOrigins).To(Equal(testCors.AllowOrigins))
				Expect(vs.Spec.Http[0].CorsPolicy.AllowMethods).To(Equal(testCors.AllowMethods))
				Expect(vs.Spec.Http[0].CorsPolicy.AllowHeaders).To(Equal(testCors.AllowHeaders))

				Expect(len(vs.Spec.Http[1].Route)).To(Equal(1))
				Expect(vs.Spec.Http[1].Route[0].Destination.Host).To(Equal(oathkeeperSvc))
				Expect(vs.Spec.Http[1].Route[0].Destination.Port.Number).To(Equal(oathkeeperSvcPort))
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

				//Verify ARs
				Expect(len(accessRules)).To(Equal(2))

				noopAccessRule := accessRules[expectedNoopRuleMatchURL]

				Expect(len(accessRules)).To(Equal(2))
				Expect(len(noopAccessRule.Spec.Authenticators)).To(Equal(1))

				Expect(noopAccessRule.Spec.Authorizer.Name).To(Equal("allow"))
				Expect(noopAccessRule.Spec.Authorizer.Config).To(BeNil())

				Expect(noopAccessRule.Spec.Authenticators[0].Handler.Name).To(Equal("noop"))
				Expect(noopAccessRule.Spec.Authenticators[0].Handler.Config).To(BeNil())

				Expect(len(noopAccessRule.Spec.Match.Methods)).To(Equal(len(apiMethods)))
				Expect(noopAccessRule.Spec.Match.Methods).To(Equal(apiMethods))
				Expect(noopAccessRule.Spec.Match.URL).To(Equal(expectedNoopRuleMatchURL))

				Expect(noopAccessRule.Spec.Upstream.URL).To(Equal(expectedRuleUpstreamURL))

				Expect(noopAccessRule.ObjectMeta.Name).To(BeEmpty())
				Expect(noopAccessRule.ObjectMeta.GenerateName).To(Equal(apiName + "-"))
				Expect(noopAccessRule.ObjectMeta.Namespace).To(Equal(apiNamespace))
				Expect(noopAccessRule.ObjectMeta.Labels[testLabelKey]).To(Equal(testLabelValue))

				Expect(noopAccessRule.ObjectMeta.OwnerReferences[0].APIVersion).To(Equal(apiAPIVersion))
				Expect(noopAccessRule.ObjectMeta.OwnerReferences[0].Kind).To(Equal(apiKind))
				Expect(noopAccessRule.ObjectMeta.OwnerReferences[0].Name).To(Equal(apiName))
				Expect(noopAccessRule.ObjectMeta.OwnerReferences[0].UID).To(Equal(apiUID))

				jwtAccessRule := accessRules[expectedJwtRuleMatchURL]

				Expect(len(jwtAccessRule.Spec.Authenticators)).To(Equal(1))

				Expect(jwtAccessRule.Spec.Authorizer.Name).To(Equal("allow"))
				Expect(jwtAccessRule.Spec.Authorizer.Config).To(BeNil())

				Expect(jwtAccessRule.Spec.Authenticators[0].Handler.Name).To(Equal("jwt"))
				Expect(jwtAccessRule.Spec.Authenticators[0].Handler.Config).NotTo(BeNil())
				Expect(string(jwtAccessRule.Spec.Authenticators[0].Handler.Config.Raw)).To(Equal(jwtConfigJSON))

				Expect(len(jwtAccessRule.Spec.Match.Methods)).To(Equal(len(apiMethods)))
				Expect(jwtAccessRule.Spec.Match.Methods).To(Equal(apiMethods))
				Expect(jwtAccessRule.Spec.Match.URL).To(Equal(expectedJwtRuleMatchURL))

				Expect(jwtAccessRule.Spec.Upstream.URL).To(Equal(expectedRuleUpstreamURL))

				Expect(jwtAccessRule.Spec.Mutators).NotTo(BeNil())
				Expect(len(jwtAccessRule.Spec.Mutators)).To(Equal(len(testMutators)))
				Expect(jwtAccessRule.Spec.Mutators[0].Handler.Name).To(Equal(testMutators[0].Name))
				Expect(jwtAccessRule.Spec.Mutators[1].Handler.Name).To(Equal(testMutators[1].Name))

				Expect(jwtAccessRule.ObjectMeta.Name).To(BeEmpty())
				Expect(jwtAccessRule.ObjectMeta.GenerateName).To(Equal(apiName + "-"))
				Expect(jwtAccessRule.ObjectMeta.Namespace).To(Equal(apiNamespace))
				Expect(jwtAccessRule.ObjectMeta.Labels[testLabelKey]).To(Equal(testLabelValue))

				Expect(jwtAccessRule.ObjectMeta.OwnerReferences[0].APIVersion).To(Equal(apiAPIVersion))
				Expect(jwtAccessRule.ObjectMeta.OwnerReferences[0].Kind).To(Equal(apiKind))
				Expect(jwtAccessRule.ObjectMeta.OwnerReferences[0].Name).To(Equal(apiName))
				Expect(jwtAccessRule.ObjectMeta.OwnerReferences[0].UID).To(Equal(apiUID))
			})

			It("should produce VS & AR for jwt & oauth authenticators for given path", func() {
				oauthConfigJSON := fmt.Sprintf(`{"required_scope": [%s]}`, toCSVList(apiScopes))

				jwtConfigJSON := fmt.Sprintf(`
					{
						"trusted_issuers": ["%s"],
						"jwks": [],
						"required_scope": [%s]
				}`, jwtIssuer, toCSVList(apiScopes))

				jwt := &gatewayv1beta1.Authenticator{
					Handler: &gatewayv1beta1.Handler{
						Name: "jwt",
						Config: &runtime.RawExtension{
							Raw: []byte(jwtConfigJSON),
						},
					},
				}
				oauth := &gatewayv1beta1.Authenticator{
					Handler: &gatewayv1beta1.Handler{
						Name: "oauth2_introspection",
						Config: &runtime.RawExtension{
							Raw: []byte(oauthConfigJSON),
						},
					},
				}

				strategies := []*gatewayv1beta1.Authenticator{jwt, oauth}

				allowRule := getRuleFor(apiPath, apiMethods, []*gatewayv1beta1.Mutator{}, strategies)
				rules := []gatewayv1beta1.Rule{allowRule}

				expectedRuleMatchURL := fmt.Sprintf("<http|https>://%s<%s>", serviceHost, apiPath)
				expectedRuleUpstreamURL := fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", serviceName, apiNamespace, servicePort)

				apiRule := getAPIRuleFor(rules)

				f := NewFactory(nil, ctrl.Log.WithName("test"), oathkeeperSvc, oathkeeperSvcPort, "https://example.com/.well-known/jwks.json", testCors, testAdditionalLabels, defaultDomain)

				desiredState := f.CalculateRequiredState(apiRule)
				vs := desiredState.virtualService
				accessRules := desiredState.accessRules

				//verify VS
				Expect(vs).NotTo(BeNil())
				Expect(len(vs.Spec.Gateways)).To(Equal(1))
				Expect(len(vs.Spec.Hosts)).To(Equal(1))
				Expect(vs.Spec.Hosts[0]).To(Equal(serviceHost))
				Expect(len(vs.Spec.Http)).To(Equal(1))

				Expect(len(vs.Spec.Http[0].Route)).To(Equal(1))
				Expect(vs.Spec.Http[0].Route[0].Destination.Host).To(Equal(oathkeeperSvc))
				Expect(vs.Spec.Http[0].Route[0].Destination.Port.Number).To(Equal(oathkeeperSvcPort))

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

				rule := accessRules[expectedRuleMatchURL]

				//Verify AR
				Expect(len(accessRules)).To(Equal(1))
				Expect(len(rule.Spec.Authenticators)).To(Equal(2))

				Expect(rule.Spec.Authorizer.Name).To(Equal("allow"))
				Expect(rule.Spec.Authorizer.Config).To(BeNil())

				Expect(rule.Spec.Authenticators[0].Handler.Name).To(Equal("jwt"))
				Expect(rule.Spec.Authenticators[0].Handler.Config).NotTo(BeNil())
				Expect(string(rule.Spec.Authenticators[0].Handler.Config.Raw)).To(Equal(jwtConfigJSON))

				Expect(rule.Spec.Authenticators[1].Handler.Name).To(Equal("oauth2_introspection"))
				Expect(rule.Spec.Authenticators[1].Handler.Config).NotTo(BeNil())
				Expect(string(rule.Spec.Authenticators[1].Handler.Config.Raw)).To(Equal(oauthConfigJSON))

				Expect(len(rule.Spec.Match.Methods)).To(Equal(len(apiMethods)))
				Expect(rule.Spec.Match.Methods).To(Equal(apiMethods))
				Expect(rule.Spec.Match.URL).To(Equal(expectedRuleMatchURL))

				Expect(rule.Spec.Upstream.URL).To(Equal(expectedRuleUpstreamURL))

				Expect(rule.ObjectMeta.Name).To(BeEmpty())
				Expect(rule.ObjectMeta.GenerateName).To(Equal(apiName + "-"))
				Expect(rule.ObjectMeta.Namespace).To(Equal(apiNamespace))
				Expect(rule.ObjectMeta.Labels[testLabelKey]).To(Equal(testLabelValue))

				Expect(rule.ObjectMeta.OwnerReferences[0].APIVersion).To(Equal(apiAPIVersion))
				Expect(rule.ObjectMeta.OwnerReferences[0].Kind).To(Equal(apiKind))
				Expect(rule.ObjectMeta.OwnerReferences[0].Name).To(Equal(apiName))
				Expect(rule.ObjectMeta.OwnerReferences[0].UID).To(Equal(apiUID))

			})

			Context("when the hostname does not contain domain name", func() {
				It("should produce VS & AR with default domain name", func() {
					noop := []*gatewayv1beta1.Authenticator{
						{
							Handler: &gatewayv1beta1.Handler{
								Name: "noop",
							},
						},
					}

					jwtConfigJSON := fmt.Sprintf(`
					{
						"trusted_issuers": ["%s"],
						"jwks": [],
						"required_scope": [%s]
				}`, jwtIssuer, toCSVList(apiScopes))

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

					testMutators := []*gatewayv1beta1.Mutator{
						{
							Handler: &gatewayv1beta1.Handler{
								Name: "noop",
							},
						},
						{
							Handler: &gatewayv1beta1.Handler{
								Name: "idtoken",
							},
						},
					}

					noopRule := getRuleFor(apiPath, apiMethods, []*gatewayv1beta1.Mutator{}, noop)
					jwtRule := getRuleFor(headersAPIPath, apiMethods, testMutators, jwt)
					rules := []gatewayv1beta1.Rule{noopRule, jwtRule}

					expectedNoopRuleMatchURL := fmt.Sprintf("<http|https>://%s<%s>", serviceHost, apiPath)
					expectedJwtRuleMatchURL := fmt.Sprintf("<http|https>://%s<%s>", serviceHost, headersAPIPath)

					apiRule := getAPIRuleFor(rules)
					apiRule.Spec.Host = &serviceHostWithNoDomain

					f := NewFactory(nil, ctrl.Log.WithName("test"), oathkeeperSvc, oathkeeperSvcPort, "https://example.com/.well-known/jwks.json", testCors, testAdditionalLabels, defaultDomain)

					desiredState := f.CalculateRequiredState(apiRule)
					vs := desiredState.virtualService
					accessRules := desiredState.accessRules

					//verify VS
					Expect(vs).NotTo(BeNil())
					Expect(len(vs.Spec.Hosts)).To(Equal(1))
					Expect(vs.Spec.Hosts[0]).To(Equal(serviceHost))

					//Verify ARs
					Expect(len(accessRules)).To(Equal(2))
					noopAccessRule := accessRules[expectedNoopRuleMatchURL]
					Expect(noopAccessRule.Spec.Match.URL).To(Equal(expectedNoopRuleMatchURL))
					jwtAccessRule := accessRules[expectedJwtRuleMatchURL]
					Expect(jwtAccessRule.Spec.Match.URL).To(Equal(expectedJwtRuleMatchURL))
				})
			})
		})
	})

	Describe("CalculateDiff", func() {
		Context("between desired state & actual state", func() {
			It("should produce patch containing VS to create & AR to create", func() {
				noop := []*gatewayv1beta1.Authenticator{
					{
						Handler: &gatewayv1beta1.Handler{
							Name: "noop",
						},
					},
				}

				noopRule := getRuleFor(apiPath, apiMethods, []*gatewayv1beta1.Mutator{}, noop)
				rules := []gatewayv1beta1.Rule{noopRule}

				apiRule := getAPIRuleFor(rules)
				expectedNoopRuleMatchURL := fmt.Sprintf("<http|https>://%s<%s>", serviceHost, apiPath)

				f := NewFactory(nil, ctrl.Log.WithName("test"), oathkeeperSvc, oathkeeperSvcPort, "https://example.com/.well-known/jwks.json", testCors, testAdditionalLabels, defaultDomain)

				desiredState := f.CalculateRequiredState(apiRule)
				actualState := &State{}

				patch := f.CalculateDiff(desiredState, actualState)

				//Verify patch
				Expect(patch.virtualService).NotTo(BeNil())
				Expect(patch.virtualService.action).To(Equal("create"))
				Expect(patch.virtualService.obj).To(Equal(desiredState.virtualService))

				Expect(patch.accessRule).NotTo(BeNil())
				Expect(len(patch.accessRule)).To(Equal(len(desiredState.accessRules)))
				Expect(patch.accessRule[expectedNoopRuleMatchURL].action).To(Equal("create"))
				Expect(patch.accessRule[expectedNoopRuleMatchURL].obj).To(Equal(desiredState.accessRules[expectedNoopRuleMatchURL]))

			})

			It("should produce patch containing VS to update, AR to create, AR to update & AR to delete", func() {
				oauthConfigJSON := fmt.Sprintf(`{"required_scope": [%s]}`, toCSVList(apiScopes))
				oauth := &gatewayv1beta1.Authenticator{
					Handler: &gatewayv1beta1.Handler{
						Name: "oauth2_introspection",
						Config: &runtime.RawExtension{
							Raw: []byte(oauthConfigJSON),
						},
					},
				}

				strategies := []*gatewayv1beta1.Authenticator{oauth}

				noop := []*gatewayv1beta1.Authenticator{
					{
						Handler: &gatewayv1beta1.Handler{
							Name: "noop",
						},
					},
				}

				noopRule := getRuleFor(headersAPIPath, apiMethods, []*gatewayv1beta1.Mutator{}, noop)
				allowRule := getRuleFor(oauthAPIPath, apiMethods, []*gatewayv1beta1.Mutator{}, strategies)

				rules := []gatewayv1beta1.Rule{noopRule, allowRule}

				apiRule := getAPIRuleFor(rules)

				f := NewFactory(nil, ctrl.Log.WithName("test"), oathkeeperSvc, oathkeeperSvcPort, "https://example.com/.well-known/jwks.json", testCors, testAdditionalLabels, defaultDomain)

				desiredState := f.CalculateRequiredState(apiRule)
				oauthNoopRuleMatchURL := fmt.Sprintf("<http|https>://%s<%s>", serviceHost, oauthAPIPath)
				expectedNoopRuleMatchURL := fmt.Sprintf("<http|https>://%s<%s>", serviceHost, headersAPIPath)
				notDesiredRuleMatchURL := fmt.Sprintf("<http|https>://%s<%s>", serviceHost, "/delete")

				labels := make(map[string]string)
				labels["myLabel"] = "should not override"

				vs := &networkingv1beta1.VirtualService{
					ObjectMeta: metav1.ObjectMeta{
						GenerateName: apiName + "-",
						Labels:       labels,
					},
				}

				noopExistingRule := &rulev1alpha1.Rule{
					ObjectMeta: metav1.ObjectMeta{
						GenerateName: apiName + "-",
						Labels:       labels,
					},
					Spec: rulev1alpha1.RuleSpec{
						Match: &rulev1alpha1.Match{
							URL: expectedNoopRuleMatchURL,
						},
					},
				}

				deleteExistingRule := &rulev1alpha1.Rule{
					ObjectMeta: metav1.ObjectMeta{
						GenerateName: apiName + "-",
						Labels:       labels,
					},
					Spec: rulev1alpha1.RuleSpec{
						Match: &rulev1alpha1.Match{
							URL: notDesiredRuleMatchURL,
						},
					},
				}

				accessRules := make(map[string]*rulev1alpha1.Rule)
				accessRules[expectedNoopRuleMatchURL] = noopExistingRule
				accessRules[notDesiredRuleMatchURL] = deleteExistingRule

				actualState := &State{virtualService: vs, accessRules: accessRules}

				patch := f.CalculateDiff(desiredState, actualState)
				vsPatch := patch.virtualService.obj.(*networkingv1beta1.VirtualService)

				//Verify patch
				Expect(patch.virtualService).NotTo(BeNil())
				Expect(patch.virtualService.action).To(Equal("update"))
				Expect(vsPatch.ObjectMeta.Labels).To(Equal(actualState.virtualService.ObjectMeta.Labels))

				//TODO verify vs spec

				Expect(len(patch.accessRule)).To(Equal(3))

				noopPatchRule := patch.accessRule[expectedNoopRuleMatchURL]
				Expect(noopPatchRule).NotTo(BeNil())
				Expect(noopPatchRule.action).To(Equal("update"))

				//TODO verify ar spec

				notDesiredPatchRule := patch.accessRule[notDesiredRuleMatchURL]
				Expect(notDesiredPatchRule).NotTo(BeNil())
				Expect(notDesiredPatchRule.action).To(Equal("delete"))

				oauthPatchRule := patch.accessRule[oauthNoopRuleMatchURL]
				Expect(oauthPatchRule).NotTo(BeNil())
				Expect(oauthPatchRule.action).To(Equal("create"))

				//TODO verify ar spec
			})
		})
	})
})

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
