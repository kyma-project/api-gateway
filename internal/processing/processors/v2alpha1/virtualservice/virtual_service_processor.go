package virtualservice

import (
	"context"
	"fmt"
	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	"github.com/kyma-project/api-gateway/internal/builders"
	"github.com/kyma-project/api-gateway/internal/processing"
	"github.com/kyma-project/api-gateway/internal/processing/default_domain"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

const defaultHttpTimeout uint32 = 180

func NewVirtualServiceProcessor(config processing.ReconciliationConfig, apiRule *gatewayv2alpha1.APIRule) VirtualServiceProcessor {
	return VirtualServiceProcessor{
		ApiRule: apiRule,
		Creator: virtualServiceCreator{
			defaultDomainName: config.DefaultDomainName,
		},
	}
}

// VirtualServiceProcessor is the generic processor that handles the Virtual Service in the reconciliation of API Rule.
type VirtualServiceProcessor struct {
	ApiRule *gatewayv2alpha1.APIRule
	Creator VirtualServiceCreator
}

// VirtualServiceCreator provides the creation of a Virtual Service using the configuration in the given APIRule.
type VirtualServiceCreator interface {
	Create(api *gatewayv2alpha1.APIRule) (*networkingv1beta1.VirtualService, error)
}

// EvaluateReconciliation evaluates the reconciliation of the Virtual Service for the given API Rule.
func (r VirtualServiceProcessor) EvaluateReconciliation(ctx context.Context, client ctrlclient.Client) ([]*processing.ObjectChange, error) {
	desired, err := r.getDesiredState(r.ApiRule)
	if err != nil {
		return make([]*processing.ObjectChange, 0), err
	}

	actual, err := r.getActualState(ctx, client, r.ApiRule)
	if err != nil {
		return make([]*processing.ObjectChange, 0), err
	}

	changes := r.getObjectChanges(desired, actual)

	return []*processing.ObjectChange{changes}, nil
}

func (r VirtualServiceProcessor) getDesiredState(api *gatewayv2alpha1.APIRule) (*networkingv1beta1.VirtualService, error) {
	return r.Creator.Create(api)
}

func (r VirtualServiceProcessor) getActualState(ctx context.Context, client ctrlclient.Client, api *gatewayv2alpha1.APIRule) (*networkingv1beta1.VirtualService, error) {
	labels := processing.GetOwnerLabelsV2alpha1(api)

	var vsList networkingv1beta1.VirtualServiceList
	if err := client.List(ctx, &vsList, ctrlclient.MatchingLabels(labels)); err != nil {
		return nil, err
	}

	if len(vsList.Items) >= 1 {
		return vsList.Items[0], nil
	} else {
		return nil, nil
	}
}

func (r VirtualServiceProcessor) getObjectChanges(desired *networkingv1beta1.VirtualService, actual *networkingv1beta1.VirtualService) *processing.ObjectChange {
	if actual != nil {
		actual.Spec = *desired.Spec.DeepCopy()
		return processing.NewObjectUpdateAction(actual)
	} else {
		return processing.NewObjectCreateAction(desired)
	}
}

type virtualServiceCreator struct {
	defaultDomainName string
}

// Create returns the Virtual Service using the configuration of the APIRule.
func (r virtualServiceCreator) Create(api *gatewayv2alpha1.APIRule) (*networkingv1beta1.VirtualService, error) {
	virtualServiceNamePrefix := fmt.Sprintf("%s-", api.ObjectMeta.Name)

	vsSpecBuilder := builders.VirtualServiceSpec()
	for _, host := range api.Spec.Hosts {
		vsSpecBuilder.AddHost(default_domain.GetHostWithDomain(string(*host), r.defaultDomainName))
	}

	vsSpecBuilder.Gateway(*api.Spec.Gateway)

	for _, rule := range api.Spec.Rules {
		httpRouteBuilder := builders.HTTPRoute()
		serviceNamespace, err := gatewayv2alpha1.FindServiceNamespace(api, rule)
		if err != nil {
			return nil, fmt.Errorf("finding service namespace: %w", err)
		}

		var host string
		var port uint32

		// Use rule level service if it exists
		if rule.Service != nil {
			host = default_domain.GetHostLocalDomain(*rule.Service.Name, serviceNamespace)
			port = *rule.Service.Port
		} else {
			// Otherwise, use service defined on APIRule spec level
			host = default_domain.GetHostLocalDomain(*api.Spec.Service.Name, serviceNamespace)
			port = *api.Spec.Service.Port
		}

		httpRouteBuilder.Route(builders.RouteDestination().Host(host).Port(port))

		matchBuilder := builders.MatchRequest().MethodRegExV2Alpha1(rule.Methods...)

		if rule.AppliesToAllPaths() {
			matchBuilder.Uri().Prefix("/")
		} else {
			matchBuilder.Uri().Regex(rule.Path)
		}

		httpRouteBuilder.Match(matchBuilder)

		httpRouteBuilder.Timeout(time.Duration(GetVirtualServiceHttpTimeout(api.Spec, rule)) * time.Second)

		headersBuilder := builders.NewHttpRouteHeadersBuilder().
			// For now, the X-Forwarded-Host header is set to the first host in the APIRule hosts list.
			// The status of this header is still under discussion in the following GitHub issue:
			// https://github.com/kyma-project/api-gateway/issues/1159
			SetHostHeader(default_domain.GetHostWithDomain(string(*api.Spec.Hosts[0]), r.defaultDomainName))

		if rule.Request != nil {
			if rule.Request.Headers != nil {
				headersBuilder.SetRequestHeaders(rule.Request.Headers)
			}

			if rule.Request.Cookies != nil {
				headersBuilder.SetRequestCookies(rule.Request.Cookies)
			}
		}

		if api.Spec.CorsPolicy != nil {
			httpRouteBuilder.CorsPolicy(builders.CorsPolicy().FromV2Alpha1ApiRuleCorsPolicy(*api.Spec.CorsPolicy))
		}
		headersBuilder.RemoveUpstreamCORSPolicyHeaders()

		httpRouteBuilder.Headers(headersBuilder.Get())

		vsSpecBuilder.HTTP(httpRouteBuilder)

	}

	vsBuilder := builders.VirtualService().
		GenerateName(virtualServiceNamePrefix).
		Namespace(api.ObjectMeta.Namespace).
		Label(processing.OwnerLabel, fmt.Sprintf("%s.%s", api.ObjectMeta.Name, api.ObjectMeta.Namespace))

	vsBuilder.Spec(vsSpecBuilder)

	return vsBuilder.Get(), nil
}

func GetVirtualServiceHttpTimeout(apiRuleSpec gatewayv2alpha1.APIRuleSpec, rule gatewayv2alpha1.Rule) uint32 {
	if rule.Timeout != nil {
		return uint32(*rule.Timeout)
	}

	if apiRuleSpec.Timeout != nil {
		return uint32(*apiRuleSpec.Timeout)
	}
	return defaultHttpTimeout
}
