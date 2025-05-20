package ory_test

import (
	"fmt"
	"net/http"
	"os"
	"testing"

	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/reporters"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func TestOry(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Ory Suite")
}

var _ = ReportAfterSuite("custom reporter", func(report types.Report) {
	logger := zap.New(zap.UseDevMode(true), zap.WriteTo(GinkgoWriter))

	if key, ok := os.LookupEnv("ARTIFACTS"); ok {
		reportsFilename := fmt.Sprintf("%s/%s", key, "junit-ory.xml")
		logger.Info("Generating reports at", "location", reportsFilename)
		err := reporters.GenerateJUnitReport(report, reportsFilename)

		if err != nil {
			logger.Error(err, "Junit Report Generation Error")
		}
	} else {
		if err := os.MkdirAll("../../../reports", 0755); err != nil {
			logger.Error(err, "could not create directory")
		}

		reportsFilename := fmt.Sprintf("%s/%s", "../../../reports", "junit-ory.xml")
		logger.Info("Generating reports at", "location", reportsFilename)
		err := reporters.GenerateJUnitReport(report, reportsFilename)

		if err != nil {
			logger.Error(err, "Junit Report Generation Error")
		}
	}
})

var testLogger = ctrl.Log.WithName("ory-test")

var methodsGet = []gatewayv1beta1.HttpMethod{http.MethodGet}
var methodsPost = []gatewayv1beta1.HttpMethod{http.MethodPost}
