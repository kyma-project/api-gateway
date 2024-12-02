package istiojwt

import (
	_ "embed"
	"encoding/base64"
	"fmt"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/global"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/hooks"
	"log"
	"path"

	"github.com/cucumber/godog"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/auth"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/helpers"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/resource"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/testcontext"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	"k8s.io/client-go/dynamic"
)

const manifestsDirectory = "testsuites/istio-jwt/manifests/"

type tokenFrom struct {
	From     string
	Prefix   string
	AsHeader bool
}

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
	template["IstioNamespace"] = t.config.IstioNamespace

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
	}
}

type testsuite struct {
	name            string
	namespace       string
	secondNamespace string
	httpClient      *helpers.RetryableHttpClient
	k8sClient       dynamic.Interface
	resourceManager *resource.Manager
	config          testcontext.Config
	oauth2Cfg       *clientcredentials.Config
}

func (t *testsuite) InitScenarios(ctx *godog.ScenarioContext) {
	initCommon(ctx, t)
	initPrefix(ctx, t)
	initRegex(ctx, t)
	initRequiredScopes(ctx, t)
	initAudience(ctx, t)
	initJwtAndAllow(ctx, t)
	initJwtAndOauth(ctx, t)
	initJwtAndNoAuth(ctx, t)
	initJwtTwoNamespaces(ctx, t)
	initJwtServiceFallback(ctx, t)
	initDiffServiceSameMethods(ctx, t)
	initJwtUnavailableIssuer(ctx, t)
	initJwtIssuerJwksNotMatch(ctx, t)
	initMutatorCookie(ctx, t)
	initMutatorHeader(ctx, t)
	initMultipleMutators(ctx, t)
	initMutatorsOverwrite(ctx, t)
	initTokenFromHeaders(ctx, t)
	initTokenFromParams(ctx, t)
	initCustomLabelSelector(ctx, t)
	initCustomCors(ctx, t)
	initDefaultCors(ctx, t)
	initExposeMethodsOnPathsAllowHandler(ctx, t)
	initExposeMethodsOnPathsNoAuthHandler(ctx, t)
	initExposeMethodsOnPathsNoopHandler(ctx, t)
	initExposeMethodsOnPathsJwtHandler(ctx, t)
	initExposeMethodsOnPathsOAuth2Handler(ctx, t)
	initv2alpha1IstioJWT(ctx, t)
	initv2alpha1NoAuthHandler(ctx, t)
	initv2alpha1NoAuthHandlerRecover(ctx, t)
}

func (t *testsuite) FeaturePath() []string {
	return []string{"testsuites/istio-jwt/features/"}
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

	secondNamespace := fmt.Sprintf("%s-2", namespace)
	t.secondNamespace = secondNamespace

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
	h := []func() error{hooks.IstioSkipVerifyJwksResolverSuiteHook(t), hooks.ApplyAndVerifyApiGatewayCrSuiteHook}

	if !t.config.IsGardener {
		h = append(h, hooks.DnsPatchForK3dSuiteHook(t))
	}
	return h
}

func (t *testsuite) AfterSuiteHooks() []func() error {
	h := []func() error{hooks.IstioSkipVerifyJwksResolverSuiteHookTeardown(t), hooks.DeleteBlockingResourcesSuiteHook, hooks.ApiGatewayCrTearDownSuiteHook}

	if !t.config.IsGardener {
		h = append(h, hooks.DnsPatchForK3dSuiteHookTeardown(t))
	}
	return h
}

func NewTestsuite(httpClient *helpers.RetryableHttpClient, k8sClient dynamic.Interface, rm *resource.Manager, config testcontext.Config) testcontext.Testsuite {
	return &testsuite{
		name:            "istio-jwt",
		httpClient:      httpClient,
		k8sClient:       k8sClient,
		resourceManager: rm,
		config:          config,
	}
}
