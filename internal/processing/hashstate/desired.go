package hashablestate

import (
	"fmt"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
)

func NewDesired() Desired {
	return Desired{
		objects: make(map[string]*securityv1beta1.AuthorizationPolicy),
	}
}

type Desired struct {
	objects map[string]*securityv1beta1.AuthorizationPolicy
}

// Add the value to the desired state. To protect against the creation of objects in that do not have the required hash
// and index labels, this function returns an error if one of these labels is missing.
func (d *Desired) Add(value *securityv1beta1.AuthorizationPolicy) error {

	index, ok := value.Labels[indexLabelName]
	if !ok {
		return fmt.Errorf("label %s not found on hashable", indexLabelName)
	}

	hash, ok := value.Labels[hashLabelName]
	if !ok {
		return fmt.Errorf("label %s not found on hashable", hashLabelName)
	}

	hashKey := createHashKey(hash, index)
	d.objects[hashKey] = value

	return nil
}

// getObjectsNotIn returns all objects in the desired state where the hash key is not present in the actual state.
func (d *Desired) getObjectsNotIn(actualState Actual) []*securityv1beta1.AuthorizationPolicy {
	var newObjects []*securityv1beta1.AuthorizationPolicy

	for desiredHashKey, desiredObject := range d.objects {

		if !actualState.containsHashkey(desiredHashKey) {
			newObjects = append(newObjects, desiredObject)
		}
	}

	return newObjects
}

func (d *Desired) containsHashkey(key string) bool {
	_, ok := d.objects[key]
	return ok
}

func (d *Desired) String() string {
	return mapKeysToString(d.objects)
}
