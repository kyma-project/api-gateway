package gateway_test

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	gatewayv2 "github.com/kyma-project/api-gateway/apis/gateway/v2"
	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	apinetworkingv1beta1 "istio.io/api/networking/v1beta1"

	gomegatypes "github.com/onsi/gomega/types"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/utils/ptr"

	"encoding/json"

	"github.com/kyma-project/api-gateway/internal/builders"
	"github.com/kyma-project/api-gateway/internal/helpers"
	"github.com/kyma-project/api-gateway/internal/processing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	rulev1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Tests needs to be executed serially because of the shared state of the JWT Handler in the API Controller.
var _ = Describe("APIRule Controller", Serial, func() {
	const (
		testNameBase               = "test"
		testIDLength               = 5
		testServiceNameBase        = "httpbin"
		testServicePort     uint32 = 443
		testPath                   = "/.*"
		testIssuer                 = "https://oauth2.example.com/"
		testJwksUri                = "https://oauth2.example.com/.well-known/jwks.json"
		defaultHttpTimeout         = time.Second * 180
	)

	var methodsGet = []gatewayv1beta1.HttpMethod{http.MethodGet}
	var methodsPut = []gatewayv1beta1.HttpMethod{http.MethodPut}
	var methodsDelete = []gatewayv1beta1.HttpMethod{http.MethodDelete}
	var methodsPost = []gatewayv1beta1.HttpMethod{http.MethodPost}
	var v2alpha1methodsGet = []gatewayv2alpha1.HttpMethod{http.MethodGet}
	var v2methodsGet = []gatewayv2.HttpMethod{http.MethodGet}

	Context("check default domain logic", func() {
		It("should have an error when creating an APIRule without a domain in cluster without kyma-gateway", func() {
			updateJwtHandlerTo(helpers.JWT_HANDLER_ISTIO)

			rule1 := testRule("/rule1", methodsGet, defaultMutators, noConfigHandler("allow"))

			apiRuleName := generateTestName(testNameBase, testIDLength)
			serviceName := generateTestName(testServiceNameBase, testIDLength)
			serviceHost := serviceName

			By("Creating APIRule")

			apiRule := testApiRule(apiRuleName, testNamespace, serviceName, testNamespace, serviceHost, testServicePort, []gatewayv1beta1.Rule{rule1})
			svc := testService(serviceName, testNamespace, testServicePort)
			defer func() {
				deleteResource(apiRule)
				deleteResource(svc)
			}()

			// when
			Expect(c.Create(context.Background(), svc)).Should(Succeed())
			Expect(c.Create(context.Background(), apiRule)).Should(Succeed())

			expectApiRuleStatus(apiRuleName, gatewayv1beta1.StatusError)

			By("Setting a full host for the APIRule should resolve the error")

			By("Updating APIRule")
			existingInstance := gatewayv1beta1.APIRule{}
			Expect(c.Get(context.Background(), client.ObjectKey{Name: apiRuleName, Namespace: testNamespace}, &existingInstance)).Should(Succeed())

			serviceHost = fmt.Sprintf("%s.local.kyma.dev", serviceName)
			existingInstance.Spec.Host = &serviceHost

			Expect(c.Update(context.Background(), &existingInstance)).Should(Succeed())

			By("Verifying APIRule after update")

			matchingLabels := matchingLabelsFunc(apiRuleName, testNamespace)

			Eventually(func(g Gomega) {
				ruleList := getRuleList(g, matchingLabels)

				//Verify All Rules point to new Service after update
				expectedUpstream := fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", serviceName, testNamespace, testServicePort)
				for i := range ruleList {
					r := ruleList[i]
					g.Expect(r.Spec.Upstream.URL).To(Equal(expectedUpstream))
				}
			}, eventuallyTimeout).Should(Succeed())

		})

		It("should succeed when creating an APIRule without a domain in cluster with kyma-gateway", func() {
			updateJwtHandlerTo(helpers.JWT_HANDLER_ISTIO)

			By("Creating Kyma gateway")

			gateway := networkingv1beta1.Gateway{
				ObjectMeta: metav1.ObjectMeta{Name: "kyma-gateway", Namespace: "kyma-system"},
				Spec: apinetworkingv1beta1.Gateway{
					Servers: []*apinetworkingv1beta1.Server{
						{
							Port: &apinetworkingv1beta1.Port{
								Protocol: "HTTPS",
							},
							Hosts: []string{
								"*.local.kyma.dev",
							},
						},
						{
							Port: &apinetworkingv1beta1.Port{
								Protocol: "HTTP",
							},
							Hosts: []string{
								"*.local.kyma.dev",
							},
						},
					},
				},
			}
			defer func() {
				deleteResource(&gateway)
			}()

			Expect(c.Create(context.Background(), &gateway)).Should(Succeed())

			By("Creating APIRule")

			rule1 := testRule("/rule1", methodsGet, defaultMutators, noConfigHandler("allow"))

			apiRuleName := generateTestName(testNameBase, testIDLength)
			serviceName := generateTestName(testServiceNameBase, testIDLength)
			serviceHost := serviceName

			apiRule := testApiRule(apiRuleName, testNamespace, serviceName, testNamespace, serviceHost, testServicePort, []gatewayv1beta1.Rule{rule1})
			svc := testService(serviceName, testNamespace, testServicePort)
			defer func() {
				deleteResource(apiRule)
				deleteResource(svc)
			}()
			// when
			Expect(c.Create(context.Background(), svc)).Should(Succeed())
			Expect(c.Create(context.Background(), apiRule)).Should(Succeed())

			expectApiRuleStatus(apiRuleName, gatewayv1beta1.StatusOK)

			By("Verifying APIRule after update")

			matchingLabels := matchingLabelsFunc(apiRuleName, testNamespace)

			Eventually(func(g Gomega) {
				ruleList := getRuleList(g, matchingLabels)

				//Verify All Rules point to new Service after update
				expectedUpstream := fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", serviceName, testNamespace, testServicePort)
				for i := range ruleList {
					r := ruleList[i]
					g.Expect(r.Spec.Upstream.URL).To(Equal(expectedUpstream))
				}
			}, eventuallyTimeout).Should(Succeed())
		})
	})

	Context("when updating the APIRule with multiple paths", func() {
		It("should create, update and delete rules depending on patch match", func() {
			updateJwtHandlerTo(helpers.JWT_HANDLER_ORY)

			rule1 := testRule("/rule1", methodsGet, defaultMutators, noConfigHandler(gatewayv1beta1.AccessStrategyNoop))
			rule2 := testRule("/rule2", methodsPut, defaultMutators, noConfigHandler(gatewayv1beta1.AccessStrategyUnauthorized))
			rule3 := testRule("/rule3", methodsDelete, defaultMutators, noConfigHandler(gatewayv1beta1.AccessStrategyAnonymous))

			apiRuleName := generateTestName(testNameBase, testIDLength)
			serviceName := generateTestName(testServiceNameBase, testIDLength)
			serviceHost := fmt.Sprintf("%s.kyma.local", serviceName)

			matchingLabels := matchingLabelsFunc(apiRuleName, testNamespace)

			pathToURLFunc := func(path string) string {
				return fmt.Sprintf("<http|https>://%s<%s>", serviceHost, path)
			}

			By("Creating APIRule")

			apiRule := testApiRule(apiRuleName, testNamespace, serviceName, testNamespace, serviceHost, testServicePort, []gatewayv1beta1.Rule{rule1, rule2, rule3})
			svc := testService(serviceName, testNamespace, testServicePort)
			defer func() {
				deleteResource(apiRule)
				deleteResource(svc)
			}()

			// when
			Expect(c.Create(context.Background(), svc)).Should(Succeed())
			Expect(c.Create(context.Background(), apiRule)).Should(Succeed())

			expectApiRuleStatus(apiRuleName, gatewayv1beta1.StatusOK)

			By("Verifying created access rules")
			Eventually(func(g Gomega) {
				ruleList := getRuleList(g, matchingLabels)
				verifyRuleList(g, ruleList, pathToURLFunc, rule1, rule2, rule3)

				//Verify All Rules point to original Service
				expectedUpstream := fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", serviceName, testNamespace, testServicePort)
				for i := range ruleList {
					r := ruleList[i]
					g.Expect(r.Spec.Upstream.URL).To(Equal(expectedUpstream))
				}

			}, eventuallyTimeout).Should(Succeed())

			By("Updating APIRule")
			existingInstance := gatewayv1beta1.APIRule{}
			Expect(c.Get(context.Background(), client.ObjectKey{Name: apiRuleName, Namespace: testNamespace}, &existingInstance)).Should(Succeed())

			rule4 := testRule("/rule4", methodsPost, defaultMutators, noConfigHandler("cookie_session"))
			existingInstance.Spec.Rules = []gatewayv1beta1.Rule{rule1, rule4}
			newServiceName := serviceName + "new"
			newServicePort := testServicePort + 3
			existingInstance.Spec.Service.Name = &newServiceName
			existingInstance.Spec.Service.Port = &newServicePort

			svcNew := testService(newServiceName, testNamespace, newServicePort)
			defer func() {
				deleteResource(svcNew)
			}()

			Expect(c.Create(context.Background(), svcNew)).Should(Succeed())
			Expect(c.Update(context.Background(), &existingInstance)).Should(Succeed())

			By("Verifying APIRule after update")

			Eventually(func(g Gomega) {
				ruleList := getRuleList(g, matchingLabels)
				verifyRuleList(g, ruleList, pathToURLFunc, rule1, rule4)

				//Verify All Rules point to new Service after update
				expectedUpstream := fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", newServiceName, testNamespace, newServicePort)
				for i := range ruleList {
					r := ruleList[i]
					g.Expect(r.Spec.Upstream.URL).To(Equal(expectedUpstream))
				}
			}, eventuallyTimeout).Should(Succeed())

		})
	})

	Context("when creating an APIRule for exposing service", func() {
		Context("on all the paths,", func() {
			Context("secured with Oauth2 introspection,", func() {
				Context("in a happy-path scenario", func() {
					It("should create a VirtualService and an AccessRule", func() {
						updateJwtHandlerTo(helpers.JWT_HANDLER_ORY)

						apiRuleName := generateTestName(testNameBase, testIDLength)
						serviceName := generateTestName(testServiceNameBase, testIDLength)
						serviceHost := fmt.Sprintf("%s.kyma.local", serviceName)

						rule := testRule(testPath, defaultMethods, defaultMutators, testOauthHandler(defaultScopes))
						apiRule := testApiRule(apiRuleName, testNamespace, serviceName, testNamespace, serviceHost, testServicePort, []gatewayv1beta1.Rule{rule})
						svc := testService(serviceName, testNamespace, testServicePort)
						defer func() {
							deleteResource(apiRule)
							deleteResource(svc)
						}()

						// when
						Expect(c.Create(context.Background(), svc)).Should(Succeed())
						Expect(c.Create(context.Background(), apiRule)).Should(Succeed())

						expectApiRuleStatus(apiRuleName, gatewayv1beta1.StatusOK)

						matchingLabels := matchingLabelsFunc(apiRuleName, testNamespace)

						By("Verifying created virtual service")
						vsList := networkingv1beta1.VirtualServiceList{}
						Eventually(func(g Gomega) {
							g.Expect(c.List(context.Background(), &vsList, matchingLabels)).Should(Succeed())
							g.Expect(vsList.Items).To(HaveLen(1))

							vs := vsList.Items[0]

							//Meta
							g.Expect(vs.Name).To(HavePrefix(apiRuleName + "-"))
							g.Expect(len(vs.Name) > len(apiRuleName)).To(BeTrue())

							expectedSpec := builders.VirtualServiceSpec().
								AddHost(serviceHost).
								Gateway(testGatewayURL).
								HTTP(builders.HTTPRoute().
									Match(builders.MatchRequest().Uri().Regex(testPath)).
									Route(builders.RouteDestination().Host(testOathkeeperSvcURL).Port(testOathkeeperPort)).
									Headers(builders.NewHttpRouteHeadersBuilder().SetHostHeader(serviceHost).Get()).
									CorsPolicy(defaultCorsPolicy).
									Timeout(defaultHttpTimeout))

							gotSpec := *expectedSpec.Get()
							g.Expect(*vs.Spec.DeepCopy()).To(Equal(*gotSpec.DeepCopy()))
						}, eventuallyTimeout).Should(Succeed())

						By("Verifying created access rules")
						expectedRuleMatchURL := fmt.Sprintf("<http|https>://%s<%s>", serviceHost, testPath)

						var rlList []rulev1alpha1.Rule
						Eventually(func(g Gomega) {
							rlList = getRuleList(g, matchingLabels)
							g.Expect(rlList).To(HaveLen(1))

							rl := rlList[0]
							g.Expect(rl.Spec.Match.URL).To(Equal(expectedRuleMatchURL))

							//Meta
							g.Expect(rl.Name).To(HavePrefix(apiRuleName + "-"))
							g.Expect(len(rl.Name) > len(apiRuleName)).To(BeTrue())

							//Spec.Upstream
							g.Expect(rl.Spec.Upstream).NotTo(BeNil())
							g.Expect(rl.Spec.Upstream.URL).To(Equal(fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", serviceName, testNamespace, testServicePort)))
							g.Expect(rl.Spec.Upstream.StripPath).To(BeNil())
							g.Expect(rl.Spec.Upstream.PreserveHost).To(BeNil())
							//Spec.Match
							g.Expect(rl.Spec.Match).NotTo(BeNil())
							g.Expect(rl.Spec.Match.URL).To(Equal(fmt.Sprintf("<http|https>://%s<%s>", serviceHost, testPath)))
							g.Expect(rl.Spec.Match.Methods).To(Equal([]string{http.MethodGet, http.MethodPut}))
							//Spec.Authenticators
							g.Expect(rl.Spec.Authenticators).To(HaveLen(1))
							g.Expect(rl.Spec.Authenticators[0].Handler).NotTo(BeNil())
							g.Expect(rl.Spec.Authenticators[0].Handler.Name).To(Equal("oauth2_introspection"))
							g.Expect(rl.Spec.Authenticators[0].Handler.Config).NotTo(BeNil())
							//Authenticators[0].Handler.Config validation
							handlerConfig := map[string]interface{}{}
							g.Expect(json.Unmarshal(rl.Spec.Authenticators[0].Config.Raw, &handlerConfig)).Should(Succeed())
							g.Expect(handlerConfig).To(HaveLen(1))
							g.Expect(asStringSlice(handlerConfig["required_scope"])).To(BeEquivalentTo(defaultScopes))
							//Spec.Authorizer
							g.Expect(rl.Spec.Authorizer).NotTo(BeNil())
							g.Expect(rl.Spec.Authorizer.Handler).NotTo(BeNil())
							g.Expect(rl.Spec.Authorizer.Handler.Name).To(Equal("allow"))
							g.Expect(rl.Spec.Authorizer.Handler.Config).To(BeNil())

							//Spec.Mutators
							g.Expect(rl.Spec.Mutators).NotTo(BeNil())
							g.Expect(len(rl.Spec.Mutators)).To(Equal(len(defaultMutators)))
							g.Expect(rl.Spec.Mutators[0].Handler.Name).To(Equal(defaultMutators[0].Name))
							g.Expect(rl.Spec.Mutators[1].Handler.Name).To(Equal(defaultMutators[1].Name))
						}, eventuallyTimeout).Should(Succeed())
					})
				})
			})

			Context("secured with JWT token authentication,", func() {
				Context("with ORY as JWT handler,", func() {
					Context("in a happy-path scenario", func() {
						It("should create a VirtualService and an AccessRules", func() {
							updateJwtHandlerTo(helpers.JWT_HANDLER_ORY)

							apiRuleName := generateTestName(testNameBase, testIDLength)
							serviceName := generateTestName(testServiceNameBase, testIDLength)
							serviceHost := fmt.Sprintf("%s.kyma.local", serviceName)

							rule1 := testRule("/img", methodsGet, defaultMutators, testOryJWTHandler(testIssuer, defaultScopes))
							rule2 := testRule("/headers", methodsGet, defaultMutators, testOryJWTHandler(testIssuer, defaultScopes))
							apiRule := testApiRule(apiRuleName, testNamespace, serviceName, testNamespace, serviceHost, testServicePort, []gatewayv1beta1.Rule{rule1, rule2})
							svc := testService(serviceName, testNamespace, testServicePort)
							defer func() {
								deleteResource(apiRule)
								deleteResource(svc)
							}()

							// when
							Expect(c.Create(context.Background(), svc)).Should(Succeed())
							Expect(c.Create(context.Background(), apiRule)).Should(Succeed())

							expectApiRuleStatus(apiRuleName, gatewayv1beta1.StatusOK)

							matchingLabels := matchingLabelsFunc(apiRuleName, testNamespace)

							By("Verifying created virtual service")
							vsList := networkingv1beta1.VirtualServiceList{}
							Eventually(func(g Gomega) {
								g.Expect(c.List(context.Background(), &vsList, matchingLabels)).Should(Succeed())
								g.Expect(vsList.Items).To(HaveLen(1))
								vs := vsList.Items[0]

								expectedSpec := builders.VirtualServiceSpec().
									AddHost(serviceHost).
									Gateway(testGatewayURL).
									HTTP(builders.HTTPRoute().
										Match(builders.MatchRequest().Uri().Regex("/img")).
										Route(builders.RouteDestination().Host(testOathkeeperSvcURL).Port(testOathkeeperPort)).
										Headers(builders.NewHttpRouteHeadersBuilder().SetHostHeader(serviceHost).Get()).
										CorsPolicy(defaultCorsPolicy).
										Timeout(defaultHttpTimeout)).
									HTTP(builders.HTTPRoute().
										Match(builders.MatchRequest().Uri().Regex("/headers")).
										Route(builders.RouteDestination().Host(testOathkeeperSvcURL).Port(testOathkeeperPort)).
										Headers(builders.NewHttpRouteHeadersBuilder().SetHostHeader(serviceHost).Get()).
										CorsPolicy(defaultCorsPolicy).
										Timeout(defaultHttpTimeout))
								gotSpec := *expectedSpec.Get()
								g.Expect(*vs.Spec.DeepCopy()).To(Equal(*gotSpec.DeepCopy()))
							}, eventuallyTimeout).Should(Succeed())

							By("Verifying created access rules")
							expectedRuleMatchURL := fmt.Sprintf("<http|https>://%s<%s>", serviceHost, "/img")

							var rlList []rulev1alpha1.Rule
							Eventually(func(g Gomega) {
								rlList = getRuleList(g, matchingLabels)
								g.Expect(rlList).To(HaveLen(2))

								rules := make(map[string]rulev1alpha1.Rule)

								for _, rule := range rlList {
									rules[rule.Spec.Match.URL] = rule
								}

								rl := rules[expectedRuleMatchURL]

								//Spec.Upstream
								g.Expect(rl.Spec.Upstream).NotTo(BeNil())
								g.Expect(rl.Spec.Upstream.URL).To(Equal(fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", serviceName, testNamespace, testServicePort)))
								g.Expect(rl.Spec.Upstream.StripPath).To(BeNil())
								g.Expect(rl.Spec.Upstream.PreserveHost).To(BeNil())
								//Spec.Match
								g.Expect(rl.Spec.Match).NotTo(BeNil())
								g.Expect(rl.Spec.Match.URL).To(Equal(expectedRuleMatchURL))
								g.Expect(rl.Spec.Match.Methods).To(Equal([]string{http.MethodGet}))
								//Spec.Authenticators
								g.Expect(rl.Spec.Authenticators).To(HaveLen(1))
								g.Expect(rl.Spec.Authenticators[0].Handler).NotTo(BeNil())
								g.Expect(rl.Spec.Authenticators[0].Handler.Name).To(Equal("jwt"))
								g.Expect(rl.Spec.Authenticators[0].Handler.Config).NotTo(BeNil())
								//Authenticators[0].Handler.Config validation
								handlerConfig := map[string]interface{}{}

								g.Expect(json.Unmarshal(rl.Spec.Authenticators[0].Config.Raw, &handlerConfig)).Should(Succeed())
								g.Expect(handlerConfig).To(HaveLen(3))
								g.Expect(asStringSlice(handlerConfig["required_scope"])).To(BeEquivalentTo(defaultScopes))
								g.Expect(asStringSlice(handlerConfig["trusted_issuers"])).To(BeEquivalentTo([]string{testIssuer}))
								//Spec.Authorizer
								g.Expect(rl.Spec.Authorizer).NotTo(BeNil())
								g.Expect(rl.Spec.Authorizer.Handler).NotTo(BeNil())
								g.Expect(rl.Spec.Authorizer.Handler.Name).To(Equal("allow"))
								g.Expect(rl.Spec.Authorizer.Handler.Config).To(BeNil())

								//Spec.Mutators
								g.Expect(rl.Spec.Mutators).NotTo(BeNil())
								g.Expect(len(rl.Spec.Mutators)).To(Equal(len(defaultMutators)))
								g.Expect(rl.Spec.Mutators[0].Handler.Name).To(Equal(defaultMutators[0].Name))
								g.Expect(rl.Spec.Mutators[1].Handler.Name).To(Equal(defaultMutators[1].Name))

								//Verify Rule2
								expectedRule2MatchURL := fmt.Sprintf("<http|https>://%s<%s>", serviceHost, "/headers")
								rl2 := rules[expectedRule2MatchURL]

								//Spec.Upstream
								g.Expect(rl2.Spec.Upstream).NotTo(BeNil())
								g.Expect(rl2.Spec.Upstream.URL).To(Equal(fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", serviceName, testNamespace, testServicePort)))
								g.Expect(rl2.Spec.Upstream.StripPath).To(BeNil())
								g.Expect(rl2.Spec.Upstream.PreserveHost).To(BeNil())
								//Spec.Match
								g.Expect(rl2.Spec.Match).NotTo(BeNil())
								g.Expect(rl2.Spec.Match.URL).To(Equal(expectedRule2MatchURL))
								g.Expect(rl2.Spec.Match.Methods).To(Equal([]string{http.MethodGet}))
								//Spec.Authenticators
								g.Expect(rl2.Spec.Authenticators).To(HaveLen(1))
								g.Expect(rl2.Spec.Authenticators[0].Handler).NotTo(BeNil())
								g.Expect(rl2.Spec.Authenticators[0].Handler.Name).To(Equal("jwt"))
								g.Expect(rl2.Spec.Authenticators[0].Handler.Config).NotTo(BeNil())
								//Authenticators[0].Handler.Config validation
								handlerConfig = map[string]interface{}{}

								g.Expect(json.Unmarshal(rl2.Spec.Authenticators[0].Config.Raw, &handlerConfig)).Should(Succeed())
								g.Expect(handlerConfig).To(HaveLen(3))
								g.Expect(asStringSlice(handlerConfig["required_scope"])).To(BeEquivalentTo(defaultScopes))
								g.Expect(asStringSlice(handlerConfig["trusted_issuers"])).To(BeEquivalentTo([]string{testIssuer}))
								//Spec.Authorizer
								g.Expect(rl2.Spec.Authorizer).NotTo(BeNil())
								g.Expect(rl2.Spec.Authorizer.Handler).NotTo(BeNil())
								g.Expect(rl2.Spec.Authorizer.Handler.Name).To(Equal("allow"))
								g.Expect(rl2.Spec.Authorizer.Handler.Config).To(BeNil())

								//Spec.Mutators
								g.Expect(rl2.Spec.Mutators).NotTo(BeNil())
								g.Expect(len(rl2.Spec.Mutators)).To(Equal(len(defaultMutators)))
								g.Expect(rl2.Spec.Mutators[0].Handler.Name).To(Equal(defaultMutators[0].Name))
								g.Expect(rl2.Spec.Mutators[1].Handler.Name).To(Equal(defaultMutators[1].Name))
							}, eventuallyTimeout).Should(Succeed())
						})
					})
				})

				Context("with Istio as JWT handler,", func() {
					Context("in a happy-path scenario", func() {
						It("should create a VirtualService, a RequestAuthentication and AuthorizationPolicies", func() {
							updateJwtHandlerTo(helpers.JWT_HANDLER_ISTIO)

							apiRuleName := generateTestName(testNameBase, testIDLength)
							serviceName := testServiceNameBase
							serviceHost := "httpbin-istio-jwt-happy-base.kyma.local"

							rule1 := testRule("/img", methodsGet, nil, testIstioJWTHandlerWithScopes(testIssuer, testJwksUri, []string{"scope-a", "scope-b"}))
							rule2 := testRule("/headers", methodsGet, nil, testIstioJWTHandlerWithScopes(testIssuer, testJwksUri, []string{"scope-c"}))
							apiRule := testApiRule(apiRuleName, testNamespace, serviceName, testNamespace, serviceHost, testServicePort, []gatewayv1beta1.Rule{rule1, rule2})
							svc := testService(serviceName, testNamespace, testServicePort)

							defer func() {
								deleteResource(apiRule)
								deleteResource(svc)
							}()

							// when
							Expect(c.Create(context.Background(), svc)).Should(Succeed())
							Expect(c.Create(context.Background(), apiRule)).Should(Succeed())

							expectApiRuleStatus(apiRuleName, gatewayv1beta1.StatusOK)

							ApiRuleNameMatchingLabels := matchingLabelsFunc(apiRuleName, testNamespace)

							By("Verifying virtual service")
							vsList := networkingv1beta1.VirtualServiceList{}
							Eventually(func(g Gomega) {
								g.Expect(c.List(context.Background(), &vsList, ApiRuleNameMatchingLabels)).Should(Succeed())
								g.Expect(vsList.Items).To(HaveLen(1))
								vs := vsList.Items[0]

								expectedSpec := builders.VirtualServiceSpec().
									AddHost(serviceHost).
									Gateway(testGatewayURL).
									HTTP(builders.HTTPRoute().
										Match(builders.MatchRequest().Uri().Regex("/img").MethodRegEx(http.MethodGet)).
										Route(builders.RouteDestination().Host(fmt.Sprintf("%s.%s.svc.cluster.local", serviceName, testNamespace)).Port(testServicePort)).
										Headers(builders.NewHttpRouteHeadersBuilder().SetHostHeader(serviceHost).Get()).
										CorsPolicy(defaultCorsPolicy).
										Timeout(defaultHttpTimeout)).
									HTTP(builders.HTTPRoute().
										Match(builders.MatchRequest().Uri().Regex("/headers").MethodRegEx(http.MethodGet)).
										Route(builders.RouteDestination().Host(fmt.Sprintf("%s.%s.svc.cluster.local", serviceName, testNamespace)).Port(testServicePort)).
										Headers(builders.NewHttpRouteHeadersBuilder().SetHostHeader(serviceHost).Get()).
										CorsPolicy(defaultCorsPolicy).
										Timeout(defaultHttpTimeout))
								gotSpec := *expectedSpec.Get()
								g.Expect(*vs.Spec.DeepCopy()).To(Equal(*gotSpec.DeepCopy()))
							}, eventuallyTimeout).Should(Succeed())

							By("Verifying request authentication")
							raList := securityv1beta1.RequestAuthenticationList{}
							Eventually(func(g Gomega) {
								g.Expect(c.List(context.Background(), &raList, ApiRuleNameMatchingLabels)).Should(Succeed())
								g.Expect(raList.Items).To(HaveLen(1))

								ra := raList.Items[0]

								g.Expect(ra.Spec.Selector.MatchLabels).To(BeEquivalentTo(map[string]string{"app": serviceName}))
								g.Expect(ra.Spec.JwtRules[0].Issuer).To(Equal(testIssuer))
								g.Expect(ra.Spec.JwtRules[0].JwksUri).To(Equal(testJwksUri))
							}, eventuallyTimeout).Should(Succeed())

							By("Verifying authorization policies")
							getByOperationPath := func(apList []*securityv1beta1.AuthorizationPolicy, path string) (*securityv1beta1.AuthorizationPolicy, error) {
								for _, ap := range apList {
									if ap.Spec.Rules[0].To[0].Operation.Paths[0] == path {
										return ap, nil
									}
								}
								return nil, fmt.Errorf("no authorization policy with operation path %s exists", path)
							}

							apList := securityv1beta1.AuthorizationPolicyList{}
							Eventually(func(g Gomega) {
								g.Expect(c.List(context.Background(), &apList, ApiRuleNameMatchingLabels)).Should(Succeed())
								g.Expect(apList.Items).To(HaveLen(2))

								hasAuthorizationPolicyWithOperationPath := func(apList []*securityv1beta1.AuthorizationPolicy, operationPath string, assertWhen func(*securityv1beta1.AuthorizationPolicy)) {
									ap, err := getByOperationPath(apList, operationPath)
									g.Expect(err).NotTo(HaveOccurred())
									g.Expect(ap.Spec.Selector.MatchLabels).To(BeEquivalentTo(map[string]string{"app": serviceName}))
									g.Expect(ap.Spec.Rules).To(HaveLen(3))

									for i := 0; i < 3; i++ {
										g.Expect(ap.Spec.Rules[i].From[0].Source.RequestPrincipals[0]).To(Equal("https://oauth2.example.com//*"))
										g.Expect(ap.Spec.Rules[i].To[0].Operation.Paths[0]).To(Equal(operationPath))
										g.Expect(ap.Spec.Rules[i].To[0].Operation.Methods).To(BeEquivalentTo([]string{http.MethodGet}))
									}

									ruleWhenKeys := append([]string{}, ap.Spec.Rules[0].When[0].Key, ap.Spec.Rules[1].When[0].Key, ap.Spec.Rules[2].When[0].Key)
									g.Expect(ruleWhenKeys).To(ContainElements("request.auth.claims[scp]", "request.auth.claims[scope]", "request.auth.claims[scopes]"))

									assertWhen(ap)
								}

								hasAuthorizationPolicyWithOperationPath(apList.Items, "/img", func(ap *securityv1beta1.AuthorizationPolicy) {
									g.Expect(ap.Spec.Rules).To(HaveLen(3))
									for i := 0; i < 3; i++ {
										g.Expect(ap.Spec.Rules[i].When).To(HaveLen(2))

										ruleWhenValues := append([]string{}, ap.Spec.Rules[0].When[0].Values...)
										ruleWhenValues = append(ruleWhenValues, ap.Spec.Rules[0].When[1].Values...)

										g.Expect(ruleWhenValues).To(ContainElements("scope-a", "scope-b"))

									}
								})

								hasAuthorizationPolicyWithOperationPath(apList.Items, "/headers", func(ap *securityv1beta1.AuthorizationPolicy) {
									g.Expect(ap.Spec.Rules).To(HaveLen(3))
									for i := 0; i < 3; i++ {
										g.Expect(ap.Spec.Rules[i].When[0].Values).To(ContainElements("scope-c"))
										g.Expect(ap.Spec.Rules[i].When[0].Values).To(ContainElements("scope-c"))
										g.Expect(ap.Spec.Rules[i].When[0].Values).To(ContainElements("scope-c"))
									}
								})
							}, eventuallyTimeout).Should(Succeed())
						})

						It("should create and update authorization policies when adding new authorization", func() {
							// given
							updateJwtHandlerTo(helpers.JWT_HANDLER_ISTIO)

							apiRuleName := generateTestName(testNameBase, testIDLength)
							serviceName := generateTestName(testServiceNameBase, testIDLength)
							serviceHost := fmt.Sprintf("%s.kyma.local", serviceName)

							authorizations := []*gatewayv1beta1.JwtAuthorization{
								{
									RequiredScopes: []string{"scope-a", "scope-b"},
								},
							}

							testIstioJWTHandlerWithAuthorizations(testIssuer, testJwksUri, authorizations)
							rule := testRule("/img", methodsGet, nil, testIstioJWTHandlerWithAuthorizations(testIssuer, testJwksUri, authorizations))
							apiRule := testApiRule(apiRuleName, testNamespace, serviceName, testNamespace, serviceHost, testServicePort, []gatewayv1beta1.Rule{rule})
							svc := testService(serviceName, testNamespace, testServicePort)

							By(fmt.Sprintf("Creating APIRule %s", apiRuleName))
							defer func() {
								deleteResource(apiRule)
								deleteResource(svc)
							}()

							// when
							Expect(c.Create(context.Background(), svc)).Should(Succeed())
							Expect(c.Create(context.Background(), apiRule)).Should(Succeed())

							Eventually(func(g Gomega) {
								createdApiRule := gatewayv1beta1.APIRule{}
								g.Expect(c.Get(context.Background(), client.ObjectKey{Name: apiRuleName, Namespace: testNamespace}, &createdApiRule)).Should(Succeed())
								g.Expect(createdApiRule.Status.APIRuleStatus).NotTo(BeNil())
								g.Expect(createdApiRule.Status.APIRuleStatus.Code).To(Equal(gatewayv1beta1.StatusOK))
								g.Expect(createdApiRule.Status.APIRuleStatus.Code).To(Equal(gatewayv1beta1.StatusOK))
							}, eventuallyTimeout).Should(Succeed())

							// when
							updatedApiRule := gatewayv1beta1.APIRule{}
							Expect(c.Get(context.Background(), client.ObjectKey{Name: apiRuleName, Namespace: testNamespace}, &updatedApiRule)).Should(Succeed())

							updatedAuthorizations := []*gatewayv1beta1.JwtAuthorization{
								{
									RequiredScopes: []string{"scope-a", "scope-c"},
								},
								{
									RequiredScopes: []string{"scope-a", "scope-d"},
								},
								{
									RequiredScopes: []string{"scope-d", "scope-b"},
								},
							}
							ruleWithScopes := testRule("/img", methodsGet, nil, testIstioJWTHandlerWithAuthorizations(testIssuer, testJwksUri, updatedAuthorizations))
							updatedApiRule.Spec.Rules = []gatewayv1beta1.Rule{ruleWithScopes}

							By(fmt.Sprintf("Updating APIRule %s with new Authorizations for /img path", apiRuleName))
							Expect(c.Update(context.Background(), &updatedApiRule)).Should(Succeed())

							// then
							matchingLabels := matchingLabelsFunc(apiRuleName, testNamespace)

							Eventually(func(g Gomega) {
								apList := securityv1beta1.AuthorizationPolicyList{}
								g.Expect(c.List(context.Background(), &apList, matchingLabels)).Should(Succeed())
								g.Expect(apList.Items).To(HaveLen(3))

								scopeAScopeCMatcher := getAuthorizationPolicyWhenScopeMatcher("scope-a", "scope-c")
								scopeAScopeDMatcher := getAuthorizationPolicyWhenScopeMatcher("scope-a", "scope-d")
								scopeDScopeBMatcher := getAuthorizationPolicyWhenScopeMatcher("scope-d", "scope-b")

								g.Expect(apList.Items).To(ContainElement(scopeAScopeCMatcher))
								g.Expect(apList.Items).To(ContainElement(scopeAScopeDMatcher))
								g.Expect(apList.Items).To(ContainElement(scopeDScopeBMatcher))
							}, eventuallyTimeout).Should(Succeed())
						})
					})
				})
			})

			Context("when service has custom label selectors,", func() {
				It("should create a RequestAuthentication and AuthorizationPolicy with custom label selector from service", func() {
					updateJwtHandlerTo(helpers.JWT_HANDLER_ISTIO)

					apiRuleName := generateTestName(testNameBase, testIDLength)
					serviceName := testServiceNameBase
					serviceHost := fmt.Sprintf("%s.kyma.local", serviceName)

					rule1 := testRule("/img", methodsGet, nil, testIstioJWTHandler(testIssuer, testJwksUri))
					apiRule := testApiRule(apiRuleName, testNamespace, serviceName, testNamespace, serviceHost, testServicePort, []gatewayv1beta1.Rule{rule1})
					svc := testService(serviceName, testNamespace, testServicePort)
					delete(svc.Spec.Selector, "app")
					svc.Spec.Selector["custom"] = serviceName
					defer func() {
						deleteResource(apiRule)
						deleteResource(svc)
					}()

					// when
					Expect(c.Create(context.Background(), svc)).Should(Succeed())
					Expect(c.Create(context.Background(), apiRule)).Should(Succeed())

					expectApiRuleStatus(apiRuleName, gatewayv1beta1.StatusOK)

					ApiRuleNameMatchingLabels := matchingLabelsFunc(apiRuleName, testNamespace)

					By("Verifying request authentication")

					raList := securityv1beta1.RequestAuthenticationList{}
					Eventually(func(g Gomega) {
						g.Expect(c.List(context.Background(), &raList, ApiRuleNameMatchingLabels)).Should(Succeed())
						g.Expect(raList.Items).To(HaveLen(1))
						g.Expect(raList.Items[0].Spec.Selector.MatchLabels).To(HaveLen(1))
						g.Expect(raList.Items[0].Spec.Selector.MatchLabels).To(HaveKeyWithValue("custom", serviceName))
					}, eventuallyTimeout).Should(Succeed())

					By("Verifying authorization policy")

					apList := securityv1beta1.AuthorizationPolicyList{}
					Eventually(func(g Gomega) {
						g.Expect(c.List(context.Background(), &apList, ApiRuleNameMatchingLabels)).Should(Succeed())
						g.Expect(apList.Items).To(HaveLen(1))
						g.Expect(apList.Items[0].Spec.Selector.MatchLabels).To(HaveLen(1))
						g.Expect(apList.Items[0].Spec.Selector.MatchLabels).To(HaveKeyWithValue("custom", serviceName))
					}, eventuallyTimeout).Should(Succeed())
				})

				It("should create a RequestAuthentication and AuthorizationPolicy with multiple custom label selectors from service", func() {
					updateJwtHandlerTo(helpers.JWT_HANDLER_ISTIO)

					apiRuleName := generateTestName(testNameBase, testIDLength)
					serviceName := testServiceNameBase
					serviceHost := fmt.Sprintf("%s.kyma.local", serviceName)

					rule1 := testRule("/img", methodsGet, nil, testIstioJWTHandler(testIssuer, testJwksUri))
					apiRule := testApiRule(apiRuleName, testNamespace, serviceName, testNamespace, serviceHost, testServicePort, []gatewayv1beta1.Rule{rule1})
					svc := testService(serviceName, testNamespace, testServicePort)
					delete(svc.Spec.Selector, "app")
					svc.Spec.Selector["custom"] = serviceName
					svc.Spec.Selector["second-custom"] = "blah"
					defer func() {
						deleteResource(apiRule)
						deleteResource(svc)
					}()

					// when
					Expect(c.Create(context.Background(), svc)).Should(Succeed())
					Expect(c.Create(context.Background(), apiRule)).Should(Succeed())

					expectApiRuleStatus(apiRuleName, gatewayv1beta1.StatusOK)

					ApiRuleNameMatchingLabels := matchingLabelsFunc(apiRuleName, testNamespace)

					By("Verifying request authentication")

					raList := securityv1beta1.RequestAuthenticationList{}
					Eventually(func(g Gomega) {
						g.Expect(c.List(context.Background(), &raList, ApiRuleNameMatchingLabels)).Should(Succeed())
						g.Expect(raList.Items).To(HaveLen(1))
						g.Expect(raList.Items[0].Spec.Selector.MatchLabels).To(HaveLen(2))
						g.Expect(raList.Items[0].Spec.Selector.MatchLabels).To(HaveKeyWithValue("custom", serviceName))
						g.Expect(raList.Items[0].Spec.Selector.MatchLabels).To(HaveKeyWithValue("second-custom", "blah"))
					}, eventuallyTimeout).Should(Succeed())

					By("Verifying authorization policy")

					apList := securityv1beta1.AuthorizationPolicyList{}
					Eventually(func(g Gomega) {
						g.Expect(c.List(context.Background(), &apList, ApiRuleNameMatchingLabels)).Should(Succeed())
						g.Expect(apList.Items).To(HaveLen(1))
						g.Expect(apList.Items[0].Spec.Selector.MatchLabels).To(HaveLen(2))
						g.Expect(apList.Items[0].Spec.Selector.MatchLabels).To(HaveKeyWithValue("custom", serviceName))
						g.Expect(apList.Items[0].Spec.Selector.MatchLabels).To(HaveKeyWithValue("second-custom", "blah"))
					}, eventuallyTimeout).Should(Succeed())
				})
			})
		})

		Context("on specified paths", func() {
			Context("with multiple endpoints secured with different authentication methods", func() {
				Context("in the happy path scenario", func() {
					It("should create a VS with corresponding matchers and access rules for each secured path", func() {
						updateJwtHandlerTo(helpers.JWT_HANDLER_ORY)

						jwtHandler := testOryJWTHandler(testIssuer, defaultScopes)
						oauthHandler := testOauthHandler(defaultScopes)
						rule1 := testRule("/img", methodsGet, defaultMutators, jwtHandler)
						rule2 := testRule("/headers", methodsGet, defaultMutators, oauthHandler)
						rule3 := testRule("/status", methodsGet, defaultMutators, noConfigHandler(gatewayv1beta1.AccessStrategyNoop))
						rule4 := testRule("/favicon", methodsGet, nil, noConfigHandler(gatewayv1beta1.AccessStrategyAllow))

						apiRuleName := generateTestName(testNameBase, testIDLength)
						serviceName := testServiceNameBase
						serviceHost := "httpbin4.kyma.local"

						apiRule := testApiRule(apiRuleName, testNamespace, serviceName, testNamespace, serviceHost, testServicePort, []gatewayv1beta1.Rule{rule1, rule2, rule3, rule4})
						svc := testService(serviceName, testNamespace, testServicePort)
						defer func() {
							deleteResource(apiRule)
							deleteResource(svc)
						}()

						// when
						Expect(c.Create(context.Background(), svc)).Should(Succeed())
						Expect(c.Create(context.Background(), apiRule)).Should(Succeed())

						expectApiRuleStatus(apiRuleName, gatewayv1beta1.StatusOK)

						matchingLabels := matchingLabelsFunc(apiRuleName, testNamespace)

						By("Verifying created virtual service")
						vsList := networkingv1beta1.VirtualServiceList{}
						Eventually(func(g Gomega) {
							g.Expect(c.List(context.Background(), &vsList, matchingLabels)).Should(Succeed())
							g.Expect(vsList.Items).To(HaveLen(1))

							vs := vsList.Items[0]

							expectedSpec := builders.VirtualServiceSpec().
								AddHost(serviceHost).
								Gateway(testGatewayURL).
								HTTP(builders.HTTPRoute().
									Match(builders.MatchRequest().Uri().Regex("/img")).
									Route(builders.RouteDestination().Host(testOathkeeperSvcURL).Port(testOathkeeperPort)).
									Headers(builders.NewHttpRouteHeadersBuilder().SetHostHeader(serviceHost).Get()).
									CorsPolicy(defaultCorsPolicy).
									Timeout(defaultHttpTimeout)).
								HTTP(builders.HTTPRoute().
									Match(builders.MatchRequest().Uri().Regex("/headers")).
									Route(builders.RouteDestination().Host(testOathkeeperSvcURL).Port(testOathkeeperPort)).
									Headers(builders.NewHttpRouteHeadersBuilder().SetHostHeader(serviceHost).Get()).
									CorsPolicy(defaultCorsPolicy).
									Timeout(defaultHttpTimeout)).
								HTTP(builders.HTTPRoute().
									Match(builders.MatchRequest().Uri().Regex("/status")).
									Route(builders.RouteDestination().Host(testOathkeeperSvcURL).Port(testOathkeeperPort)).
									Headers(builders.NewHttpRouteHeadersBuilder().SetHostHeader(serviceHost).Get()).
									CorsPolicy(defaultCorsPolicy).
									Timeout(defaultHttpTimeout)).
								HTTP(builders.HTTPRoute().
									Match(builders.MatchRequest().Uri().Regex("/favicon")).
									Route(builders.RouteDestination().Host("httpbin.atgo-system.svc.cluster.local").Port(443)). // "allow", no oathkeeper rule!
									Headers(builders.NewHttpRouteHeadersBuilder().SetHostHeader(serviceHost).Get()).
									CorsPolicy(defaultCorsPolicy).
									Timeout(defaultHttpTimeout))

							gotSpec := *expectedSpec.Get()
							g.Expect(*vs.Spec.DeepCopy()).To(Equal(*gotSpec.DeepCopy()))
						}, eventuallyTimeout).Should(Succeed())

						By("Verifying created access rules")
						for _, tc := range []struct {
							path    string
							handler string
							config  []byte
						}{
							{path: "img", handler: "jwt", config: jwtHandler.Config.Raw},
							{path: "headers", handler: "oauth2_introspection", config: oauthHandler.Config.Raw},
							{path: "status", handler: "noop", config: nil},
						} {
							expectedRuleMatchURL := fmt.Sprintf("<http|https>://%s</%s>", serviceHost, tc.path)

							var rlList []rulev1alpha1.Rule
							Eventually(func(g Gomega) {
								rlList = getRuleList(g, matchingLabels)

								// Make sure no access rules for allow and no_auth handlers are created
								g.Expect(rlList).To(HaveLen(3))

								rules := make(map[string]rulev1alpha1.Rule)

								for _, rule := range rlList {
									rules[rule.Spec.Match.URL] = rule
								}

								rl := rules[expectedRuleMatchURL]

								//Spec.Upstream
								g.Expect(rl.Spec.Upstream).NotTo(BeNil())
								g.Expect(rl.Spec.Upstream.URL).To(Equal(fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", serviceName, testNamespace, testServicePort)))
								g.Expect(rl.Spec.Upstream.StripPath).To(BeNil())
								g.Expect(rl.Spec.Upstream.PreserveHost).To(BeNil())

								//Spec.Match
								g.Expect(rl.Spec.Match).NotTo(BeNil())
								g.Expect(rl.Spec.Match.Methods).To(Equal([]string{http.MethodGet}))
								g.Expect(rl.Spec.Match.URL).To(Equal(expectedRuleMatchURL))

								//Spec.Authenticators
								g.Expect(rl.Spec.Authenticators).To(HaveLen(1))
								g.Expect(rl.Spec.Authenticators[0].Handler).NotTo(BeNil())
								g.Expect(rl.Spec.Authenticators[0].Handler.Name).To(Equal(tc.handler))

								if tc.config != nil {
									//Authenticators[0].Handler.Config validation
									g.Expect(rl.Spec.Authenticators[0].Handler.Config).NotTo(BeNil())
									g.Expect(rl.Spec.Authenticators[0].Handler.Config.Raw).To(MatchJSON(tc.config))
								}

								//Spec.Authorizer
								g.Expect(rl.Spec.Authorizer).NotTo(BeNil())
								g.Expect(rl.Spec.Authorizer.Handler).NotTo(BeNil())
								g.Expect(rl.Spec.Authorizer.Handler.Name).To(Equal("allow"))
								g.Expect(rl.Spec.Authorizer.Handler.Config).To(BeNil())

								//Spec.Mutators
								g.Expect(rl.Spec.Mutators).NotTo(BeNil())
								g.Expect(len(rl.Spec.Mutators)).To(Equal(len(defaultMutators)))
								g.Expect(rl.Spec.Mutators[0].Handler.Name).To(Equal(defaultMutators[0].Name))
								g.Expect(rl.Spec.Mutators[1].Handler.Name).To(Equal(defaultMutators[1].Name))
							}, eventuallyTimeout).Should(Succeed())
						}
					})

					jwtHandlers := []string{helpers.JWT_HANDLER_ORY, helpers.JWT_HANDLER_ISTIO}
					for _, jwtHandler := range jwtHandlers {
						Context(fmt.Sprintf("with %s as JWT handler", jwtHandler), func() {
							It("should create a VS, but no access rule for allow and no_auth handler", func() {
								updateJwtHandlerTo(jwtHandler)

								rule1 := testRule("/favicon", methodsGet, nil, noConfigHandler(gatewayv1beta1.AccessStrategyAllow))
								rule2 := testRule("/anything", methodsGet, nil, noConfigHandler(gatewayv1beta1.AccessStrategyNoAuth))

								apiRuleName := generateTestName(testNameBase, testIDLength)
								serviceName := testServiceNameBase
								serviceHost := "httpbin4.kyma.local"

								apiRule := testApiRule(apiRuleName, testNamespace, serviceName, testNamespace, serviceHost, testServicePort, []gatewayv1beta1.Rule{rule1, rule2})
								svc := testService(serviceName, testNamespace, testServicePort)

								defer func() {
									deleteResource(apiRule)
									deleteResource(svc)
								}()

								// when
								Expect(c.Create(context.Background(), svc)).Should(Succeed())
								Expect(c.Create(context.Background(), apiRule)).Should(Succeed())

								expectApiRuleStatus(apiRuleName, gatewayv1beta1.StatusOK)

								matchingLabels := matchingLabelsFunc(apiRuleName, testNamespace)

								By("Verifying created virtual service")
								vsList := networkingv1beta1.VirtualServiceList{}
								Eventually(func(g Gomega) {
									g.Expect(c.List(context.Background(), &vsList, matchingLabels)).Should(Succeed())
									g.Expect(vsList.Items).To(HaveLen(1))

									vs := vsList.Items[0]

									expectedSpec := builders.VirtualServiceSpec().
										AddHost(serviceHost).
										Gateway(testGatewayURL).
										HTTP(builders.HTTPRoute().
											Match(builders.MatchRequest().Uri().Regex("/favicon")).
											Route(builders.RouteDestination().Host("httpbin.atgo-system.svc.cluster.local").Port(443)).
											Headers(builders.NewHttpRouteHeadersBuilder().SetHostHeader(serviceHost).Get()).
											CorsPolicy(defaultCorsPolicy).
											Timeout(defaultHttpTimeout)).
										HTTP(builders.HTTPRoute().
											Match(builders.MatchRequest().Uri().Regex("/anything").MethodRegEx("GET")).
											Route(builders.RouteDestination().Host("httpbin.atgo-system.svc.cluster.local").Port(443)).
											Headers(builders.NewHttpRouteHeadersBuilder().SetHostHeader(serviceHost).Get()).
											CorsPolicy(defaultCorsPolicy).
											Timeout(defaultHttpTimeout))

									gotSpec := *expectedSpec.Get()
									g.Expect(*vs.Spec.DeepCopy()).To(Equal(*gotSpec.DeepCopy()))
								}, eventuallyTimeout).Should(Succeed())

								By("Verifying no Oathkeeper rule is created")

								var rlList []rulev1alpha1.Rule
								Eventually(func(g Gomega) {
									rlList = getRuleList(g, matchingLabels)

									g.Expect(rlList).To(HaveLen(0))

								}, eventuallyTimeout).Should(Succeed())
							})
						})
					}
				})
			})
		})
	})

	Context("Changing JWT handler in config map", func() {
		Context("Handler is ory and ApiRule with JWT handler rule exists", func() {
			Context("changing jwt handler to istio", func() {
				It("Should have validation errors for APiRule JWT handler configuration and rule is not deleted", func() {
					// given
					updateJwtHandlerTo(helpers.JWT_HANDLER_ORY)

					apiRuleName := generateTestName(testNameBase, testIDLength)
					testServiceHost := fmt.Sprintf("httpbin-%s.kyma.local", apiRuleName)

					rule := testRule("/img", methodsGet, nil, testOryJWTHandler(testIssuer, defaultScopes))
					apiRule := testApiRule(apiRuleName, testNamespace, testServiceNameBase, testNamespace, testServiceHost, testServicePort, []gatewayv1beta1.Rule{rule})
					svc := testService(testServiceNameBase, testNamespace, testServicePort)

					By("Creating ApiRule with Rule using Ory JWT handler configuration")
					defer func() {
						deleteResource(apiRule)
						deleteResource(svc)
					}()

					// when
					Expect(c.Create(context.Background(), svc)).Should(Succeed())
					Expect(c.Create(context.Background(), apiRule)).Should(Succeed())

					expectApiRuleStatus(apiRuleName, gatewayv1beta1.StatusOK)

					// when
					updateJwtHandlerTo(helpers.JWT_HANDLER_ISTIO)

					// then
					Eventually(func(g Gomega) {
						apiRule := gatewayv1beta1.APIRule{}
						g.Expect(c.Get(context.Background(), client.ObjectKey{Name: apiRuleName, Namespace: testNamespace}, &apiRule)).Should(Succeed())
						g.Expect(apiRule.Status.APIRuleStatus.Code).To(Equal(gatewayv1beta1.StatusError))
						g.Expect(apiRule.Status.APIRuleStatus.Description).To(ContainSubstring("Multiple validation errors"))

						shouldHaveRules(g, apiRuleName, testNamespace, 1)
					}, eventuallyTimeout).Should(Succeed())
				})

				It("Should create AP and RA and delete JWT Access Rule when ApiRule JWT handler configuration was updated to have valid config for istio", func() {
					// given
					updateJwtHandlerTo(helpers.JWT_HANDLER_ORY)

					apiRuleName := generateTestName(testNameBase, testIDLength)
					testServiceHost := fmt.Sprintf("httpbin-%s.kyma.local", apiRuleName)

					rule := testRule("/img", methodsGet, nil, testOryJWTHandler(testIssuer, defaultScopes))
					apiRule := testApiRule(apiRuleName, testNamespace, testServiceNameBase, testNamespace, testServiceHost, testServicePort, []gatewayv1beta1.Rule{rule})
					svc := testService(testServiceNameBase, testNamespace, testServicePort)

					By("Creating ApiRule with Rule using Ory JWT handler")
					defer func() {
						deleteResource(apiRule)
						deleteResource(svc)
					}()

					// when
					Expect(c.Create(context.Background(), svc)).Should(Succeed())
					Expect(c.Create(context.Background(), apiRule)).Should(Succeed())

					expectApiRuleStatus(apiRuleName, gatewayv1beta1.StatusOK)
					updateJwtHandlerTo(helpers.JWT_HANDLER_ISTIO)
					expectApiRuleStatus(apiRuleName, gatewayv1beta1.StatusError)

					// when
					By("Updating JWT handler configuration in ApiRule to be valid for istio")
					istioJwtRule := testRule("/img", methodsGet, nil, testIstioJWTHandler(testIssuer, testJwksUri))
					Eventually(func(g Gomega) {
						updatedApiRule := gatewayv1beta1.APIRule{}
						g.Expect(c.Get(context.Background(), client.ObjectKey{Name: apiRuleName, Namespace: testNamespace}, &updatedApiRule)).Should(Succeed())
						updatedApiRule.Spec.Rules = []gatewayv1beta1.Rule{istioJwtRule}
						g.Expect(c.Update(context.Background(), &updatedApiRule)).Should(Succeed())
					}, eventuallyTimeout).Should(Succeed())

					// then
					expectApiRuleStatus(apiRuleName, gatewayv1beta1.StatusOK)

					Eventually(func(g Gomega) {
						shouldHaveRequestAuthentications(g, apiRuleName, testNamespace, 1)
						shouldHaveAuthorizationPolicies(g, apiRuleName, testNamespace, 1)
						shouldHaveRules(g, apiRuleName, testNamespace, 0)
					}, eventuallyTimeout).Should(Succeed())
				})
			})
		})

		Context("Handler is istio and ApiRule with JWT handler specific resources exists", func() {
			Context("changing jwt handler to ory", func() {
				It("Should have validation errors for APiRule JWT handler configuration and resources are not deleted", func() {
					// given
					updateJwtHandlerTo(helpers.JWT_HANDLER_ISTIO)

					apiRuleName := generateTestName(testNameBase, testIDLength)
					testServiceHost := fmt.Sprintf("httpbin-%s.kyma.local", apiRuleName)

					rule := testRule("/img", methodsGet, nil, testIstioJWTHandler(testIssuer, testJwksUri))
					apiRule := testApiRule(apiRuleName, testNamespace, testServiceNameBase, testNamespace, testServiceHost, testServicePort, []gatewayv1beta1.Rule{rule})
					svc := testService(testServiceNameBase, testNamespace, testServicePort)

					By("Creating ApiRule with Rule using Istio JWT handler configuration")
					defer func() {
						deleteResource(apiRule)
						deleteResource(svc)
					}()

					// when
					Expect(c.Create(context.Background(), svc)).Should(Succeed())
					Expect(c.Create(context.Background(), apiRule)).Should(Succeed())

					By("Waiting until reconciliation of API Rule has finished")
					expectApiRuleStatus(apiRuleName, gatewayv1beta1.StatusOK)

					// when
					updateJwtHandlerTo(helpers.JWT_HANDLER_ORY)

					// then
					Eventually(func(g Gomega) {
						apiRule := gatewayv1beta1.APIRule{}
						g.Expect(c.Get(context.Background(), client.ObjectKey{Name: apiRuleName, Namespace: testNamespace}, &apiRule)).Should(Succeed())
						g.Expect(apiRule.Status.APIRuleStatus.Code).To(Equal(gatewayv1beta1.StatusError))
						g.Expect(apiRule.Status.APIRuleStatus.Description).To(ContainSubstring("Validation error"))

						shouldHaveRequestAuthentications(g, apiRuleName, testNamespace, 1)
						shouldHaveAuthorizationPolicies(g, apiRuleName, testNamespace, 1)
					}, eventuallyTimeout).Should(Succeed())
				})

				It("Should create Access Rule and delete RA and AP when ApiRule JWT handler configuration was updated to have valid config for ory", func() {
					// given
					updateJwtHandlerTo(helpers.JWT_HANDLER_ISTIO)

					apiRuleName := generateTestName(testNameBase, testIDLength)
					testServiceHost := fmt.Sprintf("httpbin-%s.kyma.local", apiRuleName)

					rule := testRule("/img", methodsGet, nil, testIstioJWTHandler(testIssuer, testJwksUri))
					apiRule := testApiRule(apiRuleName, testNamespace, testServiceNameBase, testNamespace, testServiceHost, testServicePort, []gatewayv1beta1.Rule{rule})
					svc := testService(testServiceNameBase, testNamespace, testServicePort)

					By("Creating ApiRule with Rule using JWT handler configuration")
					defer func() {
						deleteResource(apiRule)
						deleteResource(svc)
					}()

					// when
					Expect(c.Create(context.Background(), svc)).Should(Succeed())
					Expect(c.Create(context.Background(), apiRule)).Should(Succeed())

					By("Waiting until reconciliation of API Rule has finished")
					expectApiRuleStatus(apiRuleName, gatewayv1beta1.StatusOK)
					updateJwtHandlerTo(helpers.JWT_HANDLER_ORY)

					By("Waiting until reconciliation of API Rule has finished")
					Eventually(func(g Gomega) {
						apiRule := gatewayv1beta1.APIRule{}
						g.Expect(c.Get(context.Background(), client.ObjectKey{Name: apiRuleName, Namespace: testNamespace}, &apiRule)).Should(Succeed())
						g.Expect(apiRule.Status.APIRuleStatus).NotTo(BeNil())
						g.Expect(apiRule.Status.APIRuleStatus.Code).To(Equal(gatewayv1beta1.StatusError))
					}, eventuallyTimeout).Should(Succeed())

					// when
					By("Updating JWT handler in ApiRule to be valid for ory")
					Eventually(func(g Gomega) {
						oryJwtRule := testRule("/img", methodsGet, nil, testOryJWTHandler(testIssuer, defaultScopes))
						updatedApiRule := gatewayv1beta1.APIRule{}
						Expect(c.Get(context.Background(), client.ObjectKey{Name: apiRuleName, Namespace: testNamespace}, &updatedApiRule)).Should(Succeed())
						updatedApiRule.Spec.Rules = []gatewayv1beta1.Rule{oryJwtRule}
						Expect(c.Update(context.Background(), &updatedApiRule)).Should(Succeed())
					}, eventuallyTimeout).Should(Succeed())

					// then
					expectApiRuleStatus(apiRuleName, gatewayv1beta1.StatusOK)

					Eventually(func(g Gomega) {
						shouldHaveRequestAuthentications(g, apiRuleName, testNamespace, 0)
						shouldHaveAuthorizationPolicies(g, apiRuleName, testNamespace, 0)
						shouldHaveRules(g, apiRuleName, testNamespace, 1)
					}, eventuallyTimeout).Should(Succeed())
				})
			})
		})
	})

	Context("when creating APIRule in version v2alpha1", Ordered, func() {
		Context("respect x-validation rules only", Ordered, func() {
			BeforeAll(func() {
				updateJwtHandlerTo(helpers.JWT_HANDLER_ORY)
			})

			It("should be able to create an APIRule with noAuth=true", func() {
				apiRuleName := generateTestName(testNameBase, testIDLength)
				serviceName := testServiceNameBase
				serviceHost := gatewayv2alpha1.Host("httpbin-istio-jwt-happy-base.kyma.local")
				serviceHosts := []*gatewayv2alpha1.Host{&serviceHost}

				rule := testRulev2alpha1("/img", []gatewayv2alpha1.HttpMethod{http.MethodGet})
				rule.NoAuth = ptr.To(true)
				apiRule := testApiRulev2alpha1(apiRuleName, testNamespace, serviceName, testNamespace, serviceHosts, testServicePort, []gatewayv2alpha1.Rule{rule})
				svc := testService(serviceName, testNamespace, testServicePort)
				defer func() {
					deleteResource(apiRule)
					deleteResource(svc)
				}()

				// when
				Expect(c.Create(context.Background(), svc)).Should(Succeed())
				Expect(c.Create(context.Background(), apiRule)).Should(Succeed())

			})

			It("should be able to create an APIRule with jwt", func() {
				apiRuleName := generateTestName(testNameBase, testIDLength)
				serviceName := testServiceNameBase
				serviceHost := gatewayv2alpha1.Host("httpbin-istio-jwt-happy-base.kyma.local")
				serviceHosts := []*gatewayv2alpha1.Host{&serviceHost}

				rule := testRulev2alpha1("/img", []gatewayv2alpha1.HttpMethod{http.MethodGet})
				rule.Jwt = &gatewayv2alpha1.JwtConfig{
					Authentications: []*gatewayv2alpha1.JwtAuthentication{},
					Authorizations:  []*gatewayv2alpha1.JwtAuthorization{},
				}
				apiRule := testApiRulev2alpha1(apiRuleName, testNamespace, serviceName, testNamespace, serviceHosts, testServicePort, []gatewayv2alpha1.Rule{rule})
				svc := testService(serviceName, testNamespace, testServicePort)
				defer func() {
					deleteResource(apiRule)
					deleteResource(svc)
				}()

				// when
				Expect(c.Create(context.Background(), svc)).Should(Succeed())
				Expect(c.Create(context.Background(), apiRule)).Should(Succeed())

			})

			It("should be able to create an APIRule with jwt and noAuth=false", func() {
				apiRuleName := generateTestName(testNameBase, testIDLength)
				serviceName := testServiceNameBase
				serviceHost := gatewayv2alpha1.Host("httpbin-istio-jwt-happy-base.kyma.local")
				serviceHosts := []*gatewayv2alpha1.Host{&serviceHost}

				rule := testRulev2alpha1("/img", []gatewayv2alpha1.HttpMethod{http.MethodGet})
				rule.NoAuth = ptr.To(false)
				rule.Jwt = &gatewayv2alpha1.JwtConfig{
					Authentications: []*gatewayv2alpha1.JwtAuthentication{},
					Authorizations:  []*gatewayv2alpha1.JwtAuthorization{},
				}
				apiRule := testApiRulev2alpha1(apiRuleName, testNamespace, serviceName, testNamespace, serviceHosts, testServicePort, []gatewayv2alpha1.Rule{rule})
				svc := testService(serviceName, testNamespace, testServicePort)
				defer func() {
					deleteResource(apiRule)
					deleteResource(svc)
				}()
				// when
				Expect(c.Create(context.Background(), svc)).Should(Succeed())
				Expect(c.Create(context.Background(), apiRule)).Should(Succeed())

			})

			It("should be able to create an APIRule with jwt and mutators", func() {
				apiRuleName := generateTestName(testNameBase, testIDLength)
				serviceName := testServiceNameBase
				serviceHost := gatewayv2alpha1.Host("httpbin-istio-jwt-happy-base.kyma.local")
				serviceHosts := []*gatewayv2alpha1.Host{&serviceHost}

				rule := testRulev2alpha1("/img", []gatewayv2alpha1.HttpMethod{http.MethodGet})
				rule.Jwt = &gatewayv2alpha1.JwtConfig{
					Authentications: []*gatewayv2alpha1.JwtAuthentication{},
					Authorizations:  []*gatewayv2alpha1.JwtAuthorization{},
				}
				apiRule := testApiRulev2alpha1(apiRuleName, testNamespace, serviceName, testNamespace, serviceHosts, testServicePort, []gatewayv2alpha1.Rule{rule})
				svc := testService(serviceName, testNamespace, testServicePort)
				defer func() {
					deleteResource(apiRule)
					deleteResource(svc)
				}()

				// when
				Expect(c.Create(context.Background(), svc)).Should(Succeed())
				Expect(c.Create(context.Background(), apiRule)).Should(Succeed())

			})

			It("should fail to create an APIRule without noAuth and jwt", func() {
				apiRuleName := generateTestName(testNameBase, testIDLength)
				serviceName := testServiceNameBase
				serviceHost := gatewayv2alpha1.Host("httpbin-istio-jwt-happy-base.kyma.local")
				serviceHosts := []*gatewayv2alpha1.Host{&serviceHost}

				rule := testRulev2alpha1("/img", []gatewayv2alpha1.HttpMethod{http.MethodGet})
				apiRule := testApiRulev2alpha1(apiRuleName, testNamespace, serviceName, testNamespace, serviceHosts, testServicePort, []gatewayv2alpha1.Rule{rule})
				svc := testService(serviceName, testNamespace, testServicePort)
				defer func() {
					deleteResource(apiRule)
					deleteResource(svc)
				}()

				// when
				Expect(c.Create(context.Background(), svc)).Should(Succeed())
				err := c.Create(context.Background(), apiRule)

				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("One of the following fields must be set: noAuth, jwt, extAuth"))
			})

			It("should fail to create an APIRule with noAuth=false", func() {
				apiRuleName := generateTestName(testNameBase, testIDLength)
				serviceName := testServiceNameBase
				serviceHost := gatewayv2alpha1.Host("httpbin-istio-jwt-happy-base.kyma.local")
				serviceHosts := []*gatewayv2alpha1.Host{&serviceHost}

				rule := testRulev2alpha1("/img", []gatewayv2alpha1.HttpMethod{http.MethodGet})
				rule.NoAuth = ptr.To(false)
				apiRule := testApiRulev2alpha1(apiRuleName, testNamespace, serviceName, testNamespace, serviceHosts, testServicePort, []gatewayv2alpha1.Rule{rule})
				svc := testService(serviceName, testNamespace, testServicePort)
				defer func() {
					deleteResource(apiRule)
					deleteResource(svc)
				}()

				// when
				Expect(c.Create(context.Background(), svc)).Should(Succeed())
				err := c.Create(context.Background(), apiRule)

				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("One of the following fields must be set: noAuth, jwt, extAuth"))
			})

			It("should fail to create an APIRule with jwt and noAuth=true", func() {
				apiRuleName := generateTestName(testNameBase, testIDLength)
				serviceName := testServiceNameBase
				serviceHost := gatewayv2alpha1.Host("httpbin-istio-jwt-happy-base.kyma.local")
				serviceHosts := []*gatewayv2alpha1.Host{&serviceHost}

				rule := testRulev2alpha1("/img", []gatewayv2alpha1.HttpMethod{http.MethodGet})
				rule.NoAuth = ptr.To(true)
				rule.Jwt = &gatewayv2alpha1.JwtConfig{
					Authentications: []*gatewayv2alpha1.JwtAuthentication{},
					Authorizations:  []*gatewayv2alpha1.JwtAuthorization{},
				}
				apiRule := testApiRulev2alpha1(apiRuleName, testNamespace, serviceName, testNamespace, serviceHosts, testServicePort, []gatewayv2alpha1.Rule{rule})
				svc := testService(serviceName, testNamespace, testServicePort)
				defer func() {
					deleteResource(apiRule)
					deleteResource(svc)
				}()

				// when
				Expect(c.Create(context.Background(), svc)).Should(Succeed())
				err := c.Create(context.Background(), apiRule)

				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("One of the following fields must be set: noAuth, jwt, extAuth"))
			})

			It("should fail to create an APIRule with more than one host", func() {
				apiRuleName := generateTestName(testNameBase, testIDLength)
				serviceName := testServiceNameBase
				serviceHost := gatewayv2alpha1.Host("httpbin-istio-jwt-happy-base.kyma.local")
				secondServiceHost := gatewayv2alpha1.Host("other-istio-jwt-happy-base.kyma.local")
				serviceHosts := []*gatewayv2alpha1.Host{&serviceHost, &secondServiceHost}

				rule := testRulev2alpha1("/img", []gatewayv2alpha1.HttpMethod{http.MethodGet})
				rule.NoAuth = ptr.To(true)
				apiRule := testApiRulev2alpha1(apiRuleName, testNamespace, serviceName, testNamespace, serviceHosts, testServicePort, []gatewayv2alpha1.Rule{rule})
				svc := testService(serviceName, testNamespace, testServicePort)
				defer func() {
					deleteResource(apiRule)
					deleteResource(svc)
				}()

				// when
				Expect(c.Create(context.Background(), svc)).Should(Succeed())
				err := c.Create(context.Background(), apiRule)

				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("spec.hosts: Too many: 2: must have at most 1 items"))
			})
		})

		Context("gateway name should be valid", Ordered, func() {
			It("should create an APIRule with a valid gateway", func() {
				// given
				apiRuleName := generateTestName(testNameBase, testIDLength)
				serviceName := testServiceNameBase

				rule := testRulev2alpha1("/img", []gatewayv2alpha1.HttpMethod{http.MethodGet})
				rule.NoAuth = ptr.To(true)
				apiRule := testApiRulev2alpha1Gateway(apiRuleName, testNamespace, serviceName, testNamespace, testGatewayURL, testServicePort, []gatewayv2alpha1.Rule{rule})
				svc := testService(serviceName, testNamespace, testServicePort)
				defer func() {
					deleteResource(apiRule)
					deleteResource(svc)
				}()

				// when
				Expect(c.Create(context.Background(), svc)).Should(Succeed())
				Expect(c.Create(context.Background(), apiRule)).Should(Succeed())

			})

			invalidHelper := func(gatewayName string) {
				// given
				apiRuleName := generateTestName(testNameBase, testIDLength)
				serviceName := testServiceNameBase

				rule := testRulev2alpha1("/img", []gatewayv2alpha1.HttpMethod{http.MethodGet})
				rule.NoAuth = ptr.To(true)
				apiRule := testApiRulev2alpha1Gateway(apiRuleName, testNamespace, serviceName, testNamespace, gatewayName, testServicePort, []gatewayv2alpha1.Rule{rule})
				svc := testService(serviceName, testNamespace, testServicePort)
				defer func() {
					deleteResource(apiRule)
					deleteResource(svc)
				}()

				// when
				Expect(c.Create(context.Background(), svc)).Should(Succeed())
				err := c.Create(context.Background(), apiRule)

				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("spec.gateway: Invalid value: \"string\": Gateway must be in the namespace/name format"))
			}

			It("should not create an APIRule with an empty gateway", func() {
				invalidHelper("")
			})

			It("should not create an APIRule with too long gateway namespace name", func() {
				invalidHelper("insane-very-long-namespace-name-exceeding-sixty-three-characters/validname")
			})

			It("should not create an APIRule with too long gateway name", func() {
				invalidHelper("validnamespace/insane-very-long-namespace-name-exceeding-sixty-three-characters")
			})

			It("should not create an APIRule with just the namespace", func() {
				invalidHelper("validnamespace/")
			})

			It("should not create an APIRule with just the gateway name", func() {
				invalidHelper("/validgateway")
			})

			It("should not create an APIRule with double slashed gateway name", func() {
				invalidHelper("namespace//gateway")
			})
		})

		Context("hosts should be a valid FQDN or a short host name", Ordered, func() {
			It("should create an APIRule with a valid FQDN host", func() {
				// given
				apiRuleName := generateTestName(testNameBase, testIDLength)
				serviceName := testServiceNameBase
				serviceHost := gatewayv2alpha1.Host("test.some-example.com")
				serviceHosts := []*gatewayv2alpha1.Host{&serviceHost}

				rule := testRulev2alpha1("/img", []gatewayv2alpha1.HttpMethod{http.MethodGet})
				rule.NoAuth = ptr.To(true)
				apiRule := testApiRulev2alpha1(apiRuleName, testNamespace, serviceName, testNamespace, serviceHosts, testServicePort, []gatewayv2alpha1.Rule{rule})
				svc := testService(serviceName, testNamespace, testServicePort)
				defer func() {
					deleteResource(apiRule)
					deleteResource(svc)
				}()

				// when
				Expect(c.Create(context.Background(), svc)).Should(Succeed())
				Expect(c.Create(context.Background(), apiRule)).Should(Succeed())

			})

			It("should create an APIRule with short host name that has length of 1 character", func() {
				// given
				apiRuleName := generateTestName(testNameBase, testIDLength)
				serviceName := testServiceNameBase
				serviceHost := gatewayv2alpha1.Host("a")
				serviceHosts := []*gatewayv2alpha1.Host{&serviceHost}

				rule := testRulev2alpha1("/img", []gatewayv2alpha1.HttpMethod{http.MethodGet})
				rule.NoAuth = ptr.To(true)
				apiRule := testApiRulev2alpha1(apiRuleName, testNamespace, serviceName, testNamespace, serviceHosts, testServicePort, []gatewayv2alpha1.Rule{rule})
				svc := testService(serviceName, testNamespace, testServicePort)
				defer func() {
					deleteResource(apiRule)
					deleteResource(svc)
				}()

				// when
				Expect(c.Create(context.Background(), svc)).Should(Succeed())
				Expect(c.Create(context.Background(), apiRule)).Should(Succeed())

			})

			It("should create an APIRule with host name that has 1 char labels and 2 chars top-level domain", func() {
				// given
				apiRuleName := generateTestName(testNameBase, testIDLength)
				serviceName := testServiceNameBase
				serviceHost := gatewayv2alpha1.Host("a.b.ca")
				serviceHosts := []*gatewayv2alpha1.Host{&serviceHost}

				rule := testRulev2alpha1("/img", []gatewayv2alpha1.HttpMethod{http.MethodGet})
				rule.NoAuth = ptr.To(true)
				apiRule := testApiRulev2alpha1(apiRuleName, testNamespace, serviceName, testNamespace, serviceHosts, testServicePort, []gatewayv2alpha1.Rule{rule})
				svc := testService(serviceName, testNamespace, testServicePort)
				defer func() {
					deleteResource(apiRule)
					deleteResource(svc)
				}()

				// when
				Expect(c.Create(context.Background(), svc)).Should(Succeed())
				Expect(c.Create(context.Background(), apiRule)).Should(Succeed())

			})

			It("should create an APIRule with host name that has length of 255 characters", func() {
				// given
				apiRuleName := generateTestName(testNameBase, testIDLength)
				serviceName := testServiceNameBase
				sixtyThreeA := strings.Repeat("a", 63)
				host255 := fmt.Sprintf("%s.%s.%s.%s.com", sixtyThreeA, sixtyThreeA, sixtyThreeA, strings.Repeat("b", 59))
				serviceHost := gatewayv2alpha1.Host(host255)
				serviceHosts := []*gatewayv2alpha1.Host{&serviceHost}

				rule := testRulev2alpha1("/img", []gatewayv2alpha1.HttpMethod{http.MethodGet})
				rule.NoAuth = ptr.To(true)
				apiRule := testApiRulev2alpha1(apiRuleName, testNamespace, serviceName, testNamespace, serviceHosts, testServicePort, []gatewayv2alpha1.Rule{rule})
				svc := testService(serviceName, testNamespace, testServicePort)
				defer func() {
					deleteResource(apiRule)
					deleteResource(svc)
				}()

				// when
				Expect(host255).To(HaveLen(255))
				Expect(c.Create(context.Background(), svc)).Should(Succeed())
				Expect(c.Create(context.Background(), apiRule)).Should(Succeed())
			})

			It("should not create an APIRule with host name longer than 255 characters", func() {
				// given
				apiRuleName := generateTestName(testNameBase, testIDLength)
				serviceName := testServiceNameBase
				sixtyThreeA := strings.Repeat("a", 63)
				host256 := fmt.Sprintf("%s.%s.%s.%s.com", sixtyThreeA, sixtyThreeA, sixtyThreeA, strings.Repeat("b", 60))
				serviceHost := gatewayv2alpha1.Host(host256)
				serviceHosts := []*gatewayv2alpha1.Host{&serviceHost}

				rule := testRulev2alpha1("/img", []gatewayv2alpha1.HttpMethod{http.MethodGet})
				rule.NoAuth = ptr.To(true)
				apiRule := testApiRulev2alpha1(apiRuleName, testNamespace, serviceName, testNamespace, serviceHosts, testServicePort, []gatewayv2alpha1.Rule{rule})
				svc := testService(serviceName, testNamespace, testServicePort)
				defer func() {
					deleteResource(apiRule)
					deleteResource(svc)
				}()

				// when
				Expect(host256).To(HaveLen(256))
				Expect(c.Create(context.Background(), svc)).Should(Succeed())
				err := c.Create(context.Background(), apiRule)

				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("spec.hosts[0]: Too long: may not be longer than 255"))
			})

			invalidHelper := func(host gatewayv2alpha1.Host) {
				// given
				apiRuleName := generateTestName(testNameBase, testIDLength)
				serviceName := testServiceNameBase
				serviceHosts := []*gatewayv2alpha1.Host{ptr.To(gatewayv2alpha1.Host(host))}

				rule := testRulev2alpha1("/img", []gatewayv2alpha1.HttpMethod{http.MethodGet})
				rule.NoAuth = ptr.To(true)
				apiRule := testApiRulev2alpha1(apiRuleName, testNamespace, serviceName, testNamespace, serviceHosts, testServicePort, []gatewayv2alpha1.Rule{rule})
				svc := testService(serviceName, testNamespace, testServicePort)
				defer func() {
					deleteResource(apiRule)
					deleteResource(svc)
				}()

				// when
				Expect(c.Create(context.Background(), svc)).Should(Succeed())
				err := c.Create(context.Background(), apiRule)

				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("spec.hosts[0]: Invalid value: \"string\": Host must be a lowercase RFC 1123 label (must consist of lowercase alphanumeric characters or '-', and must start and end with an lowercase alphanumeric character) or a fully qualified domain name"))
			}

			It("should not create an APIRule with an empty host", func() {
				invalidHelper("")
			})

			It("should not create an APIRule when host name has uppercase letters", func() {
				invalidHelper("eXample.com")
				invalidHelper("example.cOm")
			})

			It("should not create an APIRule with host label longer than 63 characters", func() {
				invalidHelper(gatewayv2alpha1.Host(strings.Repeat("a", 64) + ".com"))
				invalidHelper(gatewayv2alpha1.Host("example." + strings.Repeat("a", 64)))
			})

			It("should not create an APIRule when any domain label is empty", func() {
				invalidHelper(".com")
				invalidHelper("example..com")
				invalidHelper("example.")
			})

			It("should not create an APIRule when top level domain is too short", func() {
				invalidHelper("example.c")
			})

			It("should not create an APIRule when host contains wrong characters", func() {
				invalidHelper("*example.com")
				invalidHelper("exam*ple.com")
				invalidHelper("example*.com")
				invalidHelper("example.*com")
				invalidHelper("example.co*m")
				invalidHelper("example.com*")
			})

			It("should not create an APIRule when host starts or ends with a hyphen", func() {
				invalidHelper("-example.com")
				invalidHelper("example-.com")
				invalidHelper("example.-com")
				invalidHelper("example.com-")
			})
		})

		Context("rule path validation respected", func() {
			It("should fail when path consists of a path and *", func() {
				// given
				apiRuleName := generateTestName(testNameBase, testIDLength)
				serviceName := generateTestName(testServiceNameBase, testIDLength)
				serviceHost := gatewayv2alpha1.Host("example.com")
				serviceHosts := []*gatewayv2alpha1.Host{&serviceHost}

				rule := testRulev2alpha1("/img*", []gatewayv2alpha1.HttpMethod{http.MethodGet})
				rule.NoAuth = ptr.To(true)
				apiRule := testApiRulev2alpha1(apiRuleName, testNamespace, serviceName, testNamespace, serviceHosts, testServicePort, []gatewayv2alpha1.Rule{rule})
				svc := testService(serviceName, testNamespace, testServicePort)
				defer func() {
					deleteResource(apiRule)
					deleteResource(svc)
				}()
				Expect(c.Create(context.Background(), svc)).Should(Succeed())

				// when
				err := c.Create(context.Background(), apiRule)

				// then
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("spec.rules[0].path: Invalid value: \"/img*\": spec.rules[0].path"))

			})

			It("should apply APIRule when path contains only /*", func() {
				// given
				apiRuleName := generateTestName(testNameBase, testIDLength)
				serviceName := generateTestName(testServiceNameBase, testIDLength)
				serviceHost := gatewayv2alpha1.Host("example.com")
				serviceHosts := []*gatewayv2alpha1.Host{&serviceHost}

				rule := testRulev2alpha1("/*", []gatewayv2alpha1.HttpMethod{http.MethodGet})
				rule.NoAuth = ptr.To(true)
				apiRule := testApiRulev2alpha1(apiRuleName, testNamespace, serviceName, testNamespace, serviceHosts, testServicePort, []gatewayv2alpha1.Rule{rule})
				svc := testService(serviceName, testNamespace, testServicePort)
				defer func() {
					deleteResource(apiRule)
					deleteResource(svc)
				}()
				Expect(c.Create(context.Background(), svc)).Should(Succeed())

				// when then
				Expect(c.Create(context.Background(), apiRule)).Should(Succeed())

			})

			It("should apply APIRule when path contains no *", func() {
				// given
				apiRuleName := generateTestName(testNameBase, testIDLength)
				serviceName := generateTestName(testServiceNameBase, testIDLength)
				serviceHost := gatewayv2alpha1.Host("example.com")
				serviceHosts := []*gatewayv2alpha1.Host{&serviceHost}

				rule := testRulev2alpha1("/img-new/1", []gatewayv2alpha1.HttpMethod{http.MethodGet})
				rule.NoAuth = ptr.To(true)
				apiRule := testApiRulev2alpha1(apiRuleName, testNamespace, serviceName, testNamespace, serviceHosts, testServicePort, []gatewayv2alpha1.Rule{rule})
				svc := testService(serviceName, testNamespace, testServicePort)
				defer func() {
					deleteResource(apiRule)
					deleteResource(svc)
				}()
				Expect(c.Create(context.Background(), svc)).Should(Succeed())

				// when then
				Expect(c.Create(context.Background(), apiRule)).Should(Succeed())

			})
		})
	})

	It("APIRule in status Error should reconcile to status OK when root cause of error is fixed", func() {
		// given
		updateJwtHandlerTo(helpers.JWT_HANDLER_ISTIO)

		apiRuleName := generateTestName(testNameBase, testIDLength)
		serviceName := generateTestName(testServiceNameBase, testIDLength)
		serviceHost := fmt.Sprintf("%s.kyma.local", serviceName)
		vsName := generateTestName("duplicated-host-vs", testIDLength)

		By(fmt.Sprintf("Creating virtual service for host %s", serviceHost))
		vs := virtualService(vsName, serviceHost)
		defer func() {
			deleteResource(vs)
		}()
		Expect(c.Create(context.Background(), vs)).Should(Succeed())

		By("Verifying virtual service has been created")
		Eventually(func(g Gomega) {
			createdVs := networkingv1beta1.VirtualService{}
			g.Expect(c.Get(context.Background(), client.ObjectKey{Name: vsName, Namespace: testNamespace}, &createdVs)).Should(Succeed())
		}, eventuallyTimeout).Should(Succeed())

		apiRuleLabelMatcher := matchingLabelsFunc(apiRuleName, testNamespace)

		By("Creating APIRule")
		rule := testRule("/headers", methodsGet, nil, testIstioJWTHandler(testIssuer, testJwksUri))
		apiRule := testApiRule(apiRuleName, testNamespace, serviceName, testNamespace, serviceHost, testServicePort, []gatewayv1beta1.Rule{rule})
		svc := testService(serviceName, testNamespace, testServicePort)
		defer func() {
			deleteResource(apiRule)
			deleteResource(svc)
		}()
		// when
		Expect(c.Create(context.Background(), svc)).Should(Succeed())
		Expect(c.Create(context.Background(), apiRule)).Should(Succeed())

		expectApiRuleStatus(apiRuleName, gatewayv1beta1.StatusError)

		By("Verifying virtual service for APIRule has not been created")
		verifyVirtualServiceCount(c, apiRuleLabelMatcher, 0)

		By("Deleting existing virtual service with duplicated host configuration")
		deleteResource(vs)

		By("Waiting until APIRule is reconciled after error")
		expectApiRuleStatus(apiRuleName, gatewayv1beta1.StatusOK)

		By("Verifying virtual service for APIRule has been created")
		verifyVirtualServiceCount(c, apiRuleLabelMatcher, 1)
	})

	It("APIRule in status OK should reconcile to status ERROR when an", func() {
		// given
		updateJwtHandlerTo(helpers.JWT_HANDLER_ISTIO)

		apiRuleName := generateTestName(testNameBase, testIDLength)
		serviceName := generateTestName(testServiceNameBase, testIDLength)
		serviceHost := fmt.Sprintf("%s.kyma.local", serviceName)
		vsName := generateTestName("duplicated-host-vs", testIDLength)

		By(fmt.Sprintf("Creating APIRule with host %s", serviceHost))
		rule := testRule("/headers", methodsGet, nil, testIstioJWTHandler(testIssuer, testJwksUri))
		apiRule := testApiRule(apiRuleName, testNamespace, serviceName, testNamespace, serviceHost, testServicePort, []gatewayv1beta1.Rule{rule})
		svc := testService(serviceName, testNamespace, testServicePort)

		defer func() {
			deleteResource(apiRule)
			deleteResource(svc)
		}()

		// when
		Expect(c.Create(context.Background(), svc)).Should(Succeed())
		Expect(c.Create(context.Background(), apiRule)).Should(Succeed())

		expectApiRuleStatus(apiRuleName, gatewayv1beta1.StatusOK)

		By(fmt.Sprintf("Creating virtual service for host %s", serviceHost))
		vs := virtualService(vsName, serviceHost)
		defer func() {
			deleteResource(vs)
		}()
		Expect(c.Create(context.Background(), vs)).Should(Succeed())

		By("Verifying virtual service has been created")
		Eventually(func(g Gomega) {
			createdVs := networkingv1beta1.VirtualService{}
			g.Expect(c.Get(context.Background(), client.ObjectKey{Name: vsName, Namespace: testNamespace}, &createdVs)).Should(Succeed())
		}, eventuallyTimeout).Should(Succeed())

		By("Waiting until APIRule is reconciled after error")
		expectApiRuleStatus(apiRuleName, gatewayv1beta1.StatusError)

		By("Verifying APIRule status description")
		Eventually(func(g Gomega) {
			expectedApiRule := gatewayv1beta1.APIRule{}
			g.Expect(c.Get(context.Background(), client.ObjectKey{Name: apiRuleName, Namespace: testNamespace}, &expectedApiRule)).Should(Succeed())
			g.Expect(expectedApiRule.Status.APIRuleStatus).NotTo(BeNil())
			g.Expect(expectedApiRule.Status.APIRuleStatus.Description).To(ContainSubstring("This host is occupied by another Virtual Service"))
		}, eventuallyTimeout).Should(Succeed())
	})

	It("Should recreate resources created by reconciler when they are manually deleted", func() {
		updateJwtHandlerTo(helpers.JWT_HANDLER_ISTIO)

		apiRuleName := generateTestName(testNameBase, testIDLength)
		serviceName := testServiceNameBase
		serviceHost := "httpbin-recreate-resources.kyma.local"

		rule := testRule("/img", methodsGet, nil, testIstioJWTHandlerWithScopes(testIssuer, testJwksUri, []string{"scope-a"}))
		apiRule := testApiRule(apiRuleName, testNamespace, serviceName, testNamespace, serviceHost, testServicePort, []gatewayv1beta1.Rule{rule})
		svc := testService(serviceName, testNamespace, testServicePort)
		defer func() {
			deleteResource(apiRule)
			deleteResource(svc)
		}()

		// when
		Expect(c.Create(context.Background(), svc)).Should(Succeed())
		Expect(c.Create(context.Background(), apiRule)).Should(Succeed())

		expectApiRuleStatus(apiRuleName, gatewayv1beta1.StatusOK)

		apiRuleNameMatchingLabels := matchingLabelsFunc(apiRuleName, testNamespace)

		By("Verifying that resources are created")
		verifyVirtualServiceCount(c, apiRuleNameMatchingLabels, 1)
		verifyRequestAuthenticationCount(c, apiRuleNameMatchingLabels, 1)
		verifyAuthorizationPolicyCount(c, apiRuleNameMatchingLabels, 1)

		By("Deleting Virtual Service")
		Eventually(func(g Gomega) {
			vsList := networkingv1beta1.VirtualServiceList{}
			g.Expect(c.List(context.Background(), &vsList, apiRuleNameMatchingLabels)).Should(Succeed())
			g.Expect(c.Delete(context.Background(), vsList.Items[0])).Should(Succeed())
		}, eventuallyTimeout).Should(Succeed())

		By("Deleting Request Authentication")
		raList := securityv1beta1.RequestAuthenticationList{}
		Eventually(func(g Gomega) {
			g.Expect(c.List(context.Background(), &raList, apiRuleNameMatchingLabels)).Should(Succeed())
			g.Expect(c.Delete(context.Background(), raList.Items[0])).Should(Succeed())
		}, eventuallyTimeout).Should(Succeed())

		By("Deleting Authorization Policy")
		apList := securityv1beta1.AuthorizationPolicyList{}
		Eventually(func(g Gomega) {
			g.Expect(c.List(context.Background(), &apList, apiRuleNameMatchingLabels)).Should(Succeed())
			g.Expect(c.Delete(context.Background(), apList.Items[0])).Should(Succeed())
		}, eventuallyTimeout).Should(Succeed())

		By("Verifying deleted resources are recreated")
		verifyVirtualServiceCount(c, apiRuleNameMatchingLabels, 1)
		verifyRequestAuthenticationCount(c, apiRuleNameMatchingLabels, 1)
		verifyAuthorizationPolicyCount(c, apiRuleNameMatchingLabels, 1)
	})

	It("Should delete created resources when APIRule is deleted", func() {
		updateJwtHandlerTo(helpers.JWT_HANDLER_ISTIO)

		apiRuleName := generateTestName(testNameBase, testIDLength)
		serviceName := testServiceNameBase
		serviceHost := "httpbin-delete-resources.kyma.local"

		rule := testRule("/img", methodsGet, nil, testIstioJWTHandlerWithScopes(testIssuer, testJwksUri, []string{"scope-a"}))
		apiRule := testApiRule(apiRuleName, testNamespace, serviceName, testNamespace, serviceHost, testServicePort, []gatewayv1beta1.Rule{rule})
		svc := testService(serviceName, testNamespace, testServicePort)
		defer func() {
			deleteResource(apiRule)
			deleteResource(svc)
		}()
		// when
		Expect(c.Create(context.Background(), svc)).Should(Succeed())
		Expect(c.Create(context.Background(), apiRule)).Should(Succeed())

		expectApiRuleStatus(apiRuleName, gatewayv1beta1.StatusOK)
		apiRuleNameMatchingLabels := matchingLabelsFunc(apiRuleName, testNamespace)

		verifyVirtualServiceCount(c, apiRuleNameMatchingLabels, 1)
		verifyRequestAuthenticationCount(c, apiRuleNameMatchingLabels, 1)
		verifyAuthorizationPolicyCount(c, apiRuleNameMatchingLabels, 1)

		Expect(c.Delete(context.Background(), apiRule)).Should(Succeed())

		By("Verifying resources are deleted")
		verifyVirtualServiceCount(c, apiRuleNameMatchingLabels, 0)
		verifyRequestAuthenticationCount(c, apiRuleNameMatchingLabels, 0)
		verifyAuthorizationPolicyCount(c, apiRuleNameMatchingLabels, 0)
	})

	It("should update APIRule subresources when exposed service is updated", func() {
		updateJwtHandlerTo(helpers.JWT_HANDLER_ISTIO)

		apiRuleName := generateTestName(testNameBase, testIDLength)
		serviceName := testServiceNameBase
		serviceHost := "httpbin-recreate-resources.kyma.local"

		rule := testRule("/img", methodsGet, nil, testIstioJWTHandlerWithScopes(testIssuer, testJwksUri, []string{"scope-a"}))
		apiRule := testApiRule(apiRuleName, testNamespace, serviceName, testNamespace, serviceHost, testServicePort, []gatewayv1beta1.Rule{rule})
		svc := testService(serviceName, testNamespace, testServicePort)

		// when
		Expect(c.Create(ctx, svc)).Should(Succeed())
		Expect(c.Create(ctx, apiRule)).Should(Succeed())
		defer func() {
			deleteResource(apiRule)
			deleteResource(svc)
		}()

		expectApiRuleStatus(apiRuleName, gatewayv1beta1.StatusOK)
		ap := &securityv1beta1.AuthorizationPolicy{}
		Eventually(func() {
			aps := securityv1beta1.AuthorizationPolicyList{}
			Expect(c.List(ctx, &aps, matchingLabelsFunc(apiRule.Name, apiRule.Namespace))).Should(Succeed())
			Expect(aps.Items).To(HaveLen(1))
			ap = aps.Items[0]
		})
		By("Updating service")
		svc.Spec.Selector = map[string]string{
			"app": serviceName + "-updated",
		}
		Expect(c.Update(ctx, svc)).Should(Succeed())

		By("Verifying that resources are updated selectors")
		Eventually(func() {
			got := securityv1beta1.AuthorizationPolicy{}
			Expect(c.Get(ctx, client.ObjectKeyFromObject(ap), &got)).Should(Succeed())
			Expect(got.Spec.Selector.MatchLabels["app"]).To(Equal(serviceName + "-updated"))
		})
	})

	Context("check v2alpha1 stored version", func() {
		It("should fetch the APIRule when v1beta1 is the original-version and the spec is convertible", func() {
			updateJwtHandlerTo(helpers.JWT_HANDLER_ISTIO)

			By("Creating APIRule")

			rule1 := testRule("/rule1", methodsGet, defaultMutators, noConfigHandler("no_auth"))

			apiRuleName := generateTestName(testNameBase, testIDLength)
			serviceName := generateTestName(testServiceNameBase, testIDLength)
			serviceHost := fmt.Sprintf("%s.local.kyma.dev", serviceName)

			apiRule := testApiRule(apiRuleName, testNamespace, serviceName, testNamespace, serviceHost, testServicePort, []gatewayv1beta1.Rule{rule1})
			svc := testService(serviceName, testNamespace, testServicePort)
			defer func() {
				deleteResource(apiRule)
				deleteResource(svc)
			}()
			// when
			Expect(c.Create(context.Background(), svc)).Should(Succeed())
			Expect(c.Create(context.Background(), apiRule)).Should(Succeed())

			expectApiRuleStatus(apiRuleName, gatewayv1beta1.StatusOK)

			By("Verifying APIRule fetched in version v1beta1")
			Eventually(func(g Gomega) {
				expectedApiRule := gatewayv1beta1.APIRule{}
				g.Expect(c.Get(context.Background(), client.ObjectKey{Name: apiRuleName, Namespace: testNamespace}, &expectedApiRule)).Should(Succeed())
				g.Expect(expectedApiRule.Status.APIRuleStatus.Code).To(Equal(gatewayv1beta1.StatusOK))
				g.Expect(expectedApiRule.ObjectMeta.Annotations).To(HaveKeyWithValue("gateway.kyma-project.io/original-version", "v1beta1"))
				g.Expect(expectedApiRule.ObjectMeta.Annotations).To(HaveKey("gateway.kyma-project.io/v1beta1-spec"))

				ruleV1beta1 := gatewayv1beta1.APIRule{}
				err := json.Unmarshal([]byte(expectedApiRule.Annotations["gateway.kyma-project.io/v1beta1-spec"]), &ruleV1beta1.Spec)
				g.Expect(err).To(BeNil())
				g.Expect(expectedApiRule.Spec).To(Equal(ruleV1beta1.Spec))
			}, eventuallyTimeout).Should(Succeed())

			By("Verifying APIRule fetched in version v2alpha1")
			Eventually(func(g Gomega) {
				expectedApiRule := gatewayv2alpha1.APIRule{}
				g.Expect(c.Get(context.Background(), client.ObjectKey{Name: apiRuleName, Namespace: testNamespace}, &expectedApiRule)).Should(Succeed())
				g.Expect(expectedApiRule.Status.State).To(Equal(gatewayv2alpha1.Ready))
				g.Expect(expectedApiRule.ObjectMeta.Annotations).To(HaveKeyWithValue("gateway.kyma-project.io/original-version", "v1beta1"))
				g.Expect(expectedApiRule.ObjectMeta.Annotations).To(HaveKey("gateway.kyma-project.io/v1beta1-spec"))

				ruleV1beta1 := gatewayv1beta1.APIRule{}
				err := json.Unmarshal([]byte(expectedApiRule.Annotations["gateway.kyma-project.io/v1beta1-spec"]), &ruleV1beta1.Spec)
				g.Expect(err).To(BeNil())
				g.Expect(string(*expectedApiRule.Spec.Hosts[0])).To(Equal(*ruleV1beta1.Spec.Host))
				g.Expect(expectedApiRule.Spec.Gateway).To(Equal(ruleV1beta1.Spec.Gateway))
				g.Expect(expectedApiRule.Spec.Rules).To(HaveLen(len(ruleV1beta1.Spec.Rules)))
			}, eventuallyTimeout).Should(Succeed())

			By("Verifying APIRule fetched in version v2")
			Eventually(func(g Gomega) {
				expectedApiRule := gatewayv2.APIRule{}
				g.Expect(c.Get(context.Background(), client.ObjectKey{Name: apiRuleName, Namespace: testNamespace}, &expectedApiRule)).Should(Succeed())
				g.Expect(expectedApiRule.Status.State).To(Equal(gatewayv2.Ready))
				g.Expect(expectedApiRule.ObjectMeta.Annotations).To(HaveKeyWithValue("gateway.kyma-project.io/original-version", "v1beta1"))
				g.Expect(expectedApiRule.ObjectMeta.Annotations).To(HaveKey("gateway.kyma-project.io/v1beta1-spec"))

				ruleV1beta1 := gatewayv1beta1.APIRule{}
				err := json.Unmarshal([]byte(expectedApiRule.Annotations["gateway.kyma-project.io/v1beta1-spec"]), &ruleV1beta1.Spec)
				g.Expect(err).To(BeNil())
				g.Expect(string(*expectedApiRule.Spec.Hosts[0])).To(Equal(*ruleV1beta1.Spec.Host))
				g.Expect(expectedApiRule.Spec.Gateway).To(Equal(ruleV1beta1.Spec.Gateway))
				g.Expect(expectedApiRule.Spec.Rules).To(HaveLen(len(ruleV1beta1.Spec.Rules)))
			}, eventuallyTimeout).Should(Succeed())
		})

		It("should fetch empty spec for the APIRule v2alpha1 and v2 when v1beta1 is the original-version and the spec is not convertible", func() {
			updateJwtHandlerTo(helpers.JWT_HANDLER_ISTIO)

			By("Creating APIRule")
			rule1 := testRule("/rule1", methodsGet, defaultMutators, noConfigHandler("allow"))
			rule2 := testRule("/.*", methodsGet, defaultMutators, noConfigHandler("allow"))

			apiRuleName := generateTestName(testNameBase, testIDLength)
			serviceName := generateTestName(testServiceNameBase, testIDLength)
			serviceHost := fmt.Sprintf("%s.local.kyma.dev", serviceName)

			apiRule := testApiRule(apiRuleName, testNamespace, serviceName, testNamespace, serviceHost, testServicePort, []gatewayv1beta1.Rule{rule1, rule2})
			svc := testService(serviceName, testNamespace, testServicePort)
			defer func() {
				deleteResource(apiRule)
				deleteResource(svc)
			}()
			// when
			Expect(c.Create(context.Background(), svc)).Should(Succeed())
			Expect(c.Create(context.Background(), apiRule)).Should(Succeed())

			expectApiRuleStatus(apiRuleName, gatewayv1beta1.StatusOK)

			By("Verifying APIRule fetched in version v2alpha1")
			Eventually(func(g Gomega) {
				expectedApiRule := gatewayv2alpha1.APIRule{}
				g.Expect(c.Get(context.Background(), client.ObjectKey{Name: apiRuleName, Namespace: testNamespace}, &expectedApiRule)).Should(Succeed())
				g.Expect(expectedApiRule.Status.State).To(Equal(gatewayv2alpha1.Ready))
				g.Expect(expectedApiRule.Spec.Hosts).To(BeNil())
				g.Expect(expectedApiRule.Spec.Rules).To(BeNil())
			}, eventuallyTimeout).Should(Succeed())

			By("Verifying APIRule fetched in version v2")
			Eventually(func(g Gomega) {
				expectedApiRule := gatewayv2.APIRule{}
				g.Expect(c.Get(context.Background(), client.ObjectKey{Name: apiRuleName, Namespace: testNamespace}, &expectedApiRule)).Should(Succeed())
				g.Expect(expectedApiRule.Status.State).To(Equal(gatewayv2.Ready))
				g.Expect(expectedApiRule.Spec.Hosts).To(BeNil())
				g.Expect(expectedApiRule.Spec.Rules).To(BeNil())
			}, eventuallyTimeout).Should(Succeed())
		})

		It("should fetch the APIRule when v2alpha1 is the original-version and the spec is convertible", func() {
			updateJwtHandlerTo(helpers.JWT_HANDLER_ISTIO)

			By("Creating APIRule with gateway")

			gateway := testGateway()
			rule1 := testRulev2alpha1("/rule1", v2alpha1methodsGet)
			rule1.NoAuth = ptr.To(true)

			apiRuleName := generateTestName(testNameBase, testIDLength)
			serviceName := generateTestName(testServiceNameBase, testIDLength)
			serviceHost := gatewayv2alpha1.Host(fmt.Sprintf("%s.local.kyma.dev", serviceName))
			serviceHosts := []*gatewayv2alpha1.Host{&serviceHost}

			apiRule := testApiRulev2alpha1(apiRuleName, testNamespace, serviceName, testNamespace, serviceHosts, testServicePort, []gatewayv2alpha1.Rule{rule1})
			svc := testService(serviceName, testNamespace, testServicePort)
			defer func() {
				deleteResource(&gateway)
				deleteResource(apiRule)
				deleteResource(svc)
			}()

			// when
			Expect(c.Create(context.Background(), &gateway)).Should(Succeed())
			Expect(c.Create(context.Background(), svc)).Should(Succeed())
			Expect(c.Create(context.Background(), apiRule)).Should(Succeed())

			expectV2alpha1ApiRuleStatus(apiRuleName, gatewayv2alpha1.Ready)

			By("Verifying APIRule fetched in version v2")
			Eventually(func(g Gomega) {
				expectedApiRule := gatewayv2.APIRule{}
				g.Expect(c.Get(context.Background(), client.ObjectKey{Name: apiRuleName, Namespace: testNamespace}, &expectedApiRule)).Should(Succeed())
				g.Expect(expectedApiRule.Status.State).To(Equal(gatewayv2.Ready))
				g.Expect(expectedApiRule.ObjectMeta.Annotations).To(HaveKeyWithValue("gateway.kyma-project.io/original-version", "v2alpha1"))
				g.Expect(expectedApiRule.Spec.Hosts).To(HaveLen(1))
				g.Expect(*expectedApiRule.Spec.Gateway).To(Equal(testGatewayURL))
			}, eventuallyTimeout).Should(Succeed())

			By("Verifying APIRule fetched in version v2alpha1")
			Eventually(func(g Gomega) {
				expectedApiRule := gatewayv2alpha1.APIRule{}
				g.Expect(c.Get(context.Background(), client.ObjectKey{Name: apiRuleName, Namespace: testNamespace}, &expectedApiRule)).Should(Succeed())
				g.Expect(expectedApiRule.Status.State).To(Equal(gatewayv2alpha1.Ready))
				g.Expect(expectedApiRule.ObjectMeta.Annotations).To(HaveKeyWithValue("gateway.kyma-project.io/original-version", "v2alpha1"))
				g.Expect(expectedApiRule.Spec.Hosts).To(HaveLen(1))
				g.Expect(*expectedApiRule.Spec.Gateway).To(Equal(testGatewayURL))
			}, eventuallyTimeout).Should(Succeed())
		})

		It("should fetch empty spec for the APIRule v1beta1 when v2alpha1 is the original-version and the spec is not convertible", func() {
			updateJwtHandlerTo(helpers.JWT_HANDLER_ISTIO)

			By("Creating APIRule with gateway")

			gateway := testGateway()
			rule1 := testRulev2alpha1("/rule1", v2alpha1methodsGet)
			rule1.NoAuth = ptr.To(true)

			apiRuleName := generateTestName(testNameBase, testIDLength)
			serviceName := generateTestName(testServiceNameBase, testIDLength)
			serviceHost := gatewayv2alpha1.Host(fmt.Sprintf("%s.local.kyma.dev", serviceName))
			serviceHosts := []*gatewayv2alpha1.Host{&serviceHost}

			apiRule := testApiRulev2alpha1(apiRuleName, testNamespace, serviceName, testNamespace, serviceHosts, testServicePort, []gatewayv2alpha1.Rule{rule1})
			svc := testService(serviceName, testNamespace, testServicePort)
			defer func() {
				deleteResource(&gateway)
				deleteResource(apiRule)
				deleteResource(svc)
			}()

			// when
			Expect(c.Create(context.Background(), &gateway)).Should(Succeed())
			Expect(c.Create(context.Background(), svc)).Should(Succeed())
			Expect(c.Create(context.Background(), apiRule)).Should(Succeed())

			expectV2alpha1ApiRuleStatus(apiRuleName, gatewayv2alpha1.Ready)

			By("Verifying APIRule fetched in version v1beta1")
			Eventually(func(g Gomega) {
				expectedApiRule := gatewayv1beta1.APIRule{}
				g.Expect(c.Get(context.Background(), client.ObjectKey{Name: apiRuleName, Namespace: testNamespace}, &expectedApiRule)).Should(Succeed())
				g.Expect(expectedApiRule.Status.APIRuleStatus.Code).To(Equal(gatewayv1beta1.StatusOK))
				g.Expect(expectedApiRule.Spec.Host).To(BeNil())
				g.Expect(expectedApiRule.Spec.Gateway).To(BeNil())
			}, eventuallyTimeout).Should(Succeed())
		})

		It("should fetch the APIRule when v2 is the original-version and the spec is convertible", func() {
			updateJwtHandlerTo(helpers.JWT_HANDLER_ISTIO)

			By("Creating APIRule with gateway")

			gateway := testGateway()
			rule1 := testRulev2("/rule1", v2methodsGet)
			rule1.NoAuth = ptr.To(true)

			apiRuleName := generateTestName(testNameBase, testIDLength)
			serviceName := generateTestName(testServiceNameBase, testIDLength)
			serviceHost := gatewayv2.Host(fmt.Sprintf("%s.local.kyma.dev", serviceName))
			serviceHosts := []*gatewayv2.Host{&serviceHost}

			apiRule := testApiRulev2(apiRuleName, testNamespace, serviceName, testNamespace, serviceHosts, testServicePort, []gatewayv2.Rule{rule1})
			svc := testService(serviceName, testNamespace, testServicePort)
			defer func() {
				deleteResource(&gateway)
				deleteResource(apiRule)
				deleteResource(svc)
			}()

			// when
			Expect(c.Create(context.Background(), &gateway)).Should(Succeed())
			Expect(c.Create(context.Background(), svc)).Should(Succeed())
			Expect(c.Create(context.Background(), apiRule)).Should(Succeed())

			expectV2ApiRuleStatus(apiRuleName, gatewayv2.Ready)

			By("Verifying APIRule fetched in version v2alpha1")
			Eventually(func(g Gomega) {
				expectedApiRule := gatewayv2alpha1.APIRule{}
				g.Expect(c.Get(context.Background(), client.ObjectKey{Name: apiRuleName, Namespace: testNamespace}, &expectedApiRule)).Should(Succeed())
				g.Expect(expectedApiRule.Status.State).To(Equal(gatewayv2alpha1.Ready))
				g.Expect(expectedApiRule.Spec.Hosts).To(HaveLen(1))
				g.Expect(*expectedApiRule.Spec.Gateway).To(Equal(testGatewayURL))
			}, eventuallyTimeout).Should(Succeed())

			By("Verifying APIRule fetched in version v2")
			Eventually(func(g Gomega) {
				expectedApiRule := gatewayv2.APIRule{}
				g.Expect(c.Get(context.Background(), client.ObjectKey{Name: apiRuleName, Namespace: testNamespace}, &expectedApiRule)).Should(Succeed())
				g.Expect(expectedApiRule.Status.State).To(Equal(gatewayv2.Ready))
				g.Expect(expectedApiRule.Spec.Hosts).To(HaveLen(1))
				g.Expect(*expectedApiRule.Spec.Gateway).To(Equal(testGatewayURL))
			}, eventuallyTimeout).Should(Succeed())
		})

		It("should fetch empty spec for the APIRule v1beta1 when v2 is the original-version and the spec is not convertible", func() {
			updateJwtHandlerTo(helpers.JWT_HANDLER_ISTIO)

			By("Creating APIRule with gateway")

			gateway := testGateway()
			rule1 := testRulev2("/rule1", v2methodsGet)
			rule1.NoAuth = ptr.To(true)

			apiRuleName := generateTestName(testNameBase, testIDLength)
			serviceName := generateTestName(testServiceNameBase, testIDLength)
			serviceHost := gatewayv2.Host(fmt.Sprintf("%s.local.kyma.dev", serviceName))
			serviceHosts := []*gatewayv2.Host{&serviceHost}

			apiRule := testApiRulev2(apiRuleName, testNamespace, serviceName, testNamespace, serviceHosts, testServicePort, []gatewayv2.Rule{rule1})
			svc := testService(serviceName, testNamespace, testServicePort)
			defer func() {
				deleteResource(&gateway)
				deleteResource(apiRule)
				deleteResource(svc)
			}()

			// when
			Expect(c.Create(context.Background(), &gateway)).Should(Succeed())
			Expect(c.Create(context.Background(), svc)).Should(Succeed())
			Expect(c.Create(context.Background(), apiRule)).Should(Succeed())

			expectV2ApiRuleStatus(apiRuleName, gatewayv2.Ready)

			By("Verifying APIRule fetched in version v1beta1")
			Eventually(func(g Gomega) {
				expectedApiRule := gatewayv1beta1.APIRule{}
				g.Expect(c.Get(context.Background(), client.ObjectKey{Name: apiRuleName, Namespace: testNamespace}, &expectedApiRule)).Should(Succeed())
				g.Expect(expectedApiRule.Status.APIRuleStatus.Code).To(Equal(gatewayv1beta1.StatusOK))
				g.Expect(expectedApiRule.Spec.Host).To(BeNil())
				g.Expect(expectedApiRule.Spec.Gateway).To(BeNil())
			}, eventuallyTimeout).Should(Succeed())
		})
	})
})

func verifyVirtualServiceCount(c client.Client, option client.ListOption, count int) {
	By(fmt.Sprintf("Verifying %d Virtual Service exist", count))
	Eventually(func(g Gomega) {
		vsList := networkingv1beta1.VirtualServiceList{}
		g.Expect(c.List(context.Background(), &vsList, option)).Should(Succeed())
		g.Expect(vsList.Items).To(HaveLen(count))
	}, eventuallyTimeout).Should(Succeed())
}

func verifyRequestAuthenticationCount(c client.Client, option client.ListOption, count int) {
	By(fmt.Sprintf("Verifying %d Request Authentication exist", count))
	Eventually(func(g Gomega) {
		raList := securityv1beta1.RequestAuthenticationList{}
		g.Expect(c.List(context.Background(), &raList, option)).Should(Succeed())
		g.Expect(raList.Items).To(HaveLen(count))
	}, eventuallyTimeout).Should(Succeed())
}

func verifyAuthorizationPolicyCount(c client.Client, option client.ListOption, count int) {
	By(fmt.Sprintf("Verifying %d Authorization Policy exist", count))
	Eventually(func(g Gomega) {
		apList := securityv1beta1.AuthorizationPolicyList{}
		g.Expect(c.List(context.Background(), &apList, option)).Should(Succeed())
		g.Expect(apList.Items).To(HaveLen(count))
	}, eventuallyTimeout).Should(Succeed())
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

func testRule(path string, methods []gatewayv1beta1.HttpMethod, mutators []*gatewayv1beta1.Mutator, handler *gatewayv1beta1.Handler) gatewayv1beta1.Rule {
	return gatewayv1beta1.Rule{
		Path:     path,
		Methods:  methods,
		Mutators: mutators,
		AccessStrategies: []*gatewayv1beta1.Authenticator{
			{
				Handler: handler,
			},
		},
	}
}

func testRulev2alpha1(path string, methods []gatewayv2alpha1.HttpMethod) gatewayv2alpha1.Rule {
	return gatewayv2alpha1.Rule{
		Path:    path,
		Methods: methods,
	}
}

func testRulev2(path string, methods []gatewayv2.HttpMethod) gatewayv2.Rule {
	return gatewayv2.Rule{
		Path:    path,
		Methods: methods,
	}
}

func testApiRule(name, namespace, serviceName, serviceNamespace, serviceHost string, servicePort uint32, rules []gatewayv1beta1.Rule) *gatewayv1beta1.APIRule {
	var gateway = testGatewayURL

	return &gatewayv1beta1.APIRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: gatewayv1beta1.APIRuleSpec{
			Host:    &serviceHost,
			Gateway: &gateway,
			Service: &gatewayv1beta1.Service{
				Name:      &serviceName,
				Namespace: &serviceNamespace,
				Port:      &servicePort,
			},
			Rules: rules,
		},
		Status: gatewayv1beta1.APIRuleStatus{
			APIRuleStatus: nil,
		},
	}
}

func testApiRulev2alpha1(name, namespace, serviceName, serviceNamespace string, serviceHosts []*gatewayv2alpha1.Host, servicePort uint32, rules []gatewayv2alpha1.Rule) *gatewayv2alpha1.APIRule {
	var gateway = testGatewayURL

	return &gatewayv2alpha1.APIRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: gatewayv2alpha1.APIRuleSpec{
			Hosts:   serviceHosts,
			Gateway: &gateway,
			Service: &gatewayv2alpha1.Service{
				Name:      &serviceName,
				Namespace: &serviceNamespace,
				Port:      &servicePort,
			},
			Rules: rules,
		},
	}
}

func testApiRulev2(name, namespace, serviceName, serviceNamespace string, serviceHosts []*gatewayv2.Host, servicePort uint32, rules []gatewayv2.Rule) *gatewayv2.APIRule {
	var gateway = testGatewayURL

	return &gatewayv2.APIRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: gatewayv2.APIRuleSpec{
			Hosts:   serviceHosts,
			Gateway: &gateway,
			Service: &gatewayv2.Service{
				Name:      &serviceName,
				Namespace: &serviceNamespace,
				Port:      &servicePort,
			},
			Rules: rules,
		},
	}
}

func testApiRulev2alpha1Gateway(name, namespace, serviceName, serviceNamespace, gateway string, servicePort uint32, rules []gatewayv2alpha1.Rule) *gatewayv2alpha1.APIRule {
	serviceHost := gatewayv2alpha1.Host("example.com")
	serviceHosts := []*gatewayv2alpha1.Host{&serviceHost}

	return &gatewayv2alpha1.APIRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: gatewayv2alpha1.APIRuleSpec{
			Hosts:   serviceHosts,
			Gateway: &gateway,
			Service: &gatewayv2alpha1.Service{
				Name:      &serviceName,
				Namespace: &serviceNamespace,
				Port:      &servicePort,
			},
			Rules: rules,
		},
	}
}

func testService(name, namespace string, servicePort uint32) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": name,
			},
			Ports: []corev1.ServicePort{
				{
					Name: "http",
					Port: int32(servicePort),
				},
			},
		},
	}
}

func testOryJWTHandler(issuer string, scopes []string) *gatewayv1beta1.Handler {
	configJSON := fmt.Sprintf(`{
			"trusted_issuers": ["%s"],
			"jwks": [],
			"required_scope": [%s]
		}`, issuer, toCSVList(scopes))

	return &gatewayv1beta1.Handler{
		Name: "jwt",
		Config: &runtime.RawExtension{
			Raw: []byte(configJSON),
		},
	}
}

func testIstioJWTHandler(issuer string, jwksUri string) *gatewayv1beta1.Handler {
	bytes, err := json.Marshal(gatewayv1beta1.JwtConfig{
		Authentications: []*gatewayv1beta1.JwtAuthentication{
			{
				Issuer:  issuer,
				JwksUri: jwksUri,
			},
		},
	})
	Expect(err).To(BeNil())
	return &gatewayv1beta1.Handler{
		Name: "jwt",
		Config: &runtime.RawExtension{
			Raw: bytes,
		},
	}
}

func testIstioJWTHandlerWithScopes(issuer string, jwksUri string, authorizationScopes []string) *gatewayv1beta1.Handler {
	bytes, err := json.Marshal(gatewayv1beta1.JwtConfig{
		Authentications: []*gatewayv1beta1.JwtAuthentication{
			{
				Issuer:  issuer,
				JwksUri: jwksUri,
			},
		},
		Authorizations: []*gatewayv1beta1.JwtAuthorization{
			{
				RequiredScopes: authorizationScopes,
			},
		},
	})
	Expect(err).To(BeNil())
	return &gatewayv1beta1.Handler{
		Name: "jwt",
		Config: &runtime.RawExtension{
			Raw: bytes,
		},
	}
}

func testIstioJWTHandlerWithAuthorizations(issuer string, jwksUri string, authorizations []*gatewayv1beta1.JwtAuthorization) *gatewayv1beta1.Handler {
	bytes, err := json.Marshal(gatewayv1beta1.JwtConfig{
		Authentications: []*gatewayv1beta1.JwtAuthentication{
			{
				Issuer:  issuer,
				JwksUri: jwksUri,
			},
		},
		Authorizations: authorizations,
	})
	Expect(err).To(BeNil())
	return &gatewayv1beta1.Handler{
		Name: "jwt",
		Config: &runtime.RawExtension{
			Raw: bytes,
		},
	}
}

func testOauthHandler(scopes []string) *gatewayv1beta1.Handler {
	configJSON := fmt.Sprintf(`{
		"required_scope": [%s]
	}`, toCSVList(scopes))

	return &gatewayv1beta1.Handler{
		Name: "oauth2_introspection",
		Config: &runtime.RawExtension{
			Raw: []byte(configJSON),
		},
	}
}

func testGateway() networkingv1beta1.Gateway {
	return networkingv1beta1.Gateway{
		ObjectMeta: metav1.ObjectMeta{Name: "kyma-gateway", Namespace: "kyma-system"},
		Spec: apinetworkingv1beta1.Gateway{
			Servers: []*apinetworkingv1beta1.Server{
				{
					Port: &apinetworkingv1beta1.Port{
						Protocol: "HTTPS",
					},
					Hosts: []string{
						"*.local.kyma.dev",
					},
				},
				{
					Port: &apinetworkingv1beta1.Port{
						Protocol: "HTTP",
					},
					Hosts: []string{
						"*.local.kyma.dev",
					},
				},
			},
		},
	}
}

func noConfigHandler(name string) *gatewayv1beta1.Handler {
	return &gatewayv1beta1.Handler{
		Name: name,
	}
}

// Converts a []interface{} to a string slice. Panics if given object is of other type.
func asStringSlice(in interface{}) []string {
	inSlice := in.([]interface{})

	if inSlice == nil {
		return nil
	}

	res := []string{}

	for _, v := range inSlice {
		res = append(res, v.(string))
	}

	return res
}

func generateTestName(name string, length int) string {
	rand.NewSource(time.Now().UnixNano())

	letterRunes := []rune("abcdefghijklmnopqrstuvwxyz")

	b := make([]rune, length)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return name + "-" + string(b)
}

func getRuleList(g Gomega, matchingLabels client.ListOption) []rulev1alpha1.Rule {
	res := rulev1alpha1.RuleList{}
	g.Expect(c.List(context.Background(), &res, matchingLabels)).Should(Succeed())
	return res.Items
}

func verifyRuleList(g Gomega, ruleList []rulev1alpha1.Rule, pathToURLFunc func(string) string, expected ...gatewayv1beta1.Rule) {
	g.Expect(ruleList).To(HaveLen(len(expected)))

	actual := make(map[string]rulev1alpha1.Rule)

	for _, rule := range ruleList {
		actual[rule.Spec.Match.URL] = rule
	}

	for i := range expected {
		ruleUrl := pathToURLFunc(expected[i].Path)
		g.Expect(actual[ruleUrl]).ToNot(BeNil())
		g.Expect(actual[ruleUrl].Spec.Match).ToNot(BeNil())
		g.Expect(actual[ruleUrl].Spec.Match.Methods).To(Equal(gatewayv1beta1.ConvertHttpMethodsToStrings(expected[i].Methods)))
		verifyAccessStrategies(g, actual[ruleUrl].Spec.Authenticators, expected[i].AccessStrategies)
		verifyMutators(g, actual[ruleUrl].Spec.Mutators, expected[i].Mutators)
	}
}

func verifyMutators(g Gomega, actual []*rulev1alpha1.Mutator, expected []*gatewayv1beta1.Mutator) {
	if expected == nil {
		g.Expect(actual).To(BeNil())
	} else {
		for i := 0; i < len(expected); i++ {
			verifyHandler(g, actual[i].Handler, expected[i].Handler)
		}
	}
}

func verifyAccessStrategies(g Gomega, actual []*rulev1alpha1.Authenticator, expected []*gatewayv1beta1.Authenticator) {
	if expected == nil {
		g.Expect(actual).To(BeNil())
	} else {
		for i := 0; i < len(expected); i++ {
			verifyHandler(g, actual[i].Handler, expected[i].Handler)
		}
	}
}

func verifyHandler(g Gomega, actual *rulev1alpha1.Handler, expected *gatewayv1beta1.Handler) {
	if expected == nil {
		g.Expect(actual).To(BeNil())
	} else {
		g.Expect(actual.Name).To(Equal(expected.Name))
		g.Expect(actual.Config).To(Equal(expected.Config))
	}
}

func matchingLabelsFunc(apiRuleName, namespace string) client.ListOption {
	labels := make(map[string]string)
	labels[processing.OwnerLabel] = fmt.Sprintf("%s.%s", apiRuleName, namespace)
	return client.MatchingLabels(labels)
}

func updateJwtHandlerTo(jwtHandler string) {
	cm := &corev1.ConfigMap{}
	Expect(c.Get(context.Background(), client.ObjectKey{Name: helpers.CM_NAME, Namespace: helpers.CM_NS}, cm)).Should(Succeed())

	if !strings.Contains(cm.Data[helpers.CM_KEY], jwtHandler) {
		By(fmt.Sprintf("Updating JWT handler config map to %s", jwtHandler))
		cm.Data = map[string]string{
			helpers.CM_KEY: fmt.Sprintf("jwtHandler: %s", jwtHandler),
		}
		Expect(c.Update(context.Background(), cm)).To(Succeed())

		By("Waiting until config map is updated")
		Eventually(func(g Gomega) {
			g.Expect(c.Get(context.Background(), client.ObjectKey{Name: cm.Name, Namespace: cm.Namespace}, cm)).Should(Succeed())
			g.Expect(cm.Data).To(HaveKeyWithValue(helpers.CM_KEY, fmt.Sprintf("jwtHandler: %s", jwtHandler)))
		}, eventuallyTimeout).Should(Succeed())
	}
}

func getAuthorizationPolicyWhenScopeMatcher(firstScope, secondScope string) gomegatypes.GomegaMatcher {
	var whenMatchers []gomegatypes.GomegaMatcher

	for _, key := range []string{"request.auth.claims[scp]", "request.auth.claims[scope]", "request.auth.claims[scopes]"} {
		matcher := PointTo(MatchFields(IgnoreExtras, Fields{
			"When": ContainElements(
				PointTo(MatchFields(IgnoreExtras, Fields{
					"Key":    Equal(key),
					"Values": ContainElement(firstScope),
				})),
				PointTo(MatchFields(IgnoreExtras, Fields{
					"Key":    Equal(key),
					"Values": ContainElement(secondScope),
				})),
			),
		}))
		whenMatchers = append(whenMatchers, matcher)
	}

	return PointTo(MatchFields(IgnoreExtras,
		Fields{
			"Spec": MatchFields(IgnoreExtras, Fields{
				"Rules": ContainElements(whenMatchers),
			}),
		}))
}

func deleteResource(object client.Object) {
	By(fmt.Sprintf("Deleting resource %s as part of teardown", object.GetName()))
	err := c.Delete(context.Background(), object)

	if err != nil {
		Expect(errors.IsNotFound(err)).To(BeTrue())
	}

	Eventually(func(g Gomega) {
		err := c.Get(context.Background(), client.ObjectKeyFromObject(object), object)
		g.Expect(errors.IsNotFound(err)).To(BeTrue())
	}, eventuallyTimeout).Should(Succeed())
}

func expectApiRuleStatus(apiRuleName string, statusCode gatewayv1beta1.StatusCode) {
	By(fmt.Sprintf("Verifying that ApiRule v1beta1 %s has status %s", apiRuleName, statusCode))
	Eventually(func(g Gomega) {
		expectedApiRule := gatewayv1beta1.APIRule{}
		g.Expect(func() error {
			err := c.Get(context.Background(), client.ObjectKey{Name: apiRuleName, Namespace: testNamespace}, &expectedApiRule)
			if err != nil {
				fmt.Println("Error getting APIRule:", err)
			}
			return err
		}()).Should(Succeed())
		g.Expect(expectedApiRule.Status.APIRuleStatus).NotTo(BeNil())
		g.Expect(expectedApiRule.Status.APIRuleStatus.Code).To(Equal(statusCode))
	}, eventuallyTimeout).Should(Succeed())
}

func expectV2alpha1ApiRuleStatus(apiRuleName string, state gatewayv2alpha1.State) {
	By(fmt.Sprintf("Verifying that ApiRule v2alpha1 %s has status %s", apiRuleName, state))
	Eventually(func(g Gomega) {
		expectedApiRule := gatewayv2alpha1.APIRule{}
		g.Expect(func() error {
			err := c.Get(context.Background(), client.ObjectKey{Name: apiRuleName, Namespace: testNamespace}, &expectedApiRule)
			if err != nil {
				fmt.Println("Error getting APIRule:", err)
			}
			return err
		}()).Should(Succeed())
		g.Expect(expectedApiRule.Status).NotTo(BeNil())
		g.Expect(expectedApiRule.Status.State).To(Equal(state))
	}, eventuallyTimeout).Should(Succeed())
}

func expectV2ApiRuleStatus(apiRuleName string, state gatewayv2.State) {
	By(fmt.Sprintf("Verifying that ApiRule v2alpha1 %s has status %s", apiRuleName, state))
	Eventually(func(g Gomega) {
		expectedApiRule := gatewayv2.APIRule{}
		g.Expect(func() error {
			err := c.Get(context.Background(), client.ObjectKey{Name: apiRuleName, Namespace: testNamespace}, &expectedApiRule)
			if err != nil {
				fmt.Println("Error getting APIRule:", err)
			}
			return err
		}()).Should(Succeed())
		g.Expect(expectedApiRule.Status).NotTo(BeNil())
		g.Expect(expectedApiRule.Status.State).To(Equal(state))
	}, eventuallyTimeout).Should(Succeed())
}

func virtualService(name string, host string) *networkingv1beta1.VirtualService {
	vs := &networkingv1beta1.VirtualService{}
	vs.ObjectMeta = metav1.ObjectMeta{
		Name:      name,
		Namespace: testNamespace,
	}
	vs.Spec.Hosts = []string{host}

	return vs
}
