package istio

import (
	"encoding/json"
	"fmt"
	"github.com/kyma-project/api-gateway/api/v1beta1"
	"github.com/kyma-project/api-gateway/internal/processing"
	"github.com/kyma-project/api-gateway/internal/validation"
)

type mutatorsValidator struct {
}

func (m mutatorsValidator) Validate(attributePath string, rule v1beta1.Rule) []validation.Failure {
	var failures []validation.Failure

	if !processing.IsJwtSecured(rule) {
		return nil
	}

	duplicateMutatorFailure := validateMutatorUniqueness(attributePath, rule.Mutators)
	failures = append(failures, duplicateMutatorFailure...)

	for mutatorIndex, mutator := range rule.Mutators {

		switch mutator.Name {
		case "":
			attrPath := fmt.Sprintf("%s%s[%d]%s", attributePath, ".mutators", mutatorIndex, ".handler")
			failures = append(failures, validation.Failure{AttributePath: attrPath, Message: "mutator handler cannot be empty"})
		case v1beta1.HeaderMutator:
			f := validateHeaderMutator(attributePath, mutator, mutatorIndex)
			failures = append(failures, f...)
		case v1beta1.CookieMutator:
			f := validateCookieMutator(attributePath, mutator, mutatorIndex)
			failures = append(failures, f...)
		default:
			attrPath := fmt.Sprintf("%s%s[%d]%s", attributePath, ".mutators", mutatorIndex, ".handler")
			msg := fmt.Sprintf("unsupported mutator: %s", mutator.Name)
			failures = append(failures, validation.Failure{AttributePath: attrPath, Message: msg})
		}

	}

	return failures
}

func validateHeaderMutator(attributePath string, mutator *v1beta1.Mutator, mutatorIndex int) []validation.Failure {

	if mutator.Config == nil {
		attrPath := fmt.Sprintf("%s%s[%d]%s", attributePath, ".mutators", mutatorIndex, ".handler.config")
		return []validation.Failure{
			{AttributePath: attrPath, Message: "headers cannot be empty"},
		}
	}

	var config v1beta1.HeaderMutatorConfig
	err := json.Unmarshal(mutator.Config.Raw, &config)

	if err != nil {
		return []validation.Failure{
			{AttributePath: attributePath + ".config", Message: "Can't read json: " + err.Error()},
		}
	}

	if !config.HasHeaders() {
		attrPath := fmt.Sprintf("%s%s[%d]%s", attributePath, ".mutators", mutatorIndex, ".handler.config")
		return []validation.Failure{
			{AttributePath: attrPath, Message: "headers cannot be empty"},
		}
	}

	for name := range config.Headers {
		if name == "" {
			attrPath := fmt.Sprintf("%s%s[%d]%s", attributePath, ".mutators", mutatorIndex, ".handler.config.headers.name")
			return []validation.Failure{
				{AttributePath: attrPath, Message: "cannot be empty"},
			}
		}

	}

	return nil
}

func validateCookieMutator(attributePath string, mutator *v1beta1.Mutator, mutatorIndex int) []validation.Failure {

	if mutator.Config == nil {
		attrPath := fmt.Sprintf("%s%s[%d]%s", attributePath, ".mutators", mutatorIndex, ".handler.config")
		return []validation.Failure{
			{AttributePath: attrPath, Message: "cookies cannot be empty"},
		}
	}

	var config v1beta1.CookieMutatorConfig
	err := json.Unmarshal(mutator.Config.Raw, &config)

	if err != nil {
		return []validation.Failure{
			{AttributePath: attributePath + ".config", Message: "Can't read json: " + err.Error()},
		}
	}

	if !config.HasCookies() {
		attrPath := fmt.Sprintf("%s%s[%d]%s", attributePath, ".mutators", mutatorIndex, ".handler.config")
		return []validation.Failure{
			{AttributePath: attrPath, Message: "cookies cannot be empty"},
		}
	}

	for name := range config.Cookies {
		if name == "" {
			attrPath := fmt.Sprintf("%s%s[%d]%s", attributePath, ".mutators", mutatorIndex, ".handler.config.cookies.name")
			return []validation.Failure{
				{AttributePath: attrPath, Message: "cannot be empty"},
			}
		}

	}

	return nil
}

func validateMutatorUniqueness(attributePath string, mutators []*v1beta1.Mutator) []validation.Failure {
	var failures []validation.Failure
	mutatorsSet := make(map[string]bool)

	for i, mutator := range mutators {
		if mutatorsSet[mutator.Name] {
			attrPath := fmt.Sprintf("%s%s[%d]%s%s", attributePath, ".mutators", i, ".handler.", mutator.Name)
			failures = append(failures, validation.Failure{AttributePath: attrPath, Message: "mutator for same handler already exists"})
		}
		mutatorsSet[mutator.Name] = true
	}

	return failures
}
