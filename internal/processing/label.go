package processing

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
)

var (
	//LegacyOwnerLabel .
	// deprecated: use OwnerLabelName and OwnerLabelNamespace instead.
	LegacyOwnerLabel = fmt.Sprintf("%s.%s", "apirule", gatewayv1beta1.GroupVersion.String())

	// OwnerLabelName is the label key used to identify the owner name of a resource.
	OwnerLabelName = "apirule.gateway.kyma-project.io/name"
	// OwnerLabelNamespace is the label key used to identify the owner namespace of a resource.
	OwnerLabelNamespace = "apirule.gateway.kyma-project.io/namespace"
)

type OwnerLabels struct {
	Name      string
	Namespace string
}

// GetOwnerLabelsV2alpha1 returns the owner labels for the given APIRule.
// The owner labels are still set to the old v1beta1 APIRule version.
// Do not switch the owner labels to the new APIRule version unless absolutely necessary!
// This has been done before, and it caused a lot of confusion and bugs.
// If the change for some reason has to be done, please remove the version from the LegacyOwnerLabel constant.
// deprecated: use GetLegacyOwnerLabels instead.
func GetOwnerLabelsV2alpha1(api *gatewayv2alpha1.APIRule) map[string]string {
	return map[string]string{
		LegacyOwnerLabel: fmt.Sprintf("%s.%s", api.Name, api.Namespace),
	}
}

func GetLegacyOwnerLabels(api *gatewayv1beta1.APIRule) map[string]string {
	labels := make(map[string]string)
	labels[LegacyOwnerLabel] = fmt.Sprintf("%s.%s", api.Name, api.Namespace)
	return labels
}

type Labeler interface {
	GetObjectMeta() metav1.ObjectMeta
}

func GetOwnerLabels(l Labeler) OwnerLabels {
	return OwnerLabels{
		Name:      l.GetObjectMeta().Name,
		Namespace: l.GetObjectMeta().Namespace,
	}
}
