package conditions_test

import (
	"testing"

	"github.com/kyma-project/api-gateway/tests"
	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"
)

func TestValidators(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Conditions Suite")
}

var _ = ReportAfterSuite("custom reporter", func(report types.Report) {
	tests.GenerateGinkgoJunitReport("condition-suite", report)
})
