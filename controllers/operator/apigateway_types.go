package operator

import (
	"github.com/go-logr/logr"
	"github.com/kyma-project/api-gateway/controllers"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// APIGatewayReconciler reconciles a APIGateway object
type APIGatewayReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	log           logr.Logger
	statusHandler controllers.StatusHandler
}
