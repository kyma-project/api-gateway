package v1alpha1

import (
	"encoding/json"
	"log"

	"github.com/kyma-incubator/api-gateway/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

// ConvertTo converts this ApiRule to the Hub version (v1beta1).
func (src *APIRule) ConvertTo(dstRaw conversion.Hub) error {
	log.Default().Printf("dst host: %s", src.Name)
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

	host := *src.Spec.Service.Host
	dst.Spec.Host = &host

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

	if src.Spec.Service == nil {
		log.Default().Print("conversion from v1beta1 to v1alpha1 wasn't possible as service isn't set on spec level")
		return nil
	}

	for _, rule := range src.Spec.Rules {
		if rule.Service != nil {
			log.Default().Print("conversion from v1beta1 to v1alpha1 isn't possible with rule level service definition")
			return nil
		}
	}

	host := *src.Spec.Host

	dst.Spec.Service.Host = &host

	return nil
}
