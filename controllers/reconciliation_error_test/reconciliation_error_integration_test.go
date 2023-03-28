package reconciliation_test

import (
	"context"
	"fmt"
	"github.com/kyma-project/api-gateway/internal/processing"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"math/rand"
	"strings"
	"time"

	"encoding/json"
	"github.com/kyma-project/api-gateway/internal/helpers"

	gatewayv1beta1 "github.com/kyma-project/api-gateway/api/v1beta1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	eventuallyTimeout = time.Second * 10

	testGatewayURL              = "kyma-system/kyma-gateway"
	testOathkeeperSvcURL        = "oathkeeper.kyma-system.svc.cluster.local"
	testOathkeeperPort   uint32 = 1234
	testNamespace               = "atgo-system"
)

// We use a separate package for the tests, since the manager is different configured. The manager used for these tests
// has a much shorter reconcile and reconcile on error period to be able to reproduce the expected behaviour.
var _ = Describe("APIRule Reconciliation", func() {
	const (
		testNameBase               = "reconcile-test"
		testIDLength               = 5
		testServiceNameBase        = "httpbin"
		testServicePort     uint32 = 443
		testIssuer                 = "https://oauth2.example.com/"
		testJwksUri                = "https://oauth2.example.com/.well-known/jwks.json"
	)

	Context("Handler in ConfigMap was changed to different one", func() {
		It("Should update valid Ory APIRule with error when going from Ory to Istio", func() {
			// given
			updateJwtHandlerTo(helpers.JWT_HANDLER_ORY)

			apiRuleName := generateTestName(testNameBase, testIDLength)
			serviceName := generateTestName(testServiceNameBase, testIDLength)
			testServiceHost := fmt.Sprintf("httpbin-%s.kyma.local", apiRuleName)

			rule := testRule("/img", []string{"GET"}, nil, testOryJWTHandler(testIssuer))
			apiRule := testInstance(apiRuleName, testNamespace, serviceName, testServiceHost, testServicePort, []gatewayv1beta1.Rule{rule})

			By("Create ApiRule with Rule using Ory JWT handler")
			Expect(c.Create(context.TODO(), apiRule)).Should(Succeed())
			defer func() {
				deleteApiRule(apiRule)
			}()
			expectApiRuleStatus(apiRuleName, gatewayv1beta1.StatusOK)

			// when
			updateJwtHandlerTo(helpers.JWT_HANDLER_ISTIO)

			// then
			expectApiRuleStatus(apiRuleName, gatewayv1beta1.StatusError)
		})

		It("Should update valid Istio APIRule with error from Istio to Ory", func() {
			// given
			updateJwtHandlerTo(helpers.JWT_HANDLER_ISTIO)

			apiRuleName := generateTestName(testNameBase, testIDLength)
			serviceName := generateTestName(testServiceNameBase, testIDLength)
			testServiceHost := fmt.Sprintf("httpbin-%s.kyma.local", apiRuleName)

			rule := testRule("/img", []string{"GET"}, nil, testIstioJWTHandler(testIssuer, "https://example.com/well-known/.jwks"))
			apiRule := testInstance(apiRuleName, testNamespace, serviceName, testServiceHost, testServicePort, []gatewayv1beta1.Rule{rule})

			By("Create ApiRule with Rule using Istio JWT handler")
			Expect(c.Create(context.TODO(), apiRule)).Should(Succeed())
			defer func() {
				deleteApiRule(apiRule)
			}()
			expectApiRuleStatus(apiRuleName, gatewayv1beta1.StatusOK)

			// when
			updateJwtHandlerTo(helpers.JWT_HANDLER_ORY)

			// then
			expectApiRuleStatus(apiRuleName, gatewayv1beta1.StatusError)
		})
	})

	It("APIRule in status Error should reconcile to status OK when root cause of error is fixed", func() {
		// given
		updateJwtHandlerTo(helpers.JWT_HANDLER_ISTIO)

		apiRuleName := generateTestName(testNameBase, testIDLength)
		serviceName := generateTestName(testServiceNameBase, testIDLength)
		serviceHost := fmt.Sprintf("%s.kyma.local", serviceName)
		vsName := generateTestName("duplicated-host-vs", testIDLength)

		By(fmt.Sprintf("Create Virtual service for host %s", serviceHost))
		vs := testVirtualService(vsName, serviceHost)
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

		By("Verify VirtualService is created")
		Eventually(func(g Gomega) {
			createdVs := networkingv1beta1.VirtualService{}
			g.Expect(c.Get(context.TODO(), client.ObjectKey{Name: vsName, Namespace: testNamespace}, &createdVs)).Should(Succeed())
		}, eventuallyTimeout).Should(Succeed())

		apiRuleLabelMatcher := matchingLabelsFunc(apiRuleName, testNamespace)

		By("Create APIRule")
		rule := testRule("/headers", []string{"GET"}, nil, testIstioJWTHandler(testIssuer, testJwksUri))
		instance := testInstance(apiRuleName, testNamespace, serviceName, serviceHost, testServicePort, []gatewayv1beta1.Rule{rule})

		Expect(c.Create(context.TODO(), instance)).Should(Succeed())
		defer func() {
			deleteApiRule(instance)
		}()

		expectApiRuleStatus(apiRuleName, gatewayv1beta1.StatusError)

		By("Verify VirtualService for APIRule is not created")
		Eventually(func(g Gomega) {
			vsList := networkingv1beta1.VirtualServiceList{}
			g.Expect(c.List(context.TODO(), &vsList, apiRuleLabelMatcher)).Should(Succeed())
			g.Expect(vsList.Items).To(HaveLen(0))
		}, eventuallyTimeout).Should(Succeed())

		By("Deleting existing VirtualService with duplicated host configuration")
		deleteVirtualService(vs)

		By("Waiting until APIRule is reconciled after error")
		expectApiRuleStatus(apiRuleName, gatewayv1beta1.StatusOK)

		By("Verify VirtualService for APIRule is created")
		Eventually(func(g Gomega) {
			vsList := networkingv1beta1.VirtualServiceList{}
			g.Expect(c.List(context.TODO(), &vsList, apiRuleLabelMatcher)).Should(Succeed())
			g.Expect(vsList.Items).To(HaveLen(1))
		}, eventuallyTimeout).Should(Succeed())
	})
})

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

func testOryJWTHandler(issuer string) *gatewayv1beta1.Handler {

	configJSON := fmt.Sprintf(`{
			"trusted_issuers": ["%s"],
			"jwks": []
		}`, issuer)

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

	rand.NewSource(time.Now().UnixNano())

	letterRunes := []rune("abcdefghijklmnopqrstuvwxyz")

	b := make([]rune, length)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return name + "-" + string(b)
}

func testVirtualService(name string, host string) *networkingv1beta1.VirtualService {
	vs := &networkingv1beta1.VirtualService{}
	vs.ObjectMeta = metav1.ObjectMeta{
		Name:      name,
		Namespace: testNamespace,
	}
	vs.Spec.Hosts = []string{host}

	return vs
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

		By("Waiting until CM is updated")
		Eventually(func(g Gomega) {
			g.Expect(c.Get(context.TODO(), client.ObjectKey{Name: cm.Name, Namespace: cm.Namespace}, cm)).Should(Succeed())
			g.Expect(cm.Data).To(HaveKeyWithValue(helpers.CM_KEY, fmt.Sprintf("jwtHandler: %s", jwtHandler)))
		}, eventuallyTimeout).Should(Succeed())
	}

}

func matchingLabelsFunc(apiRuleName, namespace string) client.ListOption {
	labels := make(map[string]string)
	labels[processing.OwnerLabel] = fmt.Sprintf("%s.%s", apiRuleName, namespace)
	return client.MatchingLabels(labels)
}

func deleteApiRule(apiRule *gatewayv1beta1.APIRule) {
	By(fmt.Sprintf("Deleting ApiRule %s as part of teardown", apiRule.Name))
	Expect(c.Delete(context.TODO(), apiRule)).Should(Succeed())
	Eventually(func(g Gomega) {
		a := gatewayv1beta1.APIRule{}
		err := c.Get(context.TODO(), client.ObjectKey{Name: apiRule.Name, Namespace: testNamespace}, &a)
		g.Expect(errors.IsNotFound(err)).To(BeTrue())
	}, eventuallyTimeout).Should(Succeed())
}

func deleteVirtualService(vs *networkingv1beta1.VirtualService) {
	By(fmt.Sprintf("Deleting VirtualService %s", vs.Name))
	Expect(c.Delete(context.TODO(), vs)).Should(Succeed())
	Eventually(func(g Gomega) {
		v := networkingv1beta1.VirtualService{}
		err := c.Get(context.TODO(), client.ObjectKey{Name: vs.Name, Namespace: testNamespace}, &v)
		g.Expect(errors.IsNotFound(err)).To(BeTrue())
	}, eventuallyTimeout).Should(Succeed())
}

func expectApiRuleStatus(apiRuleName string, statusCode gatewayv1beta1.StatusCode) {
	By(fmt.Sprintf("Expecting ApiRule %s to have status %s", apiRuleName, statusCode))
	Eventually(func(g Gomega) {
		expectedApiRule := gatewayv1beta1.APIRule{}
		g.Expect(c.Get(context.TODO(), client.ObjectKey{Name: apiRuleName, Namespace: testNamespace}, &expectedApiRule)).Should(Succeed())
		g.Expect(expectedApiRule.Status.APIRuleStatus).NotTo(BeNil())
		g.Expect(expectedApiRule.Status.APIRuleStatus.Code).To(Equal(statusCode))
	}, eventuallyTimeout).Should(Succeed())
	By(fmt.Sprintf("Validated that ApiRule %s has status %s", apiRuleName, statusCode))
}
