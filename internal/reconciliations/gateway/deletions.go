package gateway

import (
	"context"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func reconcilev1beta1andv2alpha1UIDeletion(ctx context.Context, k8sClient client.Client) error {
	configMapv1beta1 := v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "api-gateway-apirule-v1beta1-ui.operator.kyma-project.io",
			Namespace: "kyma-system",
		},
	}

	configMapv2alpha1 := v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "api-gateway-apirule-v2alpha1-ui.operator.kyma-project.io",
			Namespace: "kyma-system",
		},
	}

	err := client.IgnoreNotFound(k8sClient.Delete(ctx, &configMapv1beta1))
	if err != nil {
		return err
	}

	err = client.IgnoreNotFound(k8sClient.Delete(ctx, &configMapv2alpha1))
	if err != nil {
		return err
	}

	return nil
}
