package customdomain

import (
	"encoding/base64"
	"fmt"
	"github.com/cucumber/godog"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/auth"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/helpers"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/manifestprocessor"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/resource"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/testcontext"
	"golang.org/x/oauth2/clientcredentials"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"log"
	"path"
	"strings"
)

const manifestsPath = "testsuites/custom-domain/manifests/"

type scenario struct {
	parentDomainForTests string
	subdomainForTests    string
	testID               string
	suiteID              string
	namespace            string
	url                  string
	apiResourceOne       unstructured.Unstructured
	apiResourceTwo       unstructured.Unstructured
	k8sClient            dynamic.Interface
	oauth2Cfg            *clientcredentials.Config
	httpClient           *helpers.RetryableHttpClient
	resourceManager      *resource.Manager
	config               testcontext.Config
}

func initScenario(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario, err := createScenario(ts, "custom-domain")

	if err != nil {
		log.Fatalf("could not initialize custom domain endpoint err=%s", err)
	}

	ctx.Step(`^there is an unsecured endpoint$`, scenario.thereIsAnUnsecuredEndpoint)
	ctx.Step(`^calling the "([^"]*)" endpoint with any token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithAnyTokenShouldResultInStatusBetween)
	ctx.Step(`^calling the "([^"]*)" endpoint without a token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithoutATokenShouldResultInStatusBetween)
	ctx.Step(`^endpoint is secured with OAuth2$`, scenario.secureWithOAuth2)
	ctx.Step(`^calling the "([^"]*)" endpoint with an invalid token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithAInvalidTokenShouldResultInStatusBetween)
	ctx.Step(`^calling the "([^"]*)" endpoint with a valid token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithAValidTokenShouldResultInStatusBetween)
	ctx.Step(`^there is a httpbin service$`, scenario.thereIsAHttpbinService)
	ctx.Step(`^teardown httpbin service$`, scenario.teardownHttpbinService)
}

func createScenario(t *testsuite, namePrefix string) (*scenario, error) {
	ns := t.namespace
	testID := helpers.GenerateRandomTestId()
	customDomainManifestDirectory := path.Dir(manifestsPath)

	// create api-rule from file
	accessRuleOne, err := manifestprocessor.ParseSingleEntryFromFileWithTemplate("no-access-strategy.yaml", customDomainManifestDirectory, struct {
		Namespace        string
		NamePrefix       string
		TestID           string
		Subdomain        string
		GatewayName      string
		GatewayNamespace string
	}{
		Namespace:        ns,
		NamePrefix:       namePrefix,
		TestID:           testID,
		Subdomain:        t.subdomainForTests,
		GatewayName:      fmt.Sprintf("%s-%s", namePrefix, t.suiteID),
		GatewayNamespace: ns,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to process resource manifest files, details %s", err.Error())
	}
	accessRuleTwo, err := manifestprocessor.ParseSingleEntryFromFileWithTemplate("oauth-strategy.yaml", customDomainManifestDirectory, struct {
		Namespace          string
		NamePrefix         string
		TestID             string
		Subdomain          string
		GatewayName        string
		GatewayNamespace   string
		IssuerUrl          string
		EncodedCredentials string
	}{
		Namespace:          ns,
		NamePrefix:         namePrefix,
		TestID:             testID,
		Subdomain:          t.subdomainForTests,
		GatewayName:        fmt.Sprintf("%s-%s", namePrefix, t.suiteID),
		GatewayNamespace:   ns,
		IssuerUrl:          t.config.IssuerUrl,
		EncodedCredentials: base64.RawStdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", t.config.ClientID, t.config.ClientSecret))),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to process resource manifest files, details %s", err.Error())
	}

	return &scenario{
		parentDomainForTests: t.config.CustomDomain,
		subdomainForTests:    t.subdomainForTests,
		testID:               testID,
		suiteID:              t.suiteID,
		namespace:            ns,
		apiResourceOne:       accessRuleOne,
		apiResourceTwo:       accessRuleTwo,
		k8sClient:            t.k8sClient,
		httpClient:           t.httpClient,
		oauth2Cfg:            t.oauth2Cfg,
		resourceManager:      t.resourceManager,
		config:               t.config,
	}, nil
}

func (c *scenario) thereIsAnUnsecuredEndpoint() error {
	return helpers.CreateApiRule(c.resourceManager, c.k8sClient, testcontext.GetRetryOpts(), c.apiResourceOne)
}

func (c *scenario) callingTheEndpointWithAnyTokenShouldResultInStatusBetween(endpoint string, arg1, arg2 int) error {
	return c.httpClient.CallEndpointWithHeadersWithRetries(map[string]string{testcontext.AuthorizationHeaderName: testcontext.AnyToken}, fmt.Sprintf("%s/%s", c.url, strings.TrimLeft(endpoint, "/")), &helpers.StatusPredicate{LowerStatusBound: arg1, UpperStatusBound: arg2})
}

func (c *scenario) secureWithOAuth2() error {
	return helpers.CreateApiRule(c.resourceManager, c.k8sClient, testcontext.GetRetryOpts(), c.apiResourceTwo)
}

func (c *scenario) callingTheEndpointWithAInvalidTokenShouldResultInStatusBetween(endpoint string, lower int, higher int) error {
	return c.httpClient.CallEndpointWithHeadersWithRetries(map[string]string{testcontext.AuthorizationHeaderName: testcontext.AnyToken}, fmt.Sprintf("%s/%s", c.url, strings.TrimLeft(endpoint, "/")), &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher})
}

func (c *scenario) callingTheEndpointWithAValidTokenShouldResultInStatusBetween(endpoint string, lower int, higher int) error {
	url := fmt.Sprintf("%s/%s", c.url, strings.TrimLeft(endpoint, "/"))
	asserter := &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher}

	requestHeaders := make(map[string]string)

	token, err := auth.GetAccessTokenWithRetries(*c.oauth2Cfg, strings.ToLower("Opaque"), testcontext.GetRetryOpts())
	if err != nil {
		return fmt.Errorf("failed to fetch an id_token: %s", err.Error())
	}
	requestHeaders[testcontext.OpaqueHeaderName] = token

	return c.httpClient.CallEndpointWithHeadersWithRetries(requestHeaders, url, asserter)
}

func (c *scenario) callingTheEndpointWithoutATokenShouldResultInStatusBetween(endpoint string, lower int, higher int) error {
	return c.httpClient.CallEndpointWithRetries(fmt.Sprintf("%s/%s", c.url, strings.TrimLeft(endpoint, "/")), &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher})
}

func (s *scenario) thereIsAHttpbinService() error {
	customDomainManifestDirectory := path.Dir(manifestsPath)

	resources, err := manifestprocessor.ParseFromFileWithTemplate("testing-app.yaml", customDomainManifestDirectory, struct {
		Namespace string
		TestID    string
	}{
		Namespace: s.namespace,
		TestID:    s.testID,
	})
	if err != nil {
		return err
	}
	_, err = s.resourceManager.CreateResources(s.k8sClient, resources...)
	if err != nil {
		return err
	}

	s.url = fmt.Sprintf("https://httpbin-%s.%s", s.testID, s.subdomainForTests)

	return nil
}

// teardownHttpbinService deletes the httpbin service and reset the url in the scenario. This should be considered a temporary solution
// to reduce resource conumption until we implement a better way to clean up the resources by a scenario. If the test fails before this step the teardown won't be executed.
func (s *scenario) teardownHttpbinService() error {
	customDomainManifestDirectory := path.Dir(manifestsPath)

	resources, err := manifestprocessor.ParseFromFileWithTemplate("testing-app.yaml", customDomainManifestDirectory, struct {
		Namespace string
		TestID    string
	}{
		Namespace: s.namespace,
		TestID:    s.testID,
	})
	if err != nil {
		return err
	}
	err = s.resourceManager.DeleteResources(s.k8sClient, resources...)
	if err != nil {
		return err
	}

	s.url = ""

	return nil
}
