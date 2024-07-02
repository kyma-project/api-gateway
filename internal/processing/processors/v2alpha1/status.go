package v2alpha1

import (
	"github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	v1alpha2Status "github.com/kyma-project/api-gateway/internal/processing/status/v2alpha1"
)

func Base(state string) v1alpha2Status.ReconciliationV2alpha1Status {
	return v1alpha2Status.ReconciliationV2alpha1Status{
		State:       v2alpha1.State(state),
		Description: "",
	}
}
