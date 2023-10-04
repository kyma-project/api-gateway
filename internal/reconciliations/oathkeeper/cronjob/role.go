package cronjob

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	"github.com/kyma-project/api-gateway/internal/reconciliations"
	rbacv1 "k8s.io/api/rbac/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//go:embed role.yaml
var role []byte

const roleName = "ory-oathkeeper-keys-job-role"

func reconcileOryOathkeeperCronjobRole(ctx context.Context, k8sClient client.Client, apiGatewayCR v1alpha1.APIGateway) error {
	ctrl.Log.Info("Reconciling Ory Oathkeeper Cronjob Role", "name", roleName, "Namespace", reconciliations.Namespace)

	if apiGatewayCR.IsInDeletion() {
		return deleteRole(k8sClient, roleName, reconciliations.Namespace)
	}

	templateValues := make(map[string]string)
	templateValues["Name"] = roleName
	templateValues["Namespace"] = reconciliations.Namespace
	templateValues["OathkeeperName"] = oathkeeperName

	return reconciliations.ApplyResource(ctx, k8sClient, role, templateValues)
}

func deleteRole(k8sClient client.Client, name, namespace string) error {
	ctrl.Log.Info("Deleting Oathkeeper Cronjob Role if it exists", "name", name, "Namespace", namespace)
	s := rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	err := k8sClient.Delete(context.Background(), &s)

	if err != nil && !k8serrors.IsNotFound(err) {
		return fmt.Errorf("failed to delete Oathkeeper Role %s/%s: %v", namespace, name, err)
	}

	ctrl.Log.Info("Successfully deleted Oathkeeper Cronjob Role", "name", name, "Namespace", namespace)

	return nil
}
