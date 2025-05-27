package hashbasedstate

import (
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewDesired() Desired {
	return Desired{
		hashables: make(map[string]Hashable),
	}
}

type Desired struct {
	hashables map[string]Hashable
}

// Add the value to the desired state. Since setting the hashing labels is decoupled from adding objects to the state we need to protect against the creation of objects in that do not have the required hash
// and index labels, this function returns an error if one of these labels is missing.
func (d *Desired) Add(h Hashable) error {
	index, ok := h.index()
	if !ok {
		return fmt.Errorf("label %s not found on hashable", indexLabelName)
	}

	hash, ok := h.hash()
	if !ok {
		return fmt.Errorf("label %s not found on hashable", hashLabelName)
	}

	hashKey := createHashKey(hash, index)
	d.hashables[hashKey] = h

	return nil
}

// getObjectsNotIn returns all objects in the desired state where the hash key is not present in the actual state.
func (d *Desired) getObjectsNotIn(actualState Actual) []client.Object {
	var newObjects []client.Object

	for desiredHashKey, desiredHashable := range d.hashables {
		if !actualState.containsHashkey(desiredHashKey) {
			newObjects = append(newObjects, desiredHashable.ToObject())
		}
	}

	return newObjects
}

func (d *Desired) containsHashkey(key string) bool {
	_, ok := d.hashables[key]
	return ok
}
