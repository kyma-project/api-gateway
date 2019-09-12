package validation

import (
	"fmt"

	"github.com/go-logr/logr"

	gatewayv2alpha1 "github.com/kyma-incubator/api-gateway/api/v2alpha1"
	"k8s.io/apimachinery/pkg/runtime"
)

//Factory .
type Factory struct {
	Log logr.Logger
}

//Strategy .
type Strategy interface {
	Validate(gate *gatewayv2alpha1.Gate) error
}

//NewFactory .
func NewFactory(logger logr.Logger) *Factory {
	return &Factory{
		Log: logger,
	}
}

//StrategyFor .
func (f *Factory) StrategyFor(strategyName string) (Strategy, error) {
	switch strategyName {
	case gatewayv2alpha1.Allow:
		f.Log.Info("Allow validation mode detected")
		return &allow{}, nil
	case gatewayv2alpha1.Jwt:
		f.Log.Info("JWT validation mode detected")
		return &jwt{}, nil
	case gatewayv2alpha1.Oauth:
		f.Log.Info("OAUTH validation mode detected")
		return &oauth{}, nil
	default:
		return nil, fmt.Errorf("Unsupported mode: %s", strategyName)
	}
}

//configNotEmpty Verify if the config object is not empty
func configNotEmpty(config *runtime.RawExtension) bool {
	if config == nil {
		return false
	}
	return len(config.Raw) != 0
}
