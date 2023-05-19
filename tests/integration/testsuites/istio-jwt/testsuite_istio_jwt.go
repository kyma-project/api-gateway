package istiojwt

import (
	_ "embed"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/helpers"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/jwt"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/manifestprocessor"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/resource"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/testcontext"
	"golang.org/x/oauth2/clientcredentials"
	"k8s.io/client-go/dynamic"
	"path"
	"strings"

	"github.com/cucumber/godog"
)

type tokenFrom struct {
	From     string
	Prefix   string
	AsHeader bool
}

type testsuite struct {
	testcontext.Testsuite
}

func (t *testsuite) createScenario(templateFileName string, scenarioName string) *istioJwtScenario {
	ns := t.CommonResources.Namespace
	testId := helpers.GenerateRandomTestId()

	template := make(map[string]string)
	template["Namespace"] = ns
	template["NamePrefix"] = scenarioName
	template["TestID"] = testId
	template["Domain"] = t.Config.Domain
	template["GatewayName"] = t.Config.GatewayName
	template["GatewayNamespace"] = t.Config.GatewayNamespace
	template["IssuerUrl"] = t.Config.IssuerUrl
	template["EncodedCredentials"] = base64.RawStdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", t.Config.ClientID, t.Config.ClientSecret)))

	// TODO: ADD verification that the template file exists otherwise panic

	return &istioJwtScenario{
		Namespace:               ns,
		TestID:                  testId,
		Domain:                  t.Config.Domain,
		ManifestTemplate:        template,
		ApiResourceManifestPath: templateFileName,
		ApiResourceDirectory:    path.Dir("testsuites/istio-jwt/manifests/"),
		k8sClient:               t.K8sClient,
		oauth2Cfg:               t.CommonResources.Oauth2Cfg,
		httpClient:              t.HttpClient,
		resourceManager:         t.ResourceManager,
		config:                  t.Config,
	}
}

type istioJwtScenario struct {
	Namespace               string
	TestID                  string
	Domain                  string
	ApiResourceManifestPath string
	ApiResourceDirectory    string
	ManifestTemplate        map[string]string
	Url                     string
	k8sClient               dynamic.Interface
	oauth2Cfg               *clientcredentials.Config
	httpClient              *helpers.RetryableHttpClient
	resourceManager         *resource.Manager
	config                  testcontext.TestRunConfig
}

func Init(ctx *godog.ScenarioContext, t testcontext.Testsuite) {
	ts := &testsuite{t}
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
}

func (s *istioJwtScenario) theAPIRuleIsApplied() error {
	r, err := manifestprocessor.ParseFromFileWithTemplate(s.ApiResourceManifestPath, s.ApiResourceDirectory, testcontext.ResourceSeparator, s.ManifestTemplate)
	if err != nil {
		return err
	}
	return helpers.ApplyApiRule(s.resourceManager.CreateResources, s.resourceManager.UpdateResources, s.k8sClient, testcontext.GetRetryOpts(s.config), r)
}

func (s *istioJwtScenario) callingTheEndpointWithAValidToken(endpoint, tokenType, _, _ string, lower, higher int) error {
	asserter := &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher}
	tokenFrom := tokenFrom{
		From:     testcontext.AuthorizationHeaderName,
		Prefix:   testcontext.AuthorizationHeaderPrefix,
		AsHeader: true,
	}
	return s.callingEndpointWithHeadersWithRetries(fmt.Sprintf("%s/%s", s.Url, strings.TrimLeft(endpoint, "/")), tokenType, asserter, nil, &tokenFrom)
}

func (s *istioJwtScenario) callingTheEndpointWithValidTokenShouldResultInStatusBetween(endpoint, tokenType string, lower, higher int) error {
	asserter := &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher}
	tokenFrom := tokenFrom{
		From:     testcontext.AuthorizationHeaderName,
		Prefix:   testcontext.AuthorizationHeaderPrefix,
		AsHeader: true,
	}
	return s.callingEndpointWithHeadersWithRetries(fmt.Sprintf("%s/%s", s.Url, strings.TrimLeft(endpoint, "/")), tokenType, asserter, nil, &tokenFrom)
}

func (s *istioJwtScenario) callingTheEndpointWithValidTokenFromHeaderShouldResultInStatusBetween(endpoint, tokenType string, fromHeader string, prefix string, lower, higher int) error {
	asserter := &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher}
	tokenFrom := tokenFrom{
		From:     fromHeader,
		Prefix:   prefix,
		AsHeader: true,
	}
	return s.callingEndpointWithHeadersWithRetries(fmt.Sprintf("%s/%s", s.Url, strings.TrimLeft(endpoint, "/")), tokenType, asserter, nil, &tokenFrom)
}

func (s *istioJwtScenario) callingTheEndpointWithValidTokenFromParameterShouldResultInStatusBetween(endpoint, tokenType string, fromParameter string, lower, higher int) error {
	asserter := &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher}
	tokenFrom := tokenFrom{
		From:     fromParameter,
		AsHeader: false,
	}
	return s.callingEndpointWithHeadersWithRetries(fmt.Sprintf("%s/%s", s.Url, strings.TrimLeft(endpoint, "/")), tokenType, asserter, nil, &tokenFrom)
}

func (s *istioJwtScenario) callingTheEndpointWithValidTokenShouldResultInBodyContaining(endpoint, tokenType string, bodyContent string) error {
	asserter := &helpers.BodyContainsPredicate{Expected: []string{bodyContent}}
	tokenFrom := tokenFrom{
		From:     testcontext.AuthorizationHeaderName,
		Prefix:   testcontext.AuthorizationHeaderPrefix,
		AsHeader: true,
	}
	return s.callingEndpointWithHeadersWithRetries(fmt.Sprintf("%s/%s", s.Url, strings.TrimLeft(endpoint, "/")), tokenType, asserter, nil, &tokenFrom)
}

func (s *istioJwtScenario) callingEndpointWithHeadersWithRetries(url string, tokenType string, asserter helpers.HttpResponseAsserter, requestHeaders map[string]string, tokenFrom *tokenFrom) error {
	if requestHeaders == nil {
		requestHeaders = make(map[string]string)
	}

	switch tokenType {
	case "Opaque":
		token, err := jwt.GetAccessToken(*s.oauth2Cfg, strings.ToLower(tokenType))
		if err != nil {
			return fmt.Errorf("failed to fetch an id_token: %s", err.Error())
		}
		requestHeaders[testcontext.OpaqueHeaderName] = token
	case "JWT":
		token, err := jwt.GetAccessToken(*s.oauth2Cfg, strings.ToLower(tokenType))
		if err != nil {
			return fmt.Errorf("failed to fetch an id_token: %s", err.Error())
		}
		if tokenFrom.From == "" {
			return errors.New("jwt from header or parameter name not specified")
		}
		if tokenFrom.AsHeader {
			if tokenFrom.Prefix != "" {
				token = fmt.Sprintf("%s %s", tokenFrom.Prefix, token)
			}
			requestHeaders[tokenFrom.From] = token
		} else {
			url = fmt.Sprintf("%s?%s=%s", url, tokenFrom.From, token)
		}
	default:
		return fmt.Errorf("unsupported token type: %s", tokenType)
	}

	return s.httpClient.CallEndpointWithHeadersWithRetries(requestHeaders, url, asserter)
}

func (s *istioJwtScenario) callingTheEndpointWithoutTokenShouldResultInStatusBetween(endpoint string, lower, higher int) error {
	return s.httpClient.CallEndpointWithRetries(fmt.Sprintf("%s/%s", s.Url, strings.TrimLeft(endpoint, "/")), &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher})
}

func (s *istioJwtScenario) thereAreTwoNamespaces() error {
	resources, err := manifestprocessor.ParseFromFileWithTemplate("second-namespace.yaml", s.ApiResourceDirectory, testcontext.ResourceSeparator, s.ManifestTemplate)
	if err != nil {
		return err
	}
	_, err = s.resourceManager.CreateResources(s.k8sClient, resources...)
	return err
}

func (s *istioJwtScenario) thereIsAnJwtSecuredPath(path string) {
	s.ManifestTemplate["jwtSecuredPath"] = path
}

func (s *istioJwtScenario) emptyStep() {
}

func (s *istioJwtScenario) thereIsAHttpbinService() error {
	resources, err := manifestprocessor.ParseFromFileWithTemplate("testing-app.yaml", s.ApiResourceDirectory, testcontext.ResourceSeparator, s.ManifestTemplate)
	if err != nil {
		return err
	}
	_, err = s.resourceManager.CreateResources(s.k8sClient, resources...)
	if err != nil {
		return err
	}

	s.Url = fmt.Sprintf("https://httpbin-%s.%s", s.TestID, s.Domain)

	return nil
}

// teardownHttpbinService deletes the httpbin service and reset the url in the scenario. This should be considered a temporary solution
// to reduce resource conumption until we implement a better way to clean up the resources by a scenario. If the test fails before this step the teardown won't be executed.
func (s *istioJwtScenario) teardownHttpbinService() error {
	resources, err := manifestprocessor.ParseFromFileWithTemplate("testing-app.yaml", s.ApiResourceDirectory, testcontext.ResourceSeparator, s.ManifestTemplate)
	if err != nil {
		return err
	}
	err = s.resourceManager.DeleteResources(s.k8sClient, resources...)
	if err != nil {
		return err
	}

	s.Url = ""

	return nil
}
