package v2alpha1

import (
	"context"
	"fmt"
	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	"github.com/kyma-project/api-gateway/internal/validation"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func validateRules(ctx context.Context, client client.Client, parentAttributePath string, apiRule *gatewayv2alpha1.APIRule) []validation.Failure {
	var problems []validation.Failure
	rulesAttributePath := parentAttributePath + ".rules"

	rules := apiRule.Spec.Rules
	if len(rules) == 0 {
		problems = append(problems, validation.Failure{AttributePath: rulesAttributePath, Message: "No rules defined"})
		return problems
	}

	if hasPathAndMethodDuplicates(rules) {
		problems = append(problems, validation.Failure{AttributePath: rulesAttributePath, Message: "multiple rules defined for the same path and method"})
	}

	for i, rule := range rules {
		ruleAttributePath := fmt.Sprintf("%s[%d]", rulesAttributePath, i)

		if apiRule.Spec.Service == nil && rule.Service == nil {
			problems = append(problems, validation.Failure{AttributePath: ruleAttributePath + ".service", Message: "The rule must define a service, because no service is defined on spec level"})
		}

		problems = append(problems, validateJwt(ruleAttributePath, &rule)...)
		if rule.NoAuth == nil || !*rule.NoAuth {
			injectionFailures, err := validateSidecarInjection(ctx, client, ruleAttributePath, apiRule, rule)
			if err != nil {
				problems = append(problems, validation.Failure{AttributePath: ruleAttributePath, Message: fmt.Sprintf("Failed to execute sidecar injection validation, err: %s", err)})
			}

			problems = append(problems, injectionFailures...)
		}

		if rule.ExtAuth != nil {
			extAuthFailures, err := validateExtAuthProviders(ctx, client, ruleAttributePath, rule)
			if err != nil {
				problems = append(problems, validation.Failure{AttributePath: ruleAttributePath, Message: fmt.Sprintf("Failed to execute external auth provider validation, err: %s", err)})
			}

			problems = append(problems, extAuthFailures...)
		}

	}

	jwtAuthFailures := validateJwtAuthenticationEquality(rulesAttributePath, rules)
	problems = append(problems, jwtAuthFailures...)

	return problems
}

func hasPathAndMethodDuplicates(rules []gatewayv2alpha1.Rule) bool {
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
