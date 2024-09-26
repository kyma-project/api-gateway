package v1beta1

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"github.com/kyma-project/api-gateway/internal/helpers"
	"github.com/kyma-project/api-gateway/internal/processing/default_domain"
	"github.com/kyma-project/api-gateway/internal/validation"
	"google.golang.org/appengine/log"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apiv1beta1 "istio.io/api/type/v1beta1"
)

// APIRuleValidator is used to validate github.com/kyma-project/api-gateway/api/v1beta1/APIRule instances
type APIRuleValidator struct {
	ApiRule *gatewayv1beta1.APIRule

	HandlerValidator          handlerValidator
	AccessStrategiesValidator accessStrategyValidator
	MutatorsValidator         mutatorValidator
	InjectionValidator        *validation.InjectionValidator
	RulesValidator            rulesValidator
	ServiceBlockList          map[string][]string
	DomainAllowList           []string
	HostBlockList             []string
	DefaultDomainName         string
}

type accessStrategyValidator interface {
	Validate(attrPath string, accessStrategies []*gatewayv1beta1.Authenticator) []validation.Failure
}

type mutatorValidator interface {
	Validate(attrPath string, rule gatewayv1beta1.Rule) []validation.Failure
}

type rulesValidator interface {
	Validate(attrPath string, rules []gatewayv1beta1.Rule) []validation.Failure
}

// Validate performs APIRule validation
func (v *APIRuleValidator) Validate(ctx context.Context, client client.Client, vsList networkingv1beta1.VirtualServiceList, _ networkingv1beta1.GatewayList) []validation.Failure {
	var failures []validation.Failure

	//Validate service on path level if it is created
	if v.ApiRule.Spec.Service != nil {
		failures = append(failures, v.validateService(".spec.service", v.ApiRule)...)
	}
	failures = append(failures, v.validateHost(".spec.host", vsList, v.ApiRule)...)
	failures = append(failures, v.validateRules(ctx, client, ".spec.rules", v.ApiRule.Spec.Service == nil, v.ApiRule)...)

	return failures
}

func (v *APIRuleValidator) validateHost(attributePath string, vsList networkingv1beta1.VirtualServiceList, api *gatewayv1beta1.APIRule) []validation.Failure {
	var problems []validation.Failure
	if api.Spec.Host == nil {
		problems = append(problems, validation.Failure{
			AttributePath: attributePath,
			Message:       "Host was nil",
		})
		return problems
	}

	host := *api.Spec.Host
	if !default_domain.HostIncludesDomain(*api.Spec.Host) {
		if v.DefaultDomainName == "" {
			problems = append(problems, validation.Failure{
				AttributePath: attributePath,
				Message:       "Host does not contain a domain name and no default domain name is configured",
			})
		}
		host = default_domain.BuildHostWithDomain(host, v.DefaultDomainName)
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
			problems = append(problems, validation.Failure{
				AttributePath: attributePath,
				Message:       "Host is not allowlisted",
			})
		}
	}

	for _, blockedHost := range v.HostBlockList {
		host := *api.Spec.Host
		if blockedHost == host {
			subdomain := strings.Split(host, ".")[0]
			problems = append(problems, validation.Failure{
				AttributePath: attributePath,
				Message:       fmt.Sprintf("The subdomain %s is blocklisted for %s domain", subdomain, v.DefaultDomainName),
			})
		}
	}

	for _, vs := range vsList.Items {
		if occupiesHost(vs, host) && !ownedBy(vs, api) {
			problems = append(problems, validation.Failure{
				AttributePath: attributePath,
				Message:       "This host is occupied by another Virtual Service",
			})
		}
	}

	return problems
}

func (v *APIRuleValidator) validateService(attributePath string, api *gatewayv1beta1.APIRule) []validation.Failure {
	var problems []validation.Failure

	for namespace, services := range v.ServiceBlockList {
		for _, svc := range services {
			serviceNamespace := helpers.FindServiceNamespace(api, nil)
			if api != nil && svc == *api.Spec.Service.Name && namespace == serviceNamespace {
				problems = append(problems, validation.Failure{
					AttributePath: attributePath + ".name",
					Message:       fmt.Sprintf("Service %s in namespace %s is blocklisted", svc, namespace),
				})
			}
		}
	}

	return problems
}

// Validates whether all rules are defined correctly
// Checks whether all rules have service defined for them if checkForService is true
func (v *APIRuleValidator) validateRules(ctx context.Context, client client.Client, attributePath string, checkForService bool, api *gatewayv1beta1.APIRule) []validation.Failure {
	var problems []validation.Failure

	rules := api.Spec.Rules
	if len(rules) == 0 {
		problems = append(problems, validation.Failure{AttributePath: attributePath, Message: "No rules defined"})
		return problems
	}

	if hasPathAndMethodDuplicates(rules) {
		problems = append(problems, validation.Failure{AttributePath: attributePath, Message: "multiple rules defined for the same path and method"})
	}

	for i, r := range rules {
		attributePathWithRuleIndex := fmt.Sprintf("%s[%d]", attributePath, i)
		if checkForService && r.Service == nil {
			problems = append(problems, validation.Failure{AttributePath: attributePathWithRuleIndex + ".service", Message: "No service defined with no main service on spec level"})
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
						problems = append(problems, validation.Failure{
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

func (v *APIRuleValidator) validateAccessStrategies(attributePath string, accessStrategies []*gatewayv1beta1.Authenticator, selector *apiv1beta1.WorkloadSelector, namespace string) []validation.Failure {
	var problems []validation.Failure

	if len(accessStrategies) == 0 {
		problems = append(problems, validation.Failure{AttributePath: attributePath, Message: "No accessStrategies defined"})
		return problems
	}

	problems = append(problems, v.AccessStrategiesValidator.Validate(attributePath, accessStrategies)...)
	problems = append(problems, CheckForSecureAndUnsecureAccessStrategies(accessStrategies, attributePath)...)

	for i, r := range accessStrategies {
		strategyAttrPath := attributePath + fmt.Sprintf("[%d]", i)
		problems = append(problems, v.validateAccessStrategy(strategyAttrPath, r, selector, namespace)...)
	}

	return problems
}

func (v *APIRuleValidator) validateAccessStrategy(attributePath string, accessStrategy *gatewayv1beta1.Authenticator, selector *apiv1beta1.WorkloadSelector, namespace string) []validation.Failure {
	var problems []validation.Failure
	var vld handlerValidator

	switch accessStrategy.Handler.Name {
	case gatewayv1beta1.AccessStrategyAllow:
		vld = vldNoConfig
	case gatewayv1beta1.AccessStrategyNoAuth:
		vld = vldNoConfig
	case gatewayv1beta1.AccessStrategyNoop:
		vld = vldNoConfig
	case gatewayv1beta1.AccessStrategyUnauthorized:
		vld = vldNoConfig
	case gatewayv1beta1.AccessStrategyAnonymous:
		vld = vldNoConfig
	case gatewayv1beta1.AccessStrategyCookieSession:
		vld = vldNoConfig
	case gatewayv1beta1.AccessStrategyOauth2ClientCredentials:
		vld = vldDummy
	case gatewayv1beta1.AccessStrategyOauth2Introspection:
		vld = vldDummy
	case gatewayv1beta1.AccessStrategyJwt:
		vld = v.HandlerValidator
		if v.InjectionValidator != nil {
			injectionProblems, err := v.InjectionValidator.Validate(attributePath+".injection", selector, namespace)
			if err != nil {
				problems = append(problems, validation.Failure{AttributePath: attributePath + ".handler", Message: fmt.Sprintf("Could not find pod for selected service, err: %s", err)})
			} else {
				problems = append(problems, injectionProblems...)
			}
		}
	default:
		return []validation.Failure{{AttributePath: attributePath + ".handler", Message: fmt.Sprintf("Unsupported accessStrategy: %s", accessStrategy.Handler.Name)}}
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

// Validators for AccessStrategies
var vldNoConfig = &noConfigAccStrValidator{}
var vldDummy = &dummyHandlerValidator{}

type handlerValidator interface {
	Validate(attrPath string, Handler *gatewayv1beta1.Handler) []validation.Failure
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

func hasPathAndMethodDuplicates(rules []gatewayv1beta1.Rule) bool {
	duplicates := map[string]bool{}

	if len(rules) > 1 {
		for _, rule := range rules {
			if len(rule.Methods) > 0 {
				for _, method := range rule.Methods {
					tmp := fmt.Sprintf("%s:%s", rule.Path, method)
					if duplicates[tmp] {
						return true
					}
					duplicates[tmp] = true
				}
			} else {
				if duplicates[rule.Path] {
					return true
				}
				duplicates[rule.Path] = true
			}
		}
	}

	return false
}

func getOwnerLabels(api *gatewayv1beta1.APIRule) map[string]string {
	OwnerLabelv1beta1 := fmt.Sprintf("%s.%s", "apirule", gatewayv1beta1.GroupVersion.String())
	labels := make(map[string]string)
	labels[OwnerLabelv1beta1] = fmt.Sprintf("%s.%s", api.ObjectMeta.Name, api.ObjectMeta.Namespace)
	return labels
}
