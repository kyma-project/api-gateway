package modules

import (
	"bytes"
	_ "embed"
	infrahelpers "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/infrastructure"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"testing"
	"time"

	"github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/setup"
	"sigs.k8s.io/e2e-framework/klient/decoder"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
)

//go:embed operator_v1alpha1_apigateway.yaml
var apiGatewayTemplate []byte

type ApiGatewayCROptions struct {
	Template []byte
}

func WithAPIGatewayTemplate(template string) ApiGatewayCROption {
	return func(opts *ApiGatewayCROptions) {
		opts.Template = []byte(template)
	}
}

type ApiGatewayCROption func(*ApiGatewayCROptions)

func CreateApiGatewayCR(t *testing.T, options ...ApiGatewayCROption) error {
	t.Helper()
	t.Log("Creating APIGateway custom resource")
	opts := &ApiGatewayCROptions{
		Template: apiGatewayTemplate,
	}
	for _, opt := range options {
		opt(opts)
	}

	r, err := infrahelpers.ResourcesClient(t)
	if err != nil {
		t.Logf("Failed to get resources client: %v", err)
		return err
	}

	err = v1alpha1.AddToScheme(r.GetScheme())
	if err != nil {
		t.Logf("Failed to add APIGateway v1alpha1 scheme: %v", err)
		return err
	}

	icr := &v1alpha1.APIGateway{}
	err = decoder.Decode(
		bytes.NewBuffer(opts.Template),
		icr,
	)
	if err != nil {
		t.Logf("Failed to decode APIGateway custom resource template: %v", err)
		return err
	}

	err = r.Create(t.Context(), icr)
	if err != nil {
		t.Logf("Failed to create APIGateway custom resource: %v", err)
		return err
	}

	setup.DeclareCleanup(t, func() {
		t.Log("Cleaning up APIGateway after the tests")
		err := TeardownApiGatewayCR(t, options...)
		if err != nil {
			t.Logf("Failed to clean up APIGateway custom resource: %v", err)
		} else {
			t.Log("APIGateway custom resource cleaned up successfully")
		}
	})

	err = waitForAPIGatewayCRReadiness(t, r, icr)
	if err != nil {
		t.Logf("Failed to wait for APIGateway custom resource readiness: %v", err)
		return err
	}

	t.Log("APIGateway custom resource created successfully")
	return nil
}

func waitForAPIGatewayCRReadiness(t *testing.T, r *resources.Resources, icr *v1alpha1.APIGateway) error {
	t.Helper()
	t.Log("Waiting for APIGateway custom resource to be ready")

	clock := time.Now()
	err := wait.For(conditions.New(r).ResourceMatch(icr, func(obj k8s.Object) bool {
		t.Logf("Waiting for APIGateway custom resource %s to be ready", obj.GetName())
		t.Logf("Elapsed time: %s", time.Since(clock))

		icrObj, ok := obj.(*v1alpha1.APIGateway)
		if !ok {
			return false
		}
		return icrObj.Status.State == v1alpha1.Ready
	}))
	if err != nil {
		t.Logf("Failed to wait for APIGateway custom resource to be ready: %v", err)
		return err
	}

	t.Log("APIGateway custom resource is ready")
	return nil
}

const apiGWCRDeletionTimeout = time.Minute * 2

func waitForAPIGatewayCRDeletion(t *testing.T, r *resources.Resources, icr *v1alpha1.APIGateway) error {
	t.Helper()
	t.Log("Waiting for APIGateway custom resource to be deleted")

	err := wait.For(conditions.New(r).ResourceDeleted(icr), wait.WithTimeout(apiGWCRDeletionTimeout))
	if err != nil {
		t.Logf("Failed to wait for APIGateway custom resource deletion: %v", err)
		return err
	}

	t.Log("APIGateway custom resource deleted successfully")
	return nil
}

func TeardownApiGatewayCR(t *testing.T, options ...ApiGatewayCROption) error {
	t.Helper()
	t.Log("Beginning cleanup of APIGateway custom resource")
	opts := &ApiGatewayCROptions{
		Template: apiGatewayTemplate,
	}
	for _, opt := range options {
		opt(opts)
	}

	r, err := infrahelpers.ResourcesClient(t)
	if err != nil {
		t.Logf("Failed to get resources client: %v", err)
		return err
	}

	err = v1alpha1.AddToScheme(r.GetScheme())
	if err != nil {
		t.Logf("Failed to add APIGateway v1alpha1 scheme: %v", err)
		return err
	}

	icr := &v1alpha1.APIGateway{}
	t.Log("Deleting APIGateway custom resource")
	err = decoder.Decode(
		bytes.NewBuffer(opts.Template),
		icr,
	)
	if err != nil {
		t.Logf("Failed to decode APIGateway custom resource template: %v", err)
		return err
	}

	err = r.Delete(setup.GetCleanupContext(), icr)
	if err != nil {
		t.Logf("Failed to delete APIGateway custom resource: %v", err)
		if k8serrors.IsNotFound(err) {
			t.Log("APIGateway custom resource not found, nothing to delete")
			return nil
		}
		return err
	}

	return waitForAPIGatewayCRDeletion(t, r, icr)
}
