package v2alpha1

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"sigs.k8s.io/controller-runtime/pkg/client"

	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	"github.com/kyma-project/api-gateway/internal/validation"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
)

type APIRuleValidator struct {
	Api *gatewayv2alpha1.APIRule

	InjectionValidator *validation.InjectionValidator
	RulesValidator     rulesValidator
	JwtValidator       jwtValidator
	ServiceBlockList   map[string][]string
	HostBlockList      []string
	DefaultDomainName  string
}

type jwtValidator interface {
	Validate(attributePath string, handler *gatewayv2alpha1.JwtConfig) []validation.Failure
}

type jwtValidatorImpl struct{}

func (j *jwtValidatorImpl) Validate(attributePath string, jwtConfig *gatewayv2alpha1.JwtConfig) []validation.Failure {
	//TODO implement me
	return make([]validation.Failure, 0)
}

type rulesValidator interface {
	Validate(attributePath string, rules []*gatewayv2alpha1.Rule) []validation.Failure
}

type rulesValidatorImpl struct{}

func (r rulesValidatorImpl) Validate(attributePath string, rules []*gatewayv2alpha1.Rule) []validation.Failure {
	//TODO implement me
	return make([]validation.Failure, 0)
}

func NewAPIRuleValidator(ctx context.Context, client client.Client, api *gatewayv2alpha1.APIRule, defaultDomainName string) *APIRuleValidator {
	return &APIRuleValidator{
		Api:                api,
		InjectionValidator: validation.NewInjectionValidator(ctx, client),
		RulesValidator:     rulesValidatorImpl{},
		JwtValidator:       &jwtValidatorImpl{},
		DefaultDomainName:  defaultDomainName,
	}
}

// Validate performs APIRule validation
func (v *APIRuleValidator) Validate(ctx context.Context, client client.Client, vsList networkingv1beta1.VirtualServiceList) []validation.Failure {
	var failures []validation.Failure

	//Validate service on path level if it is created
	if v.Api.Spec.Service != nil {
		failures = append(failures, v.validateService(".spec.service", v.Api)...)
	}

	failures = append(failures, v.validateHosts(".spec.hosts", vsList, v.Api)...)

	return failures
}

func (v *APIRuleValidator) validateService(attributePath string, api *gatewayv2alpha1.APIRule) []validation.Failure {
	var problems []validation.Failure

	for namespace, services := range v.ServiceBlockList {
		for _, svc := range services {
			serviceNamespace := findServiceNamespace(api, nil)
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

func (v *APIRuleValidator) validateHosts(attributePath string, vsList networkingv1beta1.VirtualServiceList, api *gatewayv2alpha1.APIRule) []validation.Failure {
	var problems []validation.Failure
	if api.Spec.Hosts == nil {
		problems = append(problems, validation.Failure{
			AttributePath: attributePath,
			Message:       "Hosts was nil",
		})
		return problems
	}

	//hosts := api.Spec.Hosts
	for _, host := range api.Spec.Hosts {
		if !hostIsFQDN(string(*host)) {
			problems = append(problems, validation.Failure{
				AttributePath: attributePath,
				Message:       "Host is not fully qualified domain name",
			})
		}
	}

	for _, blockedHost := range v.HostBlockList {
		for _, host := range api.Spec.Hosts {
			if blockedHost == string(*host) {
				subdomain := strings.Split(string(*host), ".")[0]
				problems = append(problems, validation.Failure{
					AttributePath: attributePath,
					Message:       fmt.Sprintf("The subdomain %s is blocklisted for %s domain", subdomain, v.DefaultDomainName),
				})
			}
		}
	}

	for _, vs := range vsList.Items {
		for _, host := range api.Spec.Hosts {
			if occupiesHost(vs, string(*host)) && !ownedBy(vs, api) {
				problems = append(problems, validation.Failure{
					AttributePath: attributePath,
					Message:       "This host is occupied by another Virtual Service",
				})
			}
		}
	}

	return problems
}

func hostIsFQDN(host string) bool {
	if len(host) > 253 {
		return false
	}

	labelRegex := regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?$`)
	labels := strings.Split(host, ".")
	for _, label := range labels {
		if len(label) == 0 || len(label) > 63 {
			return false
		}
		if !labelRegex.MatchString(label) {
			return false
		}
	}

	return true
}

func findServiceNamespace(api *gatewayv2alpha1.APIRule, rule *gatewayv2alpha1.Rule) string {
	// Fallback direction for the upstream service namespace: Rule.Service > Spec.Service > APIRule
	if rule != nil && rule.Service != nil && rule.Service.Namespace != nil {
		return *rule.Service.Namespace
	}
	if api != nil && api.Spec.Service != nil && api.Spec.Service.Namespace != nil {
		return *api.Spec.Service.Namespace
	}
	return api.Namespace
}

func occupiesHost(vs *networkingv1beta1.VirtualService, host string) bool {
	for _, h := range vs.Spec.Hosts {
		if h == host {
			return true
		}
	}
	return false
}

func ownedBy(vs *networkingv1beta1.VirtualService, api *gatewayv2alpha1.APIRule) bool {
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

func getOwnerLabels(api *gatewayv2alpha1.APIRule) map[string]string {
	OwnerLabelv1beta1 := fmt.Sprintf("%s.%s", "apirule", gatewayv1beta1.GroupVersion.String())
	labels := make(map[string]string)
	labels[OwnerLabelv1beta1] = fmt.Sprintf("%s.%s", api.ObjectMeta.Name, api.ObjectMeta.Namespace)
	return labels
}
