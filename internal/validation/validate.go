package validation

import (
	"fmt"
	"github.com/go-logr/logr"

	gatewayv2alpha1 "github.com/kyma-incubator/api-gateway/api/v2alpha1"
	"k8s.io/apimachinery/pkg/runtime"
)

type factory struct {
	Log logr.Logger
}

type ValidationStrategy interface {
	Validate(config *runtime.RawExtension) error
}

func NewFactory(logger logr.Logger) *factory {
	return &factory{
		Log: logger,
	}
}

func (f *factory) StrategyFor(strategyName string) (ValidationStrategy, error) {
	switch strategyName {
	case gatewayv2alpha1.PASSTHROUGH:
		f.Log.Info("PASSTHROUGH validation mode detected")
		return &passthrough{}, nil
	case gatewayv2alpha1.JWT:
		f.Log.Info("JWT validation mode detected")
		return &jwt{}, nil
	case gatewayv2alpha1.OAUTH:
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
