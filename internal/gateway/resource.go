package gateway

import (
	"bytes"
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"text/template"
)

func applyResource(ctx context.Context, k8sClient client.Client, resourceManifest []byte, templateValues map[string]string) error {

	resourceBuffer, err := applyTemplateValuesToResourceManifest(resourceManifest, templateValues)
	if err != nil {
		return fmt.Errorf("failed to apply template values to resource manifest: %v", err)
	}

	resource, err := unmarshalResourceBuffer(resourceBuffer.Bytes())
	if err != nil {
		return fmt.Errorf("failed to unmarshall yaml: %v", err)
	}

	return createOrUpdateResource(ctx, k8sClient, resource)
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

func createOrUpdateResource(ctx context.Context, k8sClient client.Client, resource unstructured.Unstructured) error {
	spec, specExist := resource.Object["spec"]
	data, dataExist := resource.Object["data"]

	_, err := controllerutil.CreateOrUpdate(ctx, k8sClient, &resource, func() error {
		annotations := map[string]string{
			disclaimerKey: disclaimerValue,
		}
		resource.SetAnnotations(annotations)

		if dataExist {
			resource.Object["data"] = data
		}

		if specExist {
			resource.Object["spec"] = spec
		}

		return nil
	})

	return err
}
