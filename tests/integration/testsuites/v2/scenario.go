package v2

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/avast/retry-go/v4"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/auth"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/helpers"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/manifestprocessor"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/resource"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/testcontext"
	"golang.org/x/oauth2/clientcredentials"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
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

type tokenFrom struct {
	From     string
	Prefix   string
	AsHeader bool
}

func (s *scenario) callingTheEndpointWithAValidToken(endpoint, tokenType, audOrClaim, par1, par2 string, lower, higher int) error {
	asserter := &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher}
	tokenFrom := tokenFrom{
		From:     testcontext.AuthorizationHeaderName,
		Prefix:   testcontext.AuthorizationHeaderPrefix,
		AsHeader: true,
	}
	if audOrClaim == "audiences" {
		return s.callingEndpointWithMethodAndHeaders(fmt.Sprintf("%s/%s", s.Url, strings.TrimLeft(endpoint, "/")), http.MethodGet, tokenType, asserter, nil, &tokenFrom, helpers.RequestOptions{Audiences: []string{par1, par2}})
	} else {
		return s.callingEndpointWithMethodAndHeaders(fmt.Sprintf("%s/%s", s.Url, strings.TrimLeft(endpoint, "/")), http.MethodGet, tokenType, asserter, nil, &tokenFrom, helpers.RequestOptions{Scopes: []string{par1, par2}})
	}
}

func (s *scenario) callingTheEndpointWithValidTokenShouldResultInStatusBetween(path string, tokenType string, lower, higher int) error {
	asserter := &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher}
	return s.callingTheEndpointWithMethodWithValidToken(fmt.Sprintf("%s%s", s.Url, path), http.MethodGet, tokenType, asserter)
}

func (s *scenario) callingTheEndpointWithMethodWithValidTokenShouldResultInStatusBetween(path string, method string, tokenType string, lower, higher int) error {
	asserter := &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher}
	return s.callingTheEndpointWithMethodWithValidToken(fmt.Sprintf("%s%s", s.Url, path), method, tokenType, asserter)
}

func (s *scenario) callingTheEndpointWithMethodShouldResultInStatusBetween(path string, method string, lower, higher int) error {
	requestHeaders := make(map[string]string)
	asserter := &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher}
	return s.httpClient.CallEndpointWithHeadersAndMethod(requestHeaders, fmt.Sprintf("%s%s", s.Url, path), method, asserter)
}

func (s *scenario) callingTheEndpointWithMethodWithValidToken(url string, method string, tokenType string, asserter helpers.HttpResponseAsserter, additionalRequestHeaders ...map[string]string) error {
	requestHeaders := make(map[string]string)
	if len(additionalRequestHeaders) > 0 {
		requestHeaders = additionalRequestHeaders[0]
	}

	switch tokenType {
	case "JWT":
		tokenJwt, err := auth.GetAccessToken(*s.jwtConfig, "jwt")
		if err != nil {
			return fmt.Errorf("failed to fetch an id_token: %s", err.Error())
		}

		requestHeaders[testcontext.AuthorizationHeaderName] = fmt.Sprintf("Bearer %s", tokenJwt)
	default:
		return fmt.Errorf("unsupported token type: %s", tokenType)
	}

	return s.httpClient.CallEndpointWithHeadersAndMethod(requestHeaders, url, method, asserter)
}

func (s *scenario) theAPIRuleIsApplied() error {
	res, err := manifestprocessor.ParseSingleEntryFromFileWithTemplate(s.ApiResourceManifestPath, s.ApiResourceDirectory, s.ManifestTemplate)
	if err != nil {
		return err
	}
	return helpers.CreateApiRule(s.resourceManager, s.k8sClient, testcontext.GetRetryOpts(), res)
}

func (s *scenario) theMisconfiguredAPIRuleIsApplied() error {
	res, err := manifestprocessor.ParseSingleEntryFromFileWithTemplate(s.ApiResourceManifestPath, s.ApiResourceDirectory, s.ManifestTemplate)
	if err != nil {
		return err
	}
	_, err = s.resourceManager.CreateResources(s.k8sClient, res)
	return err
}

func (s *scenario) theAPIRulev2IsApplied() error {
	res, err := manifestprocessor.ParseSingleEntryFromFileWithTemplate(s.ApiResourceManifestPath, s.ApiResourceDirectory, s.ManifestTemplate)
	if err != nil {
		return err
	}
	return helpers.CreateApiRule(s.resourceManager, s.k8sClient, testcontext.GetRetryOpts(), res)
}

func (s *scenario) theAPIRuleTemplateFileIsSetTo(templateFileName string) {
	s.ManifestTemplate["NamePrefix"] = strings.TrimRight(templateFileName, ".yaml")
	s.ApiResourceManifestPath = templateFileName
}

func (s *scenario) templateValueIsSetTo(key, value string) {
	s.ManifestTemplate[key] = value
}

func (s *scenario) theAPIRulev2IsAppliedExpectError(errorMessage string) error {
	res, err := manifestprocessor.ParseSingleEntryFromFileWithTemplate(s.ApiResourceManifestPath, s.ApiResourceDirectory, s.ManifestTemplate)
	if err != nil {
		return err
	}
	return helpers.CreateApiRuleExpectError(s.resourceManager, s.k8sClient, testcontext.GetRetryOpts(), res, errorMessage)
}

func (s *scenario) specifiesCustomGateway(gatewayNamespace, gatewayName string) {
	s.ManifestTemplate["GatewayNamespace"] = gatewayNamespace
	s.ManifestTemplate["GatewayName"] = gatewayName
}

func (s *scenario) theAPIRuleHasStatusWithDesc(expectedState, expectedDescription string) error {
	res, err := manifestprocessor.ParseSingleEntryFromFileWithTemplate(s.ApiResourceManifestPath, s.ApiResourceDirectory, s.ManifestTemplate)
	if err != nil {
		return err
	}

	groupVersionResource, err := resource.GetGvrFromUnstructured(s.resourceManager, res)
	if err != nil {
		return err
	}

	return retry.Do(func() error {
		apiRule, err := s.resourceManager.GetResource(s.k8sClient, *groupVersionResource, res.GetNamespace(), res.GetName())
		if err != nil {
			return err
		}

		apiRuleStatus, err := helpers.GetAPIRuleStatus(apiRule)
		if err != nil {
			return err
		}

		hasExpected := apiRuleStatus.GetStatus() == expectedState && strings.Contains(apiRuleStatus.GetDescription(), expectedDescription)
		if !hasExpected {
			return fmt.Errorf("APIRule %s not in expected status %s or not containing description %s. Status: %s, Description:\n%s", apiRule.GetName(), expectedState, expectedDescription, apiRuleStatus.GetStatus(), apiRuleStatus.GetDescription())
		}

		return nil
	}, testcontext.GetRetryOpts()...)
}

func (s *scenario) getGatewayHost(name, namespace string) (string, error) {
	var host string
	err := retry.Do(func() error {
		apiRule, err := s.resourceManager.GetResource(s.k8sClient, resource.GetResourceGvr("Gateway"), namespace, name)
		if err != nil {
			return err
		}
		host = strings.TrimPrefix(apiRule.Object["spec"].(map[string]interface{})["servers"].([]interface{})[0].(map[string]interface{})["hosts"].([]interface{})[0].(string), "*.")
		return nil
	}, testcontext.GetRetryOpts()...)

	if err != nil {
		return "", err
	}
	return host, nil
}

func (s *scenario) callingTheEndpointWithMethodWithInvalidTokenShouldResultInStatusBetween(path string, method string, lower, higher int) error {
	requestHeaders := map[string]string{testcontext.AuthorizationHeaderName: testcontext.AnyToken}
	return s.httpClient.CallEndpointWithHeadersAndMethod(requestHeaders, fmt.Sprintf("%s%s", s.Url, path), method, &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher})
}

func (s *scenario) callingTheEndpointWithInvalidTokenShouldResultInStatusBetween(path string, lower, higher int) error {
	requestHeaders := map[string]string{testcontext.AuthorizationHeaderName: testcontext.AnyToken}
	return s.httpClient.CallEndpointWithHeadersWithRetries(requestHeaders, fmt.Sprintf("%s%s", s.Url, path), &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher})
}

func (s *scenario) callingTheEndpointWithoutTokenShouldResultInStatusBetween(path string, lower, higher int) error {
	return s.httpClient.CallEndpointWithRetries(fmt.Sprintf("%s/%s", s.Url, strings.TrimLeft(path, "/")), &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher})
}

func (s *scenario) inClusterCallingTheEndpointWithoutTokenShouldFail(path string) error {
	curlCommand := []string{"curl", "--fail-with-body", "--retry", "5", "-sSL", "-m", "10", fmt.Sprintf("http://httpbin-%s.%s.svc.cluster.local:8000%s", s.TestID, s.Namespace, path)}

	log, err := helpers.RunCurlInPod(s.Namespace, curlCommand)
	if err != nil {
		return nil
	}

	return fmt.Errorf("%s, %s", "Request should fail, but it succeeded", log)
}

func (s *scenario) inClusterCallingTheEndpointWithTokenShouldFail(path, tokenType string) error {
	var headers string
	switch tokenType {
	case "JWT":
		tokenJwt, err := auth.GetAccessToken(*s.jwtConfig, "jwt")
		if err != nil {
			return fmt.Errorf("failed to fetch an id_token: %s", err.Error())
		}

		headers = fmt.Sprintf("%s: Bearer %s", testcontext.AuthorizationHeaderName, tokenJwt)
	default:
		return fmt.Errorf("unsupported token type: %s", tokenType)
	}

	curlCommand := []string{"curl", "--fail-with-body", "--retry", "5", "-sSL", "-m", "10", "-H", headers, fmt.Sprintf("http://httpbin-%s.%s.svc.cluster.local:8000%s", s.TestID, s.Namespace, path)}

	log, err := helpers.RunCurlInPod(s.Namespace, curlCommand)
	if err != nil {
		return nil
	}

	return fmt.Errorf("%s, %s", "Request should fail, but it succeeded", log)
}

func (s *scenario) callingShortHostWithoutTokenShouldResultInStatusBetween(host, path string, lower, higher int) error {
	gatewayHost, err := s.getGatewayHost(s.config.GatewayName, s.config.GatewayNamespace)
	if err != nil {
		return err
	}
	return s.httpClient.CallEndpointWithRetries(fmt.Sprintf("https://%s.%s/%s", host, gatewayHost, strings.TrimLeft(path, "/")), &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher})
}

func (s *scenario) callingTheEndpointWithHeader(path, headerName, value string, lower, higher int) error {
	requestHeaders := map[string]string{headerName: value}
	return s.httpClient.CallEndpointWithHeadersWithRetries(requestHeaders, fmt.Sprintf("%s/%s", s.Url, strings.TrimLeft(path, "/")), &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher})
}

func (s *scenario) callingTheEndpointWithHeaderAndInvalidJwt(path, headerName, _, value string, lower, higher int) error {
	requestHeaders := map[string]string{headerName: value, testcontext.AuthorizationHeaderName: testcontext.AnyToken}
	return s.httpClient.CallEndpointWithHeadersWithRetries(requestHeaders, fmt.Sprintf("%s/%s", s.Url, strings.TrimLeft(path, "/")), &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher})
}

func (s *scenario) callingTheEndpointWithHeaderAndValidJwt(path, headerName, value, tokenType string, lower, higher int) error {
	requestHeaders := map[string]string{headerName: value}
	return s.callingTheEndpointWithMethodWithValidToken(fmt.Sprintf("%s/%s", s.Url, strings.TrimLeft(path, "/")), http.MethodGet, tokenType, &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher}, requestHeaders)
}

func (s *scenario) thereIsAnJwtSecuredPath(path string) {
	s.ManifestTemplate["jwtSecuredPath"] = path
}

func (s *scenario) emptyStep() {}

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

func (s *scenario) thereIsAnEndpointWithExtAuth(provider, path string) error {
	s.ManifestTemplate["extAuthPath"] = path
	s.ManifestTemplate["extAuthProvider"] = provider

	return nil
}

func (s *scenario) theEndpointHasJwtRestrictionsWithScope() error {
	return nil
}

// teardownHttpbinService deletes the httpbin service and reset the url in the scenario. This should be considered a temporary solution
// to reduce resource consumption until we implement a better way to clean up the resources by a scenario. If the test fails before this step the teardown won't be executed.
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

func (s *scenario) preflightEndpointCallResponseHeaders(endpoint, origin string, statusCode int, headerKey, headerValue string) error {
	headers := map[string]string{
		"Origin":                        origin,
		"Access-Control-Request-Method": "GET,POST,PUT,DELETE,PATCH",
	}
	return retry.Do(func() error {
		resp, err := s.httpClient.CallEndpointWithRetriesAndGetResponse(headers, nil, http.MethodOptions, s.Url+endpoint)
		if err != nil {
			return err
		}
		if resp.StatusCode != statusCode {
			return fmt.Errorf("expected response status code %d got %d", statusCode, resp.StatusCode)
		}
		rhv := resp.Header.Get(headerKey)
		if rhv != headerValue {
			return fmt.Errorf("expected header %s with value %s, got %s", headerKey, headerValue, rhv)
		}
		return nil
	}, testcontext.GetRetryOpts()...)
}

func (s *scenario) preflightEndpointCallNoResponseHeader(endpoint, origin string, statusCode int, headerKey string) error {
	headers := map[string]string{
		"Origin":                        origin,
		"Access-Control-Request-Method": "GET,POST,PUT,DELETE,PATCH",
	}
	return retry.Do(func() error {
		resp, err := s.httpClient.CallEndpointWithRetriesAndGetResponse(headers, nil, http.MethodOptions, s.Url+endpoint)
		if err != nil {
			return err
		}
		if resp.StatusCode != statusCode {
			return fmt.Errorf("expected response status code %d got %d", statusCode, resp.StatusCode)
		}
		if len(resp.Header.Values(headerKey)) > 0 {
			return fmt.Errorf("expected that the response will not contain %s header, but did", headerKey)
		}
		return nil
	}, testcontext.GetRetryOpts()...)
}

func (s *scenario) callingEndpointWithMethodAndHeaders(endpointUrl string, method string, tokenType string, asserter helpers.HttpResponseAsserter, requestHeaders map[string]string, tokenFrom *tokenFrom, options ...helpers.RequestOptions) error {
	if requestHeaders == nil {
		requestHeaders = make(map[string]string)
	}

	oCfg := *s.oauth2Cfg

	if len(options) > 0 {
		if len(oCfg.EndpointParams) == 0 {
			oCfg.EndpointParams = make(url.Values)
		}

		if len(options[0].Scopes) > 0 {
			oCfg.Scopes = options[0].Scopes
		}

		if len(options[0].Audiences) > 0 {
			oCfg.EndpointParams.Add("audience", strings.Join(options[0].Audiences, ","))
		}
	}

	token, err := auth.GetAccessToken(oCfg, strings.ToLower(tokenType))
	if err != nil {
		return fmt.Errorf("failed to fetch an id_token: %s", err.Error())
	}

	switch tokenType {
	case "Opaque":
		requestHeaders[testcontext.OpaqueHeaderName] = token
	case "JWT":
		if tokenFrom.From == "" {
			return errors.New("jwt from header or parameter name not specified")
		}
		if tokenFrom.AsHeader {
			if tokenFrom.Prefix != "" {
				token = fmt.Sprintf("%s %s", tokenFrom.Prefix, token)
			}
			requestHeaders[tokenFrom.From] = token
		} else {
			endpointUrl = fmt.Sprintf("%s?%s=%s", endpointUrl, tokenFrom.From, token)
		}
	default:
		return fmt.Errorf("unsupported token type: %s", tokenType)
	}

	return s.httpClient.CallEndpointWithHeadersAndMethod(requestHeaders, endpointUrl, method, asserter)
}

func (s *scenario) callingTheEndpointShouldResultInBodyContaining(endpoint string, header string, bodyContent string) error {
	asserter := &helpers.BodyContainsPredicate{Expected: []string{header, bodyContent}}
	return s.httpClient.CallEndpointWithRetries(fmt.Sprintf("%s/%s", s.Url, strings.TrimLeft(endpoint, "/")), asserter)
}

func (s *scenario) callingTheEndpointWithValidTokenShouldResultInBodyContaining(endpoint, tokenType string, bodyContent string) error {
	asserter := &helpers.BodyContainsPredicate{Expected: []string{bodyContent}}
	tokenFrom := tokenFrom{
		From:     testcontext.AuthorizationHeaderName,
		Prefix:   testcontext.AuthorizationHeaderPrefix,
		AsHeader: true,
	}
	return s.callingEndpointWithMethodAndHeaders(fmt.Sprintf("%s/%s", s.Url, strings.TrimLeft(endpoint, "/")), http.MethodGet, tokenType, asserter, nil, &tokenFrom)
}

func (s *scenario) apiRuleContainsOriginalVersionAnnotation(version string) error {
	res, err := manifestprocessor.ParseSingleEntryFromFileWithTemplate(s.ApiResourceManifestPath, s.ApiResourceDirectory, s.ManifestTemplate)
	if err != nil {
		return err
	}

	groupVersionResource, err := resource.GetGvrFromUnstructured(s.resourceManager, res)
	if err != nil {
		return err
	}

	return retry.Do(func() error {
		apiRule, err := s.resourceManager.GetResource(s.k8sClient, *groupVersionResource, res.GetNamespace(), res.GetName())
		if err != nil {
			return fmt.Errorf("failed to get APIRule: %w", err)
		}

		versionAnnotation := apiRule.GetAnnotations()["gateway.kyma-project.io/original-version"]
		if versionAnnotation != version {
			return fmt.Errorf("expected original version annotation to be %s, got %s", version, versionAnnotation)
		}

		return nil
	}, testcontext.GetRetryOpts()...)
}

func (s *scenario) resourceOwnedByApiRuleExists(resourceKind string) error {
	res := resource.GetResourceGvr(resourceKind)
	name := s.ManifestTemplate["NamePrefix"]
	ownerLabelSelector := fmt.Sprintf("apirule.gateway.kyma-project.io/v1beta1=%s-%s.%s", name, s.TestID, s.Namespace)
	return retry.Do(func() error {
		list, err := s.k8sClient.Resource(res).Namespace(s.Namespace).List(context.Background(), metav1.ListOptions{LabelSelector: ownerLabelSelector})
		if err != nil {
			return err
		}

		if len(list.Items) == 0 {
			return fmt.Errorf("expected at least one %s owned by APIRule, got 0", resourceKind)
		}

		return nil
	}, testcontext.GetRetryOpts()...)
}

func (s *scenario) theAPIRuleIsUpdated(manifest string) error {
	res, err := manifestprocessor.ParseSingleEntryFromFileWithTemplate(manifest, s.ApiResourceDirectory, s.ManifestTemplate)
	if err != nil {
		return err
	}
	return helpers.UpdateApiRule(s.resourceManager, s.k8sClient, testcontext.GetRetryOpts(), res)
}
