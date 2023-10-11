package cronjob

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	"github.com/kyma-project/api-gateway/internal/reconciliations"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const serviceAccountName = "ory-oathkeeper-keys-service-account"

func reconcileOryOathkeeperCronjobServiceAccount(ctx context.Context, k8sClient client.Client, _ v1alpha1.APIGateway) error {
	return deleteServiceAccount(ctx, k8sClient, serviceAccountName, reconciliations.Namespace)
}

func deleteServiceAccount(ctx context.Context, k8sClient client.Client, name, namespace string) error {
	ctrl.Log.Info("Deleting Oathkeeper Cronjob ServiceAccount if it exists", "name", name, "Namespace", namespace)
	s := corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	err := k8sClient.Delete(ctx, &s)

	if err != nil && !k8serrors.IsNotFound(err) {
		return fmt.Errorf("failed to delete Oathkeeper Cronjob ServiceAccount %s/%s: %v", namespace, name, err)
	}

	if k8serrors.IsNotFound(err) {
		ctrl.Log.Info("Skipped deletion of Oathkeeper Cronjob ServiceAccount it wasn't present", "name", name, "Namespace", namespace)
	} else {
		ctrl.Log.Info("Successfully deleted Oathkeeper Cronjob ServiceAccount", "name", name, "Namespace", namespace)
	}

	return nil
}
