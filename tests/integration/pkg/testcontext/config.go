package testcontext

import (
	"fmt"
	"github.com/avast/retry-go/v4"
	"github.com/vrischmann/envconfig"
	"time"
)

const (
	AnyToken                  = "any"
	AuthorizationHeaderName   = "Authorization"
	AuthorizationHeaderPrefix = "Bearer"
	OpaqueHeaderName          = "opaque-token"
)

type Config struct {
	CustomDomain           string `envconfig:"TEST_CUSTOM_DOMAIN,default=test.domain.kyma"`
	IssuerUrl              string `envconfig:"TEST_OIDC_ISSUER_URL"`
	ClientID               string `envconfig:"TEST_CLIENT_ID"`
	ClientSecret           string `envconfig:"TEST_CLIENT_SECRET"`
	ReqTimeout             uint   `envconfig:"TEST_REQUEST_TIMEOUT,default=120"`
	ReqDelay               uint   `envconfig:"TEST_REQUEST_DELAY,default=5"`
	Domain                 string `envconfig:"TEST_DOMAIN,default=local.kyma.dev"`
	GatewayName            string `envconfig:"TEST_GATEWAY_NAME,default=kyma-gateway"`
	GatewayNamespace       string `envconfig:"TEST_GATEWAY_NAMESPACE,default=kyma-system"`
	TestConcurrency        int    `envconfig:"TEST_CONCURRENCY,default=4"`
	APIGatewayImageVersion string `envconfig:"TEST_UPGRADE_IMG"`
}

func GetConfig() Config {
	var config Config
	if err := envconfig.Init(&config); err != nil {
		panic(fmt.Sprintf("Unable to setup test config: %v", err))
	}
	return config
}

func GetRetryOpts(config Config) []retry.Option {
	return []retry.Option{
		retry.Delay(time.Duration(config.ReqDelay) * time.Second),
		retry.Attempts(5),
		retry.DelayType(retry.FixedDelay),
	}
}
