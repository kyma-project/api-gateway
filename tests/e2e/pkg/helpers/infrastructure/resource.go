package infrastructure

import (
	"bytes"
	"testing"
	"text/template"

	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/client"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/log"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/e2e-framework/klient/decoder"
	"sigs.k8s.io/e2e-framework/klient/k8s"

	"github.com/kyma-project/api-gateway/tests/e2e/pkg/setup"
)

func CreateResourceWithTemplateValues(t *testing.T, resourceTemplate string,
	values map[string]any, opts ...decoder.DecodeOption) (k8s.Object, error) {
	t.Helper()
	resource := &unstructured.Unstructured{}

	tmpl, err := template.New("").Option("missingkey=error").Parse(resourceTemplate)
	if err != nil {
		t.Logf("Failed to parse resource template %s: %v", resourceTemplate, err)
		return nil, err
	}
	var tmplBuffer bytes.Buffer
	err = tmpl.Execute(&tmplBuffer, values)
	if err != nil {
		t.Logf("Failed to execute template for resource %s with values %v: %v", resourceTemplate, values, err)
		return nil, err
	}

	err = decoder.DecodeString(tmplBuffer.String(), resource, opts...)
	if err != nil {
		t.Logf("Failed to decode resource template\n%s\nerr=%s,\nvalues=%s", tmplBuffer.String(), err, log.StructToPrettyJson(t, values))
		return nil, err
	}

	return createResource(t, resource)
}

func UpdateResourceWithTemplateValues(t *testing.T, resourceTemplate string,
	values map[string]any, opts ...decoder.DecodeOption) (k8s.Object, error) {
	t.Helper()
	resource := &unstructured.Unstructured{}

	tmpl, err := template.New("").Option("missingkey=error").Parse(resourceTemplate)
	if err != nil {
		t.Logf("Failed to parse resource template %s: %v", resourceTemplate, err)
		return nil, err
	}
	var tmplBuffer bytes.Buffer
	err = tmpl.Execute(&tmplBuffer, values)
	if err != nil {
		t.Logf("Failed to execute template for resource %s with values %v: %v", resourceTemplate, values, err)
		return nil, err
	}

	err = decoder.DecodeString(tmplBuffer.String(), resource, opts...)
	if err != nil {
		t.Logf("Failed to decode resource template\n%s\nerr=%s,\nvalues=%s", tmplBuffer.String(), err, log.StructToPrettyJson(t, values))
		return nil, err
	}

	return updateResource(t, resource)
}

func CreateResource(t *testing.T, resourceTemplate string, opts ...decoder.DecodeOption) (k8s.Object, error) {
	t.Helper()
	resource := &unstructured.Unstructured{}
	err := decoder.DecodeString(resourceTemplate, resource, opts...)
	if err != nil {
		t.Logf("Failed to decode resource template %s: %v", resourceTemplate, err)
		return nil, err
	}

	return createResource(t, resource)
}

func createResource(t *testing.T, resource k8s.Object) (k8s.Object, error) {
	t.Helper()
	r, err := client.ResourcesClient(t)
	if err != nil {
		t.Logf("Failed to get resources client: %v", err)
		return nil, err
	}

	t.Logf("Creating %s/%s: name=\"%s\" namespace=\"%s\"",
		resource.GetObjectKind().GroupVersionKind().Kind,
		resource.GetObjectKind().GroupVersionKind().Version,
		resource.GetName(),
		resource.GetNamespace())

	setup.DeclareCleanup(t, func() {
		t.Logf("Cleaning up %s/%s: name=\"%s\" namespace=\"%s\"",
			resource.GetObjectKind().GroupVersionKind().Kind,
			resource.GetObjectKind().GroupVersionKind().Version,
			resource.GetName(),
			resource.GetNamespace())
		err := r.Delete(setup.GetCleanupContext(), resource)
		if err != nil {
			t.Logf("Failed to delete resource %s: %v", resource.GetName(), err)
			return
		}
	})

	return resource, r.Create(t.Context(), resource)
}

func updateResource(t *testing.T, resource k8s.Object) (k8s.Object, error) {
	t.Helper()
	r, err := client.ResourcesClient(t)
	if err != nil {
		t.Logf("Failed to get resources client: %v", err)
		return nil, err
	}

	t.Logf("Updating %s/%s: name=\"%s\" namespace=\"%s\"",
		resource.GetObjectKind().GroupVersionKind().Kind,
		resource.GetObjectKind().GroupVersionKind().Version,
		resource.GetName(),
		resource.GetNamespace())
	existingResource := resource.DeepCopyObject().(k8s.Object)
	err = r.Get(t.Context(), resource.GetName(), resource.GetNamespace(), existingResource)
	if err != nil {
		t.Logf("Failed to get resource %s: %v", resource.GetName(), err)
		return nil, err
	}

	resource.SetResourceVersion(existingResource.GetResourceVersion())
	resource.SetUID(existingResource.GetUID())
	resource.SetCreationTimestamp(existingResource.GetCreationTimestamp())

	t.Logf("Updating %s/%s: name=\"%s\" namespace=\"%s\"",
		resource.GetObjectKind().GroupVersionKind().Kind,
		resource.GetObjectKind().GroupVersionKind().Version,
		resource.GetName(),
		resource.GetNamespace(),
	)

	return resource, r.Update(t.Context(), resource)
}
