package processing

import (
	"context"
	"fmt"

	gatewayv2alpha1 "github.com/kyma-incubator/api-gateway/api/v2alpha1"
	istioClient "github.com/kyma-incubator/api-gateway/internal/clients/istio"
)

type jwt struct {
	vsClient *istioClient.VirtualService
	apClient *istioClient.AuthenticationPolicy
}

func (j *jwt) Process(ctx context.Context, api *gatewayv2alpha1.Gate) error {
	fmt.Println("Processing API")
	// DO STUFF HERE
	return nil
}
