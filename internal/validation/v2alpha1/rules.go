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
		injectionFailures, err := validateSidecarInjection(ctx, client, ruleAttributePath, apiRule, rule)
		if err != nil {
			problems = append(problems, validation.Failure{AttributePath: ruleAttributePath, Message: fmt.Sprintf("Failed to execute sidecar injection validation, err: %s", err)})
		}

		problems = append(problems, injectionFailures...)

		if rule.ExtAuth != nil {
			extAuthFailures, err := validateExtAuthProviders(ctx, client, ruleAttributePath, rule)
			if err != nil {
				problems = append(problems, validation.Failure{AttributePath: ruleAttributePath, Message: fmt.Sprintf("Failed to execute external auth provider validation, err: %s", err)})
			}

			problems = append(problems, extAuthFailures...)
		}

		problems = append(problems, validatePath(ruleAttributePath, rule.Path)...)
	}

	problems = append(problems, hasPathByMethodConflict(rulesAttributePath, rules)...)

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

const pathByMethodConflictTemplate = "Path %s with method %s conflicts with at least one of the previous rule paths"

func pathByMethodConflictValidationError(validationPath string, path string, method string) []validation.Failure {
	return []validation.Failure{
		{
			AttributePath: validationPath,
			Message:       fmt.Sprintf(pathByMethodConflictTemplate, path, method),
		},
	}
}

func hasPathByMethodConflict(validationPath string, rules []gatewayv2alpha1.Rule) []validation.Failure {
	pathsByMethod := make(map[gatewayv2alpha1.HttpMethod][]string)
	for _, rule := range rules {
		if len(rule.Methods) == 0 {
			pathsByMethod["NO_METHODS"] = append(pathsByMethod["NO_METHODS"], rule.Path)
		}

		for _, method := range rule.Methods {
			pathsByMethod[method] = append(pathsByMethod[method], rule.Path)
		}
	}

	for m, paths := range pathsByMethod {
		trie := segment_trie.New()
		for _, path := range paths {
			if path == "/*" {
				path = "/{**}"
			}

			tokens := token.TokenizePath(path)
			if trie.InsertAndCheckCollisions(tokens) != nil {
				return pathByMethodConflictValidationError(validationPath, path, string(m))
			}
		}
	}

	return nil
}
