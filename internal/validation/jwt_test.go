package validation_test

import (
	"testing"

	"github.com/ghodss/yaml"
	gatewayv2alpha1 "github.com/kyma-incubator/api-gateway/api/v2alpha1"
	"github.com/kyma-incubator/api-gateway/internal/validation"
	"gotest.tools/assert"
	"k8s.io/apimachinery/pkg/runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var (
	validYamlForJWT = `
issuer: http://dex.kyma.local
jwks: ["a", "b"]
mode:
  name: ALL
  config:
    scopes: ["foo", "bar"]
`
	invalidIssuer = `
issuer: this-is-not-an-url
`
	invalidJWTMode = `
issuer: http://dex.kyma.local
jwks: ["a", "b"]
mode:
  name: CLASSIFIED_MODE_DONT_USE
  config:
    top: secret
`
	logJWT = logf.Log.WithName("jwt-validate-test")
)

func TestJWTValidate(t *testing.T) {
	strategy, err := validation.NewFactory(logJWT).StrategyFor(gatewayv2alpha1.JWT)
	assert.NilError(t, err)

	jsonData, err := yaml.YAMLToJSON([]byte(invalidIssuer))
	assert.NilError(t, err)
	assert.Error(t, strategy.Validate(&runtime.RawExtension{Raw: jsonData}), "issuer field is empty or not a valid url")

	jsonData, err = yaml.YAMLToJSON([]byte(invalidJWTMode))
	assert.NilError(t, err)
	assert.Error(t, strategy.Validate(&runtime.RawExtension{Raw: jsonData}), "supplied mode is invalid: CLASSIFIED_MODE_DONT_USE, valid modes are: ALL, INCLUDE, EXCLUDE")

	jsonData, err = yaml.YAMLToJSON([]byte(validYamlForJWT))
	assert.NilError(t, err)
	assert.NilError(t, strategy.Validate(&runtime.RawExtension{Raw: jsonData}))
}
