package auth

import (
	"context"
	"crypto/tls"
	_ "embed"
	"fmt"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/manifestprocessor"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/resource"
	"k8s.io/client-go/dynamic"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"

	"golang.org/x/net/publicsuffix"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

//go:embed oauth2-server-mock.yaml
var oauth2ServerMockManifest []byte

func GetAccessToken(oauth2Cfg clientcredentials.Config, tokenType ...string) (string, error) {
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		return "", err
	}
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		Timeout: time.Second * 10,
		Jar:     jar,
	}

	if len(tokenType) > 0 {
		oauth2Cfg.EndpointParams = make(url.Values)
		oauth2Cfg.EndpointParams.Add("token_format", tokenType[0])
	}

	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, httpClient)
	token, err := oauth2Cfg.Token(ctx)
	if err != nil {
		return "", err
	}
	if !token.Valid() {
		return "", fmt.Errorf("token invalid. got: %#v", token)
	}
	if token.TokenType != "Bearer" {
		return "", fmt.Errorf("token type = %q; want %q", token.TokenType, "Bearer")
	}
	return token.AccessToken, nil
}

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

	return fmt.Sprintf("https://oauth2-mock.%s/default", domain), nil
}
