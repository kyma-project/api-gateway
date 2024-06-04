package v1beta2

import (
	"encoding/json"
	"errors"
	"github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"github.com/kyma-project/api-gateway/internal/types/ory"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
	"time"
)

var beta1to2statusConversionMap = map[v1beta1.StatusCode]State{
	v1beta1.StatusOK:      Ready,
	v1beta1.StatusError:   Error,
	v1beta1.StatusWarning: Warning,

	// StatusSkipped is not supported in v1beta2, and it happens only when another component has Error or Warning status
	// In this case, we map it to Warning
	v1beta1.StatusSkipped: Warning,
}

func convertMap(m map[v1beta1.StatusCode]State) map[State]v1beta1.StatusCode {
	inv := make(map[State]v1beta1.StatusCode)
	for k, v := range m {
		inv[v] = k
	}
	return inv
}

// The 2 => 1 map is generated automatically based on 1 => 2 map
var beta2to1statusConversionMap = convertMap(beta1to2statusConversionMap)

// Converts this ApiRule (v1beta2) to the Hub version (v1beta1)
func (apiRuleBeta2 *APIRule) ConvertTo(hub conversion.Hub) error {
	apiRuleBeta1 := hub.(*v1beta1.APIRule)

	apiRuleBeta1.ObjectMeta = apiRuleBeta2.ObjectMeta
	if apiRuleBeta1.Annotations == nil {
		apiRuleBeta1.Annotations = make(map[string]string)
	}
	apiRuleBeta1.Annotations["gateway.kyma-project.io/original-version"] = "v1beta2"

	err := convertOverJson(apiRuleBeta2.Spec.Rules, &apiRuleBeta1.Spec.Rules)
	if err != nil {
		return err
	}

	err = convertOverJson(apiRuleBeta2.Spec.Gateway, &apiRuleBeta1.Spec.Gateway)
	if err != nil {
		return err
	}
	err = convertOverJson(apiRuleBeta2.Spec.Service, &apiRuleBeta1.Spec.Service)
	if err != nil {
		return err
	}
	err = convertOverJson(apiRuleBeta2.Spec.Timeout, &apiRuleBeta1.Spec.Timeout)
	if err != nil {
		return err
	}

	// Status
	apiRuleBeta1.Status = v1beta1.APIRuleStatus{
		APIRuleStatus: &v1beta1.APIRuleResourceStatus{
			Code:        beta2to1statusConversionMap[apiRuleBeta2.Status.State],
			Description: apiRuleBeta2.Status.Description,
		},
		LastProcessedTime: apiRuleBeta2.Status.LastProcessedTime,
	}

	if apiRuleBeta2.Spec.CorsPolicy != nil {
		apiRuleBeta1.Spec.CorsPolicy = &v1beta1.CorsPolicy{}
		apiRuleBeta1.Spec.CorsPolicy.AllowHeaders = apiRuleBeta2.Spec.CorsPolicy.AllowHeaders
		apiRuleBeta1.Spec.CorsPolicy.AllowMethods = apiRuleBeta2.Spec.CorsPolicy.AllowMethods
		apiRuleBeta1.Spec.CorsPolicy.AllowOrigins = v1beta1.StringMatch(apiRuleBeta2.Spec.CorsPolicy.AllowOrigins)
		apiRuleBeta1.Spec.CorsPolicy.AllowCredentials = apiRuleBeta2.Spec.CorsPolicy.AllowCredentials
		apiRuleBeta1.Spec.CorsPolicy.ExposeHeaders = apiRuleBeta2.Spec.CorsPolicy.ExposeHeaders
		apiRuleBeta1.Spec.CorsPolicy.MaxAge = &metav1.Duration{Duration: time.Duration(apiRuleBeta2.Spec.CorsPolicy.MaxAge) * time.Second}
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

	conversionPossible, err := isFullConversionPossible(apiRuleBeta1)
	if err != nil {
		return err
	}
	if !conversionPossible {
		// We have to stop the conversion here, because we want to return an empty Spec in case we cannot fully convert the APIRule.
		return nil
	}

	err = convertOverJson(apiRuleBeta1.Spec.Rules, &apiRuleBeta2.Spec.Rules)
	if err != nil {
		return err
	}
	err = convertOverJson(apiRuleBeta1.Spec.Gateway, &apiRuleBeta2.Spec.Gateway)
	if err != nil {
		return err
	}
	err = convertOverJson(apiRuleBeta1.Spec.Service, &apiRuleBeta2.Spec.Service)
	if err != nil {
		return err
	}
	err = convertOverJson(apiRuleBeta1.Spec.Timeout, &apiRuleBeta2.Spec.Timeout)
	if err != nil {
		return err
	}

	if apiRuleBeta1.Status.APIRuleStatus != nil {
		apiRuleBeta2.Status = APIRuleStatus{
			State:             beta1to2statusConversionMap[apiRuleBeta1.Status.APIRuleStatus.Code],
			Description:       apiRuleBeta1.Status.APIRuleStatus.Description,
			LastProcessedTime: apiRuleBeta1.Status.LastProcessedTime,
		}
	}

	apiRuleBeta2.Spec.Hosts = []*Host{new(Host)}
	*apiRuleBeta2.Spec.Hosts[0] = Host(*apiRuleBeta1.Spec.Host)
	if apiRuleBeta1.Spec.CorsPolicy != nil {
		apiRuleBeta2.Spec.CorsPolicy = &CorsPolicy{}
		apiRuleBeta2.Spec.CorsPolicy.AllowHeaders = apiRuleBeta1.Spec.CorsPolicy.AllowHeaders
		apiRuleBeta2.Spec.CorsPolicy.AllowMethods = apiRuleBeta1.Spec.CorsPolicy.AllowMethods
		apiRuleBeta2.Spec.CorsPolicy.AllowOrigins = StringMatch(apiRuleBeta1.Spec.CorsPolicy.AllowOrigins)
		apiRuleBeta2.Spec.CorsPolicy.AllowCredentials = apiRuleBeta1.Spec.CorsPolicy.AllowCredentials
		apiRuleBeta2.Spec.CorsPolicy.ExposeHeaders = apiRuleBeta1.Spec.CorsPolicy.ExposeHeaders

		// metav1.Duration type for seconds is float64,
		// however the Access-Control-Max-Age header is specified in seconds without decimals.
		// In consequence, the conversion drops any values smaller than 1 second.
		// https://fetch.spec.whatwg.org/#http-responses
		apiRuleBeta2.Spec.CorsPolicy.MaxAge = uint64(apiRuleBeta1.Spec.CorsPolicy.MaxAge.Duration.Seconds())
	}

	apiRuleBeta2.Spec.Rules = []Rule{}
	for _, ruleBeta1 := range apiRuleBeta1.Spec.Rules {
		ruleBeta2 := Rule{}
		err = convertOverJson(ruleBeta1, &ruleBeta2)
		if err != nil {
			return err
		}
		for _, accessStrategy := range ruleBeta1.AccessStrategies {
			if accessStrategy.Handler.Name == "no_auth" {
				ruleBeta2.NoAuth = ptr.To(true)
			}

			if accessStrategy.Handler.Name == "jwt" && accessStrategy.Config != nil {
				jwtConfig, err := convertIstioJwtAccessStrategy(accessStrategy)
				if err != nil {
					return err
				}

				if jwtConfig.Authentications == nil && jwtConfig.Authorizations == nil {
					// If the conversion to Istio JwtConfig failed, we try to convert to Ory JwtConfig
					jwtConfig, err = convertOryJwtAccessStrategy(accessStrategy)
					if err != nil {
						return err
					}
				}

				err = convertOverJson(jwtConfig, &ruleBeta2.Jwt)
				if err != nil {
					return err
				}
			}
		}
		apiRuleBeta2.Spec.Rules = append(apiRuleBeta2.Spec.Rules, ruleBeta2)
	}

	return nil
}

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
	var jwtConfig *v1beta1.JwtConfig

	if accessStrategy.Config.Object != nil {
		jwtConfig = accessStrategy.Config.Object.(*v1beta1.JwtConfig)
	} else if accessStrategy.Config.Raw != nil {
		jwtConfig = &v1beta1.JwtConfig{}
		err := json.Unmarshal(accessStrategy.Config.Raw, jwtConfig)
		if err != nil {
			return nil, err
		}
	}

	return jwtConfig, nil
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

// isFullConversionPossible checks if the APIRule can be fully converted to v1beta2 by evaluating the access strategies.
func isFullConversionPossible(apiRule *v1beta1.APIRule) (bool, error) {
	for _, rule := range apiRule.Spec.Rules {
		for _, accessStrategy := range rule.AccessStrategies {

			if accessStrategy.Handler.Name == "no_auth" {
				continue
			}

			if accessStrategy.Handler.Name == "jwt" {
				isConvertible, err := isConvertibleJwtConfig(accessStrategy)
				if err != nil {
					return false, err
				}
				if isConvertible {
					continue
				}
			}

			return false, nil
		}

	}

	return true, nil
}

func isConvertibleJwtConfig(accessStrategy *v1beta1.Authenticator) (bool, error) {
	istioJwtConfig, err := convertIstioJwtAccessStrategy(accessStrategy)
	if err != nil {
		return false, err
	}

	if istioJwtConfig.Authentications != nil || istioJwtConfig.Authorizations != nil {
		return true, nil
	}

	var oryJwtConfig ory.JWTAccStrConfig
	err = json.Unmarshal(accessStrategy.Config.Raw, &oryJwtConfig)
	if err != nil {
		return false, err
	}
	if len(oryJwtConfig.TrustedIssuers) == 1 && len(oryJwtConfig.JWKSUrls) == 1 {
		// We can only reliably convert a single trusted issuer and JWKS URL to an Istio JwtConfig, as Istio JwtConfig does not support multiple issuers or JWKS URLs in a JwtAuthentication.
		return true, nil
	}

	return false, nil
}
