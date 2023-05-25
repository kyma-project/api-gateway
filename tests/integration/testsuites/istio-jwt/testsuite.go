package istiojwt

import (
	"context"
	_ "embed"
	"encoding/base64"
	"errors"
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

func Init(ctx *godog.ScenarioContext, test testcontext.Testsuite) error {
	ts, ok := test.(*testsuite)

	if !ok {
		return errors.New("test suite is not istio jwt")
	}

	initCommon(ctx, ts)
	initPrefix(ctx, ts)
	initRegex(ctx, ts)
	initRequiredScopes(ctx, ts)
	initAudience(ctx, ts)
	initJwtAndAllow(ctx, ts)
	initJwtAndOauth(ctx, ts)
	initJwtTwoNamespaces(ctx, ts)
	initJwtServiceFallback(ctx, ts)
	initDiffServiceSameMethods(ctx, ts)
	initJwtUnavailableIssuer(ctx, ts)
	initJwtIssuerJwksNotMatch(ctx, ts)
	initMutatorCookie(ctx, ts)
	initMutatorHeader(ctx, ts)
	initMultipleMutators(ctx, ts)
	initMutatorsOverwrite(ctx, ts)
	initTokenFromHeaders(ctx, ts)
	initTokenFromParams(ctx, ts)
	initCustomLabelSelector(ctx, ts)

	return nil
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

func (t *testsuite) Name() string {
	return t.name
}

func (t *testsuite) ResourceManager() *resource.Manager {
	return t.resourceManager
}

func (t *testsuite) K8sClient() dynamic.Interface {
	return t.k8sClient
}

func (t *testsuite) Setup() {
	namespace := fmt.Sprintf("%s-%s", t.name, helpers.GenerateRandomString(6))
	secondNamespace := fmt.Sprintf("%s-2", namespace)
	log.Printf("Using namespace: %s\n", namespace)

	oauth2Cfg := &clientcredentials.Config{
		ClientID:     t.config.ClientID,
		ClientSecret: t.config.ClientSecret,
		TokenURL:     fmt.Sprintf("%s/oauth2/token", t.config.IssuerUrl),
		AuthStyle:    oauth2.AuthStyleInHeader,
	}

	// create common resources for all scenarios
	globalCommonResources, err := manifestprocessor.ParseFromFileWithTemplate("global-commons.yaml", manifestsDirectory, struct {
		Namespace string
	}{
		Namespace: namespace,
	})
	if err != nil {
		log.Fatal(err)
	}

	// delete test namespace if the previous test namespace persists
	nsResourceSchema, ns, name := t.resourceManager.GetResourceSchemaAndNamespace(globalCommonResources[0])
	log.Printf("Delete test namespace, if exists: %s\n", name)
	err = t.resourceManager.DeleteResource(t.k8sClient, nsResourceSchema, ns, name)
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(time.Duration(t.config.ReqDelay) * time.Second)

	log.Printf("Creating common tests resources")
	_, err = t.resourceManager.CreateResources(t.k8sClient, globalCommonResources...)
	if err != nil {
		log.Fatal(err)
	}

	t.oauth2Cfg = oauth2Cfg
	t.namespace = namespace
	t.secondNamespace = secondNamespace
}

func (t *testsuite) TearDown() {
	res := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "namespaces"}
	err := t.k8sClient.Resource(res).Delete(context.Background(), t.namespace, v1.DeleteOptions{})
	if err != nil {
		log.Print(err.Error())
	}

	err = t.k8sClient.Resource(res).Delete(context.Background(), t.secondNamespace, v1.DeleteOptions{})
	if err != nil {
		log.Print(err.Error())
	}
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
