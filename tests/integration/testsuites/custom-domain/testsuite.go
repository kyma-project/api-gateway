package customdomain

import (
	"context"
	_ "embed"
	"encoding/base64"
	"fmt"
	"github.com/avast/retry-go/v4"
	"github.com/cucumber/godog"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/auth"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/global"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/helpers"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/hooks"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/manifestprocessor"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/network"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/resource"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/testcontext"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"log"
	"time"
)

const ingressServiceName = "istio-ingressgateway"
const ingressNamespaceName = "istio-system"

type testsuite struct {
	name                  string
	namespace             string
	httpClient            *helpers.RetryableHttpClient
	k8sClient             dynamic.Interface
	resourceManager       *resource.Manager
	config                testcontext.Config
	oauth2Cfg             *clientcredentials.Config
	suiteID               string
	subdomainForTests     string
	customDomainResources []unstructured.Unstructured
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

func (t *testsuite) ValidateAndFixConfig() error {
	t.config.EnforceGardener()

	err := t.config.ValidateCommon(t.resourceManager, t.k8sClient)
	if err != nil {
		return err
	}

	err = t.config.RequireGCPServiceAccount()
	if err != nil {
		return err
	}

	err = t.config.RequireCustomDomain()
	if err != nil {
		return err
	}

	return nil
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

	err := global.CreateGlobalResources(t.resourceManager, t.k8sClient, namespace, manifestsPath)
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

	t.suiteID = helpers.GenerateRandomString(8)
	log.Printf("Using suite ID: %s", t.suiteID)

	t.subdomainForTests = fmt.Sprintf("%s.%s", t.suiteID, t.config.CustomDomain)
	log.Printf("Subdomain to be used by tests: %s", t.subdomainForTests)

	return nil
}

func (t *testsuite) TearDown() {
	//Remove certificate
	res := schema.GroupVersionResource{Group: "cert.gardener.cloud", Version: "v1alpha1", Resource: "certificates"}
	err := t.k8sClient.Resource(res).Namespace("istio-system").DeleteCollection(context.Background(), v1.DeleteOptions{}, v1.ListOptions{LabelSelector: "owner=custom-domain-test"})
	if err != nil {
		log.Print(err.Error())
	}

	err = global.DeleteGlobalResources(t.resourceManager, t.k8sClient, t.namespace, manifestsPath)
	if err != nil {
		log.Print(err.Error())
	}
}

func (t *testsuite) BeforeSuiteHooks() []func() error {
	return []func() error{hooks.ApplyAndVerifyApiGatewayCrSuiteHook, t.createCustomDomainResources}
}

func (t *testsuite) AfterSuiteHooks() []func() error {
	return []func() error{t.deleteCustomDomainResources, hooks.DeleteBlockingResourcesSuiteHook, hooks.ApiGatewayCrTearDownSuiteHook}
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

func (t *testsuite) createCustomDomainResources() error {
	log.Printf("Preparing custom domain resources")

	log.Printf("Loading GCP SA json")
	gcpSAJson, err := helpers.LoadFile(t.config.GCPServiceAccountJsonPath)
	if err != nil {
		return fmt.Errorf("can't read GCP SA Account json file because of: %w", err)
	}
	log.Printf("GCP SA json loaded")

	log.Printf("Looking for ingress service and its load balancer data")
	ingress, err := helpers.GetLoadBalancerIngress(t.K8sClient(), ingressServiceName, ingressNamespaceName)
	if err != nil {
		return fmt.Errorf("can't get load balancer ingress because of: %w", err)
	}
	loadBalancerIP, err := helpers.GetLoadBalancerIp(ingress)
	if err != nil {
		return fmt.Errorf("can't determine load balancer IP: %w", err)
	}
	log.Printf("Determined load balancer IP or name: %s", loadBalancerIP)

	log.Printf("Creating custom domain resources")
	customDomainResources, err := manifestprocessor.ParseFromFileWithTemplate("resources.yaml", manifestsPath, struct {
		Namespace            string
		NamePrefix           string
		SuiteID              string
		ParentDomain         string
		Subdomain            string
		LoadBalancerIP       string
		EncodedSACredentials string
	}{
		Namespace:            t.namespace,
		NamePrefix:           "custom-domain",
		SuiteID:              t.suiteID,
		ParentDomain:         t.config.CustomDomain,
		Subdomain:            t.subdomainForTests,
		LoadBalancerIP:       loadBalancerIP.String(),
		EncodedSACredentials: base64.StdEncoding.EncodeToString(gcpSAJson),
	})
	if err != nil {
		return fmt.Errorf("failed to process common manifest files, details %s", err.Error())
	}
	t.customDomainResources = customDomainResources
	_, err = t.resourceManager.CreateResources(t.k8sClient, customDomainResources...)

	if err != nil {
		return err
	}
	log.Printf("Custom domain resources created")

	log.Printf("Waiting until the wildcard DNS record %s points to IP %s", t.subdomainForTests, loadBalancerIP)
	err = network.WaitUntilDNSReady(t.subdomainForTests, loadBalancerIP, t.getDomainRetryOpts())
	if err != nil {
		return err
	}
	log.Printf("DNS record is ready")

	return nil
}

func (t *testsuite) deleteCustomDomainResources() error {
	log.Printf("Deleting custom domain resources")
	err := t.resourceManager.DeleteResources(t.k8sClient, t.customDomainResources...)
	if err != nil {
		return err
	}
	log.Printf("Custom domain resources deleted")
	return nil
}

func (t *testsuite) getDomainRetryOpts() []retry.Option {
	retryOpts := []retry.Option{
		retry.Delay(time.Duration(10) * time.Second),
		retry.Attempts(100),
		retry.DelayType(retry.FixedDelay),
	}

	if t.config.DebugLogging {
		retryOpts = append(retryOpts, retry.OnRetry(func(n uint, err error) { log.Printf("Trial #%d failed, error: %s\n", n, err) }))
	}

	return retryOpts
}
