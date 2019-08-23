package processing

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	gatewayv2alpha1 "github.com/kyma-incubator/api-gateway/api/v2alpha1"
)

type factory struct {
	Client client.Client
	Log    logr.Logger
}

type ProcessingStrategy interface {
	Process(ctx context.Context, api *gatewayv2alpha1.Gate) error
}

func NewFactory(client client.Client, logger logr.Logger) *factory {
	return &factory{
		Client: client,
		Log:    logger,
	}
}

func (f *factory) StrategyFor(strategyName string) (ProcessingStrategy, error) {
	switch strategyName {
	case gatewayv2alpha1.PASSTHROUGH:
		f.Log.Info("PASSTHROUGH processing mode detected")
		return &passthrough{Client: f.Client}, nil
	default:
		return nil, fmt.Errorf("Unsupported mode: %s", strategyName)
	}
}
