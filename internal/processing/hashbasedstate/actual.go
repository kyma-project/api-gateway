package hashbasedstate

import "sigs.k8s.io/controller-runtime/pkg/client"

const (
	unhashedValue  = "unhashed"
	unindexedValue = "unindexed"
)

func NewActual() Actual {
	return Actual{
		hashables:         make(map[string]Hashable),
		markedForDeletion: []client.Object{},
	}
}

type Actual struct {
	hashables map[string]Hashable
	// Objects are marked for deletion for migration reasons. That means objects without required hashing labels must be
	// deleted as we can't reliably compare them. This field might be removed in the future as no objects without hashing
	// labels exist anymore.
	markedForDeletion []client.Object
}

// Add the value to the actual state. If the value does not have the hash and index labels, an error is returned.
func (a *Actual) Add(hashable Hashable) {
	index, ok := hashable.index()
	if !ok {
		index = unindexedValue
	}

	hash, ok := hashable.hash()

	if !ok {
		hash = unhashedValue
	}

	// If there are objects that do not have a hash or index set, we cannot reliably compare them, so we delete them because
	// they could be a remnant from an older state without the hash labels.
	if hash == unhashedValue || index == unindexedValue {
		a.markedForDeletion = append(a.markedForDeletion, hashable.ToObject())
	} else {
		hashKey := createHashKey(hash, index)
		a.hashables[hashKey] = hashable
	}
}

func (a *Actual) containsHashkey(key string) bool {
	_, ok := a.hashables[key]
	return ok
}
