package networkpolicy

import (
	"context"
	_ "embed"

	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

const (
	OwningResourceLabel = "api-gateway.kyma-project.io/managed-by"
)

//go:embed operator_networkpolicy.yaml
var operatorNetworkPolicy []byte

type OperatorPolicy struct {
	client.Client
	Owner   client.Object
	Enabled bool
}

func (r *OperatorPolicy) Handle(ctx context.Context) error {
	// decode static resource to get name of a NP
	policy := networkingv1.NetworkPolicy{}
	if err := yaml.Unmarshal(operatorNetworkPolicy, &policy); err != nil {
		return err
	}
	desiredSpec := policy.Spec.DeepCopy()

	if !r.Enabled {
		// delete if found
		if err := r.Get(ctx, client.ObjectKeyFromObject(&policy), &policy); err != nil {
			if errors.IsNotFound(err) {
				return nil
			}
			return err
		}
		if err := r.Delete(ctx, &policy); err != nil {
			return err
		}
		return nil
	}
	// create or update
	_, err := ctrl.CreateOrPatch(ctx, r, &policy, func() error {
		// APIGateway CR is cluster-scoped. No OwnerRef possible.
		policy.SetLabels(map[string]string{
			OwningResourceLabel: r.Owner.GetName(),
		})
		desiredSpec.DeepCopyInto(&policy.Spec)
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
