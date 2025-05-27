package oathkeeper

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	"github.com/kyma-project/api-gateway/internal/reconciliations"
)

//go:embed service-api.yaml
var serviceApi []byte

//go:embed service-metrics.yaml
var serviceMetrics []byte

//go:embed service-proxy.yaml
var serviceProxy []byte

const (
	apiServiceName     = "ory-oathkeeper-api"
	proxyServiceName   = "ory-oathkeeper-proxy"
	metricsServiceName = "ory-oathkeeper-maester-metrics"
)

func reconcileOryOathkeeperServices(ctx context.Context, k8sClient client.Client, apiGatewayCR v1alpha1.APIGateway) error {
	ctrl.Log.Info("Reconciling Ory Oathkeepers Services", "names", []string{apiServiceName, proxyServiceName, metricsServiceName}, "Namespace", reconciliations.Namespace)

	if apiGatewayCR.IsInDeletion() {
		return errors.Join(
			deleteService(ctx, k8sClient, apiServiceName, reconciliations.Namespace),
			deleteService(ctx, k8sClient, metricsServiceName, reconciliations.Namespace),
			deleteService(ctx, k8sClient, proxyServiceName, reconciliations.Namespace),
		)
	}

	templateValuesApi := make(map[string]string)
	templateValuesApi["Name"] = apiServiceName
	templateValuesApi["Namespace"] = reconciliations.Namespace

	templateValuesMetrics := make(map[string]string)
	templateValuesMetrics["Name"] = metricsServiceName
	templateValuesMetrics["Namespace"] = reconciliations.Namespace

	templateValuesProxy := make(map[string]string)
	templateValuesProxy["Name"] = proxyServiceName
	templateValuesProxy["Namespace"] = reconciliations.Namespace

	return errors.Join(
		reconciliations.ApplyResource(ctx, k8sClient, serviceApi, templateValuesApi),
		reconciliations.ApplyResource(ctx, k8sClient, serviceMetrics, templateValuesMetrics),
		reconciliations.ApplyResource(ctx, k8sClient, serviceProxy, templateValuesProxy),
	)
}

func deleteService(ctx context.Context, k8sClient client.Client, name, namespace string) error {
	ctrl.Log.Info("Deleting Oathkeeper Service if it exists", "name", name, "Namespace", namespace)
	s := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	err := k8sClient.Delete(ctx, &s)

	if err != nil && !k8serrors.IsNotFound(err) {
		return fmt.Errorf("failed to delete Oathkeeper Service %s/%s: %w", namespace, name, err)
	}

	if k8serrors.IsNotFound(err) {
		ctrl.Log.Info("Skipped deletion of Oathkeeper Service as it wasn't present", "name", name, "Namespace", namespace)
	} else {
		ctrl.Log.Info("Successfully deleted Oathkeeper Service", "name", name, "Namespace", namespace)
	}

	return nil
}
