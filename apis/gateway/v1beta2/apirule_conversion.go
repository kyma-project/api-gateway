package v1beta2

import (
	"encoding/json"

	"github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

// ConvertTo converts this ApiRule to the Hub version (v1beta1).
func (src *APIRule) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1beta1.APIRule)

	specData, err := json.Marshal(src.Spec)
	if err != nil {
		return err
	}

	err = json.Unmarshal(specData, &dst.Spec)
	if err != nil {
		return err
	}

	statusData, err := json.Marshal(src.Status)
	if err != nil {
		return err
	}

	err = json.Unmarshal(statusData, &dst.Status)
	if err != nil {
		return err
	}

	dst.ObjectMeta = src.ObjectMeta

	// Only one host is supported in v1beta1, so we use the first one from the list
	hosts := src.Spec.Hosts
	dst.Spec.Host = hosts[0]

	for _, rule := range src.Spec.Rules {
		convertedRule := v1beta1.Rule{}
		// No Auth
		if rule.NoAuth != nil && *rule.NoAuth {
			convertedRule.AccessStrategies = append(convertedRule.AccessStrategies, &v1beta1.Authenticator{
				Handler: &v1beta1.Handler{
					Name: "no_auth",
				},
			})
		}

		dst.Spec.Rules = append(dst.Spec.Rules, convertedRule)
	}

	return nil
}

// ConvertFrom converts this ApiRule from the Hub version (v1beta1).
func (dst *APIRule) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1beta1.APIRule)
	specData, err := json.Marshal(src.Spec)
	if err != nil {
		return err
	}

	err = json.Unmarshal(specData, &dst.Spec)
	if err != nil {
		return err
	}

	statusData, err := json.Marshal(src.Status)
	if err != nil {
		return err
	}

	err = json.Unmarshal(statusData, &dst.Status)
	if err != nil {
		return err
	}

	dst.ObjectMeta = src.ObjectMeta

	// Only one host is supported in v1beta1, so we use the first one from the list
	host := src.Spec.Host
	dst.Spec.Hosts = append(dst.Spec.Hosts, host)

	for _, rule := range src.Spec.Rules {
		convertedRule := Rule{}
		for _, accessStrategy := range rule.AccessStrategies {
			// No Auth
			if accessStrategy.Handler.Name == "no_auth" {
				convertedRule.NoAuth = new(bool)
				*convertedRule.NoAuth = true
			}
		}

		dst.Spec.Rules = append(dst.Spec.Rules, convertedRule)
	}

	return nil
}
