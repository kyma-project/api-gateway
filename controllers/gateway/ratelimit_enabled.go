//go:build dev_features

package gateway

import (
	gatewayv1alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func SetupRateLimit(mgr manager.Manager, scheme *runtime.Scheme) error {
	utilruntime.Must(gatewayv1alpha1.AddToScheme(scheme))
	return (&RateLimitReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr)
}
