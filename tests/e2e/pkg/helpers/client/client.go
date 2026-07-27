package client

import (
	externalv1alpha1 "github.com/kyma-project/api-gateway/apis/gateway/external/v1alpha1"
	v2 "github.com/kyma-project/api-gateway/apis/gateway/v2"
	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	"istio.io/client-go/pkg/apis/security/v1beta1"
	"k8s.io/client-go/kubernetes"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/e2e-framework/klient/conf"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sync"
	"testing"

	httphelper "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/http"
	"k8s.io/client-go/rest"
)

const KubernetesClientLogPrefix = "kube-client"

var (
	schemeOnce sync.Once
	schemeErr  error
)

func ResourcesClient(t *testing.T) (*resources.Resources, error) {
	path := conf.ResolveKubeConfigFile()
	cfg := envconf.NewWithKubeConfig(path)

	r, err := resources.New(wrapTestLog(t, cfg.Client().RESTConfig()))
	if err != nil {
		t.Logf("Failed to create resources client: %v", err)
		return nil, err
	}

	schemeOnce.Do(func() {
		if err := v2.AddToScheme(r.GetScheme()); err != nil {
			schemeErr = err
			return
		}
		if err := v1alpha3.AddToScheme(r.GetScheme()); err != nil {
			schemeErr = err
			return
		}
		if err := v1beta1.AddToScheme(r.GetScheme()); err != nil {
			schemeErr = err
			return
		}
		if err := externalv1alpha1.AddToScheme(r.GetScheme()); err != nil {
			schemeErr = err
			return
		}
	})
	if schemeErr != nil {
		t.Logf("Failed to register schemes: %v", schemeErr)
		return nil, schemeErr
	}

	return r, nil
}

func wrapTestLog(t *testing.T, cfg *rest.Config) *rest.Config {
	cfg.Wrap(func(rt http.RoundTripper) http.RoundTripper {
		return httphelper.TestLogTransportWrapper(t, KubernetesClientLogPrefix, rt)
	})
	return cfg
}

func GetClientSet(t *testing.T) (*kubernetes.Clientset, error) {
	t.Helper()
	restConfig, err := config.GetConfig()
	if err != nil {
		t.Logf("Could not create in-cluster config: err=%s", err)
		return nil, err
	}
	return kubernetes.NewForConfig(restConfig)
}
