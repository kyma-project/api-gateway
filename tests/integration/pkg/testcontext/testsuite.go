package testcontext

import (
	"crypto/tls"
	"log"
	"net/http"
	"time"

	"github.com/kyma-project/api-gateway/tests/integration/pkg/client"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/helpers"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/resource"
	"github.com/spf13/pflag"
	"github.com/vrischmann/envconfig"
	"k8s.io/client-go/dynamic"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

type Testsuite interface {
	Name() string
	Setup()
	TearDown()
	ResourceManager() *resource.Manager
	K8sClient() dynamic.Interface
}

type TestsuiteFactory func(httpClient *helpers.RetryableHttpClient, k8sClient dynamic.Interface, rm *resource.Manager, config Config) Testsuite

func New(config Config, factory TestsuiteFactory) Testsuite {
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

	ctx := factory(retryingHttpClient, k8sClient, rm, config)
	ctx.Setup()
	return ctx
}
