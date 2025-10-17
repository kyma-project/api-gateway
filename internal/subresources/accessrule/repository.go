package accessrule

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/kyma-project/api-gateway/internal/processing"
	"github.com/kyma-project/api-gateway/internal/subresources"
	rulev1alpha1 "github.com/kyma-project/api-gateway/internal/types/ory/oathkeeper-maester/api/v1alpha1"
)

// Repository provides methods to retrieve and delete AccessRule resources by owner labels
type Repository interface {
	// GetAll retrieves all AccessRule resources that match either legacy owner labels or new owner labels
	GetAll(ctx context.Context, labeler processing.Labeler) ([]*rulev1alpha1.Rule, error)
	// DeleteAll deletes all AccessRule resources that match either legacy owner labels or new owner labels
	DeleteAll(ctx context.Context, labeler processing.Labeler) error
}

type repository struct {
	client client.Client
}

// NewRepository creates a new instance of the AccessRule repository
func NewRepository(client client.Client) Repository {
	return &repository{
		client: client,
	}
}

// GetAll retrieves all AccessRule resources with both legacy and new owner labels,
// combining them into a single deduplicated list
func (r *repository) GetAll(ctx context.Context, labeler processing.Labeler) ([]*rulev1alpha1.Rule, error) {
	legacyOwnerLabels := processing.GetLegacyOwnerLabelsFromLabeler(labeler)
	newOwnerLabels := processing.GetOwnerLabels(labeler).Labels()

	// Fetch resources with legacy owner labels
	var legacyList rulev1alpha1.RuleList
	if err := r.client.List(ctx, &legacyList, client.MatchingLabels(legacyOwnerLabels)); err != nil {
		return nil, err
	}

	// Fetch resources with new owner labels
	var newList rulev1alpha1.RuleList
	if err := r.client.List(ctx, &newList, client.MatchingLabels(newOwnerLabels)); err != nil {
		return nil, err
	}

	// Convert to pointer slices for merging
	legacyPointers := make([]*rulev1alpha1.Rule, len(legacyList.Items))
	for i := range legacyList.Items {
		legacyPointers[i] = &legacyList.Items[i]
	}

	newPointers := make([]*rulev1alpha1.Rule, len(newList.Items))
	for i := range newList.Items {
		newPointers[i] = &newList.Items[i]
	}

	// Merge and deduplicate the results
	return subresources.MergeResourceSlices(legacyPointers, newPointers), nil
}

// DeleteAll retrieves and deletes all AccessRule resources with both legacy and new owner labels
func (r *repository) DeleteAll(ctx context.Context, labeler processing.Labeler) error {
	rules, err := r.GetAll(ctx, labeler)
	if err != nil {
		return err
	}

	for _, rule := range rules {
		log.Log.Info("Removing subresource", "Rule", rule.Name)
		if err := r.client.Delete(ctx, rule); err != nil {
			return err
		}
	}

	return nil
}
