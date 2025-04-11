package istio

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

// Newv1beta1VirtualServiceProcessor returns a VirtualServiceProcessor with the desired state handling specific for the Istio handler.
func Newv1beta1VirtualServiceProcessor(config processing.ReconciliationConfig, api *gatewayv1beta1.APIRule) processors.VirtualServiceProcessor {
	return processors.VirtualServiceProcessor{
		ApiRule: api,
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
	virtualServiceNamePrefix := fmt.Sprintf("%s-", api.Name)

	vsSpecBuilder := builders.VirtualServiceSpec()
	vsSpecBuilder.AddHost(default_domain.GetHostWithDomain(*api.Spec.Host, r.defaultDomainName))
	vsSpecBuilder.Gateway(*api.Spec.Gateway)
	filteredRules := processing.FilterDuplicatePaths(api.Spec.Rules)

	for _, rule := range filteredRules {
		httpRouteBuilder := builders.HTTPRoute()
		serviceNamespace := helpers.FindServiceNamespace(api, &rule)
		routeDirectlyToService := false
		if !processing.IsSecuredByOathkeeper(rule) {
			routeDirectlyToService = true
		} else if processing.IsJwtSecured(rule) {
			routeDirectlyToService = true
		}

		var host string
		var port uint32

		if routeDirectlyToService {
			// Use rule level service if it exists
			if rule.Service != nil {
				host = default_domain.GetHostLocalDomain(*rule.Service.Name, serviceNamespace)
				port = *rule.Service.Port
			} else {
				// Otherwise use service defined on APIRule spec level
				host = default_domain.GetHostLocalDomain(*api.Spec.Service.Name, serviceNamespace)
				port = *api.Spec.Service.Port
			}
		} else {
			host = r.oathkeeperSvc
			port = r.oathkeeperSvcPort
		}

		httpRouteBuilder.Route(builders.RouteDestination().Host(host).Port(port))

		matchBuilder := builders.MatchRequest()
		if restrictsExposedMethods(rule) {
			matchBuilder.MethodRegEx(rule.Methods...)
		}

		if rule.Path == "/*" {
			matchBuilder.Uri().Prefix("/")
		} else {
			matchBuilder.Uri().Regex(rule.Path)
		}

		httpRouteBuilder.Match(matchBuilder)

		if api.Spec.CorsPolicy == nil {
			httpRouteBuilder.CorsPolicy(builders.CorsPolicy().
				AllowOrigins(r.corsConfig.AllowOrigins...).
				AllowMethods(r.corsConfig.AllowMethods...).
				AllowHeaders(r.corsConfig.AllowHeaders...))
		}
		httpRouteBuilder.Timeout(processors.GetVirtualServiceHttpTimeout(api.Spec, rule))

		headersBuilder := builders.NewHttpRouteHeadersBuilder().
			SetHostHeader(default_domain.GetHostWithDomain(*api.Spec.Host, r.defaultDomainName))

		if api.Spec.CorsPolicy != nil {
			httpRouteBuilder.CorsPolicy(builders.CorsPolicy().FromApiRuleCorsPolicy(*api.Spec.CorsPolicy))
			headersBuilder.RemoveUpstreamCORSPolicyHeaders()
		}

		// We need to add mutators only for JWT secured rules, since "noop" and "oauth2_introspection" access strategies
		// create access rules and therefore use ory mutators. The "allow" access strategy does not support mutators at all.
		if processing.IsJwtSecured(rule) {
			cookieMutator, err := rule.GetCookieMutator()
			if err != nil {
				return nil, err
			}
			if cookieMutator.HasCookies() {
				headersBuilder.SetRequestCookies(cookieMutator.Cookies)
			}

			headerMutator, err := rule.GetHeaderMutator()
			if err != nil {
				return nil, err
			}
			if headerMutator.HasHeaders() {
				headersBuilder.SetRequestHeaders(headerMutator.Headers)
			}
		}

		httpRouteBuilder.Headers(headersBuilder.Get())

		vsSpecBuilder.HTTP(httpRouteBuilder)

	}

	vsBuilder := builders.VirtualService().
		GenerateName(virtualServiceNamePrefix).
		Namespace(api.ObjectMeta.Namespace).
		Label(processing.OwnerLabel, fmt.Sprintf("%s.%s", api.Name, api.Namespace))

	vsBuilder.Spec(vsSpecBuilder)

	return vsBuilder.Get(), nil
}

// restrictsExposedMethods checks if the rule has only access strategies defined that restrict access to specific HTTP methods.
func restrictsExposedMethods(rule gatewayv1beta1.Rule) bool {
	return !rule.ContainsAccessStrategy(gatewayv1beta1.AccessStrategyAllow)
}
