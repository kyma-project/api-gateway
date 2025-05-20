package processing

import (
	"testing"

	"github.com/kyma-project/api-gateway/tests"
	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"
)

func TestProcessing(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Processing Suite")
}

var _ = ReportAfterSuite("custom reporter", func(report types.Report) {
	tests.GenerateGinkgoJunitReport("processing", report)
})
