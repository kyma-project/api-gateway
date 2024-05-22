package v1beta2

import (
	"encoding/json"
	"errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"

	"github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
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

	err = convertOverJson(apiRuleBeta2.Status, &apiRuleBeta1.Status)
	if err != nil {
		return err
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

	err := convertOverJson(apiRuleBeta1.Spec.Rules, &apiRuleBeta2.Spec.Rules)
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
	err = convertOverJson(apiRuleBeta1.Status, &apiRuleBeta2.Status)
	if err != nil {
		return err
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
