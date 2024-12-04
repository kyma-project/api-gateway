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
	ValidateAndFixConfig() error
	TestConcurrency() int
}

type TestsuiteFactory func(httpClient *helpers.RetryableHttpClient, k8sClient dynamic.Interface, rm *resource.Manager, config Config) Testsuite

func New(factory TestsuiteFactory) (Testsuite, error) {
	log.Printf("Creating test suite")
	pflag.Parse()

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

	config := GetConfig()
	if err := envconfig.Init(&config); err != nil {
		log.Fatalf("Unable to setup config: %v", err)
	}
	ctx := factory(retryingHttpClient, k8sClient, rm, config)
	log.Printf("Validating test configuration")
	err = ctx.ValidateAndFixConfig()
	if err != nil {
		log.Fatalf("tests are not configured properly: %s", err)
	}
	log.Printf("Test configuration validated")

	log.Printf("Setting up the test suite")
	err = ctx.Setup()
	log.Printf("Test suite setup finished")

	return ctx, err
}
