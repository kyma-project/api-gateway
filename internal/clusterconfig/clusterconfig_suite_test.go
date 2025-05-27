package clusterconfig_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"

	"github.com/kyma-project/api-gateway/tests"
)

func TestClusterConfig(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Clusterconfig Suite")
}

var _ = ReportAfterSuite("custom reporter", func(report types.Report) {
	tests.GenerateGinkgoJunitReport("clusterconfig-suite", report)
})
