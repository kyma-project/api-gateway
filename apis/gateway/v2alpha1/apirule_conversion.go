package v2alpha1

import (
	"encoding/json"
	"time"

	"github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

var beta1toV2alpha1StatusConversionMap = map[v1beta1.StatusCode]State{
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
var alpha1to1beta1statusConversionMap = convertMap(beta1toV2alpha1StatusConversionMap)

// ConvertTo Converts this ApiRule (v2alpha1) to the Hub version (v1beta1)
func (apiRuleV2Alpha1 *APIRule) ConvertTo(hub conversion.Hub) error {
	apiRuleBeta1 := hub.(*v1beta1.APIRule)

	apiRuleBeta1.ObjectMeta = apiRuleV2Alpha1.ObjectMeta
	if apiRuleBeta1.Annotations == nil {
		apiRuleBeta1.Annotations = make(map[string]string)
	}
	apiRuleBeta1.Annotations["gateway.kyma-project.io/original-version"] = "v2alpha1"

	err := convertOverJson(apiRuleV2Alpha1.Spec.Rules, &apiRuleBeta1.Spec.Rules)
	if err != nil {
		return err
	}

	err = convertOverJson(apiRuleV2Alpha1.Spec.Gateway, &apiRuleBeta1.Spec.Gateway)
	if err != nil {
		return err
	}
	err = convertOverJson(apiRuleV2Alpha1.Spec.Service, &apiRuleBeta1.Spec.Service)
	if err != nil {
		return err
	}
	err = convertOverJson(apiRuleV2Alpha1.Spec.Timeout, &apiRuleBeta1.Spec.Timeout)
	if err != nil {
		return err
	}

	// Status
	apiRuleBeta1.Status = v1beta1.APIRuleStatus{
		APIRuleStatus: &v1beta1.APIRuleResourceStatus{
			Code:        alpha1to1beta1statusConversionMap[apiRuleV2Alpha1.Status.State],
			Description: apiRuleV2Alpha1.Status.Description,
		},
		LastProcessedTime: apiRuleV2Alpha1.Status.LastProcessedTime,
	}

	if apiRuleV2Alpha1.Spec.CorsPolicy != nil {
		apiRuleBeta1.Spec.CorsPolicy = &v1beta1.CorsPolicy{}
		apiRuleBeta1.Spec.CorsPolicy.AllowHeaders = apiRuleV2Alpha1.Spec.CorsPolicy.AllowHeaders
		apiRuleBeta1.Spec.CorsPolicy.AllowMethods = apiRuleV2Alpha1.Spec.CorsPolicy.AllowMethods
		apiRuleBeta1.Spec.CorsPolicy.AllowOrigins = v1beta1.StringMatch(apiRuleV2Alpha1.Spec.CorsPolicy.AllowOrigins)
		apiRuleBeta1.Spec.CorsPolicy.AllowCredentials = apiRuleV2Alpha1.Spec.CorsPolicy.AllowCredentials
		apiRuleBeta1.Spec.CorsPolicy.ExposeHeaders = apiRuleV2Alpha1.Spec.CorsPolicy.ExposeHeaders

		if apiRuleV2Alpha1.Spec.CorsPolicy.MaxAge != nil {
			apiRuleBeta1.Spec.CorsPolicy.MaxAge = &metav1.Duration{Duration: time.Duration(*apiRuleV2Alpha1.Spec.CorsPolicy.MaxAge) * time.Second}
		}
	}

	if len(apiRuleV2Alpha1.Spec.Hosts) > 0 {
		// Only one host is supported in v1beta1, so we use the first one from the list
		strHost := string(*apiRuleV2Alpha1.Spec.Hosts[0])
		apiRuleBeta1.Spec.Host = &strHost
	}

	if len(apiRuleV2Alpha1.Spec.Rules) > 0 {
		apiRuleBeta1.Spec.Rules = []v1beta1.Rule{}
		for _, ruleV1Alpha2 := range apiRuleV2Alpha1.Spec.Rules {
			ruleBeta1 := v1beta1.Rule{}
			err = convertOverJson(ruleV1Alpha2, &ruleBeta1)
			if err != nil {
				return err
			}
			// No Auth
			if ruleV1Alpha2.NoAuth != nil && *ruleV1Alpha2.NoAuth {
				ruleBeta1.AccessStrategies = append(ruleBeta1.AccessStrategies, &v1beta1.Authenticator{
					Handler: &v1beta1.Handler{
						Name: v1beta1.AccessStrategyNoAuth,
					},
				})
			}
			// JWT
			if ruleV1Alpha2.Jwt != nil {
				ruleBeta1.AccessStrategies = append(ruleBeta1.AccessStrategies, &v1beta1.Authenticator{
					Handler: &v1beta1.Handler{
						Name:   v1beta1.AccessStrategyJwt,
						Config: &runtime.RawExtension{Object: ruleV1Alpha2.Jwt},
					},
				})
			}

			// Mutators
			if ruleV1Alpha2.Request != nil {
				if ruleV1Alpha2.Request.Cookies != nil {
					var config runtime.RawExtension
					err := convertOverJson(ruleV1Alpha2.Request.Cookies, &config)
					if err != nil {
						return err
					}
					ruleBeta1.Mutators = append(ruleBeta1.Mutators, &v1beta1.Mutator{
						Handler: &v1beta1.Handler{
							Name:   v1beta1.CookieMutator,
							Config: &config,
						},
					})
				}

				if ruleV1Alpha2.Request.Headers != nil {
					var config runtime.RawExtension
					err := convertOverJson(ruleV1Alpha2.Request.Headers, &config)
					if err != nil {
						return err
					}
					ruleBeta1.Mutators = append(ruleBeta1.Mutators, &v1beta1.Mutator{
						Handler: &v1beta1.Handler{
							Name:   v1beta1.HeaderMutator,
							Config: &config,
						},
					})
				}
			}

			apiRuleBeta1.Spec.Rules = append(apiRuleBeta1.Spec.Rules, ruleBeta1)
		}
	}

	return nil
}

// Converts from the Hub version (v1beta1) into this ApiRule (v2alpha1)
func (apiRuleV2Alpha1 *APIRule) ConvertFrom(hub conversion.Hub) error {
	apiRuleBeta1 := hub.(*v1beta1.APIRule)

	apiRuleV2Alpha1.ObjectMeta = apiRuleBeta1.ObjectMeta

	if apiRuleBeta1.Status.APIRuleStatus != nil {
		apiRuleV2Alpha1.Status = APIRuleStatus{
			State:             beta1toV2alpha1StatusConversionMap[apiRuleBeta1.Status.APIRuleStatus.Code],
			Description:       apiRuleBeta1.Status.APIRuleStatus.Description,
			LastProcessedTime: apiRuleBeta1.Status.LastProcessedTime,
		}
	}

	conversionPossible, err := isFullConversionPossible(apiRuleBeta1)
	if err != nil {
		return err
	}
	if !conversionPossible {
		// We have to stop the conversion here, because we want to return an empty Spec in case we cannot fully convert the APIRule.
		return nil
	}

	err = convertOverJson(apiRuleBeta1.Spec.Rules, &apiRuleV2Alpha1.Spec.Rules)
	if err != nil {
		return err
	}
	err = convertOverJson(apiRuleBeta1.Spec.Gateway, &apiRuleV2Alpha1.Spec.Gateway)
	if err != nil {
		return err
	}
	err = convertOverJson(apiRuleBeta1.Spec.Service, &apiRuleV2Alpha1.Spec.Service)
	if err != nil {
		return err
	}
	err = convertOverJson(apiRuleBeta1.Spec.Timeout, &apiRuleV2Alpha1.Spec.Timeout)
	if err != nil {
		return err
	}

	if apiRuleBeta1.Spec.CorsPolicy != nil {
		apiRuleV2Alpha1.Spec.CorsPolicy = &CorsPolicy{}
		apiRuleV2Alpha1.Spec.CorsPolicy.AllowHeaders = apiRuleBeta1.Spec.CorsPolicy.AllowHeaders
		apiRuleV2Alpha1.Spec.CorsPolicy.AllowMethods = apiRuleBeta1.Spec.CorsPolicy.AllowMethods
		apiRuleV2Alpha1.Spec.CorsPolicy.AllowOrigins = StringMatch(apiRuleBeta1.Spec.CorsPolicy.AllowOrigins)
		apiRuleV2Alpha1.Spec.CorsPolicy.AllowCredentials = apiRuleBeta1.Spec.CorsPolicy.AllowCredentials
		apiRuleV2Alpha1.Spec.CorsPolicy.ExposeHeaders = apiRuleBeta1.Spec.CorsPolicy.ExposeHeaders

		// metav1.Duration type for seconds is float64,
		// however the Access-Control-Max-Age header is specified in seconds without decimals.
		// In consequence, the conversion drops any values smaller than 1 second.
		// https://fetch.spec.whatwg.org/#http-responses
		if apiRuleBeta1.Spec.CorsPolicy.MaxAge != nil {
			maxAge := uint64(apiRuleBeta1.Spec.CorsPolicy.MaxAge.Duration.Seconds())
			apiRuleV2Alpha1.Spec.CorsPolicy.MaxAge = &maxAge
		}
	}

	if apiRuleBeta1.Spec.Host != nil {
		apiRuleV2Alpha1.Spec.Hosts = []*Host{new(Host)}
		*apiRuleV2Alpha1.Spec.Hosts[0] = Host(*apiRuleBeta1.Spec.Host)
	}

	if len(apiRuleBeta1.Spec.Rules) > 0 {
		apiRuleV2Alpha1.Spec.Rules = []Rule{}
		for _, ruleBeta1 := range apiRuleBeta1.Spec.Rules {
			ruleV1Alpha2 := Rule{}
			err = convertOverJson(ruleBeta1, &ruleV1Alpha2)
			if err != nil {
				return err
			}
			for _, accessStrategy := range ruleBeta1.AccessStrategies {
				if accessStrategy.Handler.Name == v1beta1.AccessStrategyNoAuth {
					ruleV1Alpha2.NoAuth = ptr.To(true)
				}

				if accessStrategy.Handler.Name == v1beta1.AccessStrategyJwt {
					jwtConfig, err := convertToJwtConfig(accessStrategy)
					if err != nil {
						return err
					}
					err = convertOverJson(jwtConfig, &ruleV1Alpha2.Jwt)
					if err != nil {
						return err
					}
				}
			}

			if ruleBeta1.Mutators != nil {
				ruleV1Alpha2.Request = &Request{}
			}

			for _, mutator := range ruleBeta1.Mutators {
				switch mutator.Handler.Name {
				case v1beta1.HeaderMutator:
					var configStruct map[string]string

					err := json.Unmarshal(mutator.Handler.Config.Raw, &configStruct)
					if err != nil {
						return err
					}

					ruleV1Alpha2.Request.Headers = configStruct
				case v1beta1.CookieMutator:
					var configStruct map[string]string

					err := json.Unmarshal(mutator.Handler.Config.Raw, &configStruct)
					if err != nil {
						return err
					}

					ruleV1Alpha2.Request.Cookies = configStruct
				}
			}
			apiRuleV2Alpha1.Spec.Rules = append(apiRuleV2Alpha1.Spec.Rules, ruleV1Alpha2)
		}
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

// isFullConversionPossible checks if the APIRule can be fully converted to v2alpha1 by evaluating the access strategies.
func isFullConversionPossible(apiRule *v1beta1.APIRule) (bool, error) {
	for _, rule := range apiRule.Spec.Rules {
		for _, accessStrategy := range rule.AccessStrategies {

			if accessStrategy.Handler.Name == v1beta1.AccessStrategyNoAuth {
				continue
			}

			if accessStrategy.Handler.Name == v1beta1.AccessStrategyJwt {
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
