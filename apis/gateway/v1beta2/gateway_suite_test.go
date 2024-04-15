package v1beta2_test

import (
	"fmt"
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/reporters"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func TestGatewayV1beta2(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Gateway v1beta2 Suite")
}

const reportFilename = "junit-gateway-v1beta2.xml"

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
