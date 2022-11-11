package validation

import (
	"encoding/json"
	"fmt"

	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	"github.com/kyma-incubator/api-gateway/internal/helpers"
)

// jwtAccStrValidator is an accessStrategy validator for jwt ORY authenticator
type jwtAccStrValidator struct{}

func (j *jwtAccStrValidator) Validate(attrPath string, handler *gatewayv1beta1.Handler, config *helpers.Config) []Failure {
	var problems []Failure

	var template gatewayv1beta1.JWTAccStrConfig

	//TODO validate based on config.JWTHandler

	if !configNotEmpty(handler.Config) {
		problems = append(problems, Failure{AttributePath: attrPath + ".config", Message: "supplied config cannot be empty"})
		return problems
	}
	err := json.Unmarshal(handler.Config.Raw, &template)
	if err != nil {
		problems = append(problems, Failure{AttributePath: attrPath + ".config", Message: "Can't read json: " + err.Error()})
		return problems
	}
	if len(template.TrustedIssuers) > 0 {
		for i := 0; i < len(template.TrustedIssuers); i++ {
			invalid, err := isInvalidURL(template.TrustedIssuers[i])
			if invalid {
				attrPath := fmt.Sprintf("%s[%d]", attrPath+".config.trusted_issuers", i)
				problems = append(problems, Failure{AttributePath: attrPath, Message: fmt.Sprintf("value is empty or not a valid url err=%s", err)})
			}
			unsecured, err := isUnsecuredURL(template.TrustedIssuers[i])
			if unsecured {
				attrPath := fmt.Sprintf("%s[%d]", attrPath+".config.trusted_issuers", i)
				problems = append(problems, Failure{AttributePath: attrPath, Message: fmt.Sprintf("value is not a secured url err=%s", err)})
			}
		}
	}

	if len(template.JWKSUrls) > 0 {
		for i := 0; i < len(template.JWKSUrls); i++ {
			invalid, err := isInvalidURL(template.JWKSUrls[i])
			if invalid {
				attrPath := fmt.Sprintf("%s[%d]", attrPath+".config.jwks_urls", i)
				problems = append(problems, Failure{AttributePath: attrPath, Message: fmt.Sprintf("value is empty or not a valid url err=%s", err)})
			}
			unsecured, err := isUnsecuredURL(template.JWKSUrls[i])
			if unsecured {
				attrPath := fmt.Sprintf("%s[%d]", attrPath+".config.jwks_urls", i)
				problems = append(problems, Failure{AttributePath: attrPath, Message: fmt.Sprintf("value is not a secured url err=%s", err)})
			}
		}
	}

	return problems
}
