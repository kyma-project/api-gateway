package hooks

import (
	"context"
	"encoding/base64"
	errors2 "github.com/pkg/errors"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v2 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

const shootInfoConfigMapName = "shoot-info"
const shootInfoConfigMapNamespace = "kube-system"
const devDomain = "local.kyma.dev"
const v1AccessConfigMapName = "apirule-access"
const v1AccessConfigMapNamespace = "kyma-system"
const accessSigEnvVar = "APIGATEWAY_ACCESS_SIG_BASE64"
const localKymaDevSignature = "owGbwMvMwCXG+Pmv5SmepjrGNRJJzCn5yRn7Di7NyU9OzNHLrsxN1EtJLePqKGVhEONikBVTZNEKuq1/ot3ltra401qYTlYmkB4GLk4BmEhqE8MfjlXxNVnST0R6P6vkLLno6F3M80pRbpZS9yYXttS3vcmVjAxLj85ZvOYe19a9XF2ZO1Vqv3R0BbYpVMq9ernpwxWXww9YAQ=="

func createAllowAPIRuleV1Signatures(ctx context.Context, c client.Client) error {
	log.Printf("Creating allow APIRule v1 signatures")
	gardener, err := isGardener(ctx, c)
	if err != nil {
		return err
	}
	if gardener {
		log.Printf("Creating allow APIRule v1 signatures for Gardener")
		return createSignaturesForGardener(ctx, c)
	} else {
		log.Printf("Creating allow APIRule v1 signatures for local development")
		return createSignaturesForLocalDevelopment(ctx, c)
	}
}

func isGardener(ctx context.Context, c client.Client) (bool, error) {
	log.Printf("Checking if current cluster is a Gardener one")
	cm := v1.ConfigMap{}
	err := c.Get(ctx, client.ObjectKey{Name: shootInfoConfigMapName, Namespace: shootInfoConfigMapNamespace}, &cm)
	if err != nil {
		if !errors.IsNotFound(err) {
			return false, err
		}
		return false, nil
	}
	domain := cm.Data["Domain"]
	if domain == "" {
		return false, nil
	}
	if strings.Contains(domain, devDomain) {
		return false, nil
	}
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
			"domain": devDomain,
		},
	}
	if err := c.Create(ctx, cm); err != nil {
		if !errors.IsAlreadyExists(err) {
			return err
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
		return err
	}
	err = createSignature(ctx, c, localKymaDevSignature)
	if err != nil {
		return err
	}
	log.Printf("Signatures for local development created")
	return nil
}

func createSignaturesForGardener(ctx context.Context, c client.Client) error {
	log.Printf("Creating signatures for Gardener")
	signature, ok := os.LookupEnv(accessSigEnvVar)
	if !ok || signature == "" {
		return errors2.Errorf("Signature allowing APIRule v1beta1 not found in environment variable %s", accessSigEnvVar)
	}
	err := createSignature(ctx, c, signature)
	if err != nil {
		return err
	}
	log.Printf("Signatures for Gardener created")
	return nil
}

func createSignature(ctx context.Context, c client.Client, signature string) error {
	data, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return err
	}

	cm := &v1.ConfigMap{
		ObjectMeta: v2.ObjectMeta{
			Name:      v1AccessConfigMapName,
			Namespace: v1AccessConfigMapNamespace,
		},
		BinaryData: map[string][]byte{
			"access.sig": data,
		},
	}
	if err := c.Create(ctx, cm); err != nil {
		if !errors.IsAlreadyExists(err) {
			return err
		}
		log.Printf("Configmap allowing APIRule v1 access already exists, skipping")
		return nil
	}
	return nil
}
