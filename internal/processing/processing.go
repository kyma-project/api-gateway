package processing

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	istioClient "github.com/kyma-incubator/api-gateway/internal/clients/istio"
	oryClient "github.com/kyma-incubator/api-gateway/internal/clients/ory"

	gatewayv2alpha1 "github.com/kyma-incubator/api-gateway/api/v2alpha1"
)

//Factory .
type Factory struct {
	vsClient          *istioClient.VirtualService
	arClient          *oryClient.AccessRule
	Log               logr.Logger
	oathkeeperSvc     string
	oathkeeperSvcPort uint32
	JWKSURI           string
}

//Strategy .
type Strategy interface {
	Process(ctx context.Context, api *gatewayv2alpha1.Gate) error
}

//NewFactory .
func NewFactory(vsClient *istioClient.VirtualService, arClient *oryClient.AccessRule, logger logr.Logger, oathkeeperSvc string, oathkeeperSvcPort uint32, jwksURI string) *Factory {
	return &Factory{
		vsClient:          vsClient,
		arClient:          arClient,
		Log:               logger,
		oathkeeperSvc:     oathkeeperSvc,
		oathkeeperSvcPort: oathkeeperSvcPort,
		JWKSURI:           jwksURI,
	}
}

//StrategyFor .
func (f *Factory) StrategyFor(strategyName string) (Strategy, error) {
	switch strategyName {
	case gatewayv2alpha1.Allow:
		f.Log.Info("Allow processing mode detected")
		return &allow{vsClient: f.vsClient, oathkeeperSvc: f.oathkeeperSvc, oathkeeperSvcPort: f.oathkeeperSvcPort}, nil
	case gatewayv2alpha1.Jwt:
		f.Log.Info("JWT processing mode detected")
		return &jwt{vsClient: f.vsClient, arClient: f.arClient, JWKSURI: f.JWKSURI, oathkeeperSvc: f.oathkeeperSvc, oathkeeperSvcPort: f.oathkeeperSvcPort}, nil
	case gatewayv2alpha1.Oauth:
		f.Log.Info("OAUTH processing mode detected")
		return &oauth{vsClient: f.vsClient, arClient: f.arClient, oathkeeperSvc: f.oathkeeperSvc, oathkeeperSvcPort: f.oathkeeperSvcPort}, nil
	default:
		return nil, fmt.Errorf("unsupported mode: %s", strategyName)
	}
}
