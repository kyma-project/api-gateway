package operator

import (
	"context"
	_ "embed"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	disableWebhookJSONPatch = `[{"op": "remove", "path": "/spec/conversion"}]`
)

func DisableAPIRuleWebhook(k8sClient client.Client) error {
	var crd apiextensionsv1.CustomResourceDefinition
	if err := k8sClient.Get(context.Background(), types.NamespacedName{Name: "apirules.gateway.kyma-project.io"}, &crd); err != nil {
		return err
	}

	return k8sClient.Patch(
		context.Background(),
		&crd,
		client.RawPatch(types.JSONPatchType, []byte(disableWebhookJSONPatch)),
	)
}
