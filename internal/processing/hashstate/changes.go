package hashablestate

import (
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
)

// GetChanges returns the changes that need to be applied to reach the desired state by comparing the hash keys
// of the objects in the desired and actual state.
func GetChanges(desiredState Desired, actualState Actual) Changes {
	var toDelete []*securityv1beta1.AuthorizationPolicy
	var toUpdate []*securityv1beta1.AuthorizationPolicy

	for actualObjectHashKey, actualObject := range actualState.objects {

		if desiredState.containsHashkey(actualObjectHashKey) {
			// Since not all fields of the object may be included in the hash key, we need to update the desired changes in the object that is applied.
			// Additionally, we want to make sure that the object is in the expected state and possible manual changes are overwritten.
			actualObject.Spec = *desiredState.objects[actualObjectHashKey].Spec.DeepCopy()
			actualObject.Labels = desiredState.objects[actualObjectHashKey].Labels
			toUpdate = append(toUpdate, actualObject)
		} else {
			// If the actual object is no longer in the desired state we can assume that it was removed and can be deleted.
			toDelete = append(toDelete, actualObject)
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
	Create []*securityv1beta1.AuthorizationPolicy
	Delete []*securityv1beta1.AuthorizationPolicy
	Update []*securityv1beta1.AuthorizationPolicy
}
