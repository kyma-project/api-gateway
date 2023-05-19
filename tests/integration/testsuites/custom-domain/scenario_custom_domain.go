package customdomain

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/cucumber/godog"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/helpers"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/jwt"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/manifestprocessor"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/resource"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/testcontext"
	"golang.org/x/oauth2/clientcredentials"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"log"
	"net"
	"path"
	"strings"
	"time"
)

type CustomDomainScenario struct {
	domain          string
	loadBalancerIP  net.IP
	testID          string
	namespace       string
	url             string
	apiResourceOne  []unstructured.Unstructured
	apiResourceTwo  []unstructured.Unstructured
	k8sClient       dynamic.Interface
	oauth2Cfg       *clientcredentials.Config
	httpClient      *helpers.RetryableHttpClient
	resourceManager *resource.Manager
	config          testcontext.TestRunConfig
}

func Init(ctx *godog.ScenarioContext, ts testcontext.Testsuite) {
	scenario, err := CreateCustomDomainScenario(ts, "custom-domain")

	if err != nil {
		log.Fatalf("could not initialize custom domain endpoint err=%s", err)
	}

	ctx.Step(`^there is a "([^"]*)" DNS cloud credentials secret in "([^"]*)" namespace$`, scenario.thereIsAnCloudCredentialsSecret)
	ctx.Step(`^there is an "([^"]*)" service in "([^"]*)" namespace$`, scenario.thereIsAnExposedService)
	ctx.Step(`^create custom domain resources$`, scenario.createResources)
	ctx.Step(`^ensure that DNS record is ready$`, scenario.isDNSReady)
	ctx.Step(`^there is an unsecured endpoint$`, scenario.thereIsAnUnsecuredEndpoint)
	ctx.Step(`^calling the "([^"]*)" endpoint with any token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithAnyTokenShouldResultInStatusBetween)
	ctx.Step(`^calling the "([^"]*)" endpoint without a token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithoutATokenShouldResultInStatusBetween)
	ctx.Step(`^endpoint is secured with OAuth2$`, scenario.secureWithOAuth2)
	ctx.Step(`^calling the "([^"]*)" endpoint with an invalid token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithAInvalidTokenShouldResultInStatusBetween)
	ctx.Step(`^calling the "([^"]*)" endpoint with a valid token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithAValidTokenShouldResultInStatusBetween)
	ctx.Step(`^calling the "([^"]*)" endpoint without a token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithoutATokenShouldResultInStatusBetween)
}

func CreateCustomDomainScenario(t testcontext.Testsuite, namePrefix string) (*CustomDomainScenario, error) {
	ns := t.CommonResources.Namespace
	testID := helpers.GenerateRandomTestId()
	customDomainManifestDirectory := path.Join(testcontext.ManifestsDirectory, "custom-domain")

	// create common resources from files
	commonResources, err := manifestprocessor.ParseFromFileWithTemplate("testing-app.yaml", customDomainManifestDirectory, testcontext.ResourceSeparator, struct {
		Namespace string
		TestID    string
	}{
		Namespace: t.CommonResources.Namespace,
		TestID:    testID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to process common manifest files, details %s", err.Error())
	}
	_, err = t.ResourceManager.CreateResources(t.K8sClient, commonResources...)

	if err != nil {
		return nil, err
	}

	// create api-rule from file
	accessRuleOne, err := manifestprocessor.ParseFromFileWithTemplate("no-access-strategy.yaml", customDomainManifestDirectory, testcontext.ResourceSeparator, struct {
		Namespace        string
		NamePrefix       string
		TestID           string
		Domain           string
		GatewayName      string
		GatewayNamespace string
	}{Namespace: ns, NamePrefix: namePrefix, TestID: testID, Domain: fmt.Sprintf("%s.%s", testID, t.Config.CustomDomain), GatewayName: fmt.Sprintf("%s-%s", namePrefix, testID),
		GatewayNamespace: ns})
	if err != nil {
		return nil, fmt.Errorf("failed to process resource manifest files, details %s", err.Error())
	}
	accessRuleTwo, err := manifestprocessor.ParseFromFileWithTemplate("oauth-strategy.yaml", customDomainManifestDirectory, testcontext.ResourceSeparator, struct {
		Namespace          string
		NamePrefix         string
		TestID             string
		Domain             string
		GatewayName        string
		GatewayNamespace   string
		IssuerUrl          string
		EncodedCredentials string
	}{Namespace: ns, NamePrefix: namePrefix, TestID: testID, Domain: fmt.Sprintf("%s.%s", testID, t.Config.CustomDomain), GatewayName: fmt.Sprintf("%s-%s", namePrefix, testID), GatewayNamespace: ns, IssuerUrl: t.Config.IssuerUrl, EncodedCredentials: base64.RawStdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", t.Config.ClientID, t.Config.ClientSecret)))})
	if err != nil {
		return nil, fmt.Errorf("failed to process resource manifest files, details %s", err.Error())
	}

	return &CustomDomainScenario{
		domain:          t.Config.CustomDomain,
		testID:          testID,
		namespace:       ns,
		url:             fmt.Sprintf("https://httpbin-%s.%s.%s", testID, testID, t.Config.CustomDomain),
		apiResourceOne:  accessRuleOne,
		apiResourceTwo:  accessRuleTwo,
		k8sClient:       t.K8sClient,
		httpClient:      t.HttpClient,
		oauth2Cfg:       t.CommonResources.Oauth2Cfg,
		resourceManager: t.ResourceManager,
		config:          t.Config,
	}, nil
}

func (c *CustomDomainScenario) createResources() error {
	customDomainResources, err := manifestprocessor.ParseFromFileWithTemplate("resources.yaml", "manifests/custom-domain", testcontext.ResourceSeparator, struct {
		Namespace      string
		NamePrefix     string
		TestID         string
		Domain         string
		Subdomain      string
		LoadBalancerIP string
	}{Namespace: c.namespace, NamePrefix: "custom-domain", TestID: c.testID, Domain: c.domain, Subdomain: fmt.Sprintf("%s.%s", c.testID, c.domain), LoadBalancerIP: c.loadBalancerIP.String()})
	if err != nil {
		return fmt.Errorf("failed to process common manifest files, details %s", err.Error())
	}
	_, err = c.resourceManager.CreateResources(c.k8sClient, customDomainResources...)

	if err != nil {
		return err
	}

	return nil
}

func (c *CustomDomainScenario) thereIsAnCloudCredentialsSecret(secretName string, secretNamespace string) error {
	res := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "secrets"}
	_, err := c.k8sClient.Resource(res).Namespace(secretNamespace).Get(context.Background(), secretName, v1.GetOptions{})

	if err != nil {
		return fmt.Errorf("cloud credenials secret could not be found")
	}

	return nil
}

func (c *CustomDomainScenario) isDNSReady() error {
	err := wait.ExponentialBackoff(wait.Backoff{
		Duration: time.Second,
		Factor:   2,
		Steps:    10,
	}, func() (done bool, err error) {
		testName := helpers.GenerateRandomString(3)
		ips, err := net.LookupIP(fmt.Sprintf("%s.%s.%s", testName, c.testID, c.domain))
		if err != nil {
			return false, nil
		}
		if len(ips) != 0 {
			for _, ip := range ips {

				if ip.Equal(c.loadBalancerIP) {
					fmt.Printf("Found %s.%s.%s. IN A %s\n", testName, c.testID, c.domain, ip.String())
					return true, nil
				}
			}
		}
		return false, err
	})
	if err != nil {
		return fmt.Errorf("DNS record could not be looked up: %s", err)
	}
	return nil
}

func (c *CustomDomainScenario) thereIsAnExposedService(svcName string, svcNamespace string) error {
	res := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "services"}
	svc, err := c.k8sClient.Resource(res).Namespace(svcNamespace).Get(context.Background(), svcName, v1.GetOptions{})
	if err != nil {
		return fmt.Errorf("istio-ingressgateway service was not found")
	}

	ingress, found, err := unstructured.NestedSlice(svc.Object, "status", "loadBalancer", "ingress")
	if err != nil || found != true {
		return fmt.Errorf("could not get load balancer status from the service: %s", err)
	}
	loadBalancerIngress, _ := ingress[0].(map[string]interface{})

	loadBalancerIP, err := helpers.GetLoadBalancerIp(loadBalancerIngress)
	if err != nil {
		return fmt.Errorf("could not extract load balancer IP from istio service: %s", err)
	}
	c.loadBalancerIP = loadBalancerIP

	return nil
}

func (c *CustomDomainScenario) thereIsAnUnsecuredEndpoint() error {
	return helpers.ApplyApiRule(c.resourceManager.CreateResources, c.resourceManager.UpdateResources, c.k8sClient, testcontext.GetRetryOpts(c.config), c.apiResourceOne)
}

func (c *CustomDomainScenario) callingTheEndpointWithAnyTokenShouldResultInStatusBetween(endpoint string, arg1, arg2 int) error {
	return c.httpClient.CallEndpointWithHeadersWithRetries(map[string]string{testcontext.AuthorizationHeaderName: testcontext.AnyToken}, fmt.Sprintf("%s/%s", c.url, strings.TrimLeft(endpoint, "/")), &helpers.StatusPredicate{LowerStatusBound: arg1, UpperStatusBound: arg2})
}

func (c *CustomDomainScenario) secureWithOAuth2() error {
	return helpers.ApplyApiRule(c.resourceManager.UpdateResources, c.resourceManager.UpdateResources, c.k8sClient, testcontext.GetRetryOpts(c.config), c.apiResourceTwo)
}

func (c *CustomDomainScenario) callingTheEndpointWithAInvalidTokenShouldResultInStatusBetween(endpoint string, lower int, higher int) error {
	return c.httpClient.CallEndpointWithHeadersWithRetries(map[string]string{testcontext.AuthorizationHeaderName: testcontext.AnyToken}, fmt.Sprintf("%s/%s", c.url, strings.TrimLeft(endpoint, "/")), &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher})
}

func (c *CustomDomainScenario) callingTheEndpointWithAValidTokenShouldResultInStatusBetween(endpoint string, lower int, higher int) error {
	url := fmt.Sprintf("%s/%s", c.url, strings.TrimLeft(endpoint, "/"))
	asserter := &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher}

	requestHeaders := make(map[string]string)

	token, err := jwt.GetAccessToken(c.oauth2Cfg, strings.ToLower("Opaque"))
	if err != nil {
		return fmt.Errorf("failed to fetch an id_token: %s", err.Error())
	}
	requestHeaders[testcontext.OpaqueHeaderName] = token

	return c.httpClient.CallEndpointWithHeadersWithRetries(requestHeaders, url, asserter)
}

func (c *CustomDomainScenario) callingTheEndpointWithoutATokenShouldResultInStatusBetween(endpoint string, lower int, higher int) error {
	return c.httpClient.CallEndpointWithRetries(fmt.Sprintf("%s/%s", c.url, strings.TrimLeft(endpoint, "/")), &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher})
}
