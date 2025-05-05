package v1beta1

import (
	"encoding/json"
	"github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
	"slices"
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
func (r *APIRule) ConvertTo(hub conversion.Hub) error {
	apiRule := hub.(*v2alpha1.APIRule)
	apiRule.ObjectMeta = r.ObjectMeta
	apiRule.Status = convertV1beta1StatusToV2alpha1Status(r.Status)

	// if "v2", "v2alpha1" we are sure that resource is v2
	// if is not set and this method is ConvertTo, this means that we are converting from v1beta1 to v2alpha1
	//	 when client will apply new resource without specifying original-version

	if r.Annotations == nil {
		r.Annotations = make(map[string]string)
	}
	if originalVersion, ok := r.Annotations[originalVersionAnnotationKey]; !ok || !slices.Contains([]string{"v2", "v2alpha1"}, originalVersion) {
		if apiRule.Annotations == nil {
			apiRule.Annotations = make(map[string]string)
		}

		marshaledSpec, err := json.Marshal(r.Spec)
		if err != nil {
			return err
		}
		apiRule.Annotations[v1beta1SpecAnnotationKey] = string(marshaledSpec)
		apiRule.Annotations[originalVersionAnnotationKey] = "v1beta1"

		conversionPossible, err := isFullConversionPossible(r)
		if err != nil {
			return err
		}

		if !conversionPossible {
			// if conversion is not possible, we end conversion with an empty spec
			return nil
		}
	}

	err := convertOverJson(r.Spec.Gateway, &apiRule.Spec.Gateway)
	if err != nil {
		return err
	}
	err = convertOverJson(r.Spec.Service, &apiRule.Spec.Service)
	if err != nil {
		return err
	}
	err = convertOverJson(r.Spec.Timeout, &apiRule.Spec.Timeout)
	if err != nil {
		return err
	}

	if r.Spec.CorsPolicy != nil {
		apiRule.Spec.CorsPolicy = &v2alpha1.CorsPolicy{}
		apiRule.Spec.CorsPolicy.AllowHeaders = r.Spec.CorsPolicy.AllowHeaders
		apiRule.Spec.CorsPolicy.AllowMethods = r.Spec.CorsPolicy.AllowMethods
		apiRule.Spec.CorsPolicy.AllowOrigins = v2alpha1.StringMatch(r.Spec.CorsPolicy.AllowOrigins)
		apiRule.Spec.CorsPolicy.AllowCredentials = r.Spec.CorsPolicy.AllowCredentials
		apiRule.Spec.CorsPolicy.ExposeHeaders = r.Spec.CorsPolicy.ExposeHeaders
		// metav1.Duration type for seconds is float64,
		// however the Access-Control-Max-Age header is specified in seconds without decimals.
		// In consequence, the conversion drops any values smaller than 1 second.
		// https://fetch.spec.whatwg.org/#http-responses
		if r.Spec.CorsPolicy.MaxAge != nil {
			maxAge := uint64(r.Spec.CorsPolicy.MaxAge.Seconds())
			apiRule.Spec.CorsPolicy.MaxAge = &maxAge
		}
	}

	if r.Spec.Host != nil {
		host := v2alpha1.Host(*r.Spec.Host)
		apiRule.Spec.Hosts = []*v2alpha1.Host{&host}
	}

	if r.Annotations != nil {
		if _, ok := r.Annotations[v2alpha1RulesAnnotationKey]; ok && r.isV2OriginalVersion() {
			err := json.Unmarshal([]byte(r.Annotations[v2alpha1RulesAnnotationKey]), &apiRule.Spec.Rules)
			if err != nil {
				return err
			}
			return nil
		}
	}
	if len(r.Spec.Rules) > 0 {
		apiRule.Spec.Rules = []v2alpha1.Rule{}
		for _, ruleBeta1 := range r.Spec.Rules {
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
			apiRule.Spec.Rules = append(apiRule.Spec.Rules, ruleV2Alpha1)
		}

	}

	return nil
}

func (r *APIRule) isV2OriginalVersion() bool {
	if r.Annotations == nil {
		return false
	}
	if originalVersion, ok := r.Annotations[originalVersionAnnotationKey]; ok {
		return slices.Contains([]string{"v2", "v2alpha1"}, originalVersion)
	}
	return false
}

func convertV1beta1StatusToV2alpha1Status(state APIRuleStatus) v2alpha1.APIRuleStatus {

	if state.APIRuleStatus != nil && state.APIRuleStatus.Code != "" {
		return v2alpha1.APIRuleStatus{
			State:             v1beta1toV2alpha1StatusConversionMap[state.APIRuleStatus.Code],
			Description:       state.APIRuleStatus.Description,
			LastProcessedTime: state.LastProcessedTime,
		}
	}

	return v2alpha1.APIRuleStatus{
		State:             v2alpha1.Processing,
		Description:       "",
		LastProcessedTime: state.LastProcessedTime,
	}
}

func convertV2alpha1StatusToV1beta1Status(state v2alpha1.APIRuleStatus) APIRuleStatus {
	if state.State == "" {
		return APIRuleStatus{
			APIRuleStatus: &APIRuleResourceStatus{
				Code:        StatusSkipped,
				Description: "",
			},
			LastProcessedTime: state.LastProcessedTime,
		}
	}

	return APIRuleStatus{
		APIRuleStatus: &APIRuleResourceStatus{
			Code:        alpha1to1beta1statusConversionMap[state.State],
			Description: state.Description,
		},
		LastProcessedTime: state.LastProcessedTime,
	}
}

// ConvertFrom converts from the Hub version (v2alpha1) into this APIRule (v1beta1)
func (r *APIRule) ConvertFrom(hub conversion.Hub) error {
	apiRule := hub.(*v2alpha1.APIRule)
	r.ObjectMeta = apiRule.ObjectMeta

	r.Status = convertV2alpha1StatusToV1beta1Status(apiRule.Status)

	// if the original version is v1beta1, we need to convert the spec from the annotation to not lose any data
	if apiRule.Annotations[originalVersionAnnotationKey] == "v1beta1" {
		err := json.Unmarshal([]byte(apiRule.Annotations[v1beta1SpecAnnotationKey]), &r.Spec)
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
func isFullConversionPossible(apiRule *APIRule) (bool, error) {
	for _, rule := range apiRule.Spec.Rules {
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
