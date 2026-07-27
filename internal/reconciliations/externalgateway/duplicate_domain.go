package externalgateway

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"

	externalv1alpha1 "github.com/kyma-project/api-gateway/apis/gateway/external/v1alpha1"
)

func CheckExternalDomainUnique(ctx context.Context, k8sClient client.Client, current *externalv1alpha1.ExternalGateway) error {
	if current.Spec.ExternalDomain == "" {
		return nil
	}
	list := &externalv1alpha1.ExternalGatewayList{}
	if err := k8sClient.List(ctx, list); err != nil {
		return err
	}
	for i := range list.Items {
		other := &list.Items[i]
		if other.Namespace == current.Namespace && other.Name == current.Name {
			continue
		}
		if other.Spec.ExternalDomain == current.Spec.ExternalDomain {
			return NewReasonedError(
				externalv1alpha1.ReasonExternalDomainConflict,
				"externalDomain %q is also used by ExternalGateway %s/%s",
				current.Spec.ExternalDomain, other.Namespace, other.Name,
			)
		}
	}
	return nil
}
