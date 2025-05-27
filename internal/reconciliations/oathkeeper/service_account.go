package oathkeeper

import (
	"context"
	_ "embed"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	"github.com/kyma-project/api-gateway/internal/reconciliations"
)

//go:embed service_account.yaml
var serviceAccount []byte

const serviceAccountName = "ory-oathkeeper"

func reconcileOryOathkeeperServiceAccount(ctx context.Context, k8sClient client.Client, apiGatewayCR v1alpha1.APIGateway) error {
	ctrl.Log.Info("Reconciling Ory Oathkeeper ServiceAccount", "name", serviceAccountName, "Namespace", reconciliations.Namespace)

	if apiGatewayCR.IsInDeletion() {
		return deleteServiceAccount(ctx, k8sClient, serviceAccountName, reconciliations.Namespace)
	}

	templateValues := make(map[string]string)
	templateValues["Name"] = serviceAccountName
	templateValues["Namespace"] = reconciliations.Namespace

	return reconciliations.ApplyResource(ctx, k8sClient, serviceAccount, templateValues)
}

func deleteServiceAccount(ctx context.Context, k8sClient client.Client, name, namespace string) error {
	ctrl.Log.Info("Deleting Oathkeeper ServiceAccount if it exists", "name", name, "Namespace", namespace)
	s := corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	err := k8sClient.Delete(ctx, &s)

	if err != nil && !k8serrors.IsNotFound(err) {
		return fmt.Errorf("failed to delete Oathkeeper ServiceAccount %s/%s: %w", namespace, name, err)
	}

	if k8serrors.IsNotFound(err) {
		ctrl.Log.Info("Skipped deletion of Oathkeeper ServiceAccount as it wasn't present", "name", name, "Namespace", namespace)
	} else {
		ctrl.Log.Info("Successfully deleted Oathkeeper ServiceAccount", "name", name, "Namespace", namespace)
	}

	return nil
}
