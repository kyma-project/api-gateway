package oathkeeper

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

//go:embed configmap_config.yaml
var configmapConfig []byte

const configMapName = "ory-oathkeeper-config"

func reconcileOryOathkeeperConfigConfigMap(ctx context.Context, k8sClient client.Client, apiGatewayCR v1alpha1.APIGateway) error {
	ctrl.Log.Info("Reconciling Ory Config ConfigMap", "name", configMapName, "Namespace", reconciliations.Namespace)

	if apiGatewayCR.IsInDeletion() {
		return deleteConfigmap(ctx, k8sClient, configMapName, reconciliations.Namespace)
	}

	isGardenerCluster := apiGatewayCR.Spec.Gardener

	domain := "local.kyma.dev"
	if isGardenerCluster {
		d, err := reconciliations.GetGardenerDomain(ctx, k8sClient)
		if err != nil {
			return err
		}
		domain = d
	}

	templateValues := make(map[string]string)
	templateValues["Name"] = configMapName
	templateValues["Namespace"] = reconciliations.Namespace
	templateValues["Domain"] = domain

	return reconciliations.ApplyResource(ctx, k8sClient, configmapConfig, templateValues)
}

func deleteConfigmap(ctx context.Context, k8sClient client.Client, name, namespace string) error {
	ctrl.Log.Info("Deleting Oathkeeper ConfigMap if it exists", "name", name, "Namespace", namespace)
	s := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	err := k8sClient.Delete(ctx, &s)

	if err != nil && !k8serrors.IsNotFound(err) {
		return fmt.Errorf("failed to delete Oathkeeper ConfigMap %s/%s: %v", namespace, name, err)
	}

	if k8serrors.IsNotFound(err) {
		ctrl.Log.Info("Skipped deletion of Oathkeeper ConfigMap as it wasn't present", "name", name, "Namespace", namespace)
	} else {
		ctrl.Log.Info("Successfully deleted Oathkeeper ConfigMap", "name", name, "Namespace", namespace)
	}

	return nil
}
