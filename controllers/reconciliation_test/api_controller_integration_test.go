package reconciliation_test

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"encoding/json"
	"github.com/kyma-incubator/api-gateway/internal/helpers"

	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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
	eventuallyTimeout = time.Second * 10

	testGatewayURL              = "kyma-system/kyma-gateway"
	testOathkeeperSvcURL        = "oathkeeper.kyma-system.svc.cluster.local"
	testOathkeeperPort   uint32 = 1234
	testNamespace               = "atgo-system"
	testNameBase                = "test"
	testIDLength                = 5
)

var _ = Describe("APIRule Controller Reconciliation", func() {
	const testServiceName = "httpbin"
	const testServicePort uint32 = 443
	var testIssuer = "https://oauth2.example.com/"

	BeforeEach(func() {
		// We configure `ory` in ConfigMap as the default for all tests
		cm := testConfigMap(helpers.JWT_HANDLER_ORY)
		err := c.Update(context.TODO(), cm)
		if apierrors.IsInvalid(err) {
			Fail(fmt.Sprintf("failed to update configmap, got an invalid object error: %v", err))
		}
		Expect(err).NotTo(HaveOccurred())
	})

	Context("Handler in ConfigMap was changed to different one", func() {
		It("Should update valid Ory APIRule with error when going from Ory to Istio", func() {
			// given
			apiRuleName := generateTestName(testNameBase, testIDLength)
			testServiceHost := fmt.Sprintf("httpbin-%s.kyma.local", apiRuleName)

			rule := testRule("/img", []string{"GET"}, nil, testOryJWTHandler(testIssuer, []string{"test-scope"}))
			apiRule := testInstance(apiRuleName, testNamespace, testServiceName, testServiceHost, testServicePort, []gatewayv1beta1.Rule{rule})

			By("Create ApiRule with Rule using JWT handler")
			err := c.Create(context.TODO(), apiRule)
			Expect(err).NotTo(HaveOccurred())
			defer func() {
				err := c.Delete(context.TODO(), apiRule)
				Expect(err).NotTo(HaveOccurred())
			}()

			initialStateReq := reconcile.Request{NamespacedName: types.NamespacedName{Name: apiRuleName, Namespace: testNamespace}}
			Eventually(requests, eventuallyTimeout).Should(Receive(Equal(initialStateReq)))

			// when
			By("Setting JWT handler config map to istio")
			cm := testConfigMap("istio")
			err = c.Update(context.TODO(), cm)
			Expect(err).NotTo(HaveOccurred())

			cmRequest := reconcile.Request{NamespacedName: types.NamespacedName{Name: cm.Name, Namespace: cm.Namespace}}
			Eventually(requests, eventuallyTimeout).Should(Receive(Equal(cmRequest)))

			updateApiRuleReq := reconcile.Request{NamespacedName: types.NamespacedName{Name: apiRuleName, Namespace: testNamespace}}
			Eventually(requests, eventuallyTimeout).Should(Receive(Equal(updateApiRuleReq)))

			// then

			expectedApiRule := gatewayv1beta1.APIRule{}
			err = c.Get(context.TODO(), client.ObjectKey{Name: apiRuleName, Namespace: testNamespace}, &expectedApiRule)
			Expect(err).NotTo(HaveOccurred())

			Expect(expectedApiRule.Status.APIRuleStatus.Code).To(Equal(gatewayv1beta1.StatusError))
		})

		It("Should update valid Istio APIRule with error from Istio to Ory", func() {
			// given
			By("Setting JWT handler config map to istio")
			cm := testConfigMap("istio")
			err := c.Update(context.TODO(), cm)
			Expect(err).NotTo(HaveOccurred())

			apiRuleName := generateTestName(testNameBase, testIDLength)
			testServiceHost := fmt.Sprintf("httpbin-%s.kyma.local", apiRuleName)

			rule := testRule("/img", []string{"GET"}, nil, testIstioJWTHandler(testIssuer, "https://example.com/well-known/.jwks"))
			apiRule := testInstance(apiRuleName, testNamespace, testServiceName, testServiceHost, testServicePort, []gatewayv1beta1.Rule{rule})

			By("Create ApiRule with Rule using JWT handler")
			err = c.Create(context.TODO(), apiRule)
			Expect(err).NotTo(HaveOccurred())
			defer func() {
				err := c.Delete(context.TODO(), apiRule)
				Expect(err).NotTo(HaveOccurred())
			}()

			initialStateReq := reconcile.Request{NamespacedName: types.NamespacedName{Name: apiRuleName, Namespace: testNamespace}}
			Eventually(requests, eventuallyTimeout).Should(Receive(Equal(initialStateReq)))

			// when
			By("Setting JWT handler config map to ory")
			cm = testConfigMap("ory")
			err = c.Update(context.TODO(), cm)
			Expect(err).NotTo(HaveOccurred())

			cmRequest := reconcile.Request{NamespacedName: types.NamespacedName{Name: cm.Name, Namespace: cm.Namespace}}
			Eventually(requests, eventuallyTimeout).Should(Receive(Equal(cmRequest)))

			updateApiRuleReq := reconcile.Request{NamespacedName: types.NamespacedName{Name: apiRuleName, Namespace: testNamespace}}
			Eventually(requests, eventuallyTimeout).Should(Receive(Equal(updateApiRuleReq)))

			// then

			expectedApiRule := gatewayv1beta1.APIRule{}
			err = c.Get(context.TODO(), client.ObjectKey{Name: apiRuleName, Namespace: testNamespace}, &expectedApiRule)
			Expect(err).NotTo(HaveOccurred())

			Expect(expectedApiRule.Status.APIRuleStatus.Code).To(Equal(gatewayv1beta1.StatusError))
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

func generateTestName(name string, length int) string {

	rand.Seed(time.Now().UnixNano())

	letterRunes := []rune("abcdefghijklmnopqrstuvwxyz")

	b := make([]rune, length)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return name + "-" + string(b)
}
