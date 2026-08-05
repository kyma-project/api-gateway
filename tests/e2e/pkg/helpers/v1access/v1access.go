package v1access

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"strings"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
)

const (
	shootInfoConfigMapName      = "shoot-info"
	shootInfoConfigMapNamespace = "kube-system"
	domainKey                   = "domain"
	devDomain                   = "local.kyma.dev"
	V1AccessConfigMapName       = "apirule-access"
	V1AccessConfigMapNamespace  = "kyma-system"
	signatureKey                = "access.sig"
	accessSigEnvVar             = "APIGATEWAY_ACCESS_SIG_BASE64"
	localKymaDevSignature       = "xEYGAAobIJRdbtfrgZYkBehKLGT3pI8YVu22FPHyHJWVjpTzvSPa+8vQFjsiHcrLvmDfEy56Y/D9Xfq/Qtt6o41bvKMqJPUByxRiAAAAAABsb2NhbC5reW1hLmRldsKYBgAbCgAAACkFgmj7jOoioQb7y9AWOyIdysu+YN8TLnpj8P1d+r9C23qjjVu8oyok9QAAAACp7CCUXW7X64GWJAXoSixk96SPGFbtthTx8hyVlY6U870j2t8v/C1gL5Vkw9+y7sfd/GKzAZGIwlf6+XDM8U4VlHtS/CRKP155fLX9g96/jixWU7JZgCf3Yo/a5Bwjg0TYkQM="
)

func CreateAllowAPIRuleV1Signatures(ctx context.Context, r *resources.Resources) error {
	log.Printf("Creating signatures to allow APIRule v1 usage")

	gardener, err := isGardener(ctx, r)
	if err != nil {
		return fmt.Errorf("can't check whether current cluster is a Gardener one: %w", err)
	}
	if gardener {
		return createSignaturesForGardener(ctx, r)
	}
	return createSignaturesForLocalDevelopment(ctx, r)
}

func isGardener(ctx context.Context, r *resources.Resources) (bool, error) {
	log.Printf("Checking if current cluster is a Gardener one")

	cm := v1.ConfigMap{}
	err := r.Get(ctx, shootInfoConfigMapName, shootInfoConfigMapNamespace, &cm)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Printf("Shoot-info not found, it is not a Gardener cluster")
			return false, nil
		}
		return false, fmt.Errorf("can't get shoot-info configmap: %w", err)
	}

	domain := cm.Data[domainKey]
	if domain == "" {
		return false, fmt.Errorf("shoot-info configmap does not have a domain")
	}
	if strings.Contains(domain, devDomain) {
		log.Printf("Shoot-info configmap contains dev domain, it is not a Gardener cluster")
		return false, nil
	}

	log.Printf("Shoot-info configmap contains a domain, it is a Gardener cluster")
	return true, nil
}

func createShootInfoWithDevDomain(ctx context.Context, r *resources.Resources) error {
	log.Printf("Creating shoot-info configmap with dev domain")

	cm := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      shootInfoConfigMapName,
			Namespace: shootInfoConfigMapNamespace,
		},
		Data: map[string]string{
			domainKey: devDomain,
		},
	}
	if err := r.Create(ctx, cm); err != nil {
		if !errors.IsAlreadyExists(err) {
			return fmt.Errorf("can't create shoot-info configmap: %w", err)
		}
		log.Printf("shoot-info already exists, skipping")
	}
	return nil
}

func createSignaturesForLocalDevelopment(ctx context.Context, r *resources.Resources) error {
	log.Printf("Creating signatures for local development")

	if err := createShootInfoWithDevDomain(ctx, r); err != nil {
		return fmt.Errorf("can't create shoot-info with dev domain: %w", err)
	}
	if err := createSignature(ctx, r, localKymaDevSignature); err != nil {
		return fmt.Errorf("can't create configmap with signature: %w", err)
	}

	log.Printf("Signatures for local development created")
	return nil
}

func createSignaturesForGardener(ctx context.Context, r *resources.Resources) error {
	log.Printf("Creating configmap with signature for the Gardener cluster")

	signature, ok := os.LookupEnv(accessSigEnvVar)
	if !ok || signature == "" {
		return fmt.Errorf("signature allowing APIRule v1beta1 usage not found in environment variable %s", accessSigEnvVar)
	}
	if err := createSignature(ctx, r, signature); err != nil {
		return fmt.Errorf("can't create signatures for Gardener cluster: %w", err)
	}

	log.Printf("Signatures for Gardener cluster created")
	return nil
}

func createSignature(ctx context.Context, r *resources.Resources, signature string) error {
	data, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return fmt.Errorf("can't decode signature: %w", err)
	}

	cm := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      V1AccessConfigMapName,
			Namespace: V1AccessConfigMapNamespace,
		},
		BinaryData: map[string][]byte{
			signatureKey: data,
		},
	}
	if err := r.Create(ctx, cm); err != nil {
		if errors.IsAlreadyExists(err) {
			log.Printf("Configmap with a signature allowing APIRule v1 usage already exists, skipping")
			return nil
		}
		return fmt.Errorf("can't create configmap with signature allowing APIRule v1 usage: %w", err)
	}
	return nil
}
