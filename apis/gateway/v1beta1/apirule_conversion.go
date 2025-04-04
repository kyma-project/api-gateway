package v1beta1

import (
	"encoding/json"
	"github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	"k8s.io/utils/ptr"
	"maps"
	"regexp"
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
	if apiRule.Annotations == nil {
		apiRule.Annotations = make(map[string]string)
	}
	apiRule.Status = convertV1beta1StatusToV2alpha1Status(r.Status)

	convertible, err := isConvertible(r)
	if err != nil {
		return err
	}
	if !r.isV2OriginalVersion() {
		apiRule.Annotations[originalVersionAnnotationKey] = "v1beta1"
		marshaledSpec, err := json.Marshal(r.Spec)
		if err != nil {
			return err
		}
		apiRule.Annotations[v1beta1SpecAnnotationKey] = string(marshaledSpec)
	}

	if !convertible {
		// We have to stop the conversion here, because we want to return an empty Spec in case we cannot fully convert the APIRule.
		return nil
	}

	err = convertOverJson(r.Spec.Gateway, &apiRule.Spec.Gateway)
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

		if r.Spec.CorsPolicy.MaxAge != nil {
			maxAge := uint64(r.Spec.CorsPolicy.MaxAge.Duration.Seconds())
			apiRule.Spec.CorsPolicy.MaxAge = &maxAge
		}
	}

	if r.Spec.Host != nil {
		host := v2alpha1.Host(*r.Spec.Host)
		apiRule.Spec.Hosts = []*v2alpha1.Host{&host}
	}

	if _, ok := r.Annotations[v2alpha1RulesAnnotationKey]; ok && r.isV2OriginalVersion() {
		err := json.Unmarshal([]byte(r.Annotations[v2alpha1RulesAnnotationKey]), &apiRule.Spec.Rules)
		if err != nil {
			return err
		}
		return nil
	}

	if len(r.Spec.Rules) > 0 {
		apiRule.Spec.Rules = []v2alpha1.Rule{}
		for _, ruleV1beta1 := range r.Spec.Rules {
			ruleV2alpha1 := v2alpha1.Rule{}
			err = convertOverJson(ruleV1beta1, &ruleV2alpha1)
			if err != nil {
				return err
			}

			if len(ruleV1beta1.AccessStrategies) > 0 {
				for _, strategy := range ruleV1beta1.AccessStrategies {
					if strategy.Handler.Name == AccessStrategyNoAuth {
						ruleV2alpha1.NoAuth = ptr.To(true)
					}
					if strategy.Handler.Name == AccessStrategyJwt {
						jwtConfig, err := convertToJwtConfig(strategy)
						if err != nil {
							return err
						}
						err = convertOverJson(jwtConfig, &ruleV2alpha1.Jwt)
						if err != nil {
							return err
						}
					}
				}
			}

			// Mutators
			if len(ruleV1beta1.Mutators) > 0 {
				for _, mutator := range ruleV1beta1.Mutators {
					if ruleV2alpha1.Request == nil {
						ruleV2alpha1.Request = &v2alpha1.Request{}
					}

					var config map[string]string
					err := convertOverJson(mutator.Handler.Config, &config)
					if err != nil {
						return err
					}

					if mutator.Handler.Name == HeaderMutator {
						if ruleV2alpha1.Request.Headers == nil {
							ruleV2alpha1.Request.Headers = make(map[string]string)
						}
						maps.Copy(ruleV2alpha1.Request.Headers, config)
					}
					if mutator.Handler.Name == CookieMutator {
						if ruleV2alpha1.Request.Cookies == nil {
							ruleV2alpha1.Request.Cookies = make(map[string]string)
						}
						maps.Copy(ruleV2alpha1.Request.Cookies, config)
					}
				}
			}

			apiRule.Spec.Rules = append(apiRule.Spec.Rules, ruleV2alpha1)
		}

	}

	return nil
}

func (r *APIRule) isV2OriginalVersion() bool {
	return slices.Contains([]string{"v2", "v2alpha1"}, r.Annotations[originalVersionAnnotationKey])
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
	if apiRule.Annotations[originalVersionAnnotationKey] == "v1beta1" && len(apiRule.Spec.Rules) == 0 {
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

// isConvertible checks if the APIRule can be fully converted to v2alpha1 by evaluating the access strategies.
func isConvertible(apiRule *APIRule) (bool, error) {
	for _, rule := range apiRule.Spec.Rules {

		if slices.Contains([]string{"v2", "v2alpha1"}, apiRule.Annotations[originalVersionAnnotationKey]) {
			return true, nil
		}

		if !IsConvertiblePath(rule.Path) {
			return false, nil
		}

		for _, accessStrategy := range rule.AccessStrategies {

			if accessStrategy.Handler.Name == AccessStrategyNoAuth || accessStrategy.Handler.Name == "ext-auth" {
				continue
			}
			return false, nil
		}

	}

	return true, nil
}

func IsConvertiblePath(path string) bool {
	validIstioPathPattern := `^((\/([A-Za-z0-9-._~!$&'()+,;=:@]|%[0-9a-fA-F]{2})*)|(\/\{\*{1,2}\}))+$|^\/\*$`
	validPathRegex := regexp.MustCompile(validIstioPathPattern)
	return validPathRegex.MatchString(path)
}
