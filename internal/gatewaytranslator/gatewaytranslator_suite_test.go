package gatewaytranslator_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"

	"github.com/kyma-project/api-gateway/tests"
)

func TestGatewayTranslator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Gateway translator Suite")
}

var _ = ReportAfterSuite("custom reporter", func(report types.Report) {
	tests.GenerateGinkgoJunitReport("gatewaytaranslator", report)
})
