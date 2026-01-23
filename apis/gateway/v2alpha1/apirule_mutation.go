package v2alpha1

import (
	"context"

	"github.com/thoas/go-funk"
)

type MutatingWebhook struct {
}

func (in *MutatingWebhook) Default(_ context.Context, apiRule *APIRule) error {
	if apiRule.Annotations == nil {
		apiRule.Annotations = map[string]string{}
	}

	if _, ok := apiRule.Annotations["gateway.kyma-project.io/original-version"]; !ok {
		// need to set original-version to v2alpha1	because we need to know version specifier for Busola list view three tabs
		apiRule.Annotations["gateway.kyma-project.io/original-version"] = "v2alpha1"
	}

	if apiRule.Annotations["gateway.kyma-project.io/original-version"] == "v1beta1" && !funk.IsEmpty(apiRule.Spec.Rules) {
		apiRule.Annotations["gateway.kyma-project.io/original-version"] = "v2alpha1"
		return nil
	}
	return nil
}
