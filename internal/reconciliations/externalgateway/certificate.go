package externalgateway

import (
	"context"
	_ "embed"
	"fmt"

	certv1alpha1 "github.com/gardener/cert-management/pkg/apis/cert/v1alpha1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	externalv1alpha1 "github.com/kyma-project/api-gateway/apis/gateway/external/v1alpha1"
)

const (
	istioSystemNamespace = "istio-system"
	privateKeySize4096   = 4096
)

// ReconcileCertificate creates or updates the Gardener Certificate for the external gateway
func ReconcileCertificate(ctx context.Context, k8sClient client.Client, external *externalv1alpha1.ExternalGateway, internalDomain string) error {
	certName := fmt.Sprintf("%s-cert", external.GatewayName())
	secretName := fmt.Sprintf("%s-tls", external.GatewayName())

	ctrl.Log.Info("Reconciling Certificate", "name", certName, "namespace", istioSystemNamespace, "domain", internalDomain)

	cert := &certv1alpha1.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      certName,
			Namespace: istioSystemNamespace,
		},
	}

	operation, err := controllerutil.CreateOrUpdate(ctx, k8sClient, cert, func() error {
		cert.Labels = GetStandardLabels(external)
		cert.Spec = certv1alpha1.CertificateSpec{
			CommonName: &internalDomain,
			SecretName: &secretName,
			IssuerRef: &certv1alpha1.IssuerRef{
				Name: "garden",
			},
			PrivateKey: &certv1alpha1.CertificatePrivateKey{
				Size: ptr.To(certv1alpha1.PrivateKeySize(privateKeySize4096)),
			},
		}
		return nil
	})
	if err != nil {
		ctrl.Log.Error(err, "Failed to create or update Certificate", "name", certName, "namespace", istioSystemNamespace, "error", err)
		return fmt.Errorf("failed to create or update Certificate %s/%s: %w", istioSystemNamespace, certName, err)
	}

	ctrl.Log.Info("Successfully reconciled Certificate", "name", certName, "namespace", istioSystemNamespace, "operation", operation)
	return nil
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
