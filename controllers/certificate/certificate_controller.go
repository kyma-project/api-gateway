package certificate

import (
	"context"
	"crypto/tls"
	"fmt"
	"strings"
	"time"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	"github.com/go-logr/logr"
	"github.com/kyma-project/api-gateway/controllers"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/cert"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	reconciliationInterval = time.Hour * 1       // reconciliation interval of 1 hour
	maxAge                 = time.Hour * 24 * 90 // issue certificate with 90 days validity
	untilRenewal           = time.Hour * 24 * 14 // renew certificate 14 days before expiration

	certificateName = "tls.crt"
	keyName         = "tls.key"

	secretNamespace = "kyma-system"
	secretName      = "api-gateway-webhook-certificate"
	serviceName     = "api-gateway-webhook-service"

	apiRuleCRDName = "apirules.gateway.kyma-project.io"
)

var currentCertifate *tls.Certificate
var GetCertificate = func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
	if currentCertifate == nil {
		return nil, errors.New("certificate not available")
	}
	return currentCertifate, nil
}

func InitialiseCertificateSecret(ctx context.Context, client client.Client, log logr.Logger) error {
	log.Info("Initialising certficate secret", "namespace", secretNamespace, "name", secretName)

	secret := &corev1.Secret{}
	err := client.Get(ctx, types.NamespacedName{Namespace: secretNamespace, Name: secretName}, secret)
	if err != nil {
		if apierrs.IsNotFound(err) {
			log.Info("Certificate secret not found, creating a new one")
			certificate, key, err := generateNewCertificate(serviceName, secretNamespace)
			if err != nil {
				return errors.Wrap(err, "failed to generate certificate")
			}
			secret = &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{Namespace: secretNamespace, Name: secretName},
				Data: map[string][]byte{
					certificateName: certificate,
					keyName:         key,
				},
				Type: corev1.SecretTypeOpaque,
			}
			if err := client.Create(ctx, secret); err != nil {
				return errors.Wrap(err, "failed to create secret")
			}
			if err := updateCertificateInCRD(ctx, client, certificate); err != nil {
				return errors.Wrap(err, "failed to update certificate into CRD")
			}
		} else {
			return errors.Wrap(err, "failed to get certificate secret")
		}
	} else {
		log.Info("Certificate secret found", "namespace", secretNamespace, "name", secretName)
	}

	if err = parseCertificateSecret(secret, log); err != nil {
		return errors.Wrap(err, "failed to print certificate secret validity")
	}

	return nil
}

func ReadCertificateSecret(ctx context.Context, client client.Client, log logr.Logger) error {
	log.Info("Reading certficate secret", "namespace", secretNamespace, "name", secretName)

	secret := &corev1.Secret{}
	err := client.Get(ctx, types.NamespacedName{Namespace: secretNamespace, Name: secretName}, secret)
	if err != nil {
		return errors.Wrap(err, "failed to get certificate secret")
	}

	parseCertificateSecret(secret, log)

	tlsCert, err := tls.X509KeyPair(secret.Data[certificateName], secret.Data[keyName])
	if err != nil {
		return errors.Wrap(err, "failed to load certificate key pair")
	}
	currentCertifate = &tlsCert

	return nil
}

func NewCertificateReconciler(mgr manager.Manager) *Reconciler {
	return &Reconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		Log:    mgr.GetLogger().WithName("certificate-controller"),
	}
}

// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="apiextensions.k8s.io",resources=customresourcedefinitions,verbs=get;list;watch;create;update;patch;delete

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.Log.Info("Received reconciliation request", "namespace", req.Namespace, "name", req.Name)

	secret := &corev1.Secret{}
	err := r.Client.Get(ctx, types.NamespacedName{Namespace: secretNamespace, Name: secretName}, secret)
	if err != nil {
		return ctrl.Result{Requeue: false}, err
	}

	requeue, err := verifyCertificateSecret(ctx, r.Client, secret, r.Log)
	if err != nil {
		return ctrl.Result{Requeue: requeue}, err
	}

	tlsCert, err := tls.X509KeyPair(secret.Data[certificateName], secret.Data[keyName])
	if err != nil {
		return ctrl.Result{Requeue: false}, err
	}
	currentCertifate = &tlsCert

	return ctrl.Result{RequeueAfter: reconciliationInterval}, nil
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, c controllers.RateLimiterConfig) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Secret{}).
		WithEventFilter(predicate.NewPredicateFuncs(func(o client.Object) bool { return o.GetName() == secretName && o.GetNamespace() == secretNamespace })).
		WithOptions(controller.Options{
			RateLimiter: controllers.NewRateLimiter(c),
		}).
		Complete(r)
}

func verifyCertificateSecret(ctx context.Context, client client.Client, secret *corev1.Secret, log logr.Logger) (bool, error) {
	log.Info("Verifying certficate secret", "namespace", secretNamespace, "name", secretName)

	err := verifySecret(secret)
	if err == nil {
		log.Info("Certificate is still valid and does not need to be updated")
	} else {
		log.Info("Certificate verification did not succeed", "error", err.Error())
		err := generateNewCertificateSecret(ctx, client, secret)
		if err != nil {
			return true, err
		}
		parseCertificateSecret(secret, log)
	}

	return false, nil
}

func generateNewCertificateSecret(ctx context.Context, client client.Client, secret *corev1.Secret) error {
	certificate, key, err := generateNewCertificate(serviceName, secret.Namespace)
	if err != nil {
		return errors.Wrap(err, "failed to generate certificate")
	}

	mergeFrom := ctrlclient.StrategicMergeFrom(secret.DeepCopy())

	if secret.Data == nil {
		secret.Data = make(map[string][]byte)
	}

	secret.Data[certificateName] = certificate
	secret.Data[keyName] = key

	if err := client.Patch(ctx, secret, mergeFrom); err != nil {
		return errors.Wrap(err, "failed to patch secret")
	}

	if err := updateCertificateInCRD(ctx, client, certificate); err != nil {
		return errors.Wrap(err, "failed to update certificate into CRD")
	}

	return nil
}

func generateNewCertificate(serviceName, namespace string) ([]byte, []byte, error) {
	namespacedServiceName := strings.Join([]string{serviceName, namespace}, ".")
	commonName := strings.Join([]string{namespacedServiceName, "svc"}, ".")
	return GenerateSelfSignedCertificate(commonName, nil, []string{}, maxAge)
}

func updateCertificateInCRD(ctx context.Context, client client.Client, certificate []byte) error {
	crd := &apiextensionsv1.CustomResourceDefinition{}
	err := client.Get(ctx, types.NamespacedName{Name: apiRuleCRDName}, crd)
	if err != nil {
		return errors.Wrap(err, "failed to get APIRule CRD")
	}

	if contains, reason := containsConversionWebhookClientConfig(crd); !contains {
		return errors.Errorf("can not add certificate into CRD: %s", reason)
	}

	mergeFrom := ctrlclient.StrategicMergeFrom(crd.DeepCopy())
	crd.Spec.Conversion.Webhook.ClientConfig.CABundle = certificate

	if err := client.Patch(ctx, crd, mergeFrom); err != nil {
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

func parseCertificateSecret(secret *corev1.Secret, log logr.Logger) error {
	if !hasRequiredKeys(secret.Data, []string{certificateName, keyName}) {
		return fmt.Errorf("secret does not have required keys: %s, %s", certificateName, keyName)
	}

	certs, err := cert.ParseCertsPEM(secret.Data[certificateName])
	if err != nil || len(certs) == 0 {
		return errors.Wrap(err, "failed to parse certificate")
	}

	log.Info("Certificate validity", "validFrom", certs[0].NotBefore, "validUntil", certs[0].NotAfter)
	return nil
}
