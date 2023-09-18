package gateway

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"text/template"
)

//go:embed certificate.yaml
var certificateManifest []byte

func reconcileCertificate(ctx context.Context, k8sClient client.Client, name, domain, certSecretName string) error {

	resourceTemplate, err := template.New("tmpl").Option("missingkey=error").Parse(string(certificateManifest))
	if err != nil {
		return fmt.Errorf("failed to parse template yaml for Certificate %s: %v", name, err)
	}

	templateValues := make(map[string]string)
	templateValues["Name"] = name
	templateValues["Domain"] = domain
	templateValues["SecretName"] = certSecretName

	var resourceBuffer bytes.Buffer
	err = resourceTemplate.Execute(&resourceBuffer, templateValues)
	if err != nil {
		return fmt.Errorf("failed to apply parsed template yaml for Certificate %s: %v", name, err)
	}

	var dnsEntry unstructured.Unstructured
	err = yaml.Unmarshal(resourceBuffer.Bytes(), &dnsEntry)
	if err != nil {
		return fmt.Errorf("failed to decode yaml for Certificate %s: %v", name, err)
	}

	spec := dnsEntry.Object["spec"]
	_, err = controllerutil.CreateOrUpdate(ctx, k8sClient, &dnsEntry, func() error {
		annotations := map[string]string{
			disclaimerKey: disclaimerValue,
		}
		dnsEntry.SetAnnotations(annotations)
		dnsEntry.Object["spec"] = spec
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to create or update for Certificate %s: %v", name, err)
	}

	return nil

}
