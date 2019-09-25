package controllers_test

import (
	"path/filepath"
	"sync"
	"testing"

	gatewayv1alpha1 "github.com/kyma-incubator/api-gateway/api/v1alpha1"
	"github.com/kyma-incubator/api-gateway/controllers"
	"github.com/kyma-incubator/api-gateway/internal/validation"
	rulev1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	networkingv1alpha3 "knative.dev/pkg/apis/istio/v1alpha3"
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

var cfg *rest.Config
var k8sClient client.Client
var testEnv *envtest.Environment
var stopMgr chan struct{}
var mgrStopped *sync.WaitGroup
var requests chan reconcile.Request
var c client.Client

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Controller Suite",
		[]Reporter{envtest.NewlineReporter{}})
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

	err = networkingv1alpha3.AddToScheme(s)
	Expect(err).NotTo(HaveOccurred())

	mgr, err := manager.New(cfg, manager.Options{Scheme: s, MetricsBindAddress: "0"})
	Expect(err).NotTo(HaveOccurred())
	c = mgr.GetClient()

	reconciler := &controllers.APIReconciler{
		Client:            mgr.GetClient(),
		Log:               ctrl.Log.WithName("controllers").WithName("Api"),
		OathkeeperSvc:     testOathkeeperSvcURL,
		OathkeeperSvcPort: testOathkeeperPort,
		Validator: &validation.APIRule{
			DomainWhiteList: []string{"bar", "kyma.local"},
		},
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
