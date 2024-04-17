package certificate

import (
	"context"
	"crypto/x509"
	"fmt"
	"strings"
	"time"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	"github.com/kyma-project/api-gateway/controllers"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/cert"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	maxAge       = time.Hour * 24 * 90 // issue certificate with 90 days validity
	untilRenewal = time.Hour * 24 * 14 // renew certificate 14 days before expiration

	certificateName = "tls.crt"
	keyName         = "tls.key"

	secretNamespace = "kyma-system"
	secretName      = "api-gateway-webhook-certificate"
	serviceName     = "api-gateway-webhook-service"

	APIRuleCRDName = "apirules.gateway.kyma-project.io"
)

func NewCertificateReconciler(mgr manager.Manager, reconciliationInterval time.Duration) *CertificateReconciler {
	return &CertificateReconciler{
		Client:                 mgr.GetClient(),
		Scheme:                 mgr.GetScheme(),
		log:                    mgr.GetLogger().WithName("certificate-controller"),
		reconciliationInterval: reconciliationInterval,
	}
}

func (r *CertificateReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.log.Info("Received reconciliation request", "name", req.Name)

	certificateSecret := &corev1.Secret{}
	err := r.Client.Get(ctx, req.NamespacedName, certificateSecret)
	if err != nil {
		return ctrl.Result{}, err
	}

	r.log.Info("Reconciling Webhook Secret", "name", certificateSecret.Name)

	err = verifySecret(certificateSecret)
	if err == nil {
		r.log.Info("Secret certificate is still valid and does not need to be updated")
	} else {
		r.log.Info("Secret certificate is invalid", "verificationError", err.Error())
		certificate, err := createNewSecret(ctx, r.Client, certificateSecret)
		if err != nil {
			return ctrl.Result{RequeueAfter: 1 * time.Minute}, err
		}
		r.log.Info("New certificate created", "validFrom", certificate.NotBefore, "validUntil", certificate.NotAfter)
	}

	return ctrl.Result{RequeueAfter: r.reconciliationInterval}, nil
}

// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="apiextensions.k8s.io",resources=customresourcedefinitions,verbs=get;list;watch;create;update;patch;delete
func (r *CertificateReconciler) SetupWithManager(mgr ctrl.Manager, c controllers.RateLimiterConfig) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Secret{}).
		WithEventFilter(predicate.NewPredicateFuncs(func(o client.Object) bool { return o.GetName() == secretName && o.GetNamespace() == secretNamespace })).
		WithOptions(controller.Options{
			RateLimiter: controllers.NewRateLimiter(c),
		}).
		Complete(r)
}

func createNewSecret(ctx context.Context, client ctrlclient.Client, secret *corev1.Secret) (*x509.Certificate, error) {
	certificate, key, err := generateCertificate(serviceName, secret.Namespace)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate certificate")
	}

	newSecret := buildSecret(secret.Name, secret.Namespace, certificate, key)
	secret.Data = newSecret.Data

	if err := client.Update(ctx, secret); err != nil {
		return nil, errors.Wrap(err, "failed to update secret")
	}

	if err := updateCertificateInCRD(ctx, client, certificate); err != nil {
		return nil, errors.Wrap(err, "failed to update certificate into CRD")
	}

	parsedCertificates, err := cert.ParseCertsPEM(certificate)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse certificate")
	}

	return parsedCertificates[0], nil
}

func generateCertificate(serviceName, namespace string) ([]byte, []byte, error) {
	namespacedServiceName := strings.Join([]string{serviceName, namespace}, ".")
	commonName := strings.Join([]string{namespacedServiceName, "svc"}, ".")
	serviceHostname := fmt.Sprintf("%s.%s.svc.cluster.local", serviceName, namespace)
	altNames := []string{
		commonName,
		serviceName,
		namespacedServiceName,
		serviceHostname,
	}
	return generateSelfSignedCertificate(altNames[0], nil, altNames, maxAge)
}

func buildSecret(name, namespace string, certificate []byte, key []byte) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Data: map[string][]byte{
			certificateName: certificate,
			keyName:         key,
		},
	}
}

func updateCertificateInCRD(ctx context.Context, client ctrlclient.Client, certificate []byte) error {
	crd := &apiextensionsv1.CustomResourceDefinition{}
	err := client.Get(ctx, types.NamespacedName{Name: APIRuleCRDName}, crd)
	if err != nil {
		return errors.Wrap(err, "failed to get APIRule CRD")
	}

	if contains, reason := containsConversionWebhookClientConfig(crd); !contains {
		return errors.Errorf("can not add certificate into CRD: %s", reason)
	}

	crd.Spec.Conversion.Webhook.ClientConfig.CABundle = certificate
	err = client.Update(ctx, crd)
	if err != nil {
		return errors.Wrap(err, "failed to update CRD with new certificate")
	}
	return nil
}

func containsConversionWebhookClientConfig(crd *apiextensionsv1.CustomResourceDefinition) (bool, string) {
	if crd.Spec.Conversion == nil {
		return false, "conversion not found in APIRule CRD"
	}
	if crd.Spec.Conversion.Webhook == nil {
		return false, "conversion webhook not found in APIRule CRD"
	}
	if crd.Spec.Conversion.Webhook.ClientConfig == nil {
		return false, "client config for conversion webhook not found in APIRule CRD"
	}
	return true, ""
}
