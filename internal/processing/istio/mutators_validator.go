package istio

import (
	"encoding/json"
	"fmt"
	"github.com/kyma-project/api-gateway/api/v1beta1"
	"github.com/kyma-project/api-gateway/internal/validation"
)

const (
	cookieMutator = "cookie"
	headerMutator = "header"
)

type mutatorsValidator struct {
}

func (m mutatorsValidator) Validate(attributePath string, rule v1beta1.Rule) []validation.Failure {

	var failures []validation.Failure

	for mutatorIndex, mutator := range rule.Mutators {

		switch mutator.Name {
		case "":
			attrPath := fmt.Sprintf("%s%s[%d]%s", attributePath, ".mutators", mutatorIndex, ".handler")
			failures = append(failures, validation.Failure{AttributePath: attrPath, Message: "handler cannot be empty"})
		case headerMutator:
			f := validateHeaderMutator(attributePath, mutator, mutatorIndex)
			failures = append(failures, f...)
		case cookieMutator:
			f := validateCookieMutator(attributePath, mutator, mutatorIndex)
			failures = append(failures, f...)
		default:
			attrPath := fmt.Sprintf("%s%s[%d]%s", attributePath, ".mutators", mutatorIndex, ".handler")
			msg := fmt.Sprintf("unsupported handler: %s", mutator.Name)
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

	if config.Headers == nil || len(config.Headers) == 0 {
		attrPath := fmt.Sprintf("%s%s[%d]%s", attributePath, ".mutators", mutatorIndex, ".handler.config")
		return []validation.Failure{
			{AttributePath: attrPath, Message: "headers cannot be empty"},
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

	if config.Cookies == nil || len(config.Cookies) == 0 {
		attrPath := fmt.Sprintf("%s%s[%d]%s", attributePath, ".mutators", mutatorIndex, ".handler.config")
		return []validation.Failure{
			{AttributePath: attrPath, Message: "cookies cannot be empty"},
		}
	}

	return nil
}
