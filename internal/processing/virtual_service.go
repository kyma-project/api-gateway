package processing

import (
	"context"
	"fmt"
	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	"github.com/kyma-incubator/api-gateway/internal/builders"
	"github.com/kyma-incubator/api-gateway/internal/helpers"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type VirtualServiceProcessor struct {
	client            client.Client
	ctx               context.Context
	oathkeeperSvc     string
	oathkeeperSvcPort uint32
	corsConfig        *CorsConfig
	additionalLabels  map[string]string
	defaultDomainName string
}

func NewVirtualServiceProcessor(config ReconciliationConfig) VirtualServiceProcessor {
	return VirtualServiceProcessor{
		client:            config.client,
		ctx:               config.ctx,
		oathkeeperSvc:     config.oathkeeperSvc,
		oathkeeperSvcPort: config.oathkeeperSvcPort,
		corsConfig:        config.corsConfig,
		additionalLabels:  config.additionalLabels,
		defaultDomainName: config.defaultDomainName,
	}
}

func (v VirtualServiceProcessor) process(apiRule *gatewayv1beta1.APIRule) (gatewayv1beta1.StatusCode, error) {
	desiredObjectState := v.getDesiredObject(apiRule)
	actualObjectState, err := v.getActualState(v.ctx, apiRule)
	if err != nil {
		// TODO: Why is the status Skipped (not Error) when fetching of acual state fails?
		return gatewayv1beta1.StatusSkipped, err
	}

	objectsToApply := v.getDiff(desiredObjectState, actualObjectState)
	err = applyDiff(v.ctx, v.client, objectsToApply)
	if err != nil {
		// TODO: Old comment was:
		//  "We don't know exactly which object(s) are not updated properly. The safest approach is to assume nothing is correct and just use `StatusError`."
		//  Can we improve the Status here?
		return gatewayv1beta1.StatusError, err
	}

	return gatewayv1beta1.StatusOK, nil
}

func (v VirtualServiceProcessor) getDesiredObject(api *gatewayv1beta1.APIRule) *networkingv1beta1.VirtualService {
	return v.generateVirtualService(api)
}

func (v VirtualServiceProcessor) getActualState(ctx context.Context, api *gatewayv1beta1.APIRule) (*networkingv1beta1.VirtualService, error) {
	labels := getOwnerLabels(api)

	var vsList networkingv1beta1.VirtualServiceList
	if err := v.client.List(ctx, &vsList, client.MatchingLabels(labels)); err != nil {
		return nil, err
	}

	if len(vsList.Items) >= 1 {
		return vsList.Items[0], nil
	} else {
		return nil, nil
	}
}

func (v VirtualServiceProcessor) getDiff(desiredVs *networkingv1beta1.VirtualService, actualVs *networkingv1beta1.VirtualService) *ObjToPatch {
	if actualVs != nil {
		update(actualVs, desiredVs)
		return NewUpdateObjectAction(actualVs)
	} else {
		return NewCreateObjectAction(desiredVs)
	}
}

func (v VirtualServiceProcessor) generateVirtualService(api *gatewayv1beta1.APIRule) *networkingv1beta1.VirtualService {
	virtualServiceNamePrefix := fmt.Sprintf("%s-", api.ObjectMeta.Name)
	ownerRef := generateOwnerRef(api)

	vsSpecBuilder := builders.VirtualServiceSpec()
	vsSpecBuilder.Host(helpers.GetHostWithDomain(*api.Spec.Host, v.defaultDomainName))
	vsSpecBuilder.Gateway(*api.Spec.Gateway)
	filteredRules := filterDuplicatePaths(api.Spec.Rules)

	for _, rule := range filteredRules {
		httpRouteBuilder := builders.HTTPRoute()
		host, port := v.oathkeeperSvc, v.oathkeeperSvcPort
		serviceNamespace := helpers.FindServiceNamespace(api, &rule)

		if !isSecured(rule) {
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
			AllowOrigins(v.corsConfig.AllowOrigins...).
			AllowMethods(v.corsConfig.AllowMethods...).
			AllowHeaders(v.corsConfig.AllowHeaders...))
		httpRouteBuilder.Headers(builders.Headers().
			SetHostHeader(helpers.GetHostWithDomain(*api.Spec.Host, v.defaultDomainName)))
		vsSpecBuilder.HTTP(httpRouteBuilder)

	}

	vsBuilder := builders.VirtualService().
		GenerateName(virtualServiceNamePrefix).
		Namespace(api.ObjectMeta.Namespace).
		Owner(builders.OwnerReference().From(&ownerRef)).
		Label(OwnerLabel, fmt.Sprintf("%s.%s", api.ObjectMeta.Name, api.ObjectMeta.Namespace)).
		Label(OwnerLabelv1alpha1, fmt.Sprintf("%s.%s", api.ObjectMeta.Name, api.ObjectMeta.Namespace))

	for k, v := range v.additionalLabels {
		vsBuilder.Label(k, v)
	}

	vsBuilder.Spec(vsSpecBuilder)

	return vsBuilder.Get()
}

func update(existing, required *networkingv1beta1.VirtualService) {
	existing.Spec = *required.Spec.DeepCopy()
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
