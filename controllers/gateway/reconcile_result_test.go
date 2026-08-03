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

func TestIsAPIRuleV2_V1beta1AnnotationReturnsFalse(t *testing.T) {
	apiRule := &gatewayv2alpha1.APIRule{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				"gateway.kyma-project.io/original-version": "v1beta1",
			},
		},
	}
	if isAPIRuleV2(apiRule) {
		t.Fatal("expected isAPIRuleV2 to return false for v1beta1-annotated APIRule")
	}
}

func TestIsAPIRuleV2_NoAnnotationReturnsTrue(t *testing.T) {
	apiRule := &gatewayv2alpha1.APIRule{}
	if !isAPIRuleV2(apiRule) {
		t.Fatal("expected isAPIRuleV2 to return true when annotation is absent")
	}
}

func TestIsAPIRuleV2_NonV1beta1AnnotationReturnsTrue(t *testing.T) {
	apiRule := &gatewayv2alpha1.APIRule{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				"gateway.kyma-project.io/original-version": "v2alpha1",
			},
		},
	}
	if !isAPIRuleV2(apiRule) {
		t.Fatal("expected isAPIRuleV2 to return true for non-v1beta1 annotation")
	}
}

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
