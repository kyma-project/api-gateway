package gateway

import (
	"context"
	"errors"
	"testing"

	"github.com/go-logr/logr"
	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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
	result, err := r.updateStatus(context.Background(), logr.Discard(), apiRule.DeepCopy(), true)

	if !errors.Is(err, errReconcileWithBackoff) {
		t.Fatalf("expected sentinel backoff error, got: %v", err)
	}

	if result.RequeueAfter != 0 {
		t.Fatalf("expected zero explicit requeue delay for rate limited retry, got: %v", result.RequeueAfter)
	}
}
