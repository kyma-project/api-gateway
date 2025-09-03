package v1beta1

import (
	"encoding/json"
	"github.com/kyma-project/api-gateway/internal/gatewaytranslator"
	"slices"

	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/conversion"

	"github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
)

const (
	v1beta1SpecAnnotationKey     = "gateway.kyma-project.io/v1beta1-spec"
	v2alpha1RulesAnnotationKey   = "gateway.kyma-project.io/v2alpha1-rules"
	originalVersionAnnotationKey = "gateway.kyma-project.io/original-version"
)

var v1beta1toV2alpha1StatusConversionMap = map[StatusCode]v2alpha1.State{
	StatusOK:      v2alpha1.Ready,
	StatusError:   v2alpha1.Error,
	StatusWarning: v2alpha1.Warning,
	// StatusSkipped is not supported in v2alpha1, and it happens only when another component has Error or Warning status
	// In this case, we map it to Warning
	StatusSkipped: v2alpha1.Warning,
}

var alpha1to1beta1statusConversionMap = map[v2alpha1.State]StatusCode{
	v2alpha1.Ready:    StatusOK,
	v2alpha1.Error:    StatusError,
	v2alpha1.Warning:  StatusWarning,
	v2alpha1.Deleting: StatusOK,
}

// ConvertTo Converts APIRule (v1beta1) to the Hub version (v2alpha1)
func (ruleV1 *APIRule) ConvertTo(hub conversion.Hub) error {
	ruleV2 := hub.(*v2alpha1.APIRule)
	ruleV2.ObjectMeta = ruleV1.ObjectMeta

	if ruleV1.Status.APIRuleStatus != nil && ruleV1.Status.APIRuleStatus.Code != "" {
		ruleV2.Status = v2alpha1.APIRuleStatus{
			State:             v1beta1toV2alpha1StatusConversionMap[ruleV1.Status.APIRuleStatus.Code],
			Description:       ruleV1.Status.APIRuleStatus.Description,
			LastProcessedTime: ruleV1.Status.LastProcessedTime,
		}
	}

	// if "v2", "v2alpha1" we are sure that resource is v2
	// if is not set and this method is ConvertTo, this means that we are converting from v1beta1 to v2alpha1
	//	 when client will apply new resource without specifying original-version

	if ruleV1.Annotations == nil {
		ruleV1.Annotations = make(map[string]string)
	}
	if originalVersion, ok := ruleV1.Annotations[originalVersionAnnotationKey]; !ok || !slices.Contains([]string{"v2", "v2alpha1"}, originalVersion) {
		if ruleV2.Annotations == nil {
			ruleV2.Annotations = make(map[string]string)
		}

		marshaledSpec, err := json.Marshal(ruleV1.Spec)
		if err != nil {
			return err
		}
		ruleV2.Annotations[v1beta1SpecAnnotationKey] = string(marshaledSpec)
		ruleV2.Annotations[originalVersionAnnotationKey] = "v1beta1"

	}
	conversionPossible, err := isFullConversionPossible(ruleV1)
	if err != nil {
		return err
	}

	if ruleV1.Spec.Gateway != nil && gatewaytranslator.IsOldGatewayNameFormat(*ruleV1.Spec.Gateway) {
		convertedGatewayName, err := gatewaytranslator.TranslateGatewayNameToNewFormat(*ruleV1.Spec.Gateway, ruleV1.Namespace)
		if err != nil {
			return err
		}

		ruleV1.Spec.Gateway = &convertedGatewayName
	}
	err = convertOverJson(ruleV1.Spec.Gateway, &ruleV2.Spec.Gateway)
	if err != nil {
		return err
	}

	err = convertOverJson(ruleV1.Spec.Service, &ruleV2.Spec.Service)
	if err != nil {
		return err
	}
	err = convertOverJson(ruleV1.Spec.Timeout, &ruleV2.Spec.Timeout)
	if err != nil {
		return err
	}

	if ruleV1.Spec.CorsPolicy != nil {
		ruleV2.Spec.CorsPolicy = &v2alpha1.CorsPolicy{}
		ruleV2.Spec.CorsPolicy.AllowHeaders = ruleV1.Spec.CorsPolicy.AllowHeaders
		ruleV2.Spec.CorsPolicy.AllowMethods = ruleV1.Spec.CorsPolicy.AllowMethods
		ruleV2.Spec.CorsPolicy.AllowOrigins = v2alpha1.StringMatch(ruleV1.Spec.CorsPolicy.AllowOrigins)
		ruleV2.Spec.CorsPolicy.AllowCredentials = ruleV1.Spec.CorsPolicy.AllowCredentials
		ruleV2.Spec.CorsPolicy.ExposeHeaders = ruleV1.Spec.CorsPolicy.ExposeHeaders
		// metav1.Duration type for seconds is float64,
		// however the Access-Control-Max-Age header is specified in seconds without decimals.
		// In consequence, the conversion drops any values smaller than 1 second.
		// https://fetch.spec.whatwg.org/#http-responses
		if ruleV1.Spec.CorsPolicy.MaxAge != nil {
			maxAge := uint64(ruleV1.Spec.CorsPolicy.MaxAge.Seconds())
			ruleV2.Spec.CorsPolicy.MaxAge = &maxAge
		}
	}

	if ruleV1.Spec.Host != nil {
		host := v2alpha1.Host(*ruleV1.Spec.Host)
		ruleV2.Spec.Hosts = []*v2alpha1.Host{&host}
	}
	if !conversionPossible {
		// if conversion is not possible, we end conversion with an empty rules array
		return nil
	}

	if ruleV1.Annotations != nil {
		if _, ok := ruleV1.Annotations[v2alpha1RulesAnnotationKey]; ok && ruleV1.isV2OriginalVersion() {
			err := json.Unmarshal([]byte(ruleV1.Annotations[v2alpha1RulesAnnotationKey]), &ruleV2.Spec.Rules)
			if err != nil {
				return err
			}
			return nil
		}
	}
	if len(ruleV1.Spec.Rules) > 0 {
		ruleV2.Spec.Rules = []v2alpha1.Rule{}
		for _, ruleBeta1 := range ruleV1.Spec.Rules {
			ruleV2Alpha1 := v2alpha1.Rule{}
			err = convertOverJson(ruleBeta1, &ruleV2Alpha1)
			if err != nil {
				return err
			}
			for _, accessStrategy := range ruleBeta1.AccessStrategies {
				if accessStrategy.Name == AccessStrategyNoAuth {
					ruleV2Alpha1.NoAuth = ptr.To(true)
				}

				if accessStrategy.Name == AccessStrategyJwt {
					jwtConfig, err := convertToJwtConfig(accessStrategy)
					if err != nil {
						return err
					}
					err = convertOverJson(jwtConfig, &ruleV2Alpha1.Jwt)
					if err != nil {
						return err
					}
				}
			}

			if ruleBeta1.Mutators != nil {
				ruleV2Alpha1.Request = &v2alpha1.Request{}
			}

			for _, mutator := range ruleBeta1.Mutators {
				switch mutator.Name {
				case HeaderMutator:
					var headersRaw struct {
						Headers map[string]string `json:"headers"`
					}

					err := json.Unmarshal(mutator.Config.Raw, &headersRaw)
					if err != nil {
						return err
					}

					ruleV2Alpha1.Request.Headers = headersRaw.Headers
				case CookieMutator:
					var cookiesRaw struct {
						Cookies map[string]string `json:"cookies"`
					}

					err := json.Unmarshal(mutator.Config.Raw, &cookiesRaw)
					if err != nil {
						return err
					}

					ruleV2Alpha1.Request.Cookies = cookiesRaw.Cookies
				}
			}
			ruleV2.Spec.Rules = append(ruleV2.Spec.Rules, ruleV2Alpha1)
		}

	}

	return nil
}

func (ruleV1 *APIRule) isV2OriginalVersion() bool {
	if ruleV1.Annotations == nil {
		return false
	}
	if originalVersion, ok := ruleV1.Annotations[originalVersionAnnotationKey]; ok {
		return slices.Contains([]string{"v2", "v2alpha1"}, originalVersion)
	}
	return false
}

// ConvertFrom converts from the Hub version (v2alpha1) into this APIRule (v1beta1)
func (ruleV1 *APIRule) ConvertFrom(hub conversion.Hub) error {
	ruleV2 := hub.(*v2alpha1.APIRule)
	ruleV1.ObjectMeta = ruleV2.ObjectMeta
	if ruleV2.Status.State != "" {
		ruleV1.Status = APIRuleStatus{
			APIRuleStatus: &APIRuleResourceStatus{
				Code:        alpha1to1beta1statusConversionMap[ruleV2.Status.State],
				Description: ruleV2.Status.Description,
			},
			LastProcessedTime: ruleV2.Status.LastProcessedTime,
		}
	}

	// if the original version is v1beta1, we need to convert the spec from the annotation to not lose any data
	if ruleV2.Annotations[originalVersionAnnotationKey] == "v1beta1" {
		err := json.Unmarshal([]byte(ruleV2.Annotations[v1beta1SpecAnnotationKey]), &ruleV1.Spec)
		if err != nil {
			return err
		}
		return nil
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

// isFullConversionPossible checks if the APIRule can be fully converted to v2 by evaluating the access strategies and path.
func isFullConversionPossible(ruleV1 *APIRule) (bool, error) {
	for _, rule := range ruleV1.Spec.Rules {
		if !isConvertiblePath(rule.Path) {
			return false, nil
		}
		for _, accessStrategy := range rule.AccessStrategies {
			if accessStrategy.Name == AccessStrategyNoAuth {
				continue
			}

			if accessStrategy.Name == AccessStrategyJwt {
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
