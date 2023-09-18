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

func reconcileResource(ctx context.Context, k8sClient client.Client, resourceManifest []byte, templateValues map[string]string) error {

	resourceTemplate, err := template.New("tmpl").Option("missingkey=error").Parse(string(resourceManifest))
	if err != nil {
		return fmt.Errorf("failed to parse template yaml: %v", err)
	}

	var resourceBuffer bytes.Buffer
	err = resourceTemplate.Execute(&resourceBuffer, templateValues)
	if err != nil {
		return fmt.Errorf("failed to apply parsed template yaml: %v", err)
	}

	var resource unstructured.Unstructured
	err = yaml.Unmarshal(resourceBuffer.Bytes(), &resource)
	if err != nil {
		return fmt.Errorf("failed to decode yaml: %v", err)
	}

	spec := resource.Object["spec"]
	_, err = controllerutil.CreateOrUpdate(ctx, k8sClient, &resource, func() error {
		annotations := map[string]string{
			disclaimerKey: disclaimerValue,
		}
		resource.SetAnnotations(annotations)
		resource.Object["spec"] = spec
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to create or update: %v", err)
	}

	return nil

}
