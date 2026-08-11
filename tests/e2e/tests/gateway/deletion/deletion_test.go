package deletion

import (
	"context"
	_ "embed"
	"fmt"
	"strings"
	"testing"

	"github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	oryv1alpha1 "github.com/kyma-project/api-gateway/internal/types/ory/oathkeeper-maester/api/v1alpha1"
	apiruleasserts "github.com/kyma-project/api-gateway/tests/e2e/pkg/asserts/apirule"
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

		require.NoError(t, modulehelpers.DeleteAPIGateway(t, r, "kyma-system", "default"))
		require.NoError(t, modulehelpers.WaitUntilAPIGatewayDeleted(t, r, "kyma-system", "default"))
	})

	t.Run("Deleting API-Gateway CR with APIRule present", func(t *testing.T) {
		modulehelpers.SetupBaseCR(t)
		testBackground, err := testsetup.SetupRandomNamespaceWithHttpbin(t, testsetup.WithPrefix("deletion-apirule"))
		require.NoError(t, err, "Failed to setup test background with httpbin")

		customGatewayNamespace := "kyma-system"
		customGatewayName := "some-other-gateway"
		customGatewayHost := fmt.Sprintf("custom-gw-%s.example.com", testBackground.TestName)

		err = setupCustomGateway(t, customGatewayNamespace, customGatewayName, customGatewayHost)
		require.NoError(t, err, "Failed to create custom Istio Gateway")

		kymaGatewayDomain, err := domain.GetFromGateway(t, customGatewayName, customGatewayNamespace)
		require.NoError(t, err, "Failed to get domain from custom gateway")

		apiRuleName := "kyma-rule"
		host := fmt.Sprintf("httpbin-%s.%s", testBackground.TestName, kymaGatewayDomain)
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
		require.NoError(t, modulehelpers.DeleteAPIGateway(t, r, "kyma-system", "default"))

		require.NoError(t, modulehelpers.WaitUntilAPIGatewayExists(t, r, "kyma-system", "default"), "APIGateway CR should still exist while APIRule is present")
		require.NoError(t, modulehelpers.WaitUntilAPIGatewayCRHasState(t, r, "kyma-system", "default", v1alpha1.Warning, "There are APIRule(s) that block the deletion of API-Gateway CR. Please take a look at kyma-system/api-gateway-controller-manager logs to see more information about the warning"),
			"APIGateway CR should be in Warning state while APIRule is present")

		require.NoError(t, r.Delete(t.Context(), apiRule), "Failed to delete APIRule resource")

		require.NoError(t, modulehelpers.WaitUntilAPIGatewayDeleted(t, r, "kyma-system", "default"))
	})

	t.Run("Deleting API-Gateway CR with ORY Oathkeeper Rule present", func(t *testing.T) {
		r, err := e2eclient.ResourcesClient(t)
		require.NoError(t, err)
		require.NoError(t, e2eclient.RegisterAdditionalSchemes(r, oryv1alpha1.AddToScheme))
		require.NoError(t, v1access.CreateAllowAPIRuleV1Signatures(context.Background(), r, t))
		modulehelpers.SetupBaseCR(t)

		const oryRuleNamespace = "kyma-system"
		const oryRuleName = "ory-rule"

		_, err = infrahelpers.CreateResourceWithTemplateValues(
			t,
			ORYRule,
			map[string]any{
				"Name":      oryRuleName,
				"Namespace": oryRuleNamespace,
			},
		)
		if err != nil {
			if strings.Contains(err.Error(), "no matches for oathkeeper.ory.sh/v1alpha1") {
				require.FailNow(t, "oathkeeper.ory.sh/v1alpha1 Rule CRD is not installed in this cluster")
			}
			require.NoError(t, err, "Failed to create ORY Rule resource")
		}

		var currentOryRule oryv1alpha1.Rule
		require.NoError(t, r.Get(t.Context(), oryRuleName, oryRuleNamespace, &currentOryRule))

		require.NoError(t, modulehelpers.DeleteAPIGateway(t, r, "kyma-system", "default"))

		require.NoError(t, modulehelpers.WaitUntilAPIGatewayExists(t, r, "kyma-system", "default"), "APIGateway CR should still exist while ORY Rule is present")
		require.NoError(t, modulehelpers.WaitUntilAPIGatewayCRHasState(t, r, "kyma-system", "default", v1alpha1.Warning, "There are ORY Oathkeeper Rule(s) that block the deletion of API-Gateway CR. Please take a look at kyma-system/api-gateway-controller-manager logs to see more information about the warning"),
			"APIGateway CR should be in Warning state while ORY Rule is present")

		require.NoError(t, r.Delete(t.Context(), &currentOryRule), "Failed to delete ORY Rule resource")

		require.NoError(t, modulehelpers.WaitUntilAPIGatewayDeleted(t, r, "kyma-system", "default"))

	})
}
