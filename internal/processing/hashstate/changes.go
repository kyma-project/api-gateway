package hashablestate

import (
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
)

func GetChanges(desiredState Desired, actualState Actual) StateChanges {
	var toDelete []*securityv1beta1.AuthorizationPolicy
	var toUpdate []*securityv1beta1.AuthorizationPolicy

	for actualObjectHashKey, actualObject := range actualState.objects {

		if desiredState.containsHashkey(actualObjectHashKey) {
			// If the object is still the same, we don't really need to update it. Nevertheless, we want to add it for an
			// update to make sure that the object is in the expected state and possible manual changes are overwritten.
			toUpdate = append(toUpdate, actualObject)
		} else {
			// If the actual object is no longer in the desired state we can assume that it was removed and can be deleted.
			toDelete = append(toDelete, actualObject)
		}
	}

	toDelete = append(toDelete, actualState.markedForDeletion...)

	// We know that all objects that are in the desired state but not in the actual state must be new objects and need to be created.
	toCreate := desiredState.getObjectsNotIn(actualState)

	return StateChanges{
		Create: toCreate,
		Delete: toDelete,
		Update: toUpdate,
	}
}

type StateChanges struct {
	Create []*securityv1beta1.AuthorizationPolicy
	Delete []*securityv1beta1.AuthorizationPolicy
	Update []*securityv1beta1.AuthorizationPolicy
}
