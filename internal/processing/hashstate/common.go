// Package hashablestate provides types and functions to compare objects by a hash and a position in a yaml sequence.
//
// The hashablestate package should only be used objects unmarshalled from a yaml sequences as they have a defined order.
// The comparison is based on labels for a hash and an index put on a kubernetes object. The hash label holds the hash that
// represents the object and the index value holds the position of the object in the sequence. Both of this information is
// then used to identify if an object was changed, removed or newly added.
package hashablestate

import (
	"fmt"
	"github.com/kyma-project/api-gateway/internal/helpers"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	"strconv"
	"strings"
)

const (
	hashLabelName  string = "gateway.kyma-project.io/hash"
	indexLabelName string = "gateway.kyma-project.io/index"
)

// createHashKey returns a key in the format of "hash:index".
func createHashKey(hashValue, indexValue string) string {
	return fmt.Sprintf("%s:%s", hashValue, indexValue)
}

// AddHashingLabels adds labels to the desired object to be able to compare it with the actual object in the cluster.
func AddHashingLabels(ap *securityv1beta1.AuthorizationPolicy, indexInYaml int) error {

	// We add the index as a label to be able to compare later if something has been changed or not. We can make the assumption
	// that the index is the same if nothing has changed, since authorizations in yaml are a sequence and the order for sequences
	// is static (https://yaml.org/spec/1.2/spec.html#id2764044).
	ap.Labels[indexLabelName] = strconv.Itoa(indexInYaml)

	hash, err := helpers.GetAuthorizationPolicyHash(ap)
	if err != nil {
		return err
	}

	ap.Labels[hashLabelName] = hash

	return nil
}

func mapKeysToString(m map[string]*securityv1beta1.AuthorizationPolicy) string {
	keys := make([]string, len(m))
	for key, _ := range m {
		keys = append(keys, key)
	}

	return fmt.Sprintf("Keys in state: %s", strings.Join(keys, ","))
}
