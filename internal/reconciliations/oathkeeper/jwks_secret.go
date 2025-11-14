package oathkeeper

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/go-jose/go-jose/v3"
	"github.com/google/uuid"
	"github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	"github.com/kyma-project/api-gateway/internal/reconciliations"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	jwksAlg  = "RS256"
	jwksBits = 3072
)

const secretName = "ory-oathkeeper-jwks-secret"

func reconcileOryJWKSSecret(ctx context.Context, k8sClient client.Client, apiGatewayCR v1alpha1.APIGateway) error {
	ctrl.Log.Info("Reconciling Ory JWKS Secret", "name", secretName, "Namespace", reconciliations.Namespace)

	if apiGatewayCR.IsInDeletion() {
		return deleteSecret(ctx, k8sClient, secretName, reconciliations.Namespace)
	}

	data, err := generateJWKS()
	if err != nil {
		return err
	}

	secret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: secretName, Namespace: reconciliations.Namespace},
		Data: map[string][]byte{
			"jwks.json": data,
		},
	}

	secretMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&secret)
	if err != nil {
		return err
	}
	secretUnstructured := unstructured.Unstructured{Object: secretMap}
	secretUnstructured.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "",
		Version: "v1",
		Kind:    "Secret",
	})

	return reconciliations.CreateOrUpdateResource(ctx, k8sClient, secretUnstructured)
}

func deleteSecret(ctx context.Context, k8sClient client.Client, name, namespace string) error {
	ctrl.Log.Info("Deleting Oathkeeper JWKS Secret if it exists", "name", name, "Namespace", namespace)
	s := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	err := k8sClient.Delete(ctx, &s)

	if err != nil && !k8serrors.IsNotFound(err) {
		return fmt.Errorf("failed to delete Oathkeeper Secret %s/%s: %v", namespace, name, err)
	}

	if k8serrors.IsNotFound(err) {
		ctrl.Log.Info("Skipped deletion of Oathkeeper Secret as it wasn't present", "name", name, "Namespace", namespace)
	} else {
		ctrl.Log.Info("Successfully deleted Oathkeeper Secret", "name", name, "Namespace", namespace)
	}

	return nil
}

func generateJWKS() ([]byte, error) {
	id := uuid.New().String()
	key, err := rsa.GenerateKey(rand.Reader, jwksBits)
	if err != nil {
		return nil, errors.Wrap(err, "jwks: unable to generate RSA key")
	}

	jwks := &jose.JSONWebKeySet{
		Keys: []jose.JSONWebKey{
			{
				Algorithm:    jwksAlg,
				Use:          "sig",
				Key:          key,
				KeyID:        id,
				Certificates: []*x509.Certificate{},
			},
		},
	}

	data, err := json.Marshal(jwks)
	if err != nil {
		return nil, err
	}

	return data, nil
}
