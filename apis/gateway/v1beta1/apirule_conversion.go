package v1beta1

import (
	"encoding/json"
	"log"

	"github.com/kyma-project/api-gateway/apis/gateway/v1beta2"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

// Hub marks this type as a conversion hub.
func (*APIRule) Hub() {}

// ConvertTo converts this ApiRule to the Hub version (v1beta2).
func (src *APIRule) ConvertTo(dstRaw conversion.Hub) error {
	log.Default().Printf("dst host: %s", src.Name)
	dst := dstRaw.(*v1beta2.APIRule)

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

	host := *src.Spec.Host
	dst.Spec.Hosts = []*string{&host}

	for _, rule := range src.Spec.Rules {
		convertedRule := v1beta2.Rule{}
		for _, accessStrategy := range rule.AccessStrategies {
			if accessStrategy.Handler.Name == AccessStrategyJwt {
				convertedRule.Jwt = &v1beta2.JwtConfig{}
			}

		}

		dst.Spec.Rules = append(dst.Spec.Rules, convertedRule)
	}

	return nil
}

// ConvertFrom converts this ApiRule from the Hub version (v1beta2).
func (dst *APIRule) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1beta2.APIRule)
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

	for _, rule := range src.Spec.Rules {
		if rule.Service != nil {
			log.Default().Print("conversion from v1beta1 to v1alpha1 isn't possible with rule level service definition")
			return nil
		}
	}

	hosts := src.Spec.Hosts
	dst.Spec.Host = hosts[0]

	return nil
}
