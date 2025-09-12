package infrastructure

import (
	"testing"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/api-gateway/tests/e2e/pkg/setup"
)

type NamespaceOptions struct {
	Labels              map[string]string
	IgnoreAlreadyExists bool
}

func WithLabels(labels map[string]string) NamespaceOption {
	return func(opts *NamespaceOptions) {
		opts.Labels = labels
	}
}

func IgnoreAlreadyExists() NamespaceOption {
	return func(opts *NamespaceOptions) {
		opts.IgnoreAlreadyExists = true
	}
}

func WithSidecarInjectionEnabled() NamespaceOption {
	return func(opts *NamespaceOptions) {
		if opts.Labels == nil {
			opts.Labels = make(map[string]string)
		}
		opts.Labels["istio-injection"] = "enabled"
	}
}

type NamespaceOption func(*NamespaceOptions)

func CreateNamespace(t *testing.T, name string, options ...NamespaceOption) error {
	t.Helper()
	opts := &NamespaceOptions{
		Labels: nil,
	}
	for _, opt := range options {
		opt(opts)
	}

	r, err := ResourcesClient(t)
	if err != nil {
		t.Logf("Failed to get resources client: %v", err)
		return err
	}

	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: name}}
	if opts.Labels != nil {
		ns.Labels = opts.Labels
	}

	t.Log("Creating namespace: ", name)

	err = r.Create(t.Context(), ns)
	if err != nil {
		if opts.IgnoreAlreadyExists && k8serrors.IsAlreadyExists(err) {
			t.Logf("Namespace %s already exists, ignoring error as per options", name)
			return nil
		}
		return err
	}

	setup.DeclareCleanup(t, func() {
		t.Log("Deleting namespace: ", name)
		err := DeleteNamespace(t, name)
		if err != nil {
			t.Logf("Failed to delete namespace %s: %v", name, err)
		}
	})
	return nil
}

func DeleteNamespace(t *testing.T, name string) error {
	t.Helper()
	r, err := ResourcesClient(t)
	if err != nil {
		t.Logf("Failed to get resources client: %v", err)
		return err
	}
	return r.Delete(setup.GetCleanupContext(), &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: name}})
}
