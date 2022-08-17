package processing

import (
	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	"github.com/kyma-incubator/api-gateway/internal/builders"
	k8sMeta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func isSecured(rule gatewayv1beta1.Rule) bool {
	if len(rule.Mutators) > 0 {
		return true
	}
	for _, strat := range rule.AccessStrategies {
		if strat.Name != "allow" {
			return true
		}
	}
	return false
}

func generateOwnerRef(api *gatewayv1beta1.APIRule) k8sMeta.OwnerReference {
	return *builders.OwnerReference().
		Name(api.ObjectMeta.Name).
		APIVersion(api.TypeMeta.APIVersion).
		Kind(api.TypeMeta.Kind).
		UID(api.ObjectMeta.UID).
		Controller(true).
		Get()
}
