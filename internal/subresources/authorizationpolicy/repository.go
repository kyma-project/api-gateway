package authorizationpolicy

import (
	"context"

	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kyma-project/api-gateway/internal/processing"
	"github.com/kyma-project/api-gateway/internal/subresources"
)

// Repository provides methods to retrieve AuthorizationPolicy resources by owner labels
type Repository interface {
	// GetAll retrieves all AuthorizationPolicy resources that match either legacy owner labels or new owner labels
	GetAll(ctx context.Context, labeler processing.Labeler) ([]*securityv1beta1.AuthorizationPolicy, error)
}

type repository struct {
	client client.Client
}

// NewRepository creates a new instance of the AuthorizationPolicy repository
func NewRepository(client client.Client) Repository {
	return &repository{
		client: client,
	}
}

// GetAll retrieves all AuthorizationPolicy resources with both legacy and new owner labels,
// combining them into a single deduplicated list
func (r *repository) GetAll(ctx context.Context, labeler processing.Labeler) ([]*securityv1beta1.AuthorizationPolicy, error) {
	legacyOwnerLabels := processing.GetLegacyOwnerLabelsFromLabeler(labeler)
	newOwnerLabels := processing.GetOwnerLabels(labeler).Labels()

	// Fetch resources with legacy owner labels
	var legacyList securityv1beta1.AuthorizationPolicyList
	if err := r.client.List(ctx, &legacyList, client.MatchingLabels(legacyOwnerLabels)); err != nil {
		return nil, err
	}

	// Fetch resources with new owner labels
	var newList securityv1beta1.AuthorizationPolicyList
	if err := r.client.List(ctx, &newList, client.MatchingLabels(newOwnerLabels)); err != nil {
		return nil, err
	}

	// Merge and deduplicate the results
	return subresources.MergeResourceSlices(legacyList.Items, newList.Items), nil
}
