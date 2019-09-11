package controllers_test

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"encoding/json"

	gatewayv2alpha1 "github.com/kyma-incubator/api-gateway/api/v2alpha1"
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

	testGatewayURL              = "kyma-gateway.kyma-system.svc.cluster.local"
	testOathkeeperSvcURL        = "oathkeeper.kyma-system.svc.cluster.local"
	testOathkeeperPort   uint32 = 1234
	testNamespace               = "padu-system"
	testNameBase                = "test"
	testIDLength                = 5
)

var _ = Describe("Gate Controller", func() {
	const testServiceName = "httpbin"
	const testServiceHost = "httpbin.kyma.local"
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

	Context("when creating a Gate for exposing service", func() {
		Context("on all the paths,", func() {
			Context("secured with Oauth2 introspection,", func() {
				Context("in a happy-path scenario", func() {
					It("should create a VirtualService and an AccessRule", func() {
						configJSON := fmt.Sprintf(`{}`)

						testName := generateTestName(testNameBase, testIDLength)

						var authStrategyName = gatewayv2alpha1.Oauth

						instance := testInstance(authStrategyName, configJSON, testName, testNamespace, testServiceName, testServiceHost, testServicePort, testPath, testMethods, testScopes, testMutators)

						err := c.Create(context.TODO(), instance)
						if apierrors.IsInvalid(err) {
							Fail(fmt.Sprintf("failed to create object, got an invalid object error: %v", err))
							return
						}
						Expect(err).NotTo(HaveOccurred())
						defer c.Delete(context.TODO(), instance)

						expectedRequest := reconcile.Request{NamespacedName: types.NamespacedName{Name: testName, Namespace: testNamespace}}

						Eventually(requests, timeout).Should(Receive(Equal(expectedRequest)))

						//Verify VirtualService
						expectedVSName := testName + "-" + testServiceName
						expectedVSNamespace := testNamespace
						vs := networkingv1alpha3.VirtualService{}
						err = c.Get(context.TODO(), client.ObjectKey{Name: expectedVSName, Namespace: expectedVSNamespace}, &vs)
						Expect(err).NotTo(HaveOccurred())

						//Meta
						verifyOwnerReference(vs.ObjectMeta, testName, gatewayv2alpha1.GroupVersion.String(), "Gate")
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
						expectedRuleName := testName + "-" + testServiceName
						expectedRuleNamespace := testNamespace
						rl := rulev1alpha1.Rule{}
						err = c.Get(context.TODO(), client.ObjectKey{Name: expectedRuleName, Namespace: expectedRuleNamespace}, &rl)
						Expect(err).NotTo(HaveOccurred())

						//Meta
						verifyOwnerReference(rl.ObjectMeta, testName, gatewayv2alpha1.GroupVersion.String(), "Gate")

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
					It("should create a VirtualService and an AccessRule", func() {
						configJSON := fmt.Sprintf(`
							{
								"issuer": "%s",
								"jwks": []
							}`, testIssuer)
						fmt.Printf("---\n%s\n---", configJSON)
						testName := generateTestName(testNameBase, testIDLength)

						var authStrategyName = gatewayv2alpha1.Jwt
						instance := testInstance(authStrategyName, configJSON, testName, testNamespace, testServiceName, testServiceHost, testServicePort, "/.*", []string{"GET"}, testScopes, testMutators)

						err := c.Create(context.TODO(), instance)
						if apierrors.IsInvalid(err) {
							Fail(fmt.Sprintf("failed to create object, got an invalid object error: %v", err))
							return
						}
						Expect(err).NotTo(HaveOccurred())
						defer c.Delete(context.TODO(), instance)

						expectedRequest := reconcile.Request{NamespacedName: types.NamespacedName{Name: testName, Namespace: testNamespace}}

						Eventually(requests, timeout).Should(Receive(Equal(expectedRequest)))
						//Verify VirtualService
						expectedVSName := testName + "-" + testServiceName
						expectedVSNamespace := testNamespace
						vs := networkingv1alpha3.VirtualService{}
						err = c.Get(context.TODO(), client.ObjectKey{Name: expectedVSName, Namespace: expectedVSNamespace}, &vs)
						Expect(err).NotTo(HaveOccurred())

						//Meta
						verifyOwnerReference(vs.ObjectMeta, testName, gatewayv2alpha1.GroupVersion.String(), "Gate")
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
						expectedRuleName := testName + "-" + testServiceName
						expectedRuleNamespace := testNamespace
						rl := rulev1alpha1.Rule{}
						err = c.Get(context.TODO(), client.ObjectKey{Name: expectedRuleName, Namespace: expectedRuleNamespace}, &rl)
						Expect(err).NotTo(HaveOccurred())

						//Meta
						verifyOwnerReference(rl.ObjectMeta, testName, gatewayv2alpha1.GroupVersion.String(), "Gate")

						//Spec.Upstream
						Expect(rl.Spec.Upstream).NotTo(BeNil())
						Expect(rl.Spec.Upstream.URL).To(Equal(fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", testServiceName, testNamespace, testServicePort)))
						Expect(rl.Spec.Upstream.StripPath).To(BeNil())
						Expect(rl.Spec.Upstream.PreserveHost).To(BeNil())
						//Spec.Match
						Expect(rl.Spec.Match).NotTo(BeNil())
						Expect(rl.Spec.Match.URL).To(Equal(fmt.Sprintf("<http|https>://%s<%s>", testServiceHost, testPath)))
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
						Expect(handlerConfig).To(HaveLen(2))
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

	err = c.Watch(&source.Kind{Type: &gatewayv2alpha1.Gate{}}, &handler.EnqueueRequestForObject{})
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

func testInstance(authStrategyName, configJSON, name, namespace, serviceName, serviceHost string, servicePort uint32, path string, methods []string, scopes []string, mutators []*rulev1alpha1.Mutator) *gatewayv2alpha1.Gate {
	rawCfg := &runtime.RawExtension{
		Raw: []byte(configJSON),
	}

	var gateway = testGatewayURL

	return &gatewayv2alpha1.Gate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: gatewayv2alpha1.GateSpec{
			Gateway: &gateway,
			Service: &gatewayv2alpha1.Service{
				Host: &serviceHost,
				Name: &serviceName,
				Port: &servicePort,
			},
			Auth: &gatewayv2alpha1.AuthStrategy{
				Name:   &authStrategyName,
				Config: rawCfg,
			},
			Paths: []gatewayv2alpha1.Path{
				{
					Path:    path,
					Scopes:  scopes,
					Methods: methods,
				},
			},
			Mutators: mutators,
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
