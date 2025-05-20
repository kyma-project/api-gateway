package oathkeeper

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	"github.com/kyma-project/api-gateway/internal/reconciliations"
	v1 "k8s.io/api/policy/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//go:embed pod_disruption_budget.yaml
var pdb []byte

const pdbName = "ory-oathkeeper"

func reconcileOathkeeperPdb(ctx context.Context, k8sClient client.Client, apiGatewayCR v1alpha1.APIGateway) error {
	ctrl.Log.Info("Reconciling Oathkeeper Pod Disruption Budget", "name", pdbName, "Namespace", reconciliations.Namespace)

	if apiGatewayCR.IsInDeletion() {
		return deletePdb(ctx, k8sClient, pdbName, reconciliations.Namespace)
	}

	templateValues := make(map[string]string)
	templateValues["Name"] = pdbName
	templateValues["Namespace"] = reconciliations.Namespace

	return reconciliations.ApplyResource(ctx, k8sClient, pdb, templateValues)
}

func deletePdb(ctx context.Context, k8sClient client.Client, name, namespace string) error {
	ctrl.Log.Info("Deleting Oathkeeper Pod Disruption Budget if it exists", "name", name, "Namespace", namespace)
	s := v1.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	err := k8sClient.Delete(ctx, &s)

	if err != nil && !k8serrors.IsNotFound(err) {
		return fmt.Errorf("failed to delete Oathkeeper Pod Disruption Budget %s/%s: %v", namespace, name, err)
	}

	if k8serrors.IsNotFound(err) {
		ctrl.Log.Info("Skipped deletion of Oathkeeper Pod Disruption Budget as it wasn't present", "name", name, "Namespace", namespace)
	} else {
		ctrl.Log.Info("Successfully deleted Oathkeeper Pod Disruption Budget", "name", name, "Namespace", namespace)
	}

	return nil
}
