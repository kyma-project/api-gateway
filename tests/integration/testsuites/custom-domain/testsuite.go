package customdomain

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/cucumber/godog"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/auth"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/helpers"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/hooks"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/manifestprocessor"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/resource"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/testcontext"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"log"
)

type testsuite struct {
	name            string
	namespace       string
	httpClient      *helpers.RetryableHttpClient
	k8sClient       dynamic.Interface
	resourceManager *resource.Manager
	config          testcontext.Config
	oauth2Cfg       *clientcredentials.Config
}

func (t *testsuite) InitScenarios(ctx *godog.ScenarioContext) {
	initScenario(ctx, t)
}

func (t *testsuite) FeaturePath() []string {
	return []string{"testsuites/custom-domain/features/"}
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
	namespace := fmt.Sprintf("%s-%s", t.name, helpers.GenerateRandomString())
	log.Printf("Using namespace: %s\n", namespace)

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

	oauth2Cfg := &clientcredentials.Config{
		ClientID:     t.config.ClientID,
		ClientSecret: t.config.ClientSecret,
		TokenURL:     tokenURL,
		AuthStyle:    oauth2.AuthStyleInHeader,
	}

	// create common resources for all scenarios
	globalCommonResources, err := manifestprocessor.ParseFromFileWithTemplate("global-commons.yaml", manifestsPath, struct {
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

	t.oauth2Cfg = oauth2Cfg
	t.namespace = namespace

	return nil
}

func (t *testsuite) TearDown() {
	//Remove certificate
	res := schema.GroupVersionResource{Group: "cert.gardener.cloud", Version: "v1alpha1", Resource: "certificates"}
	err := t.k8sClient.Resource(res).Namespace("istio-system").DeleteCollection(context.Background(), v1.DeleteOptions{}, v1.ListOptions{LabelSelector: "owner=custom-domain-test"})
	if err != nil {
		log.Print(err.Error())
	}

	res = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "namespaces"}
	err = t.k8sClient.Resource(res).Delete(context.Background(), t.namespace, v1.DeleteOptions{})
	if err != nil {
		log.Print(err.Error())
	}
}

func (t *testsuite) BeforeSuiteHooks() []func() error {
	return []func() error{hooks.ApplyAndVerifyApiGatewayCrSuiteHook}
}

func (t *testsuite) AfterSuiteHooks() []func() error {
	return []func() error{hooks.DeleteBlockingResourcesSuiteHook, hooks.ApiGatewayCrTearDownSuiteHook}
}

func NewTestsuite(httpClient *helpers.RetryableHttpClient, k8sClient dynamic.Interface, rm *resource.Manager, config testcontext.Config) testcontext.Testsuite {
	return &testsuite{
		name:            "custom-domain",
		httpClient:      httpClient,
		k8sClient:       k8sClient,
		resourceManager: rm,
		config:          config,
	}
}
