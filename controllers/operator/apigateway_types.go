package operator

import (
	"github.com/go-logr/logr"
	"github.com/kyma-project/api-gateway/internal/operator/reconciliations/api_gateway"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// APIGatewayReconciler reconciles a APIGateway object
type APIGatewayReconciler struct {
	client.Client
	Scheme                   *runtime.Scheme
	log                      logr.Logger
	apiGatewayReconciliation api_gateway.ApiGatewayReconciliation
}
