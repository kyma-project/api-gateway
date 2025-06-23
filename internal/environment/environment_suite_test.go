package environment_test

import (
	"github.com/kyma-project/api-gateway/tests"
	"github.com/onsi/ginkgo/v2/types"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestEnvironment(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "API-Gateway Helpers Suite")
}

var _ = ReportAfterSuite("custom reporter", func(report types.Report) {
	tests.GenerateGinkgoJunitReport("api-gateway-helpers-suite", report)
})

func createFakeClient(objects ...client.Object) client.Client {
	Expect(corev1.AddToScheme(scheme.Scheme)).Should(Succeed())

	return fake.NewClientBuilder().WithScheme(scheme.Scheme).WithObjects(objects...).WithStatusSubresource(objects...).Build()
}
