package testcontext

import (
	"crypto/tls"
	"log"
	"net/http"
	"os"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"time"

	"github.com/cucumber/godog"

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
	FeaturePath() []string
	InitScenarios(ctx *godog.ScenarioContext)
	Setup() error
	TearDown()
	ResourceManager() *resource.Manager
	K8sClient() dynamic.Interface
	BeforeSuiteHooks() []func() error
	AfterSuiteHooks() []func() error
}

type TestsuiteFactory func(httpClient *helpers.RetryableHttpClient, k8sClient dynamic.Interface, rm *resource.Manager, config Config) Testsuite

func New(config Config, factory TestsuiteFactory) (Testsuite, error) {
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

	retryingHttpClient := helpers.NewClientWithRetry(httpClient, GetRetryOpts())

	k8sClient, err := client.GetDynamicClient()
	if err != nil {
		return nil, err
	}
	logf.SetLogger(zap.New(zap.WriteTo(os.Stdout), zap.UseDevMode(true)))

	rm := resource.NewManager(GetRetryOpts())

	ctx := factory(retryingHttpClient, k8sClient, rm, config)
	err = ctx.Setup()
	return ctx, err
}
