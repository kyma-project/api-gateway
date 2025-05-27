package gateway

import (
	"context"
	"encoding/json"
	"fmt"

	apierrs "k8s.io/apimachinery/pkg/api/errors"

	"github.com/kyma-project/api-gateway/internal/validation"
)

const (
	configMapName      = "api-gateway-config"
	configMapNamespace = "kyma-system"
)

func (r *APIRuleReconciler) reconcileConfigMap(ctx context.Context, isCMReconcile bool) (finishReconciliation bool) {
	err := r.Config.ReadFromConfigMap(ctx, r.Client)
	if err != nil {
		if apierrs.IsNotFound(err) {
			r.Log.Info(fmt.Sprintf(`ConfigMap %s in namespace %s was not found {"controller": "APIRule"}, will use default config`, configMapName, configMapNamespace))
			r.Config.ResetToDefault()
		} else {
			r.Log.Error(err, fmt.Sprintf(`could not read ConfigMap %s in namespace %s {"controller": "APIRule"}`, configMapName, configMapNamespace))
			r.Config.Reset()
		}
	}

	if isCMReconcile {
		r.Log.Info("Starting ConfigMap reconciliation")
		configValidationFailures := validation.ValidateConfig(r.Config)
		r.Log.Info("ConfigMap changed", "config", r.Config)
		if len(configValidationFailures) > 0 {
			failuresJson, _ := json.Marshal(configValidationFailures)
			r.Log.Error(err, fmt.Sprintf(`Config validation failure {"controller": "APIRule", "failures": %s}`, string(failuresJson)))
		}
		r.Log.Info("ConfigMap reconciliation finished")
		return true
	}
	return false
}
