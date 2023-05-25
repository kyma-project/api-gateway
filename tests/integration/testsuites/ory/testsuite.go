package ory

import (
	"context"
	_ "embed"
	"encoding/base64"
	"fmt"
	"github.com/cucumber/godog"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/helpers"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/manifestprocessor"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/resource"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/testcontext"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"log"
	"path"
	"time"
)

const manifestsDirectory = "testsuites/ory/manifests/"

func (t *testsuite) createScenario(templateFileName string, scenarioName string) *scenario {
	ns := t.namespace
	testId := helpers.GenerateRandomTestId()

	template := make(map[string]string)
	template["Namespace"] = ns
	template["NamePrefix"] = scenarioName
	template["TestID"] = testId
	template["Domain"] = t.config.Domain
	template["GatewayName"] = t.config.GatewayName
	template["GatewayNamespace"] = t.config.GatewayNamespace
	template["IssuerUrl"] = t.config.IssuerUrl
	template["EncodedCredentials"] = base64.RawStdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", t.config.ClientID, t.config.ClientSecret)))

	return &scenario{
		Namespace:               ns,
		TestID:                  testId,
		Domain:                  t.config.Domain,
		ManifestTemplate:        template,
		ApiResourceManifestPath: templateFileName,
		ApiResourceDirectory:    path.Dir(manifestsDirectory),
		k8sClient:               t.K8sClient(),
		oauth2Cfg:               t.oauth2Cfg,
		httpClient:              t.httpClient,
		resourceManager:         t.ResourceManager(),
		config:                  t.config,
		jwtConfig:               t.jwtConfig,
	}
}

type testsuite struct {
	name            string
	namespace       string
	httpClient      *helpers.RetryableHttpClient
	k8sClient       dynamic.Interface
	resourceManager *resource.Manager
	config          testcontext.Config
	oauth2Cfg       *clientcredentials.Config
	jwtConfig       *clientcredentials.Config
}

func (t *testsuite) InitScenarios(ctx *godog.ScenarioContext) {
	initScenarioOAuth2JWTOnePath(ctx, t)
	initScenarioOAuth2JWTTwoPaths(ctx, t)
	initScenarioOAuth2Endpoint(ctx, t)
	initScenarioServicePerPath(ctx, t)
	initScenarioUnsecuredEndpoint(ctx, t)
	initScenarioSecuredToUnsecuredEndpoint(ctx, t)
	initScenarioUnsecuredToSecuredEndpointOauthJwt(ctx, t)
}

func (t *testsuite) FeaturePath() string {
	return "testsuites/ory/features/"
}

func (t *testsuite) Name() string {
	return t.name
}

func (t *testsuite) ResourceManager() *resource.Manager {
	return t.resourceManager
}

func (t *testsuite) K8sClient() dynamic.Interface {
	return t.k8sClient
}

func (t *testsuite) Setup() error {
	oauthClientID := helpers.GenerateRandomString(8)
	oauthClientSecret := helpers.GenerateRandomString(8)

	oauthSuffix := helpers.GenerateRandomString(6)
	oauthSecretName := fmt.Sprintf("%s-secret-%s", t.name, oauthSuffix)
	oauthClientName := fmt.Sprintf("%s-client-%s", t.name, oauthSuffix)

	namespace := fmt.Sprintf("%s-%s", t.name, helpers.GenerateRandomString(6))
	log.Printf("Using namespace: %s\n", namespace)
	log.Printf("Using OAuth2Client with name: %s, secretName: %s\n", oauthClientName, oauthSecretName)

	hydraAddress := fmt.Sprintf("oauth2.%s", t.config.Domain)

	oauth2Cfg := &clientcredentials.Config{
		ClientID:     oauthClientID,
		ClientSecret: oauthClientSecret,
		TokenURL:     fmt.Sprintf("https://%s/oauth2/token", hydraAddress),
		Scopes:       []string{"read"},
		AuthStyle:    oauth2.AuthStyleInHeader,
	}

	jwtConfig := &clientcredentials.Config{
		ClientID:     t.config.ClientID,
		ClientSecret: t.config.ClientSecret,
		TokenURL:     fmt.Sprintf("%s/oauth2/token", t.config.IssuerUrl),
		AuthStyle:    oauth2.AuthStyleInHeader,
	}

	// create common resources for all scenarios
	globalCommonResources, err := manifestprocessor.ParseFromFileWithTemplate("global-commons.yaml", manifestsDirectory, struct {
		Namespace         string
		OauthClientSecret string
		OauthClientID     string
		OauthSecretName   string
	}{
		Namespace:         namespace,
		OauthClientSecret: base64.StdEncoding.EncodeToString([]byte(oauthClientSecret)),
		OauthClientID:     base64.StdEncoding.EncodeToString([]byte(oauthClientID)),
		OauthSecretName:   oauthSecretName,
	})
	if err != nil {
		return err
	}

	// delete test namespace if the previous test namespace persists
	nsResourceSchema, ns, name := t.resourceManager.GetResourceSchemaAndNamespace(globalCommonResources[0])
	log.Printf("Delete test namespace, if exists: %s\n", name)
	err = t.resourceManager.DeleteResource(t.k8sClient, nsResourceSchema, ns, name)
	if err != nil {
		return err
	}

	time.Sleep(time.Duration(t.config.ReqDelay) * time.Second)

	log.Printf("Creating common tests resources")
	_, err = t.resourceManager.CreateResources(t.k8sClient, globalCommonResources...)
	if err != nil {
		return err
	}

	hydraClientResource, err := manifestprocessor.ParseFromFileWithTemplate("hydra-client.yaml", manifestsDirectory, struct {
		Namespace       string
		OauthClientName string
		OauthSecretName string
	}{
		Namespace:       namespace,
		OauthClientName: oauthClientName,
		OauthSecretName: oauthSecretName,
	})
	if err != nil {
		return err
	}
	log.Printf("Creating hydra client resources")

	_, err = t.resourceManager.CreateResources(t.k8sClient, hydraClientResource...)
	if err != nil {
		return err
	}

	// Let's wait a bit to register client in hydra
	time.Sleep(time.Duration(t.config.ReqDelay) * time.Second)

	// Get HydraClient Status
	hydraClientResourceSchema, ns, name := t.resourceManager.GetResourceSchemaAndNamespace(hydraClientResource[0])
	clientStatus, err := t.resourceManager.GetStatus(t.k8sClient, hydraClientResourceSchema, ns, name)
	errorStatus, ok := clientStatus["reconciliationError"].(map[string]interface{})
	if err != nil || !ok {
		return fmt.Errorf("error retrieving Oauth2Client status: %+v | %+v", err, ok)
	}
	if len(errorStatus) != 0 {
		return fmt.Errorf("Invalid status in Oauth2Client resource: %+v", errorStatus)
	}

	t.oauth2Cfg = oauth2Cfg
	t.namespace = namespace
	t.jwtConfig = jwtConfig

	return nil
}

func (t *testsuite) TearDown() {
	res := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "namespaces"}
	err := t.k8sClient.Resource(res).Delete(context.Background(), t.namespace, v1.DeleteOptions{})
	if err != nil {
		log.Print(err.Error())
	}
}

func NewTestsuite(httpClient *helpers.RetryableHttpClient, k8sClient dynamic.Interface, rm *resource.Manager, config testcontext.Config) testcontext.Testsuite {

	return &testsuite{
		name:            "ory",
		httpClient:      httpClient,
		k8sClient:       k8sClient,
		resourceManager: rm,
		config:          config,
	}
}
