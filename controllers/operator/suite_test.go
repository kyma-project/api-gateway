/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package operator

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	ratelimitv1alpha1 "github.com/kyma-project/api-gateway/apis/gateway/ratelimit/v1alpha1"
	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	operatorv1alpha1 "github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	"github.com/kyma-project/api-gateway/controllers"
	"github.com/kyma-project/api-gateway/internal/reconciliations/oathkeeper"
	"github.com/kyma-project/api-gateway/internal/resources"
	"github.com/kyma-project/api-gateway/tests"
	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"
	oryv1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
	networkingv1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	schedulingv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

const (
	testNamespace    = "kyma-system"
	apiGatewayCRName = "default"

	eventuallyTimeout = time.Second * 20

	kymaNamespace   = "kyma-system"
	kymaGatewayName = "kyma-gateway"

	controlledList = `
resources:
  - GroupVersionKind:
      group: gateway.kyma-project.io
      version: v1beta1
      kind: APIRule
  - GroupVersionKind:
      group: networking.istio.io
      version: v1beta1
      kind: VirtualService
    ControlledList:
      - name: "istio-healthz"
        namespace: "istio-system"
`
)

var (
	cfg       *rest.Config
	k8sClient client.Client
	testEnv   *envtest.Environment
	ctx       context.Context
	cancel    context.CancelFunc
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "API Gateway Operator Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))
	ctx, cancel = context.WithCancel(context.Background())

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.FromSlash("../../config/crd/bases"),
			filepath.FromSlash("../../hack/crds"),
		},
		ErrorIfCRDPathMissing: true,
	}

	var err error
	// cfg is defined in this file globally.
	cfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	s := getTestScheme()

	//+kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: s})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	createCommonTestResources(k8sClient)

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme: s,
		Metrics: metricsserver.Options{
			BindAddress: "0",
		},
	})
	Expect(err).NotTo(HaveOccurred())

	rateLimiterCfg := controllers.RateLimiterConfig{
		Burst:            200,
		Frequency:        30,
		FailureBaseDelay: 1 * time.Second,
		FailureMaxDelay:  10 * time.Second,
	}

	resources.ReadResourcesFileHandle = func(name string) ([]byte, error) {
		return []byte(controlledList), nil
	}

	Expect(NewAPIGatewayReconciler(mgr, oathkeeperReconcilerWithoutVerification{}).SetupWithManager(mgr, rateLimiterCfg)).Should(Succeed())

	go func() {
		defer GinkgoRecover()
		err := mgr.Start(ctx)
		// A workaround for DeadlineExceeded error is introduced, since this started occurring during the teardown
		// after adding Oathkeeper reconciliation.
		if !errors.Is(err, context.DeadlineExceeded) {
			Expect(err).Should(Succeed())
		} else {
			println("Context deadline exceeded during tearing down", err.Error())
		}
	}()
})

var _ = AfterSuite(func() {
	/*
		 Provided solution for timeout issue waiting for kubeapiserver
			https://github.com/kubernetes-sigs/controller-runtime/issues/1571#issuecomment-1005575071
	*/
	cancel()
	By("Tearing down the test environment")
	err := retry.OnError(wait.Backoff{
		Duration: 500 * time.Millisecond,
		Steps:    150,
	}, func(err error) bool {
		return true
	}, func() error {
		return testEnv.Stop()
	})
	Expect(err).NotTo(HaveOccurred())
})

var _ = ReportAfterSuite("custom reporter", func(report types.Report) {
	tests.GenerateGinkgoJunitReport("api-gateway-operator-suite", report)
})

func createCommonTestResources(k8sClient client.Client) {
	kymaSystemNs := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: testNamespace},
		Spec:       corev1.NamespaceSpec{},
	}
	Expect(k8sClient.Create(context.Background(), kymaSystemNs)).Should(Succeed())

	istioSystemNs := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: "istio-system"},
		Spec:       corev1.NamespaceSpec{},
	}
	Expect(k8sClient.Create(context.Background(), istioSystemNs)).Should(Succeed())
}

func createFakeClient(objects ...client.Object) client.Client {

	crds := []apiextensionsv1.CustomResourceDefinition{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "gateways.networking.istio.io",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "virtualservices.networking.istio.io",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "dnsentries.dns.gardener.cloud",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "certificates.cert.gardener.cloud",
			},
		},
	}

	fakeClient := fake.NewClientBuilder().WithScheme(getTestScheme()).WithObjects(objects...).WithStatusSubresource(objects...).Build()

	for _, crd := range crds {
		Expect(fakeClient.Create(context.Background(), &crd)).To(Succeed())
	}

	return fakeClient
}

func getTestScheme() *runtime.Scheme {
	s := runtime.NewScheme()
	utilruntime.Must(corev1.AddToScheme(s))
	utilruntime.Must(appsv1.AddToScheme(s))
	utilruntime.Must(rbacv1.AddToScheme(s))
	utilruntime.Must(policyv1.AddToScheme(s))
	utilruntime.Must(operatorv1alpha1.AddToScheme(s))
	utilruntime.Must(autoscalingv2.AddToScheme(s))
	utilruntime.Must(securityv1beta1.AddToScheme(s))
	utilruntime.Must(schedulingv1.AddToScheme(s))
	utilruntime.Must(apiextensionsv1.AddToScheme(s))
	utilruntime.Must(gatewayv1beta1.AddToScheme(s))
	utilruntime.Must(networkingv1alpha3.AddToScheme(s))
	utilruntime.Must(networkingv1beta1.AddToScheme(s))
	utilruntime.Must(oryv1alpha1.AddToScheme(s))
	utilruntime.Must(ratelimitv1alpha1.AddToScheme(s))

	return s
}

type oathkeeperReconcilerWithoutVerification struct {
}

func (o oathkeeperReconcilerWithoutVerification) ReconcileAndVerifyReadiness(
	ctx context.Context,
	client client.Client,
	apiGateway *operatorv1alpha1.APIGateway,
) controllers.Status {
	// We don't want to wait for Oathkeeper to be ready in the tests, because the implemented logic doesn't work in unit and envTest-based tests.
	return oathkeeper.Reconcile(ctx, client, apiGateway)
}
