package accessrule

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kyma-project/api-gateway/internal/processing"
	"github.com/kyma-project/api-gateway/internal/subresources"
	rulev1alpha1 "github.com/kyma-project/api-gateway/internal/types/ory/oathkeeper-maester/api/v1alpha1"
)

var grv = schema.GroupVersionKind{
	Group:   "oathkeeper.ory.sh",
	Kind:    "Rule",
	Version: "v1alpha1",
}

// Repository provides methods to retrieve and delete AccessRule resources by owner labels
type Repository interface {
	// GetAll retrieves all AccessRule resources that match either legacy owner labels or new owner labels
	GetAll(ctx context.Context, labeler processing.Labeler) ([]*rulev1alpha1.Rule, error)
	// DeleteAll deletes all AccessRule resources that match either legacy owner labels or new owner labels
	DeleteAll(ctx context.Context, labeler processing.Labeler) error
}

// NewRepository creates a new instance of the AccessRule repository
func NewRepository(client client.Client) Repository {
	return subresources.NewRepository[*rulev1alpha1.Rule](client, grv)
}
