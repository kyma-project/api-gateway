package oathkeeper

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	"github.com/kyma-project/api-gateway/internal/clusterconfig"
	"github.com/kyma-project/api-gateway/internal/reconciliations"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	hpaName = "ory-oathkeeper"
)

//go:embed hpa.yaml
var hpa []byte

func reconcileOathkeeperHPA(ctx context.Context, k8sClient client.Client, apiGatewayCR v1alpha1.APIGateway) error {

	clusterSize, err := clusterconfig.EvaluateClusterSize(ctx, k8sClient)
	if err != nil {
		return err
	}

	ctrl.Log.Info("Reconciling Ory Oathkeeper HPA", "Cluster size", clusterSize, "name", hpaName, "Namespace", reconciliations.Namespace)

	if clusterSize == clusterconfig.Evaluation || apiGatewayCR.IsInDeletion() {
		return deleteHPA(ctx, k8sClient, hpaName)
	}

	templateValues := make(map[string]string)
	templateValues["Name"] = hpaName
	templateValues["Namespace"] = reconciliations.Namespace
	templateValues["DeploymentName"] = deploymentName

	return reconciliations.ApplyResource(ctx, k8sClient, hpa, templateValues)
}

func deleteHPA(ctx context.Context, k8sClient client.Client, name string) error {
	ctrl.Log.Info("Deleting HPA if it exists", "name", name, "Namespace", reconciliations.Namespace)
	c := autoscalingv2.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: reconciliations.Namespace,
		},
	}
	err := k8sClient.Delete(ctx, &c)

	if err != nil && !k8serrors.IsNotFound(err) {
		return fmt.Errorf("failed to delete HPA %s/%s: %v", reconciliations.Namespace, name, err)
	}

	if k8serrors.IsNotFound(err) {
		ctrl.Log.Info("Skipped deletion of HPA as it wasn't present", "name", name, "Namespace", reconciliations.Namespace)
	} else {
		ctrl.Log.Info("Successfully deleted HPA", "name", name, "Namespace", deploymentName)
	}

	return nil
}
