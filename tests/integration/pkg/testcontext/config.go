package testcontext

import (
	"fmt"
	"github.com/avast/retry-go"
	"github.com/vrischmann/envconfig"
	"time"
)

const (
	ManifestsDirectory        = "manifests"
	globalCommonResourcesFile = "global-commons.yaml"
	ResourceSeparator         = "---"
	ExportResultVar           = "EXPORT_RESULT"
	CucumberFileName          = "cucumber-report.json"
	AnyToken                  = "any"
	AuthorizationHeaderName   = "Authorization"
	AuthorizationHeaderPrefix = "Bearer"
	OpaqueHeaderName          = "opaque-token"
	DefaultNS                 = "kyma-system"
	ConfigMapName             = "api-gateway-config"
)

type TestRunConfig struct {
	// TODO: Custom domain is only relevant in custom domain testsuite
	CustomDomain     string `envconfig:"TEST_CUSTOM_DOMAIN,default=test.domain.kyma"`
	IssuerUrl        string `envconfig:"TEST_OIDC_ISSUER_URL"`
	ClientID         string `envconfig:"TEST_CLIENT_ID"`
	ClientSecret     string `envconfig:"TEST_CLIENT_SECRET"`
	ReqTimeout       uint   `envconfig:"TEST_REQUEST_TIMEOUT,default=180"`
	ReqDelay         uint   `envconfig:"TEST_REQUEST_DELAY,default=5"`
	Domain           string `envconfig:"TEST_DOMAIN,default=local.kyma.dev"`
	GatewayName      string `envconfig:"TEST_GATEWAY_NAME,default=kyma-gateway"`
	GatewayNamespace string `envconfig:"TEST_GATEWAY_NAMESPACE,default=kyma-system"`
	TestConcurrency  int    `envconfig:"TEST_CONCURRENCY,default=1"`
}

func GetConfig() TestRunConfig {
	var config TestRunConfig
	if err := envconfig.Init(&config); err != nil {
		panic(fmt.Sprintf("Unable to setup test config: %v", err))
	}
	return config
}

func GetRetryOpts(config TestRunConfig) []retry.Option {
	return []retry.Option{
		retry.Delay(time.Duration(config.ReqDelay) * time.Second),
		retry.Attempts(config.ReqTimeout / config.ReqDelay),
		retry.DelayType(retry.FixedDelay),
	}
}
