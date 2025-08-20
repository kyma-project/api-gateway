package apirule

import (
	v2 "github.com/kyma-project/api-gateway/apis/gateway/v2"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/infrastructure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"testing"
)

func HasState(t *testing.T, name, namespace string, state v2.State) bool {
	t.Helper()

	r, err := infrastructure.ResourcesClient(t)
	require.NoError(t, err)

	var apiRule v2.APIRule
	require.NoError(t, r.Get(t.Context(), name, namespace, &apiRule))

	return apiRule.Status.State == state
}

func WaitUntilReady(t *testing.T, name, namespace string) {
	t.Helper()

	r, err := infrastructure.ResourcesClient(t)
	require.NoError(t, err)

	var apiRule v2.APIRule
	require.NoError(t, r.Get(t.Context(), name, namespace, &apiRule))

	err = wait.For(conditions.New(r).ResourceMatch(&apiRule, func(obj k8s.Object) bool {
		ar, ok := obj.(*v2.APIRule)
		if !ok {
			t.Fatalf("Expected object of type v2.APIRule, got %T", obj)
		}
		return ar.Status.State == v2.Ready
	}))
	assert.NoError(t, err)
	if err != nil {
		t.Logf("APIRule %s/%s status: %+v", namespace, name, apiRule.Status)
	}
}
