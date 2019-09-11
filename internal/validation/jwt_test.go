package validation_test

import (
	"k8s.io/apimachinery/pkg/types"
	"testing"

	"github.com/ghodss/yaml"
	gatewayv2alpha1 "github.com/kyma-incubator/api-gateway/api/v2alpha1"
	"github.com/kyma-incubator/api-gateway/internal/validation"
	"gotest.tools/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	logJWT = logf.Log.WithName("jwt-validate-test")
)

func TestJWTValidate(t *testing.T) {
	strategy, err := validation.NewFactory(logJWT).StrategyFor(gatewayv2alpha1.Jwt)
	assert.NilError(t, err)

	jsonData, err := yaml.YAMLToJSON([]byte(invalidIssuer))
	assert.NilError(t, err)
	invalidIssuerGate := getJWTGate(&runtime.RawExtension{Raw: jsonData})
	assert.Error(t, strategy.Validate(invalidIssuerGate), "issuer field is empty or not a valid url")

	jsonData, err = yaml.YAMLToJSON([]byte(validYamlForJWT))
	validGate := getJWTGate(&runtime.RawExtension{Raw: jsonData})
	assert.NilError(t, err)
	assert.NilError(t, strategy.Validate(validGate))
}

func getJWTGate(config *runtime.RawExtension) *gatewayv2alpha1.Gate {
	var apiUID types.UID = "eab0f1c8-c417-11e9-bf11-4ac644044351"
	var apiGateway = "some-gateway"
	var serviceName = "test-service"
	var serviceHost = "myService.myDomain.com"
	var servicePort uint32 = 8080

	return &gatewayv2alpha1.Gate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-gate",
			UID:       apiUID,
			Namespace: "test-namespace",
		},
		TypeMeta: metav1.TypeMeta{
			APIVersion: "gateway.kyma-project.io/v2alpha1",
			Kind:       "Gate",
		},
		Spec: gatewayv2alpha1.GateSpec{
			Gateway: &apiGateway,
			Service: &gatewayv2alpha1.Service{
				Name: &serviceName,
				Host: &serviceHost,
				Port: &servicePort,
			},
			Auth: &gatewayv2alpha1.AuthStrategy{
				Config: config,
			},
			Paths: []gatewayv2alpha1.Path{
				{
					Path:    "/.*",
					Scopes:  []string{},
					Methods: []string{},
				},
			},
		},
	}
}
