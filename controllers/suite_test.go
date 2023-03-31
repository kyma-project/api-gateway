package controllers_test

import (
	"context"
	"fmt"
	"github.com/kyma-project/api-gateway/internal/builders"
	"os"
	"path/filepath"
	"testing"
	"time"

	rulev1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
	"istio.io/api/networking/v1beta1"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"

	gatewayv1beta1 "github.com/kyma-project/api-gateway/api/v1beta1"
	"github.com/kyma-project/api-gateway/controllers"
	"github.com/kyma-project/api-gateway/internal/helpers"
	"github.com/kyma-project/api-gateway/internal/processing"

	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/reporters"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"

	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	// +kubebuilder:scaffold:imports
)

const (
	eventuallyTimeout           = time.Second * 5
	testNamespace               = "atgo-system"
	testGatewayURL              = "kyma-system/kyma-gateway"
	testOathkeeperSvcURL        = "oathkeeper.kyma-system.svc.cluster.local"
	testOathkeeperPort   uint32 = 1234
)

var (
	cfg       *rest.Config
	k8sClient client.Client
	testEnv   *envtest.Environment
	c         client.Client
	ctx       context.Context
	cancel    context.CancelFunc

	defaultMethods  = []string{"GET", "PUT"}
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
	TestAllowMethods = []string{"GET", "POST", "PUT", "DELETE"}
	TestAllowHeaders = []string{"header1", "header2"}

	defaultCorsPolicy = builders.CorsPolicy().
				AllowHeaders(TestAllowHeaders...).
				AllowMethods(TestAllowMethods...).
				AllowOrigins(TestAllowOrigins...)
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Controller Suite")
}

var _ = BeforeSuite(func(specCtx SpecContext) {
	logf.SetLogger(zap.New(zap.UseDevMode(true), zap.WriteTo(GinkgoWriter)))
	ctx, cancel = context.WithCancel(context.TODO())

	By("Bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{filepath.Join("..", "config", "crd", "bases"), filepath.Join("..", "hack")},
	}

	var err error
	cfg, err = testEnv.Start()
	Expect(err).ToNot(HaveOccurred())
	Expect(cfg).ToNot(BeNil())

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).ToNot(HaveOccurred())
	Expect(k8sClient).ToNot(BeNil())

	s := runtime.NewScheme()

	Expect(gatewayv1beta1.AddToScheme(s)).Should(Succeed())
	Expect(rulev1alpha1.AddToScheme(s)).Should(Succeed())
	Expect(networkingv1beta1.AddToScheme(s)).Should(Succeed())
	Expect(securityv1beta1.AddToScheme(s)).Should(Succeed())
	Expect(corev1.AddToScheme(s)).Should(Succeed())

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:             s,
		MetricsBindAddress: "0",
	})
	Expect(err).NotTo(HaveOccurred())

	c, err = client.New(cfg, client.Options{Scheme: s})
	Expect(err).NotTo(HaveOccurred())

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: testNamespace},
		Spec:       corev1.NamespaceSpec{},
	}
	Expect(c.Create(context.TODO(), ns)).Should(Succeed())

	nsKyma := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: helpers.CM_NS},
		Spec:       corev1.NamespaceSpec{},
	}
	Expect(c.Create(context.TODO(), nsKyma)).Should(Succeed())

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      helpers.CM_NAME,
			Namespace: helpers.CM_NS,
		},
		Data: map[string]string{
			helpers.CM_KEY: fmt.Sprintf("jwtHandler: %s", helpers.JWT_HANDLER_ORY),
		},
	}
	Expect(c.Create(context.TODO(), cm)).Should(Succeed())

	apiReconciler := &controllers.APIRuleReconciler{
		Client:            mgr.GetClient(),
		Scheme:            mgr.GetScheme(),
		Log:               ctrl.Log.WithName("controllers").WithName("Api"),
		OathkeeperSvc:     testOathkeeperSvcURL,
		OathkeeperSvcPort: testOathkeeperPort,
		DomainAllowList:   []string{"bar", "kyma.local"},
		CorsConfig: &processing.CorsConfig{
			AllowOrigins: TestAllowOrigins,
			AllowMethods: TestAllowMethods,
			AllowHeaders: TestAllowHeaders,
		},
		GeneratedObjectsLabels: map[string]string{},
		Config:                 &helpers.Config{},

		ReconcilePeriod:        time.Second * 2,
		OnErrorReconcilePeriod: time.Second * 2,
	}

	Expect(apiReconciler.SetupWithManager(mgr)).Should(Succeed())

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
		reportsFilename := fmt.Sprintf("%s/%s", key, "junit-controller.xml")
		logger.Info("Generating reports at", "location", reportsFilename)
		err := reporters.GenerateJUnitReport(report, reportsFilename)

		if err != nil {
			logger.Error(err, "Junit Report Generation Error")
		}
	} else {
		if err := os.MkdirAll("../reports", 0755); err != nil {
			logger.Error(err, "could not create directory")
		}

		reportsFilename := fmt.Sprintf("%s/%s", "../reports", "junit-controller.xml")
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
	g.Expect(c.List(context.TODO(), &list, matchingLabels)).Should(Succeed())
	g.Expect(list.Items).To(HaveLen(len))
}

// shouldHaveRequestAuthentications verifies that the expected number of request authentications exists for the APIRule
func shouldHaveRequestAuthentications(g Gomega, apiRuleName, testNamespace string, len int) {
	matchingLabels := matchingLabelsFunc(apiRuleName, testNamespace)
	list := securityv1beta1.RequestAuthenticationList{}
	g.Expect(c.List(context.TODO(), &list, matchingLabels)).Should(Succeed())
	g.Expect(list.Items).To(HaveLen(len))
}

// shouldHaveAuthorizationPolicies verifies that the expected number of authorization policies exists for the APIRule
func shouldHaveAuthorizationPolicies(g Gomega, apiRuleName, testNamespace string, len int) {
	matchingLabels := matchingLabelsFunc(apiRuleName, testNamespace)
	list := securityv1beta1.AuthorizationPolicyList{}
	g.Expect(c.List(context.TODO(), &list, matchingLabels)).Should(Succeed())
	g.Expect(list.Items).To(HaveLen(len))
}

// shouldHaveRules verifies that the expected number of rules exists for the APIRule
func shouldHaveRules(g Gomega, apiRuleName, testNamespace string, len int) {
	matchingLabels := matchingLabelsFunc(apiRuleName, testNamespace)
	list := rulev1alpha1.RuleList{}
	g.Expect(c.List(context.TODO(), &list, matchingLabels)).Should(Succeed())
	g.Expect(list.Items).To(HaveLen(len))
}
