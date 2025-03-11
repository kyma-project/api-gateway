package v1beta1

import (
	"encoding/json"
	"fmt"
	"github.com/kyma-project/api-gateway/apis/gateway/versions"
	"istio.io/api/networking/v1beta1"
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
	v1beta1SpecAnnotationKey = "gateway.kyma-project.io/v1beta1-spec"
)

var v1beta1toV2alpha1StatusConversionMap = map[StatusCode]v2alpha1.State{
	StatusOK:      v2alpha1.Ready,
	StatusError:   v2alpha1.Error,
	StatusWarning: v2alpha1.Warning,
	// StatusSkipped is not supported in v2alpha1, and it happens only when another component has Error or Warning status
	// In this case, we map it to Warning
	StatusSkipped: v2alpha1.Warning,
}

var alpha1to1beta1statusConversionMap = convertMap(v1beta1toV2alpha1StatusConversionMap)

// ConvertTo Converts APIRule (v1beta1) to the Hub version (v2alpha1)
func (r *APIRule) ConvertTo(hub conversion.Hub) error {
	apiRule := hub.(*v2alpha1.APIRule)
	apiRule.ObjectMeta = r.ObjectMeta
	if apiRule.Annotations == nil {
		apiRule.Annotations = make(map[string]string)
	}
	apiRule.Annotations["gateway.kyma-project.io/original-version"] = "v1beta1"
	marshaledSpec, err := json.Marshal(r.Spec)
	if err != nil {
		return err
	}

	apiRule.Annotations[v1beta1SpecAnnotationKey] = string(marshaledSpec)

	apiRule.Status = v2alpha1.APIRuleStatus{
		LastProcessedTime: r.Status.LastProcessedTime,
		State:             v1beta1toV2alpha1StatusConversionMap[r.Status.APIRuleStatus.Code],
		Description:       r.Status.APIRuleStatus.Description,
	}

	conversionPossible, err := isFullConversionPossible(r)
	if err != nil {
		return err
	}

	if !conversionPossible {
		// We have to stop the conversion here, because we want to return an empty Spec in case we cannot fully convert the APIRule.
		return nil
	}

	err = convertOverJson(r.Spec.Rules, &apiRule.Spec.Rules)
	if err != nil {
		return err
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
					if slices.Index([]string{AccessStrategyNoAuth, AccessStrategyNoop, AccessStrategyAllow}, strategy.Handler.Name) >= 0 {
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

func (r *APIRule) ConvertFrom(hub conversion.Hub) error {
	apiRule := hub.(*v2alpha1.APIRule)

	r.ObjectMeta = apiRule.ObjectMeta
	r.Status.LastProcessedTime = apiRule.Status.LastProcessedTime
	r.Status.APIRuleStatus = &APIRuleResourceStatus{
		Code:        convertMap(v1beta1toV2alpha1StatusConversionMap)[apiRule.Status.State],
		Description: apiRule.Status.Description,
	}
	// if the original version is v1beta1, we need to convert the spec from the annotation to not lose any data
	if apiRule.Annotations["gateway.kyma-project.io/original-version"] == "v1beta1" {
		err := convertOverJson(apiRule.Annotations[v1beta1SpecAnnotationKey], &r.Spec)
		if err != nil {
			return err
		}
		return nil
	}
	err := convertOverJson(apiRule.Spec.Rules, &r.Spec.Rules)
	if err != nil {
		return err
	}

	err = convertOverJson(apiRule.Spec.Gateway, &r.Spec.Gateway)
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

		if r.Spec.CorsPolicy.MaxAge != nil {
			r.Spec.CorsPolicy.MaxAge = &v1.Duration{Duration: time.Duration(*apiRule.Spec.CorsPolicy.MaxAge) * time.Second}
		}
	}

	if len(apiRule.Spec.Hosts) > 0 {
		*r.Spec.Host = string(*apiRule.Spec.Hosts[0])
	}

	// todo finish the conversion

	if apiRule.Spec.CorsPolicy != nil {
		r.Spec.CorsPolicy = &v1beta1.CorsPolicy{}
		r.Spec.CorsPolicy.AllowHeaders = apiRule.Spec.CorsPolicy.AllowHeaders
		r.Spec.CorsPolicy.AllowMethods = apiRule.Spec.CorsPolicy.AllowMethods
		r.Spec.CorsPolicy.AllowOrigins = v1beta1.StringMatch(apiRule.Spec.CorsPolicy.AllowOrigins)
		r.Spec.CorsPolicy.AllowCredentials = apiRule.Spec.CorsPolicy.AllowCredentials
		r.Spec.CorsPolicy.ExposeHeaders = apiRule.Spec.CorsPolicy.ExposeHeaders

		if apiRule.Spec.CorsPolicy.MaxAge != nil {
			r.Spec.CorsPolicy.MaxAge = &metav1.Duration{Duration: time.Duration(*apiRule.Spec.CorsPolicy.MaxAge) * time.Second}
		}
	}

	if len(apiRule.Spec.Hosts) > 0 {
		// Only one host is supported in v1beta1, so we use the first one from the list
		strHost := string(*apiRule.Spec.Hosts[0])
		r.Spec.Host = &strHost
	}

	if len(apiRule.Spec.Rules) > 0 {
		marshaledApiRules, err := json.Marshal(apiRule.Spec.Rules)
		if err != nil {
			return err
		}
		if len(r.Annotations) == 0 {
			r.Annotations = make(map[string]string)
		}
		r.Annotations[v2alpha1RulesAnnotationKey] = string(marshaledApiRules)

		r.Spec.Rules = []v1beta1.Rule{}
		for _, ruleV2Alpha1 := range apiRule.Spec.Rules {
			ruleBeta1 := v1beta1.Rule{}
			err = convertOverJson(ruleV2Alpha1, &ruleBeta1)
			if err != nil {
				return err
			}

			// ExtAuth
			if ruleV2Alpha1.ExtAuth != nil {
				ruleBeta1.AccessStrategies = append(ruleBeta1.AccessStrategies, &v1beta1.Authenticator{
					Handler: &v1beta1.Handler{
						Name: "ext-auth",
					},
				})
			}

			// NoAuth
			if ruleV2Alpha1.NoAuth != nil && *ruleV2Alpha1.NoAuth {
				ruleBeta1.AccessStrategies = append(ruleBeta1.AccessStrategies, &v1beta1.Authenticator{
					Handler: &v1beta1.Handler{
						Name: v1beta1.AccessStrategyNoAuth,
					},
				})
			}
			// JWT
			if ruleV2Alpha1.Jwt != nil {
				ruleBeta1.AccessStrategies = append(ruleBeta1.AccessStrategies, &v1beta1.Authenticator{
					Handler: &v1beta1.Handler{
						Name:   v1beta1.AccessStrategyJwt,
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
					ruleBeta1.Mutators = append(ruleBeta1.Mutators, &v1beta1.Mutator{
						Handler: &v1beta1.Handler{
							Name:   v1beta1.CookieMutator,
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
					ruleBeta1.Mutators = append(ruleBeta1.Mutators, &v1beta1.Mutator{
						Handler: &v1beta1.Handler{
							Name:   v1beta1.HeaderMutator,
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

func convertMap(m map[StatusCode]v2alpha1.State) map[v2alpha1.State]StatusCode {
	inv := make(map[v2alpha1.State]StatusCode)
	for k, v := range m {
		inv[v] = k
	}
	return inv
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
func isFullConversionPossible(apiRule *APIRule) (bool, error) {
	for _, rule := range apiRule.Spec.Rules {

		if !isConvertablePath(rule.Path) {
			return false, nil
		}

		for _, accessStrategy := range rule.AccessStrategies {

			if accessStrategy.Handler.Name == AccessStrategyNoAuth || accessStrategy.Handler.Name == "ext-auth" {
				continue
			}

			if accessStrategy.Handler.Name == AccessStrategyJwt {
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

func isConvertablePath(path string) bool {
	validIstioPathPattern := `^((\/([A-Za-z0-9-._~!$&'()+,;=:@]|%[0-9a-fA-F]{2})*)|(\/\{\*{1,2}\}))+$|^\/\*$`
	validPathRegex := regexp.MustCompile(validIstioPathPattern)
	return validPathRegex.MatchString(path)
}
