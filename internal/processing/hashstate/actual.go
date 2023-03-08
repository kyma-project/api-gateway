package hashablestate

import (
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
)

const (
	unhashedValue  = "unhashed"
	unindexedValue = "unindexed"
)

func NewActual() Actual {
	return Actual{
		objects:           make(map[string]*securityv1beta1.AuthorizationPolicy),
		markedForDeletion: []*securityv1beta1.AuthorizationPolicy{},
	}
}

type Actual struct {
	objects map[string]*securityv1beta1.AuthorizationPolicy
	// Objects are marked for deletion for migration reasons. That means objects without required hashing labels must be
	// deleted as we can't reliably compare them. This field might be removed in the future as no objects without hashing
	// labels exist anymore.
	markedForDeletion []*securityv1beta1.AuthorizationPolicy
}

// Add the value to the desired state. If the value does not have the hash and index labels, an error is returned.
func (a *Actual) Add(value *securityv1beta1.AuthorizationPolicy) {
	index, ok := value.Labels[indexLabelName]
	if !ok {
		index = unindexedValue
	}

	hash, ok := value.Labels[hashLabelName]

	if !ok {
		hash = unhashedValue
	}

	// If there are objects that do not have a hash or index set, we cannot reliably compare them, so we delete them because
	// they could be a remnant from an older state without the hash labels.
	if hash == unhashedValue || index == unindexedValue {
		a.markedForDeletion = append(a.markedForDeletion, value)
	} else {
		hashKey := createHashKey(hash, index)
		a.objects[hashKey] = value
	}
}

func (a *Actual) containsHashkey(key string) bool {
	_, ok := a.objects[key]
	return ok
}
