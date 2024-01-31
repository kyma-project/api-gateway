package auth

import (
	"context"
	"crypto/tls"
	_ "embed"
	"fmt"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/manifestprocessor"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
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

// PatchIstiodDeploymentWithEnvironmentVariables patches the istiod deployment with the given environment variables.
func PatchIstiodDeploymentWithEnvironmentVariables(resourceMgr *resource.Manager, k8sClient dynamic.Interface, namespace string, environmentVariables map[string]string) error {
	log.Printf("Patching istiod deployment with environment variables: %v", environmentVariables)
	res, err := resourceMgr.GetResource(k8sClient, schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}, namespace, "istiod")
	if err != nil {
		return fmt.Errorf("could not get istiod deployment: %s", err.Error())
	}

	containers, found, err := unstructured.NestedSlice(res.Object, "spec", "template", "spec", "containers")
	if err != nil || !found {
		return fmt.Errorf("could not find containers in istiod deployment: %s", err.Error())
	}
	if len(containers) != 1 {
		return fmt.Errorf("istiod deployment contains more than one container")
	}

	env, found, err := unstructured.NestedSlice(containers[0].(map[string]interface{}), "env")
	if err != nil || !found {
		return fmt.Errorf("could not find env in istiod deployment: %s", err.Error())
	}

	for key, value := range environmentVariables {
		env = append(env, map[string]interface{}{"name": key, "value": value})
	}

	err = unstructured.SetNestedSlice(containers[0].(map[string]interface{}), env, "env")
	if err != nil {
		return fmt.Errorf("could not set env in istiod deployment: %s", err.Error())
	}

	err = unstructured.SetNestedSlice(res.Object, containers, "spec", "template", "spec", "containers")
	if err != nil {
		return fmt.Errorf("could not set containers in istiod deployment: %s", err.Error())
	}

	err = resourceMgr.UpdateResource(k8sClient, schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}, namespace, "istiod", *res)
	if err != nil {
		return fmt.Errorf("could not update istiod deployment: %s", err.Error())
	}

	return nil
}

// RemoveEnvironmentVariableFromIstiodDeployment removes the given environment variable from the deployment.
func RemoveEnvironmentVariableFromIstiodDeployment(resourceMgr *resource.Manager, k8sClient dynamic.Interface, namespace string, environmentVariableName string) error {
	log.Printf("Removing environment variable %s from istiod deployment", environmentVariableName)
	res, err := resourceMgr.GetResource(k8sClient, schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}, namespace, "istiod")
	if err != nil {
		return fmt.Errorf("could not get istiod deployment: %s", err.Error())
	}

	containers, found, err := unstructured.NestedSlice(res.Object, "spec", "template", "spec", "containers")
	if err != nil || !found {
		return fmt.Errorf("could not find containers in istiod deployment: %s", err.Error())
	}
	if len(containers) != 1 {
		return fmt.Errorf("istiod deployment contains more than one container")
	}

	env, found, err := unstructured.NestedSlice(containers[0].(map[string]interface{}), "env")
	if err != nil || !found {
		return fmt.Errorf("could not find env in istiod deployment: %s", err.Error())
	}

	for i, v := range env {
		if v.(map[string]interface{})["name"] == environmentVariableName {
			env = append(env[:i], env[i+1:]...)
			break
		}
	}

	err = unstructured.SetNestedSlice(containers[0].(map[string]interface{}), env, "env")
	if err != nil {
		return fmt.Errorf("could not set env in istiod deployment: %s", err.Error())
	}

	err = unstructured.SetNestedSlice(res.Object, containers, "spec", "template", "spec", "containers")
	if err != nil {
		return fmt.Errorf("could not set containers in istiod deployment: %s", err.Error())
	}

	err = resourceMgr.UpdateResource(k8sClient, schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}, namespace, "istiod", *res)
	if err != nil {
		return fmt.Errorf("could not update istiod deployment: %s", err.Error())
	}

	return nil
}
