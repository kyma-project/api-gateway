package virtualservice

import (
	"context"

	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kyma-project/api-gateway/internal/processing"
	"github.com/kyma-project/api-gateway/internal/subresources"
)

var grv = schema.GroupVersionKind{
	Group:   "networking.istio.io",
	Kind:    "VirtualService",
	Version: "v1beta1",
}

// Repository provides methods to retrieve and delete VirtualService resources by owner labels
type Repository interface {
	// GetAll retrieves all VirtualService resources that match either legacy owner labels or new owner labels
	GetAll(ctx context.Context, labeler processing.Labeler) ([]*networkingv1beta1.VirtualService, error)
	// DeleteAll deletes all VirtualService resources that match either legacy owner labels or new owner labels
	DeleteAll(ctx context.Context, labeler processing.Labeler) error
}

// NewRepository creates a new instance of the VirtualService repository
func NewRepository(client client.Client) Repository {
	return subresources.NewRepository[*networkingv1beta1.VirtualService](client, grv)
}
