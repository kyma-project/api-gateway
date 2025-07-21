package infrastructure

import (
	v2 "github.com/kyma-project/api-gateway/apis/gateway/v2"
	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	"istio.io/client-go/pkg/apis/security/v1beta1"
	"net/http"
	"sigs.k8s.io/e2e-framework/klient/conf"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sync/atomic"
	"testing"

	httphelper "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/http"
	"k8s.io/client-go/rest"
)

const KubernetesClientLogPrefix = "kube-client"

var isInitialized atomic.Bool

func ResourcesClient(t *testing.T) (*resources.Resources, error) {
	path := conf.ResolveKubeConfigFile()
	cfg := envconf.NewWithKubeConfig(path)

	r, err := resources.New(wrapTestLog(t, cfg.Client().RESTConfig()))
	if err != nil {
		t.Logf("Failed to create resources client: %v", err)
		return nil, err
	}

	if !isInitialized.Load() {
		err = v2.AddToScheme(r.GetScheme())
		if err != nil {
			t.Logf("Failed to add v2 scheme: %v", err)
			return nil, err
		}

		err = v1alpha3.AddToScheme(r.GetScheme())
		if err != nil {
			t.Logf("Failed to add v1alpha3 scheme: %v", err)
			return nil, err
		}

		err = v1beta1.AddToScheme(r.GetScheme())
		if err != nil {
			t.Logf("Failed to add v1beta1 scheme: %v", err)
			return nil, err
		}
		isInitialized.Store(true)
	}

	return r, nil
}

func wrapTestLog(t *testing.T, cfg *rest.Config) *rest.Config {
	cfg.Wrap(func(rt http.RoundTripper) http.RoundTripper {
		return httphelper.TestLogTransportWrapper(t, KubernetesClientLogPrefix, rt)
	})
	return cfg
}
