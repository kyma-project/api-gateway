package hooks

import (
	"context"
	"encoding/base64"
	"fmt"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v2 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

const (
	shootInfoConfigMapName      = "shoot-info"
	shootInfoConfigMapNamespace = "kube-system"
	domainKey                   = "domain"
	devDomain                   = "local.kyma.dev"
	v1AccessConfigMapName       = "apirule-access"
	v1AccessConfigMapNamespace  = "kyma-system"
	signatureKey                = "access.sig"
	accessSigEnvVar             = "APIGATEWAY_ACCESS_SIG_BASE64"
	localKymaDevSignature       = "xEYGAAobIJRdbtfrgZYkBehKLGT3pI8YVu22FPHyHJWVjpTzvSPa+8vQFjsiHcrLvmDfEy56Y/D9Xfq/Qtt6o41bvKMqJPUByxRiAAAAAABsb2NhbC5reW1hLmRldsKYBgAbCgAAACkFgmj7jOoioQb7y9AWOyIdysu+YN8TLnpj8P1d+r9C23qjjVu8oyok9QAAAACp7CCUXW7X64GWJAXoSixk96SPGFbtthTx8hyVlY6U870j2t8v/C1gL5Vkw9+y7sfd/GKzAZGIwlf6+XDM8U4VlHtS/CRKP155fLX9g96/jixWU7JZgCf3Yo/a5Bwjg0TYkQM="
)

func createAllowAPIRuleV1Signatures(ctx context.Context, c client.Client) error {
	log.Printf("Creating signatures to allow APIRule v1 usage")
	gardener, err := isGardener(ctx, c)
	if err != nil {
		return fmt.Errorf("can't check whether current cluster is a Gardener one: %w", err)
	}
	if gardener {
		return createSignaturesForGardener(ctx, c)
	} else {
		return createSignaturesForLocalDevelopment(ctx, c)
	}
}

func isGardener(ctx context.Context, c client.Client) (bool, error) {
	log.Printf("Checking if current cluster is a Gardener one")
	cm := v1.ConfigMap{}
	err := c.Get(ctx, client.ObjectKey{Name: shootInfoConfigMapName, Namespace: shootInfoConfigMapNamespace}, &cm)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Printf("Shoot-info not found, it is not a Gardener cluster")
		} else {
			return false, fmt.Errorf("can't get shoot-info configmap: %w", err)
		}
		return false, nil
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

func createShootInfoWithDevDomain(ctx context.Context, c client.Client) error {
	log.Printf("Creating shoot-info configmap with dev domain")
	cm := &v1.ConfigMap{
		ObjectMeta: v2.ObjectMeta{
			Name:      shootInfoConfigMapName,
			Namespace: shootInfoConfigMapNamespace,
		},
		Data: map[string]string{
			domainKey: devDomain,
		},
	}
	if err := c.Create(ctx, cm); err != nil {
		if !errors.IsAlreadyExists(err) {
			return fmt.Errorf("can't create shoot-info configmap: %w", err)
		}
		log.Printf("shoot-info already exists, skipping")
		return nil
	}
	log.Printf("shoot-info configmap with dev domain created")
	return nil
}

func createSignaturesForLocalDevelopment(ctx context.Context, c client.Client) error {
	log.Printf("Creating signatures for local development")
	err := createShootInfoWithDevDomain(ctx, c)
	if err != nil {
		return fmt.Errorf("can't create shoot-info with dev domain: %w", err)
	}
	err = createSignature(ctx, c, localKymaDevSignature)
	if err != nil {
		return fmt.Errorf("can't create configmap with signature: %w", err)
	}
	log.Printf("Signatures for local development created")
	return nil
}

func createSignaturesForGardener(ctx context.Context, c client.Client) error {
	log.Printf("Creating configmap with signature for the Gardener cluster")
	signature, ok := os.LookupEnv(accessSigEnvVar)
	if !ok || signature == "" {
		return fmt.Errorf("signature allowing APIRule v1beta1 usage not found in environment variable %s", accessSigEnvVar)
	}
	err := createSignature(ctx, c, signature)
	if err != nil {
		return fmt.Errorf("can't create signatures for Gardener cluster: %w", err)
	}
	log.Printf("Signatures for Gardener cluster created")
	return nil
}

func createSignature(ctx context.Context, c client.Client, signature string) error {
	data, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return fmt.Errorf("can't decode signature: %w", err)
	}

	cm := &v1.ConfigMap{
		ObjectMeta: v2.ObjectMeta{
			Name:      v1AccessConfigMapName,
			Namespace: v1AccessConfigMapNamespace,
		},
		BinaryData: map[string][]byte{
			signatureKey: data,
		},
	}
	if err := c.Create(ctx, cm); err != nil {
		if errors.IsAlreadyExists(err) {
			log.Printf("Configmap with a signature allowing APIRule v1 usage already exists, skipping")
		} else {
			return fmt.Errorf("can't create configmap with signature allowing APIRule v1 usage: %w", err)
		}
		return nil
	}
	return nil
}
