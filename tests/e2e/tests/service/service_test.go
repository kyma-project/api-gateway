package service

import (
	_ "embed"
	"fmt"
	"net/http"
	"testing"

	apiruleasserts "github.com/kyma-project/api-gateway/tests/e2e/pkg/asserts/apirule"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/domain"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/httpbin"
	infrahelpers "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/infrastructure"
	modulehelpers "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/modules"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/oauth2"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/testsetup"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/e2e-framework/klient/decoder"
)

//go:embed service_fallback.yaml
var APIRuleServiceFallback string

//go:embed service_two_namespaces.yaml
var APIRuleServiceTwoNamespaces string

//go:embed service_diff_same_methods.yaml
var APIRuleServiceDiffSameMethods string

//go:embed service_custom_label_selector.yaml
var APIRuleServiceCustomLabelSelector string

func TestAPIRuleDifferentServices(t *testing.T) {
	require.NoError(t, modulehelpers.CreateIstioOperatorCR(t))
	require.NoError(t, modulehelpers.CreateApiGatewayCR(t))

	t.Run("Endpoints exposed in APIRule should fallback to service defined on root level when there is no service defined on rule level", func(t *testing.T) {
		t.Parallel()
		testBackground, err := testsetup.SetupRandomNamespaceWithOauth2MockAndHttpbin(t, testsetup.WithPrefix("services"))
		require.NoError(t, err, "Failed to setup test background with httpbin")
		kymaGatewayDomain, err := domain.GetFromGateway(t, "kyma-gateway", "kyma-system")
		require.NoError(t, err, "Failed to get domain from kyma-gateway")

		createdApirule, err := infrahelpers.CreateResourceWithTemplateValues(
			t,
			APIRuleServiceFallback,
			map[string]any{
				"Name":                         testBackground.TestName,
				"Host":                         testBackground.TestName,
				"ServiceName":                  testBackground.TargetServiceName,
				"ServicePort":                  testBackground.TargetServicePort,
				"Gateway":                      "kyma-system/kyma-gateway",
				"Issuer":                       testBackground.Provider.GetIssuerURL(),
				"JwksUri":                      testBackground.Provider.GetJwksURI(),
				"JwtSecuredPathWithService":    "/headers",
				"JwtSecuredPathWithoutService": "/ip",
			},
			decoder.MutateNamespace(testBackground.Namespace),
		)
		require.NoError(t, err, "Failed to create APIRule resource")
		require.NotEmpty(t, createdApirule, "Created APIRule resource should not be empty")

		apiruleasserts.WaitUntilReady(t, testBackground.TestName, testBackground.Namespace)

		urlWithServiceDefined := fmt.Sprintf("https://%s.%s%s", testBackground.TestName, kymaGatewayDomain, "/headers")
		oauth2.AssertEndpointWithProvider(
			t,
			testBackground.Provider,
			urlWithServiceDefined,
			http.MethodGet,
		)

		urlWithoutServiceDefined := fmt.Sprintf("https://%s.%s%s", testBackground.TestName, kymaGatewayDomain, "/ip")
		oauth2.AssertEndpointWithProvider(
			t,
			testBackground.Provider,
			urlWithoutServiceDefined,
			http.MethodGet,
		)
	})

	t.Run("Exposing endpoints in two namespaces", func(t *testing.T) {
		t.Parallel()
		testBackground, err := testsetup.SetupRandomNamespaceWithOauth2MockAndHttpbin(t, testsetup.WithPrefix("service"))
		require.NoError(t, err, "Failed to setup test background with oauth and httpbin")
		testBackgroundOtherNamespace, err := testsetup.SetupRandomNamespaceWithHttpbin(t, testsetup.WithPrefix("service"))
		require.NoError(t, err, "Failed to setup test background with httpbin")
		kymaGatewayDomain, err := domain.GetFromGateway(t, "kyma-gateway", "kyma-system")
		require.NoError(t, err, "Failed to get domain from kyma-gateway")

		createdApirule, err := infrahelpers.CreateResourceWithTemplateValues(
			t,
			APIRuleServiceTwoNamespaces,
			map[string]any{
				"Name":                         testBackground.TestName,
				"Host":                         testBackground.TestName,
				"ServiceName":                  testBackground.TargetServiceName,
				"ServicePort":                  testBackground.TargetServicePort,
				"ServiceNamespace":             testBackground.Namespace,
				"Gateway":                      "kyma-system/kyma-gateway",
				"Issuer":                       testBackground.Provider.GetIssuerURL(),
				"JwksUri":                      testBackground.Provider.GetJwksURI(),
				"JwtSecuredPath":               "/get",
				"OtherNamespaceJwtSecuredPath": "/ip",
				"OtherNamespace":               testBackgroundOtherNamespace.Namespace,
				"OtherNamespaceServiceName":    testBackgroundOtherNamespace.TargetServiceName,
				"OtherNamespaceServicePort":    testBackgroundOtherNamespace.TargetServicePort,
			},
			decoder.MutateNamespace(testBackground.Namespace),
		)
		require.NoError(t, err, "Failed to create APIRule resource")
		require.NotEmpty(t, createdApirule, "Created APIRule resource should not be empty")

		apiruleasserts.WaitUntilReady(t, testBackground.TestName, testBackground.Namespace)

		url := fmt.Sprintf("https://%s.%s%s", testBackground.TestName, kymaGatewayDomain, "/get")
		urlOtherNamespace := fmt.Sprintf("https://%s.%s%s", testBackground.TestName, kymaGatewayDomain, "/ip")

		oauth2.AssertEndpointWithProvider(
			t,
			testBackground.Provider,
			url,
			http.MethodGet,
		)
		oauth2.AssertEndpointWithProvider(
			t,
			testBackground.Provider,
			urlOtherNamespace,
			http.MethodGet,
		)
	})

	t.Run("Exposing different services with same methods", func(t *testing.T) {
		t.Parallel()
		testBackground, err := testsetup.SetupRandomNamespaceWithOauth2MockAndHttpbin(t, testsetup.WithPrefix("services"))
		require.NoError(t, err, "Failed to setup test background with httpbin")
		svcName, svcPort, err := httpbin.DeploySecondHttpbin(t, testBackground.Namespace)
		require.NoError(t, err, "Failed to deploy second httpbin service")

		kymaGatewayDomain, err := domain.GetFromGateway(t, "kyma-gateway", "kyma-system")
		require.NoError(t, err, "Failed to get domain from kyma-gateway")

		createdApirule, err := infrahelpers.CreateResourceWithTemplateValues(
			t,
			APIRuleServiceDiffSameMethods,
			map[string]any{
				"Name":              testBackground.TestName,
				"Host":              testBackground.TestName,
				"ServiceName":       testBackground.TargetServiceName,
				"ServicePort":       testBackground.TargetServicePort,
				"SecondServiceName": svcName,
				"SecondServicePort": svcPort,
				"Gateway":           "kyma-system/kyma-gateway",
				"Issuer":            testBackground.Provider.GetIssuerURL(),
				"JwksUri":           testBackground.Provider.GetJwksURI(),
			},
			decoder.MutateNamespace(testBackground.Namespace),
		)
		require.NoError(t, err, "Failed to create APIRule resource")
		require.NotEmpty(t, createdApirule, "Created APIRule resource should not be empty")

		apiruleasserts.WaitUntilReady(t, testBackground.TestName, testBackground.Namespace)

		requests := []struct {
			path                   string
			method                 string
			expectedResponseStatus int
		}{
			{path: "/ip", method: http.MethodGet, expectedResponseStatus: http.StatusOK},
			{path: "/headers", method: http.MethodGet, expectedResponseStatus: http.StatusOK},
			{path: "/ip", method: http.MethodPost, expectedResponseStatus: http.StatusOK},
			{path: "/headers", method: http.MethodPost, expectedResponseStatus: http.StatusOK},
		}

		for _, r := range requests {
			url := fmt.Sprintf("https://%s.%s%s", testBackground.TestName, kymaGatewayDomain, r.path)
			oauth2.AssertEndpointWithProvider(
				t,
				testBackground.Provider,
				url,
				r.method,
			)
		}
	})

}
