package maester

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	"github.com/kyma-project/api-gateway/internal/reconciliations"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//go:embed service_account.yaml
var serviceAccount []byte

const ServiceAccountName = "oathkeeper-maester-account"

func reconcileOryOathkeeperMaesterServiceAccount(ctx context.Context, k8sClient client.Client, apiGatewayCR v1alpha1.APIGateway) error {
	ctrl.Log.Info("Reconciling Ory Maester ServiceAccount", "name", ServiceAccountName, "Namespace", reconciliations.Namespace)

	if apiGatewayCR.IsInDeletion() {
		return deleteServiceAccount(ctx, k8sClient, ServiceAccountName, reconciliations.Namespace)
	}

	templateValues := make(map[string]string)
	templateValues["Name"] = ServiceAccountName
	templateValues["Namespace"] = reconciliations.Namespace

	err := reconciliations.ApplyResource(ctx, k8sClient, serviceAccount, templateValues)
	if err != nil {
		return err
	}

	return waitForServiceAccount(ctx, k8sClient)
}

func deleteServiceAccount(ctx context.Context, k8sClient client.Client, name, namespace string) error {
	ctrl.Log.Info("Deleting Oathkeeper Maester ServiceAccount if it exists", "name", name, "Namespace", namespace)
	s := corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	err := k8sClient.Delete(ctx, &s)

	if err != nil && !k8serrors.IsNotFound(err) {
		return fmt.Errorf("failed to delete Oathkeeper Maester ServiceAccount %s/%s: %v", namespace, name, err)
	}

	if k8serrors.IsNotFound(err) {
		ctrl.Log.Info("Skipped deletion of Oathkeeper Maester ServiceAccount as it wasn't present", "name", name, "Namespace", namespace)
	} else {
		ctrl.Log.Info("Successfully deleted Oathkeeper Maester ServiceAccount", "name", name, "Namespace", namespace)
	}

	return nil
}

func waitForServiceAccount(ctx context.Context, k8sClient client.Client) error {
	return retry.Do(func() error {
		serviceAccount := corev1.ServiceAccount{}
		return k8sClient.Get(ctx, types.NamespacedName{Name: ServiceAccountName, Namespace: reconciliations.Namespace}, &serviceAccount)
	}, retry.Attempts(10), retry.Delay(2*time.Second), retry.DelayType(retry.FixedDelay))
}
