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

const roleBindingName = "ory-oathkeeper-keys-job-role-binding"

func reconcileOryOathkeeperCronjobRoleBinding(_ context.Context, k8sClient client.Client, _ v1alpha1.APIGateway) error {
	return deleteRoleBinding(k8sClient, roleBindingName, reconciliations.Namespace)
}

func deleteRoleBinding(k8sClient client.Client, name, namespace string) error {
	ctrl.Log.Info("Deleting Oathkeeper Cronjob RoleBinding if it exists", "name", name, "Namespace", namespace)
	s := rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	err := k8sClient.Delete(context.Background(), &s)

	if err != nil && !k8serrors.IsNotFound(err) {
		return fmt.Errorf("failed to delete Oathkeeper RoleBinding %s/%s: %v", namespace, name, err)
	}

	if k8serrors.IsNotFound(err) {
		ctrl.Log.Info("Skipped deletion of Oathkeeper Cronjob RoleBindingas it wasn't present", "name", name, "Namespace", namespace)
	} else {
		ctrl.Log.Info("Successfully deleted Oathkeeper Cronjob RoleBinding", "name", name, "Namespace", namespace)
	}

	return nil
}
