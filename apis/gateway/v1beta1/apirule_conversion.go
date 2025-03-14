package v1beta1

import (
	"encoding/json"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"k8s.io/utils/strings/slices"
	"maps"
	"regexp"
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/conversion"

	"github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
)

const (
	v1beta1RulesAnnotationKey    = "gateway.kyma-project.io/v1beta1-rules"
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
	v2alpha1.Ready:   StatusOK,
	v2alpha1.Error:   StatusError,
	v2alpha1.Warning: StatusWarning,
}

// ConvertTo Converts APIRule (v1beta1) to the Hub version (v2alpha1)
func (r *APIRule) ConvertTo(hub conversion.Hub) error {
	apiRule := hub.(*v2alpha1.APIRule)
	apiRule.ObjectMeta = r.ObjectMeta
	if apiRule.Annotations == nil {
		apiRule.Annotations = make(map[string]string)
	}

	if !slices.Contains([]string{"v2", "v2alpha1"}, r.Annotations[originalVersionAnnotationKey]) {
		apiRule.Annotations[originalVersionAnnotationKey] = "v1beta1"
	}
	marshaledRules, err := json.Marshal(r.Spec.Rules)
	if err != nil {
		return err
	}

	apiRule.Annotations[v1beta1RulesAnnotationKey] = string(marshaledRules)

	apiRule.Status = v2alpha1.APIRuleStatus{
		LastProcessedTime: r.Status.LastProcessedTime,
		State:             v1beta1toV2alpha1StatusConversionMap[r.Status.APIRuleStatus.Code],
		Description:       r.Status.APIRuleStatus.Description,
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

	convertible, err := isConvertible(r)
	if err != nil {
		return err
	}

	if !convertible {
		// We have to stop the conversion here, because we want to return an empty Spec in case we cannot fully convert the APIRule.
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

// ConvertFrom converts from the Hub version (v2alpha1) into this APIRule (v1beta1)
func (r *APIRule) ConvertFrom(hub conversion.Hub) error {
	apiRule := hub.(*v2alpha1.APIRule)

	r.ObjectMeta = apiRule.ObjectMeta
	r.Status.LastProcessedTime = apiRule.Status.LastProcessedTime
	r.Status.APIRuleStatus = &APIRuleResourceStatus{
		Code:        alpha1to1beta1statusConversionMap[apiRule.Status.State],
		Description: apiRule.Status.Description,
	}

	err := convertOverJson(apiRule.Spec.Gateway, &r.Spec.Gateway)
	if err != nil {
		return err
	}
	err = convertOverJson(apiRule.Spec.Service, &r.Spec.Service)
	if err != nil {
		return err
	}
	err = convertOverJson(apiRule.Spec.Timeout, &r.Spec.Timeout)
	if err != nil {
		return err
	}
	if apiRule.Spec.CorsPolicy != nil {
		r.Spec.CorsPolicy = &CorsPolicy{}
		r.Spec.CorsPolicy.AllowHeaders = apiRule.Spec.CorsPolicy.AllowHeaders
		r.Spec.CorsPolicy.AllowMethods = apiRule.Spec.CorsPolicy.AllowMethods
		r.Spec.CorsPolicy.AllowOrigins = StringMatch(apiRule.Spec.CorsPolicy.AllowOrigins)
		r.Spec.CorsPolicy.AllowCredentials = apiRule.Spec.CorsPolicy.AllowCredentials
		r.Spec.CorsPolicy.ExposeHeaders = apiRule.Spec.CorsPolicy.ExposeHeaders

		if apiRule.Spec.CorsPolicy.MaxAge != nil {
			r.Spec.CorsPolicy.MaxAge = &v1.Duration{Duration: time.Duration(*apiRule.Spec.CorsPolicy.MaxAge) * time.Second}
		}
	}

	if len(apiRule.Spec.Hosts) > 0 {
		// Only one host is supported in v1beta1, so we use the first one from the list
		strHost := string(*apiRule.Spec.Hosts[0])
		r.Spec.Host = &strHost
	}

	// if the original version is v1beta1, we need to convert the spec from the annotation to not lose any data
	if apiRule.Annotations[originalVersionAnnotationKey] == "v1beta1" {
		err := json.Unmarshal([]byte(apiRule.Annotations[v1beta1RulesAnnotationKey]), &r.Spec.Rules)
		if err != nil {
			return err
		}
		return nil
	}

	if len(apiRule.Spec.Rules) > 0 {
		r.Spec.Rules = []Rule{}
		for _, ruleV2Alpha1 := range apiRule.Spec.Rules {
			ruleBeta1 := Rule{}
			err = convertOverJson(ruleV2Alpha1, &ruleBeta1)
			if err != nil {
				return err
			}

			// ExtAuth
			if ruleV2Alpha1.ExtAuth != nil {
				ruleBeta1.AccessStrategies = append(ruleBeta1.AccessStrategies, &Authenticator{
					Handler: &Handler{
						Name: "ext-auth",
					},
				})
			}

			// NoAuth
			if ruleV2Alpha1.NoAuth != nil && *ruleV2Alpha1.NoAuth {
				ruleBeta1.AccessStrategies = append(ruleBeta1.AccessStrategies, &Authenticator{
					Handler: &Handler{
						Name: AccessStrategyNoAuth,
					},
				})
			}
			// JWT
			if ruleV2Alpha1.Jwt != nil {
				ruleBeta1.AccessStrategies = append(ruleBeta1.AccessStrategies, &Authenticator{
					Handler: &Handler{
						Name:   AccessStrategyJwt,
						Config: &runtime.RawExtension{Object: ruleV2Alpha1.Jwt},
					},
				})
			}

			// Mutators
			if ruleV2Alpha1.Request != nil {
				if ruleV2Alpha1.Request.Cookies != nil {
					var config runtime.RawExtension
					err := convertOverJson(ruleV2Alpha1.Request.Cookies, &config)
					if err != nil {
						return err
					}
					ruleBeta1.Mutators = append(ruleBeta1.Mutators, &Mutator{
						Handler: &Handler{
							Name:   CookieMutator,
							Config: &config,
						},
					})
				}

				if ruleV2Alpha1.Request.Headers != nil {
					var config runtime.RawExtension
					err := convertOverJson(ruleV2Alpha1.Request.Headers, &config)
					if err != nil {
						return err
					}
					ruleBeta1.Mutators = append(ruleBeta1.Mutators, &Mutator{
						Handler: &Handler{
							Name:   HeaderMutator,
							Config: &config,
						},
					})
				}
			}

			r.Spec.Rules = append(r.Spec.Rules, ruleBeta1)
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

// isConvertible checks if the APIRule can be fully converted to v2alpha1 by evaluating the access strategies.
func isConvertible(apiRule *APIRule) (bool, error) {
	for _, rule := range apiRule.Spec.Rules {

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
