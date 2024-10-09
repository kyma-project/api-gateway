package v2alpha1

import (
	"errors"
	"fmt"
	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	"strings"

	"github.com/kyma-project/api-gateway/internal/validation"
)

func validateJwt(parentAttributePath string, rule *gatewayv2alpha1.Rule) []validation.Failure {
	var failures []validation.Failure

	if rule.Jwt != nil {
		jwtAttributePath := parentAttributePath + ".jwt"

		failures = append(failures, hasInvalidAuthorizations(jwtAttributePath, rule.Jwt.Authorizations)...)
		failures = append(failures, hasInvalidAuthentications(jwtAttributePath, rule.Jwt.Authentications)...)
	} else if rule.ExtAuth != nil && rule.ExtAuth.Restrictions != nil {
		extAuthAttributePath := parentAttributePath + ".extAuth"

		failures = append(failures, hasInvalidAuthorizations(extAuthAttributePath, rule.ExtAuth.Restrictions.Authorizations)...)
		failures = append(failures, hasInvalidAuthentications(extAuthAttributePath, rule.ExtAuth.Restrictions.Authentications)...)
	}

	return failures
}

func hasInvalidRequiredScopes(authorization gatewayv2alpha1.JwtAuthorization) error {
	if authorization.RequiredScopes == nil {
		return nil
	}
	if len(authorization.RequiredScopes) == 0 {
		return errors.New("value is empty")
	}
	for _, scope := range authorization.RequiredScopes {
		if scope == "" {
			return errors.New("scope value is empty")
		}
	}
	return nil
}

func hasInvalidAudiences(authorization gatewayv2alpha1.JwtAuthorization) error {
	if authorization.Audiences == nil {
		return nil
	}
	if len(authorization.Audiences) == 0 {
		return errors.New("value is empty")
	}
	for _, audience := range authorization.Audiences {
		if audience == "" {
			return errors.New("audience value is empty")
		}
	}
	return nil
}

func hasInvalidAuthentications(parentAttributePath string, authentications []*gatewayv2alpha1.JwtAuthentication) []validation.Failure {
	var failures []validation.Failure
	authenticationsAttrPath := parentAttributePath + ".authentications"

	hasFromHeaders, hasFromParams := false, false
	if len(authentications) == 0 {
		return []validation.Failure{
			{
				AttributePath: authenticationsAttrPath,
				Message:       "A JWT config must have at least one authentication",
			},
		}
	}
	for i, authentication := range authentications {
		issuerErr := validateJwtIssuer(authentication.Issuer)
		if issuerErr != nil {
			attrPath := fmt.Sprintf("%s[%d]%s", authenticationsAttrPath, i, ".issuer")
			failures = append(failures, validation.Failure{AttributePath: attrPath, Message: fmt.Sprintf("value is empty or not a valid uri err=%s", issuerErr)})
		}

		invalidJwksUri, err := validation.IsInvalidURI(authentication.JwksUri)
		if invalidJwksUri {
			attrPath := fmt.Sprintf("%s[%d]%s", authenticationsAttrPath, i, ".jwksUri")
			failures = append(failures, validation.Failure{AttributePath: attrPath, Message: fmt.Sprintf("value is empty or not a valid uri err=%s", err)})
		}
		if len(authentication.FromHeaders) > 0 {
			if hasFromParams {
				attrPath := fmt.Sprintf("%s[%d]%s", authenticationsAttrPath, i, ".fromHeaders")
				failures = append(failures, validation.Failure{AttributePath: attrPath, Message: "mixture of multiple fromHeaders and fromParams is not supported"})
			}
			hasFromHeaders = true
		}
		if len(authentication.FromParams) > 0 {
			if hasFromHeaders {
				attrPath := fmt.Sprintf("%s[%d]%s", authenticationsAttrPath, i, ".fromParams")
				failures = append(failures, validation.Failure{AttributePath: attrPath, Message: "mixture of multiple fromHeaders and fromParams is not supported"})
			}
			hasFromParams = true
		}
		if len(authentication.FromHeaders) > 1 {
			attrPath := fmt.Sprintf("%s[%d]%s", authenticationsAttrPath, i, ".fromHeaders")
			failures = append(failures, validation.Failure{AttributePath: attrPath, Message: "multiple fromHeaders are not supported"})
		}
		if len(authentication.FromParams) > 1 {
			attrPath := fmt.Sprintf("%s[%d]%s", authenticationsAttrPath, i, ".fromParams")
			failures = append(failures, validation.Failure{AttributePath: attrPath, Message: "multiple fromParams are not supported"})
		}
	}
	return failures
}

func hasInvalidAuthorizations(parentAttributePath string, authorizations []*gatewayv2alpha1.JwtAuthorization) []validation.Failure {
	var failures []validation.Failure
	authorizationsAttrPath := parentAttributePath + ".authorizations"

	if authorizations == nil {
		return nil
	}
	if len(authorizations) == 0 {
		return append(failures, validation.Failure{AttributePath: authorizationsAttrPath, Message: "authorizations defined, but no configuration exists"})
	}

	for i, authorization := range authorizations {
		if authorization == nil {
			attrPath := fmt.Sprintf("%s[%d]", authorizationsAttrPath, i)
			failures = append(failures, validation.Failure{AttributePath: attrPath, Message: "authorization is empty"})
			continue
		}

		err := hasInvalidRequiredScopes(*authorization)
		if err != nil {
			attrPath := fmt.Sprintf("%s[%d]%s", authorizationsAttrPath, i, ".requiredScopes")
			failures = append(failures, validation.Failure{AttributePath: attrPath, Message: err.Error()})
		}

		err = hasInvalidAudiences(*authorization)
		if err != nil {
			attrPath := fmt.Sprintf("%s[%d]%s", authorizationsAttrPath, i, ".audiences")
			failures = append(failures, validation.Failure{AttributePath: attrPath, Message: err.Error()})
		}
	}

	return failures
}

// validateJwtAuthenticationEquality validates that all JWT authorizations with the same issuer and JWKS URI have the
// same configuration
func validateJwtAuthenticationEquality(parentAttributePath string, rules []gatewayv2alpha1.Rule) []validation.Failure {
	var failures []validation.Failure
	jwtAuths := map[string]*gatewayv2alpha1.JwtAuthentication{}

	for ruleIndex, rule := range rules {

		if rule.Jwt == nil {
			continue
		}

		for authenticationIndex, authentication := range rule.Jwt.Authentications {
			authAttributePath := fmt.Sprintf("%s[%d].jwt.authentications[%d]", parentAttributePath, ruleIndex, authenticationIndex)

			jwtAuthKey := authentication.Issuer + authentication.JwksUri
			if jwtAuths[jwtAuthKey] != nil && !jwtAuthenticationsEqual(authentication, jwtAuths[jwtAuthKey]) {
				failures = append(failures, validation.Failure{AttributePath: authAttributePath, Message: "multiple jwt configurations that differ for the same issuer"})
			} else {
				jwtAuths[jwtAuthKey] = authentication
			}
		}
	}
	return failures
}

func jwtAuthenticationsEqual(auth1 *gatewayv2alpha1.JwtAuthentication, auth2 *gatewayv2alpha1.JwtAuthentication) bool {
	if auth1.Issuer != auth2.Issuer || auth1.JwksUri != auth2.JwksUri {
		return false
	}
	if len(auth1.FromHeaders) != len(auth2.FromHeaders) {
		return false
	}
	for i, auth1FromHeader := range auth1.FromHeaders {
		if auth1FromHeader.Name != auth2.FromHeaders[i].Name || auth1FromHeader.Prefix != auth2.FromHeaders[i].Prefix {
			return false
		}
	}
	if len(auth1.FromParams) != len(auth2.FromParams) {
		return false
	}
	for i, auth1FromParam := range auth1.FromParams {
		if auth1FromParam != auth2.FromParams[i] {
			return false
		}
	}
	return true
}

func validateJwtIssuer(issuer string) error {
	if issuer == "" {
		return errors.New("value is empty")
	}

	// If issuer contains ':' it must be a valid URI, see https://datatracker.ietf.org/doc/html/rfc7519#section-4.1.1
	if strings.Contains(issuer, ":") {
		if isInvalid, err := validation.IsInvalidURI(issuer); isInvalid {
			return err
		}
	}

	return nil
}
