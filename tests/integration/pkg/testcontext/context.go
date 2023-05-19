package testcontext

import (
	"context"
	"crypto/tls"
	"fmt"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"log"
	"net/http"
	"time"

	"github.com/kyma-project/api-gateway/tests/integration/pkg/client"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/helpers"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/manifestprocessor"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/resource"
	"github.com/spf13/pflag"
	"github.com/vrischmann/envconfig"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"

	"k8s.io/client-go/dynamic"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

type Context struct {
	Name            string
	HttpClient      *helpers.RetryableHttpClient
	K8sClient       dynamic.Interface
	ResourceManager *resource.Manager
	Config          Config
	CommonResources commonResources
}

func New(name string, config Config) Context {
	pflag.Parse()

	if err := envconfig.Init(&config); err != nil {
		log.Fatalf("Unable to setup config: %v", err)
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		Timeout: time.Second * 10,
	}

	retryingHttpClient := helpers.NewClientWithRetry(httpClient, GetRetryOpts(config))

	k8sClient, err := client.GetDynamicClient()
	if err != nil {
		panic(err)
	}

	rm := resource.NewManager(GetRetryOpts(config))

	t := Context{
		Name:            name,
		HttpClient:      retryingHttpClient,
		K8sClient:       k8sClient,
		Config:          config,
		ResourceManager: rm,
	}

	t.setupCommonResources(config)
	return t
}

type commonResources struct {
	Oauth2Cfg       *clientcredentials.Config
	Namespace       string
	secondNamespace string
}

func (t *Context) setupCommonResources(config Config) {
	namespace := fmt.Sprintf("%s-%s", t.Name, helpers.GenerateRandomString(6))
	secondNamespace := fmt.Sprintf("%s-2", namespace)
	log.Printf("Using namespace: %s\n", namespace)

	oauth2Cfg := &clientcredentials.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		TokenURL:     fmt.Sprintf("%s/oauth2/token", config.IssuerUrl),
		AuthStyle:    oauth2.AuthStyleInHeader,
	}

	// create common resources for all scenarios
	globalCommonResources, err := manifestprocessor.ParseFromFileWithTemplate("global-commons.yaml", "manifests", struct {
		Namespace string
	}{
		Namespace: namespace,
	})
	if err != nil {
		log.Fatal(err)
	}

	// delete test namespace if the previous test namespace persists
	nsResourceSchema, ns, name := t.ResourceManager.GetResourceSchemaAndNamespace(globalCommonResources[0])
	log.Printf("Delete test namespace, if exists: %s\n", name)
	err = t.ResourceManager.DeleteResource(t.K8sClient, nsResourceSchema, ns, name)
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(time.Duration(config.ReqDelay) * time.Second)

	log.Printf("Creating common tests resources")
	_, err = t.ResourceManager.CreateResources(t.K8sClient, globalCommonResources...)
	if err != nil {
		log.Fatal(err)
	}

	t.CommonResources = commonResources{
		Oauth2Cfg:       oauth2Cfg,
		Namespace:       namespace,
		secondNamespace: secondNamespace,
	}
}

func (t *Context) TearDownCommonResources() {
	res := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "namespaces"}
	err := t.K8sClient.Resource(res).Delete(context.Background(), t.CommonResources.Namespace, v1.DeleteOptions{})
	if err != nil {
		log.Print(err.Error())
	}

	err = t.K8sClient.Resource(res).Delete(context.Background(), t.CommonResources.secondNamespace, v1.DeleteOptions{})
	if err != nil {
		log.Print(err.Error())
	}
}
