package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kyma-project/api-gateway/internal/validation"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
)

const (
	configMapName      = "api-gateway-config"
	configMapNamespace = "kyma-system"
)

func (r *APIRuleReconciler) reconcileConfigMap(ctx context.Context, isCMReconcile bool) (finishReconciliation bool) {
	r.Log.Info("Starting ConfigMap reconciliation")
	err := r.Config.ReadFromConfigMap(ctx, r.Client)
	if err != nil {
		if apierrs.IsNotFound(err) {
			r.Log.Info(fmt.Sprintf(`ConfigMap %s in namespace %s was not found {"controller": "ApiRule"}, will use default config`, configMapName, configMapNamespace))
			r.Config.ResetToDefault()
		} else {
			r.Log.Error(err, fmt.Sprintf(`could not read ConfigMap %s in namespace %s {"controller": "ApiRule"}`, configMapName, configMapNamespace))
			r.Config.Reset()
		}
	}
	if isCMReconcile {
		configValidationFailures := validation.ValidateConfig(r.Config)
		r.Log.Info("ConfigMap changed", "config", r.Config)
		if len(configValidationFailures) > 0 {
			failuresJson, _ := json.Marshal(configValidationFailures)
			r.Log.Error(err, fmt.Sprintf(`Config validation failure {"controller": "ApiRule", "failures": %s}`, string(failuresJson)))
		}
		r.Log.Info("ConfigMap reconciliation finished")
		return true
	}
	return false
}
