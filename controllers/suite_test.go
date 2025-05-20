package controllers

import (
	"testing"

	"github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	"github.com/kyma-project/api-gateway/tests"
	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestValidators(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Controllers Suite")
}

var _ = ReportAfterSuite("custom reporter", func(report types.Report) {
	tests.GenerateGinkgoJunitReport("controllers-suite", report)
})

func createFakeClient(objects ...client.Object) client.Client {
	Expect(v1alpha1.AddToScheme(scheme.Scheme)).Should(Succeed())

	return fake.NewClientBuilder().WithScheme(scheme.Scheme).WithObjects(objects...).WithStatusSubresource(objects...).Build()
}
