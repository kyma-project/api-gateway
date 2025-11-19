package validation

import (
	_ "embed"
	"testing"

	apiruleasserts "github.com/kyma-project/api-gateway/tests/e2e/pkg/asserts/apirule"
	infrahelpers "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/infrastructure"
	modulehelpers "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/modules"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/testsetup"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/e2e-framework/klient/decoder"
)

//go:embed validation_error.yaml
var APIRuleValidationError string

func TestAPIRuleValidation(t *testing.T) {
	require.NoError(t, modulehelpers.CreateIstioOperatorCR(t))
	require.NoError(t, modulehelpers.CreateApiGatewayCR(t))

	t.Run("Validation errors on misconfigured APIRule with multiple jwt configurations", func(t *testing.T) {
		testBackground, err := testsetup.SetupRandomNamespaceWithOauth2MockAndHttpbin(t, testsetup.WithPrefix("asterisk"))
		require.NoError(t, err, "Failed to setup test background with httpbin")

		createdApirule, err := infrahelpers.CreateResourceWithTemplateValues(
			t,
			APIRuleValidationError,
			map[string]any{
				"Name":        testBackground.TestName,
				"Host":        testBackground.TestName,
				"ServiceName": testBackground.TargetServiceName,
				"ServicePort": testBackground.TargetServicePort,
				"Gateway":     "kyma-system/kyma-gateway",
				"Issuer":      testBackground.Provider.GetIssuerURL(),
				"JwksUri":     testBackground.Provider.GetJwksURI(),
			},
			decoder.MutateNamespace(testBackground.Namespace),
		)
		require.NoError(t, err, "Failed to create APIRule resource")
		require.NotEmpty(t, createdApirule, "Created APIRule resource should not be empty")

		apiruleasserts.WaitForStatusDesc(t, testBackground.TestName, testBackground.Namespace, "multiple jwt configurations")
	})
}
