package v2alpha1

import (
	"context"
	"k8s.io/apimachinery/pkg/runtime"
)

type MutatingWebhook struct {
}

func (in *MutatingWebhook) Default(_ context.Context, obj runtime.Object) error {
	apiRule := obj.(*APIRule)

	if apiRule.Annotations == nil {
		apiRule.Annotations = map[string]string{}
	}

	if _, ok := apiRule.Annotations["gateway.kyma-project.io/original-version"]; !ok {
		// need to set original-version to v2alpha1	because we need to know version specifier for Busola list view three tabs
		apiRule.Annotations["gateway.kyma-project.io/original-version"] = "v2alpha1"
	}

	if apiRule.Annotations["gateway.kyma-project.io/original-version"] == "v1beta1" {
		apiRule.Annotations["gateway.kyma-project.io/original-version"] = "v2alpha1"
		return nil
	}
	return nil
}
