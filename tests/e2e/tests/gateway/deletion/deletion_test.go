package deletion

import (
	"context"
	_ "embed"
	"fmt"
	"strings"
	"testing"

	oryv1alpha1 "github.com/kyma-project/api-gateway/internal/types/ory/oathkeeper-maester/api/v1alpha1"
	apiruleasserts "github.com/kyma-project/api-gateway/tests/e2e/pkg/asserts/apirule"
	apigatewayasserts "github.com/kyma-project/api-gateway/tests/e2e/pkg/asserts/gateway"
	e2eclient "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/client"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/domain"
	extgwhelper "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/extgateway"
	infrahelpers "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/infrastructure"
	modulehelpers "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/modules"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/testsetup"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/v1access"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/e2e-framework/klient/decoder"
)

const (
	apiGatewayCRNamespace = "kyma-system"
	apiGatewayCRName      = "default"
)

//go:embed apirule.yaml
var APIRule string

//go:embed oryrule.yaml
var ORYRule string

//go:embed customGw.yaml
var customDomainGatewayTemplate string

// setupCustomGateway creates a custom Istio Gateway with a self-signed TLS certificate
// for the given host. This is required before calling domain.GetFromGateway.
func setupCustomGateway(t *testing.T, namespace, name, host string) error {
	t.Helper()

	certPEM, keyPEM, err := extgwhelper.GenerateServerTLSCert(t, host)
	if err != nil {
		return fmt.Errorf("generating server TLS cert for %s: %w", host, err)
	}

	tlsSecretName := name + "-tls"
	if err := extgwhelper.CreateServerTLSSecret(t, tlsSecretName, certPEM, keyPEM); err != nil {
		return fmt.Errorf("creating TLS secret %s: %w", tlsSecretName, err)
	}

	_, err = infrahelpers.CreateResourceWithTemplateValues(
		t,
		customDomainGatewayTemplate,
		map[string]any{
			"Name":          name,
			"Host":          host,
			"TLSSecretName": tlsSecretName,
		},
		decoder.MutateNamespace(namespace),
	)
	if err != nil {
		return fmt.Errorf("creating custom Istio Gateway %s/%s: %w", namespace, name, err)
	}

	return nil
}

func TestDeletion(t *testing.T) {
	t.Run("Deleting API-Gateway CR without blocking resources", func(t *testing.T) {
		modulehelpers.SetupBaseCR(t)
		r, err := e2eclient.ResourcesClient(t)
		require.NoError(t, err)

		require.NoError(t, modulehelpers.DeleteAPIGateway(t, r, apiGatewayCRNamespace, apiGatewayCRName))
		apigatewayasserts.AssertAPIGatewayCRDeleted(t, r, apiGatewayCRNamespace, apiGatewayCRName)
	})

	t.Run("Deleting API-Gateway CR with APIRule present", func(t *testing.T) {
		modulehelpers.SetupBaseCR(t)
		testBackground, err := testsetup.SetupRandomNamespaceWithHttpbin(t, testsetup.WithPrefix("deletion-apirule"))
		require.NoError(t, err, "Failed to setup test background with httpbin")

		customGatewayNamespace := apiGatewayCRNamespace
		customGatewayName := testBackground.TestName + "-gw"
		customGatewayHost := fmt.Sprintf("custom-gw-%s.example.com", testBackground.TestName)

		err = setupCustomGateway(t, customGatewayNamespace, customGatewayName, customGatewayHost)
		require.NoError(t, err, "Failed to create custom Istio Gateway")

		customGatewayDomain, err := domain.GetFromGateway(t, customGatewayName, customGatewayNamespace)
		require.NoError(t, err, "Failed to get domain from custom gateway")

		apiRuleName := "kyma-rule"
		host := fmt.Sprintf("httpbin-%s.%s", testBackground.TestName, customGatewayDomain)
		apiRule, err := infrahelpers.CreateResourceWithTemplateValues(
			t,
			APIRule,
			map[string]any{
				"Name":        apiRuleName,
				"Host":        host,
				"Gateway":     fmt.Sprintf("%s/%s", customGatewayNamespace, customGatewayName),
				"ServiceName": testBackground.TargetServiceName,
				"ServicePort": testBackground.TargetServicePort,
			},
			decoder.MutateNamespace(testBackground.Namespace),
		)
		require.NoError(t, err, "Failed to create APIRule resource")

		apiruleasserts.WaitUntilReady(t, apiRuleName, testBackground.Namespace)

		r, err := e2eclient.ResourcesClient(t)
		require.NoError(t, err)
		require.NoError(t, modulehelpers.DeleteAPIGateway(t, r, apiGatewayCRNamespace, apiGatewayCRName))

		require.NoError(t, modulehelpers.AssertAPIGatewayExists(t, r, apiGatewayCRNamespace, apiGatewayCRName), "APIGateway CR should still exist while APIRule is present")
		apigatewayasserts.AssertAPIGatewayCRWarning(t, r, apiGatewayCRNamespace, apiGatewayCRName)
		apigatewayasserts.AssertAPIGatewayCRDescriptionContains(t, r, apiGatewayCRNamespace, apiGatewayCRName, "There are APIRule(s) that block the deletion of API-Gateway CR")

		require.NoError(t, r.Delete(t.Context(), apiRule), "Failed to delete APIRule resource")

		apigatewayasserts.AssertAPIGatewayCRDeleted(t, r, apiGatewayCRNamespace, apiGatewayCRName)
	})

	t.Run("Deleting API-Gateway CR with ORY Oathkeeper Rule present", func(t *testing.T) {
		r, err := e2eclient.ResourcesClient(t)
		require.NoError(t, err)
		require.NoError(t, e2eclient.RegisterAdditionalSchemes(r, oryv1alpha1.AddToScheme))
		require.NoError(t, v1access.CreateAllowAPIRuleV1Signatures(context.Background(), r, t))
		modulehelpers.SetupBaseCR(t)

		const oryRuleNamespace = apiGatewayCRNamespace
		const oryRuleName = "ory-rule"

		_, err = infrahelpers.CreateResourceWithTemplateValues(
			t,
			ORYRule,
			map[string]any{
				"Namespace": oryRuleNamespace,
			},
		)
		if err != nil {
			if strings.Contains(err.Error(), "no matches for oathkeeper.ory.sh/v1alpha1") {
				t.Skip("oathkeeper.ory.sh/v1alpha1 Rule CRD is not installed in this cluster")
			}
			require.NoError(t, err, "Failed to create ORY Rule resource")
		}

		var currentOryRule oryv1alpha1.Rule
		require.NoError(t, r.Get(t.Context(), oryRuleName, oryRuleNamespace, &currentOryRule))

		require.NoError(t, modulehelpers.DeleteAPIGateway(t, r, apiGatewayCRNamespace, apiGatewayCRName))

		require.NoError(t, modulehelpers.AssertAPIGatewayExists(t, r, apiGatewayCRNamespace, apiGatewayCRName), "APIGateway CR should still exist while ORY Rule is present")
		apigatewayasserts.AssertAPIGatewayCRWarning(t, r, apiGatewayCRNamespace, apiGatewayCRName)
		apigatewayasserts.AssertAPIGatewayCRDescriptionContains(t, r, apiGatewayCRNamespace, apiGatewayCRName, "There are ORY Oathkeeper Rule(s) that block the deletion of API-Gateway CR")

		require.NoError(t, r.Delete(t.Context(), &currentOryRule), "Failed to delete ORY Rule resource")

		apigatewayasserts.AssertAPIGatewayCRDeleted(t, r, apiGatewayCRNamespace, apiGatewayCRName)
	})
}
