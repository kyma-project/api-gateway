package gateway

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

	certv1alpha1 "github.com/gardener/cert-management/pkg/apis/cert/v1alpha1"
	"github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	"github.com/kyma-project/api-gateway/internal/reconciliations"
	"github.com/kyma-project/api-gateway/internal/version"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	kymaGatewayCertificateName = "kyma-tls-cert"
	// Istio IngressGateway requires the TLS secret to be present in the same namespace, that's why we have to use istio-system
	certificateDefaultNamespace = "istio-system"
	kymaGatewayCertSecretName   = "kyma-gateway-certs"
)

var (
	ErrCertificateError   = errors.New("kyma gateway certificate is in error state")
	ErrCertificatePending = errors.New("kyma gateway certificate is in pending state")
)

//go:embed certificate.yaml
var certificateManifest []byte

func reconcileKymaGatewayCertificate(ctx context.Context, k8sClient client.Client, apiGatewayCR v1alpha1.APIGateway, domain string) error {
	isEnabled := isKymaGatewayEnabled(apiGatewayCR)
	ctrl.Log.Info("Reconciling Certificate", "KymaGatewayEnabled", isEnabled, "name", kymaGatewayCertificateName, "namespace", certificateDefaultNamespace)

	if !isEnabled || apiGatewayCR.IsInDeletion() {
		return deleteCertificate(ctx, k8sClient, kymaGatewayCertificateName)
	}

	return reconcileCertificate(ctx, k8sClient, kymaGatewayCertificateName, domain, kymaGatewayCertSecretName)
}

// verifyKymaGatewayCertificateReady returns nil when the Kyma gateway Certificate is ready (or does not exist yet),
// errCertificateError when it failed, and errCertificateNotReady while it is still being issued.
func verifyKymaGatewayCertificateReady(ctx context.Context, k8sClient client.Client) error {
	var cert certv1alpha1.Certificate
	if err := k8sClient.Get(ctx, client.ObjectKey{Namespace: certificateDefaultNamespace, Name: kymaGatewayCertificateName}, &cert); err != nil {
		return client.IgnoreNotFound(err)
	}
	switch cert.Status.State {
	case certv1alpha1.StateReady:
		return nil
	case certv1alpha1.StateError:
		return ErrCertificateError
	default:
		return ErrCertificatePending
	}
}

func reconcileCertificate(ctx context.Context, k8sClient client.Client, name, domain, certSecretName string) error {
	ctrl.Log.Info("Reconciling Certificate", "name", name, "namespace", certificateDefaultNamespace, "domain", domain, "secretName", certSecretName)
	templateValues := make(map[string]string)
	templateValues["Name"] = name
	templateValues["Namespace"] = certificateDefaultNamespace
	templateValues["Domain"] = domain
	templateValues["SecretName"] = certSecretName
	templateValues["Version"] = version.GetModuleVersion()

	return reconciliations.ApplyResource(ctx, k8sClient, certificateManifest, templateValues)
}

func deleteCertificate(ctx context.Context, k8sClient client.Client, name string) error {
	ctrl.Log.Info("Deleting Certificate if it exists", "name", name, "namespace", certificateDefaultNamespace)
	c := certv1alpha1.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: certificateDefaultNamespace,
		},
	}
	err := k8sClient.Delete(ctx, &c)

	if err != nil && !k8serrors.IsNotFound(err) {
		return fmt.Errorf("failed to delete Certificate %s/%s: %v", certificateDefaultNamespace, name, err)
	}

	if k8serrors.IsNotFound(err) {
		ctrl.Log.Info("Skipped deletion of Certificate as it wasn't present", "name", name, "namespace", certificateDefaultNamespace)
	} else {
		ctrl.Log.Info("Successfully deleted Certificate", "name", name, "namespace", certificateDefaultNamespace)
	}

	return nil
}
