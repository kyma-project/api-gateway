package clusterconfig_test

import (
	"github.com/kyma-project/api-gateway/tests"
	"testing"

	"github.com/onsi/ginkgo/v2/types"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestClusterConfig(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Clusterconfig Suite")
}

var _ = ReportAfterSuite("custom reporter", func(report types.Report) {
	tests.GenerateGinkgoJunitReport("clusterconfig-suite", report)
})
