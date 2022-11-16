package processor

import (
	"context"
	"fmt"
	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	"github.com/kyma-incubator/api-gateway/internal/builders"
	"github.com/kyma-incubator/api-gateway/internal/helpers"
	"github.com/kyma-incubator/api-gateway/internal/processing"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type VirtualService struct {
	client            client.Client
	ctx               context.Context
	oathkeeperSvc     string
	oathkeeperSvcPort uint32
	corsConfig        *processing.CorsConfig
	additionalLabels  map[string]string
	defaultDomainName string
}

func NewVirtualService(config processing.ReconciliationConfig) VirtualService {
	return VirtualService{
		client:            config.Client,
		ctx:               config.Ctx,
		oathkeeperSvc:     config.OathkeeperSvc,
		oathkeeperSvcPort: config.OathkeeperSvcPort,
		corsConfig:        config.CorsConfig,
		additionalLabels:  config.AdditionalLabels,
		defaultDomainName: config.DefaultDomainName,
	}
}

func (p VirtualService) EvaluateReconciliation(apiRule *gatewayv1beta1.APIRule) ([]*processing.ObjectChange, gatewayv1beta1.StatusCode, error) {
	desired := p.getDesiredState(apiRule)
	actual, err := p.getActualState(p.ctx, apiRule)
	if err != nil {
		return make([]*processing.ObjectChange, 0), gatewayv1beta1.StatusSkipped, err
	}

	c := p.getObjectChanges(desired, actual)

	return []*processing.ObjectChange{c}, gatewayv1beta1.StatusOK, nil
}

func (p VirtualService) getDesiredState(api *gatewayv1beta1.APIRule) *networkingv1beta1.VirtualService {
	return p.generateVirtualService(api)
}

func (p VirtualService) getActualState(ctx context.Context, api *gatewayv1beta1.APIRule) (*networkingv1beta1.VirtualService, error) {
	labels := processing.GetOwnerLabels(api)

	var vsList networkingv1beta1.VirtualServiceList
	if err := p.client.List(ctx, &vsList, client.MatchingLabels(labels)); err != nil {
		return nil, err
	}

	if len(vsList.Items) >= 1 {
		return vsList.Items[0], nil
	} else {
		return nil, nil
	}
}

func (p VirtualService) getObjectChanges(desiredVs *networkingv1beta1.VirtualService, actualVs *networkingv1beta1.VirtualService) *processing.ObjectChange {
	if actualVs != nil {
		actualVs.Spec = *desiredVs.Spec.DeepCopy()
		return processing.NewObjectUpdateAction(actualVs)
	} else {
		return processing.NewObjectCreateAction(desiredVs)
	}
}

func (p VirtualService) generateVirtualService(api *gatewayv1beta1.APIRule) *networkingv1beta1.VirtualService {
	virtualServiceNamePrefix := fmt.Sprintf("%s-", api.ObjectMeta.Name)
	ownerRef := processing.GenerateOwnerRef(api)

	vsSpecBuilder := builders.VirtualServiceSpec()
	vsSpecBuilder.Host(helpers.GetHostWithDomain(*api.Spec.Host, p.defaultDomainName))
	vsSpecBuilder.Gateway(*api.Spec.Gateway)
	filteredRules := filterDuplicatePaths(api.Spec.Rules)

	for _, rule := range filteredRules {
		httpRouteBuilder := builders.HTTPRoute()
		host, port := p.oathkeeperSvc, p.oathkeeperSvcPort
		serviceNamespace := helpers.FindServiceNamespace(api, &rule)

		if !processing.IsSecured(rule) {
			// Use rule level service if it exists
			if rule.Service != nil {
				host = fmt.Sprintf("%s.%s.svc.cluster.local", *rule.Service.Name, *serviceNamespace)
				port = *rule.Service.Port
			} else {
				// Otherwise use service defined on APIRule spec level
				host = fmt.Sprintf("%s.%s.svc.cluster.local", *api.Spec.Service.Name, *serviceNamespace)
				port = *api.Spec.Service.Port
			}
		}

		httpRouteBuilder.Route(builders.RouteDestination().Host(host).Port(port))
		httpRouteBuilder.Match(builders.MatchRequest().Uri().Regex(rule.Path))
		httpRouteBuilder.CorsPolicy(builders.CorsPolicy().
			AllowOrigins(p.corsConfig.AllowOrigins...).
			AllowMethods(p.corsConfig.AllowMethods...).
			AllowHeaders(p.corsConfig.AllowHeaders...))
		httpRouteBuilder.Headers(builders.Headers().
			SetHostHeader(helpers.GetHostWithDomain(*api.Spec.Host, p.defaultDomainName)))
		vsSpecBuilder.HTTP(httpRouteBuilder)

	}

	vsBuilder := builders.VirtualService().
		GenerateName(virtualServiceNamePrefix).
		Namespace(api.ObjectMeta.Namespace).
		Owner(builders.OwnerReference().From(&ownerRef)).
		Label(processing.OwnerLabel, fmt.Sprintf("%s.%s", api.ObjectMeta.Name, api.ObjectMeta.Namespace)).
		Label(processing.OwnerLabelv1alpha1, fmt.Sprintf("%s.%s", api.ObjectMeta.Name, api.ObjectMeta.Namespace))

	for k, v := range p.additionalLabels {
		vsBuilder.Label(k, v)
	}

	vsBuilder.Spec(vsSpecBuilder)

	return vsBuilder.Get()
}

func filterDuplicatePaths(rules []gatewayv1beta1.Rule) []gatewayv1beta1.Rule {
	duplicates := make(map[string]bool)
	var filteredRules []gatewayv1beta1.Rule
	for _, rule := range rules {
		if _, exists := duplicates[rule.Path]; !exists {
			duplicates[rule.Path] = true
			filteredRules = append(filteredRules, rule)
		}
	}

	return filteredRules
}
