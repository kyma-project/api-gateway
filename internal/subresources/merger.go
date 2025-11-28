package subresources

import (
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// MergeResourceSlices combines two slices of Kubernetes resources, deduplicating by name and namespace.
// When duplicates are found, the resource with the higher Generation is preserved as it represents the more recent version.
// This is useful when fetching resources with different label sets (e.g., legacy and new owner labels).
func MergeResourceSlices[T client.Object](slice1, slice2 []T) []T {
	// Deduplicate by name and namespace, keeping the resource with higher generation
	type resourceKey struct {
		name      string
		namespace string
	}
	resourceMap := make(map[resourceKey]T)

	// Process slice1 first
	for _, item := range slice1 {
		key := resourceKey{
			name:      item.GetName(),
			namespace: item.GetNamespace(),
		}
		resourceMap[key] = item
	}

	// Process slice2, replacing items only if they have a higher generation
	for _, item := range slice2 {
		key := resourceKey{
			name:      item.GetName(),
			namespace: item.GetNamespace(),
		}
		if existing, exists := resourceMap[key]; exists {
			// Keep the resource with higher generation
			if item.GetGeneration() > existing.GetGeneration() {
				resourceMap[key] = item
			}
		} else {
			resourceMap[key] = item
		}
	}

	// Convert map to slice
	var result []T
	for _, item := range resourceMap {
		result = append(result, item)
	}

	return result
}
