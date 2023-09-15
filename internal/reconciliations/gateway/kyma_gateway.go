package gateway

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	"k8s.io/apimachinery/pkg/util/yaml"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"text/template"
)

const (
	disclaimerKey   = "apigateways.operator.kyma-project.io/managed-by-disclaimer"
	disclaimerValue = "DO NOT EDIT - This resource is managed by Kyma.\nAny modifications are discarded and the resource is reverted to the original state."
)

//go:embed kyma_gateway.yaml
var manifest []byte

func ReconcileKymaGateway(ctx context.Context, k8sClient client.Client, apiGatewayCR v1alpha1.APIGateway) error {

	if apiGatewayCR.Spec.EnableKymaGateway == nil || *apiGatewayCR.Spec.EnableKymaGateway == false {
		return nil
	}

	resourceTemplate, err := template.New("tmpl").Option("missingkey=error").Parse(string(manifest))
	if err != nil {
		return fmt.Errorf("failed to parse template for Kyma gateway yaml: %v", err)
	}

	templateValues := make(map[string]string)
	domain, err := getKymaGatewayDomain(ctx, k8sClient)
	if err != nil {
		return fmt.Errorf("failed to get Kyma gateway domain: %v", err)
	}
	templateValues["DomainName"] = domain

	var resourceBuffer bytes.Buffer
	err = resourceTemplate.Execute(&resourceBuffer, templateValues)
	if err != nil {
		return fmt.Errorf("failed to apply parsed template for Kyma gateway yaml: %v", err)
	}

	var gateway unstructured.Unstructured
	err = yaml.Unmarshal(resourceBuffer.Bytes(), &gateway)
	if err != nil {
		return fmt.Errorf("failed to decode Kyma gateway yaml: %v", err)
	}

	spec := gateway.Object["spec"]
	_, err = controllerutil.CreateOrUpdate(ctx, k8sClient, &gateway, func() error {
		annotations := map[string]string{
			disclaimerKey: disclaimerValue,
		}
		gateway.SetAnnotations(annotations)
		gateway.Object["spec"] = spec
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to create or update Kyma gateway: %v", err)
	}

	return nil
}
