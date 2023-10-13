package maester

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

//go:embed cluster_role_binding.yaml
var clusterRoleBinding []byte

const roleBindingName = "oathkeeper-maester-role-binding"

func reconcileOryOathkeeperMaesterClusterRoleBinding(ctx context.Context, k8sClient client.Client, apiGatewayCR v1alpha1.APIGateway) error {
	ctrl.Log.Info("Reconciling Ory Oathkeeper Maester ClusterRoleBinding", "name", roleBindingName)

	if apiGatewayCR.IsInDeletion() {
		return deleteRoleBinding(ctx, k8sClient, roleBindingName)
	}

	templateValues := make(map[string]string)
	templateValues["Name"] = roleBindingName
	templateValues["ServiceAccountName"] = ServiceAccountName
	templateValues["ServiceAccountNamespace"] = reconciliations.Namespace
	templateValues["ClusterRoleName"] = roleName

	return reconciliations.ApplyResource(ctx, k8sClient, clusterRoleBinding, templateValues)
}

func deleteRoleBinding(ctx context.Context, k8sClient client.Client, name string) error {
	ctrl.Log.Info("Deleting Oathkeeper Maester ClusterRoleBinding if it exists", "name", name)
	s := rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	err := k8sClient.Delete(ctx, &s)

	if err != nil && !k8serrors.IsNotFound(err) {
		return fmt.Errorf("failed to delete Oathkeeper Maester ClusterRoleBinding %s: %v", name, err)
	}

	if k8serrors.IsNotFound(err) {
		ctrl.Log.Info("Skipped deletion of Oathkeeper Maester ClusterRoleBinding as it wasn't present", "name", name)
	} else {
		ctrl.Log.Info("Successfully deleted Oathkeeper Maester ClusterRoleBinding", "name", name)
	}

	return nil
}
