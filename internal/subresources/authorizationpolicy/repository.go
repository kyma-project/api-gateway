package authorizationpolicy

import (
	"context"

	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kyma-project/api-gateway/internal/processing"
	"github.com/kyma-project/api-gateway/internal/subresources"
)

var grv = schema.GroupVersionKind{
	Group:   "security.istio.io",
	Kind:    "AuthorizationPolicy",
	Version: "v1beta1",
}

// Repository provides methods to retrieve and delete AuthorizationPolicy resources by owner labels
type Repository interface {
	// GetAll retrieves all AuthorizationPolicy resources that match either legacy owner labels or new owner labels
	GetAll(ctx context.Context, labeler processing.Labeler) ([]*securityv1beta1.AuthorizationPolicy, error)
	// DeleteAll deletes all AuthorizationPolicy resources that match either legacy owner labels or new owner labels
	DeleteAll(ctx context.Context, labeler processing.Labeler) error
}

// NewRepository creates a new instance of the AuthorizationPolicy repository
func NewRepository(client client.Client) Repository {
	return subresources.NewRepository[*securityv1beta1.AuthorizationPolicy](client, grv)
}
