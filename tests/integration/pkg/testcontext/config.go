package testcontext

import (
	"fmt"
	"log"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/vrischmann/envconfig"
)

const (
	AnyToken                  = "any"
	AuthorizationHeaderName   = "Authorization"
	AuthorizationHeaderPrefix = "Bearer"
	OpaqueHeaderName          = "opaque-token"
)

type Config struct {
	CustomDomain              string `envconfig:"TEST_CUSTOM_DOMAIN,optional"`
	OIDCConfigUrl             string `envconfig:"TEST_OIDC_CONFIG_URL,optional"`
	IssuerUrl                 string `envconfig:"-"`
	ClientID                  string `envconfig:"TEST_CLIENT_ID,optional"`
	ClientSecret              string `envconfig:"TEST_CLIENT_SECRET,optional"`
	ReqAttempts               uint   `envconfig:"TEST_REQUEST_ATTEMPTS,default=60"`
	ReqDelay                  uint   `envconfig:"TEST_REQUEST_DELAY,default=5"`
	Domain                    string `envconfig:"TEST_DOMAIN,default=local.kyma.dev"`
	GatewayName               string `envconfig:"TEST_GATEWAY_NAME,default=kyma-gateway"`
	GatewayNamespace          string `envconfig:"TEST_GATEWAY_NAMESPACE,default=kyma-system"`
	TestConcurrency           int    `envconfig:"TEST_CONCURRENCY,default=4"`
	IstioNamespace            string `envconfig:"TEST_ISTIO_NAMESPACE,default=istio-system"`
	IsGardener                bool   `envconfig:"IS_GARDENER,default=false"`
	GCPServiceAccountJsonPath string `envconfig:"TEST_SA_ACCESS_KEY_PATH,optional"`
	DebugLogging              bool   `envconfig:"TEST_DEBUG_LOGGING,default=false"`
}

var (
	retryOpts []retry.Option
)

func GetConfig() Config {
	var config Config
	if err := envconfig.Init(&config); err != nil {
		panic(fmt.Sprintf("Unable to setup test config: %v", err))
	}
	return config
}

func GetRetryOpts() []retry.Option {
	if retryOpts == nil {
		var config = GetConfig()

		retryOpts = []retry.Option{
			retry.Delay(time.Duration(config.ReqDelay) * time.Second),
			retry.Attempts(config.ReqAttempts),
			retry.DelayType(retry.FixedDelay),
		}

		if config.DebugLogging {
			retryOpts = append(retryOpts, retry.OnRetry(func(n uint, err error) { log.Printf("Trial #%d failed, error: %s\n", n, err) }))
		}
	}

	return retryOpts
}
