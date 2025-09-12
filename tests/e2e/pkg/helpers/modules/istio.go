package modules

import (
	"bytes"
	_ "embed"
	"testing"
	"time"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/e2e-framework/klient/decoder"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"

	infrahelpers "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/infrastructure"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/setup"
)

//go:embed operator_v1alpha2_istio_ext_authorizers.yaml
var IstioExtAuthorizersTemplate string

//go:embed operator_v1alpha2_istio_default.yaml
var IstioDefaultTemplate string

type IstioCROptions struct {
	Template []byte
}

type IstioCROption func(options *IstioCROptions)

func WithIstioOperatorTemplate(template string) IstioCROption {
	return func(opts *IstioCROptions) {
		opts.Template = []byte(template)
	}
}

func CreateIstioOperatorCR(t *testing.T, options ...IstioCROption) error {
	t.Helper()
	t.Log("Creating Istio custom resource")
	opts := &IstioCROptions{
		Template: []byte(IstioDefaultTemplate),
	}
	for _, opt := range options {
		opt(opts)
	}

	r, err := infrahelpers.ResourcesClient(t)
	if err != nil {
		t.Logf("Failed to get resources client: %v", err)
		return err
	}

	icr := &unstructured.Unstructured{}
	err = decoder.Decode(
		bytes.NewBuffer(opts.Template),
		icr,
	)
	if err != nil {
		t.Logf("Failed to decode Istio custom resource template: %v", err)
		return err
	}

	err = r.Create(t.Context(), icr)
	if err != nil {
		t.Logf("Failed to create Istio custom resource: %v", err)
		return err
	}

	setup.DeclareCleanup(t, func() {
		t.Log("Cleaning up Istio after the tests")
		err := TeardownIstioCR(t, options...)
		if err != nil {
			t.Logf("Failed to clean up APIGateway custom resource: %v", err)
		} else {
			t.Log("APIGateway custom resource cleaned up successfully")
		}
	})

	return nil
}

func TeardownIstioCR(t *testing.T, options ...IstioCROption) error {
	t.Helper()
	t.Log("Beginning cleanup of Istio custom resource")
	opts := &IstioCROptions{
		Template: []byte(IstioDefaultTemplate),
	}
	for _, opt := range options {
		opt(opts)
	}

	r, err := infrahelpers.ResourcesClient(t)
	if err != nil {
		t.Logf("Failed to get resources client: %v", err)
		return err
	}

	icr := &unstructured.Unstructured{}
	t.Log("Deleting Istio custom resource")
	err = decoder.Decode(
		bytes.NewBuffer(opts.Template),
		icr,
	)
	if err != nil {
		t.Logf("Failed to decode Istio custom resource template: %v", err)
		return err
	}

	err = r.Delete(setup.GetCleanupContext(), icr)
	if err != nil {
		t.Logf("Failed to delete Istio custom resource: %v", err)
		if k8serrors.IsNotFound(err) {
			t.Log("Istio custom resource not found, nothing to delete")
			return nil
		}
		return err
	}

	return waitForIstioCRDeletion(t, r, icr)
}

var istioCRDeletionTimeout = 2 * time.Minute

func waitForIstioCRDeletion(t *testing.T, r *resources.Resources, istioCR *unstructured.Unstructured) error {
	t.Helper()
	t.Log("Waiting for Istio custom resource to be deleted")

	err := wait.For(conditions.New(r).ResourceDeleted(istioCR), wait.WithTimeout(istioCRDeletionTimeout))
	if err != nil {
		t.Logf("Failed to wait for Istio custom resource deletion: %v", err)
		return err
	}

	t.Log("Istio custom resource deleted successfully")
	return nil
}
