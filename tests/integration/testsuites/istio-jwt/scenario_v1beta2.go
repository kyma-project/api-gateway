package istiojwt

import (
	"context"
	"fmt"

	"github.com/avast/retry-go/v4"
	"github.com/cucumber/godog"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/manifestprocessor"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/testcontext"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func initV1Beta2IstioJWT(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("v1beta2-istio-jwt.yaml", "v1beta2-istio-jwt")

	ctx.Step(`v1beta2IstioJWT: There is a httpbin service$`, scenario.thereIsAHttpbinService)
	ctx.Step(`v1beta2IstioJWT: There is an endpoint secured with JWT on path "([^"]*)"`, scenario.thereIsAnJwtSecuredPath)
	ctx.Step(`v1beta2IstioJWT: The APIRule is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(`v1beta2IstioJWT: Calling the "([^"]*)" endpoint without a token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithoutTokenShouldResultInStatusBetween)
	ctx.Step(`v1beta2IstioJWT: Calling the "([^"]*)" endpoint with an invalid token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithInvalidTokenShouldResultInStatusBetween)
	ctx.Step(`v1beta2IstioJWT: Calling the "([^"]*)" endpoint with a valid "([^"]*)" token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithValidTokenShouldResultInStatusBetween)
	ctx.Step(`v1beta2IstioJWT: Teardown httpbin service$`, scenario.teardownHttpbinService)
}

func initV1Beta2NoAuthHandler(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("v1beta2-no-auth-handler.yaml", "v1beta2-no-auth-handler")

	ctx.Step(`^v1beta2NoAuthHandler: There is a httpbin service$`, scenario.thereIsAHttpbinService)
	ctx.Step(`^v1beta2NoAuthHandler: The APIRule is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(`^v1beta2NoAuthHandler: Calling the "([^"]*)" endpoint with "([^"]*)" method with any token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithMethodWithInvalidTokenShouldResultInStatusBetween)
	ctx.Step(`^v1beta2NoAuthHandler: Teardown httpbin service$`, scenario.teardownHttpbinService)
}

func initV1Beta2NoAuthHandlerRecover(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("v1beta2-no-auth-handler.yaml", "v1beta2-no-auth-handler-recover")

	ctx.Step(`^v1beta2NoAuthHandlerRecover: There is a httpbin service$`, scenario.thereIsAHttpbinService)
	ctx.Step(`^v1beta2NoAuthHandlerRecover: The APIRule is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(`^v1beta2NoAuthHandlerRecover: Certificate secret is reset$`, scenario.certificateSecretReset)
	ctx.Step(`^v1beta2NoAuthHandlerRecover: Certificate secret is rotated$`, scenario.certificateSecretRotated)
	ctx.Step(`^v1beta2NoAuthHandlerRecover: Calling the "([^"]*)" endpoint with "([^"]*)" method with any token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithMethodWithInvalidTokenShouldResultInStatusBetween)
	ctx.Step(`^v1beta2NoAuthHandlerRecover: Teardown httpbin service$`, scenario.teardownHttpbinService)
}

func (s *scenario) certificateSecretReset() error {
	r, err := manifestprocessor.ParseFromFileWithTemplate("certificate-secret.yaml", s.ApiResourceDirectory, s.ManifestTemplate)
	if err != nil {
		return err
	}

	_, err = s.resourceManager.CreateOrUpdateResources(s.k8sClient, r...)
	if err != nil {
		return err
	}

	return nil
}

func (s *scenario) certificateSecretRotated() error {
	secretGVR := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "secrets"}

	return retry.Do(func() error {
		unstructured, err := s.k8sClient.Resource(secretGVR).Namespace("kyma-system").Get(context.Background(), "api-gateway-webhook-certificate", v1.GetOptions{})
		if err != nil {
			return errors.Wrap(err, "could not get certificate secret")
		}

		var secret corev1.Secret
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(unstructured.Object, &secret)
		if err != nil {
			return errors.Wrap(err, "could not convert unstructured to secret")
		}

		if len(secret.Data["tls.crt"]) == 0 || len(secret.Data["tls.key"]) == 0 {
			return fmt.Errorf("certificate secret is empty")
		}
		return nil
	}, testcontext.GetRetryOpts()...)
}
