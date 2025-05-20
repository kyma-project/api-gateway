package reconciliations_test

import (
	"testing"

	certv1alpha1 "github.com/gardener/cert-management/pkg/apis/cert/v1alpha1"
	dnsv1alpha1 "github.com/gardener/external-dns-management/pkg/apis/dns/v1alpha1"
	"github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	"github.com/kyma-project/api-gateway/tests"
	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"
	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func createFakeClient(objects ...client.Object) client.Client {
	Expect(v1alpha1.AddToScheme(scheme.Scheme)).Should(Succeed())
	Expect(corev1.AddToScheme(scheme.Scheme)).Should(Succeed())
	Expect(v1alpha3.AddToScheme(scheme.Scheme)).Should(Succeed())
	Expect(v1beta1.AddToScheme(scheme.Scheme)).Should(Succeed())
	Expect(dnsv1alpha1.AddToScheme(scheme.Scheme)).Should(Succeed())
	Expect(certv1alpha1.AddToScheme(scheme.Scheme)).Should(Succeed())
	Expect(networkingv1beta1.AddToScheme(scheme.Scheme)).Should(Succeed())

	return fake.NewClientBuilder().WithScheme(scheme.Scheme).WithObjects(objects...).Build()
}

func TestResources(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Resources Suite")
}

var _ = ReportAfterSuite("custom reporter", func(report types.Report) {
	tests.GenerateGinkgoJunitReport("resources-suite", report)
})
