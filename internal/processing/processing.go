package processing

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	istioClient "github.com/kyma-incubator/api-gateway/internal/clients/istio"

	gatewayv2alpha1 "github.com/kyma-incubator/api-gateway/api/v2alpha1"
)

type factory struct {
	vsClient *istioClient.VirtualService
	Log      logr.Logger
}

type ProcessingStrategy interface {
	Process(ctx context.Context, api *gatewayv2alpha1.Gate) error
}

func NewFactory(vsClient *istioClient.VirtualService, logger logr.Logger) *factory {
	return &factory{
		vsClient: vsClient,
		Log:      logger,
	}
}

func (f *factory) StrategyFor(strategyName string) (ProcessingStrategy, error) {
	switch strategyName {
	case gatewayv2alpha1.PASSTHROUGH:
		f.Log.Info("PASSTHROUGH processing mode detected")
		return &passthrough{vsClient: f.vsClient}, nil
	default:
		return nil, fmt.Errorf("Unsupported mode: %s", strategyName)
	}
}
