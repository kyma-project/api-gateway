package ratelimit_test

import (
	"github.com/kyma-project/api-gateway/tests"
	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"
	"testing"
)

func Test(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Ratelimit Controller Suite")
}

var _ = ReportAfterSuite("custom reporter", func(report types.Report) {
	tests.GenerateGinkgoJunitReport("ratelimit-controller-suite", report)
})
