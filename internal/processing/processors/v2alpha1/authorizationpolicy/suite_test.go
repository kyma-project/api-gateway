package authorizationpolicy_test

import (
	"github.com/kyma-project/api-gateway/tests"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"
)

func Test(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "authorizationpolicy v1alpha2 Suite")
}

var _ = ReportAfterSuite("custom reporter", func(report types.Report) {
	tests.GenerateGinkgoJunitReport("authorizationpolicy-v1alpha2-suite", report)
})
