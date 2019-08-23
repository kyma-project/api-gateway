package validation_test

import (
	"testing"

	gatewayv2alpha1 "github.com/kyma-incubator/api-gateway/api/v2alpha1"
	"github.com/kyma-incubator/api-gateway/internal/validation"
	"gotest.tools/assert"
	"k8s.io/apimachinery/pkg/runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var (
	validYaml    = ``
	notValidYaml = `
config:
  foo: bar
`
	log = logf.Log.WithName("passthrough-validate-test")
)

func TestPassthroughValidate(t *testing.T) {
	strategy, err := validation.NewFactory(log).StrategyFor(gatewayv2alpha1.PASSTHROUGH)
	assert.NilError(t, err)

	valid := &runtime.RawExtension{Raw: []byte(validYaml)}
	assert.NilError(t, strategy.Validate(valid))

	notValid := &runtime.RawExtension{Raw: []byte(notValidYaml)}
	assert.Error(t, strategy.Validate(notValid), "passthrough mode requires empty configuration")
}
