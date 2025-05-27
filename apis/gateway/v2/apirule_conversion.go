package v2

import (
	"encoding/json"

	"sigs.k8s.io/controller-runtime/pkg/conversion"

	"github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
)

const OriginalVersionAnnotation = "gateway.kyma-project.io/original-version"

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

	ruleV2alpha1.Annotations[OriginalVersionAnnotation] = "v2"

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
