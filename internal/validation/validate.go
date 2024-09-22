package validation

import (
	"context"
	"fmt"

	"github.com/kyma-project/api-gateway/internal/helpers"
	"golang.org/x/exp/slices"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ApiRuleValidator interface {
	Validate(ctx context.Context, client client.Client, vsList networkingv1beta1.VirtualServiceList, gwList networkingv1beta1.GatewayList) []Failure
}

// Failure carries validation failures for a single attribute of an object.
type Failure struct {
	AttributePath string
	Message       string
}

func ValidateConfig(config *helpers.Config) []Failure {
	var problems []Failure

	if config == nil {
		problems = append(problems, Failure{
			Message: "Configuration is missing",
		})
	} else {
		if !slices.Contains([]string{helpers.JWT_HANDLER_ORY, helpers.JWT_HANDLER_ISTIO}, config.JWTHandler) {
			problems = append(problems, Failure{
				Message: fmt.Sprintf("Unsupported JWT Handler: %s", config.JWTHandler),
			})
		}
	}

	return problems
}
