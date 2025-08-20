package operator

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	"github.com/kyma-project/api-gateway/controllers"
	"github.com/kyma-project/api-gateway/controllers/gateway"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// APIGatewayReconciler reconciles a APIGateway object
type APIGatewayReconciler struct {
	client.Client
	Scheme                   *runtime.Scheme
	log                      logr.Logger
	oathkeeperReconciler     ReadyVerifyingReconciler
	apiRuleReconcilerStarter gateway.APIRuleReconcilerStarter
}

type ReadyVerifyingReconciler interface {
	// ReconcileAndVerifyReadiness runs the reconciliation and verifies that the resource is ready.
	ReconcileAndVerifyReadiness(ctx context.Context, k8sClient client.Client, apiGatewayCR *v1alpha1.APIGateway) controllers.Status
}
