package processing

import (
	"fmt"
	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
)

var (
	//OwnerLabel .
	OwnerLabel = fmt.Sprintf("%s.%s", "apirule", "gateway.kyma-project.io/v1beta1")
)

// GetOwnerLabelsV2alpha1 returns the owner labels for the given APIRule.
// The owner labels are still set to the old v1beta1 APIRule version.
// Do not switch the owner labels to the new APIRule version unless absolutely necessary!
// This has been done before, and it caused a lot of confusion and bugs.
// If the change for some reason has to be done, please remove the version from the OwnerLabel constant.
func GetOwnerLabelsV2alpha1(api *gatewayv2alpha1.APIRule) map[string]string {
	return map[string]string{
		OwnerLabel: fmt.Sprintf("%s.%s", api.Name, api.Namespace),
	}
}

func GetOwnerLabels(api *gatewayv2alpha1.APIRule) map[string]string {
	labels := make(map[string]string)
	labels[OwnerLabel] = fmt.Sprintf("%s.%s", api.Name, api.Namespace)
	return labels
}
