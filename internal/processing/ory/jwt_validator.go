package ory

import (
	"encoding/json"
	"fmt"

	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	"github.com/kyma-incubator/api-gateway/internal/types/ory"
	"github.com/kyma-incubator/api-gateway/internal/validation"
)

type handlerValidator struct{}

func (o *handlerValidator) Validate(attributePath string, handler *gatewayv1beta1.Handler) []validation.Failure {
	var problems []validation.Failure
	var template ory.JWTAccStrConfig

	if !validation.ConfigNotEmpty(handler.Config) {
		problems = append(problems, validation.Failure{AttributePath: attributePath + ".config", Message: "supplied config cannot be empty"})
		return problems
	}

	err := json.Unmarshal(handler.Config.Raw, &template)
	if err != nil {
		problems = append(problems, validation.Failure{AttributePath: attributePath + ".config", Message: "Can't read json: " + err.Error()})
		return problems
	}

	// The https:// configuration for TrustedIssuers is not necessary in terms of security best practices,
	// however it is part of "secure by default" configuration, as this is the most common use case for iss claim.
	// If we want to allow some weaker configurations, we should have a dedicated configuration which allows that.
	if len(template.TrustedIssuers) > 0 {
		for i := 0; i < len(template.TrustedIssuers); i++ {
			invalid, err := validation.IsInvalidURL(template.TrustedIssuers[i])
			if invalid {
				attrPath := fmt.Sprintf("%s[%d]", attributePath+".config.trusted_issuers", i)
				problems = append(problems, validation.Failure{AttributePath: attrPath, Message: fmt.Sprintf("value is empty or not a valid url err=%s", err)})
			}
			unsecured, err := validation.IsUnsecuredURL(template.TrustedIssuers[i])
			if unsecured {
				attrPath := fmt.Sprintf("%s[%d]", attributePath+".config.trusted_issuers", i)
				problems = append(problems, validation.Failure{AttributePath: attrPath, Message: fmt.Sprintf("value is not a secured url err=%s", err)})
			}
		}
	}

	if len(template.JWKSUrls) > 0 {
		for i := 0; i < len(template.JWKSUrls); i++ {
			invalid, err := validation.IsInvalidURL(template.JWKSUrls[i])
			if invalid {
				attrPath := fmt.Sprintf("%s[%d]", attributePath+".config.jwks_urls", i)
				problems = append(problems, validation.Failure{AttributePath: attrPath, Message: fmt.Sprintf("value is empty or not a valid url err=%s", err)})
			}
			unsecured, err := validation.IsUnsecuredURL(template.JWKSUrls[i])
			if unsecured {
				attrPath := fmt.Sprintf("%s[%d]", attributePath+".config.jwks_urls", i)
				problems = append(problems, validation.Failure{AttributePath: attrPath, Message: fmt.Sprintf("value is not a secured url err=%s", err)})
			}
		}
	}

	return problems
}
