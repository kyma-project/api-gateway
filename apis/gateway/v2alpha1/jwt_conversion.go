package v2alpha1

import (
	"encoding/json"
	"errors"
	"github.com/kyma-project/api-gateway/apis/gateway/shared"
	"github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"github.com/kyma-project/api-gateway/internal/types/ory"
)

func convertOryJwtAccessStrategy(accessStrategy *v1beta1.Authenticator) (*shared.JwtConfig, error) {
	var oryJwtConfig ory.JWTAccStrConfig
	err := json.Unmarshal(accessStrategy.Config.Raw, &oryJwtConfig)
	if err != nil {
		return nil, err
	}

	// We can only reliably convert a single trusted issuer and JWKS URL to an Istio JwtConfig, as Istio JwtConfig does not support multiple issuers or JWKS URLs in a JwtAuthentication.
	// If an Ory JwtConfig has multiple trusted issuers or JWKS URLs the assumption of this function is, that this already has been checked earlier in the conversion.
	authentication := shared.JwtAuthentication{
		Issuer:  oryJwtConfig.TrustedIssuers[0],
		JwksUri: oryJwtConfig.JWKSUrls[0],
	}

	authorization := shared.JwtAuthorization{
		RequiredScopes: oryJwtConfig.RequiredScopes,
		Audiences:      oryJwtConfig.TargetAudience,
	}

	jwtConfig := &shared.JwtConfig{
		Authentications: []*shared.JwtAuthentication{&authentication},
		Authorizations:  []*shared.JwtAuthorization{&authorization},
	}

	return jwtConfig, nil
}

func convertIstioJwtAccessStrategy(accessStrategy *v1beta1.Authenticator) (*shared.JwtConfig, error) {
	var jwtConfig *shared.JwtConfig

	convertedFromObject := false
	if accessStrategy.Config.Object != nil {
		jwtConfig, convertedFromObject = accessStrategy.Config.Object.(*shared.JwtConfig)
	}

	if accessStrategy.Config.Raw != nil && !convertedFromObject {
		jwtConfig = &shared.JwtConfig{}
		err := json.Unmarshal(accessStrategy.Config.Raw, jwtConfig)
		if err != nil {
			return nil, err
		}
	}

	if jwtConfig == nil {
		return nil, errors.New("could not convert %v handler to v2alpha1.JwtConfig")
	}
	return jwtConfig, nil
}

func isConvertibleJwtConfig(accessStrategy *v1beta1.Authenticator) (bool, error) {

	if accessStrategy.Config == nil {
		return false, nil
	}

	istioJwtConfig, err := convertIstioJwtAccessStrategy(accessStrategy)
	if err != nil {
		return false, err
	}

	if len(istioJwtConfig.Authentications) > 0 {
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

func convertToJwtConfig(accessStrategy *v1beta1.Authenticator) (*shared.JwtConfig, error) {
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
