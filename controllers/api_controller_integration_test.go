package controllers_test

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/kyma-incubator/api-gateway/internal/processing"

	"encoding/json"

	gatewayv1alpha1 "github.com/kyma-incubator/api-gateway/api/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	rulev1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	networkingv1alpha3 "knative.dev/pkg/apis/istio/v1alpha3"
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
	testGatewayURL              = "kyma-gateway.kyma-system.svc.cluster.local"
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
	var testMethods = []string{"GET", "PUT"}
	var testScopes = []string{"foo", "bar"}
	var testMutators = []*rulev1alpha1.Mutator{
		{
			Handler: &rulev1alpha1.Handler{
				Name: "noop",
			},
		},
		{
			Handler: &rulev1alpha1.Handler{
				Name: "idtoken",
			},
		},
	}

	Context("when creating an APIRule for exposing service", func() {

		It("Should report validation errors in CR status", func() {

			configJSON := fmt.Sprintf(`{
							"required_scope": [%s]
						}`, toCSVList(testScopes))

			nonEmptyConfig := &rulev1alpha1.Handler{
				Name: "noop",
				Config: &runtime.RawExtension{
					Raw: []byte(configJSON),
				},
			}

			apiRuleName := generateTestName(testNameBase, testIDLength)
			testServiceHost := "httpbin.kyma.local"
			rule := testRule(testPath, testMethods, testMutators, nonEmptyConfig)
			instance := testInstance(apiRuleName, testNamespace, testServiceName, testServiceHost, testServicePort, []gatewayv1alpha1.Rule{rule})
			instance.Spec.Rules = append(instance.Spec.Rules, instance.Spec.Rules[0]) //Duplicate entry
			instance.Spec.Rules = append(instance.Spec.Rules, instance.Spec.Rules[0]) //Duplicate entry

			err := c.Create(context.TODO(), instance)
			if apierrors.IsInvalid(err) {
				Fail(fmt.Sprintf("failed to create object, got an invalid object error: %v", err))
				return
			}
			Expect(err).NotTo(HaveOccurred())
			defer c.Delete(context.TODO(), instance)

			expectedRequest := reconcile.Request{NamespacedName: types.NamespacedName{Name: apiRuleName, Namespace: testNamespace}}

			Eventually(requests, timeout).Should(Receive(Equal(expectedRequest)))

			//Verify APIRule
			created := gatewayv1alpha1.APIRule{}
			err = c.Get(context.TODO(), client.ObjectKey{Name: apiRuleName, Namespace: testNamespace}, &created)
			Expect(err).NotTo(HaveOccurred())
			Expect(created.Status.APIRuleStatus.Code).To(Equal(gatewayv1alpha1.StatusError))
			Expect(created.Status.APIRuleStatus.Description).To(ContainSubstring("Multiple validation errors:"))
			Expect(created.Status.APIRuleStatus.Description).To(ContainSubstring("Attribute \".spec.rules\": multiple rules defined for the same path"))
			Expect(created.Status.APIRuleStatus.Description).To(ContainSubstring("Attribute \".spec.rules[0].accessStrategies[0].config\": strategy: noop does not support configuration"))
			Expect(created.Status.APIRuleStatus.Description).To(ContainSubstring("Attribute \".spec.rules[1].accessStrategies[0].config\": strategy: noop does not support configuration"))
			Expect(created.Status.APIRuleStatus.Description).To(ContainSubstring("1 more error(s)..."))

			//Verify VirtualService is not created
			vsList := networkingv1alpha3.VirtualServiceList{}
			err = c.List(context.TODO(), &vsList)
			Expect(err).NotTo(HaveOccurred())
			Expect(vsList.Items).To(HaveLen(0))
		})

		Context("on all the paths,", func() {
			Context("secured with Oauth2 introspection,", func() {
				Context("in a happy-path scenario", func() {
					It("should create a VirtualService and an AccessRule", func() {
						configJSON := fmt.Sprintf(`{
							"required_scope": [%s]
						}`, toCSVList(testScopes))

						oauthConfig := &rulev1alpha1.Handler{
							Name: "oauth2_introspection",
							Config: &runtime.RawExtension{
								Raw: []byte(configJSON),
							},
						}

						apiRuleName := generateTestName(testNameBase, testIDLength)
						testServiceHost := "httpbin2.kyma.local"
						rule := testRule(testPath, testMethods, testMutators, oauthConfig)
						instance := testInstance(apiRuleName, testNamespace, testServiceName, testServiceHost, testServicePort, []gatewayv1alpha1.Rule{rule})

						err := c.Create(context.TODO(), instance)
						if apierrors.IsInvalid(err) {
							Fail(fmt.Sprintf("failed to create object, got an invalid object error: %v", err))
							return
						}
						Expect(err).NotTo(HaveOccurred())
						defer c.Delete(context.TODO(), instance)

						expectedRequest := reconcile.Request{NamespacedName: types.NamespacedName{Name: apiRuleName, Namespace: testNamespace}}

						Eventually(requests, timeout).Should(Receive(Equal(expectedRequest)))

						labels := make(map[string]string)
						labels[processing.OwnerLabel] = fmt.Sprintf("%s.%s", apiRuleName, testNamespace)
						matchingLabelsFunc := client.MatchingLabels(labels)

						//Verify VirtualService
						vsList := networkingv1alpha3.VirtualServiceList{}
						err = c.List(context.TODO(), &vsList, matchingLabelsFunc)
						Expect(err).NotTo(HaveOccurred())
						Expect(vsList.Items).To(HaveLen(1))
						vs := vsList.Items[0]

						//Meta
						Expect(vs.Name).To(HavePrefix(apiRuleName + "-"))
						Expect(len(vs.Name) > len(apiRuleName)).To(BeTrue())

						verifyOwnerReference(vs.ObjectMeta, apiRuleName, gatewayv1alpha1.GroupVersion.String(), kind)
						//Spec.Hosts
						Expect(vs.Spec.Hosts).To(HaveLen(1))
						Expect(vs.Spec.Hosts[0]).To(Equal(testServiceHost))
						//Spec.Gateways
						Expect(vs.Spec.Gateways).To(HaveLen(1))
						Expect(vs.Spec.Gateways[0]).To(Equal(testGatewayURL))
						//Spec.HTTP
						Expect(vs.Spec.HTTP).To(HaveLen(1))
						////// HTTP.Match[]
						Expect(vs.Spec.HTTP[0].Match).To(HaveLen(1))
						/////////// Match[].URI
						Expect(vs.Spec.HTTP[0].Match[0].URI).NotTo(BeNil())
						Expect(vs.Spec.HTTP[0].Match[0].URI.Exact).To(BeEmpty())
						Expect(vs.Spec.HTTP[0].Match[0].URI.Prefix).To(BeEmpty())
						Expect(vs.Spec.HTTP[0].Match[0].URI.Suffix).To(BeEmpty())
						Expect(vs.Spec.HTTP[0].Match[0].URI.Regex).To(Equal(testPath))
						Expect(vs.Spec.HTTP[0].Match[0].Scheme).To(BeNil())
						Expect(vs.Spec.HTTP[0].Match[0].Method).To(BeNil())
						Expect(vs.Spec.HTTP[0].Match[0].Authority).To(BeNil())
						Expect(vs.Spec.HTTP[0].Match[0].Headers).To(BeNil())
						Expect(vs.Spec.HTTP[0].Match[0].Port).To(BeZero())
						Expect(vs.Spec.HTTP[0].Match[0].SourceLabels).To(BeNil())
						Expect(vs.Spec.HTTP[0].Match[0].Gateways).To(BeNil())
						////// HTTP.Route[]
						Expect(vs.Spec.HTTP[0].Route).To(HaveLen(1))
						Expect(vs.Spec.HTTP[0].Route[0].Destination.Host).To(Equal(testOathkeeperSvcURL))
						Expect(vs.Spec.HTTP[0].Route[0].Destination.Subset).To(Equal(""))
						Expect(vs.Spec.HTTP[0].Route[0].Destination.Port.Name).To(Equal(""))
						Expect(vs.Spec.HTTP[0].Route[0].Destination.Port.Number).To(Equal(testOathkeeperPort))
						Expect(vs.Spec.HTTP[0].Route[0].Weight).To(BeZero())
						Expect(vs.Spec.HTTP[0].Route[0].Headers).To(BeNil())
						//Others
						Expect(vs.Spec.HTTP[0].Rewrite).To(BeNil())
						Expect(vs.Spec.HTTP[0].WebsocketUpgrade).To(BeFalse())
						Expect(vs.Spec.HTTP[0].Timeout).To(BeEmpty())
						Expect(vs.Spec.HTTP[0].Retries).To(BeNil())
						Expect(vs.Spec.HTTP[0].Fault).To(BeNil())
						Expect(vs.Spec.HTTP[0].Mirror).To(BeNil())
						Expect(vs.Spec.HTTP[0].DeprecatedAppendHeaders).To(BeNil())
						Expect(vs.Spec.HTTP[0].Headers).To(BeNil())
						Expect(vs.Spec.HTTP[0].RemoveResponseHeaders).To(BeNil())
						Expect(vs.Spec.HTTP[0].CorsPolicy).To(BeNil())
						//Spec.TCP
						Expect(vs.Spec.TCP).To(BeNil())
						//Spec.TLS
						Expect(vs.Spec.TLS).To(BeNil())

						//Verify Rule
						expectedRuleMatchURL := fmt.Sprintf("<http|https>://%s<%s>", testServiceHost, testPath)

						rlList := rulev1alpha1.RuleList{}

						err = c.List(context.TODO(), &rlList, matchingLabelsFunc)
						Expect(err).NotTo(HaveOccurred())

						Expect(rlList.Items).To(HaveLen(1))

						rules := make(map[string]rulev1alpha1.Rule)

						for _, rule := range rlList.Items {
							rules[rule.Spec.Match.URL] = rule
						}

						rl := rules[expectedRuleMatchURL]

						//Meta
						Expect(rl.Name).To(HavePrefix(apiRuleName + "-"))
						Expect(len(rl.Name) > len(apiRuleName)).To(BeTrue())

						verifyOwnerReference(rl.ObjectMeta, apiRuleName, gatewayv1alpha1.GroupVersion.String(), kind)

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
				Context("in a happy-path scenario", func() {
					It("should create a VirtualService and an AccessRules", func() {
						configJSON := fmt.Sprintf(`
							{
								"trusted_issuers": ["%s"],
								"jwks": [],
								"required_scope": [%s]
						}`, testIssuer, toCSVList(testScopes))
						jwtConfig := &rulev1alpha1.Handler{
							Name: "jwt",
							Config: &runtime.RawExtension{
								Raw: []byte(configJSON),
							},
						}

						apiRuleName := generateTestName(testNameBase, testIDLength)
						testServiceHost := "httpbin3.kyma.local"
						rule1 := testRule("/img", []string{"GET"}, testMutators, jwtConfig)
						rule2 := testRule("/headers", []string{"GET"}, testMutators, jwtConfig)
						instance := testInstance(apiRuleName, testNamespace, testServiceName, testServiceHost, testServicePort, []gatewayv1alpha1.Rule{rule1, rule2})

						err := c.Create(context.TODO(), instance)
						if apierrors.IsInvalid(err) {
							Fail(fmt.Sprintf("failed to create object, got an invalid object error: %v", err))
							return
						}
						Expect(err).NotTo(HaveOccurred())
						defer c.Delete(context.TODO(), instance)

						expectedRequest := reconcile.Request{NamespacedName: types.NamespacedName{Name: apiRuleName, Namespace: testNamespace}}

						Eventually(requests, timeout).Should(Receive(Equal(expectedRequest)))

						labels := make(map[string]string)
						labels[processing.OwnerLabel] = fmt.Sprintf("%s.%s", apiRuleName, testNamespace)
						matchingLabelsFunc := client.MatchingLabels(labels)

						//Verify VirtualService
						vsList := networkingv1alpha3.VirtualServiceList{}
						err = c.List(context.TODO(), &vsList, matchingLabelsFunc)
						Expect(err).NotTo(HaveOccurred())
						Expect(vsList.Items).To(HaveLen(1))
						vs := vsList.Items[0]

						//Meta
						verifyOwnerReference(vs.ObjectMeta, apiRuleName, gatewayv1alpha1.GroupVersion.String(), kind)
						//Spec.Hosts
						Expect(vs.Spec.Hosts).To(HaveLen(1))
						Expect(vs.Spec.Hosts[0]).To(Equal(testServiceHost))
						//Spec.Gateways
						Expect(vs.Spec.Gateways).To(HaveLen(1))
						Expect(vs.Spec.Gateways[0]).To(Equal(testGatewayURL))
						//Spec.HTTP
						Expect(vs.Spec.HTTP).To(HaveLen(2))
						////// HTTP.Match[]
						Expect(vs.Spec.HTTP[0].Match).To(HaveLen(1))
						/////////// Match[].URI
						Expect(vs.Spec.HTTP[0].Match[0].URI).NotTo(BeNil())
						Expect(vs.Spec.HTTP[0].Match[0].URI.Exact).To(BeEmpty())
						Expect(vs.Spec.HTTP[0].Match[0].URI.Prefix).To(BeEmpty())
						Expect(vs.Spec.HTTP[0].Match[0].URI.Suffix).To(BeEmpty())
						Expect(vs.Spec.HTTP[0].Match[0].URI.Regex).To(Equal("/img"))
						Expect(vs.Spec.HTTP[0].Match[0].Scheme).To(BeNil())
						Expect(vs.Spec.HTTP[0].Match[0].Method).To(BeNil())
						Expect(vs.Spec.HTTP[0].Match[0].Authority).To(BeNil())
						Expect(vs.Spec.HTTP[0].Match[0].Headers).To(BeNil())
						Expect(vs.Spec.HTTP[0].Match[0].Port).To(BeZero())
						Expect(vs.Spec.HTTP[0].Match[0].SourceLabels).To(BeNil())
						Expect(vs.Spec.HTTP[0].Match[0].Gateways).To(BeNil())
						////// HTTP.Route[]
						Expect(vs.Spec.HTTP[0].Route).To(HaveLen(1))
						Expect(vs.Spec.HTTP[0].Route[0].Destination.Host).To(Equal(testOathkeeperSvcURL))
						Expect(vs.Spec.HTTP[0].Route[0].Destination.Subset).To(Equal(""))
						Expect(vs.Spec.HTTP[0].Route[0].Destination.Port.Name).To(Equal(""))
						Expect(vs.Spec.HTTP[0].Route[0].Destination.Port.Number).To(Equal(testOathkeeperPort))
						Expect(vs.Spec.HTTP[0].Route[0].Weight).To(BeZero())
						Expect(vs.Spec.HTTP[0].Route[0].Headers).To(BeNil())
						//Others
						Expect(vs.Spec.HTTP[0].Rewrite).To(BeNil())
						Expect(vs.Spec.HTTP[0].WebsocketUpgrade).To(BeFalse())
						Expect(vs.Spec.HTTP[0].Timeout).To(BeEmpty())
						Expect(vs.Spec.HTTP[0].Retries).To(BeNil())
						Expect(vs.Spec.HTTP[0].Fault).To(BeNil())
						Expect(vs.Spec.HTTP[0].Mirror).To(BeNil())
						Expect(vs.Spec.HTTP[0].DeprecatedAppendHeaders).To(BeNil())
						Expect(vs.Spec.HTTP[0].Headers).To(BeNil())
						Expect(vs.Spec.HTTP[0].RemoveResponseHeaders).To(BeNil())
						Expect(vs.Spec.HTTP[0].CorsPolicy).To(BeNil())
						//Spec.TCP
						Expect(vs.Spec.TCP).To(BeNil())
						//Spec.TLS
						Expect(vs.Spec.TLS).To(BeNil())

						//Verify Rule1
						expectedRuleMatchURL := fmt.Sprintf("<http|https>://%s<%s>", testServiceHost, "/img")

						rlList := rulev1alpha1.RuleList{}

						err = c.List(context.TODO(), &rlList, matchingLabelsFunc)
						Expect(err).NotTo(HaveOccurred())

						Expect(len(rlList.Items)).To(Equal(2))

						rules := make(map[string]rulev1alpha1.Rule)

						for _, rule := range rlList.Items {
							rules[rule.Spec.Match.URL] = rule
						}

						rl := rules[expectedRuleMatchURL]

						//Meta
						verifyOwnerReference(rl.ObjectMeta, apiRuleName, gatewayv1alpha1.GroupVersion.String(), kind)

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
						verifyOwnerReference(rl2.ObjectMeta, apiRuleName, gatewayv1alpha1.GroupVersion.String(), "APIRule")

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
		})

		Context("on specified paths", func() {
			Context("with multiple endpoints secured with different authentication methods", func() {
				Context("in the happy path scenario", func() {
					It("should create a VS with corresponding matchers and access rules for each secured path", func() {

						configJWT := fmt.Sprintf(`
							{
								"trusted_issuers": ["%s"],
								"jwks": [],
								"required_scope": [%s]
						}`, testIssuer, toCSVList(testScopes))

						configOAuth := fmt.Sprintf(`{
							"required_scope": [%s]
						}`, toCSVList(testScopes))

						jwtHandler := &rulev1alpha1.Handler{
							Name: "jwt",
							Config: &runtime.RawExtension{
								Raw: []byte(configJWT),
							},
						}

						oauthHandler := &rulev1alpha1.Handler{
							Name: "oauth2_introspection",
							Config: &runtime.RawExtension{
								Raw: []byte(configOAuth),
							},
						}

						noopHandler := &rulev1alpha1.Handler{
							Name: "noop",
						}

						allowHandler := &rulev1alpha1.Handler{
							Name: "allow",
						}

						rule1 := testRule("/img", []string{"GET"}, testMutators, jwtHandler)
						rule2 := testRule("/headers", []string{"GET"}, testMutators, oauthHandler)
						rule3 := testRule("/status", []string{"GET"}, testMutators, noopHandler)
						rule4 := testRule("/favicon", []string{"GET"}, nil, allowHandler)

						apiRuleName := generateTestName(testNameBase, testIDLength)
						testServiceHost := "httpbin4.kyma.local"
						instance := testInstance(apiRuleName, testNamespace, testServiceName, testServiceHost, testServicePort, []gatewayv1alpha1.Rule{rule1, rule2, rule3, rule4})

						err := c.Create(context.TODO(), instance)
						if apierrors.IsInvalid(err) {
							Fail(fmt.Sprintf("failed to create object, got an invalid object error: %v", err))
							return
						}
						Expect(err).NotTo(HaveOccurred())
						defer c.Delete(context.TODO(), instance)

						expectedRequest := reconcile.Request{NamespacedName: types.NamespacedName{Name: apiRuleName, Namespace: testNamespace}}
						Eventually(requests, timeout).Should(Receive(Equal(expectedRequest)))

						labels := make(map[string]string)
						labels[processing.OwnerLabel] = fmt.Sprintf("%s.%s", apiRuleName, testNamespace)
						matchingLabelsFunc := client.MatchingLabels(labels)

						//Verify VirtualService
						vsList := networkingv1alpha3.VirtualServiceList{}
						err = c.List(context.TODO(), &vsList, matchingLabelsFunc)
						Expect(err).NotTo(HaveOccurred())
						Expect(vsList.Items).To(HaveLen(1))
						vs := vsList.Items[0]

						//Meta
						verifyOwnerReference(vs.ObjectMeta, apiRuleName, gatewayv1alpha1.GroupVersion.String(), kind)
						//Spec.Hosts
						Expect(vs.Spec.Hosts).To(HaveLen(1))
						Expect(vs.Spec.Hosts[0]).To(Equal(testServiceHost))
						//Spec.Gateways
						Expect(vs.Spec.Gateways).To(HaveLen(1))
						Expect(vs.Spec.Gateways[0]).To(Equal(testGatewayURL))
						//Spec.HTTP
						Expect(vs.Spec.HTTP).To(HaveLen(4))
						//HTTP.Match[]
						Expect(vs.Spec.HTTP[0].Match).To(HaveLen(1))

						Expect(vs.Spec.HTTP[0].Match[0].URI.Regex).To(Equal(rule1.Path))
						Expect(vs.Spec.HTTP[1].Match[0].URI.Regex).To(Equal(rule2.Path))
						Expect(vs.Spec.HTTP[2].Match[0].URI.Regex).To(Equal(rule3.Path))
						Expect(vs.Spec.HTTP[3].Match[0].URI.Regex).To(Equal(rule4.Path))

						for _, h := range vs.Spec.HTTP {

							//Match[].URI
							Expect(h.Match[0].URI).NotTo(BeNil())
							Expect(h.Match[0].URI.Exact).To(BeEmpty())
							Expect(h.Match[0].URI.Prefix).To(BeEmpty())
							Expect(h.Match[0].URI.Suffix).To(BeEmpty())
							Expect(h.Match[0].Scheme).To(BeNil())
							Expect(h.Match[0].Method).To(BeNil())
							Expect(h.Match[0].Authority).To(BeNil())
							Expect(h.Match[0].Headers).To(BeNil())
							Expect(h.Match[0].Port).To(BeZero())
							Expect(h.Match[0].SourceLabels).To(BeNil())
							Expect(h.Match[0].Gateways).To(BeNil())

							//HTTP.Route[]
							Expect(h.Route).To(HaveLen(1))

							url, port := testOathkeeperSvcURL, testOathkeeperPort
							if h.Match[0].URI.Regex == "/favicon" { // allow, no oathkeeper rule
								url, port = "httpbin.atgo-system.svc.cluster.local", 443
							}
							Expect(h.Route[0].Destination.Host).To(Equal(url))
							Expect(h.Route[0].Destination.Subset).To(Equal(""))
							Expect(h.Route[0].Destination.Port.Name).To(Equal(""))
							Expect(h.Route[0].Destination.Port.Number).To(Equal(port))
							Expect(h.Route[0].Weight).To(BeZero())
							Expect(h.Route[0].Headers).To(BeNil())

							//Others
							Expect(h.Rewrite).To(BeNil())
							Expect(h.WebsocketUpgrade).To(BeFalse())
							Expect(h.Timeout).To(BeEmpty())
							Expect(h.Retries).To(BeNil())
							Expect(h.Fault).To(BeNil())
							Expect(h.Mirror).To(BeNil())
							Expect(h.DeprecatedAppendHeaders).To(BeNil())
							Expect(h.Headers).To(BeNil())
							Expect(h.RemoveResponseHeaders).To(BeNil())
							Expect(h.CorsPolicy).To(BeNil())
						}

						//Spec.TCP
						Expect(vs.Spec.TCP).To(BeNil())
						//Spec.TLS
						Expect(vs.Spec.TLS).To(BeNil())

						//Verify Rules
						for _, tc := range []struct {
							path    string
							handler string
							config  []byte
						}{
							{path: "img", handler: "jwt", config: []byte(configJWT)},
							{path: "headers", handler: "oauth2_introspection", config: []byte(configOAuth)},
							{path: "status", handler: "noop", config: nil},
						} {
							expectedRuleMatchURL := fmt.Sprintf("<http|https>://%s</%s>", testServiceHost, tc.path)

							rlList := rulev1alpha1.RuleList{}

							err = c.List(context.TODO(), &rlList, matchingLabelsFunc)
							Expect(err).NotTo(HaveOccurred())

							Expect(len(rlList.Items)).To(Equal(3))

							rules := make(map[string]rulev1alpha1.Rule)

							for _, rule := range rlList.Items {
								rules[rule.Spec.Match.URL] = rule
							}

							rl := rules[expectedRuleMatchURL]

							//Meta
							verifyOwnerReference(rl.ObjectMeta, apiRuleName, gatewayv1alpha1.GroupVersion.String(), kind)

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

	err = c.Watch(&source.Kind{Type: &gatewayv1alpha1.APIRule{}}, &handler.EnqueueRequestForObject{})
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

func testRule(path string, methods []string, mutators []*rulev1alpha1.Mutator, config *rulev1alpha1.Handler) gatewayv1alpha1.Rule {
	return gatewayv1alpha1.Rule{
		Path:     path,
		Methods:  methods,
		Mutators: mutators,
		AccessStrategies: []*rulev1alpha1.Authenticator{
			{
				Handler: config,
			},
		},
	}
}

func testInstance(name, namespace, serviceName, serviceHost string, servicePort uint32, rules []gatewayv1alpha1.Rule) *gatewayv1alpha1.APIRule {
	var gateway = testGatewayURL

	return &gatewayv1alpha1.APIRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: gatewayv1alpha1.APIRuleSpec{
			Gateway: &gateway,
			Service: &gatewayv1alpha1.Service{
				Host: &serviceHost,
				Name: &serviceName,
				Port: &servicePort,
			},
			Rules: rules,
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

//Converts a []interface{} to a string slice. Panics if given object is of other type.
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
