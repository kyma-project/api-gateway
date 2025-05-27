package v2alpha1

import (
	_ "embed"
	"encoding/base64"
	"fmt"
	"log"
	"path"

	"github.com/cucumber/godog"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	"k8s.io/client-go/dynamic"

	"github.com/kyma-project/api-gateway/tests/integration/pkg/auth"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/global"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/helpers"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/hooks"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/resource"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/testcontext"
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

func (t *testsuite) ValidateAndFixConfig() error {
	return t.config.ValidateCommon(t.resourceManager, t.k8sClient)
}

func (t *testsuite) TestConcurrency() int { return t.config.TestConcurrency }

func (t *testsuite) ResourceManager() *resource.Manager {
	return t.resourceManager
}

func (t *testsuite) K8sClient() dynamic.Interface {
	return t.k8sClient
}

func (t *testsuite) Setup() error {
	namespace := global.GenerateNamespaceName(t.name)
	t.namespace = namespace
	log.Printf("Using namespace: %s", namespace)

	err := global.CreateGlobalResources(t.resourceManager, t.k8sClient, namespace, manifestsDirectory)
	if err != nil {
		return err
	}

	issuerUrl, tokenUrl, err := auth.EnsureOAuth2Server(t.resourceManager, t.k8sClient, namespace, t.config, testcontext.GetRetryOpts())
	if err != nil {
		return err
	}
	t.config.IssuerUrl = issuerUrl

	t.oauth2Cfg = &clientcredentials.Config{
		ClientID:     t.config.ClientID,
		ClientSecret: t.config.ClientSecret,
		TokenURL:     tokenUrl,
		AuthStyle:    oauth2.AuthStyleInHeader,
		Scopes:       []string{"read"},
	}

	t.jwtConfig = &clientcredentials.Config{
		ClientID:     t.config.ClientID,
		ClientSecret: t.config.ClientSecret,
		TokenURL:     tokenUrl,
		AuthStyle:    oauth2.AuthStyleInHeader,
	}

	return nil
}

func (t *testsuite) TearDown() {
	err := global.DeleteGlobalResources(t.resourceManager, t.k8sClient, t.namespace, manifestsDirectory)
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
