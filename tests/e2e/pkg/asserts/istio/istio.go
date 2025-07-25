package istio

import (
	"fmt"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/infrastructure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	"istio.io/client-go/pkg/apis/security/v1beta1"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"testing"
)

func VirtualServiceOwnedByAPIRuleExists(t *testing.T, resourceNamespace, apiRuleName, apiRuleNamespace string) {
	t.Helper()

	r, err := infrastructure.ResourcesClient(t)
	require.NoError(t, err)

	ownerLabel := fmt.Sprintf("apirule.gateway.kyma-project.io/v1beta1=%s.%s", apiRuleName, apiRuleNamespace)
	var virtualServiceList v1alpha3.VirtualServiceList

	err = r.WithNamespace(resourceNamespace).List(
		t.Context(),
		&virtualServiceList,
		resources.WithLabelSelector(ownerLabel),
		)

	require.NoError(t, err, "Failed to list VirtualServices")
	assert.Equal(t, 1, len(virtualServiceList.Items), "Expected 1 VirtualService")
}

func AuthorizationPolicyOwnedByAPIRuleExists(t *testing.T, resourceNamespace, apiRuleName, apiRuleNamespace string) {
	t.Helper()

	r, err := infrastructure.ResourcesClient(t)
	require.NoError(t, err)

	ownerLabel := fmt.Sprintf("apirule.gateway.kyma-project.io/v1beta1=%s.%s", apiRuleName, apiRuleNamespace)
	var authorizationPolicyList v1beta1.AuthorizationPolicyList

	err = r.WithNamespace(resourceNamespace).List(
		t.Context(),
		&authorizationPolicyList,
		resources.WithLabelSelector(ownerLabel),
	)

	require.NoError(t, err, "Failed to list AuthorizationPolicies")
	assert.Equal(t, 1, len(authorizationPolicyList.Items), "Expected 1 AuthorizationPolicy")
}