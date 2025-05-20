// Package hashbasedstate provides types and functions to compare objects by a hash and a position in a yaml sequence.
//
// The hashbasedstate package should only be used objects unmarshalled from a yaml sequences as they have a defined order.
// The comparison is based on labels for a hash and an index put on a kubernetes object. The hash label holds the hash that
// represents the object and the index value holds the position of the object in the sequence. Both of this information is
// then used to identify if an object was changed, removed or newly added.
//
// Since this comparison is based on the order of objects, it means that adding a new object before an existing object in the
// sequence triggers an update for all the following objects, since their position in the sequence has changed.
package hashbasedstate

import (
	"fmt"
	"strconv"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	hashLabelName  string = "gateway.kyma-project.io/hash"
	indexLabelName string = "gateway.kyma-project.io/index"
)

// createHashKey returns a key in the format of "hash:index".
func createHashKey(hashValue, indexValue string) string {
	return fmt.Sprintf("%s:%s", hashValue, indexValue)
}

// addHashingLabels adds labels to the desired object to be able to compare it with the actual object in the cluster.
func addHashingLabels(o client.Object, hash string, indexInYaml int) {

	// We add the index as a label to be able to compare later if something has been changed or not. We can make the assumption
	// that the index is the same if nothing has changed, since authorizations in yaml are a sequence and the order for sequences
	// is static (https://yaml.org/spec/1.2/spec.html#id2764044).
	o.GetLabels()[indexLabelName] = strconv.Itoa(indexInYaml)
	o.GetLabels()[hashLabelName] = hash

}
