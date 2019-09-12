package validation_test

import (
	"testing"

	"k8s.io/apimachinery/pkg/types"

	gatewayv2alpha1 "github.com/kyma-incubator/api-gateway/api/v2alpha1"
	"github.com/kyma-incubator/api-gateway/internal/validation"
	"gotest.tools/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var (
	log = logf.Log.WithName("allow-validate-test")
)

func TestPassthroughValidate(t *testing.T) {
	strategy, err := validation.NewFactory(log).StrategyFor(gatewayv2alpha1.Allow)
	assert.NilError(t, err)

	valid := getPassthroughValidGate()
	assert.NilError(t, strategy.Validate(valid))

	notValid := getPassthroughNotValidGate()
	assert.Error(t, strategy.Validate(notValid), "allow mode does not support scopes")
}

func getPassthroughValidGate() *gatewayv2alpha1.Gate {
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
			Paths: []gatewayv2alpha1.Path{
				{
					Path:    "/.*",
					Methods: []string{"GET"},
				},
			},
		},
	}
}

func getPassthroughNotValidGate() *gatewayv2alpha1.Gate {
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
			Paths: []gatewayv2alpha1.Path{
				{
					Path:    "/.*",
					Methods: []string{"GET"},
					Scopes:  []string{"read", "write"},
				},
			},
		},
	}
}
