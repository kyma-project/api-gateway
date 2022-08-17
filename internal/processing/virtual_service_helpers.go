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

func (f *Factory) generateVirtualService(api *gatewayv1beta1.APIRule) *networkingv1beta1.VirtualService {
	virtualServiceNamePrefix := fmt.Sprintf("%s-", api.ObjectMeta.Name)
	ownerRef := generateOwnerRef(api)

	vsSpecBuilder := builders.VirtualServiceSpec()
	vsSpecBuilder.Host(helpers.GetHostWithDomain(*api.Spec.Host, f.defaultDomainName))
	vsSpecBuilder.Gateway(*api.Spec.Gateway)

	for _, rule := range api.Spec.Rules {

		httpRouteBuilder := builders.HTTPRoute()
		host, port := f.oathkeeperSvc, f.oathkeeperSvcPort

		if !isSecured(rule) {
			// Use rule level service if it exists
			if rule.Service != nil {
				host = fmt.Sprintf("%s.%s.svc.cluster.local", *rule.Service.Name, api.ObjectMeta.Namespace)
				port = *rule.Service.Port
			} else {
				// Otherwise use service defined on APIRule spec level
				host = fmt.Sprintf("%s.%s.svc.cluster.local", *api.Spec.Service.Name, api.ObjectMeta.Namespace)
				port = *api.Spec.Service.Port
			}
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
