package v2alpha1

import (
	"context"
	_ "embed"
	"encoding/base64"
	"fmt"
	"log"
	"path"

	"github.com/kyma-project/api-gateway/tests/integration/pkg/hooks"

	"github.com/cucumber/godog"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/auth"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/helpers"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/manifestprocessor"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/resource"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/testcontext"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

const manifestsDirectory = "testsuites/v2alpha1/manifests/"

func (t *testsuite) createScenario() *scenario {
	ns := t.namespace
	testId := helpers.GenerateRandomTestId()

	template := make(map[string]string)
	template["Namespace"] = ns
	template["TestID"] = testId
	template["Domain"] = t.config.Domain
	template["GatewayName"] = t.config.GatewayName
	template["GatewayNamespace"] = t.config.GatewayNamespace
	template["IssuerUrl"] = t.config.IssuerUrl
	template["EncodedCredentials"] = base64.RawStdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", t.config.ClientID, t.config.ClientSecret)))
	template["extAuthProvider"] = "sample-ext-authz-http"

	return &scenario{
		Namespace:            ns,
		TestID:               testId,
		Domain:               t.config.Domain,
		ManifestTemplate:     template,
		ApiResourceDirectory: path.Dir(manifestsDirectory),
		k8sClient:            t.K8sClient(),
		oauth2Cfg:            t.oauth2Cfg,
		httpClient:           t.httpClient,
		resourceManager:      t.ResourceManager(),
		config:               t.config,
		jwtConfig:            t.jwtConfig,
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
	initScenario(ctx, t)
}

func (t *testsuite) FeaturePath() []string {
	return []string{"testsuites/v2alpha1/features/"}
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
	namespace := fmt.Sprintf("%s-%s", t.name, helpers.GenerateRandomString(6))
	log.Printf("Using namespace: %s\n", namespace)

	// create common resources for all scenarios
	globalCommonResources, err := manifestprocessor.ParseFromFileWithTemplate("global-commons.yaml", manifestsDirectory, struct {
		Namespace string
	}{
		Namespace: namespace,
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

	log.Printf("Creating common tests resources")
	_, err = t.resourceManager.CreateResources(t.k8sClient, globalCommonResources...)
	if err != nil {
		return err
	}

	var tokenURL string
	if t.config.OIDCConfigUrl == "empty" {
		issuerUrl, err := auth.ApplyOAuth2MockServer(t.resourceManager, t.k8sClient, namespace, t.config.Domain)
		if err != nil {
			return err
		}
		t.config.IssuerUrl = fmt.Sprintf("http://mock-oauth2-server.%s.svc.cluster.local", namespace)
		tokenURL = fmt.Sprintf("%s/oauth2/token", issuerUrl)
	} else {
		oidcConfiguration, err := helpers.GetOIDCConfiguration(t.config.OIDCConfigUrl)
		if err != nil {
			return err
		}
		t.config.IssuerUrl = oidcConfiguration.Issuer
		tokenURL = oidcConfiguration.TokenEndpoint
	}

	t.oauth2Cfg = &clientcredentials.Config{
		ClientID:     t.config.ClientID,
		ClientSecret: t.config.ClientSecret,
		TokenURL:     tokenURL,
		AuthStyle:    oauth2.AuthStyleInHeader,
		Scopes:       []string{"read"},
	}

	t.namespace = namespace

	t.jwtConfig = &clientcredentials.Config{
		ClientID:     t.config.ClientID,
		ClientSecret: t.config.ClientSecret,
		TokenURL:     tokenURL,
		AuthStyle:    oauth2.AuthStyleInHeader,
	}

	return nil
}

func (t *testsuite) TearDown() {
	res := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "namespaces"}
	err := t.k8sClient.Resource(res).Delete(context.Background(), t.namespace, v1.DeleteOptions{})
	if err != nil {
		log.Print(err.Error())
	}
}

func (t *testsuite) BeforeSuiteHooks() []func() error {
	return []func() error{hooks.ExtAuthorizerInstallHook(t), hooks.ApplyAndVerifyApiGatewayWithoutOathkeeperCrSuiteHook}
}

func (t *testsuite) AfterSuiteHooks() []func() error {
	return []func() error{hooks.DeleteBlockingResourcesSuiteHook, hooks.ApiGatewayCrTearDownSuiteHook, hooks.ExtAuthorizerRemoveHook(t)}
}

func NewTestsuite(httpClient *helpers.RetryableHttpClient, k8sClient dynamic.Interface, rm *resource.Manager, config testcontext.Config) testcontext.Testsuite {
	return &testsuite{
		name:            "v2alpha1",
		httpClient:      httpClient,
		k8sClient:       k8sClient,
		resourceManager: rm,
		config:          config,
	}
}
