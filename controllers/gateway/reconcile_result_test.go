package gateway

import (
	"context"
	"errors"
	"testing"

	"github.com/go-logr/logr"
	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestUpdateStatus_ReconcileErrorTriggersBackoff(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := gatewayv2alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("add scheme: %v", err)
	}

	apiRule := &gatewayv2alpha1.APIRule{
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
	result, err := r.updateStatus(context.Background(), logr.Discard(), apiRule.DeepCopy(), true)

	if !errors.Is(err, errReconcileWithBackoff) {
		t.Fatalf("expected sentinel backoff error, got: %v", err)
	}

	if result.RequeueAfter != 0 {
		t.Fatalf("expected zero explicit requeue delay for rate limited retry, got: %v", result.RequeueAfter)
	}
}
