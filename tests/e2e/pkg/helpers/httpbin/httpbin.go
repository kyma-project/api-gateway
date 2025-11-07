package httpbin

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/client"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/setup"
	"github.com/pkg/errors"
	"sigs.k8s.io/e2e-framework/klient/decoder"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
)

//go:embed manifest.yaml
var manifest []byte

//go:embed manifest_second.yaml
var manifestSecondHttpbin []byte

func DeployHttpbin(t *testing.T, namespace string) (svcName string, svcPort int, err error) {
	t.Helper()

	r, err := client.ResourcesClient(t)
	if err != nil {
		t.Logf("Failed to get resources client: %v", err)
		return "", 0, fmt.Errorf("failed to get resources client: %w", err)
	}

	return "httpbin", 8000, start(t, r, manifest, namespace)
}

func DeploySecondHttpbin(t *testing.T, namespace string) (svcName string, svcPort int, err error) {
	t.Helper()

	r, err := client.ResourcesClient(t)
	if err != nil {
		t.Logf("Failed to get resources client: %v", err)
		return "", 0, fmt.Errorf("failed to get resources client: %w", err)
	}

	return "httpbin-2", 8000, start(t, r, manifestSecondHttpbin, namespace)
}

func start(t *testing.T, r *resources.Resources, manifest []byte, namespace string) error {
	err := decoder.DecodeEach(
		t.Context(),
		bytes.NewBuffer(manifest),
		decoder.CreateHandler(r),
		decoder.MutateNamespace(namespace),
	)
	if err != nil {
		t.Logf("Failed to deploy mock: %v", err)
		return err
	}

	setup.DeclareCleanup(t, func() {
		t.Logf("Cleaning up httpbin in namespace %s", namespace)
		err := decoder.DecodeEach(
			setup.GetCleanupContext(),
			bytes.NewBuffer(manifest),
			decoder.DeleteHandler(r),
			decoder.MutateNamespace(namespace),
		)
		if err != nil {
			t.Logf("Failed to clean up httpbin: %v", err)
		} else {
			t.Logf("Successfully cleaned up httpbin in namespace %s", namespace)
		}
	})

	return wait.For(conditions.New(r).DeploymentAvailable("httpbin", namespace))
}

type HttpBinBodyWithHeaders map[string][]string

// TODO: WHAT IS THAT? TO REMOVE
func GetHttpbinBodyWithHeadersFromResponse(response *http.Response) (*HttpBinBodyWithHeaders, error) {
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, errors.Errorf("failed to read response body: %s", err.Error())
	}

	httpbinBodyWIthHeaders := map[string]interface{}{}
	err = json.Unmarshal(responseBody, &httpbinBodyWIthHeaders)
	if err != nil {
		return nil, errors.Errorf("failed to unmarshal response body: %s", err.Error())
	}
	return httpbinBodyWIthHeaders["headers"].(*HttpBinBodyWithHeaders), nil
}
