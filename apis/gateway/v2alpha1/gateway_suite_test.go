package v2alpha1_test

import (
	"fmt"
	"github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	rulev1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/reporters"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func TestGatewayv2alpha1(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Gateway v2alpha1 Suite")
}

const reportFilename = "junit-gateway-v2alpha1.xml"

var _ = ReportAfterSuite("custom reporter", func(report types.Report) {
	logger := zap.New(zap.UseDevMode(true), zap.WriteTo(GinkgoWriter))

	if key, ok := os.LookupEnv("ARTIFACTS"); ok {
		reportsPath := fmt.Sprintf("%s/%s", key, reportFilename)
		logger.Info("Generating reports at", "location", reportsPath)
		err := reporters.GenerateJUnitReport(report, reportsPath)

		if err != nil {
			logger.Error(err, "Junit Report Generation Error")
		}
	} else {
		if err := os.MkdirAll("../../reports", 0755); err != nil {
			logger.Error(err, "could not create directory")
		}

		reportsPath := fmt.Sprintf("%s/%s", "../../reports", reportFilename)
		logger.Info("Generating reports at", "location", reportsPath)
		err := reporters.GenerateJUnitReport(report, reportsPath)

		if err != nil {
			logger.Error(err, "Junit Report Generation Error")
		}
	}
})

func createFakeClient(objs ...client.Object) client.Client {
	scheme := runtime.NewScheme()
	err := networkingv1beta1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())
	err = rulev1alpha1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())
	err = securityv1beta1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())
	err = v2alpha1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())
	err = corev1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	return fake.NewClientBuilder().WithScheme(scheme).WithObjects(objs...).Build()
}
