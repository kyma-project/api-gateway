package virtualservice

import (
	"context"

	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/kyma-project/api-gateway/internal/processing"
	"github.com/kyma-project/api-gateway/internal/subresources"
)

// Repository provides methods to retrieve and delete VirtualService resources by owner labels
type Repository interface {
	// GetAll retrieves all VirtualService resources that match either legacy owner labels or new owner labels
	GetAll(ctx context.Context, labeler processing.Labeler) ([]*networkingv1beta1.VirtualService, error)
	// DeleteAll deletes all VirtualService resources that match either legacy owner labels or new owner labels
	DeleteAll(ctx context.Context, labeler processing.Labeler) error
}

type repository struct {
	client client.Client
}

// NewRepository creates a new instance of the VirtualService repository
func NewRepository(client client.Client) Repository {
	return &repository{
		client: client,
	}
}

// GetAll retrieves all VirtualService resources with both legacy and new owner labels,
// combining them into a single deduplicated list
func (r *repository) GetAll(ctx context.Context, labeler processing.Labeler) ([]*networkingv1beta1.VirtualService, error) {
	legacyOwnerLabels := processing.GetLegacyOwnerLabelsFromLabeler(labeler)
	newOwnerLabels := processing.GetOwnerLabels(labeler).Labels()

	// Fetch resources with legacy owner labels
	var legacyList networkingv1beta1.VirtualServiceList
	if err := r.client.List(ctx, &legacyList, client.MatchingLabels(legacyOwnerLabels)); err != nil {
		return nil, err
	}

	// Fetch resources with new owner labels
	var newList networkingv1beta1.VirtualServiceList
	if err := r.client.List(ctx, &newList, client.MatchingLabels(newOwnerLabels)); err != nil {
		return nil, err
	}

	// Merge and deduplicate the results
	return subresources.MergeResourceSlices(legacyList.Items, newList.Items), nil
}

// DeleteAll retrieves and deletes all VirtualService resources with both legacy and new owner labels
func (r *repository) DeleteAll(ctx context.Context, labeler processing.Labeler) error {
	virtualServices, err := r.GetAll(ctx, labeler)
	if err != nil {
		return err
	}

	for _, vs := range virtualServices {
		log.Log.Info("Removing subresource", "VirtualService", vs.Name)
		if err := r.client.Delete(ctx, vs); err != nil {
			return err
		}
	}

	return nil
}
