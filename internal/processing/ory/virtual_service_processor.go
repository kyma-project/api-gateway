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
func NewVirtualServiceProcessor(config processing.ReconciliationConfig) processors.VirtualServiceProcessor {
	return processors.VirtualServiceProcessor{
		Creator: virtualServiceCreator{
			oathkeeperSvc:     config.OathkeeperSvc,
			oathkeeperSvcPort: config.OathkeeperSvcPort,
			corsConfig:        config.CorsConfig,
			additionalLabels:  config.AdditionalLabels,
			defaultDomainName: config.DefaultDomainName,
		},
	}
}

type virtualServiceCreator struct {
	oathkeeperSvc     string
	oathkeeperSvcPort uint32
	corsConfig        *processing.CorsConfig
	defaultDomainName string
	additionalLabels  map[string]string
}

// Create returns the Virtual Service using the configuration of the APIRule.
func (r virtualServiceCreator) Create(api *gatewayv1beta1.APIRule) (*networkingv1beta1.VirtualService, error) {
	virtualServiceNamePrefix := fmt.Sprintf("%s-", api.ObjectMeta.Name)

	vsSpecBuilder := builders.VirtualServiceSpec()
	vsSpecBuilder.Host(default_domain.GetHostWithDomain(*api.Spec.Host, r.defaultDomainName))
	vsSpecBuilder.Gateway(*api.Spec.Gateway)
	filteredRules := processing.FilterDuplicatePaths(api.Spec.Rules)

	for _, rule := range filteredRules {
		httpRouteBuilder := builders.HTTPRoute()
		host, port := r.oathkeeperSvc, r.oathkeeperSvcPort
		serviceNamespace := helpers.FindServiceNamespace(api, &rule)

		if !processing.IsSecured(rule) {
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

		corsPolicy := builders.CorsPolicy().
			AllowOrigins(r.corsConfig.AllowOrigins...).
			AllowHeaders(r.corsConfig.AllowHeaders...)
		if len(rule.Methods) >= 1 {
			corsPolicy.AllowMethods(rule.Methods...)
		} else {
			corsPolicy.AllowMethods(r.corsConfig.AllowMethods...)
		}

		httpRouteBuilder.Route(builders.RouteDestination().Host(host).Port(port))
		httpRouteBuilder.Match(builders.MatchRequest().Uri().Regex(rule.Path))
		httpRouteBuilder.CorsPolicy(corsPolicy)
		httpRouteBuilder.Headers(builders.NewHttpRouteHeadersBuilder().
			SetHostHeader(default_domain.GetHostWithDomain(*api.Spec.Host, r.defaultDomainName)).Get())
		httpRouteBuilder.Timeout(processors.GetVirtualServiceHttpTimeout(api.Spec, rule))
		vsSpecBuilder.HTTP(httpRouteBuilder)

	}

	vsBuilder := builders.VirtualService().
		GenerateName(virtualServiceNamePrefix).
		Namespace(api.ObjectMeta.Namespace).
		Label(processing.OwnerLabel, fmt.Sprintf("%s.%s", api.ObjectMeta.Name, api.ObjectMeta.Namespace))

	for k, v := range r.additionalLabels {
		vsBuilder.Label(k, v)
	}

	vsBuilder.Spec(vsSpecBuilder)

	return vsBuilder.Get(), nil
}
