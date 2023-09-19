package gateway

import (
	"context"
	_ "embed"
	"fmt"
	certv1alpha1 "github.com/gardener/cert-management/pkg/apis/cert/v1alpha1"
	"github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	kymaGatewayCertificateName = "kyma-gateway"
	// Per default the certificate secret needs to be in the same namespace as the gateway using it
	certificateDefaultNamespace = "istio-system"
	kymaGatewayCertSecretName   = "kyma-gateway-certs"
)

//go:embed certificate.yaml
var certificateManifest []byte

//go:embed certificate_secret_fallback.yaml
var nonGardenerCertificateSecretManifest []byte

func reconcileKymaGatewayCertificate(ctx context.Context, k8sClient client.Client, apiGatewayCR v1alpha1.APIGateway, domain string) error {

	isEnabled := isKymaGatewayEnabled(apiGatewayCR)
	ctrl.Log.Info("Reconciling Certificate", "KymaGatewayEnabled", isEnabled, "Name", kymaGatewayCertificateName, "Namespace", certificateDefaultNamespace)

	if !isEnabled {
		return deleteCertificate(k8sClient, kymaGatewayCertificateName)
	}

	return reconcileCertificate(ctx, k8sClient, kymaGatewayCertificateName, domain, kymaGatewayCertSecretName)
}

func reconcileCertificate(ctx context.Context, k8sClient client.Client, name, domain, certSecretName string) error {

	ctrl.Log.Info("Reconciling Certificate", "Name", name, "Namespace", certificateDefaultNamespace, "Domain", domain, "SecretName", certSecretName)
	templateValues := make(map[string]string)
	templateValues["Name"] = name
	templateValues["Namespace"] = certificateDefaultNamespace
	templateValues["Domain"] = domain
	templateValues["SecretName"] = certSecretName

	return reconcileResource(ctx, k8sClient, certificateManifest, templateValues)
}

func deleteCertificate(k8sClient client.Client, name string) error {
	ctrl.Log.Info("Deleting Certificate if it exists", "Name", name, "Namespace", certificateDefaultNamespace)
	c := certv1alpha1.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: certificateDefaultNamespace,
		},
	}
	err := k8sClient.Delete(context.TODO(), &c)

	if err != nil && !k8serrors.IsNotFound(err) {
		return fmt.Errorf("failed to delete DNSEntry %s/%s: %v", certificateDefaultNamespace, name, err)
	}

	if err == nil {
		ctrl.Log.Info("Successfully deleted DNSEntry", "Name", name, "Namespace", certificateDefaultNamespace)
	}

	return nil
}

func reconcileNonGardenerCertificateSecret(ctx context.Context, k8sClient client.Client) error {

	ctrl.Log.Info("Reconciling fallback certificate secret", "Name", kymaGatewayCertSecretName, "Namespace", certificateDefaultNamespace)
	templateValues := make(map[string]string)
	templateValues["Name"] = kymaGatewayCertSecretName
	templateValues["Namespace"] = certificateDefaultNamespace

	return reconcileResource(ctx, k8sClient, nonGardenerCertificateSecretManifest, templateValues)
}
