package gateway_test

import (
	"context"
	"fmt"
	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"math/rand"
	"strings"
	"time"

	gomegatypes "github.com/onsi/gomega/types"
	"k8s.io/apimachinery/pkg/api/errors"

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

	Context("when creating an APIRule for exposing service", func() {

		Context("on all the paths,", func() {
			Context("secured with Oauth2 introspection,", func() {
				Context("in a happy-path scenario", func() {
					It("should create a VirtualService and an AccessRule", func() {
						updateJwtHandlerTo(helpers.JWT_HANDLER_ISTIO)

						apiRuleName := generateTestName(testNameBase, testIDLength)
						serviceName := generateTestName(testServiceNameBase, testIDLength)
						serviceHost := fmt.Sprintf("%s.kyma.local", serviceName)

						rule := testRule(testPath, defaultMethods, defaultMutators, testOauthHandler(defaultScopes))
						apiRule := testApiRule(apiRuleName, testNamespace, serviceName, testNamespace, serviceHost, testServicePort, []gatewayv1beta1.Rule{rule})
						svc := testService(serviceName, testNamespace, testServicePort)

						// when
						Expect(c.Create(context.TODO(), svc)).Should(Succeed())
						Expect(c.Create(context.TODO(), apiRule)).Should(Succeed())
						defer func() {
							apiRuleTeardown(apiRule)
							serviceTeardown(svc)
						}()

						expectApiRuleStatus(apiRuleName, gatewayv1beta1.StatusOK)

						matchingLabels := matchingLabelsFunc(apiRuleName, testNamespace)

						By("Verifying created virtual service")
						vsList := networkingv1beta1.VirtualServiceList{}
						Eventually(func(g Gomega) {
							g.Expect(c.List(context.TODO(), &vsList, matchingLabels)).Should(Succeed())
							g.Expect(vsList.Items).To(HaveLen(1))

							vs := vsList.Items[0]

							//Meta
							g.Expect(vs.Name).To(HavePrefix(apiRuleName + "-"))
							g.Expect(len(vs.Name) > len(apiRuleName)).To(BeTrue())

							expectedSpec := builders.VirtualServiceSpec().
								Host(serviceHost).
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
							g.Expect(rl.Spec.Match.Methods).To(Equal(defaultMethods))
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

				Context("with Istio as JWT handler,", func() {
					Context("in a happy-path scenario", func() {
						It("should create a VirtualService, a RequestAuthentication and AuthorizationPolicies", func() {
							updateJwtHandlerTo(helpers.JWT_HANDLER_ISTIO)

							apiRuleName := generateTestName(testNameBase, testIDLength)
							serviceName := testServiceNameBase
							serviceHost := "httpbin-istio-jwt-happy-base.kyma.local"

							rule1 := testRule("/img", []string{"GET"}, nil, testIstioJWTHandlerWithScopes(testIssuer, testJwksUri, []string{"scope-a", "scope-b"}))
							rule2 := testRule("/headers", []string{"GET"}, nil, testIstioJWTHandlerWithScopes(testIssuer, testJwksUri, []string{"scope-c"}))
							apiRule := testApiRule(apiRuleName, testNamespace, serviceName, testNamespace, serviceHost, testServicePort, []gatewayv1beta1.Rule{rule1, rule2})
							svc := testService(serviceName, testNamespace, testServicePort)

							// when
							Expect(c.Create(context.TODO(), svc)).Should(Succeed())
							Expect(c.Create(context.TODO(), apiRule)).Should(Succeed())
							defer func() {
								apiRuleTeardown(apiRule)
								serviceTeardown(svc)
							}()

							expectApiRuleStatus(apiRuleName, gatewayv1beta1.StatusOK)

							ApiRuleNameMatchingLabels := matchingLabelsFunc(apiRuleName, testNamespace)

							By("Verifying virtual service")
							vsList := networkingv1beta1.VirtualServiceList{}
							Eventually(func(g Gomega) {
								g.Expect(c.List(context.TODO(), &vsList, ApiRuleNameMatchingLabels)).Should(Succeed())
								g.Expect(vsList.Items).To(HaveLen(1))
								vs := vsList.Items[0]

								expectedSpec := builders.VirtualServiceSpec().
									Host(serviceHost).
									Gateway(testGatewayURL).
									HTTP(builders.HTTPRoute().
										Match(builders.MatchRequest().Uri().Regex("/img")).
										Route(builders.RouteDestination().Host(fmt.Sprintf("%s.%s.svc.cluster.local", serviceName, testNamespace)).Port(testServicePort)).
										Headers(builders.NewHttpRouteHeadersBuilder().SetHostHeader(serviceHost).Get()).
										CorsPolicy(defaultCorsPolicy).
										Timeout(defaultHttpTimeout)).
									HTTP(builders.HTTPRoute().
										Match(builders.MatchRequest().Uri().Regex("/headers")).
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
								g.Expect(c.List(context.TODO(), &raList, ApiRuleNameMatchingLabels)).Should(Succeed())
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
								g.Expect(c.List(context.TODO(), &apList, ApiRuleNameMatchingLabels)).Should(Succeed())
								g.Expect(apList.Items).To(HaveLen(2))

								hasAuthorizationPolicyWithOperationPath := func(apList []*securityv1beta1.AuthorizationPolicy, operationPath string, assertWhen func(*securityv1beta1.AuthorizationPolicy)) {
									ap, err := getByOperationPath(apList, operationPath)
									g.Expect(err).NotTo(HaveOccurred())
									g.Expect(ap.Spec.Selector.MatchLabels).To(BeEquivalentTo(map[string]string{"app": serviceName}))
									g.Expect(ap.Spec.Rules).To(HaveLen(3))

									for i := 0; i < 3; i++ {
										g.Expect(ap.Spec.Rules[i].From[0].Source.RequestPrincipals[0]).To(Equal("https://oauth2.example.com//*"))
										g.Expect(ap.Spec.Rules[i].To[0].Operation.Paths[0]).To(Equal(operationPath))
										g.Expect(ap.Spec.Rules[i].To[0].Operation.Methods).To(BeEquivalentTo([]string{"GET"}))
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
							rule := testRule("/img", []string{"GET"}, nil, testIstioJWTHandlerWithAuthorizations(testIssuer, testJwksUri, authorizations))
							apiRule := testApiRule(apiRuleName, testNamespace, serviceName, testNamespace, serviceHost, testServicePort, []gatewayv1beta1.Rule{rule})
							svc := testService(serviceName, testNamespace, testServicePort)

							By(fmt.Sprintf("Creating APIRule %s", apiRuleName))

							// when
							Expect(c.Create(context.TODO(), svc)).Should(Succeed())
							Expect(c.Create(context.TODO(), apiRule)).Should(Succeed())
							defer func() {
								apiRuleTeardown(apiRule)
								serviceTeardown(svc)
							}()

							Eventually(func(g Gomega) {
								createdApiRule := gatewayv1beta1.APIRule{}
								g.Expect(c.Get(context.TODO(), client.ObjectKey{Name: apiRuleName, Namespace: testNamespace}, &createdApiRule)).Should(Succeed())
								g.Expect(createdApiRule.Status.APIRuleStatus).NotTo(BeNil())
								g.Expect(createdApiRule.Status.APIRuleStatus.Code).To(Equal(gatewayv1beta1.StatusOK))
								g.Expect(createdApiRule.Status.APIRuleStatus.Code).To(Equal(gatewayv1beta1.StatusOK))
								g.Expect(createdApiRule.Status.VirtualServiceStatus.Code).To(Equal(gatewayv1beta1.StatusOK))
								g.Expect(createdApiRule.Status.AuthorizationPolicyStatus.Code).To(Equal(gatewayv1beta1.StatusOK))
								g.Expect(createdApiRule.Status.RequestAuthenticationStatus.Code).To(Equal(gatewayv1beta1.StatusOK))
							}, eventuallyTimeout).Should(Succeed())

							// when
							updatedApiRule := gatewayv1beta1.APIRule{}
							Expect(c.Get(context.TODO(), client.ObjectKey{Name: apiRuleName, Namespace: testNamespace}, &updatedApiRule)).Should(Succeed())

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
							ruleWithScopes := testRule("/img", []string{"GET"}, nil, testIstioJWTHandlerWithAuthorizations(testIssuer, testJwksUri, updatedAuthorizations))
							updatedApiRule.Spec.Rules = []gatewayv1beta1.Rule{ruleWithScopes}

							By(fmt.Sprintf("Updating APIRule %s with new Authorizations for /img path", apiRuleName))
							Expect(c.Update(context.TODO(), &updatedApiRule)).Should(Succeed())

							// then
							matchingLabels := matchingLabelsFunc(apiRuleName, testNamespace)

							Eventually(func(g Gomega) {
								apList := securityv1beta1.AuthorizationPolicyList{}
								g.Expect(c.List(context.TODO(), &apList, matchingLabels)).Should(Succeed())
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

					rule1 := testRule("/img", []string{"GET"}, nil, testIstioJWTHandler(testIssuer, testJwksUri))
					apiRule := testApiRule(apiRuleName, testNamespace, serviceName, testNamespace, serviceHost, testServicePort, []gatewayv1beta1.Rule{rule1})
					svc := testService(serviceName, testNamespace, testServicePort)
					delete(svc.Spec.Selector, "app")
					svc.Spec.Selector["custom"] = serviceName

					// when
					Expect(c.Create(context.TODO(), svc)).Should(Succeed())
					Expect(c.Create(context.TODO(), apiRule)).Should(Succeed())
					defer func() {
						apiRuleTeardown(apiRule)
						serviceTeardown(svc)
					}()

					expectApiRuleStatus(apiRuleName, gatewayv1beta1.StatusOK)

					ApiRuleNameMatchingLabels := matchingLabelsFunc(apiRuleName, testNamespace)

					By("Verifying request authentication")

					raList := securityv1beta1.RequestAuthenticationList{}
					Eventually(func(g Gomega) {
						g.Expect(c.List(context.TODO(), &raList, ApiRuleNameMatchingLabels)).Should(Succeed())
						g.Expect(raList.Items).To(HaveLen(1))
						g.Expect(raList.Items[0].Spec.Selector.MatchLabels).To(HaveLen(1))
						g.Expect(raList.Items[0].Spec.Selector.MatchLabels).To(HaveKeyWithValue("custom", serviceName))
					}, eventuallyTimeout).Should(Succeed())

					By("Verifying authorization policy")

					apList := securityv1beta1.AuthorizationPolicyList{}
					Eventually(func(g Gomega) {
						g.Expect(c.List(context.TODO(), &apList, ApiRuleNameMatchingLabels)).Should(Succeed())
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

					rule1 := testRule("/img", []string{"GET"}, nil, testIstioJWTHandler(testIssuer, testJwksUri))
					apiRule := testApiRule(apiRuleName, testNamespace, serviceName, testNamespace, serviceHost, testServicePort, []gatewayv1beta1.Rule{rule1})
					svc := testService(serviceName, testNamespace, testServicePort)
					delete(svc.Spec.Selector, "app")
					svc.Spec.Selector["custom"] = serviceName
					svc.Spec.Selector["second-custom"] = "blah"

					// when
					Expect(c.Create(context.TODO(), svc)).Should(Succeed())
					Expect(c.Create(context.TODO(), apiRule)).Should(Succeed())
					defer func() {
						apiRuleTeardown(apiRule)
						serviceTeardown(svc)
					}()

					expectApiRuleStatus(apiRuleName, gatewayv1beta1.StatusOK)

					ApiRuleNameMatchingLabels := matchingLabelsFunc(apiRuleName, testNamespace)

					By("Verifying request authentication")

					raList := securityv1beta1.RequestAuthenticationList{}
					Eventually(func(g Gomega) {
						g.Expect(c.List(context.TODO(), &raList, ApiRuleNameMatchingLabels)).Should(Succeed())
						g.Expect(raList.Items).To(HaveLen(1))
						g.Expect(raList.Items[0].Spec.Selector.MatchLabels).To(HaveLen(2))
						g.Expect(raList.Items[0].Spec.Selector.MatchLabels).To(HaveKeyWithValue("custom", serviceName))
						g.Expect(raList.Items[0].Spec.Selector.MatchLabels).To(HaveKeyWithValue("second-custom", "blah"))
					}, eventuallyTimeout).Should(Succeed())

					By("Verifying authorization policy")

					apList := securityv1beta1.AuthorizationPolicyList{}
					Eventually(func(g Gomega) {
						g.Expect(c.List(context.TODO(), &apList, ApiRuleNameMatchingLabels)).Should(Succeed())
						g.Expect(apList.Items).To(HaveLen(1))
						g.Expect(apList.Items[0].Spec.Selector.MatchLabels).To(HaveLen(2))
						g.Expect(apList.Items[0].Spec.Selector.MatchLabels).To(HaveKeyWithValue("custom", serviceName))
						g.Expect(apList.Items[0].Spec.Selector.MatchLabels).To(HaveKeyWithValue("second-custom", "blah"))
					}, eventuallyTimeout).Should(Succeed())
				})
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
		Expect(c.Create(context.TODO(), vs)).Should(Succeed())
		defer func() {
			By(fmt.Sprintf("Deleting VirtualService %s as part of teardown", vs.Name))
			Eventually(func(g Gomega) {
				_ = c.Delete(context.TODO(), vs)
				v := networkingv1beta1.VirtualService{}
				err := c.Get(context.TODO(), client.ObjectKey{Name: vs.Name, Namespace: testNamespace}, &v)
				g.Expect(errors.IsNotFound(err)).To(BeTrue())
			}, eventuallyTimeout).Should(Succeed())
		}()

		By("Verifying virtual service has been created")
		Eventually(func(g Gomega) {
			createdVs := networkingv1beta1.VirtualService{}
			g.Expect(c.Get(context.TODO(), client.ObjectKey{Name: vsName, Namespace: testNamespace}, &createdVs)).Should(Succeed())
		}, eventuallyTimeout).Should(Succeed())

		apiRuleLabelMatcher := matchingLabelsFunc(apiRuleName, testNamespace)

		By("Creating APIRule")
		rule := testRule("/headers", []string{"GET"}, nil, testIstioJWTHandler(testIssuer, testJwksUri))
		apiRule := testApiRule(apiRuleName, testNamespace, serviceName, testNamespace, serviceHost, testServicePort, []gatewayv1beta1.Rule{rule})
		svc := testService(serviceName, testNamespace, testServicePort)

		// when
		Expect(c.Create(context.TODO(), svc)).Should(Succeed())
		Expect(c.Create(context.TODO(), apiRule)).Should(Succeed())
		defer func() {
			apiRuleTeardown(apiRule)
			serviceTeardown(svc)
		}()

		expectApiRuleStatus(apiRuleName, gatewayv1beta1.StatusError)

		By("Verifying virtual service for APIRule has not been created")
		verifyVirtualServiceCount(c, apiRuleLabelMatcher, 0)

		By("Deleting existing virtual service with duplicated host configuration")
		deleteVirtualService(vs)

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
		rule := testRule("/headers", []string{"GET"}, nil, testIstioJWTHandler(testIssuer, testJwksUri))
		apiRule := testApiRule(apiRuleName, testNamespace, serviceName, testNamespace, serviceHost, testServicePort, []gatewayv1beta1.Rule{rule})
		svc := testService(serviceName, testNamespace, testServicePort)

		// when
		Expect(c.Create(context.TODO(), svc)).Should(Succeed())
		Expect(c.Create(context.TODO(), apiRule)).Should(Succeed())
		defer func() {
			apiRuleTeardown(apiRule)
			serviceTeardown(svc)
		}()

		expectApiRuleStatus(apiRuleName, gatewayv1beta1.StatusOK)

		By(fmt.Sprintf("Creating virtual service for host %s", serviceHost))
		vs := virtualService(vsName, serviceHost)
		Expect(c.Create(context.TODO(), vs)).Should(Succeed())
		defer func() {
			By(fmt.Sprintf("Deleting VirtualService %s as part of teardown", vs.Name))
			Eventually(func(g Gomega) {
				_ = c.Delete(context.TODO(), vs)
				v := networkingv1beta1.VirtualService{}
				err := c.Get(context.TODO(), client.ObjectKey{Name: vs.Name, Namespace: testNamespace}, &v)
				g.Expect(errors.IsNotFound(err)).To(BeTrue())
			}, eventuallyTimeout).Should(Succeed())
		}()

		By("Verifying virtual service has been created")
		Eventually(func(g Gomega) {
			createdVs := networkingv1beta1.VirtualService{}
			g.Expect(c.Get(context.TODO(), client.ObjectKey{Name: vsName, Namespace: testNamespace}, &createdVs)).Should(Succeed())
		}, eventuallyTimeout).Should(Succeed())

		By("Waiting until APIRule is reconciled after error")
		expectApiRuleStatus(apiRuleName, gatewayv1beta1.StatusError)

		By("Verifying APIRule status description")
		Eventually(func(g Gomega) {
			expectedApiRule := gatewayv1beta1.APIRule{}
			g.Expect(c.Get(context.TODO(), client.ObjectKey{Name: apiRuleName, Namespace: testNamespace}, &expectedApiRule)).Should(Succeed())
			g.Expect(expectedApiRule.Status.APIRuleStatus).NotTo(BeNil())
			g.Expect(expectedApiRule.Status.APIRuleStatus.Description).To(ContainSubstring("This host is occupied by another Virtual Service"))
		}, eventuallyTimeout).Should(Succeed())
	})

	It("Should recreate resources created by reconciler when they are manually deleted", func() {
		updateJwtHandlerTo(helpers.JWT_HANDLER_ISTIO)

		apiRuleName := generateTestName(testNameBase, testIDLength)
		serviceName := testServiceNameBase
		serviceHost := "httpbin-recreate-resources.kyma.local"

		rule := testRule("/img", []string{"GET"}, nil, testIstioJWTHandlerWithScopes(testIssuer, testJwksUri, []string{"scope-a"}))
		apiRule := testApiRule(apiRuleName, testNamespace, serviceName, testNamespace, serviceHost, testServicePort, []gatewayv1beta1.Rule{rule})
		svc := testService(serviceName, testNamespace, testServicePort)

		// when
		Expect(c.Create(context.TODO(), svc)).Should(Succeed())
		Expect(c.Create(context.TODO(), apiRule)).Should(Succeed())
		defer func() {
			apiRuleTeardown(apiRule)
			serviceTeardown(svc)
		}()

		expectApiRuleStatus(apiRuleName, gatewayv1beta1.StatusOK)

		apiRuleNameMatchingLabels := matchingLabelsFunc(apiRuleName, testNamespace)

		By("Verifying that resources are created")
		verifyVirtualServiceCount(c, apiRuleNameMatchingLabels, 1)
		verifyRequestAuthenticationCount(c, apiRuleNameMatchingLabels, 1)
		verifyAuthorizationPolicyCount(c, apiRuleNameMatchingLabels, 1)

		By("Deleting Virtual Service")
		Eventually(func(g Gomega) {
			vsList := networkingv1beta1.VirtualServiceList{}
			g.Expect(c.List(context.TODO(), &vsList, apiRuleNameMatchingLabels)).Should(Succeed())
			g.Expect(c.Delete(context.TODO(), vsList.Items[0])).Should(Succeed())
		}, eventuallyTimeout).Should(Succeed())

		By("Deleting Request Authentication")
		raList := securityv1beta1.RequestAuthenticationList{}
		Eventually(func(g Gomega) {
			g.Expect(c.List(context.TODO(), &raList, apiRuleNameMatchingLabels)).Should(Succeed())
			g.Expect(c.Delete(context.TODO(), raList.Items[0])).Should(Succeed())
		}, eventuallyTimeout).Should(Succeed())

		By("Deleting Authorization Policy")
		apList := securityv1beta1.AuthorizationPolicyList{}
		Eventually(func(g Gomega) {
			g.Expect(c.List(context.TODO(), &apList, apiRuleNameMatchingLabels)).Should(Succeed())
			g.Expect(c.Delete(context.TODO(), apList.Items[0])).Should(Succeed())
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

		rule := testRule("/img", []string{"GET"}, nil, testIstioJWTHandlerWithScopes(testIssuer, testJwksUri, []string{"scope-a"}))
		apiRule := testApiRule(apiRuleName, testNamespace, serviceName, testNamespace, serviceHost, testServicePort, []gatewayv1beta1.Rule{rule})
		svc := testService(serviceName, testNamespace, testServicePort)

		// when
		Expect(c.Create(context.TODO(), svc)).Should(Succeed())
		Expect(c.Create(context.TODO(), apiRule)).Should(Succeed())
		defer func() {
			apiRuleTeardown(apiRule)
			serviceTeardown(svc)
		}()

		expectApiRuleStatus(apiRuleName, gatewayv1beta1.StatusOK)
		apiRuleNameMatchingLabels := matchingLabelsFunc(apiRuleName, testNamespace)

		verifyVirtualServiceCount(c, apiRuleNameMatchingLabels, 1)
		verifyRequestAuthenticationCount(c, apiRuleNameMatchingLabels, 1)
		verifyAuthorizationPolicyCount(c, apiRuleNameMatchingLabels, 1)

		deleteApiRule(apiRule)

		By("Verifying resources are deleted")
		verifyVirtualServiceCount(c, apiRuleNameMatchingLabels, 0)
		verifyRequestAuthenticationCount(c, apiRuleNameMatchingLabels, 0)
		verifyAuthorizationPolicyCount(c, apiRuleNameMatchingLabels, 0)
	})
})

func verifyVirtualServiceCount(c client.Client, option client.ListOption, count int) {
	By(fmt.Sprintf("Verifying %d Virtual Service exist", count))
	Eventually(func(g Gomega) {
		vsList := networkingv1beta1.VirtualServiceList{}
		g.Expect(c.List(context.TODO(), &vsList, option)).Should(Succeed())
		g.Expect(vsList.Items).To(HaveLen(count))
	}, eventuallyTimeout).Should(Succeed())
}

func verifyRequestAuthenticationCount(c client.Client, option client.ListOption, count int) {
	By(fmt.Sprintf("Verifying %d Request Authentication exist", count))
	Eventually(func(g Gomega) {
		raList := securityv1beta1.RequestAuthenticationList{}
		g.Expect(c.List(context.TODO(), &raList, option)).Should(Succeed())
		g.Expect(raList.Items).To(HaveLen(count))
	}, eventuallyTimeout).Should(Succeed())
}

func verifyAuthorizationPolicyCount(c client.Client, option client.ListOption, count int) {
	By(fmt.Sprintf("Verifying %d Authorization Policy exist", count))
	Eventually(func(g Gomega) {
		apList := securityv1beta1.AuthorizationPolicyList{}
		g.Expect(c.List(context.TODO(), &apList, option)).Should(Succeed())
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

func testRule(path string, methods []string, mutators []*gatewayv1beta1.Mutator, handler *gatewayv1beta1.Handler) gatewayv1beta1.Rule {
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
	g.Expect(c.List(context.TODO(), &res, matchingLabels)).Should(Succeed())
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
		g.Expect(actual[ruleUrl].Spec.Match.Methods).To(Equal(expected[i].Methods))
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
	Expect(c.Get(context.TODO(), client.ObjectKey{Name: helpers.CM_NAME, Namespace: helpers.CM_NS}, cm)).Should(Succeed())

	if !strings.Contains(cm.Data[helpers.CM_KEY], jwtHandler) {
		By(fmt.Sprintf("Updating JWT handler config map to %s", jwtHandler))
		cm.Data = map[string]string{
			helpers.CM_KEY: fmt.Sprintf("jwtHandler: %s", jwtHandler),
		}
		Expect(c.Update(context.TODO(), cm)).To(Succeed())

		By("Waiting until config map is updated")
		Eventually(func(g Gomega) {
			g.Expect(c.Get(context.TODO(), client.ObjectKey{Name: cm.Name, Namespace: cm.Namespace}, cm)).Should(Succeed())
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

func deleteApiRule(apiRule *gatewayv1beta1.APIRule) {
	By(fmt.Sprintf("Deleting ApiRule %s", apiRule.Name))
	Expect(c.Delete(context.TODO(), apiRule)).Should(Succeed())
	Eventually(func(g Gomega) {
		a := gatewayv1beta1.APIRule{}
		err := c.Get(context.TODO(), client.ObjectKey{Name: apiRule.Name, Namespace: testNamespace}, &a)
		g.Expect(errors.IsNotFound(err)).To(BeTrue())
	}, eventuallyTimeout).Should(Succeed())
}

func apiRuleTeardown(apiRule *gatewayv1beta1.APIRule) {
	By(fmt.Sprintf("Deleting ApiRule %s as part of teardown", apiRule.Name))
	err := c.Delete(context.TODO(), apiRule)

	if err != nil {
		Expect(errors.IsNotFound(err)).To(BeTrue())
	}

	Eventually(func(g Gomega) {
		a := gatewayv1beta1.APIRule{}
		err := c.Get(context.TODO(), client.ObjectKey{Name: apiRule.Name, Namespace: testNamespace}, &a)
		g.Expect(errors.IsNotFound(err)).To(BeTrue())
	}, eventuallyTimeout).Should(Succeed())
}

func serviceTeardown(svc *corev1.Service) {
	By(fmt.Sprintf("Deleting Service %s as part of teardown", svc.Name))
	err := c.Delete(context.TODO(), svc)

	if err != nil {
		Expect(errors.IsNotFound(err)).To(BeTrue())
	}

	Eventually(func(g Gomega) {
		a := corev1.Service{}
		err := c.Get(context.TODO(), client.ObjectKey{Name: svc.Name, Namespace: testNamespace}, &a)
		g.Expect(errors.IsNotFound(err)).To(BeTrue())
	}, eventuallyTimeout).Should(Succeed())
}

func expectApiRuleStatus(apiRuleName string, statusCode gatewayv1beta1.StatusCode) {
	By(fmt.Sprintf("Verifying that ApiRule %s has status %s", apiRuleName, statusCode))
	Eventually(func(g Gomega) {
		expectedApiRule := gatewayv1beta1.APIRule{}
		g.Expect(c.Get(context.TODO(), client.ObjectKey{Name: apiRuleName, Namespace: testNamespace}, &expectedApiRule)).Should(Succeed())
		g.Expect(expectedApiRule.Status.APIRuleStatus).NotTo(BeNil())
		g.Expect(expectedApiRule.Status.APIRuleStatus.Code).To(Equal(statusCode))
	}, eventuallyTimeout).Should(Succeed())
}

func deleteVirtualService(vs *networkingv1beta1.VirtualService) {
	By(fmt.Sprintf("Deleting virtual service %s", vs.Name))
	Expect(c.Delete(context.TODO(), vs)).Should(Succeed())
	Eventually(func(g Gomega) {
		v := networkingv1beta1.VirtualService{}
		err := c.Get(context.TODO(), client.ObjectKey{Name: vs.Name, Namespace: testNamespace}, &v)
		g.Expect(errors.IsNotFound(err)).To(BeTrue())
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
