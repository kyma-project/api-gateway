package gateway_test

import (
	"context"
	"fmt"
	"github.com/kyma-project/api-gateway/internal/metrics"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	"github.com/kyma-project/api-gateway/controllers"
	"github.com/kyma-project/api-gateway/controllers/gateway"
	"github.com/kyma-project/api-gateway/internal/builders"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	rulev1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
	"istio.io/api/networking/v1beta1"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/kyma-project/api-gateway/internal/helpers"
	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/reporters"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"

	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	// +kubebuilder:scaffold:imports
)

const (
	eventuallyTimeout    = time.Second * 5
	testNamespace        = "atgo-system"
	testGatewayURL       = "kyma-system/kyma-gateway"
	testOathkeeperSvcURL = "oathkeeper.kyma-system.svc.cluster.local"
	testOathkeeperPort   = 1234
)

var (
	cfg     *rest.Config
	testEnv *envtest.Environment
	c       client.Client
	ctx     context.Context
	cancel  context.CancelFunc

	defaultMethods  = []gatewayv1beta1.HttpMethod{http.MethodGet, http.MethodPut}
	defaultScopes   = []string{"foo", "bar"}
	defaultMutators = []*gatewayv1beta1.Mutator{
		{
			Handler: noConfigHandler("noop"),
		},
		{
			Handler: noConfigHandler("idToken"),
		},
	}

	TestAllowOrigins = []*v1beta1.StringMatch{{MatchType: &v1beta1.StringMatch_Regex{Regex: ".*"}}}
	TestAllowMethods = []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete}
	TestAllowHeaders = []string{"header1", "header2"}

	defaultCorsPolicy = builders.CorsPolicy().
		AllowHeaders(TestAllowHeaders...).
		AllowMethods(TestAllowMethods...).
		AllowOrigins(TestAllowOrigins...)
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "API Rule Controller Suite")
}

var _ = BeforeSuite(func(specCtx SpecContext) {
	logf.SetLogger(zap.New(zap.UseDevMode(true), zap.WriteTo(GinkgoWriter)))
	ctx, cancel = context.WithCancel(context.Background())

	By("Bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.FromSlash("../../config/crd/bases"),
			filepath.FromSlash("../../hack/crds"),
		},
	}

	var err error
	cfg, err = testEnv.Start()
	Expect(err).ToNot(HaveOccurred())
	Expect(cfg).ToNot(BeNil())

	s := runtime.NewScheme()

	Expect(gatewayv1beta1.AddToScheme(s)).Should(Succeed())
	Expect(gatewayv2alpha1.AddToScheme(s)).Should(Succeed())
	Expect(rulev1alpha1.AddToScheme(s)).Should(Succeed())
	Expect(networkingv1beta1.AddToScheme(s)).Should(Succeed())
	Expect(securityv1beta1.AddToScheme(s)).Should(Succeed())
	Expect(corev1.AddToScheme(s)).Should(Succeed())
	Expect(apiextensionsv1.AddToScheme(s)).Should(Succeed())

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme: s,
		Metrics: metricsserver.Options{
			BindAddress: "0",
		},
	})
	Expect(err).NotTo(HaveOccurred())

	c, err = client.New(cfg, client.Options{Scheme: s})
	Expect(err).NotTo(HaveOccurred())

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: testNamespace},
		Spec:       corev1.NamespaceSpec{},
	}
	Expect(c.Create(context.Background(), ns)).Should(Succeed())

	nsKyma := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: helpers.CM_NS},
		Spec:       corev1.NamespaceSpec{},
	}
	Expect(c.Create(context.Background(), nsKyma)).Should(Succeed())

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      helpers.CM_NAME,
			Namespace: helpers.CM_NS,
		},
		Data: map[string]string{
			helpers.CM_KEY: fmt.Sprintf("jwtHandler: %s", helpers.JWT_HANDLER_ORY),
		},
	}
	Expect(c.Create(context.Background(), cm)).Should(Succeed())

	reconcilerConfig := gateway.ApiRuleReconcilerConfiguration{
		OathkeeperSvcAddr:         testOathkeeperSvcURL,
		OathkeeperSvcPort:         testOathkeeperPort,
		CorsAllowOrigins:          "regex:.*",
		CorsAllowMethods:          "GET,POST,PUT,DELETE",
		CorsAllowHeaders:          "header1,header2",
		ReconciliationPeriod:      2,
		ErrorReconciliationPeriod: 2,
	}

	apiGatewayMetrics := metrics.NewApiGatewayMetrics()

	apiReconciler := gateway.NewApiRuleReconciler(mgr, reconcilerConfig, apiGatewayMetrics)
	rateLimiterCfg := controllers.RateLimiterConfig{
		Burst:            200,
		Frequency:        30,
		FailureBaseDelay: 1 * time.Second,
		FailureMaxDelay:  10 * time.Second,
	}

	Expect(apiReconciler.SetupWithManager(mgr, rateLimiterCfg)).Should(Succeed())

	go func() {
		defer GinkgoRecover()
		Expect(mgr.Start(ctx)).Should(Succeed())
	}()

}, NodeTimeout(60*time.Second))

var _ = AfterSuite(func() {
	/*
		 Provided solution for timeout issue waiting for kubeapiserver
			https://github.com/kubernetes-sigs/controller-runtime/issues/1571#issuecomment-1005575071
	*/
	cancel()
	By("Tearing down the test environment")
	err := testEnv.Stop()
	// Set 4 with random
	if err != nil {
		time.Sleep(4 * time.Second)
	}
	err = testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())

})

var _ = ReportAfterSuite("custom reporter", func(report types.Report) {
	logger := zap.New(zap.UseDevMode(true), zap.WriteTo(GinkgoWriter))

	if key, ok := os.LookupEnv("ARTIFACTS"); ok {
		reportsFilename := fmt.Sprintf("%s/%s", key, "junit-api-rule-controller.xml")
		logger.Info("Generating reports at", "location", reportsFilename)
		err := reporters.GenerateJUnitReport(report, reportsFilename)

		if err != nil {
			logger.Error(err, "Junit Report Generation Error")
		}
	} else {
		if err := os.MkdirAll("../reports", 0755); err != nil {
			logger.Error(err, "could not create directory")
		}

		reportsFilename := "../../reports/junit-api-rule-controller.xml"
		logger.Info("Generating reports at", "location", reportsFilename)
		err := reporters.GenerateJUnitReport(report, reportsFilename)

		if err != nil {
			logger.Error(err, "Junit Report Generation Error")
		}
	}
})

// shouldHaveVirtualServices verifies that the expected number of virtual services exists for the APIRule
func shouldHaveVirtualServices(g Gomega, apiRuleName, testNamespace string, len int) {
	matchingLabels := matchingLabelsFunc(apiRuleName, testNamespace)
	list := securityv1beta1.RequestAuthenticationList{}
	g.Expect(c.List(context.Background(), &list, matchingLabels)).Should(Succeed())
	g.Expect(list.Items).To(HaveLen(len))
}

// shouldHaveRequestAuthentications verifies that the expected number of request authentications exists for the APIRule
func shouldHaveRequestAuthentications(g Gomega, apiRuleName, testNamespace string, len int) {
	matchingLabels := matchingLabelsFunc(apiRuleName, testNamespace)
	list := securityv1beta1.RequestAuthenticationList{}
	g.Expect(c.List(context.Background(), &list, matchingLabels)).Should(Succeed())
	g.Expect(list.Items).To(HaveLen(len))
}

// shouldHaveAuthorizationPolicies verifies that the expected number of authorization policies exists for the APIRule
func shouldHaveAuthorizationPolicies(g Gomega, apiRuleName, testNamespace string, len int) {
	matchingLabels := matchingLabelsFunc(apiRuleName, testNamespace)
	list := securityv1beta1.AuthorizationPolicyList{}
	g.Expect(c.List(context.Background(), &list, matchingLabels)).Should(Succeed())
	g.Expect(list.Items).To(HaveLen(len))
}

// shouldHaveRules verifies that the expected number of rules exists for the APIRule
func shouldHaveRules(g Gomega, apiRuleName, testNamespace string, len int) {
	matchingLabels := matchingLabelsFunc(apiRuleName, testNamespace)
	list := rulev1alpha1.RuleList{}
	g.Expect(c.List(context.Background(), &list, matchingLabels)).Should(Succeed())
	g.Expect(list.Items).To(HaveLen(len))
}
