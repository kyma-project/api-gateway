package hashbasedstate

import (
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetChanges returns the changes that need to be applied to reach the desired state by comparing the hash keys
// of the objects in the desired and actual state.
func GetChanges(desiredState Desired, actualState Actual) Changes {
	var toDelete []client.Object
	var toUpdate []client.Object

	for actualHashKey, actual := range actualState.hashables {

		if desiredState.containsHashkey(actualHashKey) {
			// Since not all fields of the object may be included in the hash key, we need to update the desired changes in the object that is applied.
			// Additionally, we want to make sure that the object is in the expected state and possible manual changes are overwritten.
			updated := actual.updateSpec(desiredState.hashables[actualHashKey])
			toUpdate = append(toUpdate, updated)
		} else {
			// If the actual object is no longer in the desired state we can assume that it was removed and can be deleted.
			toDelete = append(toDelete, actual.ToObject())
		}
	}

	toDelete = append(toDelete, actualState.markedForDeletion...)

	// We know that all objects that are in the desired state but not in the actual state must be new objects and need to be created.
	toCreate := desiredState.getObjectsNotIn(actualState)

	return Changes{
		Create: toCreate,
		Delete: toDelete,
		Update: toUpdate,
	}
}

// Changes that need to be applied to reach the desired state
type Changes struct {
	Create []client.Object
	Delete []client.Object
	Update []client.Object
}

func (c Changes) String() string {
	toCreate := make([]string, len(c.Create))
	for _, ap := range c.Create {
		toCreate = append(toCreate, ap.GetName())
	}

	toUpdate := make([]string, len(c.Update))
	for _, ap := range c.Update {
		toUpdate = append(toUpdate, ap.GetName())
	}

	toDelete := make([]string, len(c.Delete))
	for _, ap := range c.Delete {
		toDelete = append(toDelete, ap.GetName())
	}

	return fmt.Sprintf("Create: %s; Delete: %s; Update: %s", toCreate, toDelete, toUpdate)
}

type Hashable interface {
	index() (string, bool)
	hash() (string, bool)
	updateSpec(h Hashable) client.Object
	ToObject() client.Object
}
