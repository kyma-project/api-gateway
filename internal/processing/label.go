package processing

import (
	"fmt"

	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
)

const (
	ModuleLabelKey       = "kyma-project.io/module"
	K8sManagedByLabelKey = "app.kubernetes.io/managed-by"
	K8sComponentLabelKey = "app.kubernetes.io/component"
	K8sPartOfLabelKey    = "app.kubernetes.io/part-of"
	ApiGatewayLabelValue = "api-gateway"
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

func (l OwnerLabels) Labels() map[string]string {
	return map[string]string{
		OwnerLabelName:      l.Name,
		OwnerLabelNamespace: l.Namespace,
	}
}

func (o OwnerLabels) Owns(labels map[string]string) bool {
	name, nameOk := labels[OwnerLabelName]
	namespace, namespaceOk := labels[OwnerLabelNamespace]
	return nameOk && namespaceOk && name == o.Name && namespace == o.Namespace
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

// GetLegacyOwnerLabelsFromLabeler returns the legacy owner labels for any object that implements Labeler
func GetLegacyOwnerLabelsFromLabeler(l Labeler) map[string]string {
	labels := make(map[string]string)
	labels[LegacyOwnerLabel] = fmt.Sprintf("%s.%s", l.GetName(), l.GetNamespace())
	return labels
}

type Labeler interface {
	GetName() string
	GetNamespace() string
}

func GetOwnerLabels(l Labeler) OwnerLabels {
	return OwnerLabels{
		Name:      l.GetName(),
		Namespace: l.GetNamespace(),
	}
}
