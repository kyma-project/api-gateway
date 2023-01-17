package v1alpha1

import (
	"encoding/json"
	"errors"
	"log"

	"github.com/kyma-incubator/api-gateway/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

// ConvertTo converts this ApiRule to the Hub version (v1beta1).
func (src *APIRule) ConvertTo(dstRaw conversion.Hub) error {
	dst, ok := dstRaw.(*v1beta1.APIRule)
	if !ok {
		return errors.New("Couldn't get v1alpha1 spec")
	}

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

	if src.Spec.Service == nil || src.Spec.Service.Host == nil {
		log.Default().Printf("conversion from v1alpha1 to v1beta1 wasn't possible as service or service.host was nil for %s", src.Name)
		return nil
	}

	host := *src.Spec.Service.Host
	dst.Spec.Host = &host

	return nil
}

// ConvertFrom converts this ApiRule from the Hub version (v1beta1).
func (dst *APIRule) ConvertFrom(srcRaw conversion.Hub) error {
	src, ok := srcRaw.(*v1beta1.APIRule)
	if !ok {
		return errors.New("Couldn't get v1beta1 spec")
	}
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

	if src.Spec.Host == nil {
		log.Default().Printf("conversion from v1beta1 to v1alpha1 wasn't possible as host was nil for %s", src.Name)
		return nil
	}

	host := *src.Spec.Host

	dst.Spec.Service.Host = &host

	return nil
}
