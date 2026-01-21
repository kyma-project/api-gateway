package subresources

import (
	"context"
	"golang.org/x/exp/maps"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/kyma-project/api-gateway/internal/processing"
)

type IRepository[T client.Object] interface {
	GetAll(ctx context.Context, labeler processing.Labeler) ([]T, error)
	// DeleteAll deletes all AccessRule resources that match either legacy owner labels or new owner labels
	DeleteAll(ctx context.Context, labeler processing.Labeler) error
}

type Repository[T client.Object] struct {
	client           client.Client
	groupVersionKind schema.GroupVersionKind
}

func NewRepository[T client.Object](c client.Client, grv schema.GroupVersionKind) *Repository[T] {
	return &Repository[T]{
		client:           c,
		groupVersionKind: grv,
	}
}

// GetAll retrieves all AccessRule resources with both legacy and new owner labels,
// combining them into a single deduplicated list
func (r *Repository[T]) GetAll(ctx context.Context, labeler processing.Labeler) ([]T, error) {
	legacyOwnerLabels := processing.GetLegacyOwnerLabelsFromLabeler(labeler)
	newOwnerLabels := processing.GetOwnerLabels(labeler).Labels()
	legacyList := unstructured.UnstructuredList{}
	if len(maps.Values(legacyOwnerLabels)[0]) <= 63 {
		// Fetch resources with legacyList owner labels
		legacyList.SetGroupVersionKind(r.groupVersionKind)
		if err := r.client.List(ctx, &legacyList, client.MatchingLabels(legacyOwnerLabels)); err != nil {
			return nil, err
		}
	}
	// Fetch resources with new owner labels
	newList := unstructured.UnstructuredList{}
	newList.SetGroupVersionKind(r.groupVersionKind)
	if err := r.client.List(ctx, &newList, client.MatchingLabels(newOwnerLabels)); err != nil {
		return nil, err
	}

	// Convert to pointer slices for merging
	legacyPointers := make([]T, len(legacyList.Items))
	for i := range legacyList.Items {
		// convert unstructured to typed object

		objUnstructured := &legacyList.Items[i]
		obj := new(T)
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(objUnstructured.Object, &obj)
		if err != nil {
			return nil, err
		}
		legacyPointers[i] = *obj
	}

	newPointers := make([]T, len(newList.Items))
	for i := range newList.Items {
		// convert unstructured to typed object
		objUnstructured := &newList.Items[i]
		obj := new(T)
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(objUnstructured.Object, obj)
		if err != nil {
			return nil, err
		}
		newPointers[i] = *obj
	}

	// Merge and deduplicate the results
	return MergeResourceSlices(legacyPointers, newPointers), nil
}

// DeleteAll retrieves and deletes all AccessRule resources with both legacy and new owner labels
func (r *Repository[T]) DeleteAll(ctx context.Context, labeler processing.Labeler) error {
	resources, err := r.GetAll(ctx, labeler)
	if err != nil {
		return err
	}
	for _, resource := range resources {
		log.Log.Info("Removing subresource", r.groupVersionKind.Kind, resource.GetName())
		if err := r.client.Delete(ctx, resource); err != nil {
			return err
		}
	}

	return nil
}
