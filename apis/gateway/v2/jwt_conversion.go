package v2

import (
	"encoding/json"
	"errors"
	"github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"github.com/kyma-project/api-gateway/internal/types/ory"
)

func convertOryJwtAccessStrategy(accessStrategy *v1beta1.Authenticator) (*v1beta1.JwtConfig, error) {
	var oryJwtConfig ory.JWTAccStrConfig
	err := json.Unmarshal(accessStrategy.Config.Raw, &oryJwtConfig)
	if err != nil {
		return nil, err
	}

	// We can only reliably convert a single trusted issuer and JWKS URL to an Istio JwtConfig, as Istio JwtConfig does not support multiple issuers or JWKS URLs in a JwtAuthentication.
	// If an Ory JwtConfig has multiple trusted issuers or JWKS URLs the assumption of this function is, that this already has been checked earlier in the conversion.
	authentication := v1beta1.JwtAuthentication{
		Issuer:  oryJwtConfig.TrustedIssuers[0],
		JwksUri: oryJwtConfig.JWKSUrls[0],
	}

	authorization := v1beta1.JwtAuthorization{
		RequiredScopes: oryJwtConfig.RequiredScopes,
		Audiences:      oryJwtConfig.TargetAudience,
	}

	jwtConfig := &v1beta1.JwtConfig{
		Authentications: []*v1beta1.JwtAuthentication{&authentication},
		Authorizations:  []*v1beta1.JwtAuthorization{&authorization},
	}

	return jwtConfig, nil
}

func convertIstioJwtAccessStrategy(accessStrategy *v1beta1.Authenticator) (*v1beta1.JwtConfig, error) {

	if accessStrategy.Config.Object != nil {
		jwtConfig, ok := accessStrategy.Config.Object.(*v1beta1.JwtConfig)
		if ok {
			return jwtConfig, nil
		}
	}

	if accessStrategy.Config.Raw != nil {
		var jwtConfig v1beta1.JwtConfig
		err := json.Unmarshal(accessStrategy.Config.Raw, &jwtConfig)
		if err != nil {
			return nil, err
		}
		return &jwtConfig, nil
	}

	return nil, errors.New("no raw config to convert")
}
