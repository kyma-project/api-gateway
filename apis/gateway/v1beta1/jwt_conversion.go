package v1beta1

import (
	"encoding/json"
	"errors"
	"github.com/kyma-project/api-gateway/internal/types/ory"
)

func convertOryJwtAccessStrategy(accessStrategy *Authenticator) (*JwtConfig, error) {
	var oryJwtConfig ory.JWTAccStrConfig
	err := json.Unmarshal(accessStrategy.Config.Raw, &oryJwtConfig)
	if err != nil {
		return nil, err
	}

	// We can only reliably convert a single trusted issuer and JWKS URL to an Istio JwtConfig, as Istio JwtConfig does not support multiple issuers or JWKS URLs in a JwtAuthentication.
	// If an Ory JwtConfig has multiple trusted issuers or JWKS URLs the assumption of this function is, that this already has been checked earlier in the conversion.
	authentication := JwtAuthentication{
		Issuer:  oryJwtConfig.TrustedIssuers[0],
		JwksUri: oryJwtConfig.JWKSUrls[0],
	}

	authorization := JwtAuthorization{
		RequiredScopes: oryJwtConfig.RequiredScopes,
		Audiences:      oryJwtConfig.TargetAudience,
	}

	jwtConfig := &JwtConfig{
		Authentications: []*JwtAuthentication{&authentication},
		Authorizations:  []*JwtAuthorization{&authorization},
	}

	return jwtConfig, nil
}

func convertIstioJwtAccessStrategy(accessStrategy *Authenticator) (*JwtConfig, error) {

	if accessStrategy.Config.Object != nil {
		jwtConfig, ok := accessStrategy.Config.Object.(*JwtConfig)
		if ok {
			return jwtConfig, nil
		}
	}

	if accessStrategy.Config.Raw != nil {
		var jwtConfig JwtConfig
		err := json.Unmarshal(accessStrategy.Config.Raw, &jwtConfig)
		if err != nil {
			return nil, err
		}
		return &jwtConfig, nil
	}

	return nil, errors.New("no raw config to convert")
}

func isConvertibleJwtConfig(accessStrategy *Authenticator) (bool, error) {
	if accessStrategy.Config == nil {
		return false, nil
	}

	istioJwtConfig, err := convertIstioJwtAccessStrategy(accessStrategy)
	if err != nil {
		return false, err
	}

	if len(istioJwtConfig.Authentications) > 0 || len(istioJwtConfig.Authorizations) > 0 {
		return true, nil
	}

	var oryJwtConfig ory.JWTAccStrConfig
	err = json.Unmarshal(accessStrategy.Config.Raw, &oryJwtConfig)
	if err != nil {
		return false, err
	}

	// We only consider an Ory JwtConfig convertible if it has exactly one trusted issuer and one JWKS URL, since Istio JwtConfig
	// requires an issuer and a JWKS URI to be valid and does not support multiple issuers or JWKS URLs in a JwtAuthentication.
	if len(oryJwtConfig.TrustedIssuers) == 1 && len(oryJwtConfig.JWKSUrls) == 1 {
		return true, nil
	}

	return false, nil
}

func convertToJwtConfig(accessStrategy *Authenticator) (*JwtConfig, error) {
	jwtConfig, err := convertIstioJwtAccessStrategy(accessStrategy)
	if err != nil {
		return nil, err
	}

	if jwtConfig.Authentications == nil && jwtConfig.Authorizations == nil {
		// If the conversion to Istio JwtConfig failed, we try to convert to Ory JwtConfig
		jwtConfig, err = convertOryJwtAccessStrategy(accessStrategy)
		if err != nil {
			return nil, err
		}
	}
	return jwtConfig, nil
}
