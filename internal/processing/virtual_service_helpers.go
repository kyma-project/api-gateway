package processing

import (
	"fmt"

	"github.com/kyma-incubator/api-gateway/internal/helpers"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"

	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	"github.com/kyma-incubator/api-gateway/internal/builders"
)

func (f *Factory) updateVirtualService(existing, required *networkingv1beta1.VirtualService) {
	existing.Spec = *required.Spec.DeepCopy()
}

func (f *Factory) generateVirtualService(api *gatewayv1beta1.APIRule, config *helpers.Config) *networkingv1beta1.VirtualService {
	virtualServiceNamePrefix := fmt.Sprintf("%s-", api.ObjectMeta.Name)
	ownerRef := generateOwnerRef(api)

	vsSpecBuilder := builders.VirtualServiceSpec()
	vsSpecBuilder.Host(helpers.GetHostWithDomain(*api.Spec.Host, f.defaultDomainName))
	vsSpecBuilder.Gateway(*api.Spec.Gateway)
	filteredRules := filterDuplicatePaths(api.Spec.Rules)

	for _, rule := range filteredRules {
		httpRouteBuilder := builders.HTTPRoute()
		serviceNamespace := helpers.FindServiceNamespace(api, &rule)

		routeDirectlyToService := false

		if !isSecured(rule) {
			routeDirectlyToService = true
		} else if isJwtSecured(rule) && config.JWTHandler == helpers.JWT_HANDLER_ISTIO {
			routeDirectlyToService = true
		}

		var host string
		var port uint32

		if routeDirectlyToService {
			// Use rule level service if it exists
			if rule.Service != nil {
				host = helpers.GetHostLocalDomain(*rule.Service.Name, *serviceNamespace)
				port = *rule.Service.Port
			} else {
				// Otherwise use service defined on APIRule spec level
				host = helpers.GetHostLocalDomain(*api.Spec.Service.Name, *serviceNamespace)
				port = *api.Spec.Service.Port
			}
		} else {
			host = f.oathkeeperSvc
			port = f.oathkeeperSvcPort
		}

		httpRouteBuilder.Route(builders.RouteDestination().Host(host).Port(port))
		httpRouteBuilder.Match(builders.MatchRequest().Uri().Regex(rule.Path))
		httpRouteBuilder.CorsPolicy(builders.CorsPolicy().
			AllowOrigins(f.corsConfig.AllowOrigins...).
			AllowMethods(f.corsConfig.AllowMethods...).
			AllowHeaders(f.corsConfig.AllowHeaders...))
		httpRouteBuilder.Headers(builders.Headers().
			SetHostHeader(helpers.GetHostWithDomain(*api.Spec.Host, f.defaultDomainName)))
		vsSpecBuilder.HTTP(httpRouteBuilder)
	}

	vsBuilder := builders.VirtualService().
		GenerateName(virtualServiceNamePrefix).
		Namespace(api.ObjectMeta.Namespace).
		Owner(builders.OwnerReference().From(&ownerRef)).
		Label(OwnerLabel, fmt.Sprintf("%s.%s", api.ObjectMeta.Name, api.ObjectMeta.Namespace)).
		Label(OwnerLabelv1alpha1, fmt.Sprintf("%s.%s", api.ObjectMeta.Name, api.ObjectMeta.Namespace))

	for k, v := range f.additionalLabels {
		vsBuilder.Label(k, v)
	}

	vsBuilder.Spec(vsSpecBuilder)

	return vsBuilder.Get()
}
