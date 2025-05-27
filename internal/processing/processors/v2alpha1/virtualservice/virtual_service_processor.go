package virtualservice

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	"github.com/kyma-project/api-gateway/internal/builders"
	"github.com/kyma-project/api-gateway/internal/helpers"
	"github.com/kyma-project/api-gateway/internal/processing"
	"github.com/kyma-project/api-gateway/internal/processing/defaultdomain"
)

const defaultHttpTimeout uint32 = 180

var (
	envoyTemplatesTranslation = map[string]string{
		`{**}`: `([A-Za-z0-9-._~!$&'()*+,;=:@/]|%[0-9a-fA-F]{2})*`,
		`{*}`:  `([A-Za-z0-9-._~!$&'()*+,;=:@]|%[0-9a-fA-F]{2})+`,
	}
)

func NewVirtualServiceProcessor(_ processing.ReconciliationConfig, apiRule *gatewayv2alpha1.APIRule, gateway *networkingv1beta1.Gateway) Processor {
	return Processor{
		APIRule: apiRule,
		Creator: virtualServiceCreator{
			gateway: gateway,
		},
	}
}

// Processor is the generic processor that handles the Virtual Service in the reconciliation of API Rule.
type Processor struct {
	APIRule *gatewayv2alpha1.APIRule
	Creator Creator
}

// Creator provides the creation of a Virtual Service using the configuration in the given APIRule.
type Creator interface {
	Create(api *gatewayv2alpha1.APIRule) (*networkingv1beta1.VirtualService, error)
}

// EvaluateReconciliation evaluates the reconciliation of the Virtual Service for the given API Rule.
func (r Processor) EvaluateReconciliation(ctx context.Context, client ctrlclient.Client) ([]*processing.ObjectChange, error) {
	desired, err := r.getDesiredState(r.APIRule)
	if err != nil {
		return make([]*processing.ObjectChange, 0), err
	}

	actual, err := r.getActualState(ctx, client, r.APIRule)
	if err != nil {
		return make([]*processing.ObjectChange, 0), err
	}

	changes := r.getObjectChanges(desired, actual)

	return []*processing.ObjectChange{changes}, nil
}

func (r Processor) getDesiredState(api *gatewayv2alpha1.APIRule) (*networkingv1beta1.VirtualService, error) {
	return r.Creator.Create(api)
}

func (r Processor) getActualState(ctx context.Context, client ctrlclient.Client, api *gatewayv2alpha1.APIRule) (*networkingv1beta1.VirtualService, error) {
	labels := processing.GetOwnerLabelsV2alpha1(api)

	var vsList networkingv1beta1.VirtualServiceList
	if err := client.List(ctx, &vsList, ctrlclient.MatchingLabels(labels)); err != nil {
		return nil, err
	}

	if len(vsList.Items) >= 1 {
		return vsList.Items[0], nil
	}
	return nil, nil
}

func (r Processor) getObjectChanges(desired *networkingv1beta1.VirtualService, actual *networkingv1beta1.VirtualService) *processing.ObjectChange {
	if actual != nil {
		actual.Spec = *desired.Spec.DeepCopy()
		return processing.NewObjectUpdateAction(actual)
	}
	return processing.NewObjectCreateAction(desired)
}

type virtualServiceCreator struct {
	gateway *networkingv1beta1.Gateway
}

// Create returns the Virtual Service using the configuration of the APIRule.
func (r virtualServiceCreator) Create(api *gatewayv2alpha1.APIRule) (*networkingv1beta1.VirtualService, error) {
	virtualServiceNamePrefix := fmt.Sprintf("%s-", api.Name)

	vsSpecBuilder := builders.VirtualServiceSpec()
	hosts, gatewayDomain, err := getHostsAndDomainFromAPIRule(api, r)
	if err != nil {
		return nil, fmt.Errorf("getting hosts from api rule: %w", err)
	}

	for _, host := range hosts {
		vsSpecBuilder.AddHost(host)
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
			host = defaultdomain.GetHostLocalDomain(*rule.Service.Name, serviceNamespace)
			port = *rule.Service.Port
		} else {
			// Otherwise, use service defined on APIRule spec level
			host = defaultdomain.GetHostLocalDomain(*api.Spec.Service.Name, serviceNamespace)
			port = *api.Spec.Service.Port
		}

		httpRouteBuilder.Route(builders.RouteDestination().Host(host).Port(port))

		matchBuilder := builders.MatchRequest().MethodRegExV2Alpha1(rule.Methods...)

		if rule.AppliesToAllPaths() {
			matchBuilder.Uri().Prefix("/")
		} else {
			matchBuilder.Uri().Regex(prepareRegexPath(rule.Path))
		}

		httpRouteBuilder.Match(matchBuilder)

		httpRouteBuilder.Timeout(time.Duration(GetVirtualServiceHTTPTimeout(api.Spec, rule)) * time.Second)

		headersBuilder := builders.NewHttpRouteHeadersBuilder().
			// For now, the X-Forwarded-Host header is set to the first host in the APIRule hosts list.
			// The status of this header is still under discussion in the following GitHub issue:
			// https://github.com/kyma-project/api-gateway/issues/1159
			SetHostHeader(defaultdomain.GetHostWithDomain(string(*api.Spec.Hosts[0]), gatewayDomain))

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
		Label(processing.OwnerLabel, fmt.Sprintf("%s.%s", api.Name, api.Namespace))

	vsBuilder.Spec(vsSpecBuilder)

	return vsBuilder.Get(), nil
}

func prepareRegexPath(path string) string {
	for key, replace := range envoyTemplatesTranslation {
		path = strings.ReplaceAll(path, key, replace)
	}

	return fmt.Sprintf("^%s$", path)
}

func GetVirtualServiceHTTPTimeout(apiRuleSpec gatewayv2alpha1.APIRuleSpec, rule gatewayv2alpha1.Rule) uint32 {
	if rule.Timeout != nil {
		return uint32(*rule.Timeout)
	}

	if apiRuleSpec.Timeout != nil {
		return uint32(*apiRuleSpec.Timeout)
	}
	return defaultHttpTimeout
}

// getHostsAndDomainFromAPIRule extracts all FQDNs for which the APIRule should match.
// If the APIRule contains short host names, it will use the domain of the specified gateway to generate FQDNs for them.
// This is done by concatenating the short host name with the wildcard domain of the gateway.
// For example, if the short host name is "foo" and the gateway defines itself to ".*.example.com"
// the resulting FQDN will be "foo.example.com".
// For FQDN host names, it will just return the host name as is.
//
// Returns:
//   - a slice of FQDN host names.
//   - the domain used by the gateway.
//   - an error if the gateway is not provided and short host names are used.
func getHostsAndDomainFromAPIRule(api *gatewayv2alpha1.APIRule, r virtualServiceCreator) ([]string, string, error) {
	var hosts []string
	var gatewayDomain string

	if r.gateway != nil {
		for _, server := range r.gateway.Spec.GetServers() {
			if len(server.GetHosts()) > 0 {
				gatewayDomain = strings.TrimPrefix(server.GetHosts()[0], "*.")

				// This break statement here ensures that the host used for the gateway is the first one.
				// Possibly it might be better to return an error if there are multiple different hosts in the same gateway.
				break
			}
		}
	}

	for _, h := range api.Spec.Hosts {
		host := string(*h)
		if !helpers.IsShortHostName(host) {
			hosts = append(hosts, host)
		} else {
			if r.gateway == nil {
				return nil, "", errors.New("gateway must be provided when using short host name")
			}

			if gatewayDomain == "" {
				return nil, "", errors.New("gateway with host definition must be provided when using short host name")
			}
			hosts = append(hosts, defaultdomain.GetHostWithDomain(host, gatewayDomain))
		}
	}

	return hosts, gatewayDomain, nil
}
