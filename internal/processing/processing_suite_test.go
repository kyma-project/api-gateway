package processing

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	"testing"
)

func TestProcessing(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecsWithDefaultAndCustomReporters(t, "Processing Suite",
		[]Reporter{printer.NewProwReporter("api-gateway-processing-testsuite")})
}
