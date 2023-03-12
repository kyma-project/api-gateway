package controllers_test

import (
	"context"
	"fmt"
	gomegatypes "github.com/onsi/gomega/types"
	"math/rand"
	"time"

	"encoding/json"
	"github.com/kyma-project/api-gateway/internal/builders"
	"github.com/kyma-project/api-gateway/internal/helpers"
	"github.com/kyma-project/api-gateway/internal/processing"

	gatewayv1beta1 "github.com/kyma-project/api-gateway/api/v1beta1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	rulev1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var _ = Describe("APIRule Controller", func() {

	const (
		testNameBase               = "test"
		testIDLength               = 5
		testServiceNameBase        = "httpbin"
		testServicePort     uint32 = 443
		testPath                   = "/.*"
		testIssuer                 = "https://oauth2.example.com/"
		testJwksUri                = "https://oauth2.example.com/.well-known/jwks.json"
	)

	BeforeEach(func() {
		// We configure `ory` in ConfigMap as the default for all tests
		setHandlerConfigMap(helpers.JWT_HANDLER_ORY)
	})

	Context("when updating the APIRule with multiple paths", func() {

		It("should create, update and delete rules depending on patch match", func() {
			rule1 := testRule("/rule1", []string{"GET"}, defaultMutators, noConfigHandler("noop"))
			rule2 := testRule("/rule2", []string{"PUT"}, defaultMutators, noConfigHandler("unauthorized"))
			rule3 := testRule("/rule3", []string{"DELETE"}, defaultMutators, noConfigHandler("anonymous"))

			apiRuleName := generateTestName(testNameBase, testIDLength)
			serviceName := generateTestName(testServiceNameBase, testIDLength)
			serviceHost := fmt.Sprintf("%s.kyma.local", serviceName)

			matchingLabels := matchingLabelsFunc(apiRuleName, testNamespace)

			pathToURLFunc := func(path string) string {
				return fmt.Sprintf("<http|https>://%s<%s>", serviceHost, path)
			}

			By("Create APIRule")

			instance := testInstance(apiRuleName, testNamespace, serviceName, serviceHost, testServicePort, []gatewayv1beta1.Rule{rule1, rule2, rule3})
			err := c.Create(context.TODO(), instance)

			if apierrors.IsInvalid(err) {
				Fail(fmt.Sprintf("failed to create object, got an invalid object error: %v", err))
			}
			Expect(err).NotTo(HaveOccurred())
			defer func() {
				err := c.Delete(context.TODO(), instance)
				Expect(err).NotTo(HaveOccurred())
			}()

			expectedRequest := reconcile.Request{NamespacedName: types.NamespacedName{Name: apiRuleName, Namespace: testNamespace}}
			Eventually(requests, eventuallyTimeout).Should(Receive(Equal(expectedRequest)))

			By("Verify before update")

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

			By("Update APIRule")
			existingInstance := gatewayv1beta1.APIRule{}
			err = c.Get(context.TODO(), client.ObjectKey{Name: apiRuleName, Namespace: testNamespace}, &existingInstance)
			Expect(err).NotTo(HaveOccurred())
			rule4 := testRule("/rule4", []string{"POST"}, defaultMutators, noConfigHandler("cookie_session"))
			existingInstance.Spec.Rules = []gatewayv1beta1.Rule{rule1, rule4}
			newServiceName := serviceName + "new"
			newServicePort := testServicePort + 3
			existingInstance.Spec.Service.Name = &newServiceName
			existingInstance.Spec.Service.Port = &newServicePort

			err = c.Update(context.TODO(), &existingInstance)
			Expect(err).NotTo(HaveOccurred())
			expectedRequest = reconcile.Request{NamespacedName: types.NamespacedName{Name: apiRuleName, Namespace: testNamespace}}
			Eventually(requests, eventuallyTimeout).Should(Receive(Equal(expectedRequest)))

			By("Verify after update")

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
						setHandlerConfigMap(helpers.JWT_HANDLER_ORY)

						apiRuleName := generateTestName(testNameBase, testIDLength)
						serviceName := generateTestName(testServiceNameBase, testIDLength)
						serviceHost := fmt.Sprintf("%s.kyma.local", serviceName)

						rule := testRule(testPath, defaultMethods, defaultMutators, testOauthHandler(defaultScopes))
						instance := testInstance(apiRuleName, testNamespace, serviceName, serviceHost, testServicePort, []gatewayv1beta1.Rule{rule})

						err := c.Create(context.TODO(), instance)
						if apierrors.IsInvalid(err) {
							Fail(fmt.Sprintf("failed to create object, got an invalid object error: %v", err))
							return
						}
						Expect(err).NotTo(HaveOccurred())
						defer func() {
							err := c.Delete(context.TODO(), instance)
							Expect(err).NotTo(HaveOccurred())
						}()

						expectedRequest := reconcile.Request{NamespacedName: types.NamespacedName{Name: apiRuleName, Namespace: testNamespace}}

						Eventually(requests, eventuallyTimeout).Should(Receive(Equal(expectedRequest)))

						matchingLabels := matchingLabelsFunc(apiRuleName, testNamespace)

						//Verify VirtualService
						vsList := networkingv1beta1.VirtualServiceList{}
						Eventually(func(g Gomega) {
							err = c.List(context.TODO(), &vsList, matchingLabels)
							g.Expect(err).NotTo(HaveOccurred())
							g.Expect(vsList.Items).To(HaveLen(1))
						}, eventuallyTimeout).Should(Succeed())
						vs := vsList.Items[0]

						//Meta
						Expect(vs.Name).To(HavePrefix(apiRuleName + "-"))
						Expect(len(vs.Name) > len(apiRuleName)).To(BeTrue())

						expectedSpec := builders.VirtualServiceSpec().
							Host(serviceHost).
							Gateway(testGatewayURL).
							HTTP(builders.HTTPRoute().
								Match(builders.MatchRequest().Uri().Regex(testPath)).
								Route(builders.RouteDestination().Host(testOathkeeperSvcURL).Port(testOathkeeperPort)).
								Headers(builders.Headers().SetHostHeader(serviceHost)).
								CorsPolicy(defaultCorsPolicy).
								Timeout(time.Second * helpers.DEFAULT_HTTP_TIMEOUT))

						gotSpec := *expectedSpec.Get()
						Expect(*vs.Spec.DeepCopy()).To(Equal(*gotSpec.DeepCopy()))

						//Verify Rule
						expectedRuleMatchURL := fmt.Sprintf("<http|https>://%s<%s>", serviceHost, testPath)

						var rlList []rulev1alpha1.Rule
						Eventually(func(g Gomega) {
							rlList = getRuleList(g, matchingLabels)
							g.Expect(rlList).To(HaveLen(1))
						}, eventuallyTimeout).Should(Succeed())

						rl := rlList[0]
						Expect(rl.Spec.Match.URL).To(Equal(expectedRuleMatchURL))

						//Meta
						Expect(rl.Name).To(HavePrefix(apiRuleName + "-"))
						Expect(len(rl.Name) > len(apiRuleName)).To(BeTrue())

						//Spec.Upstream
						Expect(rl.Spec.Upstream).NotTo(BeNil())
						Expect(rl.Spec.Upstream.URL).To(Equal(fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", serviceName, testNamespace, testServicePort)))
						Expect(rl.Spec.Upstream.StripPath).To(BeNil())
						Expect(rl.Spec.Upstream.PreserveHost).To(BeNil())
						//Spec.Match
						Expect(rl.Spec.Match).NotTo(BeNil())
						Expect(rl.Spec.Match.URL).To(Equal(fmt.Sprintf("<http|https>://%s<%s>", serviceHost, testPath)))
						Expect(rl.Spec.Match.Methods).To(Equal(defaultMethods))
						//Spec.Authenticators
						Expect(rl.Spec.Authenticators).To(HaveLen(1))
						Expect(rl.Spec.Authenticators[0].Handler).NotTo(BeNil())
						Expect(rl.Spec.Authenticators[0].Handler.Name).To(Equal("oauth2_introspection"))
						Expect(rl.Spec.Authenticators[0].Handler.Config).NotTo(BeNil())
						//Authenticators[0].Handler.Config validation
						handlerConfig := map[string]interface{}{}
						err = json.Unmarshal(rl.Spec.Authenticators[0].Config.Raw, &handlerConfig)
						Expect(err).NotTo(HaveOccurred())
						Expect(handlerConfig).To(HaveLen(1))
						Expect(asStringSlice(handlerConfig["required_scope"])).To(BeEquivalentTo(defaultScopes))
						//Spec.Authorizer
						Expect(rl.Spec.Authorizer).NotTo(BeNil())
						Expect(rl.Spec.Authorizer.Handler).NotTo(BeNil())
						Expect(rl.Spec.Authorizer.Handler.Name).To(Equal("allow"))
						Expect(rl.Spec.Authorizer.Handler.Config).To(BeNil())

						//Spec.Mutators
						Expect(rl.Spec.Mutators).NotTo(BeNil())
						Expect(len(rl.Spec.Mutators)).To(Equal(len(defaultMutators)))
						Expect(rl.Spec.Mutators[0].Handler.Name).To(Equal(defaultMutators[0].Name))
						Expect(rl.Spec.Mutators[1].Handler.Name).To(Equal(defaultMutators[1].Name))
					})
				})
			})

			Context("secured with JWT token authentication,", func() {
				Context("with ORY as JWT handler,", func() {
					Context("in a happy-path scenario", func() {
						It("should create a VirtualService and an AccessRules", func() {
							apiRuleName := generateTestName(testNameBase, testIDLength)
							serviceName := generateTestName(testServiceNameBase, testIDLength)
							serviceHost := fmt.Sprintf("%s.kyma.local", serviceName)

							rule1 := testRule("/img", []string{"GET"}, defaultMutators, testOryJWTHandler(testIssuer, defaultScopes))
							rule2 := testRule("/headers", []string{"GET"}, defaultMutators, testOryJWTHandler(testIssuer, defaultScopes))
							instance := testInstance(apiRuleName, testNamespace, serviceName, serviceHost, testServicePort, []gatewayv1beta1.Rule{rule1, rule2})

							err := c.Create(context.TODO(), instance)
							if apierrors.IsInvalid(err) {
								Fail(fmt.Sprintf("failed to create object, got an invalid object error: %v", err))
								return
							}
							Expect(err).NotTo(HaveOccurred())
							defer func() {
								err := c.Delete(context.TODO(), instance)
								Expect(err).NotTo(HaveOccurred())
							}()

							expectedRequest := reconcile.Request{NamespacedName: types.NamespacedName{Name: apiRuleName, Namespace: testNamespace}}

							Eventually(requests, eventuallyTimeout).Should(Receive(Equal(expectedRequest)))

							matchingLabels := matchingLabelsFunc(apiRuleName, testNamespace)

							//Verify VirtualService
							vsList := networkingv1beta1.VirtualServiceList{}
							Eventually(func(g Gomega) {
								err = c.List(context.TODO(), &vsList, matchingLabels)
								g.Expect(err).NotTo(HaveOccurred())
								g.Expect(vsList.Items).To(HaveLen(1))
							}, eventuallyTimeout).Should(Succeed())

							vs := vsList.Items[0]

							expectedSpec := builders.VirtualServiceSpec().
								Host(serviceHost).
								Gateway(testGatewayURL).
								HTTP(builders.HTTPRoute().
									Match(builders.MatchRequest().Uri().Regex("/img")).
									Route(builders.RouteDestination().Host(testOathkeeperSvcURL).Port(testOathkeeperPort)).
									Headers(builders.Headers().SetHostHeader(serviceHost)).
									CorsPolicy(defaultCorsPolicy).
									Timeout(time.Second * helpers.DEFAULT_HTTP_TIMEOUT)).
								HTTP(builders.HTTPRoute().
									Match(builders.MatchRequest().Uri().Regex("/headers")).
									Route(builders.RouteDestination().Host(testOathkeeperSvcURL).Port(testOathkeeperPort)).
									Headers(builders.Headers().SetHostHeader(serviceHost)).
									CorsPolicy(defaultCorsPolicy).
									Timeout(time.Second * helpers.DEFAULT_HTTP_TIMEOUT))
							gotSpec := *expectedSpec.Get()
							Expect(*vs.Spec.DeepCopy()).To(Equal(*gotSpec.DeepCopy()))

							//Verify Rule1
							expectedRuleMatchURL := fmt.Sprintf("<http|https>://%s<%s>", serviceHost, "/img")

							var rlList []rulev1alpha1.Rule
							Eventually(func(g Gomega) {
								rlList = getRuleList(g, matchingLabels)
								g.Expect(rlList).To(HaveLen(2))
							}, eventuallyTimeout).Should(Succeed())

							rules := make(map[string]rulev1alpha1.Rule)

							for _, rule := range rlList {
								rules[rule.Spec.Match.URL] = rule
							}

							rl := rules[expectedRuleMatchURL]

							//Spec.Upstream
							Expect(rl.Spec.Upstream).NotTo(BeNil())
							Expect(rl.Spec.Upstream.URL).To(Equal(fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", serviceName, testNamespace, testServicePort)))
							Expect(rl.Spec.Upstream.StripPath).To(BeNil())
							Expect(rl.Spec.Upstream.PreserveHost).To(BeNil())
							//Spec.Match
							Expect(rl.Spec.Match).NotTo(BeNil())
							Expect(rl.Spec.Match.URL).To(Equal(expectedRuleMatchURL))
							Expect(rl.Spec.Match.Methods).To(Equal([]string{"GET"}))
							//Spec.Authenticators
							Expect(rl.Spec.Authenticators).To(HaveLen(1))
							Expect(rl.Spec.Authenticators[0].Handler).NotTo(BeNil())
							Expect(rl.Spec.Authenticators[0].Handler.Name).To(Equal("jwt"))
							Expect(rl.Spec.Authenticators[0].Handler.Config).NotTo(BeNil())
							//Authenticators[0].Handler.Config validation
							handlerConfig := map[string]interface{}{}

							err = json.Unmarshal(rl.Spec.Authenticators[0].Config.Raw, &handlerConfig)
							Expect(err).NotTo(HaveOccurred())
							Expect(handlerConfig).To(HaveLen(3))
							Expect(asStringSlice(handlerConfig["required_scope"])).To(BeEquivalentTo(defaultScopes))
							Expect(asStringSlice(handlerConfig["trusted_issuers"])).To(BeEquivalentTo([]string{testIssuer}))
							//Spec.Authorizer
							Expect(rl.Spec.Authorizer).NotTo(BeNil())
							Expect(rl.Spec.Authorizer.Handler).NotTo(BeNil())
							Expect(rl.Spec.Authorizer.Handler.Name).To(Equal("allow"))
							Expect(rl.Spec.Authorizer.Handler.Config).To(BeNil())

							//Spec.Mutators
							Expect(rl.Spec.Mutators).NotTo(BeNil())
							Expect(len(rl.Spec.Mutators)).To(Equal(len(defaultMutators)))
							Expect(rl.Spec.Mutators[0].Handler.Name).To(Equal(defaultMutators[0].Name))
							Expect(rl.Spec.Mutators[1].Handler.Name).To(Equal(defaultMutators[1].Name))

							//Verify Rule2
							expectedRule2MatchURL := fmt.Sprintf("<http|https>://%s<%s>", serviceHost, "/headers")
							rl2 := rules[expectedRule2MatchURL]

							//Spec.Upstream
							Expect(rl2.Spec.Upstream).NotTo(BeNil())
							Expect(rl2.Spec.Upstream.URL).To(Equal(fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", serviceName, testNamespace, testServicePort)))
							Expect(rl2.Spec.Upstream.StripPath).To(BeNil())
							Expect(rl2.Spec.Upstream.PreserveHost).To(BeNil())
							//Spec.Match
							Expect(rl2.Spec.Match).NotTo(BeNil())
							Expect(rl2.Spec.Match.URL).To(Equal(expectedRule2MatchURL))
							Expect(rl2.Spec.Match.Methods).To(Equal([]string{"GET"}))
							//Spec.Authenticators
							Expect(rl2.Spec.Authenticators).To(HaveLen(1))
							Expect(rl2.Spec.Authenticators[0].Handler).NotTo(BeNil())
							Expect(rl2.Spec.Authenticators[0].Handler.Name).To(Equal("jwt"))
							Expect(rl2.Spec.Authenticators[0].Handler.Config).NotTo(BeNil())
							//Authenticators[0].Handler.Config validation
							handlerConfig = map[string]interface{}{}

							err = json.Unmarshal(rl2.Spec.Authenticators[0].Config.Raw, &handlerConfig)
							Expect(err).NotTo(HaveOccurred())
							Expect(handlerConfig).To(HaveLen(3))
							Expect(asStringSlice(handlerConfig["required_scope"])).To(BeEquivalentTo(defaultScopes))
							Expect(asStringSlice(handlerConfig["trusted_issuers"])).To(BeEquivalentTo([]string{testIssuer}))
							//Spec.Authorizer
							Expect(rl2.Spec.Authorizer).NotTo(BeNil())
							Expect(rl2.Spec.Authorizer.Handler).NotTo(BeNil())
							Expect(rl2.Spec.Authorizer.Handler.Name).To(Equal("allow"))
							Expect(rl2.Spec.Authorizer.Handler.Config).To(BeNil())

							//Spec.Mutators
							Expect(rl2.Spec.Mutators).NotTo(BeNil())
							Expect(len(rl2.Spec.Mutators)).To(Equal(len(defaultMutators)))
							Expect(rl2.Spec.Mutators[0].Handler.Name).To(Equal(defaultMutators[0].Name))
							Expect(rl2.Spec.Mutators[1].Handler.Name).To(Equal(defaultMutators[1].Name))
						})
					})
				})

				Context("with Istio as JWT handler,", func() {
					Context("in a happy-path scenario", func() {
						It("should create a VirtualService, a RequestAuthentication and AuthorizationPolicies", func() {
							cm := testConfigMap(helpers.JWT_HANDLER_ISTIO)
							err := c.Update(context.TODO(), cm)

							if apierrors.IsInvalid(err) {
								Fail(fmt.Sprintf("failed to update configmap, got an invalid object error: %v", err))
							}
							Expect(err).NotTo(HaveOccurred())

							expectedRequest := reconcile.Request{NamespacedName: types.NamespacedName{Name: cm.Name, Namespace: cm.Namespace}}
							Eventually(requests, eventuallyTimeout).Should(Receive(Equal(expectedRequest)))

							apiRuleName := generateTestName(testNameBase, testIDLength)
							serviceName := testServiceNameBase
							serviceHost := "httpbin-istio-jwt-happy-base.kyma.local"

							rule1 := testRule("/img", []string{"GET"}, nil, testIstioJWTHandlerWithScopes(testIssuer, testJwksUri, []string{"scope-a", "scope-b"}))
							rule2 := testRule("/headers", []string{"GET"}, nil, testIstioJWTHandlerWithScopes(testIssuer, testJwksUri, []string{"scope-c"}))
							instance := testInstance(apiRuleName, testNamespace, serviceName, serviceHost, testServicePort, []gatewayv1beta1.Rule{rule1, rule2})

							err = c.Create(context.TODO(), instance)
							if apierrors.IsInvalid(err) {
								Fail(fmt.Sprintf("failed to create object, got an invalid object error: %v", err))
								return
							}
							Expect(err).NotTo(HaveOccurred())
							defer func() {
								err := c.Delete(context.TODO(), instance)
								Expect(err).NotTo(HaveOccurred())
							}()

							expectedRequest = reconcile.Request{NamespacedName: types.NamespacedName{Name: apiRuleName, Namespace: testNamespace}}
							Eventually(requests, eventuallyTimeout).Should(Receive(Equal(expectedRequest)))

							matchingLabels := matchingLabelsFunc(apiRuleName, testNamespace)

							//Verify VirtualService
							vsList := networkingv1beta1.VirtualServiceList{}
							Eventually(func(g Gomega) {
								err = c.List(context.TODO(), &vsList, matchingLabels)
								g.Expect(err).NotTo(HaveOccurred())
								g.Expect(vsList.Items).To(HaveLen(1))
							}, eventuallyTimeout).Should(Succeed())

							vs := vsList.Items[0]

							expectedSpec := builders.VirtualServiceSpec().
								Host(serviceHost).
								Gateway(testGatewayURL).
								HTTP(builders.HTTPRoute().
									Match(builders.MatchRequest().Uri().Regex("/img")).
									Route(builders.RouteDestination().Host(fmt.Sprintf("%s.%s.svc.cluster.local", serviceName, testNamespace)).Port(testServicePort)).
									Headers(builders.Headers().SetHostHeader(serviceHost)).
									CorsPolicy(defaultCorsPolicy).
									Timeout(time.Second * helpers.DEFAULT_HTTP_TIMEOUT)).
								HTTP(builders.HTTPRoute().
									Match(builders.MatchRequest().Uri().Regex("/headers")).
									Route(builders.RouteDestination().Host(fmt.Sprintf("%s.%s.svc.cluster.local", serviceName, testNamespace)).Port(testServicePort)).
									Headers(builders.Headers().SetHostHeader(serviceHost)).
									CorsPolicy(defaultCorsPolicy).
									Timeout(time.Second * helpers.DEFAULT_HTTP_TIMEOUT))
							gotSpec := *expectedSpec.Get()
							Expect(*vs.Spec.DeepCopy()).To(Equal(*gotSpec.DeepCopy()))

							// Verify RequestAuthentication
							raList := securityv1beta1.RequestAuthenticationList{}
							Eventually(func(g Gomega) {
								err = c.List(context.TODO(), &raList, matchingLabels)
								Expect(err).NotTo(HaveOccurred())
								Expect(raList.Items).To(HaveLen(1))
							}, eventuallyTimeout).Should(Succeed())

							ra := raList.Items[0]

							Expect(ra.Spec.Selector.MatchLabels).To(BeEquivalentTo(map[string]string{"app": serviceName}))
							Expect(ra.Spec.JwtRules[0].Issuer).To(Equal(testIssuer))
							Expect(ra.Spec.JwtRules[0].JwksUri).To(Equal(testJwksUri))

							// Verify AuthorizationPolicies
							apList := securityv1beta1.AuthorizationPolicyList{}
							Eventually(func(g Gomega) {
								err = c.List(context.TODO(), &apList, matchingLabels)
								g.Expect(err).NotTo(HaveOccurred())
								g.Expect(apList.Items).To(HaveLen(2))
							}, eventuallyTimeout).Should(Succeed())

							getByOperationPath := func(apList []*securityv1beta1.AuthorizationPolicy, path string) (*securityv1beta1.AuthorizationPolicy, error) {
								for _, ap := range apList {
									if ap.Spec.Rules[0].To[0].Operation.Paths[0] == path {
										return ap, nil
									}
								}
								return nil, fmt.Errorf("no authorization policy with operation path %s exists", path)
							}

							hasAuthorizationPolicyWithOperationPath := func(apList []*securityv1beta1.AuthorizationPolicy, operationPath string, assertWhen func(*securityv1beta1.AuthorizationPolicy)) {
								ap, err := getByOperationPath(apList, operationPath)
								Expect(err).NotTo(HaveOccurred())
								Expect(ap.Spec.Selector.MatchLabels).To(BeEquivalentTo(map[string]string{"app": serviceName}))
								Expect(ap.Spec.Rules).To(HaveLen(3))

								for i := 0; i < 3; i++ {
									Expect(ap.Spec.Rules[i].From[0].Source.RequestPrincipals[0]).To(Equal("*"))
									Expect(ap.Spec.Rules[i].To[0].Operation.Paths[0]).To(Equal(operationPath))
									Expect(ap.Spec.Rules[i].To[0].Operation.Methods).To(BeEquivalentTo([]string{"GET"}))
								}

								ruleWhenKeys := append([]string{}, ap.Spec.Rules[0].When[0].Key, ap.Spec.Rules[1].When[0].Key, ap.Spec.Rules[2].When[0].Key)
								Expect(ruleWhenKeys).To(ContainElements("request.auth.claims[scp]", "request.auth.claims[scope]", "request.auth.claims[scopes]"))

								assertWhen(ap)
							}

							hasAuthorizationPolicyWithOperationPath(apList.Items, "/img", func(ap *securityv1beta1.AuthorizationPolicy) {
								Expect(ap.Spec.Rules).To(HaveLen(3))
								for i := 0; i < 3; i++ {
									Expect(ap.Spec.Rules[i].When).To(HaveLen(2))

									ruleWhenValues := append([]string{}, ap.Spec.Rules[0].When[0].Values...)
									ruleWhenValues = append(ruleWhenValues, ap.Spec.Rules[0].When[1].Values...)

									Expect(ruleWhenValues).To(ContainElements("scope-a", "scope-b"))

								}
							})

							hasAuthorizationPolicyWithOperationPath(apList.Items, "/headers", func(ap *securityv1beta1.AuthorizationPolicy) {
								Expect(ap.Spec.Rules).To(HaveLen(3))
								for i := 0; i < 3; i++ {
									Expect(ap.Spec.Rules[i].When[0].Values).To(ContainElements("scope-c"))
									Expect(ap.Spec.Rules[i].When[0].Values).To(ContainElements("scope-c"))
									Expect(ap.Spec.Rules[i].When[0].Values).To(ContainElements("scope-c"))
								}
							})

						})

						It("should create and update authorization policies when adding new authorization", func() {
							// given
							cm := testConfigMap(helpers.JWT_HANDLER_ISTIO)
							err := c.Update(context.TODO(), cm)

							if apierrors.IsInvalid(err) {
								Fail(fmt.Sprintf("failed to update configmap, got an invalid object error: %v", err))
							}
							Expect(err).NotTo(HaveOccurred())

							expectedRequest := reconcile.Request{NamespacedName: types.NamespacedName{Name: cm.Name, Namespace: cm.Namespace}}
							Eventually(requests, eventuallyTimeout).Should(Receive(Equal(expectedRequest)))

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
							instance := testInstance(apiRuleName, testNamespace, serviceName, serviceHost, testServicePort, []gatewayv1beta1.Rule{rule})

							err = c.Create(context.TODO(), instance)
							if apierrors.IsInvalid(err) {
								Fail(fmt.Sprintf("failed to create object, got an invalid object error: %v", err))
								return
							}
							Expect(err).NotTo(HaveOccurred())
							defer func() {
								err := c.Delete(context.TODO(), instance)
								Expect(err).NotTo(HaveOccurred())
							}()

							expectedRequest = reconcile.Request{NamespacedName: types.NamespacedName{Name: apiRuleName, Namespace: testNamespace}}
							Eventually(requests, eventuallyTimeout).Should(Receive(Equal(expectedRequest)))

							Eventually(func(g Gomega) {
								createdApiRule := gatewayv1beta1.APIRule{}
								err = c.Get(context.TODO(), client.ObjectKey{Name: apiRuleName, Namespace: testNamespace}, &createdApiRule)
								g.Expect(err).NotTo(HaveOccurred())
								g.Expect(createdApiRule.Status.APIRuleStatus.Code).To(Equal(gatewayv1beta1.StatusOK))
								g.Expect(createdApiRule.Status.VirtualServiceStatus.Code).To(Equal(gatewayv1beta1.StatusOK))
								g.Expect(createdApiRule.Status.AuthorizationPolicyStatus.Code).To(Equal(gatewayv1beta1.StatusOK))
								g.Expect(createdApiRule.Status.RequestAuthenticationStatus.Code).To(Equal(gatewayv1beta1.StatusOK))
							}, eventuallyTimeout).Should(Succeed())

							// when
							updatedApiRule := gatewayv1beta1.APIRule{}
							err = c.Get(context.TODO(), client.ObjectKey{Name: apiRuleName, Namespace: testNamespace}, &updatedApiRule)
							Expect(err).NotTo(HaveOccurred())

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

							err = c.Update(context.TODO(), &updatedApiRule)
							Expect(err).NotTo(HaveOccurred())

							apiRuleUpdated := reconcile.Request{NamespacedName: types.NamespacedName{Name: apiRuleName, Namespace: testNamespace}}
							Eventually(requests, eventuallyTimeout).Should(Receive(Equal(apiRuleUpdated)))

							// then
							matchingLabels := matchingLabelsFunc(apiRuleName, testNamespace)

							apList := securityv1beta1.AuthorizationPolicyList{}
							Eventually(func(g Gomega) {
								err = c.List(context.TODO(), &apList, matchingLabels)
								g.Expect(err).NotTo(HaveOccurred())
								g.Expect(apList.Items).To(HaveLen(3))
							}, eventuallyTimeout).Should(Succeed())

							scopeAScopeCMatcher := getAuthorizationPolicyWhenScopeMatcher("scope-a", "scope-c")
							scopeAScopeDMatcher := getAuthorizationPolicyWhenScopeMatcher("scope-a", "scope-d")
							scopeDScopeBMatcher := getAuthorizationPolicyWhenScopeMatcher("scope-d", "scope-b")

							Expect(apList.Items).To(ContainElement(scopeAScopeCMatcher))
							Expect(apList.Items).To(ContainElement(scopeAScopeDMatcher))
							Expect(apList.Items).To(ContainElement(scopeDScopeBMatcher))
						})
					})
				})
			})
		})

		Context("on specified paths", func() {
			Context("with multiple endpoints secured with different authentication methods", func() {
				Context("in the happy path scenario", func() {
					It("should create a VS with corresponding matchers and access rules for each secured path", func() {
						jwtHandler := testOryJWTHandler(testIssuer, defaultScopes)
						oauthHandler := testOauthHandler(defaultScopes)
						rule1 := testRule("/img", []string{"GET"}, defaultMutators, jwtHandler)
						rule2 := testRule("/headers", []string{"GET"}, defaultMutators, oauthHandler)
						rule3 := testRule("/status", []string{"GET"}, defaultMutators, noConfigHandler("noop"))
						rule4 := testRule("/favicon", []string{"GET"}, nil, noConfigHandler("allow"))

						apiRuleName := generateTestName(testNameBase, testIDLength)
						serviceName := testServiceNameBase
						serviceHost := "httpbin4.kyma.local"

						instance := testInstance(apiRuleName, testNamespace, serviceName, serviceHost, testServicePort, []gatewayv1beta1.Rule{rule1, rule2, rule3, rule4})

						err := c.Create(context.TODO(), instance)
						if apierrors.IsInvalid(err) {
							Fail(fmt.Sprintf("failed to create object, got an invalid object error: %v", err))
							return
						}
						Expect(err).NotTo(HaveOccurred())
						defer func() {
							err := c.Delete(context.TODO(), instance)
							Expect(err).NotTo(HaveOccurred())
						}()

						expectedRequest := reconcile.Request{NamespacedName: types.NamespacedName{Name: apiRuleName, Namespace: testNamespace}}
						Eventually(requests, eventuallyTimeout).Should(Receive(Equal(expectedRequest)))

						matchingLabels := matchingLabelsFunc(apiRuleName, testNamespace)

						//Verify VirtualService

						vsList := networkingv1beta1.VirtualServiceList{}
						Eventually(func(g Gomega) {
							err = c.List(context.TODO(), &vsList, matchingLabels)
							g.Expect(err).NotTo(HaveOccurred())
							g.Expect(vsList.Items).To(HaveLen(1))
						}, eventuallyTimeout).Should(Succeed())

						vs := vsList.Items[0]

						expectedSpec := builders.VirtualServiceSpec().
							Host(serviceHost).
							Gateway(testGatewayURL).
							HTTP(builders.HTTPRoute().
								Match(builders.MatchRequest().Uri().Regex("/img")).
								Route(builders.RouteDestination().Host(testOathkeeperSvcURL).Port(testOathkeeperPort)).
								Headers(builders.Headers().SetHostHeader(serviceHost)).
								CorsPolicy(defaultCorsPolicy).
								Timeout(time.Second * helpers.DEFAULT_HTTP_TIMEOUT)).
							HTTP(builders.HTTPRoute().
								Match(builders.MatchRequest().Uri().Regex("/headers")).
								Route(builders.RouteDestination().Host(testOathkeeperSvcURL).Port(testOathkeeperPort)).
								Headers(builders.Headers().SetHostHeader(serviceHost)).
								CorsPolicy(defaultCorsPolicy).
								Timeout(time.Second * helpers.DEFAULT_HTTP_TIMEOUT)).
							HTTP(builders.HTTPRoute().
								Match(builders.MatchRequest().Uri().Regex("/status")).
								Route(builders.RouteDestination().Host(testOathkeeperSvcURL).Port(testOathkeeperPort)).
								Headers(builders.Headers().SetHostHeader(serviceHost)).
								CorsPolicy(defaultCorsPolicy).
								Timeout(time.Second * helpers.DEFAULT_HTTP_TIMEOUT)).
							HTTP(builders.HTTPRoute().
								Match(builders.MatchRequest().Uri().Regex("/favicon")).
								Route(builders.RouteDestination().Host("httpbin.atgo-system.svc.cluster.local").Port(443)). // "allow", no oathkeeper rule!
								Headers(builders.Headers().SetHostHeader(serviceHost)).
								CorsPolicy(defaultCorsPolicy).
								Timeout(time.Second * helpers.DEFAULT_HTTP_TIMEOUT))

						gotSpec := *expectedSpec.Get()
						Expect(*vs.Spec.DeepCopy()).To(Equal(*gotSpec.DeepCopy()))

						//Verify Rules
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
								g.Expect(rlList).To(HaveLen(3))
							}, eventuallyTimeout).Should(Succeed())

							rules := make(map[string]rulev1alpha1.Rule)

							for _, rule := range rlList {
								rules[rule.Spec.Match.URL] = rule
							}

							rl := rules[expectedRuleMatchURL]

							//Spec.Upstream
							Expect(rl.Spec.Upstream).NotTo(BeNil())
							Expect(rl.Spec.Upstream.URL).To(Equal(fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", serviceName, testNamespace, testServicePort)))
							Expect(rl.Spec.Upstream.StripPath).To(BeNil())
							Expect(rl.Spec.Upstream.PreserveHost).To(BeNil())

							//Spec.Match
							Expect(rl.Spec.Match).NotTo(BeNil())
							Expect(rl.Spec.Match.Methods).To(Equal([]string{"GET"}))
							Expect(rl.Spec.Match.URL).To(Equal(expectedRuleMatchURL))

							//Spec.Authenticators
							Expect(rl.Spec.Authenticators).To(HaveLen(1))
							Expect(rl.Spec.Authenticators[0].Handler).NotTo(BeNil())
							Expect(rl.Spec.Authenticators[0].Handler.Name).To(Equal(tc.handler))

							if tc.config != nil {
								//Authenticators[0].Handler.Config validation
								Expect(rl.Spec.Authenticators[0].Handler.Config).NotTo(BeNil())
								Expect(rl.Spec.Authenticators[0].Handler.Config.Raw).To(MatchJSON(tc.config))
							}

							//Spec.Authorizer
							Expect(rl.Spec.Authorizer).NotTo(BeNil())
							Expect(rl.Spec.Authorizer.Handler).NotTo(BeNil())
							Expect(rl.Spec.Authorizer.Handler.Name).To(Equal("allow"))
							Expect(rl.Spec.Authorizer.Handler.Config).To(BeNil())

							//Spec.Mutators
							Expect(rl.Spec.Mutators).NotTo(BeNil())
							Expect(len(rl.Spec.Mutators)).To(Equal(len(defaultMutators)))
							Expect(rl.Spec.Mutators[0].Handler.Name).To(Equal(defaultMutators[0].Name))
							Expect(rl.Spec.Mutators[1].Handler.Name).To(Equal(defaultMutators[1].Name))
						}

						//make sure no rule for "/favicon" path has been created
						name := fmt.Sprintf("%s-%s-3", apiRuleName, serviceName)
						Expect(c.Get(context.TODO(), client.ObjectKey{Name: name, Namespace: testNamespace}, &rulev1alpha1.Rule{})).To(HaveOccurred())
					})
				})
			})
		})
	})

	Context("Changing JWT handler in config map", func() {
		Context("Handler is ory and ApiRule with JWT handler rule exists", func() {
			Context("changing jwt handler to istio", func() {
				It("Should have validation errors for APiRule JWT handler configuration and rule is not deleted", func() {
					// given
					By("Setting JWT handler config map to ory")
					cm := testConfigMap(helpers.JWT_HANDLER_ORY)
					err := c.Update(context.TODO(), cm)
					Expect(err).NotTo(HaveOccurred())

					apiRuleName := generateTestName(testNameBase, testIDLength)
					testServiceHost := fmt.Sprintf("httpbin-%s.kyma.local", apiRuleName)

					rule := testRule("/img", []string{"GET"}, nil, testOryJWTHandler(testIssuer, defaultScopes))
					instance := testInstance(apiRuleName, testNamespace, testServiceNameBase, testServiceHost, testServicePort, []gatewayv1beta1.Rule{rule})

					By("Create ApiRule with Rule using JWT handler")
					err = c.Create(context.TODO(), instance)
					Expect(err).NotTo(HaveOccurred())
					defer func() {
						err := c.Delete(context.TODO(), instance)
						Expect(err).NotTo(HaveOccurred())
					}()

					initialStateReq := reconcile.Request{NamespacedName: types.NamespacedName{Name: apiRuleName, Namespace: testNamespace}}
					Eventually(requests, eventuallyTimeout).Should(Receive(Equal(initialStateReq)))

					By("Waiting until reconciliation of API Rule is finished")
					Eventually(func(g Gomega) {
						apiRule := gatewayv1beta1.APIRule{}
						err = c.Get(context.TODO(), client.ObjectKey{Name: apiRuleName, Namespace: testNamespace}, &apiRule)
						Expect(err).NotTo(HaveOccurred())

						Expect(apiRule.Status.APIRuleStatus.Code).To(Equal(gatewayv1beta1.StatusOK))
					}, eventuallyTimeout).Should(Succeed())

					By("Updating JWT handler config map to istio")
					cm = testConfigMap(helpers.JWT_HANDLER_ISTIO)
					err = c.Update(context.TODO(), cm)
					Expect(err).NotTo(HaveOccurred())

					cmChangedReq := reconcile.Request{NamespacedName: types.NamespacedName{Name: cm.Name, Namespace: cm.Namespace}}
					Eventually(requests, eventuallyTimeout).Should(Receive(Equal(cmChangedReq)))

					By("Waiting until reconciliation of CM is finished")
					Eventually(func(g Gomega) {
						err = c.Get(context.TODO(), client.ObjectKey{Name: cm.Name, Namespace: cm.Namespace}, cm)
						Expect(err).NotTo(HaveOccurred())

						Expect(cm.Data).To(HaveKeyWithValue(helpers.CM_KEY, "jwtHandler: istio"))
					}, eventuallyTimeout).Should(Succeed())

					// when
					triggerApiRuleReconciliation(apiRuleName)

					// then
					Eventually(func(g Gomega) {
						shouldHaveRules(g, apiRuleName, testNamespace, 1)
					}, eventuallyTimeout).Should(Succeed())

					apiRule := gatewayv1beta1.APIRule{}
					err = c.Get(context.TODO(), client.ObjectKey{Name: apiRuleName, Namespace: testNamespace}, &apiRule)
					Expect(err).NotTo(HaveOccurred())

					Expect(apiRule.Status.APIRuleStatus.Code).To(Equal(gatewayv1beta1.StatusError))
					Expect(apiRule.Status.APIRuleStatus.Description).To(ContainSubstring("Validation error"))
				})

				It("Should create AP and RA and delete JWT Access Rule when ApiRule JWT handler configuration was updated to have valid config for istio", func() {
					// given
					By("Setting JWT handler config map to ory")
					cm := testConfigMap(helpers.JWT_HANDLER_ORY)
					err := c.Update(context.TODO(), cm)
					Expect(err).NotTo(HaveOccurred())

					apiRuleName := generateTestName(testNameBase, testIDLength)
					testServiceHost := fmt.Sprintf("httpbin-%s.kyma.local", apiRuleName)

					rule := testRule("/img", []string{"GET"}, nil, testOryJWTHandler(testIssuer, defaultScopes))
					apiRule := testInstance(apiRuleName, testNamespace, testServiceNameBase, testServiceHost, testServicePort, []gatewayv1beta1.Rule{rule})

					By("Create ApiRule with Rule using JWT handler")
					err = c.Create(context.TODO(), apiRule)
					Expect(err).NotTo(HaveOccurred())
					defer func() {
						err := c.Delete(context.TODO(), apiRule)
						Expect(err).NotTo(HaveOccurred())
					}()

					initialStateReq := reconcile.Request{NamespacedName: types.NamespacedName{Name: apiRuleName, Namespace: testNamespace}}
					Eventually(requests, eventuallyTimeout).Should(Receive(Equal(initialStateReq)))

					By("Updating JWT handler config map to istio")
					cm = testConfigMap(helpers.JWT_HANDLER_ISTIO)
					err = c.Update(context.TODO(), cm)
					Expect(err).NotTo(HaveOccurred())

					cmRequest := reconcile.Request{NamespacedName: types.NamespacedName{Name: cm.Name, Namespace: cm.Namespace}}
					Eventually(requests, eventuallyTimeout).Should(Receive(Equal(cmRequest)))

					// when
					By("Updating JWT handler in ApiRule to be valid for istio")
					istioJwtRule := testRule("/img", []string{"GET"}, nil, testIstioJWTHandler(testIssuer, testJwksUri))
					updatedApiRule := gatewayv1beta1.APIRule{}
					err = c.Get(context.TODO(), client.ObjectKey{Name: apiRuleName, Namespace: testNamespace}, &updatedApiRule)
					Expect(err).NotTo(HaveOccurred())
					updatedApiRule.Spec.Rules = []gatewayv1beta1.Rule{istioJwtRule}
					err = c.Update(context.TODO(), &updatedApiRule)
					Expect(err).NotTo(HaveOccurred())

					updateApiRuleReq := reconcile.Request{NamespacedName: types.NamespacedName{Name: apiRuleName, Namespace: testNamespace}}
					Eventually(requests, eventuallyTimeout).Should(Receive(Equal(updateApiRuleReq)))

					// then
					Eventually(func(g Gomega) {
						shouldHaveRequestAuthentications(g, apiRuleName, testNamespace, 1)
					}, eventuallyTimeout).Should(Succeed())

					Eventually(func(g Gomega) {
						shouldHaveAuthorizationPolicies(g, apiRuleName, testNamespace, 1)
					}, eventuallyTimeout).Should(Succeed())

					Eventually(func(g Gomega) {
						shouldHaveRules(g, apiRuleName, testNamespace, 0)
					}, eventuallyTimeout).Should(Succeed())

					expectedApiRule := gatewayv1beta1.APIRule{}
					err = c.Get(context.TODO(), client.ObjectKey{Name: apiRuleName, Namespace: testNamespace}, &expectedApiRule)
					Expect(err).NotTo(HaveOccurred())

					Expect(expectedApiRule.Status.APIRuleStatus.Code).To(Equal(gatewayv1beta1.StatusOK))
				})

			})
		})

		Context("Handler is istio and ApiRule with JWT handler specific resources exists", func() {

			Context("changing jwt handler to ory", func() {
				It("Should have validation errors for APiRule JWT handler configuration and resources are not deleted", func() {
					// given
					By("Setting JWT handler config map to istio")
					cm := testConfigMap(helpers.JWT_HANDLER_ISTIO)
					err := c.Update(context.TODO(), cm)
					Expect(err).NotTo(HaveOccurred())

					expectedRequest := reconcile.Request{NamespacedName: types.NamespacedName{Name: cm.Name, Namespace: cm.Namespace}}
					Eventually(requests, eventuallyTimeout).Should(Receive(Equal(expectedRequest)))

					apiRuleName := generateTestName(testNameBase, testIDLength)
					testServiceHost := fmt.Sprintf("httpbin-%s.kyma.local", apiRuleName)

					rule := testRule("/img", []string{"GET"}, nil, testIstioJWTHandler(testIssuer, testJwksUri))
					instance := testInstance(apiRuleName, testNamespace, testServiceNameBase, testServiceHost, testServicePort, []gatewayv1beta1.Rule{rule})

					By("Create ApiRule with Rule using JWT handler")
					err = c.Create(context.TODO(), instance)
					Expect(err).NotTo(HaveOccurred())
					defer func() {
						err := c.Delete(context.TODO(), instance)
						Expect(err).NotTo(HaveOccurred())
					}()

					initialStateReq := reconcile.Request{NamespacedName: types.NamespacedName{Name: apiRuleName, Namespace: testNamespace}}
					Eventually(requests, eventuallyTimeout).Should(Receive(Equal(initialStateReq)))

					By("Waiting until reconciliation of API Rule is finished")
					Eventually(func(g Gomega) {
						apiRule := gatewayv1beta1.APIRule{}
						err = c.Get(context.TODO(), client.ObjectKey{Name: apiRuleName, Namespace: testNamespace}, &apiRule)
						Expect(err).NotTo(HaveOccurred())

						Expect(apiRule.Status.APIRuleStatus.Code).To(Equal(gatewayv1beta1.StatusOK))
					}, eventuallyTimeout).Should(Succeed())

					By("Updating JWT handler config map to ory")
					cm = testConfigMap(helpers.JWT_HANDLER_ORY)
					err = c.Update(context.TODO(), cm)
					Expect(err).NotTo(HaveOccurred())

					cmChangedReq := reconcile.Request{NamespacedName: types.NamespacedName{Name: cm.Name, Namespace: cm.Namespace}}
					Eventually(requests, eventuallyTimeout).Should(Receive(Equal(cmChangedReq)))

					By("Waiting until reconciliation of CM is finished")
					Eventually(func(g Gomega) {
						err = c.Get(context.TODO(), client.ObjectKey{Name: cm.Name, Namespace: cm.Namespace}, cm)
						Expect(err).NotTo(HaveOccurred())

						Expect(cm.Data).To(HaveKeyWithValue(helpers.CM_KEY, "jwtHandler: ory"))
					}, eventuallyTimeout).Should(Succeed())

					// when
					triggerApiRuleReconciliation(apiRuleName)

					// then
					Eventually(func(g Gomega) {
						shouldHaveRequestAuthentications(g, apiRuleName, testNamespace, 1)
					}, eventuallyTimeout).Should(Succeed())

					Eventually(func(g Gomega) {
						shouldHaveAuthorizationPolicies(g, apiRuleName, testNamespace, 1)
					}, eventuallyTimeout).Should(Succeed())

					apiRule := gatewayv1beta1.APIRule{}
					err = c.Get(context.TODO(), client.ObjectKey{Name: apiRuleName, Namespace: testNamespace}, &apiRule)
					Expect(err).NotTo(HaveOccurred())

					Expect(apiRule.Status.APIRuleStatus.Code).To(Equal(gatewayv1beta1.StatusError))
					Expect(apiRule.Status.APIRuleStatus.Description).To(ContainSubstring("Validation error"))
				})

				It("Should create Access Rule and delete RA and AP when ApiRule JWT handler configuration was updated to have valid config for ory", func() {
					// given
					By("Setting JWT handler config map to istio")
					cm := testConfigMap(helpers.JWT_HANDLER_ISTIO)
					err := c.Update(context.TODO(), cm)
					Expect(err).NotTo(HaveOccurred())

					cmRequest := reconcile.Request{NamespacedName: types.NamespacedName{Name: cm.Name, Namespace: cm.Namespace}}
					Eventually(requests, eventuallyTimeout).Should(Receive(Equal(cmRequest)))

					apiRuleName := generateTestName(testNameBase, testIDLength)
					testServiceHost := fmt.Sprintf("httpbin-%s.kyma.local", apiRuleName)

					rule := testRule("/img", []string{"GET"}, nil, testIstioJWTHandler(testIssuer, testJwksUri))
					apiRule := testInstance(apiRuleName, testNamespace, testServiceNameBase, testServiceHost, testServicePort, []gatewayv1beta1.Rule{rule})

					By("Create ApiRule with Rule using JWT handler")
					err = c.Create(context.TODO(), apiRule)
					Expect(err).NotTo(HaveOccurred())
					defer func() {
						err := c.Delete(context.TODO(), apiRule)
						Expect(err).NotTo(HaveOccurred())
					}()

					initialStateReq := reconcile.Request{NamespacedName: types.NamespacedName{Name: apiRuleName, Namespace: testNamespace}}
					Eventually(requests, eventuallyTimeout).Should(Receive(Equal(initialStateReq)))

					By("Updating JWT handler config map to ory")
					cm = testConfigMap(helpers.JWT_HANDLER_ORY)
					err = c.Update(context.TODO(), cm)
					Expect(err).NotTo(HaveOccurred())

					cmRequest = reconcile.Request{NamespacedName: types.NamespacedName{Name: cm.Name, Namespace: cm.Namespace}}
					Eventually(requests, eventuallyTimeout).Should(Receive(Equal(cmRequest)))

					// when
					By("Updating JWT handler in ApiRule to be valid for ory")
					oryJwtRule := testRule("/img", []string{"GET"}, nil, testOryJWTHandler(testIssuer, defaultScopes))
					updatedApiRule := gatewayv1beta1.APIRule{}
					err = c.Get(context.TODO(), client.ObjectKey{Name: apiRuleName, Namespace: testNamespace}, &updatedApiRule)
					Expect(err).NotTo(HaveOccurred())
					updatedApiRule.Spec.Rules = []gatewayv1beta1.Rule{oryJwtRule}
					err = c.Update(context.TODO(), &updatedApiRule)
					Expect(err).NotTo(HaveOccurred())

					updateApiRuleReq := reconcile.Request{NamespacedName: types.NamespacedName{Name: apiRuleName, Namespace: testNamespace}}
					Eventually(requests, eventuallyTimeout).Should(Receive(Equal(updateApiRuleReq)))

					// then
					Eventually(func(g Gomega) {
						shouldHaveRequestAuthentications(g, apiRuleName, testNamespace, 0)
					}, eventuallyTimeout).Should(Succeed())

					Eventually(func(g Gomega) {
						shouldHaveAuthorizationPolicies(g, apiRuleName, testNamespace, 0)
					}, eventuallyTimeout).Should(Succeed())

					Eventually(func(g Gomega) {
						shouldHaveRules(g, apiRuleName, testNamespace, 1)
					}, eventuallyTimeout).Should(Succeed())

					expectedApiRule := gatewayv1beta1.APIRule{}
					err = c.Get(context.TODO(), client.ObjectKey{Name: apiRuleName, Namespace: testNamespace}, &expectedApiRule)
					Expect(err).NotTo(HaveOccurred())

					Expect(expectedApiRule.Status.APIRuleStatus.Code).To(Equal(gatewayv1beta1.StatusOK))
				})
			})
		})
	})
})

func add(mgr manager.Manager, r reconcile.Reconciler) error {
	c, err := controller.New("api-gateway-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &gatewayv1beta1.APIRule{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &corev1.ConfigMap{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
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

func testInstance(name, namespace, serviceName, serviceHost string, servicePort uint32, rules []gatewayv1beta1.Rule) *gatewayv1beta1.APIRule {
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
				Name: &serviceName,
				Port: &servicePort,
			},
			Rules: rules,
		},
	}
}

func testConfigMap(jwtHandler string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      helpers.CM_NAME,
			Namespace: helpers.CM_NS,
		},
		Data: map[string]string{
			helpers.CM_KEY: fmt.Sprintf("jwtHandler: %s", jwtHandler),
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

	rand.Seed(time.Now().UnixNano())

	letterRunes := []rune("abcdefghijklmnopqrstuvwxyz")

	b := make([]rune, length)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return name + "-" + string(b)
}

func getRuleList(g Gomega, matchingLabels client.ListOption) []rulev1alpha1.Rule {
	res := rulev1alpha1.RuleList{}
	err := c.List(context.TODO(), &res, matchingLabels)
	g.Expect(err).NotTo(HaveOccurred())
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

func triggerApiRuleReconciliation(apiRuleName string) {
	By("Trigger Reconcile of ApiRule")
	reconciledApiRule := gatewayv1beta1.APIRule{}
	err := c.Get(context.TODO(), client.ObjectKey{Name: apiRuleName, Namespace: testNamespace}, &reconciledApiRule)
	// We check for a different generation before reconciliation of APiRule, therefore we need to do a dummy change to
	// trigger the reconciliation of the ApiRule.
	newDummyHost := fmt.Sprintf("dummy-change.%s", *reconciledApiRule.Spec.Host)
	reconciledApiRule.Spec.Host = &newDummyHost
	Expect(err).NotTo(HaveOccurred())
	err = c.Update(context.TODO(), &reconciledApiRule)
	Expect(err).NotTo(HaveOccurred())

	reconcileApiReq := reconcile.Request{NamespacedName: types.NamespacedName{Name: apiRuleName, Namespace: testNamespace}}
	Eventually(requests, eventuallyTimeout).Should(Receive(Equal(reconcileApiReq)))
}

func setHandlerConfigMap(handler string) {
	cm := testConfigMap(handler)
	err := c.Update(context.TODO(), cm)
	if apierrors.IsInvalid(err) {
		Fail(fmt.Sprintf("failed to update configmap, got an invalid object error: %v", err))
	}
	Expect(err).NotTo(HaveOccurred())
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
