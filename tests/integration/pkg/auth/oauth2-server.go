package auth

import (
	_ "embed"
	"fmt"
	"github.com/avast/retry-go/v4"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/helpers"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/manifestprocessor"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/resource"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/testcontext"
	"k8s.io/client-go/dynamic"
	"log"
)

const oauthServerMockDeploymentName = "mock-oauth2-server-deployment"

//go:embed oauth2-server-mock.yaml
var oauth2ServerMockManifest []byte

// ApplyOAuth2MockServer creates OAuth2  mock server deployment, service and virtual service. Returns the issuer URL of the mock server.
func ApplyOAuth2MockServer(resourceMgr *resource.Manager, k8sClient dynamic.Interface, namespace string, domain string) (string, string, error) {

	templateData := struct {
		Domain    string
		Namespace string
	}{
		Domain:    domain,
		Namespace: namespace,
	}

	resources, err := manifestprocessor.ParseWithTemplate(oauth2ServerMockManifest, templateData)
	if err != nil {
		return "", "", err
	}

	_, err = resourceMgr.CreateResources(k8sClient, resources...)
	if err != nil {
		return "", "", err
	}

	issuerUrl := fmt.Sprintf("http://mock-oauth2-server.%s.svc.cluster.local", namespace)
	tokenUrl := fmt.Sprintf("https://oauth2-mock.%s/oauth2/token", domain)

	return issuerUrl, tokenUrl, nil
}

func EnsureOAuth2Server(resourceMgr *resource.Manager, k8sClient dynamic.Interface, namespace string, config testcontext.Config, retryOpts []retry.Option) (string, string, error) {
	if config.OIDCConfigUrl == "" {
		log.Printf("OIDC url not provided, deploying OAuth2 mock")
		issuerUrl, tokenUrl, err := ApplyOAuth2MockServer(resourceMgr, k8sClient, namespace, config.Domain)
		if err != nil {
			return "", "", err
		}
		err = helpers.WaitForDeployment(resourceMgr, k8sClient, namespace, oauthServerMockDeploymentName, retryOpts)
		if err != nil {
			return "", "", fmt.Errorf("OAuth2 mock server deployment does not work: %w", err)
		}
		log.Printf("OAuth2 mock deployed")
		return issuerUrl, tokenUrl, nil
	} else {
		log.Printf("OIDC url provided, getting configuration")
		oidcConfiguration, err := helpers.GetOIDCConfiguration(config.OIDCConfigUrl)
		if err != nil {
			return "", "", err
		}
		log.Printf("OIDC configuration received")
		issuerUrl := oidcConfiguration.Issuer
		tokenUrl := oidcConfiguration.TokenEndpoint
		return issuerUrl, tokenUrl, nil
	}
}
