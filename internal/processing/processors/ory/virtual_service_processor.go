package ory

import (
	"fmt"

	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"github.com/kyma-project/api-gateway/internal/builders"
	"github.com/kyma-project/api-gateway/internal/helpers"
	"github.com/kyma-project/api-gateway/internal/processing"
	"github.com/kyma-project/api-gateway/internal/processing/default_domain"
	"github.com/kyma-project/api-gateway/internal/processing/processors"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
)

// NewVirtualServiceProcessor returns a VirtualServiceProcessor with the desired state handling specific for the Ory handler.
func NewVirtualServiceProcessor(config processing.ReconciliationConfig, apiRule *gatewayv1beta1.APIRule) processors.VirtualServiceProcessor {
	return processors.VirtualServiceProcessor{
		ApiRule: apiRule,
		Creator: virtualServiceCreator{
			oathkeeperSvc:     config.OathkeeperSvc,
			oathkeeperSvcPort: config.OathkeeperSvcPort,
			corsConfig:        config.CorsConfig,
			defaultDomainName: config.DefaultDomainName,
		},
	}
}

type virtualServiceCreator struct {
	oathkeeperSvc     string
	oathkeeperSvcPort uint32
	corsConfig        *processing.CorsConfig
	defaultDomainName string
}

// Create returns the Virtual Service using the configuration of the APIRule.
func (r virtualServiceCreator) Create(api *gatewayv1beta1.APIRule) (*networkingv1beta1.VirtualService, error) {
	virtualServiceNamePrefix := fmt.Sprintf("%s-", api.ObjectMeta.Name)

	vsSpecBuilder := builders.VirtualServiceSpec()
	vsSpecBuilder.AddHost(default_domain.GetHostWithDomain(*api.Spec.Host, r.defaultDomainName))
	vsSpecBuilder.Gateway(*api.Spec.Gateway)
	filteredRules := processing.FilterDuplicatePaths(api.Spec.Rules)

	for _, rule := range filteredRules {
		httpRouteBuilder := builders.HTTPRoute()
		host, port := r.oathkeeperSvc, r.oathkeeperSvcPort
		serviceNamespace := helpers.FindServiceNamespace(api, &rule)

		if !processing.IsSecuredByOathkeeper(rule) {
			// Use rule level service if it exists
			if rule.Service != nil {
				host = fmt.Sprintf("%s.%s.svc.cluster.local", *rule.Service.Name, serviceNamespace)
				port = *rule.Service.Port
			} else {
				// Otherwise use service defined on APIRule spec level
				host = fmt.Sprintf("%s.%s.svc.cluster.local", *api.Spec.Service.Name, serviceNamespace)
				port = *api.Spec.Service.Port
			}
		}

		matchBuilder := builders.MatchRequest().Uri().Regex(rule.Path)

		// Only in case of the no_auth we want to restrict the methods, because it would be a breaking
		// change for other handlers and in case of allow we expose all methods.
		if rule.ContainsAccessStrategy(gatewayv1beta1.AccessStrategyNoAuth) {
			matchBuilder.MethodRegEx(rule.Methods...)
		}

		httpRouteBuilder.Match(matchBuilder)
		httpRouteBuilder.Route(builders.RouteDestination().Host(host).Port(port))
		if api.Spec.CorsPolicy == nil {
			httpRouteBuilder.CorsPolicy(builders.CorsPolicy().
				AllowOrigins(r.corsConfig.AllowOrigins...).
				AllowMethods(r.corsConfig.AllowMethods...).
				AllowHeaders(r.corsConfig.AllowHeaders...))
		}

		headersBuilder := builders.NewHttpRouteHeadersBuilder().
			SetHostHeader(default_domain.GetHostWithDomain(*api.Spec.Host, r.defaultDomainName))

		if api.Spec.CorsPolicy != nil {
			httpRouteBuilder.CorsPolicy(builders.CorsPolicy().FromApiRuleCorsPolicy(*api.Spec.CorsPolicy))
			headersBuilder.RemoveUpstreamCORSPolicyHeaders()
		}

		httpRouteBuilder.Headers(headersBuilder.Get())
		httpRouteBuilder.Timeout(processors.GetVirtualServiceHttpTimeout(api.Spec, rule))
		vsSpecBuilder.HTTP(httpRouteBuilder)

	}

	vsBuilder := builders.VirtualService().
		GenerateName(virtualServiceNamePrefix).
		Namespace(api.ObjectMeta.Namespace).
		Label(processing.OwnerLabel, fmt.Sprintf("%s.%s", api.ObjectMeta.Name, api.ObjectMeta.Namespace))

	vsBuilder.Spec(vsSpecBuilder)

	return vsBuilder.Get(), nil
}
