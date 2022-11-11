package controllers_test

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/kyma-incubator/api-gateway/internal/builders"
	"github.com/kyma-incubator/api-gateway/internal/helpers"
	"github.com/kyma-incubator/api-gateway/internal/processing"
	istioint "github.com/kyma-incubator/api-gateway/internal/types/istio"

	"encoding/json"

	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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

const (
	timeout = time.Second * 5

	kind                        = "APIRule"
	testGatewayURL              = "kyma-system/kyma-gateway"
	testOathkeeperSvcURL        = "oathkeeper.kyma-system.svc.cluster.local"
	testOathkeeperPort   uint32 = 1234
	testNamespace               = "atgo-system"
	testNameBase                = "test"
	testIDLength                = 5
)

var _ = Describe("APIRule Controller", func() {
	const testServiceName = "httpbin"
	const testServicePort uint32 = 443
	const testPath = "/.*"
	var testIssuer = "https://oauth2.example.com/"
	var testJwksUri = "https://oauth2.example.com/.well-known/jwks.json"
	var testMethods = []string{"GET", "PUT"}
	var testScopes = []string{"foo", "bar"}
	var testMutators = []*gatewayv1beta1.Mutator{
		{
			Handler: noConfigHandler("noop"),
		},
		{
			Handler: noConfigHandler("idToken"),
		},
	}

	var corsPolicyBuilder = builders.CorsPolicy().
		AllowHeaders(TestAllowHeaders...).
		AllowMethods(TestAllowMethods...).
		AllowOrigins(TestAllowOrigins...)

	BeforeEach(func() {
		// We configure `ory` in ConfigMap as the default for all tests
		cm := testConfigMap(helpers.JWT_HANDLER_ORY)
		err := c.Update(context.TODO(), cm)
		if apierrors.IsInvalid(err) {
			Fail(fmt.Sprintf("failed to update configmap, got an invalid object error: %v", err))
		}
		Expect(err).NotTo(HaveOccurred())
	})

	Context("when updating the APIRule with multiple paths", func() {

		It("should create, update and delete rules depending on patch match", func() {
			rule1 := testRule("/rule1", []string{"GET"}, testMutators, noConfigHandler("noop"))
			rule2 := testRule("/rule2", []string{"PUT"}, testMutators, noConfigHandler("unauthorized"))
			rule3 := testRule("/rule3", []string{"DELETE"}, testMutators, noConfigHandler("anonymous"))

			apiRuleName := generateTestName(testNameBase, testIDLength)
			testServiceHost := "httpbin5.kyma.local"

			matchingLabels := matchingLabelsFunc(apiRuleName, testNamespace)

			pathToURLFunc := func(path string) string {
				return fmt.Sprintf("<http|https>://%s<%s>", testServiceHost, path)
			}

			By("Create APIRule")

			instance := testInstance(apiRuleName, testNamespace, testServiceName, testServiceHost, testServicePort, []gatewayv1beta1.Rule{rule1, rule2, rule3})
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
			Eventually(requests, timeout).Should(Receive(Equal(expectedRequest)))

			By("Verify before update")

			ruleList := getRuleList(matchingLabels)
			verifyRuleList(ruleList, pathToURLFunc, rule1, rule2, rule3)
			expectedUpstream := fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", testServiceName, testNamespace, testServicePort)
			//Verify All Rules point to original Service
			for i := range ruleList {
				r := ruleList[i]
				Expect(r.Spec.Upstream.URL).To(Equal(expectedUpstream))
			}

			By("Update APIRule")
			existingInstance := gatewayv1beta1.APIRule{}
			err = c.Get(context.TODO(), client.ObjectKey{Name: apiRuleName, Namespace: testNamespace}, &existingInstance)
			Expect(err).NotTo(HaveOccurred())
			rule4 := testRule("/rule4", []string{"POST"}, testMutators, noConfigHandler("cookie_session"))
			existingInstance.Spec.Rules = []gatewayv1beta1.Rule{rule1, rule4}
			newServiceName := testServiceName + "new"
			newServicePort := testServicePort + 3
			existingInstance.Spec.Service.Name = &newServiceName
			existingInstance.Spec.Service.Port = &newServicePort

			err = c.Update(context.TODO(), &existingInstance)
			Expect(err).NotTo(HaveOccurred())
			expectedRequest = reconcile.Request{NamespacedName: types.NamespacedName{Name: apiRuleName, Namespace: testNamespace}}
			Eventually(requests, timeout).Should(Receive(Equal(expectedRequest)))

			By("Verify after update")
			time.Sleep(1 * time.Second) //Otherwise K8s client fetches old Rules.
			ruleList = getRuleList(matchingLabels)
			verifyRuleList(ruleList, pathToURLFunc, rule1, rule4)
			//Verify All Rules point to new Service after update
			expectedUpstream = fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", newServiceName, testNamespace, newServicePort)
			for i := range ruleList {
				r := ruleList[i]
				Expect(r.Spec.Upstream.URL).To(Equal(expectedUpstream))
			}
		})
	})

	Context("when creating an APIRule for exposing service", func() {

		It("Should report validation errors in CR status", func() {

			invalidConfig := testOauthHandler(testScopes)
			invalidConfig.Name = "noop"

			apiRuleName := generateTestName(testNameBase, testIDLength)
			testServiceHost := "httpbin.kyma.local"
			rule := testRule(testPath, testMethods, testMutators, invalidConfig)
			instance := testInstance(apiRuleName, testNamespace, testServiceName, testServiceHost, testServicePort, []gatewayv1beta1.Rule{rule})
			instance.Spec.Rules = append(instance.Spec.Rules, instance.Spec.Rules[0]) //Duplicate entry
			instance.Spec.Rules = append(instance.Spec.Rules, instance.Spec.Rules[0]) //Duplicate entry

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

			Eventually(requests, timeout).Should(Receive(Equal(expectedRequest)))

			//Verify APIRule
			created := gatewayv1beta1.APIRule{}
			err = c.Get(context.TODO(), client.ObjectKey{Name: apiRuleName, Namespace: testNamespace}, &created)
			Expect(err).NotTo(HaveOccurred())
			Expect(created.Status.APIRuleStatus.Code).To(Equal(gatewayv1beta1.StatusError))
			Expect(created.Status.APIRuleStatus.Description).To(ContainSubstring("Multiple validation errors:"))
			Expect(created.Status.APIRuleStatus.Description).To(ContainSubstring("Attribute \".spec.rules\": multiple rules defined for the same path and method"))
			Expect(created.Status.APIRuleStatus.Description).To(ContainSubstring("Attribute \".spec.rules[0].accessStrategies[0].config\": strategy: noop does not support configuration"))
			Expect(created.Status.APIRuleStatus.Description).To(ContainSubstring("Attribute \".spec.rules[1].accessStrategies[0].config\": strategy: noop does not support configuration"))
			Expect(created.Status.APIRuleStatus.Description).To(ContainSubstring("1 more error(s)..."))

			//Verify VirtualService is not created
			vsList := networkingv1beta1.VirtualServiceList{}
			err = c.List(context.TODO(), &vsList, matchingLabelsFunc(apiRuleName, testNamespace))
			Expect(err).NotTo(HaveOccurred())
			Expect(vsList.Items).To(HaveLen(0))
		})

		Context("on all the paths,", func() {
			Context("secured with Oauth2 introspection,", func() {
				Context("in a happy-path scenario", func() {
					It("should create a VirtualService and an AccessRule", func() {
						apiRuleName := generateTestName(testNameBase, testIDLength)
						testServiceHost := "httpbin2.kyma.local"
						rule := testRule(testPath, testMethods, testMutators, testOauthHandler(testScopes))
						instance := testInstance(apiRuleName, testNamespace, testServiceName, testServiceHost, testServicePort, []gatewayv1beta1.Rule{rule})

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

						Eventually(requests, timeout).Should(Receive(Equal(expectedRequest)))

						matchingLabels := matchingLabelsFunc(apiRuleName, testNamespace)

						//Verify VirtualService
						vsList := networkingv1beta1.VirtualServiceList{}
						err = c.List(context.TODO(), &vsList, matchingLabels)
						Expect(err).NotTo(HaveOccurred())
						Expect(vsList.Items).To(HaveLen(1))
						vs := vsList.Items[0]

						//Meta
						Expect(vs.Name).To(HavePrefix(apiRuleName + "-"))
						Expect(len(vs.Name) > len(apiRuleName)).To(BeTrue())

						verifyOwnerReference(vs.ObjectMeta, apiRuleName, gatewayv1beta1.GroupVersion.String(), kind)

						expectedSpec := builders.VirtualServiceSpec().
							Host(testServiceHost).
							Gateway(testGatewayURL).
							HTTP(builders.HTTPRoute().
								Match(builders.MatchRequest().Uri().Regex(testPath)).
								Route(builders.RouteDestination().Host(testOathkeeperSvcURL).Port(testOathkeeperPort)).
								Headers(builders.Headers().SetHostHeader(testServiceHost)).
								CorsPolicy(corsPolicyBuilder))

						gotSpec := *expectedSpec.Get()
						Expect(*vs.Spec.DeepCopy()).To(Equal(*gotSpec.DeepCopy()))

						//Verify Rule
						expectedRuleMatchURL := fmt.Sprintf("<http|https>://%s<%s>", testServiceHost, testPath)

						rlList := getRuleList(matchingLabels)
						Expect(rlList).To(HaveLen(1))
						rl := rlList[0]
						Expect(rl.Spec.Match.URL).To(Equal(expectedRuleMatchURL))

						//Meta
						Expect(rl.Name).To(HavePrefix(apiRuleName + "-"))
						Expect(len(rl.Name) > len(apiRuleName)).To(BeTrue())

						verifyOwnerReference(rl.ObjectMeta, apiRuleName, gatewayv1beta1.GroupVersion.String(), kind)

						//Spec.Upstream
						Expect(rl.Spec.Upstream).NotTo(BeNil())
						Expect(rl.Spec.Upstream.URL).To(Equal(fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", testServiceName, testNamespace, testServicePort)))
						Expect(rl.Spec.Upstream.StripPath).To(BeNil())
						Expect(rl.Spec.Upstream.PreserveHost).To(BeNil())
						//Spec.Match
						Expect(rl.Spec.Match).NotTo(BeNil())
						Expect(rl.Spec.Match.URL).To(Equal(fmt.Sprintf("<http|https>://%s<%s>", testServiceHost, testPath)))
						Expect(rl.Spec.Match.Methods).To(Equal(testMethods))
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
						Expect(asStringSlice(handlerConfig["required_scope"])).To(BeEquivalentTo(testScopes))
						//Spec.Authorizer
						Expect(rl.Spec.Authorizer).NotTo(BeNil())
						Expect(rl.Spec.Authorizer.Handler).NotTo(BeNil())
						Expect(rl.Spec.Authorizer.Handler.Name).To(Equal("allow"))
						Expect(rl.Spec.Authorizer.Handler.Config).To(BeNil())

						//Spec.Mutators
						Expect(rl.Spec.Mutators).NotTo(BeNil())
						Expect(len(rl.Spec.Mutators)).To(Equal(len(testMutators)))
						Expect(rl.Spec.Mutators[0].Handler.Name).To(Equal(testMutators[0].Name))
						Expect(rl.Spec.Mutators[1].Handler.Name).To(Equal(testMutators[1].Name))
					})
				})
			})

			Context("secured with JWT token authentication,", func() {
				Context("with ORY as JWT handler,", func() {
					Context("in a happy-path scenario", func() {
						It("should create a VirtualService and an AccessRules", func() {
							apiRuleName := generateTestName(testNameBase, testIDLength)
							testServiceHost := "httpbin3.kyma.local"
							rule1 := testRule("/img", []string{"GET"}, testMutators, testOryJWTHandler(testIssuer, testScopes))
							rule2 := testRule("/headers", []string{"GET"}, testMutators, testOryJWTHandler(testIssuer, testScopes))
							instance := testInstance(apiRuleName, testNamespace, testServiceName, testServiceHost, testServicePort, []gatewayv1beta1.Rule{rule1, rule2})

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

							Eventually(requests, timeout).Should(Receive(Equal(expectedRequest)))

							matchingLabels := matchingLabelsFunc(apiRuleName, testNamespace)

							//Verify VirtualService
							vsList := networkingv1beta1.VirtualServiceList{}
							err = c.List(context.TODO(), &vsList, matchingLabels)
							Expect(err).NotTo(HaveOccurred())
							Expect(vsList.Items).To(HaveLen(1))
							vs := vsList.Items[0]

							//Meta
							verifyOwnerReference(vs.ObjectMeta, apiRuleName, gatewayv1beta1.GroupVersion.String(), kind)

							expectedSpec := builders.VirtualServiceSpec().
								Host(testServiceHost).
								Gateway(testGatewayURL).
								HTTP(builders.HTTPRoute().
									Match(builders.MatchRequest().Uri().Regex("/img")).
									Route(builders.RouteDestination().Host(testOathkeeperSvcURL).Port(testOathkeeperPort)).
									Headers(builders.Headers().SetHostHeader(testServiceHost)).
									CorsPolicy(corsPolicyBuilder)).
								HTTP(builders.HTTPRoute().
									Match(builders.MatchRequest().Uri().Regex("/headers")).
									Route(builders.RouteDestination().Host(testOathkeeperSvcURL).Port(testOathkeeperPort)).
									Headers(builders.Headers().SetHostHeader(testServiceHost)).
									CorsPolicy(corsPolicyBuilder))
							gotSpec := *expectedSpec.Get()
							Expect(*vs.Spec.DeepCopy()).To(Equal(*gotSpec.DeepCopy()))

							//Verify Rule1
							expectedRuleMatchURL := fmt.Sprintf("<http|https>://%s<%s>", testServiceHost, "/img")

							rlList := getRuleList(matchingLabels)

							Expect(rlList).To(HaveLen(2))

							rules := make(map[string]rulev1alpha1.Rule)

							for _, rule := range rlList {
								rules[rule.Spec.Match.URL] = rule
							}

							rl := rules[expectedRuleMatchURL]

							//Meta
							verifyOwnerReference(rl.ObjectMeta, apiRuleName, gatewayv1beta1.GroupVersion.String(), kind)

							//Spec.Upstream
							Expect(rl.Spec.Upstream).NotTo(BeNil())
							Expect(rl.Spec.Upstream.URL).To(Equal(fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", testServiceName, testNamespace, testServicePort)))
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
							Expect(asStringSlice(handlerConfig["required_scope"])).To(BeEquivalentTo(testScopes))
							Expect(asStringSlice(handlerConfig["trusted_issuers"])).To(BeEquivalentTo([]string{testIssuer}))
							//Spec.Authorizer
							Expect(rl.Spec.Authorizer).NotTo(BeNil())
							Expect(rl.Spec.Authorizer.Handler).NotTo(BeNil())
							Expect(rl.Spec.Authorizer.Handler.Name).To(Equal("allow"))
							Expect(rl.Spec.Authorizer.Handler.Config).To(BeNil())

							//Spec.Mutators
							Expect(rl.Spec.Mutators).NotTo(BeNil())
							Expect(len(rl.Spec.Mutators)).To(Equal(len(testMutators)))
							Expect(rl.Spec.Mutators[0].Handler.Name).To(Equal(testMutators[0].Name))
							Expect(rl.Spec.Mutators[1].Handler.Name).To(Equal(testMutators[1].Name))

							//Verify Rule2
							expectedRule2MatchURL := fmt.Sprintf("<http|https>://%s<%s>", testServiceHost, "/headers")
							rl2 := rules[expectedRule2MatchURL]

							//Meta
							verifyOwnerReference(rl2.ObjectMeta, apiRuleName, gatewayv1beta1.GroupVersion.String(), "APIRule")

							//Spec.Upstream
							Expect(rl2.Spec.Upstream).NotTo(BeNil())
							Expect(rl2.Spec.Upstream.URL).To(Equal(fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", testServiceName, testNamespace, testServicePort)))
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
							Expect(asStringSlice(handlerConfig["required_scope"])).To(BeEquivalentTo(testScopes))
							Expect(asStringSlice(handlerConfig["trusted_issuers"])).To(BeEquivalentTo([]string{testIssuer}))
							//Spec.Authorizer
							Expect(rl2.Spec.Authorizer).NotTo(BeNil())
							Expect(rl2.Spec.Authorizer.Handler).NotTo(BeNil())
							Expect(rl2.Spec.Authorizer.Handler.Name).To(Equal("allow"))
							Expect(rl2.Spec.Authorizer.Handler.Config).To(BeNil())

							//Spec.Mutators
							Expect(rl2.Spec.Mutators).NotTo(BeNil())
							Expect(len(rl2.Spec.Mutators)).To(Equal(len(testMutators)))
							Expect(rl2.Spec.Mutators[0].Handler.Name).To(Equal(testMutators[0].Name))
							Expect(rl2.Spec.Mutators[1].Handler.Name).To(Equal(testMutators[1].Name))
						})
					})
				})

				Context("with Istio as JWT handler,", func() {
					Context("in a happy-path scenario", func() {
						It("should create a VirtualService, a RequestAuthentication and AuthorizationPolicies", func() {
							cm := testConfigMap("istio")
							err := c.Update(context.TODO(), cm)

							if apierrors.IsInvalid(err) {
								Fail(fmt.Sprintf("failed to update configmap, got an invalid object error: %v", err))
							}
							Expect(err).NotTo(HaveOccurred())

							expectedRequest := reconcile.Request{NamespacedName: types.NamespacedName{Name: cm.Name, Namespace: cm.Namespace}}
							Eventually(requests, timeout).Should(Receive(Equal(expectedRequest)))

							apiRuleName := generateTestName(testNameBase, testIDLength)
							testServiceHost := "httpbin-istio-jwt-happy-base.kyma.local"

							rule1 := testRule("/img", []string{"GET"}, nil, testIstioJWTHandler(testIssuer, testJwksUri))
							rule2 := testRule("/headers", []string{"GET"}, nil, testIstioJWTHandler(testIssuer, testJwksUri))
							instance := testInstance(apiRuleName, testNamespace, testServiceName, testServiceHost, testServicePort, []gatewayv1beta1.Rule{rule1, rule2})

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
							Eventually(requests, timeout).Should(Receive(Equal(expectedRequest)))

							matchingLabels := matchingLabelsFunc(apiRuleName, testNamespace)

							//Verify VirtualService
							vsList := networkingv1beta1.VirtualServiceList{}
							err = c.List(context.TODO(), &vsList, matchingLabels)
							Expect(err).NotTo(HaveOccurred())
							Expect(vsList.Items).To(HaveLen(1))
							vs := vsList.Items[0]

							verifyOwnerReference(vs.ObjectMeta, apiRuleName, gatewayv1beta1.GroupVersion.String(), kind)

							expectedSpec := builders.VirtualServiceSpec().
								Host(testServiceHost).
								Gateway(testGatewayURL).
								HTTP(builders.HTTPRoute().
									Match(builders.MatchRequest().Uri().Regex("/img")).
									Route(builders.RouteDestination().Host(fmt.Sprintf("%s.%s.svc.cluster.local", testServiceName, testNamespace)).Port(testServicePort)).
									Headers(builders.Headers().SetHostHeader(testServiceHost)).
									CorsPolicy(corsPolicyBuilder)).
								HTTP(builders.HTTPRoute().
									Match(builders.MatchRequest().Uri().Regex("/headers")).
									Route(builders.RouteDestination().Host(fmt.Sprintf("%s.%s.svc.cluster.local", testServiceName, testNamespace)).Port(testServicePort)).
									Headers(builders.Headers().SetHostHeader(testServiceHost)).
									CorsPolicy(corsPolicyBuilder))
							gotSpec := *expectedSpec.Get()
							Expect(*vs.Spec.DeepCopy()).To(Equal(*gotSpec.DeepCopy()))

							// Verify RequestAuthentication
							raList := securityv1beta1.RequestAuthenticationList{}
							err = c.List(context.TODO(), &raList, matchingLabels)
							Expect(err).NotTo(HaveOccurred())
							Expect(raList.Items).To(HaveLen(1))
							ra := raList.Items[0]

							verifyOwnerReference(ra.ObjectMeta, apiRuleName, gatewayv1beta1.GroupVersion.String(), kind)

							Expect(ra.Spec.Selector.MatchLabels).To(BeEquivalentTo(map[string]string{"app": testServiceName}))
							Expect(ra.Spec.JwtRules[0].Issuer).To(Equal(testIssuer))
							Expect(ra.Spec.JwtRules[0].JwksUri).To(Equal(testJwksUri))

							// Verify AuthorizationPolicies
							apList := securityv1beta1.AuthorizationPolicyList{}
							err = c.List(context.TODO(), &apList, matchingLabels)
							Expect(err).NotTo(HaveOccurred())
							Expect(apList.Items).To(HaveLen(2))

							hasAuthorizationPolicyWithOperationPath := func(apList []*securityv1beta1.AuthorizationPolicy, operationPath string) {

								getByOperationPath := func(apList []*securityv1beta1.AuthorizationPolicy, path string) (*securityv1beta1.AuthorizationPolicy, error) {
									for _, ap := range apList {
										if ap.Spec.Rules[0].To[0].Operation.Paths[0] == path {
											return ap, nil
										}
									}
									return nil, fmt.Errorf("no authorization policy with operation path %s exists", path)
								}

								ap, err := getByOperationPath(apList, operationPath)
								Expect(err).NotTo(HaveOccurred())

								verifyOwnerReference(ap.ObjectMeta, apiRuleName, gatewayv1beta1.GroupVersion.String(), kind)

								Expect(ap.Spec.Selector.MatchLabels).To(BeEquivalentTo(map[string]string{"app": testServiceName}))
								Expect(ap.Spec.Rules[0].From[0].Source.RequestPrincipals[0]).To(Equal("*"))
								Expect(ap.Spec.Rules[0].To[0].Operation.Paths[0]).To(Equal(operationPath))
								Expect(ap.Spec.Rules[0].To[0].Operation.Methods).To(BeEquivalentTo([]string{"GET"}))
							}

							hasAuthorizationPolicyWithOperationPath(apList.Items, "/img")
							hasAuthorizationPolicyWithOperationPath(apList.Items, "/headers")

						})
					})
				})
			})
		})

		Context("on specified paths", func() {
			Context("with multiple endpoints secured with different authentication methods", func() {
				Context("in the happy path scenario", func() {
					It("should create a VS with corresponding matchers and access rules for each secured path", func() {
						jwtHandler := testOryJWTHandler(testIssuer, testScopes)
						oauthHandler := testOauthHandler(testScopes)
						rule1 := testRule("/img", []string{"GET"}, testMutators, jwtHandler)
						rule2 := testRule("/headers", []string{"GET"}, testMutators, oauthHandler)
						rule3 := testRule("/status", []string{"GET"}, testMutators, noConfigHandler("noop"))
						rule4 := testRule("/favicon", []string{"GET"}, nil, noConfigHandler("allow"))

						apiRuleName := generateTestName(testNameBase, testIDLength)
						testServiceHost := "httpbin4.kyma.local"
						instance := testInstance(apiRuleName, testNamespace, testServiceName, testServiceHost, testServicePort, []gatewayv1beta1.Rule{rule1, rule2, rule3, rule4})

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
						Eventually(requests, timeout).Should(Receive(Equal(expectedRequest)))

						matchingLabels := matchingLabelsFunc(apiRuleName, testNamespace)

						//Verify VirtualService
						vsList := networkingv1beta1.VirtualServiceList{}
						err = c.List(context.TODO(), &vsList, matchingLabels)
						Expect(err).NotTo(HaveOccurred())
						Expect(vsList.Items).To(HaveLen(1))
						vs := vsList.Items[0]

						//Meta
						verifyOwnerReference(vs.ObjectMeta, apiRuleName, gatewayv1beta1.GroupVersion.String(), kind)

						expectedSpec := builders.VirtualServiceSpec().
							Host(testServiceHost).
							Gateway(testGatewayURL).
							HTTP(builders.HTTPRoute().
								Match(builders.MatchRequest().Uri().Regex("/img")).
								Route(builders.RouteDestination().Host(testOathkeeperSvcURL).Port(testOathkeeperPort)).
								Headers(builders.Headers().SetHostHeader(testServiceHost)).
								CorsPolicy(corsPolicyBuilder)).
							HTTP(builders.HTTPRoute().
								Match(builders.MatchRequest().Uri().Regex("/headers")).
								Route(builders.RouteDestination().Host(testOathkeeperSvcURL).Port(testOathkeeperPort)).
								Headers(builders.Headers().SetHostHeader(testServiceHost)).
								CorsPolicy(corsPolicyBuilder)).
							HTTP(builders.HTTPRoute().
								Match(builders.MatchRequest().Uri().Regex("/status")).
								Route(builders.RouteDestination().Host(testOathkeeperSvcURL).Port(testOathkeeperPort)).
								Headers(builders.Headers().SetHostHeader(testServiceHost)).
								CorsPolicy(corsPolicyBuilder)).
							HTTP(builders.HTTPRoute().
								Match(builders.MatchRequest().Uri().Regex("/favicon")).
								Route(builders.RouteDestination().Host("httpbin.atgo-system.svc.cluster.local").Port(443)). // "allow", no oathkeeper rule!
								Headers(builders.Headers().SetHostHeader(testServiceHost)).
								CorsPolicy(corsPolicyBuilder))

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
							expectedRuleMatchURL := fmt.Sprintf("<http|https>://%s</%s>", testServiceHost, tc.path)

							rlList := getRuleList(matchingLabels)
							Expect(rlList).To(HaveLen(3))

							rules := make(map[string]rulev1alpha1.Rule)

							for _, rule := range rlList {
								rules[rule.Spec.Match.URL] = rule
							}

							rl := rules[expectedRuleMatchURL]

							//Meta
							verifyOwnerReference(rl.ObjectMeta, apiRuleName, gatewayv1beta1.GroupVersion.String(), kind)

							//Spec.Upstream
							Expect(rl.Spec.Upstream).NotTo(BeNil())
							Expect(rl.Spec.Upstream.URL).To(Equal(fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", testServiceName, testNamespace, testServicePort)))
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
							Expect(len(rl.Spec.Mutators)).To(Equal(len(testMutators)))
							Expect(rl.Spec.Mutators[0].Handler.Name).To(Equal(testMutators[0].Name))
							Expect(rl.Spec.Mutators[1].Handler.Name).To(Equal(testMutators[1].Name))
						}

						//make sure no rule for "/favicon" path has been created
						name := fmt.Sprintf("%s-%s-3", apiRuleName, testServiceName)
						Expect(c.Get(context.TODO(), client.ObjectKey{Name: name, Namespace: testNamespace}, &rulev1alpha1.Rule{})).To(HaveOccurred())
					})
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

func verifyOwnerReference(m metav1.ObjectMeta, name, version, kind string) {
	Expect(m.OwnerReferences).To(HaveLen(1))
	Expect(m.OwnerReferences[0].APIVersion).To(Equal(version))
	Expect(m.OwnerReferences[0].Kind).To(Equal(kind))
	Expect(m.OwnerReferences[0].Name).To(Equal(name))
	Expect(m.OwnerReferences[0].UID).NotTo(BeEmpty())
	Expect(*m.OwnerReferences[0].Controller).To(BeTrue())
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

	bytes, err := json.Marshal(istioint.JwtConfig{
		Authentications: []istioint.JwtAuth{
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

func getRuleList(matchingLabels client.ListOption) []rulev1alpha1.Rule {
	res := rulev1alpha1.RuleList{}
	err := c.List(context.TODO(), &res, matchingLabels)
	Expect(err).NotTo(HaveOccurred())
	return res.Items
}

func verifyRuleList(ruleList []rulev1alpha1.Rule, pathToURLFunc func(string) string, expected ...gatewayv1beta1.Rule) {

	Expect(ruleList).To(HaveLen(len(expected)))

	actual := make(map[string]rulev1alpha1.Rule)

	for _, rule := range ruleList {
		actual[rule.Spec.Match.URL] = rule
	}

	for i := range expected {
		ruleUrl := pathToURLFunc(expected[i].Path)
		Expect(actual[ruleUrl].Spec.Match.Methods).To(Equal(expected[i].Methods))
		//Expect(actual[ruleUrl].Spec.Authenticators).To(Equal(expected[i].AccessStrategies))
		verifyAccessStrategies(actual[ruleUrl].Spec.Authenticators, expected[i].AccessStrategies)
		//Expect(actual[ruleUrl].Spec.Mutators).To(Equal(expected[i].Mutators))
		verifyMutators(actual[ruleUrl].Spec.Mutators, expected[i].Mutators)
	}
}
func verifyMutators(actual []*rulev1alpha1.Mutator, expected []*gatewayv1beta1.Mutator) {
	if expected == nil {
		Expect(actual).To(BeNil())
	} else {
		for i := 0; i < len(expected); i++ {
			verifyHandler(actual[i].Handler, expected[i].Handler)
		}
	}
}
func verifyAccessStrategies(actual []*rulev1alpha1.Authenticator, expected []*gatewayv1beta1.Authenticator) {
	if expected == nil {
		Expect(actual).To(BeNil())
	} else {
		for i := 0; i < len(expected); i++ {
			verifyHandler(actual[i].Handler, expected[i].Handler)
		}
	}
}

func verifyHandler(actual *rulev1alpha1.Handler, expected *gatewayv1beta1.Handler) {
	if expected == nil {
		Expect(actual).To(BeNil())
	} else {
		Expect(actual.Name).To(Equal(expected.Name))
		Expect(actual.Config).To(Equal(expected.Config))
	}
}

func matchingLabelsFunc(apiRuleName, namespace string) client.ListOption {
	labels := make(map[string]string)
	labels[processing.OwnerLabel] = fmt.Sprintf("%s.%s", apiRuleName, namespace)
	return client.MatchingLabels(labels)
}
