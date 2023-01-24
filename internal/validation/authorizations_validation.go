package validation

import (
	"errors"
	"github.com/kyma-incubator/api-gateway/api/v1beta1"
)

func HasInvalidScopes(authorization v1beta1.JwtAuthorization) (bool, error) {
	if len(authorization.RequiredScopes) == 0 {
		return true, errors.New("value is empty")
	}
	for _, scope := range authorization.RequiredScopes {
		if scope == "" {
			return true, errors.New("scope value is empty")
		}
	}
	return false, nil
}
