package httpbin

import (
	"bytes"
	_ "embed"
	"fmt"
	infrahelpers "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/infrastructure"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/setup"
	"sigs.k8s.io/e2e-framework/klient/decoder"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"testing"
)

//go:embed manifest.yaml
var manifest []byte

type Options struct {
	Namespace string
}

func WithNamespace(ns string) Option {
	return func(o *Options) {
		o.Namespace = ns
	}
}

type Option func(*Options)

func DeployHttpbin(t *testing.T, options ...Option) (svcName string, svcPort int, err error) {
	t.Helper()
	opts := &Options{
		Namespace: "default",
	}
	for _, opt := range options {
		opt(opts)
	}

	r, err := infrahelpers.ResourcesClient(t)
	if err != nil {
		t.Logf("Failed to get resources client: %v", err)
		return "", 0, fmt.Errorf("failed to get resources client: %w", err)
	}

	err = infrahelpers.CreateNamespace(t, opts.Namespace, infrahelpers.IgnoreAlreadyExists(), infrahelpers.WithSidecarInjectionEnabled())
	if err != nil {
		t.Logf("Failed to create namespace: %v", err)
		return "", 0, fmt.Errorf("failed to create namespace %s: %w", opts.Namespace, err)
	}

	// No further cleanup is needed as the namespace will be deleted
	// as part of Namespace cleanup.
	// setup.DeclareCleanup(t, func() {})

	return "httpbin", 8000, start(t, r, opts)
}

func start(t *testing.T, r *resources.Resources, options *Options) error {
	err := decoder.DecodeEach(
		t.Context(),
		bytes.NewBuffer(manifest),
		decoder.CreateHandler(r),
		decoder.MutateNamespace(options.Namespace),
	)
	if err != nil {
		t.Logf("Failed to deploy mock: %v", err)
		return err
	}

	setup.DeclareCleanup(t, func() {
		t.Logf("Cleaning up httpbin in namespace %s", options.Namespace)
		err := decoder.DecodeEach(
			setup.GetCleanupContext(),
			bytes.NewBuffer(manifest),
			decoder.DeleteHandler(r),
			decoder.MutateNamespace(options.Namespace),
		)
		if err != nil {
			t.Logf("Failed to clean up httpbin: %v", err)
		} else {
			t.Logf("Successfully cleaned up httpbin in namespace %s", options.Namespace)
		}
	})

	return wait.For(conditions.New(r).DeploymentAvailable("httpbin", options.Namespace))
}
