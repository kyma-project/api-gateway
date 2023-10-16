package client

import (
	"os"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	"k8s.io/client-go/discovery"
	memory "k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"

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

	err = agopv1alpha1.AddToScheme(c.Scheme())
	if err != nil {
		panic(err)
	}
	err = v1beta1.AddToScheme(c.Scheme())
	if err != nil {
		panic(err)
	}
	err = networkingv1beta1.AddToScheme(c.Scheme())
	if err != nil {
		panic(err)
	}

	return c
}
