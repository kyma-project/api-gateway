package gateway

import (
	"fmt"
	"github.com/go-logr/logr"
	"github.com/kyma-project/api-gateway/internal/helpers"
	"github.com/kyma-project/api-gateway/internal/processing"
	"github.com/kyma-project/api-gateway/internal/validation"
	"github.com/pkg/errors"
	"istio.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"strings"
	"time"
)

// APIRuleReconciler reconciles a APIRule object
type APIRuleReconciler struct {
	processing.ReconciliationConfig
	client.Client
	Log                    logr.Logger
	Scheme                 *runtime.Scheme
	Config                 *helpers.Config
	ReconcilePeriod        time.Duration
	OnErrorReconcilePeriod time.Duration
}

type ApiRuleReconcilerConfiguration struct {
	OathkeeperSvcAddr                                    string
	OathkeeperSvcPort                                    uint
	AllowListedDomains                                   string
	BlockListedServices                                  string
	DomainName                                           string
	CorsAllowOrigins, CorsAllowMethods, CorsAllowHeaders string
	AdditionalLabels                                     map[string]string
	ReconciliationPeriod                                 uint
	ErrorReconciliationPeriod                            uint
}

func NewApiRuleReconciler(mgr manager.Manager, config ApiRuleReconcilerConfiguration) (*APIRuleReconciler, error) {

	const blockListedSubdomains string = "api"

	if config.AllowListedDomains != "" {
		for _, domain := range getList(config.AllowListedDomains) {
			if !validation.ValidateDomainName(domain) {
				ctrl.Log.Error(fmt.Errorf("invalid domain in domain-allowlist"), "unable to create controller", "controller", "Api")
				os.Exit(1)
			}
		}
	}

	serviceBlockList, err := getNamespaceServiceMap(config.BlockListedServices)
	if err != nil {
		return nil, err
	}

	hostBlockList, err := getHostBlockListFrom(blockListedSubdomains, config.DomainName)
	if err != nil {
		return nil, err
	}

	return &APIRuleReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("Api"),
		ReconciliationConfig: processing.ReconciliationConfig{
			OathkeeperSvc:     config.OathkeeperSvcAddr,
			OathkeeperSvcPort: uint32(config.OathkeeperSvcPort),
			CorsConfig: &processing.CorsConfig{
				AllowHeaders: getList(config.CorsAllowHeaders),
				AllowMethods: getList(config.CorsAllowMethods),
				AllowOrigins: getStringMatch(config.CorsAllowOrigins),
			},
			AdditionalLabels:  config.AdditionalLabels,
			DefaultDomainName: config.DomainName,
			ServiceBlockList:  serviceBlockList,
			DomainAllowList:   getList(config.AllowListedDomains),
			HostBlockList:     hostBlockList,
		},
		Scheme:                 mgr.GetScheme(),
		Config:                 &helpers.Config{},
		ReconcilePeriod:        time.Duration(config.ReconciliationPeriod) * time.Second,
		OnErrorReconcilePeriod: time.Duration(config.ErrorReconciliationPeriod) * time.Second,
	}, nil
}

func getHostBlockListFrom(blockListedSubdomains string, domainName string) ([]string, error) {
	var result []string
	for _, subdomain := range getList(blockListedSubdomains) {
		if !validation.ValidateSubdomainName(subdomain) {
			return nil, errors.Errorf("invalid subdomain in subdomain-blocklist")
		}
		blockedHost := strings.Join([]string{subdomain, domainName}, ".")
		result = append(result, blockedHost)
	}
	return result, nil
}

func getNamespaceServiceMap(raw string) (map[string][]string, error) {
	result := make(map[string][]string)
	for _, s := range getList(raw) {
		if !validation.ValidateServiceName(s) {
			return nil, errors.Errorf("invalid service in service-blocklist")
		}
		namespacedService := strings.Split(s, ".")
		namespace := namespacedService[1]
		service := namespacedService[0]
		result[namespace] = append(result[namespace], service)
	}
	return result, nil
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
		switch {
		case matchType == "regex":
			stringMatch = regex(value)
		case matchType == "prefix":
			stringMatch = prefix(value)
		case matchType == "exact":
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
