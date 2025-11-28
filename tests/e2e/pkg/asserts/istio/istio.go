package istio

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	"istio.io/client-go/pkg/apis/security/v1beta1"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"

	"github.com/kyma-project/api-gateway/internal/processing"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/infrastructure"
)

func VirtualServiceOwnedByAPIRuleExists(t *testing.T, resourceNamespace, apiRuleName, apiRuleNamespace string) {
	t.Helper()

	r, err := infrastructure.ResourcesClient(t)
	require.NoError(t, err)

	var virtualServiceList v1alpha3.VirtualServiceList

	err = r.WithNamespace(resourceNamespace).List(
		t.Context(),
		&virtualServiceList,
		resources.WithLabelSelector(fmt.Sprintf("%s=%s", processing.OwnerLabelName, apiRuleName)),
		resources.WithLabelSelector(fmt.Sprintf("%s=%s", processing.OwnerLabelNamespace, apiRuleNamespace)),
	)

	require.NoError(t, err, "Failed to list VirtualServices")
	assert.Equal(t, 1, len(virtualServiceList.Items), "Expected 1 VirtualService")
}

func AuthorizationPolicyOwnedByAPIRuleExists(t *testing.T, resourceNamespace, apiRuleName, apiRuleNamespace string, numberOfPolicies int) {
	t.Helper()

	r, err := infrastructure.ResourcesClient(t)
	require.NoError(t, err)

	var authorizationPolicyList v1beta1.AuthorizationPolicyList

	err = r.WithNamespace(resourceNamespace).List(
		t.Context(),
		&authorizationPolicyList,
		resources.WithLabelSelector(fmt.Sprintf("%s=%s", processing.OwnerLabelName, apiRuleName)),
		resources.WithLabelSelector(fmt.Sprintf("%s=%s", processing.OwnerLabelNamespace, apiRuleNamespace)),
	)

	require.NoError(t, err, "Failed to list AuthorizationPolicies")
	assert.Equal(t, numberOfPolicies, len(authorizationPolicyList.Items), "Expected AuthorizationPolicies", numberOfPolicies)
}
