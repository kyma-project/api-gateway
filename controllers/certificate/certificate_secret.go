package certificate

import (
	"context"
	"crypto/tls"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var currentCertificate *tls.Certificate

func GetCertificate(*tls.ClientHelloInfo) (*tls.Certificate, error) {
	if currentCertificate == nil {
		return nil, errors.New("certificate not available")
	}
	return currentCertificate, nil
}

func InitialiseCertificateSecret(ctx context.Context, client client.Client, log logr.Logger) error {
	log.Info("Initialising certificate secret", "namespace", secretNamespace, "name", secretName)

	secret := &corev1.Secret{}
	err := client.Get(ctx, types.NamespacedName{Namespace: secretNamespace, Name: secretName}, secret)
	if err != nil {
		if apierrs.IsNotFound(err) {
			log.Info("Certificate secret not found, creating a new one")
			deployment := &appsv1.Deployment{}
			err = client.Get(ctx, types.NamespacedName{Namespace: "kyma-system", Name: "api-gateway-controller-manager"}, deployment)
			if err != nil {
				return errors.Wrap(err, "failed to get api-gateway-controller-manager deployment")
			}
			certificate, key, err := generateNewCertificate(serviceName, secretNamespace)
			if err != nil {
				return errors.Wrap(err, "failed to generate certificate")
			}
			secret = &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: secretNamespace,
					Name:      secretName,
					Labels: map[string]string{
						"kyma-project.io/module": "api-gateway",
					},
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion: "apps/v1",
							Kind:       "Deployment",
							Name:       "api-gateway-controller-manager",
							UID:        deployment.UID,
						},
					},
				},
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
			if err := updateCertificateInMutatingWebhookConfigurationCR(ctx, client, certificate); err != nil {
				return errors.Wrap(err, "failed to update certificate into MutatingWebhookConfiguration CR")
			}
		} else {
			return errors.Wrap(err, "failed to get certificate secret")
		}
	} else {
		if certificate, ok := secret.Data[certificateName]; ok {
			if err := updateCertificateInMutatingWebhookConfigurationCR(ctx, client, certificate); err != nil {
				return errors.Wrap(err, "failed to update certificate into MutatingWebhookConfiguration CR during initialization")
			}
			log.Info("MutatingWebhookConfiguration updated with CABundle")
		}

		log.Info("Certificate secret found", "namespace", secretNamespace, "name", secretName)
	}

	if err = parseCertificateSecret(secret, log); err != nil {
		return errors.Wrap(err, "failed to get parse certificate secret")
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

	if err = parseCertificateSecret(secret, log); err != nil {
		return errors.Wrap(err, "failed to get parse certificate secret")
	}

	tlsCert, err := tls.X509KeyPair(secret.Data[certificateName], secret.Data[keyName])
	if err != nil {
		return errors.Wrap(err, "failed to load certificate key pair")
	}
	currentCertificate = &tlsCert

	return nil
}
