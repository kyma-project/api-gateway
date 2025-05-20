package maester

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	"github.com/kyma-project/api-gateway/internal/reconciliations"
	rbacv1 "k8s.io/api/rbac/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//go:embed cluster_role.yaml
var clusterRole []byte

const roleName = "oathkeeper-maester-role"

func reconcileOryOathkeeperMaesterClusterRole(ctx context.Context, k8sClient client.Client, apiGatewayCR v1alpha1.APIGateway) error {
	ctrl.Log.Info("Reconciling Ory Oathkeeper Maester ClusterRole", "name", roleName)

	if apiGatewayCR.IsInDeletion() {
		return deleteClusterRole(ctx, k8sClient, roleName)
	}

	templateValues := make(map[string]string)
	templateValues["Name"] = roleName

	err := reconciliations.ApplyResource(ctx, k8sClient, clusterRole, templateValues)
	if err != nil {
		return err
	}

	return waitForClusterRole(ctx, k8sClient)
}

func deleteClusterRole(ctx context.Context, k8sClient client.Client, name string) error {
	ctrl.Log.Info("Deleting Oathkeeper Maester ClusterRole if it exists", "name", name)
	s := rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	err := k8sClient.Delete(ctx, &s)

	if err != nil && !k8serrors.IsNotFound(err) {
		return fmt.Errorf("failed to delete Oathkeeper Maester ClusterRole %s: %v", name, err)
	}

	if k8serrors.IsNotFound(err) {
		ctrl.Log.Info("Skipped deletion of Oathkeeper Maester ClusterRole as it wasn't present", "name", name)
	} else {
		ctrl.Log.Info("Successfully deleted Oathkeeper Maester ClusterRole", "name", name)
	}

	return nil
}

func waitForClusterRole(ctx context.Context, k8sClient client.Client) error {
	return retry.Do(func() error {
		var clusterRole rbacv1.ClusterRole
		return k8sClient.Get(ctx, types.NamespacedName{Name: roleName}, &clusterRole)
	}, retry.Attempts(10), retry.Delay(2*time.Second), retry.DelayType(retry.FixedDelay))
}
