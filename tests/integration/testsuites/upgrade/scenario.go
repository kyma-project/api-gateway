package upgrade

import (
	"errors"
	"fmt"
	"strings"
	"time"

	v1 "k8s.io/api/apps/v1"

	"github.com/avast/retry-go/v4"
	"github.com/cucumber/godog"
	apirulev1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/auth"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/helpers"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/manifestprocessor"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/resource"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/testcontext"
	"golang.org/x/oauth2/clientcredentials"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

var deploymentGVR = schema.GroupVersionResource{
	Group:    "apps",
	Version:  "v1",
	Resource: "deployments",
}

const apiGatewayNS, apiGatewayName = "kyma-system", "api-gateway-controller-manager"

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
	httpClient              *helpers.RetryableHttpClient
	resourceManager         *resource.Manager
	config                  testcontext.Config
	APIGatewayImageVersion  string
}

func (s *scenario) theAPIRuleIsApplied() error {
	r, err := manifestprocessor.ParseFromFileWithTemplate(s.ApiResourceManifestPath, s.ApiResourceDirectory, s.ManifestTemplate)
	if err != nil {
		return err
	}
	return helpers.ApplyApiRule(s.resourceManager.CreateOrUpdateResources, s.resourceManager.UpdateResources, s.k8sClient, testcontext.GetRetryOpts(), r)
}

func (s *scenario) callingTheEndpointWithValidTokenShouldResultInStatusBetween(endpoint, tokenType string, lower, higher int) error {
	asserter := &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher}
	tokenFrom := tokenFrom{
		From:     testcontext.AuthorizationHeaderName,
		Prefix:   testcontext.AuthorizationHeaderPrefix,
		AsHeader: true,
	}
	return s.callingEndpointWithHeadersWithRetries(fmt.Sprintf("%s/%s", s.Url, strings.TrimLeft(endpoint, "/")), tokenType, asserter, nil, &tokenFrom)
}

func (s *scenario) callingEndpointWithHeadersWithRetries(url string, tokenType string, asserter helpers.HttpResponseAsserter, requestHeaders map[string]string, tokenFrom *tokenFrom) error {
	if requestHeaders == nil {
		requestHeaders = make(map[string]string)
	}

	token, err := auth.GetAccessToken(*s.oauth2Cfg, strings.ToLower(tokenType))
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
			url = fmt.Sprintf("%s?%s=%s", url, tokenFrom.From, token)
		}
	default:
		return fmt.Errorf("unsupported token type: %s", tokenType)
	}

	return s.httpClient.CallEndpointWithHeadersWithRetries(requestHeaders, url, asserter)
}

func (s *scenario) callingTheEndpointWithoutTokenShouldResultInStatusBetween(endpoint string, lower, higher int) error {
	return s.httpClient.CallEndpointWithRetries(fmt.Sprintf("%s/%s", s.Url, strings.TrimLeft(endpoint, "/")), &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher})
}

func (s *scenario) callingTheEndpointWithInvalidTokenShouldResultInStatusBetween(path string, lower, higher int) error {
	requestHeaders := map[string]string{testcontext.AuthorizationHeaderName: testcontext.AnyToken}
	return s.httpClient.CallEndpointWithHeadersWithRetries(requestHeaders, fmt.Sprintf("%s%s", s.Url, path), &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher})
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

func (s *scenario) thereIsAnJwtSecuredPath(path string) {
	s.ManifestTemplate["jwtSecuredPath"] = path
}

func (s *scenario) upgradeApiGateway() error {
	const manifestDirectory = "testsuites/upgrade/manifests"
	const manifestFileName = "upgrade-test-generated-operator-manifest.yaml"

	manifestCrds, err := manifestprocessor.ParseYamlFromFile(manifestFileName, manifestDirectory)
	if err != nil {
		return err
	}

	return retry.Do(func() error {
		var apiGatewayDeployment v1.Deployment

		apiGatewayUnstructured, err := s.resourceManager.GetResource(s.k8sClient, deploymentGVR, apiGatewayNS, apiGatewayName)
		if err != nil {
			return err
		}

		err = runtime.DefaultUnstructuredConverter.FromUnstructured(apiGatewayUnstructured.UnstructuredContent(), &apiGatewayDeployment)
		if err != nil {
			return err
		}

		err = s.resourceManager.MergeAndUpdateOrCreateResources(s.k8sClient, manifestCrds)
		if err != nil {
			return err
		}

		apiGatewayUnstructured, err = s.resourceManager.GetResource(s.k8sClient, deploymentGVR, apiGatewayNS, apiGatewayName)
		if err != nil {
			return err
		}

		err = runtime.DefaultUnstructuredConverter.FromUnstructured(apiGatewayUnstructured.UnstructuredContent(), &apiGatewayDeployment)
		if err != nil {
			return err
		}

		if currentImage := apiGatewayDeployment.Spec.Template.Spec.Containers[0].Image; currentImage != s.APIGatewayImageVersion {
			return fmt.Errorf("image is not updated, expected: %s, got: %s", s.APIGatewayImageVersion, currentImage)
		}

		return nil
	}, testcontext.GetRetryOpts()...)

}

func (s *scenario) reconciliationHappened(numberOfSeconds int) error {
	apiRules, err := manifestprocessor.ParseFromFileWithTemplate(s.ApiResourceManifestPath, s.ApiResourceDirectory, s.ManifestTemplate)
	if err != nil {
		return err
	}

	return retry.Do(func() error {
		for _, apiRule := range apiRules {
			var apiRuleStructured apirulev1beta1.APIRule
			res, err := s.resourceManager.GetResource(s.k8sClient, schema.GroupVersionResource{
				Group:    apirulev1beta1.GroupVersion.Group,
				Version:  apirulev1beta1.GroupVersion.Version,
				Resource: "apirules",
			}, apiRule.GetNamespace(), apiRule.GetName(), retry.Attempts(1))

			if err != nil {
				return err
			}

			err = runtime.DefaultUnstructuredConverter.FromUnstructured(res.UnstructuredContent(), &apiRuleStructured)
			if err != nil {
				return err
			}

			if time.Since(apiRuleStructured.Status.LastProcessedTime.Time) > time.Second*time.Duration(numberOfSeconds) {
				return fmt.Errorf("reconcilation didn't happened in last %d seconds", numberOfSeconds)
			}
		}
		return nil
	}, retry.Attempts(uint(numberOfSeconds)), retry.Delay(time.Second))
}

func initCommon(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("istio-jwt-common.yaml", "api-gateway-upgrade")

	ctx.Step(`Upgrade: There is a httpbin service$`, scenario.thereIsAHttpbinService)
	ctx.Step(`Upgrade: There is an endpoint secured with JWT on path "([^"]*)"`, scenario.thereIsAnJwtSecuredPath)
	ctx.Step(`Upgrade: The APIRule is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(`Upgrade: Calling the "([^"]*)" endpoint without a token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithoutTokenShouldResultInStatusBetween)
	ctx.Step(`Upgrade: Calling the "([^"]*)" endpoint with an invalid token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithInvalidTokenShouldResultInStatusBetween)
	ctx.Step(`Upgrade: Calling the "([^"]*)" endpoint with a valid "([^"]*)" token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithValidTokenShouldResultInStatusBetween)
	ctx.Step(`Upgrade: API Gateway is upgraded to current branch version$`, scenario.upgradeApiGateway)
	ctx.Step(`Upgrade: Teardown httpbin service$`, scenario.teardownHttpbinService)
	ctx.Step(`Upgrade: A reconciliation happened in the last (\d+) seconds$`, scenario.reconciliationHappened)
}
