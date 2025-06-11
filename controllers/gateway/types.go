package gateway

import (
	"github.com/kyma-project/api-gateway/internal/environment"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/kyma-project/api-gateway/internal/helpers"
	"github.com/kyma-project/api-gateway/internal/metrics"
	"github.com/kyma-project/api-gateway/internal/processing"
	"istio.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// APIRuleReconciler reconciles a APIRule object
type APIRuleReconciler struct {
	processing.ReconciliationConfig
	client.Client
	Log                      logr.Logger
	Scheme                   *runtime.Scheme
	Config                   *helpers.Config
	ReconcilePeriod          time.Duration
	OnErrorReconcilePeriod   time.Duration
	MigrationReconcilePeriod time.Duration
	Metrics                  *metrics.ApiGatewayMetrics
	EnvironmentalConfig      *environment.Config
}

type ApiRuleReconcilerConfiguration struct {
	OathkeeperSvcAddr                                    string
	OathkeeperSvcPort                                    uint
	CorsAllowOrigins, CorsAllowMethods, CorsAllowHeaders string
	ReconciliationPeriod                                 uint
	ErrorReconciliationPeriod                            uint
	MigrationReconciliationPeriod                        uint
}

func NewApiRuleReconciler(mgr manager.Manager, config ApiRuleReconcilerConfiguration, apiGatewayMetrics *metrics.ApiGatewayMetrics, environmentalConfig *environment.Config) *APIRuleReconciler {
	return &APIRuleReconciler{
		Client: mgr.GetClient(),
		Log:    mgr.GetLogger().WithName("apirule-controller"),
		ReconciliationConfig: processing.ReconciliationConfig{
			OathkeeperSvc:     config.OathkeeperSvcAddr,
			OathkeeperSvcPort: uint32(config.OathkeeperSvcPort),
			CorsConfig: &processing.CorsConfig{
				AllowHeaders: getList(config.CorsAllowHeaders),
				AllowMethods: getList(config.CorsAllowMethods),
				AllowOrigins: getStringMatch(config.CorsAllowOrigins),
			},
		},
		Scheme:                   mgr.GetScheme(),
		Config:                   &helpers.Config{},
		ReconcilePeriod:          time.Duration(config.ReconciliationPeriod) * time.Second,
		OnErrorReconcilePeriod:   time.Duration(config.ErrorReconciliationPeriod) * time.Second,
		MigrationReconcilePeriod: time.Duration(config.MigrationReconciliationPeriod) * time.Second,
		Metrics:                  apiGatewayMetrics,
		EnvironmentalConfig:      environmentalConfig,
	}
}

func getList(raw string) []string {
	var result []string
	for _, s := range strings.Split(raw, ",") {
		trim := strings.TrimSpace(s)
		if trim != "" {
			result = append(result, trim)
		}
	}
	return result
}

func getStringMatch(raw string) []*v1beta1.StringMatch {
	var result []*v1beta1.StringMatch
	for _, s := range getList(raw) {
		matchTypePair := strings.SplitN(s, ":", 2)
		matchType := matchTypePair[0]
		value := matchTypePair[1]
		var stringMatch *v1beta1.StringMatch
		switch matchType {
		case "regex":
			stringMatch = regex(value)
		case "prefix":
			stringMatch = prefix(value)
		case "exact":
			stringMatch = exact(value)
		}
		result = append(result, stringMatch)
	}
	return result
}

func regex(val string) *v1beta1.StringMatch {
	return &v1beta1.StringMatch{
		MatchType: &v1beta1.StringMatch_Regex{Regex: val},
	}
}

func prefix(val string) *v1beta1.StringMatch {
	return &v1beta1.StringMatch{
		MatchType: &v1beta1.StringMatch_Prefix{Prefix: val},
	}
}

func exact(val string) *v1beta1.StringMatch {
	return &v1beta1.StringMatch{
		MatchType: &v1beta1.StringMatch_Exact{Exact: val},
	}
}
