package certificate_test

import (
	"context"
	_ "embed"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

//go:generate cp ../../internal/reconciliations/gateway/apirule_crd.yaml apirule_crd.yaml
//go:generate cp ../../config/conversion_webhook/apirule_mutating_webhook.yaml apirule_mutating_webhook.yaml
//go:embed apirule_crd.yaml
var APIRuleCRD []byte

//go:embed apirule_mutating_webhook.yaml
var APIRuleMutatingWebhook []byte

func ApplyAPIRuleCRD(k8sClient client.Client) error {
	var resource unstructured.Unstructured
	err := yaml.Unmarshal(APIRuleCRD, &resource)
	if err != nil {
		return err
	}
	if err := client.IgnoreAlreadyExists(
		k8sClient.Create(context.Background(), &resource),
	); err != nil {
		return err
	}

	return nil
}

func ApplyAPIRuleMutatingWebhook(k8sClient client.Client) error {
	var resource unstructured.Unstructured
	err := yaml.Unmarshal(APIRuleMutatingWebhook, &resource)
	if err != nil {
		return err
	}
	resource.SetName("api-gateway-mutating-webhook-configuration")
	if err := client.IgnoreAlreadyExists(
		k8sClient.Create(context.Background(), &resource),
	); err != nil {
		return err
	}

	return nil
}
