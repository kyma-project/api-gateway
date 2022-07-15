package validation

import (
	"encoding/json"
	"fmt"

	gatewayv1alpha1 "github.com/kyma-incubator/api-gateway/api/v1alpha1"
)

//jwtAccStrValidator is an accessStrategy validator for jwt ORY authenticator
type jwtAccStrValidator struct{}

func (j *jwtAccStrValidator) Validate(attributePath string, handler *gatewayv1alpha1.Handler) []Failure {
	var problems []Failure

	var template gatewayv1alpha1.JWTAccStrConfig

	if !configNotEmpty(handler.Config) {
		problems = append(problems, Failure{AttributePath: attributePath + ".config", Message: "supplied config cannot be empty"})
		return problems
	}
	err := json.Unmarshal(handler.Config.Raw, &template)
	if err != nil {
		problems = append(problems, Failure{AttributePath: attributePath + ".config", Message: "Can't read json: " + err.Error()})
		return problems
	}
	if len(template.TrustedIssuers) > 0 {
		for i := 0; i < len(template.TrustedIssuers); i++ {
			if isInvalidURL(template.TrustedIssuers[i]) {
				attrPath := fmt.Sprintf("%s[%d]", attributePath+".config.trusted_issuers", i)
				problems = append(problems, Failure{AttributePath: attrPath, Message: "value is empty or not a valid url"})
			}
			if isUnsecuredURL(template.TrustedIssuers[i]) {
				attrPath := fmt.Sprintf("%s[%d]", attributePath+".config.trusted_issuers", i)
				problems = append(problems, Failure{AttributePath: attrPath, Message: "value is not a secured url"})
			}
		}
	}

	if len(template.JWKSUrls) > 0 {
		for i := 0; i < len(template.JWKSUrls); i++ {
			if isInvalidURL(template.JWKSUrls[i]) {
				attrPath := fmt.Sprintf("%s[%d]", attributePath+".config.jwks_urls", i)
				problems = append(problems, Failure{AttributePath: attrPath, Message: "value is empty or not a valid url"})
			}
			if isUnsecuredURL(template.JWKSUrls[i]) {
				attrPath := fmt.Sprintf("%s[%d]", attributePath+".config.jwks_urls", i)
				problems = append(problems, Failure{AttributePath: attrPath, Message: "value is not a secured url"})
			}
		}
	}

	return problems
}
