package processing_test

import "testing"

func TestSubresourceDeletion(t *testing.T) {
	// given
	strategies := []*gatewayv1beta1.Authenticator{
		{
			Handler: &gatewayv1beta1.Handler{
				Name: "noop",
			},
		},
	}

	allowRule := GetRuleWithServiceFor(ApiPath, ApiMethods, []*gatewayv1beta1.Mutator{}, strategies, nil)
	rules := []gatewayv1beta1.Rule{allowRule}

	apiRule := GetAPIRuleFor(rules)
	client := GetFakeClient()
}
