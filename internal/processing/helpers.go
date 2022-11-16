package processing

import (
	"fmt"
	gatewayv1alpha1 "github.com/kyma-incubator/api-gateway/api/v1alpha1"
	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	"github.com/kyma-incubator/api-gateway/internal/builders"
	k8sMeta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	//OwnerLabel .
	OwnerLabel = fmt.Sprintf("%s.%s", "apirule", gatewayv1beta1.GroupVersion.String())
	//OwnerLabelv1alpha1 .
	OwnerLabelv1alpha1 = fmt.Sprintf("%s.%s", "apirule", gatewayv1alpha1.GroupVersion.String())
)

func IsSecured(rule gatewayv1beta1.Rule) bool {
	if len(rule.Mutators) > 0 {
		return true
	}
	for _, strat := range rule.AccessStrategies {
		// TODO This considers an APIRule only as "not secured" when the strategy is "allow". Isn't "noop" also
		//  relevant for marking it as not secured?
		if strat.Name != "allow" {
			return true
		}
	}
	return false
}

func HasPathDuplicates(rules []gatewayv1beta1.Rule) bool {
	duplicates := map[string]bool{}
	for _, rule := range rules {
		if duplicates[rule.Path] {
			return true
		}
		duplicates[rule.Path] = true
	}

	return false
}

func GenerateOwnerRef(api *gatewayv1beta1.APIRule) k8sMeta.OwnerReference {
	return *builders.OwnerReference().
		Name(api.ObjectMeta.Name).
		APIVersion(api.TypeMeta.APIVersion).
		Kind(api.TypeMeta.Kind).
		UID(api.ObjectMeta.UID).
		Controller(true).
		Get()
}

func GetOwnerLabels(api *gatewayv1beta1.APIRule) map[string]string {
	labels := make(map[string]string)
	labels[OwnerLabelv1alpha1] = fmt.Sprintf("%s.%s", api.ObjectMeta.Name, api.ObjectMeta.Namespace)
	return labels
}
