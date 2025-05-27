package maester

import (
	"context"
	_ "embed"
	"fmt"

	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	"github.com/kyma-project/api-gateway/internal/reconciliations"
)

//go:embed peer_authentication.yaml
var peerAuthentication []byte

const peerAuthenticationName = "ory-oathkeeper-maester-metrics"

func reconcileOryOathkeeperPeerAuthentication(ctx context.Context, k8sClient client.Client, apiGatewayCR v1alpha1.APIGateway) error {
	ctrl.Log.Info("Reconciling Ory Config PeerAuthentication", "name", peerAuthenticationName, "Namespace", reconciliations.Namespace)

	if apiGatewayCR.IsInDeletion() {
		return deletePeerAuthentication(ctx, k8sClient, peerAuthenticationName, reconciliations.Namespace)
	}

	templateValues := make(map[string]string)
	templateValues["Name"] = peerAuthenticationName
	templateValues["Namespace"] = reconciliations.Namespace

	return reconciliations.ApplyResource(ctx, k8sClient, peerAuthentication, templateValues)
}

func deletePeerAuthentication(ctx context.Context, k8sClient client.Client, name, namespace string) error {
	ctrl.Log.Info("Deleting Oathkeeper Maester PeerAuthentication if it exists", "name", name, "Namespace", namespace)
	s := securityv1beta1.PeerAuthentication{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	err := k8sClient.Delete(ctx, &s)

	if err != nil && !k8serrors.IsNotFound(err) {
		return fmt.Errorf("failed to delete Oathkeeper Maester PeerAuthentication %s/%s: %w", namespace, name, err)
	}

	if k8serrors.IsNotFound(err) {
		ctrl.Log.Info("Skipped deletion of Oathkeeper Maester PeerAuthentication as it wasn't present", "name", name, "Namespace", namespace)
	} else {
		ctrl.Log.Info("Successfully deleted Oathkeeper Maester PeerAuthentication", "name", name, "Namespace", namespace)
	}

	return nil
}
