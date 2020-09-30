package controllers_test

import (
	"context"
	"istio.io/api/networking/v1beta1"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"path/filepath"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	"sync"
	"testing"

	"github.com/kyma-incubator/api-gateway/internal/processing"

	gatewayv1alpha1 "github.com/kyma-incubator/api-gateway/api/v1alpha1"
	"github.com/kyma-incubator/api-gateway/controllers"
	rulev1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	// +kubebuilder:scaffold:imports
)

var (
	cfg        *rest.Config
	k8sClient  client.Client
	testEnv    *envtest.Environment
	stopMgr    chan struct{}
	mgrStopped *sync.WaitGroup
	requests   chan reconcile.Request
	c          client.Client

	TestAllowOrigins = []*v1beta1.StringMatch{{MatchType: &v1beta1.StringMatch_Regex{Regex: ".*"}}}
	TestAllowMethods = []string{"GET", "POST", "PUT", "DELETE"}
	TestAllowHeaders = []string{"header1", "header2"}
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Controller Suite",
		[]Reporter{printer.NewlineReporter{}})
}

var _ = BeforeSuite(func(done Done) {
	logf.SetLogger(zap.LoggerTo(GinkgoWriter, true))

	By("bootstrapping test environment")
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

	err = rulev1alpha1.AddToScheme(s)
	Expect(err).NotTo(HaveOccurred())

	err = gatewayv1alpha1.AddToScheme(s)
	Expect(err).NotTo(HaveOccurred())

	err = networkingv1beta1.AddToScheme(s)
	Expect(err).NotTo(HaveOccurred())

	err = corev1.AddToScheme(s)
	Expect(err).NotTo(HaveOccurred())

	mgr, err := manager.New(cfg, manager.Options{Scheme: s, MetricsBindAddress: "0"})
	Expect(err).NotTo(HaveOccurred())

	c, err = client.New(cfg, client.Options{Scheme: s})
	Expect(err).NotTo(HaveOccurred())

	ns := &corev1.Namespace{
		ObjectMeta: v1.ObjectMeta{Name: testNamespace},
		Spec:       corev1.NamespaceSpec{},
	}
	err = c.Create(context.TODO(), ns)
	Expect(err).NotTo(HaveOccurred())

	reconciler := &controllers.APIReconciler{
		Client:            mgr.GetClient(),
		Log:               ctrl.Log.WithName("controllers").WithName("Api"),
		OathkeeperSvc:     testOathkeeperSvcURL,
		OathkeeperSvcPort: testOathkeeperPort,
		DomainWhiteList:   []string{"bar", "kyma.local"},
		CorsConfig: &processing.CorsConfig{
			AllowOrigins: TestAllowOrigins,
			AllowMethods: TestAllowMethods,
			AllowHeaders: TestAllowHeaders,
		},
		GeneratedObjectsLabels: map[string]string{},
	}

	var recFn reconcile.Reconciler
	recFn, requests = SetupTestReconcile(reconciler)

	Expect(add(mgr, recFn)).To(Succeed())

	stopMgr, mgrStopped = StartTestManager(mgr)

	close(done)
}, 60)

var _ = AfterSuite(func() {
	close(stopMgr)
	mgrStopped.Wait()

	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).ToNot(HaveOccurred())
})

// SetupTestReconcile returns a reconcile.Reconcile implementation that delegates to inner and
// writes the request to requests after Reconcile is finished.
func SetupTestReconcile(inner reconcile.Reconciler) (reconcile.Reconciler, chan reconcile.Request) {
	requests := make(chan reconcile.Request)
	fn := reconcile.Func(func(req reconcile.Request) (reconcile.Result, error) {
		result, err := inner.Reconcile(req)
		requests <- req
		return result, err
	})
	return fn, requests
}

// StartTestManager adds recFn
func StartTestManager(mgr manager.Manager) (chan struct{}, *sync.WaitGroup) {
	stop := make(chan struct{})
	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		Expect(mgr.Start(stop)).NotTo(HaveOccurred())
	}()
	return stop, wg
}
