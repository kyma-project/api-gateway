package auth

import (
	_ "embed"
	"fmt"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/manifestprocessor"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/resource"
	"k8s.io/client-go/dynamic"
	"log"
)

//go:embed oauth2-server-mock.yaml
var oauth2ServerMockManifest []byte

// ApplyOAuth2MockServer creates OAuth2  mock server deployment, service and virtual service. Returns the issuer URL of the mock server.
// Additional Info can be found in the documentation of the OAuth2 mock server: https://github.com/navikt/mock-oauth2-server.
func ApplyOAuth2MockServer(resourceMgr *resource.Manager, k8sClient dynamic.Interface, namespace string, domain string) (string, error) {

	log.Printf("Applying OAuth2 mock server")
	templateData := struct {
		Domain    string
		Namespace string
	}{
		Domain:    domain,
		Namespace: namespace,
	}

	resources, err := manifestprocessor.ParseWithTemplate(oauth2ServerMockManifest, templateData)
	if err != nil {
		return "", err
	}

	_, err = resourceMgr.CreateResources(k8sClient, resources...)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("https://oauth2-mock.%s", domain), nil
}
