package patch

import (
	"context"
	"fmt"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	removeRequiredGatewayPatch = "/spec/versions/%d/schema/openAPIV3Schema/properties/spec/required"
	removeRequiredMethodsPatch = "/spec/versions/%d/schema/openAPIV3Schema/properties/spec/properties/rules/items/required"
)

func removeRequiredGateway(k8sClient client.Client, versionIndex int) error {
	patch := fmt.Sprintf(removeRequiredGatewayPatch, versionIndex)
	// Create a patch to remove the "gateway" required field
	patchData := []byte(fmt.Sprintf(`[{"op": "remove", "path": "%s"}]`, patch))

	// Apply the patch to the APIRule CRD
	return k8sClient.Patch(context.Background(), &apiextensionsv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "apirules.gateway.kyma-project.io",
		},
	}, client.RawPatch(types.JSONPatchType, patchData))
}

func removeRequiredMethods(k8sClient client.Client, versionIndex int) error {
	patch := fmt.Sprintf(removeRequiredMethodsPatch, versionIndex)
	// Create a patch to remove the "methods" required field
	patchData := []byte(fmt.Sprintf(`[{"op": "remove", "path": "%s"}]`, patch))

	// Apply the patch to the APIRule CRD
	return k8sClient.Patch(context.Background(), &apiextensionsv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "apirules.gateway.kyma-project.io",
		},
	}, client.RawPatch(types.JSONPatchType, patchData))
}

func Removev2alpha1VersionRequiredFields(k8sClient client.Client) error {
	var crd apiextensionsv1.CustomResourceDefinition
	if err := k8sClient.Get(context.Background(),
		types.NamespacedName{Name: "apirules.gateway.kyma-project.io"}, &crd); err != nil {
		return fmt.Errorf("failed to get APIRule CRD: %w", err)
	}

	var versionIndex int
	for i, version := range crd.Spec.Versions {
		if version.Name == "v2alpha1" {
			versionIndex = i
			break
		}
	}

	if err := removeRequiredGateway(k8sClient, versionIndex); err != nil {
		return fmt.Errorf("failed to remove required 'gateway' field: %w", err)
	}

	if err := removeRequiredMethods(k8sClient, versionIndex); err != nil {
		return fmt.Errorf("failed to remove required 'methods' field: %w", err)
	}

	return nil
}