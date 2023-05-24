package ory

import (
	_ "embed"
	"fmt"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/helpers"
	jwt "github.com/kyma-project/api-gateway/tests/integration/pkg/jwt"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/manifestprocessor"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/resource"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/testcontext"
	"golang.org/x/oauth2/clientcredentials"
	"k8s.io/client-go/dynamic"
	"strings"
)

type scenario struct {
	Namespace               string
	TestID                  string
	Domain                  string
	ApiResourceManifestPath string
	ApiResourceDirectory    string
	ManifestTemplate        map[string]string
	Url                     string
	k8sClient               dynamic.Interface
	oauth2Cfg               *clientcredentials.Config
	jwtConfig               *clientcredentials.Config
	httpClient              *helpers.RetryableHttpClient
	resourceManager         *resource.Manager
	config                  testcontext.Config
}

func (s *scenario) callingTheEndpointWithValidTokenShouldResultInStatusBetween(path string, tokenType string, lower, higher int) error {
	asserter := &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher}
	return s.callingTheEndpointWithValidToken(fmt.Sprintf("%s%s", s.Url, path), tokenType, asserter)
}

func (s *scenario) callingTheEndpointWithValidToken(url string, tokenType string, asserter helpers.HttpResponseAsserter) error {

	requestHeaders := make(map[string]string)

	switch tokenType {
	case "JWT":
		tokenJwt, err := jwt.GetAccessToken(*s.jwtConfig, strings.ToLower(tokenType))
		if err != nil {
			return fmt.Errorf("failed to fetch an id_token: %s", err.Error())
		}
		requestHeaders[testcontext.AuthorizationHeaderName] = fmt.Sprintf("Bearer %s", tokenJwt)

	case "OAuth2":
		tokenOauth, err := jwt.GetHydraOAuth2AccessToken(*s.oauth2Cfg)
		if err != nil {
			return err
		}
		requestHeaders[testcontext.AuthorizationHeaderName] = fmt.Sprintf("Bearer %s", tokenOauth)
	default:
		return fmt.Errorf("unsupported token type: %s", tokenType)
	}

	return s.httpClient.CallEndpointWithHeadersWithRetries(requestHeaders, url, asserter)
}

func (s *scenario) thereIsAnOauth2Endpoint() error {
	err := s.thereIsAHttpbinService()
	if err != nil {
		return err
	}

	r, err := manifestprocessor.ParseFromFileWithTemplate(s.ApiResourceManifestPath, s.ApiResourceDirectory, s.ManifestTemplate)
	if err != nil {
		return err
	}
	return helpers.ApplyApiRule(s.resourceManager.CreateResources, s.resourceManager.UpdateResources, s.k8sClient, testcontext.GetRetryOpts(s.config), r)
}

func (s *scenario) theManifestIsApplied() error {
	r, err := manifestprocessor.ParseFromFileWithTemplate(s.ApiResourceManifestPath, s.ApiResourceDirectory, s.ManifestTemplate)
	if err != nil {
		return err
	}
	return helpers.ApplyApiRule(s.resourceManager.CreateResources, s.resourceManager.UpdateResources, s.k8sClient, testcontext.GetRetryOpts(s.config), r)
}

func (s *scenario) callingTheEndpointWithInvalidTokenShouldResultInStatusBetween(path string, lower, higher int) error {
	requestHeaders := map[string]string{testcontext.AuthorizationHeaderName: testcontext.AnyToken}
	return s.httpClient.CallEndpointWithHeadersWithRetries(requestHeaders, fmt.Sprintf("%s%s", s.Url, path), &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher})
}

func (s *scenario) callingTheEndpointWithoutTokenShouldResultInStatusBetween(path string, lower, higher int) error {
	return s.httpClient.CallEndpointWithRetries(fmt.Sprintf("%s/%s", s.Url, strings.TrimLeft(path, "/")), &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher})
}

func (s *scenario) thereIsAHttpbinService() error {
	resources, err := manifestprocessor.ParseFromFileWithTemplate("testing-app.yaml", s.ApiResourceDirectory, s.ManifestTemplate)
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
func (s *scenario) teardownHttpbinService() error {
	resources, err := manifestprocessor.ParseFromFileWithTemplate("testing-app.yaml", s.ApiResourceDirectory, s.ManifestTemplate)
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

func (s *scenario) callingTheEndpointsWithInvalidTokenShouldResultInStatusBetween(path1, path2 string, lower, higher int) error {
	err := s.callingTheEndpointWithInvalidTokenShouldResultInStatusBetween(path1, lower, higher)
	if err != nil {
		return err
	}
	return s.callingTheEndpointWithInvalidTokenShouldResultInStatusBetween(path2, lower, higher)
}

func (s *scenario) callingTheEndpointsWithoutTokenShouldResultInStatusBetween(path1, path2 string, lower, higher int) error {
	err := s.callingTheEndpointWithoutTokenShouldResultInStatusBetween(path1, lower, higher)
	if err != nil {
		return err
	}
	return s.callingTheEndpointWithoutTokenShouldResultInStatusBetween(path2, lower, higher)
}
