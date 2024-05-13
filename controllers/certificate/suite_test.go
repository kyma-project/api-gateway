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

package certificate_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/kyma-project/api-gateway/tests"
	. "github.com/onsi/ginkgo/v2"
	gikgotypes "github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/retry"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	//+kubebuilder:scaffold:imports
)

const (
	testNamespace     = "kyma-system"
	eventuallyTimeout = time.Second * 20
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

	RunSpecs(t, "API Gateway Certificate Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))
	ctx, cancel = context.WithCancel(context.Background())

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "..", "config", "crd", "bases"), filepath.Join("..", "..", "hack")},
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
	patchAPIRuleCRD(k8sClient)

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme: s,
		Metrics: metricsserver.Options{
			BindAddress: "0",
		},
	})
	Expect(err).NotTo(HaveOccurred())

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

var _ = ReportAfterSuite("custom reporter", func(report gikgotypes.Report) {
	tests.GenerateGinkgoJunitReport("api-gateway-certificate-suite", report)
})

func createCommonTestResources(k8sClient client.Client) {
	kymaSystemNs := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: testNamespace},
		Spec:       corev1.NamespaceSpec{},
	}
	Expect(k8sClient.Create(context.TODO(), kymaSystemNs)).Should(Succeed())
}

func patchAPIRuleCRD(k8sClient client.Client) {
	crd := &apiextensionsv1.CustomResourceDefinition{}
	Expect(k8sClient.Get(ctx, types.NamespacedName{Name: "apirules.gateway.kyma-project.io"}, crd)).Should(Succeed())

	mergeFrom := ctrlclient.StrategicMergeFrom(crd.DeepCopy())
	crd.Spec.Conversion.Strategy = apiextensionsv1.WebhookConverter
	crd.Spec.Conversion.Webhook = &apiextensionsv1.WebhookConversion{
		ClientConfig: &apiextensionsv1.WebhookClientConfig{
			Service: &apiextensionsv1.ServiceReference{
				Namespace: testNamespace,
				Name:      "api-gateway-webhook-service",
				Path:      ptr.To("/convert"),
				Port:      ptr.To(int32(9443)),
			},
		},
		ConversionReviewVersions: []string{"v1beta1", "v1beta2"},
	}

	Expect(k8sClient.Patch(ctx, crd, mergeFrom)).Should(Succeed())
}

func createFakeClient(objects ...client.Object) client.Client {
	fakeClient := fake.NewClientBuilder().WithScheme(getTestScheme()).WithObjects(objects...).WithStatusSubresource(objects...).Build()

	return fakeClient
}

func getTestScheme() *runtime.Scheme {
	s := runtime.NewScheme()
	utilruntime.Must(corev1.AddToScheme(s))
	utilruntime.Must(apiextensionsv1.AddToScheme(s))
	return s
}
