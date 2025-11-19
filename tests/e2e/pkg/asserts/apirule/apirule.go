package apirule

import (
	"strings"
	"testing"
	"time"

	v2 "github.com/kyma-project/api-gateway/apis/gateway/v2"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
)

func HasState(t *testing.T, name, namespace string, state v2.State) bool {
	t.Helper()

	r, err := client.ResourcesClient(t)
	require.NoError(t, err)

	var apiRule v2.APIRule
	require.NoError(t, r.Get(t.Context(), name, namespace, &apiRule))

	return apiRule.Status.State == state
}

func HasStatusDescription(t *testing.T, name, namespace string, statusDesc string) bool {
	t.Helper()

	r, err := client.ResourcesClient(t)
	require.NoError(t, err)

	var apiRule v2.APIRule
	require.NoError(t, r.Get(t.Context(), name, namespace, &apiRule))

	return apiRule.Status.Description == statusDesc
}
func ContainsInStatusDescription(t *testing.T, name, namespace string, substringStatusDesc string) bool {
	t.Helper()

	r, err := client.ResourcesClient(t)
	require.NoError(t, err)

	var apiRule v2.APIRule
	require.NoError(t, r.Get(t.Context(), name, namespace, &apiRule))
	return strings.Contains(apiRule.Status.Description, substringStatusDesc)
}

func HasAnnotation(t *testing.T, name, namespace string, key, value string) bool {
	t.Helper()

	r, err := client.ResourcesClient(t)
	require.NoError(t, err)

	var apiRule v2.APIRule
	require.NoError(t, r.Get(t.Context(), name, namespace, &apiRule))

	return apiRule.Annotations[key] == value
}

func WaitUntilReady(t *testing.T, name, namespace string) {
	t.Helper()

	r, err := client.ResourcesClient(t)
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

func WaitUntilError(t *testing.T, name, namespace string) {
	t.Helper()

	r, err := client.ResourcesClient(t)
	require.NoError(t, err)

	var apiRule v2.APIRule
	require.NoError(t, r.Get(t.Context(), name, namespace, &apiRule))

	err = wait.For(conditions.New(r).ResourceMatch(&apiRule, func(obj k8s.Object) bool {
		ar, ok := obj.(*v2.APIRule)
		if !ok {
			t.Fatalf("Expected object of type v2.APIRule, got %T", obj)
		}
		return ar.Status.State == v2.Error
	}))
	assert.NoError(t, err)
	if err != nil {
		t.Logf("APIRule %s/%s status: %+v", namespace, name, apiRule.Status)
	}
}

func WaitForStatusDesc(t *testing.T, name, namespace, shouldContain string) {
	t.Helper()

	r, err := client.ResourcesClient(t)
	require.NoError(t, err)

	var apiRule v2.APIRule
	require.NoError(t, r.Get(t.Context(), name, namespace, &apiRule))

	err = wait.For(conditions.New(r).ResourceMatch(&apiRule, func(obj k8s.Object) bool {
		ar, ok := obj.(*v2.APIRule)
		if !ok {
			t.Fatalf("Expected object of type v2.APIRule, got %T", obj)
		}
		return strings.Contains(ar.Status.Description, shouldContain)
		// we introduce timeout bellow that equals 75 sec
		// it's a workaround for the APIRule creation spec
		// some percent of tests runs end up with freshly created APIRule in error status
		// caused by inability to fetch Gateway host (error msg: Validation errors: Attribute 'spec.gateway': Could not get specified Gateway)
		// the reconciliation is requeued for the given APIRule with 60s interval
		// it gets fixed in the next reconciliation
		// therefore assumption is that after 75 seconds APIRule should be ready
		// otherwise it is some other not related problem
	}), wait.WithTimeout(time.Second*75))
	assert.NoError(t, err)
	if err != nil {
		t.Logf("APIRule %s/%s status: %+v", namespace, name, apiRule.Status)
	}
}
