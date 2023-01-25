package validation

import (
	"errors"
	"github.com/kyma-incubator/api-gateway/api/v1beta1"
)

func HasInvalidScopes(authorization v1beta1.JwtAuthorization) error {
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
