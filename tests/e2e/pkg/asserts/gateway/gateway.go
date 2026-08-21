package gateway

import (
	"strings"
	"testing"
	"time"

	"github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
)

const assertTimeout = 2 * time.Minute

// AssertAPIGatewayCRReady polls until the named APIGateway CR reaches Ready state.
func AssertAPIGatewayCRReady(t *testing.T, r *resources.Resources, namespace, name string) {
	t.Helper()
	assertState(t, r, namespace, name, v1alpha1.Ready)
}

// AssertAPIGatewayCRWarning polls until the named APIGateway CR reaches Warning state.
func AssertAPIGatewayCRWarning(t *testing.T, r *resources.Resources, namespace, name string) {
	t.Helper()
	assertState(t, r, namespace, name, v1alpha1.Warning)
}

// AssertAPIGatewayCRDescriptionContains polls until the CR's status.description contains substr.
func AssertAPIGatewayCRDescriptionContains(t *testing.T, r *resources.Resources, namespace, name string, substr string) {
	t.Helper()
	cr := &v1alpha1.APIGateway{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace}}
	err := wait.For(
		conditions.New(r).ResourceMatch(cr, func(obj k8s.Object) bool {
			ag, ok := obj.(*v1alpha1.APIGateway)
			return ok && strings.Contains(ag.Status.Description, substr)
		}),
		wait.WithTimeout(assertTimeout),
	)
	require.NoError(t, err, "APIGateway CR %s/%s status description did not contain %q", namespace, name, substr)
}

// AssertAPIGatewayCRDeleted polls until the named APIGateway CR is gone.
func AssertAPIGatewayCRDeleted(t *testing.T, r *resources.Resources, namespace, name string) {
	t.Helper()
	cr := &v1alpha1.APIGateway{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace}}
	err := wait.For(conditions.New(r).ResourceDeleted(cr), wait.WithTimeout(assertTimeout))
	require.NoError(t, err, "APIGateway CR %s/%s was not deleted", namespace, name)
}

func assertState(t *testing.T, r *resources.Resources, namespace, name string, expectedState v1alpha1.State) {
	t.Helper()
	cr := &v1alpha1.APIGateway{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace}}
	err := wait.For(
		conditions.New(r).ResourceMatch(cr, func(obj k8s.Object) bool {
			ag, ok := obj.(*v1alpha1.APIGateway)
			return ok && ag.Status.State == expectedState
		}),
		wait.WithTimeout(assertTimeout),
	)
	require.NoError(t, err, "APIGateway CR %s/%s did not reach state %s", namespace, name, expectedState)
}
