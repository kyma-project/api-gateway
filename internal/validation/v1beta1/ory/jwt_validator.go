package ory

import (
	"encoding/json"
	"fmt"
	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"

	"github.com/kyma-project/api-gateway/internal/types/ory"
	"github.com/kyma-project/api-gateway/internal/validation"
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

	problems = append(problems, checkForIstioConfig(attributePath, handler)...)

	// The https:// configuration for TrustedIssuers is not necessary in terms of security best practices,
	// however, it is part of "secure by default" configuration, as this is the most common use case for iss claim.
	// If we want to allow some weaker configurations, we should have a dedicated configuration which allows that.
	if len(template.TrustedIssuers) > 0 {
		for i := 0; i < len(template.TrustedIssuers); i++ {
			invalid, err := validation.IsInvalidURI(template.TrustedIssuers[i])
			if invalid {
				attrPath := fmt.Sprintf("%s[%d]", attributePath+".config.trusted_issuers", i)
				problems = append(problems, validation.Failure{AttributePath: attrPath, Message: fmt.Sprintf("value is empty or not a valid uri err=%s", err)})
			}
		}
	}

	if len(template.JWKSUrls) > 0 {
		for i := 0; i < len(template.JWKSUrls); i++ {
			invalid, err := validation.IsInvalidURI(template.JWKSUrls[i])
			if invalid {
				attrPath := fmt.Sprintf("%s[%d]", attributePath+".config.jwks_urls", i)
				problems = append(problems, validation.Failure{AttributePath: attrPath, Message: fmt.Sprintf("value is empty or not a valid uri err=%s", err)})
			}
		}
	}

	return problems
}

func checkForIstioConfig(attributePath string, handler *gatewayv1beta1.Handler) (problems []validation.Failure) {
	var template gatewayv1beta1.JwtConfig
	err := json.Unmarshal(handler.Config.Raw, &template)
	if err != nil {
		return []validation.Failure{{AttributePath: attributePath + ".config", Message: "Can't read json: " + err.Error()}}
	}

	if len(template.Authentications) > 0 {
		return []validation.Failure{{AttributePath: attributePath + ".config" + ".authentications", Message: "Configuration for authentications is not supported with Ory handler"}}
	}

	return problems
}
