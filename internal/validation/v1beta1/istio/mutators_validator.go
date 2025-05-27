package istio

import (
	"encoding/json"
	"fmt"

	"github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"github.com/kyma-project/api-gateway/internal/processing"
	"github.com/kyma-project/api-gateway/internal/validation"
)

// mutatorsValidator is used to validate Istio-based mutator configurations. Since currently only the jwt access strategy
// supports these mutators, validation is skipped for rules without jwt access strategy.
type mutatorsValidator struct{}

func (m mutatorsValidator) Validate(attributePath string, rule v1beta1.Rule) []validation.Failure {
	var failures []validation.Failure

	if !processing.IsJwtSecured(rule) {
		return nil
	}

	basePath := fmt.Sprintf("%s%s", attributePath, ".mutators")
	duplicateMutatorFailure := validateMutatorUniqueness(basePath, rule.Mutators)
	failures = append(failures, duplicateMutatorFailure...)

	for mutatorIndex, mutator := range rule.Mutators {
		handlerPath := fmt.Sprintf("%s[%d]%s", basePath, mutatorIndex, ".handler")

		switch mutator.Name {
		case "":
			failures = append(failures, validation.Failure{AttributePath: handlerPath, Message: "mutator handler cannot be empty"})
		case v1beta1.HeaderMutator:
			f := validateHeaderMutator(handlerPath, mutator)
			failures = append(failures, f...)
		case v1beta1.CookieMutator:
			f := validateCookieMutator(handlerPath, mutator)
			failures = append(failures, f...)
		default:
			msg := fmt.Sprintf("unsupported mutator: %s", mutator.Name)
			failures = append(failures, validation.Failure{AttributePath: handlerPath, Message: msg})
		}
	}

	return failures
}

func validateHeaderMutator(handlerPath string, mutator *v1beta1.Mutator) []validation.Failure {
	configPath := fmt.Sprintf("%s%s", handlerPath, ".config")

	if mutator.Config == nil {
		attrPath := fmt.Sprintf("%s%s", handlerPath, ".config")
		return []validation.Failure{
			{AttributePath: attrPath, Message: "headers cannot be empty"},
		}
	}

	var config v1beta1.HeaderMutatorConfig
	err := json.Unmarshal(mutator.Config.Raw, &config)

	if err != nil {
		return []validation.Failure{
			{AttributePath: configPath, Message: "Can't read json: " + err.Error()},
		}
	}

	if !config.HasHeaders() {
		return []validation.Failure{
			{AttributePath: configPath, Message: "headers cannot be empty"},
		}
	}

	for name := range config.Headers {
		if name == "" {
			attrPath := fmt.Sprintf("%s%s", configPath, ".headers.name")
			return []validation.Failure{
				{AttributePath: attrPath, Message: "cannot be empty"},
			}
		}
	}
	return nil
}

func validateCookieMutator(handlerPath string, mutator *v1beta1.Mutator) []validation.Failure {
	configPath := fmt.Sprintf("%s%s", handlerPath, ".config")

	if mutator.Config == nil {
		return []validation.Failure{
			{AttributePath: configPath, Message: "cookies cannot be empty"},
		}
	}

	var config v1beta1.CookieMutatorConfig
	err := json.Unmarshal(mutator.Config.Raw, &config)

	if err != nil {
		return []validation.Failure{
			{AttributePath: configPath, Message: "Can't read json: " + err.Error()},
		}
	}

	if !config.HasCookies() {
		return []validation.Failure{
			{AttributePath: configPath, Message: "cookies cannot be empty"},
		}
	}

	for name := range config.Cookies {
		if name == "" {
			attrPath := fmt.Sprintf("%s%s", configPath, ".cookies.name")
			return []validation.Failure{
				{AttributePath: attrPath, Message: "cannot be empty"},
			}
		}
	}
	return nil
}

func validateMutatorUniqueness(basePath string, mutators []*v1beta1.Mutator) []validation.Failure {
	var failures []validation.Failure
	mutatorsSet := make(map[string]bool)

	for i, mutator := range mutators {
		if mutatorsSet[mutator.Name] {
			attrPath := fmt.Sprintf("%s[%d]%s%s", basePath, i, ".handler.", mutator.Name)
			failures = append(failures, validation.Failure{AttributePath: attrPath, Message: "mutator for same handler already exists"})
		}
		mutatorsSet[mutator.Name] = true
	}

	return failures
}
