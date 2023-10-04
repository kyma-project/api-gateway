package cronjob

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	"github.com/kyma-project/api-gateway/internal/reconciliations"
	schedulingv1 "k8s.io/api/batch/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//go:embed cronjob.yaml
var cronjob []byte

const (
	cronjobName    = "oathkeeper-jwks-rotator"
	oathkeeperName = "ory-oathkeeper"
)

func reconcileOryOathkeeperCronjob(ctx context.Context, k8sClient client.Client, apiGatewayCR v1alpha1.APIGateway) error {
	ctrl.Log.Info("Reconciling Ory Oathkeeper Cronjob", "name", cronjobName, "Namespace", reconciliations.Namespace)

	if apiGatewayCR.IsInDeletion() {
		return deleteCronjob(k8sClient, roleName, reconciliations.Namespace)
	}

	templateValues := make(map[string]string)
	templateValues["Name"] = roleName
	templateValues["Namespace"] = reconciliations.Namespace
	templateValues["OathkeeperName"] = oathkeeperName
	templateValues["ServiceAccountName"] = serviceAccountName

	return reconciliations.ApplyResource(ctx, k8sClient, cronjob, templateValues)
}

func deleteCronjob(k8sClient client.Client, name, namespace string) error {
	ctrl.Log.Info("Deleting Oathkeeper Cronjob if it exists", "name", name, "Namespace", namespace)
	s := schedulingv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	err := k8sClient.Delete(context.Background(), &s)

	if err != nil && !k8serrors.IsNotFound(err) {
		return fmt.Errorf("failed to delete Oathkeeper Cronjob %s/%s: %v", namespace, name, err)
	}

	ctrl.Log.Info("Successfully deleted Oathkeeper Cronjob", "name", name, "Namespace", namespace)

	return nil
}

func ReconcileCronjob(ctx context.Context, k8sClient client.Client, apiGatewayCR v1alpha1.APIGateway) error {
	return errors.Join(
		reconcileOryOathkeeperCronjobServiceAccount(ctx, k8sClient, apiGatewayCR),
		reconcileOryOathkeeperCronjobRole(ctx, k8sClient, apiGatewayCR),
		reconcileOryOathkeeperCronjobRoleBinding(ctx, k8sClient, apiGatewayCR),
		reconcileOryOathkeeperCronjob(ctx, k8sClient, apiGatewayCR),
	)
}
