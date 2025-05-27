package processing

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"

	"github.com/kyma-project/api-gateway/tests"
)

func TestProcessing(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Processing Suite")
}

var _ = ReportAfterSuite("custom reporter", func(report types.Report) {
	tests.GenerateGinkgoJunitReport("processing", report)
})
