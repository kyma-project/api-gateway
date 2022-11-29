package validation

import (
	"encoding/json"
	"fmt"

	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	"github.com/kyma-incubator/api-gateway/internal/helpers"
	istioint "github.com/kyma-incubator/api-gateway/internal/types/istio"
	oryint "github.com/kyma-incubator/api-gateway/internal/types/ory"
)

// jwtAccStrValidator is an accessStrategy validator for jwt ORY authenticator
type jwtAccStrValidator struct{}

func (j *jwtAccStrValidator) Validate(attrPath string, handler *gatewayv1beta1.Handler, config *helpers.Config) []Failure {
	var problems []Failure

	if !ConfigNotEmpty(handler.Config) {
		problems = append(problems, Failure{AttributePath: attrPath + ".config", Message: "supplied config cannot be empty"})
		return problems
	}

	switch config.JWTHandler {
	case helpers.JWT_HANDLER_ORY:
		problems = j.validateOryConfig(attrPath, handler)
	case helpers.JWT_HANDLER_ISTIO:
		problems = j.validateIstioConfig(attrPath, handler)
	}

	return problems
}

func (j *jwtAccStrValidator) validateOryConfig(attrPath string, handler *gatewayv1beta1.Handler) []Failure {
	var problems []Failure
	var template oryint.JWTAccStrConfig

	err := json.Unmarshal(handler.Config.Raw, &template)
	if err != nil {
		problems = append(problems, Failure{AttributePath: attrPath + ".config", Message: "Can't read json: " + err.Error()})
		return problems
	}

	if len(template.TrustedIssuers) > 0 {
		for i := 0; i < len(template.TrustedIssuers); i++ {
			invalid, err := IsInvalidURL(template.TrustedIssuers[i])
			if invalid {
				attrPath := fmt.Sprintf("%s[%d]", attrPath+".config.trusted_issuers", i)
				problems = append(problems, Failure{AttributePath: attrPath, Message: fmt.Sprintf("value is empty or not a valid url err=%s", err)})
			}
			unsecured, err := IsUnsecuredURL(template.TrustedIssuers[i])
			if unsecured {
				attrPath := fmt.Sprintf("%s[%d]", attrPath+".config.trusted_issuers", i)
				problems = append(problems, Failure{AttributePath: attrPath, Message: fmt.Sprintf("value is not a secured url err=%s", err)})
			}
		}
	}

	if len(template.JWKSUrls) > 0 {
		for i := 0; i < len(template.JWKSUrls); i++ {
			invalid, err := IsInvalidURL(template.JWKSUrls[i])
			if invalid {
				attrPath := fmt.Sprintf("%s[%d]", attrPath+".config.jwks_urls", i)
				problems = append(problems, Failure{AttributePath: attrPath, Message: fmt.Sprintf("value is empty or not a valid url err=%s", err)})
			}
			unsecured, err := IsUnsecuredURL(template.JWKSUrls[i])
			if unsecured {
				attrPath := fmt.Sprintf("%s[%d]", attrPath+".config.jwks_urls", i)
				problems = append(problems, Failure{AttributePath: attrPath, Message: fmt.Sprintf("value is not a secured url err=%s", err)})
			}
		}
	}

	return problems
}

func (j *jwtAccStrValidator) validateIstioConfig(attrPath string, handler *gatewayv1beta1.Handler) []Failure {
	var problems []Failure
	var template istioint.JwtConfig

	err := json.Unmarshal(handler.Config.Raw, &template)
	if err != nil {
		problems = append(problems, Failure{AttributePath: attrPath + ".config", Message: "Can't read json: " + err.Error()})
		return problems
	}

	for i, auth := range template.Authentications {
		invalidIssuer, err := IsInvalidURL(auth.Issuer)
		if invalidIssuer {
			attrPath := fmt.Sprintf("%s%s[%d]%s", attrPath, ".config.authentications", i, ".issuer")
			problems = append(problems, Failure{AttributePath: attrPath, Message: fmt.Sprintf("value is empty or not a valid url err=%s", err)})
		}
		unsecuredIssuer, err := IsUnsecuredURL(auth.Issuer)
		if unsecuredIssuer {
			attrPath := fmt.Sprintf("%s%s[%d]%s", attrPath, ".config.authentications", i, ".issuer")
			problems = append(problems, Failure{AttributePath: attrPath, Message: fmt.Sprintf("value is not a secured url err=%s", err)})
		}
		invalidJwksUri, err := IsInvalidURL(auth.JwksUri)
		if invalidJwksUri {
			attrPath := fmt.Sprintf("%s%s[%d]%s", attrPath, ".config.authentications", i, ".jwksUri")
			problems = append(problems, Failure{AttributePath: attrPath, Message: fmt.Sprintf("value is empty or not a valid url err=%s", err)})
		}
		unsecuredJwksUri, err := IsUnsecuredURL(auth.JwksUri)
		if unsecuredJwksUri {
			attrPath := fmt.Sprintf("%s%s[%d]%s", attrPath, ".config.authentications", i, ".jwksUri")
			problems = append(problems, Failure{AttributePath: attrPath, Message: fmt.Sprintf("value is not a secured url err=%s", err)})
		}
	}

	return problems
}
