package gateway

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	"github.com/kyma-project/api-gateway/internal/clusterconfig"
	"github.com/kyma-project/api-gateway/internal/reconciliations"
	appsv1 "k8s.io/api/apps/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	deploymentName = "ory-oathkeeper"
)

//go:embed deployment_light.yaml
var deploymentLight []byte

func reconcileOathkeeperDeployment(ctx context.Context, k8sClient client.Client, apiGatewayCR v1alpha1.APIGateway, domain string) error {

	clusterSize, err := clusterconfig.EvaluateClusterSize(ctx, k8sClient)
	if err != nil {
		return err
	}

	ctrl.Log.Info("Reconciling Ory Oathkeeper Deployment", "Cluster size", clusterSize, "name", deploymentName, "namespace", namespace)

	if apiGatewayCR.IsInDeletion() {
		return deleteDeployment(k8sClient, deploymentName)
	}

	if clusterSize == clusterconfig.Evaluation {
		return reconcileDeployment(ctx, k8sClient, deploymentName, &deploymentLight)
	}
	return reconcileDeployment(ctx, k8sClient, deploymentName, &[]byte{})
}

func reconcileDeployment(ctx context.Context, k8sClient client.Client, name string, deploymentManifest *[]byte) error {

	ctrl.Log.Info("Reconciling Deployment", "name", name, "namespace", namespace)
	templateValues := make(map[string]string)
	templateValues["Name"] = name
	templateValues["Namespace"] = namespace

	return reconciliations.ApplyResource(ctx, k8sClient, *deploymentManifest, templateValues)
}

func deleteDeployment(k8sClient client.Client, name string) error {
	ctrl.Log.Info("Deleting Deployment if it exists", "name", name, "namespace", namespace)
	c := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	err := k8sClient.Delete(context.Background(), &c)

	if err != nil && !k8serrors.IsNotFound(err) {
		return fmt.Errorf("failed to delete Deployment %s/%s: %v", namespace, name, err)
	}

	if k8serrors.IsNotFound(err) {
		ctrl.Log.Info("Skipped deletion of Deployment as it wasn't present", "name", name, "namespace", namespace)
	} else {
		ctrl.Log.Info("Successfully deleted Deployment", "name", name, "namespace", deploymentName)
	}

	return nil
}
