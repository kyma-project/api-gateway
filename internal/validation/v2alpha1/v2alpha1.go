package v2alpha1

import (
	"context"
	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	"github.com/kyma-project/api-gateway/internal/validation"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type APIRuleValidator struct {
	api *gatewayv2alpha1.APIRule

	// TODO: I don't know if those validators are enough, for now I added some boilerplate code
	InjectionValidator *validation.InjectionValidator
	RulesValidator     rulesValidator
	JwtValidator       jwtValidator

	DefaultDomainName string
}

type jwtValidator interface {
	Validate(attributePath string, handler *gatewayv2alpha1.JwtConfig) []validation.Failure
}

type jwtValidatorImpl struct{}

func (j *jwtValidatorImpl) Validate(attributePath string, jwtConfig *gatewayv2alpha1.JwtConfig) []validation.Failure {
	//TODO implement me
	return make([]validation.Failure, 0)
}

type rulesValidator interface {
	Validate(attributePath string, rules []*gatewayv2alpha1.Rule) []validation.Failure
}

type rulesValidatorImpl struct{}

func (r rulesValidatorImpl) Validate(attributePath string, rules []*gatewayv2alpha1.Rule) []validation.Failure {
	//TODO implement me
	return make([]validation.Failure, 0)
}

func NewAPIRuleValidator(ctx context.Context, client client.Client, api *gatewayv2alpha1.APIRule, defaultDomainName string) *APIRuleValidator {
	return &APIRuleValidator{
		api:                api,
		InjectionValidator: validation.NewInjectionValidator(ctx, client),
		RulesValidator:     rulesValidatorImpl{},
		JwtValidator:       &jwtValidatorImpl{},
		DefaultDomainName:  defaultDomainName,
	}
}

// TODO: Actually Validate
func (*APIRuleValidator) Validate(ctx context.Context, client client.Client, vsList networkingv1beta1.VirtualServiceList) []validation.Failure {
	//TODO implement me
	return make([]validation.Failure, 0)
}
