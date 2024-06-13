package v2alpha1

import (
	"encoding/json"
	"errors"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"

	"github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"github.com/kyma-project/api-gateway/internal/types/ory"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

const (
	v1beta1DeprecatedTemplate = "APIRule in version v1beta1 has been deprecated. To request APIRule v1beta1, use the command 'kubectl get -n %s apirules.v1beta1.gateway.kyma-project.io %s'. See APIRule v2alpha1 documentation and consider migrating to the newer version."
)

var beta1to2statusConversionMap = map[v1beta1.StatusCode]State{
	v1beta1.StatusOK:      Ready,
	v1beta1.StatusError:   Error,
	v1beta1.StatusWarning: Warning,

	// StatusSkipped is not supported in v2alpha1, and it happens only when another component has Error or Warning status
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
var 2alpha1to1statusConversionMap = convertMap(beta1to2statusConversionMap)

// Converts this ApiRule (v2alpha1) to the Hub version (v1beta1)
func (apiRule2Alpha1 *APIRule) ConvertTo(hub conversion.Hub) error {
	apiRuleBeta1 := hub.(*v1beta1.APIRule)

	apiRuleBeta1.ObjectMeta = apiRule2Alpha1.ObjectMeta
	if apiRuleBeta1.Annotations == nil {
		apiRuleBeta1.Annotations = make(map[string]string)
	}
	apiRuleBeta1.Annotations["gateway.kyma-project.io/original-version"] = "v2alpha1"

	err := convertOverJson(apiRule2Alpha1.Spec.Rules, &apiRuleBeta1.Spec.Rules)
	if err != nil {
		return err
	}

	err = convertOverJson(apiRule2Alpha1.Spec.Gateway, &apiRuleBeta1.Spec.Gateway)
	if err != nil {
		return err
	}
	err = convertOverJson(apiRule2Alpha1.Spec.Service, &apiRuleBeta1.Spec.Service)
	if err != nil {
		return err
	}
	err = convertOverJson(apiRule2Alpha1.Spec.Timeout, &apiRuleBeta1.Spec.Timeout)
	if err != nil {
		return err
	}

	// Status
	apiRuleBeta1.Status = v1beta1.APIRuleStatus{
		APIRuleStatus: &v1beta1.APIRuleResourceStatus{
			Code:        2alpha1to1statusConversionMap[apiRule2Alpha1.Status.State],
			Description: apiRule2Alpha1.Status.Description,
		},
		LastProcessedTime: apiRule2Alpha1.Status.LastProcessedTime,
	}

	if apiRule2Alpha1.Spec.CorsPolicy != nil {
		apiRuleBeta1.Spec.CorsPolicy = &v1beta1.CorsPolicy{}
		apiRuleBeta1.Spec.CorsPolicy.AllowHeaders = apiRule2Alpha1.Spec.CorsPolicy.AllowHeaders
		apiRuleBeta1.Spec.CorsPolicy.AllowMethods = apiRule2Alpha1.Spec.CorsPolicy.AllowMethods
		apiRuleBeta1.Spec.CorsPolicy.AllowOrigins = v1beta1.StringMatch(apiRule2Alpha1.Spec.CorsPolicy.AllowOrigins)
		apiRuleBeta1.Spec.CorsPolicy.AllowCredentials = apiRule2Alpha1.Spec.CorsPolicy.AllowCredentials
		apiRuleBeta1.Spec.CorsPolicy.ExposeHeaders = apiRule2Alpha1.Spec.CorsPolicy.ExposeHeaders

		if apiRule2Alpha1.Spec.CorsPolicy.MaxAge != nil {
			apiRuleBeta1.Spec.CorsPolicy.MaxAge = &metav1.Duration{Duration: time.Duration(*apiRule2Alpha1.Spec.CorsPolicy.MaxAge) * time.Second}
		}
	}

	// Only one host is supported in v1beta1, so we use the first one from the list
	strHost := string(*apiRule2Alpha1.Spec.Hosts[0])
	apiRuleBeta1.Spec.Host = &strHost

	apiRuleBeta1.Spec.Rules = []v1beta1.Rule{}
	for _, rule2alpha1 := range apiRule2Alpha1.Spec.Rules {
		ruleBeta1 := v1beta1.Rule{}
		err = convertOverJson(rule2alpha1, &ruleBeta1)
		if err != nil {
			return err
		}
		// No Auth
		if rule2alpha1.NoAuth != nil && *rule2alpha1.NoAuth {
			ruleBeta1.AccessStrategies = append(ruleBeta1.AccessStrategies, &v1beta1.Authenticator{
				Handler: &v1beta1.Handler{
					Name: "no_auth",
				},
			})
		}
		// JWT
		if rule2alpha1.Jwt != nil {
			ruleBeta1.AccessStrategies = append(ruleBeta1.AccessStrategies, &v1beta1.Authenticator{
				Handler: &v1beta1.Handler{
					Name:   "jwt",
					Config: &runtime.RawExtension{Object: rule2alpha1.Jwt},
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

// Converts from the Hub version (v1beta1) into this ApiRule (v2alpha1)
func (apiRule2Alpha1 *APIRule) ConvertFrom(hub conversion.Hub) error {
	apiRuleBeta1 := hub.(*v1beta1.APIRule)

	apiRule2Alpha1.ObjectMeta = apiRuleBeta1.ObjectMeta

	err := convertOverJson(apiRuleBeta1.Spec.Rules, &apiRule2Alpha1.Spec.Rules)
	if err != nil {
		return err
	}
	err = convertOverJson(apiRuleBeta1.Spec.Gateway, &apiRule2Alpha1.Spec.Gateway)
	if err != nil {
		return err
	}
	err = convertOverJson(apiRuleBeta1.Spec.Service, &apiRule2Alpha1.Spec.Service)
	if err != nil {
		return err
	}
	err = convertOverJson(apiRuleBeta1.Spec.Timeout, &apiRule2Alpha1.Spec.Timeout)
	if err != nil {
		return err
	}

	if apiRuleBeta1.Status.APIRuleStatus != nil {
		apiRule2Alpha1.Status = APIRuleStatus{
			State:             beta1to2statusConversionMap[apiRuleBeta1.Status.APIRuleStatus.Code],
			Description:       apiRuleBeta1.Status.APIRuleStatus.Description,
			LastProcessedTime: apiRuleBeta1.Status.LastProcessedTime,
		}
	}

	apiRule2Alpha1.Spec.Hosts = []*Host{new(Host)}
	*apiRule2Alpha1.Spec.Hosts[0] = Host(*apiRuleBeta1.Spec.Host)
	if apiRuleBeta1.Spec.CorsPolicy != nil {
		apiRule2Alpha1.Spec.CorsPolicy = &CorsPolicy{}
		apiRule2Alpha1.Spec.CorsPolicy.AllowHeaders = apiRuleBeta1.Spec.CorsPolicy.AllowHeaders
		apiRule2Alpha1.Spec.CorsPolicy.AllowMethods = apiRuleBeta1.Spec.CorsPolicy.AllowMethods
		apiRule2Alpha1.Spec.CorsPolicy.AllowOrigins = StringMatch(apiRuleBeta1.Spec.CorsPolicy.AllowOrigins)
		apiRule2Alpha1.Spec.CorsPolicy.AllowCredentials = apiRuleBeta1.Spec.CorsPolicy.AllowCredentials
		apiRule2Alpha1.Spec.CorsPolicy.ExposeHeaders = apiRuleBeta1.Spec.CorsPolicy.ExposeHeaders

		// metav1.Duration type for seconds is float64,
		// however the Access-Control-Max-Age header is specified in seconds without decimals.
		// In consequence, the conversion drops any values smaller than 1 second.
		// https://fetch.spec.whatwg.org/#http-responses
		if apiRuleBeta1.Spec.CorsPolicy.MaxAge != nil {
			maxAge := uint64(apiRuleBeta1.Spec.CorsPolicy.MaxAge.Duration.Seconds())
			apiRule2Alpha1.Spec.CorsPolicy.MaxAge = &maxAge
		}
	}

	apiRule2Alpha1.Spec.Rules = []Rule{}
	for _, ruleBeta1 := range apiRuleBeta1.Spec.Rules {
		rule2alpha1 := Rule{}
		err = convertOverJson(ruleBeta1, &rule2alpha1)
		if err != nil {
			return err
		}
		for _, accessStrategy := range ruleBeta1.AccessStrategies {
			if accessStrategy.Handler.Name == "no_auth" { // No Auth
				rule2alpha1.NoAuth = ptr.To(true)
			} else if accessStrategy.Handler.Name == "jwt" && accessStrategy.Config != nil { // JWT
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
				if jwtConfig.Authentications == nil && jwtConfig.Authorizations == nil { // v1beta1 ory jwt config
					var oryJwtConfig ory.JWTAccStrConfig
					_ = json.Unmarshal(accessStrategy.Config.Raw, &oryJwtConfig)
					if len(oryJwtConfig.JWKSUrls) > 0 {
						return fmt.Errorf(v1beta1DeprecatedTemplate, apiRuleBeta1.Namespace, apiRuleBeta1.Name)
					}
				}
				err = convertOverJson(jwtConfig, &rule2alpha1.Jwt)
				if err != nil {
					return err
				}
			} else {
				return fmt.Errorf(v1beta1DeprecatedTemplate, apiRuleBeta1.Namespace, apiRuleBeta1.Name)
			}
		}
		apiRule2Alpha1.Spec.Rules = append(apiRule2Alpha1.Spec.Rules, rule2alpha1)
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
