package oathkeeper

import (
	"context"
	_ "embed"
	"fmt"
	"strconv"

	"github.com/avast/retry-go/v4"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	"github.com/kyma-project/api-gateway/internal/clusterconfig"
	"github.com/kyma-project/api-gateway/internal/reconciliations"
	"github.com/kyma-project/api-gateway/internal/reconciliations/oathkeeper/maester"
)

const (
	deploymentName = "ory-oathkeeper"
)

//go:embed deployment_light.yaml
var deploymentLight []byte

//go:embed deployment.yaml
var deployment []byte

func reconcileOathkeeperDeployment(ctx context.Context, k8sClient client.Client, apiGatewayCR v1alpha1.APIGateway) error {
	clusterSize, err := clusterconfig.EvaluateClusterSize(ctx, k8sClient)
	if err != nil {
		return err
	}

	ctrl.Log.Info("Reconciling Ory Oathkeeper Deployment", "Cluster size", clusterSize, "name", deploymentName, "Namespace", reconciliations.Namespace)

	if apiGatewayCR.IsInDeletion() {
		return deleteDeployment(ctx, k8sClient, deploymentName)
	}

	if clusterSize == clusterconfig.Evaluation {
		return reconcileDeployment(ctx, k8sClient, deploymentName, &deploymentLight)
	}
	return reconcileDeployment(ctx, k8sClient, deploymentName, &deployment)
}

func reconcileDeployment(ctx context.Context, k8sClient client.Client, name string, deploymentManifest *[]byte) error {
	ctrl.Log.Info("Reconciling Deployment", "name", name, "Namespace", reconciliations.Namespace)

	// As we have no replicas configured in the manifest for production because it is set by HPA, we read the replicas from the current configuration.
	// This way we avoid that the replicas are reset to 1 by the configuration in the manifest during reconciliation and then updated again by the HPA.
	replicas, err := getReplicasForDeployment(ctx, k8sClient)
	if err != nil {
		return err
	}

	templateValues := make(map[string]string)
	templateValues["Name"] = name
	templateValues["Namespace"] = reconciliations.Namespace
	templateValues["Replicas"] = replicas
	templateValues["ServiceAccountName"] = maester.ServiceAccountName

	return reconciliations.ApplyResource(ctx, k8sClient, *deploymentManifest, templateValues)
}

func deleteDeployment(ctx context.Context, k8sClient client.Client, name string) error {
	ctrl.Log.Info("Deleting Deployment if it exists", "name", name, "Namespace", reconciliations.Namespace)
	c := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: reconciliations.Namespace,
		},
	}
	err := k8sClient.Delete(ctx, &c)

	if err != nil && !k8serrors.IsNotFound(err) {
		return fmt.Errorf("failed to delete Deployment %s/%s: %w", reconciliations.Namespace, name, err)
	}

	if k8serrors.IsNotFound(err) {
		ctrl.Log.Info("Skipped deletion of Deployment as it wasn't present", "name", name, "Namespace", reconciliations.Namespace)
	} else {
		ctrl.Log.Info("Successfully deleted Deployment", "name", name, "Namespace", deploymentName)
	}

	return nil
}

func waitForOathkeeperDeploymentToBeReady(ctx context.Context, k8sClient client.Client, cfg RetryConfig) error {
	return retry.Do(func() error {
		var dep appsv1.Deployment
		err := k8sClient.Get(ctx, client.ObjectKey{
			Namespace: reconciliations.Namespace,
			Name:      deploymentName,
		}, &dep)

		if err != nil {
			return err
		}

		if hasMinimumReplicasAvailable(dep) {
			return nil
		}

		return fmt.Errorf("oathkeeper deployment does not have minimum replicas available. unavailable replicas  %d", dep.Status.UnavailableReplicas)
	}, retry.Attempts(cfg.Attempts), retry.Delay(cfg.Delay), retry.DelayType(retry.FixedDelay))
}

func getReplicasForDeployment(ctx context.Context, k8sClient client.Client) (string, error) {
	var dep appsv1.Deployment
	err := k8sClient.Get(ctx, client.ObjectKey{
		Namespace: reconciliations.Namespace,
		Name:      deploymentName,
	}, &dep)

	if k8serrors.IsNotFound(err) {
		return "", nil
	} else if err != nil {
		return "", err
	}

	if dep.Spec.Replicas == nil {
		return "", nil
	}

	return strconv.Itoa(int(*dep.Spec.Replicas)), nil
}

func hasMinimumReplicasAvailable(dep appsv1.Deployment) bool {
	for _, condition := range dep.Status.Conditions {
		if condition.Type == appsv1.DeploymentAvailable && condition.Status == corev1.ConditionTrue {
			return true
		}
	}

	return false
}
