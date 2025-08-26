package v2

import (
	"encoding/json"

	"github.com/thoas/go-funk"
	"sigs.k8s.io/controller-runtime/pkg/conversion"

	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
)

const (
	OriginalVersionAnnotation = "gateway.kyma-project.io/original-version"
	v1beta1SpecAnnotationKey  = "gateway.kyma-project.io/v1beta1-spec"
)

// ConvertTo Converts this ApiRule (v2) to the Hub version (v2alpha1)
func (ruleV2 *APIRule) ConvertTo(hub conversion.Hub) error {
	ruleV2alpha1 := hub.(*v2alpha1.APIRule)
	ruleV2alpha1.ObjectMeta = ruleV2.ObjectMeta
	err := convertOverJson(ruleV2.Status, &ruleV2alpha1.Status)
	if err != nil {
		return err
	}
	if ruleV2alpha1.Annotations == nil {
		ruleV2alpha1.Annotations = make(map[string]string)
	}

	val, exists := ruleV2alpha1.Annotations[OriginalVersionAnnotation]
	if !exists || val == "v1beta1" && !funk.IsEmpty(ruleV2.Spec.Rules) {
		ruleV2alpha1.Annotations[OriginalVersionAnnotation] = "v2"
	}

	err = convertOverJson(ruleV2.Spec, &ruleV2alpha1.Spec)
	if err != nil {
		return err
	}

	return nil
}

// ConvertFrom converts from the Hub version (v2alpha1) into this ApiRule (v2)
func (ruleV2 *APIRule) ConvertFrom(hub conversion.Hub) error {
	ruleV2alpha1 := hub.(*v2alpha1.APIRule)
	ruleV2.ObjectMeta = ruleV2alpha1.ObjectMeta

	err := convertOverJson(ruleV2alpha1.Status, &ruleV2.Status)
	if err != nil {
		return err
	}

	err = convertOverJson(ruleV2alpha1.Spec, &ruleV2.Spec)
	if err != nil {
		return err
	}

	if ruleV2alpha1.Spec.Gateway == nil {
		if ruleV2alpha1.Annotations[OriginalVersionAnnotation] == "v1beta1" {
			ruleV1 := &gatewayv1beta1.APIRule{}
			err := json.Unmarshal([]byte(ruleV2alpha1.Annotations[v1beta1SpecAnnotationKey]), &ruleV1.Spec)
			if err != nil {
				return err
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

			if ruleV2.Spec.CorsPolicy != nil {
				ruleV2.Spec.CorsPolicy = &CorsPolicy{}
				ruleV2.Spec.CorsPolicy.AllowHeaders = ruleV1.Spec.CorsPolicy.AllowHeaders
				ruleV2.Spec.CorsPolicy.AllowMethods = ruleV1.Spec.CorsPolicy.AllowMethods
				ruleV2.Spec.CorsPolicy.AllowOrigins = StringMatch(ruleV1.Spec.CorsPolicy.AllowOrigins)
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
				host := Host(*ruleV1.Spec.Host)
				ruleV2.Spec.Hosts = []*Host{&host}
			}
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
