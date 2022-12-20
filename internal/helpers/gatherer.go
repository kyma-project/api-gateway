package helpers

import (
	"context"
	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ListApiRule lists all ApiRules on the cluster
func ListApiRule(ctx context.Context, kubeclient client.Client) (*gatewayv1beta1.APIRuleList, error) {
	list := gatewayv1beta1.APIRuleList{}
	err := kubeclient.List(ctx, &list)
	if err != nil {
		return nil, err
	}
	return &list, nil
}
