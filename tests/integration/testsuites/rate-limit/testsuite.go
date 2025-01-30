package ratelimit

import (
	_ "embed"
	"fmt"
	"github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/manifestprocessor"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"log"
	"path"

	"github.com/cucumber/godog"
	"k8s.io/client-go/dynamic"

	"github.com/kyma-project/api-gateway/tests/integration/pkg/global"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/helpers"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/resource"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/testcontext"
)

const manifestsDirectory = "testsuites/rate-limit/manifests/"

func (t *testsuite) createScenario() *scenario {
	testId := helpers.GenerateRandomTestId()

	template := make(map[string]string)
	template["Namespace"] = t.namespace
	template["TestID"] = testId
	template["Domain"] = t.config.Domain
	template["GatewayName"] = t.config.GatewayName
	template["GatewayNamespace"] = t.config.GatewayNamespace

	return &scenario{
		Namespace:            t.namespace,
		TestID:               testId,
		Domain:               t.config.Domain,
		ManifestTemplate:     template,
		ApiResourceDirectory: path.Dir(manifestsDirectory),
		k8sClient:            t.K8sClient(),
		httpClient:           t.httpClient,
		resourceManager:      t.ResourceManager(),
		config:               t.config,
	}
}

type testsuite struct {
	name            string
	namespace       string
	httpClient      *helpers.RetryableHttpClient
	k8sClient       dynamic.Interface
	resourceManager *resource.Manager
	config          testcontext.Config
}

func (t *testsuite) InitScenarios(ctx *godog.ScenarioContext) {
	initScenario(ctx, t)
}

func (t *testsuite) FeaturePath() []string {
	return []string{"testsuites/rate-limit/features/"}
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
	t.namespace = global.GenerateNamespaceName(t.name)
	log.Printf("Using namespace: %s", t.namespace)
	err := global.CreateGlobalResources(t.resourceManager, t.k8sClient, t.namespace, manifestsDirectory)
	if err != nil {
		return err
	}
	apiGateway, err := manifestprocessor.ParseFromFileWithTemplate("api-gateway.yaml", manifestsDirectory, struct{}{})
	if err != nil {
		return err
	}

	_, err = t.resourceManager.CreateResourcesWithoutNS(t.k8sClient, apiGateway...)
	if err != nil {
		return fmt.Errorf("could not create APIGateway, details %s", err.Error())
	}

	return nil
}

func (t *testsuite) TearDown() {
	err := global.DeleteGlobalResources(t.resourceManager, t.k8sClient, t.namespace, manifestsDirectory)
	if err != nil {
		log.Print(err.Error())
	}

	err = t.resourceManager.DeleteResourceWithoutNS(t.k8sClient, schema.GroupVersionResource{
		Group:    v1alpha1.GroupVersion.Group,
		Version:  v1alpha1.GroupVersion.Version,
		Resource: "apigateways",
	}, "default")
	if err != nil {
		log.Printf("Could not remove APIGateway, details: %s,", err.Error())
	}
}

func (t *testsuite) BeforeSuiteHooks() []func() error {
	return []func() error{}
}

func (t *testsuite) AfterSuiteHooks() []func() error {
	return []func() error{}
}

func NewTestsuite(httpClient *helpers.RetryableHttpClient, k8sClient dynamic.Interface, rm *resource.Manager, config testcontext.Config) testcontext.Testsuite {
	return &testsuite{
		name:            "rate-limit",
		httpClient:      httpClient,
		k8sClient:       k8sClient,
		resourceManager: rm,
		config:          config,
	}
}
