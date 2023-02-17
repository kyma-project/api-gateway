package validation

import (
	"errors"
	"fmt"
	"github.com/kyma-project/api-gateway/api/v1beta1"
)

func hasInvalidRequiredScopes(authorization v1beta1.JwtAuthorization) error {
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

func hasInvalidAudiences(authorization v1beta1.JwtAuthorization) error {
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

func HasInvalidAuthorizations(attributePath string, authorizations []*v1beta1.JwtAuthorization) (failures []Failure) {
	if authorizations == nil {
		return nil
	}
	if len(authorizations) == 0 {
		attrPath := fmt.Sprintf("%s%s", attributePath, ".config.authorizations")
		failures = append(failures, Failure{AttributePath: attrPath, Message: "value is empty"})
		return
	}

	for i, authorization := range authorizations {

		if authorization == nil {
			attrPath := fmt.Sprintf("%s%s[%d]", attributePath, ".config.authorizations", i)
			failures = append(failures, Failure{AttributePath: attrPath, Message: "authorization is empty"})
			continue
		}

		err := hasInvalidRequiredScopes(*authorization)
		if err != nil {
			attrPath := fmt.Sprintf("%s%s[%d]%s", attributePath, ".config.authorizations", i, ".requiredScopes")
			failures = append(failures, Failure{AttributePath: attrPath, Message: err.Error()})
		}

		err = hasInvalidAudiences(*authorization)
		if err != nil {
			attrPath := fmt.Sprintf("%s%s[%d]%s", attributePath, ".config.authorizations", i, ".audiences")
			failures = append(failures, Failure{AttributePath: attrPath, Message: err.Error()})
		}
	}

	return
}
