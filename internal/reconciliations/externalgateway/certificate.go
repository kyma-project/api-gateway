package externalgateway

import (
	"context"
	_ "embed"
	"fmt"

	certv1alpha1 "github.com/gardener/cert-management/pkg/apis/cert/v1alpha1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	externalv1alpha1 "github.com/kyma-project/api-gateway/apis/gateway/external/v1alpha1"
	"github.com/kyma-project/api-gateway/internal/reconciliations"
)

const (
	istioSystemNamespace = "istio-system"
)

//go:embed certificate.yaml
var certificateManifest []byte

// ReconcileCertificate creates or updates the Gardener Certificate for the external gateway
func ReconcileCertificate(ctx context.Context, k8sClient client.Client, external *externalv1alpha1.ExternalGateway, internalDomain string) error {
	certName := fmt.Sprintf("%s-cert", external.GatewayName())
	secretName := fmt.Sprintf("%s-tls", external.GatewayName())

	ctrl.Log.Info("Reconciling Certificate", "name", certName, "namespace", istioSystemNamespace, "domain", internalDomain)

	templateValues := map[string]string{
		"Name":                     certName,
		"Namespace":                istioSystemNamespace,
		"SecretName":               secretName,
		"Domain":                   internalDomain,
		"ExternalGatewayName":      external.Name,
		"ExternalGatewayNamespace": external.Namespace,
	}

	return reconciliations.ApplyResource(ctx, k8sClient, certificateManifest, templateValues)
}

// DeleteCertificate deletes the Certificate resource
func DeleteCertificate(ctx context.Context, k8sClient client.Client, gatewayName string) error {
	certName := fmt.Sprintf("%s-cert", gatewayName)

	ctrl.Log.Info("Deleting Certificate if it exists", "name", certName, "namespace", istioSystemNamespace)

	cert := &certv1alpha1.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      certName,
			Namespace: istioSystemNamespace,
		},
	}

	err := k8sClient.Delete(ctx, cert)
	if err != nil && !k8serrors.IsNotFound(err) {
		return fmt.Errorf("failed to delete Certificate %s/%s: %w", istioSystemNamespace, certName, err)
	}

	if k8serrors.IsNotFound(err) {
		ctrl.Log.Info("Skipped deletion of Certificate as it wasn't present", "name", certName)
	} else {
		ctrl.Log.Info("Successfully deleted Certificate", "name", certName)
	}

	return nil
}
