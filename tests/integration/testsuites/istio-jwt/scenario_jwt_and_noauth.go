package istiojwt

import (
	"fmt"
	"github.com/avast/retry-go/v4"
	"github.com/cucumber/godog"
	apirulev1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/manifestprocessor"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"time"
)

func initJwtAndNoAuth(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("istio-jwt-and-noauth.yaml", "istio-jwt-noauth")

	ctx.Step(`JwtAndNoAuth: There is a httpbin service$`, scenario.thereIsAHttpbinService)
	ctx.Step(`JwtAndNoAuth: There is an endpoint secured with JWT and no_auth on path "([^"]*)"$`, scenario.thereIsAnJwtSecuredPath)
	ctx.Step(`JwtAndNoAuth: There is an endpoint with handler "([^"]*)" on path "([^"]*)"$`, scenario.thereIsAnEndpointWithHandler)
	ctx.Step(`JwtAndNoAuth: Create APIRule$`, scenario.createAPIRule)
	ctx.Step(`JwtAndNoAuth: The APIRule has "([^"]*)" status$`, scenario.theAPIRuleHasAStatus)
	ctx.Step(`JwtAndNoAuth: Teardown httpbin service$`, scenario.teardownHttpbinService)
}

func initJwtAndNoAuthMethods(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("istio-jwt-and-noauth-separate-methods.yaml", "istio-jwt-noauth-separate-methods")

	ctx.Step(`JwtAndNoAuthMethods: There is a httpbin service$`, scenario.thereIsAHttpbinService)
	ctx.Step(`JwtAndNoAuthMethods: There is an endpoint secured with JWT and no_auth on path "([^"]*)"$`, scenario.thereIsAnJwtSecuredPath)
	ctx.Step(`JwtAndNoAuthMethods: There is an endpoint with handler "([^"]*)" on path "([^"]*)"$`, scenario.thereIsAnEndpointWithHandler)
	ctx.Step(`JwtAndNoAuthMethods: The APIRule is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(`JwtAndNoAuthMethods: The APIRule has "([^"]*)" status$`, scenario.theAPIRuleHasAStatus)
	ctx.Step(`JwtAndNoAuthMethods: Teardown httpbin service$`, scenario.teardownHttpbinService)
}

func (s *scenario) theAPIRuleHasAStatus(status string) error {
	apiRules, err := manifestprocessor.ParseFromFileWithTemplate(s.ApiResourceManifestPath, s.ApiResourceDirectory, s.ManifestTemplate)
	if err != nil {
		return err
	}
	if len(apiRules) > 1 {
		return fmt.Errorf("more than one APIRule found in the manifests")
	}

	return retry.Do(func() error {
		var apiRuleStructured apirulev1beta1.APIRule
		res, err := s.resourceManager.GetResource(s.k8sClient, schema.GroupVersionResource{
			Group:    apirulev1beta1.GroupVersion.Group,
			Version:  apirulev1beta1.GroupVersion.Version,
			Resource: "apirules",
		}, apiRules[0].GetNamespace(), apiRules[0].GetName(), retry.Attempts(1))

		if err != nil {
			return err
		}

		err = runtime.DefaultUnstructuredConverter.FromUnstructured(res.UnstructuredContent(), &apiRuleStructured)
		if err != nil {
			return err
		}

		if apiRuleStructured.Status.APIRuleStatus == nil {
			return fmt.Errorf("status not found")
		}

		if apiRuleStructured.Status.APIRuleStatus.Code != stringToStatus(status) {
			return fmt.Errorf("expected status %s, got %s", status, apiRuleStructured.Status.APIRuleStatus.Code)
		}

		return nil
	}, retry.Attempts(5), retry.Delay(time.Second*2))
}

func stringToStatus(status string) apirulev1beta1.StatusCode {
	switch status {
	case "OK":
		return apirulev1beta1.StatusOK
	case "SKIPPED":
		return apirulev1beta1.StatusSkipped
	case "ERROR":
		return apirulev1beta1.StatusError
	case "WARNING":
		return apirulev1beta1.StatusWarning
	default:
		return apirulev1beta1.StatusError
	}
}
