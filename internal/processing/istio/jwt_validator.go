package istio

import (
	"encoding/json"
	"fmt"

	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	istiojwt "github.com/kyma-incubator/api-gateway/internal/types/istio"
	"github.com/kyma-incubator/api-gateway/internal/validation"
)

type handlerValidator struct{}

func (o *handlerValidator) Validate(attributePath string, handler *gatewayv1beta1.Handler) []validation.Failure {
	var problems []validation.Failure
	var template istiojwt.JwtConfig

	if !validation.ConfigNotEmpty(handler.Config) {
		problems = append(problems, validation.Failure{AttributePath: attributePath + ".config", Message: "supplied config cannot be empty"})
		return problems
	}

	err := json.Unmarshal(handler.Config.Raw, &template)
	if err != nil {
		problems = append(problems, validation.Failure{AttributePath: attributePath + ".config", Message: "Can't read json: " + err.Error()})
		return problems
	}

	for i, auth := range template.Authentications {
		invalidIssuer, err := validation.IsInvalidURL(auth.Issuer)
		if invalidIssuer {
			attrPath := fmt.Sprintf("%s%s[%d]%s", attributePath, ".config.authentications", i, ".issuer")
			problems = append(problems, validation.Failure{AttributePath: attrPath, Message: fmt.Sprintf("value is empty or not a valid url err=%s", err)})
		}
		// The https:// configuration for TrustedIssuers is not necessary in terms of security best practices, 
		// however it is part of "secure by default" configuration, as this is the most common use case for iss claim.
	    // If we want to allow some weaker configurations, we should have a dedicated configuration which allows that.
		unsecuredIssuer, err := validation.IsUnsecuredURL(auth.Issuer)
		if unsecuredIssuer {
			attrPath := fmt.Sprintf("%s%s[%d]%s", attributePath, ".config.authentications", i, ".issuer")
			problems = append(problems, validation.Failure{AttributePath: attrPath, Message: fmt.Sprintf("value is not a secured url err=%s", err)})
		}
		invalidJwksUri, err := validation.IsInvalidURL(auth.JwksUri)
		if invalidJwksUri {
			attrPath := fmt.Sprintf("%s%s[%d]%s", attributePath, ".config.authentications", i, ".jwksUri")
			problems = append(problems, validation.Failure{AttributePath: attrPath, Message: fmt.Sprintf("value is empty or not a valid url err=%s", err)})
		}
		unsecuredJwksUri, err := validation.IsUnsecuredURL(auth.JwksUri)
		if unsecuredJwksUri {
			attrPath := fmt.Sprintf("%s%s[%d]%s", attributePath, ".config.authentications", i, ".jwksUri")
			problems = append(problems, validation.Failure{AttributePath: attrPath, Message: fmt.Sprintf("value is not a secured url err=%s", err)})
		}
	}

	return problems
}
