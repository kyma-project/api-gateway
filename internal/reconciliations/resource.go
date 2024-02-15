package reconciliations

import (
	"bytes"
	"context"
	"fmt"
	"text/template"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	disclaimerKey   = "apigateways.operator.kyma-project.io/managed-by-disclaimer"
	disclaimerValue = "DO NOT EDIT - This resource is managed by Kyma.\nAny modifications are discarded and the resource is reverted to the original state."

	labelModuleKey   = "kyma-project.io/module"
	labelModuleValue = "api-gateway"

	Namespace = "kyma-system"
)

func ApplyResource(ctx context.Context, k8sClient client.Client, resourceManifest []byte, templateValues map[string]string) error {
	resource, err := CreateUnstructuredResource(resourceManifest, templateValues)
	if err != nil {
		return err
	}

	return CreateOrUpdateResource(ctx, k8sClient, resource)
}

func CreateUnstructuredResource(resourceManifest []byte, templateValues map[string]string) (unstructured.Unstructured, error) {
	resourceBuffer, err := applyTemplateValuesToResourceManifest(resourceManifest, templateValues)
	if err != nil {
		return unstructured.Unstructured{}, fmt.Errorf("failed to apply template values to resource manifest: %v", err)
	}

	resource, err := unmarshalResourceBuffer(resourceBuffer.Bytes())
	if err != nil {
		return unstructured.Unstructured{}, fmt.Errorf("failed to unmarshall yaml: %v", err)
	}

	return resource, nil
}

func applyTemplateValuesToResourceManifest(resourceManifest []byte, templateValues map[string]string) (bytes.Buffer, error) {
	var resourceBuffer bytes.Buffer

	resourceTemplate, err := template.New("tmpl").Option("missingkey=error").Parse(string(resourceManifest))
	if err != nil {
		return resourceBuffer, err
	}

	err = resourceTemplate.Execute(&resourceBuffer, templateValues)
	if err != nil {
		return resourceBuffer, err
	}

	return resourceBuffer, nil
}

func unmarshalResourceBuffer(resourceBuffer []byte) (unstructured.Unstructured, error) {
	var resource unstructured.Unstructured

	err := yaml.Unmarshal(resourceBuffer, &resource)
	if err != nil {
		return resource, err
	}

	return resource, nil
}

func CreateOrUpdateResource(ctx context.Context, k8sClient client.Client, resource unstructured.Unstructured) error {
	spec, specExist := resource.Object["spec"]
	data, dataExist := resource.Object["data"]
	labels := resource.GetLabels()

	_, err := controllerutil.CreateOrUpdate(ctx, k8sClient, &resource, func() error {
		annotations := map[string]string{
			disclaimerKey: disclaimerValue,
		}
		resource.SetAnnotations(annotations)

		if len(labels) == 0 {
			labels = map[string]string{}
		}
		labels[labelModuleKey] = labelModuleValue
		resource.SetLabels(labels)

		if specExist {
			resource.Object["spec"] = spec
		}

		if dataExist {
			resource.Object["data"] = data
		}

		return nil
	})

	return err
}
