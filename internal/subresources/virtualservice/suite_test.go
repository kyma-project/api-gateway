package virtualservice_test

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

func TestVirtualService(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "VirtualService Repository Suite")
}

var _ = ReportAfterSuite("custom reporter", func(report types.Report) {
	logger := zap.New(zap.UseDevMode(true), zap.WriteTo(GinkgoWriter))

	if key, ok := os.LookupEnv("ARTIFACTS"); ok {
		reportsFilename := fmt.Sprintf("%s/%s", key, "junit-virtualservice.xml")
		logger.Info("Generating reports at", "location", reportsFilename)
		err := reporters.GenerateJUnitReport(report, reportsFilename)

		if err != nil {
			logger.Error(err, "Junit Report Generation Error")
		}
	} else {
		if err := os.MkdirAll("../../../reports", 0755); err != nil {
			logger.Error(err, "could not create directory")
		}

		reportsFilename := fmt.Sprintf("%s/%s", "../../../reports", "junit-virtualservice.xml")
		logger.Info("Generating reports at", "location", reportsFilename)
		err := reporters.GenerateJUnitReport(report, reportsFilename)

		if err != nil {
			logger.Error(err, "Junit Report Generation Error")
		}
	}
})
