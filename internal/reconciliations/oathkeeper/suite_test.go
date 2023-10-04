package oathkeeper_test

import (
	"github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	"github.com/kyma-project/api-gateway/tests"
	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"
	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

func TestValidators(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Oathkeeper Suite")
}

var _ = ReportAfterSuite("custom reporter", func(report types.Report) {
	tests.GenerateGinkgoJunitReport("oauthkeeper-suite", report)
})

func createFakeClient(objects ...client.Object) client.Client {
	Expect(v1alpha1.AddToScheme(scheme.Scheme)).Should(Succeed())
	Expect(corev1.AddToScheme(scheme.Scheme)).Should(Succeed())
	Expect(v1alpha3.AddToScheme(scheme.Scheme)).Should(Succeed())
	Expect(v1beta1.AddToScheme(scheme.Scheme)).Should(Succeed())
	Expect(apiextensionsv1.AddToScheme(scheme.Scheme)).Should(Succeed())
	Expect(securityv1beta1.AddToScheme(scheme.Scheme)).Should(Succeed())

	return fake.NewClientBuilder().WithScheme(scheme.Scheme).WithObjects(objects...).Build()
}
