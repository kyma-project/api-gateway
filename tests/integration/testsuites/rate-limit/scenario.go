package ratelimit

import (
	"fmt"

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
}

func (s *scenario) callingEndpointWithHeadersNTimesShouldResultWithStatusCode(endpoint string, n, expectedStatusCode int) error {
	url := s.Url + endpoint
	headers := map[string]string{
		"X-Rate-Limited": "true",
	}
	for i := 0; i < n; i++ {
		if i == n-1 {
			err := s.httpClient.CallEndpointWithHeadersWithRetries(headers, url, &helpers.StatusPredicate{LowerStatusBound: expectedStatusCode, UpperStatusBound: expectedStatusCode})
			if err != nil {
				return err
			}
			return nil
		}
		err := s.httpClient.CallEndpointWithHeadersWithRetries(headers, url, &helpers.StatusPredicate{LowerStatusBound: 200, UpperStatusBound: 200})
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *scenario) callingEndpointNTimesShouldResultWithStatusCode(endpoint string, n, expectedStatusCode int) error {
	url := s.Url + endpoint

	for i := 0; i < n; i++ {
		if i == n-1 {
			err := s.httpClient.CallEndpointWithRetries(url, &helpers.StatusPredicate{LowerStatusBound: expectedStatusCode, UpperStatusBound: expectedStatusCode})
			if err != nil {
				return err
			}
			return nil
		}
		err := s.httpClient.CallEndpointWithRetries(url, &helpers.StatusPredicate{LowerStatusBound: 200, UpperStatusBound: 200})
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *scenario) rateLimitWithPathBaseConfigurationApplied() error {
	resources, err := manifestprocessor.ParseFromFileWithTemplate("ratelimit-path-based.yaml", s.ApiResourceDirectory, s.ManifestTemplate)
	if err != nil {
		return err
	}

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

func (s *scenario) rateLimitWithPathAndHeaderBaseConfigurationApplied() error {
	resources, err := manifestprocessor.ParseFromFileWithTemplate("ratelimit-path-and-header-based.yaml", s.ApiResourceDirectory, s.ManifestTemplate)
	if err != nil {
		return err
	}

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

func (s *scenario) rateLimitWithHeaderBaseConfigurationApplied() error {
	resources, err := manifestprocessor.ParseFromFileWithTemplate("ratelimit-header-based.yaml", s.ApiResourceDirectory, s.ManifestTemplate)
	if err != nil {
		return err
	}

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
