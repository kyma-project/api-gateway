package external_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	gardenerCertv1alpha1 "github.com/gardener/cert-management/pkg/apis/cert/v1alpha1"
	gardenerDNSv1alpha1 "github.com/gardener/external-dns-management/pkg/apis/dns/v1alpha1"
	networkingv1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	externalv1alpha1 "github.com/kyma-project/api-gateway/apis/gateway/external/v1alpha1"
	"github.com/kyma-project/api-gateway/controllers"
	"github.com/kyma-project/api-gateway/controllers/gateway/external"
)

const (
	testNamespace = "test-namespace"
	kubeSystemNs  = "kube-system"
	istioSystemNs = "istio-system"
)

var (
	cfg       *rest.Config
	testEnv   *envtest.Environment
	k8sClient client.Client
	ctx       context.Context
	cancel    context.CancelFunc
)

func TestMain(m *testing.M) {

	logf.SetLogger(zap.New(zap.UseDevMode(true)))
	ctx, cancel = context.WithCancel(context.Background())

	// Create scheme with all required types
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		panic(err)
	}
	if err := externalv1alpha1.AddToScheme(scheme); err != nil {
		panic(err)
	}
	if err := networkingv1alpha3.AddToScheme(scheme); err != nil {
		panic(err)
	}
	if err := networkingv1beta1.AddToScheme(scheme); err != nil {
		panic(err)
	}
	if err := gardenerDNSv1alpha1.AddToScheme(scheme); err != nil {
		panic(err)
	}
	if err := gardenerCertv1alpha1.AddToScheme(scheme); err != nil {
		panic(err)
	}

	// Setup test environment
	testEnv = &envtest.Environment{
		CRDInstallOptions: envtest.CRDInstallOptions{
			Scheme: scheme,
		},

		CRDDirectoryPaths: []string{
			filepath.FromSlash("../../../config/crd/bases"),
			filepath.FromSlash("../../../hack/crds"),
			filepath.FromSlash("../../../hack/crds/gardener"),
		},
	}

	var err error
	cfg, err = testEnv.Start()
	if err != nil {
		logf.Log.Error(err, "Failed to start envtest. Integration tests require envtest binaries. Run 'make test' instead of 'go test'.")
		panic(err)
	}

	// Create manager
	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme: scheme,
		Metrics: metricsserver.Options{
			BindAddress: "0",
		},
	})
	if err != nil {
		panic(err)
	}

	// Setup controller
	reconciler := external.NewExternalGatewayReconciler(mgr)
	rateLimiterCfg := controllers.RateLimiterConfig{
		Burst:            200,
		Frequency:        30,
		FailureBaseDelay: 1 * time.Second,
		FailureMaxDelay:  10 * time.Second,
	}
	if err := reconciler.SetupWithManager(mgr, rateLimiterCfg); err != nil {
		panic(err)
	}

	// Create k8s client
	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme})
	if err != nil {
		panic(err)
	}

	// Create test namespaces
	testNs := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: testNamespace},
	}
	if err := k8sClient.Create(ctx, testNs); err != nil {
		panic(err)
	}

	// kube-system is created by envtest automatically, so we don't need to create it
	// Just verify it exists
	kubeSystemNsObj := &corev1.Namespace{}
	if err := k8sClient.Get(ctx, client.ObjectKey{Name: kubeSystemNs}, kubeSystemNsObj); err != nil {
		panic(err)
	}

	istioSystemNsObj := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: istioSystemNs},
	}
	if err := k8sClient.Create(ctx, istioSystemNsObj); err != nil {
		panic(err)
	}

	// Create regions ConfigMap
	externalRegionsConfigMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "external-gateway-regions",
			Namespace: testNamespace,
		},
		Data: map[string]string{
			"regions.yaml": `
- Provider: aws
  Region: us-east-1
  CertSubjects:
    - "CN=test-ca-aws"
- Provider: gcp
  Region: europe-west1
  CertSubjects:
    - "CN=test-ca-gcp"
`,
		},
	}
	if err := k8sClient.Create(ctx, externalRegionsConfigMap); err != nil {
		panic(err)
	}

	// Start manager
	go func() {
		if err := mgr.Start(ctx); err != nil {
			panic(err)
		}
	}()

	// Wait for manager to be ready and CRDs to be established
	// The controller needs time to start and the CRD needs to be fully registered in the API server
	time.Sleep(5 * time.Second)

	// Run tests
	code := m.Run()

	// Cleanup
	cancel()
	if err := testEnv.Stop(); err != nil {
		time.Sleep(4 * time.Second)
		_ = testEnv.Stop()
	}

	// Exit with test result code
	if code != 0 {
		os.Exit(code)
	}
}
