package externalgateway

import (
	"context"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	externalv1alpha1 "github.com/kyma-project/api-gateway/apis/gateway/external/v1alpha1"
)

func newEG(name, ns, domain string) *externalv1alpha1.ExternalGateway {
	return &externalv1alpha1.ExternalGateway{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns},
		Spec:       externalv1alpha1.ExternalGatewaySpec{ExternalDomain: domain},
	}
}

func TestCheckExternalDomainUnique_NoConflict(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = externalv1alpha1.AddToScheme(scheme)
	existing := newEG("other", "ns", "other.example.com")
	current := newEG("me", "ns", "api.example.com")
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(existing, current).Build()

	if err := CheckExternalDomainUnique(context.Background(), c, current); err != nil {
		t.Fatalf("expected no conflict, got %v", err)
	}
}

func TestCheckExternalDomainUnique_Conflict(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = externalv1alpha1.AddToScheme(scheme)
	existing := newEG("other", "other-ns", "api.example.com")
	current := newEG("me", "ns", "api.example.com")
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(existing, current).Build()

	err := CheckExternalDomainUnique(context.Background(), c, current)
	if err == nil {
		t.Fatal("expected conflict error, got nil")
	}
	reason, ok := ErrorReason(err)
	if !ok || reason != externalv1alpha1.ReasonExternalDomainConflict {
		t.Fatalf("expected reason %s, got %q ok=%v", externalv1alpha1.ReasonExternalDomainConflict, reason, ok)
	}
}

func TestCheckExternalDomainUnique_SelfIgnored(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = externalv1alpha1.AddToScheme(scheme)
	current := newEG("me", "ns", "api.example.com")
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(current).Build()

	if err := CheckExternalDomainUnique(context.Background(), c, current); err != nil {
		t.Fatalf("expected no error (self must be ignored), got %v", err)
	}
}

func TestCheckExternalDomainUnique_EmptyDomainSkipped(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = externalv1alpha1.AddToScheme(scheme)
	other := newEG("other", "ns", "")
	current := newEG("me", "ns", "")
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(other, current).Build()

	if err := CheckExternalDomainUnique(context.Background(), c, current); err != nil {
		t.Fatalf("empty externalDomain should not conflict, got %v", err)
	}
}
