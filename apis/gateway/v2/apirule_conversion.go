package v2

import (
	"encoding/json"

	"sigs.k8s.io/controller-runtime/pkg/conversion"

	"github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
)

// ConvertTo Converts this ApiRule (v2) to the Hub version (v2alpha1)
func (apiRule *APIRule) ConvertTo(hub conversion.Hub) error {
	apiRuleV2alpha1 := hub.(*v2alpha1.APIRule)
	err := convertOverJson(apiRule, &apiRuleV2alpha1)
	if err != nil {
		return err
	}

	return nil
}

// ConvertFrom converts from the Hub version (v2alpha1) into this ApiRule (v2)
func (apiRule *APIRule) ConvertFrom(hub conversion.Hub) error {
	apiRuleV2alpha1 := hub.(*v2alpha1.APIRule)
	err := convertOverJson(apiRuleV2alpha1, &apiRule)
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
