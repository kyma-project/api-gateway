package ratelimit

import (
	"fmt"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"net/http"
	"net/url"

	"github.com/avast/retry-go/v4"
	"k8s.io/client-go/dynamic"

	"github.com/kyma-project/api-gateway/tests/integration/pkg/helpers"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/manifestprocessor"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/resource"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/testcontext"
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
	httpClient              *helpers.RetryableHttpClient
	resourceManager         *resource.Manager
	config                  testcontext.Config
	rateLimitResources      []unstructured.Unstructured
}

func (s *scenario) callingEndpointWithHeadersNTimesShouldResultWithStatusCode(endpoint string, expectedStatusCode int) error {
	endpointUrl, err := url.Parse(s.Url + endpoint)
	if err != nil {
		return err
	}
	httpClient := s.httpClient.HttpClient()
	req := &http.Request{
		URL:    endpointUrl,
		Method: http.MethodGet,
		Header: map[string][]string{
			"X-Rate-Limited": {"true"},
		},
	}
	err = retry.Do(func() error {
		response, err := httpClient.Do(req)
		if err != nil {
			return err
		}
		if response.StatusCode != expectedStatusCode {
			return fmt.Errorf("status code %d is not match expected status code  %d", response.StatusCode, expectedStatusCode)
		}
		return nil
	}, testcontext.GetRetryOpts()...,
	)
	if err != nil {
		return err
	}
	return nil
}

func (s *scenario) callingEndpointNTimesShouldResultWithStatusCode(endpoint string, expectedStatusCode int) error {
	endpointUrl, err := url.Parse(s.Url + endpoint)
	if err != nil {
		return err
	}
	httpClient := s.httpClient.HttpClient()
	req := &http.Request{
		URL:    endpointUrl,
		Method: http.MethodGet,
	}
	err = retry.Do(func() error {
		response, err := httpClient.Do(req)
		if err != nil {
			return err
		}
		if response.StatusCode != expectedStatusCode {
			return fmt.Errorf("status code %d is not match expected status code  %d", response.StatusCode, expectedStatusCode)
		}
		return nil
	}, testcontext.GetRetryOpts()...,
	)
	if err != nil {
		return err
	}
	return nil
}

func (s *scenario) rateLimitWithPathBaseConfigurationApplied() error {
	resources, err := manifestprocessor.ParseFromFileWithTemplate("ratelimit-path-based.yaml", s.ApiResourceDirectory, s.ManifestTemplate)
	if err != nil {
		return err
	}
	s.rateLimitResources = resources

	_, err = s.resourceManager.CreateResources(s.k8sClient, resources...)
	if err != nil {
		return err
	}

	err = helpers.WaitForRateLimit(s.resourceManager, s.k8sClient, s.Namespace, "ratelimit-path-sample", testcontext.GetRetryOpts())
	if err != nil {
		return err
	}
	return nil
}

func (s *scenario) rateLimitTargetingIngressWithPathBaseConfigurationApplied() error {
	resources, err := manifestprocessor.ParseFromFileWithTemplate("ratelimit-ingressgateway-path-based.yaml", s.ApiResourceDirectory, s.ManifestTemplate)
	if err != nil {
		return err
	}
	s.rateLimitResources = resources

	_, err = s.resourceManager.CreateResources(s.k8sClient, resources...)
	if err != nil {
		return err
	}

	err = helpers.WaitForRateLimit(s.resourceManager, s.k8sClient, "istio-system", "ratelimit-ingressgateway-path-sample", testcontext.GetRetryOpts())
	if err != nil {
		return err
	}
	return nil
}

func (s *scenario) rateLimitWithPathAndHeaderBaseConfigurationApplied() error {
	resources, err := manifestprocessor.ParseFromFileWithTemplate("ratelimit-path-and-header-based.yaml", s.ApiResourceDirectory, s.ManifestTemplate)
	if err != nil {
		return err
	}
	s.rateLimitResources = resources

	_, err = s.resourceManager.CreateResources(s.k8sClient, resources...)
	if err != nil {
		return err
	}

	err = helpers.WaitForRateLimit(s.resourceManager, s.k8sClient, s.Namespace, "ratelimit-path-header-sample", testcontext.GetRetryOpts())
	if err != nil {
		return err
	}
	return nil
}

func (s *scenario) rateLimitTargetingIngressWithPathAndHeaderBaseConfigurationApplied() error {
	resources, err := manifestprocessor.ParseFromFileWithTemplate("ratelimit-ingressgateway-path-and-header-based.yaml", s.ApiResourceDirectory, s.ManifestTemplate)
	if err != nil {
		return err
	}
	s.rateLimitResources = resources

	_, err = s.resourceManager.CreateResources(s.k8sClient, resources...)
	if err != nil {
		return err
	}

	err = helpers.WaitForRateLimit(s.resourceManager, s.k8sClient, "istio-system", "ratelimit-ingressgateway-path-header-sample", testcontext.GetRetryOpts())
	if err != nil {
		return err
	}
	return nil
}

func (s *scenario) rateLimitWithHeaderBaseConfigurationApplied() error {
	resources, err := manifestprocessor.ParseFromFileWithTemplate("ratelimit-header-based.yaml", s.ApiResourceDirectory, s.ManifestTemplate)
	if err != nil {
		return err
	}
	s.rateLimitResources = resources

	_, err = s.resourceManager.CreateResources(s.k8sClient, resources...)
	if err != nil {
		return err
	}

	err = helpers.WaitForRateLimit(s.resourceManager, s.k8sClient, s.Namespace, "ratelimit-header-sample", testcontext.GetRetryOpts())
	if err != nil {
		return err
	}
	return nil
}

func (s *scenario) rateLimitTargetingIngressWithHeaderBaseConfigurationApplied() error {
	resources, err := manifestprocessor.ParseFromFileWithTemplate("ratelimit-ingressgateway-header-based.yaml", s.ApiResourceDirectory, s.ManifestTemplate)
	if err != nil {
		return err
	}
	s.rateLimitResources = resources

	_, err = s.resourceManager.CreateResources(s.k8sClient, resources...)
	if err != nil {
		return err
	}

	err = helpers.WaitForRateLimit(s.resourceManager, s.k8sClient, "istio-system", "ratelimit-ingressgateway-header-sample", testcontext.GetRetryOpts())
	if err != nil {
		return err
	}
	return nil
}

func (s *scenario) tearDownRateLimit() error {
	return s.resourceManager.DeleteResources(s.k8sClient, s.rateLimitResources...)
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

	err = helpers.WaitForDeployment(s.resourceManager, s.k8sClient, s.Namespace, fmt.Sprintf("httpbin-%s", s.TestID), testcontext.GetRetryOpts())
	if err != nil {
		return err
	}

	return nil
}

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
