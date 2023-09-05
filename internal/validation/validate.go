package validation

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"google.golang.org/appengine/log"
	"strings"

	"github.com/kyma-project/api-gateway/internal/helpers"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"

	gatewayv1alpha1 "github.com/kyma-project/api-gateway/api/v1alpha1"
	gatewayv1beta1 "github.com/kyma-project/api-gateway/api/v1beta1"
	apiv1beta1 "istio.io/api/type/v1beta1"
	"k8s.io/utils/strings/slices"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Validators for AccessStrategies
var vldNoConfig = &noConfigAccStrValidator{}
var vldDummy = &dummyHandlerValidator{}

type handlerValidator interface {
	Validate(attrPath string, Handler *gatewayv1beta1.Handler) []Failure
}

type accessStrategyValidator interface {
	Validate(attrPath string, accessStrategies []*gatewayv1beta1.Authenticator) []Failure
}

type mutatorValidator interface {
	Validate(attrPath string, rule gatewayv1beta1.Rule) []Failure
}

type injectionValidator interface {
	Validate(attrPath string, service *apiv1beta1.WorkloadSelector, namespace string) ([]Failure, error)
}

type rulesValidator interface {
	Validate(attrPath string, rules []gatewayv1beta1.Rule) []Failure
}

// APIRuleValidator is used to validate github.com/kyma-project/api-gateway/api/v1beta1/APIRule instances
type APIRuleValidator struct {
	HandlerValidator          handlerValidator
	AccessStrategiesValidator accessStrategyValidator
	MutatorsValidator         mutatorValidator
	InjectionValidator        injectionValidator
	RulesValidator            rulesValidator
	ServiceBlockList          map[string][]string
	DomainAllowList           []string
	HostBlockList             []string
	DefaultDomainName         string
}

// Failure carries validation failures for a single attribute of an object.
type Failure struct {
	AttributePath string
	Message       string
}

// Validate performs APIRule validation
func (v *APIRuleValidator) Validate(ctx context.Context, client client.Client, api *gatewayv1beta1.APIRule, vsList networkingv1beta1.VirtualServiceList) []Failure {
	var failures []Failure

	//Validate service on path level if it is created
	if api.Spec.Service != nil {
		failures = append(failures, v.validateService(".spec.service", api)...)
	}
	failures = append(failures, v.validateHost(".spec.host", vsList, api)...)
	failures = append(failures, v.validateGateway(".spec.gateway", api.Spec.Gateway)...)
	failures = append(failures, v.validateRules(ctx, client, ".spec.rules", api.Spec.Service == nil, api)...)

	return failures
}

func (v *APIRuleValidator) ValidateConfig(config *helpers.Config) []Failure {
	var problems []Failure

	if config == nil {
		problems = append(problems, Failure{
			Message: "Configuration is missing",
		})
	} else {
		if !slices.Contains([]string{helpers.JWT_HANDLER_ORY, helpers.JWT_HANDLER_ISTIO}, config.JWTHandler) {
			problems = append(problems, Failure{
				Message: fmt.Sprintf("Unsupported JWT Handler: %s", config.JWTHandler),
			})
		}
	}

	return problems
}

func (v *APIRuleValidator) validateHost(attributePath string, vsList networkingv1beta1.VirtualServiceList, api *gatewayv1beta1.APIRule) []Failure {
	var problems []Failure
	if api.Spec.Host == nil {
		problems = append(problems, Failure{
			AttributePath: attributePath,
			Message:       "Host was nil",
		})
		return problems
	}

	host := *api.Spec.Host
	if !helpers.HostIncludesDomain(*api.Spec.Host) {
		if v.DefaultDomainName == "" {
			problems = append(problems, Failure{
				AttributePath: attributePath,
				Message:       "Host does not contain a domain name and no default domain name is configured",
			})
		}
		host = helpers.GetHostWithDefaultDomain(host, v.DefaultDomainName)
	} else if len(v.DomainAllowList) > 0 {
		// Do the allowList check only if the list is actually provided AND the default domain name is not used.
		domainFound := false
		for _, domain := range v.DomainAllowList {
			// service host containing duplicated allowlisted domain should be rejected.
			// for example `my-lambda.kyma.local.kyma.local`
			// service host containing allowlisted domain but only as a part of bigger domain should also be rejected
			// for example `my-lambda.kyma.local.com` when only `kyma.local` is allowlisted
			if count := strings.Count(host, domain); count == 1 && strings.HasSuffix(host, domain) {
				domainFound = true
			}
		}
		if !domainFound {
			problems = append(problems, Failure{
				AttributePath: attributePath,
				Message:       "Host is not allowlisted",
			})
		}
	}

	for _, blockedHost := range v.HostBlockList {
		host := *api.Spec.Host
		if blockedHost == host {
			subdomain := strings.Split(host, ".")[0]
			problems = append(problems, Failure{
				AttributePath: attributePath,
				Message:       fmt.Sprintf("The subdomain %s is blocklisted for %s domain", subdomain, v.DefaultDomainName),
			})
		}
	}

	for _, vs := range vsList.Items {
		if occupiesHost(vs, host) && !ownedBy(vs, api) {
			problems = append(problems, Failure{
				AttributePath: attributePath,
				Message:       "This host is occupied by another Virtual Service",
			})
		}
	}

	return problems
}

func (v *APIRuleValidator) validateService(attributePath string, api *gatewayv1beta1.APIRule) []Failure {
	var problems []Failure

	for namespace, services := range v.ServiceBlockList {
		for _, svc := range services {
			serviceNamespace := helpers.FindServiceNamespace(api, nil)
			if api != nil && svc == *api.Spec.Service.Name && namespace == serviceNamespace {
				problems = append(problems, Failure{
					AttributePath: attributePath + ".name",
					Message:       fmt.Sprintf("Service %s in namespace %s is blocklisted", svc, namespace),
				})
			}
		}
	}

	return problems
}

func (v *APIRuleValidator) validateGateway(attributePath string, gateway *string) []Failure {
	return nil
}

// Validates whether all rules are defined correctly
// Checks whether all rules have service defined for them if checkForService is true
func (v *APIRuleValidator) validateRules(ctx context.Context, client client.Client, attributePath string, checkForService bool, api *gatewayv1beta1.APIRule) []Failure {
	var problems []Failure

	rules := api.Spec.Rules
	if len(rules) == 0 {
		problems = append(problems, Failure{AttributePath: attributePath, Message: "No rules defined"})
		return problems
	}

	if hasPathAndMethodDuplicates(rules) {
		problems = append(problems, Failure{AttributePath: attributePath, Message: "multiple rules defined for the same path and method"})
	}

	for i, r := range rules {
		attributePathWithRuleIndex := fmt.Sprintf("%s[%d]", attributePath, i)
		problems = append(problems, v.validateMethods(attributePathWithRuleIndex+".methods", r.Methods)...)
		if checkForService && r.Service == nil {
			problems = append(problems, Failure{AttributePath: attributePathWithRuleIndex + ".service", Message: "No service defined with no main service on spec level"})
		}
		if r.Service != nil {
			labelSelector, err := helpers.GetLabelSelectorFromService(ctx, client, r.Service, api, &r)
			if err != nil {
				l, errorCtx := logr.FromContext(ctx)
				if errorCtx != nil {
					log.Errorf(ctx, "No logger in context: %s", errorCtx)
				} else {
					l.Info("Couldn't get label selectors for service", "error", err)
				}
			}
			problems = append(problems, v.validateAccessStrategies(attributePathWithRuleIndex+".accessStrategies", r.AccessStrategies, labelSelector, helpers.FindServiceNamespace(api, &r))...)
			for namespace, services := range v.ServiceBlockList {
				for _, svc := range services {
					serviceNamespace := helpers.FindServiceNamespace(api, &r)
					if svc == *r.Service.Name && namespace == serviceNamespace {
						problems = append(problems, Failure{
							AttributePath: attributePathWithRuleIndex + ".service.name",
							Message:       fmt.Sprintf("Service %s in namespace %s is blocklisted", svc, namespace),
						})
					}
				}
			}
		} else if api.Spec.Service != nil {
			labelSelector, err := helpers.GetLabelSelectorFromService(ctx, client, api.Spec.Service, api, nil)
			if err != nil {
				l, errorCtx := logr.FromContext(ctx)
				if errorCtx != nil {
					log.Errorf(ctx, "No logger in context: %s", errorCtx)
				} else {
					l.Info("Couldn't get label selectors for service", "error", err)
				}
			}
			problems = append(problems, v.validateAccessStrategies(attributePathWithRuleIndex+".accessStrategies", r.AccessStrategies, labelSelector, helpers.FindServiceNamespace(api, &r))...)
		}

		if v.MutatorsValidator != nil {
			mutatorFailures := v.MutatorsValidator.Validate(attributePathWithRuleIndex, r)
			problems = append(problems, mutatorFailures...)
		}

	}

	if v.RulesValidator != nil {
		rulesFailures := v.RulesValidator.Validate(".spec.rules", rules)
		problems = append(problems, rulesFailures...)
	}

	return problems
}

func (v *APIRuleValidator) validateMethods(attributePath string, methods []string) []Failure {
	return nil
}

func (v *APIRuleValidator) validateAccessStrategies(attributePath string, accessStrategies []*gatewayv1beta1.Authenticator, selector *apiv1beta1.WorkloadSelector, namespace string) []Failure {
	var problems []Failure

	if len(accessStrategies) == 0 {
		problems = append(problems, Failure{AttributePath: attributePath, Message: "No accessStrategies defined"})
		return problems
	}

	problems = append(problems, v.AccessStrategiesValidator.Validate(attributePath, accessStrategies)...)

	for i, r := range accessStrategies {
		strategyAttrPath := attributePath + fmt.Sprintf("[%d]", i)
		problems = append(problems, v.validateAccessStrategy(strategyAttrPath, r, selector, namespace)...)
	}

	return problems
}

func (v *APIRuleValidator) validateAccessStrategy(attributePath string, accessStrategy *gatewayv1beta1.Authenticator, selector *apiv1beta1.WorkloadSelector, namespace string) []Failure {
	var problems []Failure
	var vld handlerValidator

	switch accessStrategy.Handler.Name {
	case "allow": //our internal constant, does not exist in ORY
		vld = vldNoConfig
	case "noop":
		vld = vldNoConfig
	case "unauthorized":
		vld = vldNoConfig
	case "anonymous":
		vld = vldNoConfig
	case "cookie_session":
		vld = vldNoConfig
	case "oauth2_client_credentials":
		vld = vldDummy
	case "oauth2_introspection":
		vld = vldDummy
	case "jwt":
		vld = v.HandlerValidator
		if v.InjectionValidator != nil {
			injectionProblems, err := v.InjectionValidator.Validate(attributePath+".injection", selector, namespace)
			if err != nil {
				problems = append(problems, Failure{AttributePath: attributePath + ".handler", Message: fmt.Sprintf("Could not find pod for selected service, err: %s", err)})
			} else {
				problems = append(problems, injectionProblems...)
			}
		}
	default:
		return []Failure{{AttributePath: attributePath + ".handler", Message: fmt.Sprintf("Unsupported accessStrategy: %s", accessStrategy.Handler.Name)}}
	}

	return append(problems, vld.Validate(attributePath, accessStrategy.Handler)...)
}

func occupiesHost(vs *networkingv1beta1.VirtualService, host string) bool {
	for _, h := range vs.Spec.Hosts {
		if h == host {
			return true
		}
	}
	return false
}

func getOwnerLabels(api *gatewayv1beta1.APIRule) map[string]string {
	OwnerLabelv1alpha1 := fmt.Sprintf("%s.%s", "apirule", gatewayv1alpha1.GroupVersion.String())
	labels := make(map[string]string)
	labels[OwnerLabelv1alpha1] = fmt.Sprintf("%s.%s", api.ObjectMeta.Name, api.ObjectMeta.Namespace)
	return labels
}

func ownedBy(vs *networkingv1beta1.VirtualService, api *gatewayv1beta1.APIRule) bool {
	ownerLabels := getOwnerLabels(api)
	vsLabels := vs.GetLabels()

	for key, label := range ownerLabels {
		val, ok := vsLabels[key]
		if ok {
			return val == label
		}
	}
	return false
}
