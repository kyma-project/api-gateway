package v1beta2

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"github.com/kyma-project/api-gateway/internal/types/ory"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

// Converts this ApiRule (v1beta2) to the Hub version (v1beta1)
func (apiRuleBeta2 *APIRule) ConvertTo(hub conversion.Hub) error {
	apiRuleBeta1 := hub.(*v1beta1.APIRule)

	apiRuleBeta1.ObjectMeta = apiRuleBeta2.ObjectMeta
	if apiRuleBeta1.Annotations == nil {
		apiRuleBeta1.Annotations = make(map[string]string)
	}
	apiRuleBeta1.Annotations["gateway.kyma-project.io/original-version"] = "v1beta2"

	err := convertOverJson(apiRuleBeta2.Spec, &apiRuleBeta1.Spec)
	if err != nil {
		return err
	}
	err = convertOverJson(apiRuleBeta2.Status, &apiRuleBeta1.Status)
	if err != nil {
		return err
	}

	// Only one host is supported in v1beta1, so we use the first one from the list
	strHost := string(*apiRuleBeta2.Spec.Hosts[0])
	apiRuleBeta1.Spec.Host = &strHost

	apiRuleBeta1.Spec.Rules = []v1beta1.Rule{}
	for _, ruleBeta2 := range apiRuleBeta2.Spec.Rules {
		ruleBeta1 := v1beta1.Rule{}
		err = convertOverJson(ruleBeta2, &ruleBeta1)
		if err != nil {
			return err
		}
		// No Auth
		if ruleBeta2.NoAuth != nil && *ruleBeta2.NoAuth {
			ruleBeta1.AccessStrategies = append(ruleBeta1.AccessStrategies, &v1beta1.Authenticator{
				Handler: &v1beta1.Handler{
					Name: "no_auth",
				},
			})
		}
		// JWT
		if ruleBeta2.Jwt != nil {
			ruleBeta1.AccessStrategies = append(ruleBeta1.AccessStrategies, &v1beta1.Authenticator{
				Handler: &v1beta1.Handler{
					Name:   "jwt",
					Config: &runtime.RawExtension{Object: ruleBeta2.Jwt},
				},
			})
		}
		if len(ruleBeta1.AccessStrategies) == 0 {
			return errors.New("either jwt is configured or noAuth must be set to true in a rule")
		}
		apiRuleBeta1.Spec.Rules = append(apiRuleBeta1.Spec.Rules, ruleBeta1)
	}

	return nil
}

// Converts from the Hub version (v1beta1) into this ApiRule (v1beta2)
func (apiRuleBeta2 *APIRule) ConvertFrom(hub conversion.Hub) error {
	apiRuleBeta1 := hub.(*v1beta1.APIRule)

	apiRuleBeta2.ObjectMeta = apiRuleBeta1.ObjectMeta

	err := convertOverJson(apiRuleBeta1.Spec, &apiRuleBeta2.Spec)
	if err != nil {
		return err
	}
	err = convertOverJson(apiRuleBeta1.Status, &apiRuleBeta2.Status)
	if err != nil {
		return err
	}

	apiRuleBeta2.Spec.Hosts = []*Host{new(Host)}
	*apiRuleBeta2.Spec.Hosts[0] = Host(*apiRuleBeta1.Spec.Host)

	apiRuleBeta2.Spec.Rules = []Rule{}
	for _, ruleBeta1 := range apiRuleBeta1.Spec.Rules {
		ruleBeta2 := Rule{}
		err = convertOverJson(ruleBeta1, &ruleBeta2)
		if err != nil {
			return err
		}
		for _, accessStrategy := range ruleBeta1.AccessStrategies {
			// No Auth
			if accessStrategy.Handler.Name == "no_auth" {
				ruleBeta2.NoAuth = ptr.To(true)
			}
			// JWT
			if accessStrategy.Handler.Name == "jwt" && accessStrategy.Config != nil {
				var jwtConfig *v1beta1.JwtConfig
				if accessStrategy.Config.Object != nil {
					jwtConfig = accessStrategy.Config.Object.(*v1beta1.JwtConfig)
				} else if accessStrategy.Config.Raw != nil {
					jwtConfig = &v1beta1.JwtConfig{}
					err = json.Unmarshal(accessStrategy.Config.Raw, jwtConfig)
					if err != nil {
						return err
					}
				}
				if jwtConfig.Authentications == nil && jwtConfig.Authorizations == nil {
					// best effort to convert ory jwt v1beta1 to v1beta2
					var oryJwtConfig ory.JWTAccStrConfig
					_ = json.Unmarshal(accessStrategy.Config.Raw, &oryJwtConfig)
					if len(oryJwtConfig.JWKSUrls) > 0 && len(oryJwtConfig.JWKSUrls) == len(oryJwtConfig.TrustedIssuers) {
						for index, jwksUrl := range oryJwtConfig.JWKSUrls {
							jwtAuthentication := v1beta1.JwtAuthentication{}
							jwtAuthentication.JwksUri = jwksUrl
							jwtAuthentication.Issuer = oryJwtConfig.TrustedIssuers[index]
							jwtConfig.Authentications = append(jwtConfig.Authentications, &jwtAuthentication)
						}
						if len(oryJwtConfig.RequiredScopes) > 0 || len(oryJwtConfig.TargetAudience) > 0 {
							jwtAuthorization := v1beta1.JwtAuthorization{}
							jwtAuthorization.RequiredScopes = append(jwtAuthorization.RequiredScopes, oryJwtConfig.RequiredScopes...)
							jwtAuthorization.Audiences = append(jwtAuthorization.Audiences, oryJwtConfig.TargetAudience...)
							jwtConfig.Authorizations = append(jwtConfig.Authorizations, &jwtAuthorization)
						}
					}
				}
				err = convertOverJson(jwtConfig, &ruleBeta2.Jwt)
				if err != nil {
					return err
				}
			}
			// OAuth2
			if strings.Contains(accessStrategy.Handler.Name, "oauth2") {
				return errors.New("oauth2 access strategy is not supported in v1beta2, please migrate to v1beta2 or request it in v1beta1 version")
			}
		}
		apiRuleBeta2.Spec.Rules = append(apiRuleBeta2.Spec.Rules, ruleBeta2)
	}

	return nil
}

func convertOverJson(src any, dst any) error {
	data, err := json.Marshal(src)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, dst)
	if err != nil {
		return err
	}

	return nil
}
