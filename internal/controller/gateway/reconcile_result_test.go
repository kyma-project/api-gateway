package gateway

import (
	"context"
	"errors"
	"testing"

	"github.com/go-logr/logr"
	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	meta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestUpdateStatus_ReconcileErrorTriggersBackoff(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := gatewayv1beta1.AddToScheme(scheme); err != nil {
		t.Fatalf("add scheme: %v", err)
	}

	apiRule := &gatewayv1beta1.APIRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-apirule",
			Namespace: "default",
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(apiRule).
		WithObjects(apiRule).
		Build()

	r := &APIRuleReconciler{Client: fakeClient}
	result, err := r.updateStatus(context.Background(), logr.Discard(), apiRule.DeepCopy(), nil, true, false)

	if !errors.Is(err, errReconcileWithBackoff) {
		t.Fatalf("expected sentinel backoff error, got: %v", err)
	}

	if result.RequeueAfter != 0 {
		t.Fatalf("expected zero explicit requeue delay for rate limited retry, got: %v", result.RequeueAfter)
	}
}

func TestUpdateStatus_RemoveStalePrunesConditionsAbsentFromFreshSet(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := gatewayv2alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("add scheme: %v", err)
	}

	// The stored CR carries a RequestAuthenticationConfigured condition from a prior
	// pass. This pass does not emit RA, so the object passed to updateStatus omits it;
	// the terminal write (removeStale=true) must prune the stale condition.
	stored := &gatewayv2alpha1.APIRule{
		ObjectMeta: metav1.ObjectMeta{Name: "test-apirule", Namespace: "default"},
		Status: gatewayv2alpha1.APIRuleStatus{
			Conditions: []metav1.Condition{
				gatewayv2alpha1.VirtualServiceConfiguredCondition(1),
				gatewayv2alpha1.RequestAuthenticationConfiguredCondition(1),
				gatewayv2alpha1.AuthorizationPolicyConfiguredCondition(1),
				gatewayv2alpha1.DependenciesReadyCondition(1),
				gatewayv2alpha1.ReadyCondition(1),
			},
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(stored).
		WithObjects(stored).
		Build()

	// toUpdate is the cached object: it still carries the stale RA condition (as the
	// upsert-merge of this pass's fresh set onto the prior status would). The pure
	// desired set passed as the arg omits RA — that arg, not the object, is the prune
	// authority. This reproduces the cache-contamination bug: deriving the authority
	// from toUpdate.Status.Conditions would wrongly keep RA.
	toUpdate := &gatewayv2alpha1.APIRule{
		ObjectMeta: metav1.ObjectMeta{Name: "test-apirule", Namespace: "default"},
		Status: gatewayv2alpha1.APIRuleStatus{
			Conditions: []metav1.Condition{
				gatewayv2alpha1.VirtualServiceConfiguredCondition(1),
				gatewayv2alpha1.RequestAuthenticationConfiguredCondition(1),
				gatewayv2alpha1.AuthorizationPolicyConfiguredCondition(1),
				gatewayv2alpha1.DependenciesReadyCondition(1),
				gatewayv2alpha1.ReadyCondition(1),
			},
		},
	}

	// Pure desired set this pass: everything except RA.
	desired := []metav1.Condition{
		gatewayv2alpha1.VirtualServiceConfiguredCondition(1),
		gatewayv2alpha1.AuthorizationPolicyConfiguredCondition(1),
		gatewayv2alpha1.DependenciesReadyCondition(1),
		gatewayv2alpha1.ReadyCondition(1),
	}

	r := &APIRuleReconciler{Client: fakeClient}
	if _, err := r.updateStatus(context.Background(), logr.Discard(), toUpdate, desired, false, true); err != nil {
		t.Fatalf("updateStatus: %v", err)
	}

	got := &gatewayv2alpha1.APIRule{}
	if err := fakeClient.Get(context.Background(), client.ObjectKeyFromObject(stored), got); err != nil {
		t.Fatalf("get: %v", err)
	}

	if meta.FindStatusCondition(got.Status.Conditions, gatewayv2alpha1.ConditionTypeRequestAuthentication) != nil {
		t.Errorf("expected stale RequestAuthenticationConfigured to be pruned")
	}
	for _, typ := range []string{
		gatewayv2alpha1.ConditionTypeVirtualService,
		gatewayv2alpha1.ConditionTypeAuthorizationPolicy,
		gatewayv2alpha1.ConditionTypeDependencies,
		gatewayv2alpha1.ConditionTypeReady,
	} {
		if meta.FindStatusCondition(got.Status.Conditions, typ) == nil {
			t.Errorf("expected fresh condition %q to be kept", typ)
		}
	}
}

func TestUpdateStatus_NoPruneWhenRemoveStaleFalse(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := gatewayv2alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("add scheme: %v", err)
	}

	stored := &gatewayv2alpha1.APIRule{
		ObjectMeta: metav1.ObjectMeta{Name: "test-apirule", Namespace: "default"},
		Status: gatewayv2alpha1.APIRuleStatus{
			Conditions: []metav1.Condition{
				gatewayv2alpha1.RequestAuthenticationConfiguredCondition(1),
			},
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(stored).
		WithObjects(stored).
		Build()

	// Desired set omits RA, but removeStale=false (error/processing write), so the
	// stored RA condition must survive — merge upserts, it never prunes.
	toUpdate := &gatewayv2alpha1.APIRule{
		ObjectMeta: metav1.ObjectMeta{Name: "test-apirule", Namespace: "default"},
		Status: gatewayv2alpha1.APIRuleStatus{
			State:       gatewayv2alpha1.Error,
			Description: "some error",
		},
	}

	r := &APIRuleReconciler{Client: fakeClient}
	if _, err := r.updateStatus(context.Background(), logr.Discard(), toUpdate, nil, true, false); err != nil {
		if !errors.Is(err, errReconcileWithBackoff) {
			t.Fatalf("updateStatus: %v", err)
		}
	}

	got := &gatewayv2alpha1.APIRule{}
	if err := fakeClient.Get(context.Background(), client.ObjectKeyFromObject(stored), got); err != nil {
		t.Fatalf("get: %v", err)
	}
	if meta.FindStatusCondition(got.Status.Conditions, gatewayv2alpha1.ConditionTypeRequestAuthentication) == nil {
		t.Errorf("expected condition to be kept when removeStale is false")
	}
}
