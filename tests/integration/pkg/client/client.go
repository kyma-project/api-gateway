package client

import (
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"os"

	ratelimit "github.com/kyma-project/api-gateway/apis/gateway/ratelimit/v1alpha1"
	v2 "github.com/kyma-project/api-gateway/apis/gateway/v2"

	"sigs.k8s.io/controller-runtime/pkg/client"

	oryv1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	"k8s.io/client-go/discovery"
	memory "k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"

	agopv1alpha1 "github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
)

const kubeconfigEnvName = "KUBECONFIG"

func loadKubeConfigOrDie() (*rest.Config, error) {
	if kubeconfig, ok := os.LookupEnv(kubeconfigEnvName); ok {
		cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, err
		}
		return cfg, nil
	}

	if _, err := os.Stat(clientcmd.RecommendedHomeFile); os.IsNotExist(err) {
		cfg, err := rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
		return cfg, nil
	}

	cfg, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func GetDynamicClient() (dynamic.Interface, error) {
	config, err := loadKubeConfigOrDie()
	if err != nil {
		return nil, err
	}

	client, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func GetDiscoveryMapper() (*restmapper.DeferredDiscoveryRESTMapper, error) {
	config, err := loadKubeConfigOrDie()
	if err != nil {
		return nil, err
	}
	dc, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, err
	}
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))

	return mapper, nil
}

func GetK8sClient() client.Client {
	config, err := loadKubeConfigOrDie()
	if err != nil {
		panic(err)
	}

	c, err := client.New(config, client.Options{})
	if err != nil {
		panic(err)
	}

	err = apiextensionsv1.AddToScheme(c.Scheme())
	if err != nil {
		panic(err)
	}
	err = agopv1alpha1.AddToScheme(c.Scheme())
	if err != nil {
		panic(err)
	}
	err = v1beta1.AddToScheme(c.Scheme())
	if err != nil {
		panic(err)
	}
	err = v2alpha1.AddToScheme(c.Scheme())
	if err != nil {
		panic(err)
	}
	err = v2.AddToScheme(c.Scheme())
	if err != nil {
		panic(err)
	}
	err = networkingv1beta1.AddToScheme(c.Scheme())
	if err != nil {
		panic(err)
	}
	err = oryv1alpha1.AddToScheme(c.Scheme())
	if err != nil {
		panic(err)
	}
	err = ratelimit.AddToScheme(c.Scheme())
	if err != nil {
		panic(err)
	}

	return c
}

// GetK8SConfig returns a Kubernetes client configuration.
func GetK8SConfig() (*rest.Config, error) {
	config, err := loadKubeConfigOrDie()
	if err != nil {
		return nil, err
	}

	return config, nil
}
