package validation

import (
	"fmt"
	"github.com/kyma-project/api-gateway/api/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

// We limit the maximum timeout to 65 minutes as this is the lowest timeout supported by the supported hyperscaler load balancers.
// The limiting hyperscaler is AWS with a supported timeout of up to 4000s.
const maximumTimeout = 65 * time.Minute

type timeoutValidator struct {
}

func (timeoutValidator) validateApiRule(apiRule v1beta1.APIRule) []Failure {
	return validate(apiRule.Spec.Timeout, "spec.timeout")
}

func (timeoutValidator) validateRule(rule v1beta1.Rule, failureAttributePath string) []Failure {
	return validate(rule.Timeout, fmt.Sprintf("%s.timeout", failureAttributePath))
}

func validate(timeout *metav1.Duration, failureAttributePath string) []Failure {
	if timeout == nil {
		return nil
	}

	if timeout.Duration > maximumTimeout {
		return []Failure{{
			AttributePath: failureAttributePath,
			Message:       "Timeout must not exceed 65m",
		}}
	}

	if timeout.Duration <= 0 {
		return []Failure{{
			AttributePath: failureAttributePath,
			Message:       "Timeout must not be 0 or lower",
		}}
	}

	return nil
}
