package ory

import (
	"context"
	_ "embed"
	"fmt"
	"net/http"
	"strings"

	"github.com/avast/retry-go/v4"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/auth"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/helpers"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/manifestprocessor"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/resource"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/testcontext"
	"golang.org/x/oauth2/clientcredentials"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
)

type scenario struct {
	name                    string
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
	return s.callingTheEndpointWithMethodWithValidToken(fmt.Sprintf("%s%s", s.Url, path), http.MethodGet, tokenType, asserter)
}

func (s *scenario) callingTheEndpointWithMethodWithValidTokenShouldResultInStatusBetween(path string, method string, tokenType string, lower, higher int) error {
	asserter := &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher}
	return s.callingTheEndpointWithMethodWithValidToken(fmt.Sprintf("%s%s", s.Url, path), method, tokenType, asserter)
}

func (s *scenario) callingTheEndpointWithMethodWithValidToken(url string, method string, tokenType string, asserter helpers.HttpResponseAsserter) error {

	requestHeaders := make(map[string]string)

	switch tokenType {
	case "OAuth2":
		tokenOpaque, err := auth.GetAccessToken(*s.oauth2Cfg, "opaque")
		if err != nil {
			return fmt.Errorf("failed to fetch an id_token: %s", err.Error())
		}

		requestHeaders[testcontext.AuthorizationHeaderName] = fmt.Sprintf("Bearer %s", tokenOpaque)
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

func (s *scenario) thereIsAHttpbinServiceAndApiRuleIsApplied() error {
	err := s.thereIsAHttpbinService()
	if err != nil {
		return err
	}

	res, err := manifestprocessor.ParseSingleEntryFromFileWithTemplate(s.ApiResourceManifestPath, s.ApiResourceDirectory, s.ManifestTemplate)
	if err != nil {
		return err
	}
	return helpers.CreateApiRule(s.resourceManager, s.k8sClient, testcontext.GetRetryOpts(), res)
}

func (s *scenario) theAPIRuleIsApplied() error {
	res, err := manifestprocessor.ParseSingleEntryFromFileWithTemplate(s.ApiResourceManifestPath, s.ApiResourceDirectory, s.ManifestTemplate)
	if err != nil {
		return err
	}
	return helpers.CreateApiRule(s.resourceManager, s.k8sClient, testcontext.GetRetryOpts(), res)
}

func (s *scenario) theAPIRuleIsUpdated(manifest string) error {
	res, err := manifestprocessor.ParseSingleEntryFromFileWithTemplate(manifest, s.ApiResourceDirectory, s.ManifestTemplate)
	if err != nil {
		return err
	}
	return helpers.UpdateApiRule(s.resourceManager, s.k8sClient, testcontext.GetRetryOpts(), res)
}

func (s *scenario) theAPIRuleIsUpdatedToV2alpha1(manifest string) error {
	res, err := manifestprocessor.ParseSingleEntryFromFileWithTemplate(manifest, s.ApiResourceDirectory, s.ManifestTemplate)
	if err != nil {
		return err
	}
	return helpers.UpdateApiRuleV2alpha1(s.resourceManager, s.k8sClient, testcontext.GetRetryOpts(), res)
}

func (s *scenario) theAPIRuleIsDeletedUsingv2alpha1Version() error {
	res, err := manifestprocessor.ParseSingleEntryFromFileWithTemplate(s.ApiResourceManifestPath, s.ApiResourceDirectory, s.ManifestTemplate)
	if err != nil {
		return err
	}

	groupVersionResource, err := resource.GetGvrFromUnstructured(s.resourceManager, res)
	if err != nil {
		return err
	}
	groupVersionResource.Version = "v2alpha1"

	return s.resourceManager.DeleteResource(s.k8sClient, *groupVersionResource, res.GetNamespace(), res.GetName())
}

func (s *scenario) theAPIRuleHasStatus(expectedStatus string) error {
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

		apiRuleStatus, err := helpers.GetAPIRuleStatusV2(apiRule)
		if err != nil {
			return fmt.Errorf("APIRule %s not in expected status %s. Error getting status: %w", apiRule.GetName(), expectedStatus, err)
		}

		if apiRuleStatus.Status.State != expectedStatus {
			return fmt.Errorf("APIRule %s not in expected status %s. Status: %s, Desc:\n%s", apiRule.GetName(), expectedStatus, apiRuleStatus.Status.State, apiRuleStatus.Status.Description)
		}

		return nil
	}, testcontext.GetRetryOpts()...)
}

func (s *scenario) theAPIRuleIsNotFound() error {
	res, err := manifestprocessor.ParseSingleEntryFromFileWithTemplate(s.ApiResourceManifestPath, s.ApiResourceDirectory, s.ManifestTemplate)
	if err != nil {
		return err
	}

	gvr, err := resource.GetGvrFromUnstructured(s.resourceManager, res)
	if err != nil {
		return err
	}

	return retry.Do(func() error {
		_, err = s.k8sClient.Resource(*gvr).Namespace(res.GetNamespace()).Get(context.Background(), res.GetName(), metav1.GetOptions{})

		if apierrors.IsNotFound(err) {
			return nil
		}
		if err != nil {
			return err
		}

		return fmt.Errorf("expected that APIRule %s not to exist, but it exists", res.GetName())
	}, testcontext.GetRetryOpts()...)
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

func (s *scenario) thereIsAHelloWorldService() error {
	resources, err := manifestprocessor.ParseFromFileWithTemplate("testing-app-helloworld.yaml", s.ApiResourceDirectory, s.ManifestTemplate)
	if err != nil {
		return err
	}
	_, err = s.resourceManager.CreateResources(s.k8sClient, resources...)
	if err != nil {
		return err
	}

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

func (s *scenario) teardownHelloWorldService() error {
	resources, err := manifestprocessor.ParseFromFileWithTemplate("testing-app-helloworld.yaml", s.ApiResourceDirectory, s.ManifestTemplate)
	if err != nil {
		return err
	}
	err = s.resourceManager.DeleteResources(s.k8sClient, resources...)
	if err != nil {
		return err
	}

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
