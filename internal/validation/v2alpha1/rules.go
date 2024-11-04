package v2alpha1

import (
	"context"
	"fmt"
	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	"github.com/kyma-project/api-gateway/internal/path/segment_trie"
	"github.com/kyma-project/api-gateway/internal/path/token"
	"github.com/kyma-project/api-gateway/internal/validation"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

func validateRules(ctx context.Context, client client.Client, parentAttributePath string, apiRule *gatewayv2alpha1.APIRule) []validation.Failure {
	var problems []validation.Failure
	rulesAttributePath := parentAttributePath + ".rules"

	rules := apiRule.Spec.Rules
	if len(rules) == 0 {
		problems = append(problems, validation.Failure{AttributePath: rulesAttributePath, Message: "No rules defined"})
		return problems
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

		problems = append(problems, validatePath(ruleAttributePath, rule.Path)...)
	}

	if path, method, conflict := hasPathByMethodConflict(rules); conflict {
		problems = append(problems, validation.Failure{AttributePath: rulesAttributePath, Message: fmt.Sprintf("Path %s with method %s conflicts with at least one of the other defined paths", path, method)})
	}

	jwtAuthFailures := validateJwtAuthenticationEquality(rulesAttributePath, rules)
	problems = append(problems, jwtAuthFailures...)

	return problems
}

func validatePath(validationPath string, rulePath string) (problems []validation.Failure) {
	problems = append(problems, validateEnvoyTemplate(validationPath+".path", rulePath)...)
	return problems
}

func validateEnvoyTemplate(validationPath string, path string) []validation.Failure {
	if strings.Count(path, "{**}") == 0 {
		return nil
	}

	if strings.Count(path, "{**}") > 1 {
		return []validation.Failure{
			{
				AttributePath: validationPath,
				Message:       "Only one {**} operator is allowed in the path.",
			},
		}
	}

	segments := strings.Split(strings.TrimLeft(path, "/"), "/")

	var last bool
	for _, segment := range segments {
		if segment == "{*}" || segment == "{**}" {
			if last {
				return []validation.Failure{
					{
						AttributePath: validationPath,
						Message:       "The {**} operator must be the last operator in the path.",
					},
				}
			}
			last = segment == "{**}"
		}
	}

	return nil
}

func hasPathByMethodConflict(rules []gatewayv2alpha1.Rule) (path string, method gatewayv2alpha1.HttpMethod, conflict bool) {
	rulesByMethod := map[gatewayv2alpha1.HttpMethod][]gatewayv2alpha1.Rule{}
	for _, rule := range rules {
		for _, method := range rule.Methods {
			rulesByMethod[method] = append(rulesByMethod[method], rule)
		}
	}

	for m, rules := range rulesByMethod {
		trie := segment_trie.New()
		for _, rule := range rules {
			path = rule.Path
			if rule.Path == "/*" {
				path = "/{**}"
			}

			tokens := token.TokenizePath(path)
			if trie.InsertAndCheckCollisions(tokens) != nil {
				return rule.Path, m, true
			}
		}
	}

	return "", "", false
}
