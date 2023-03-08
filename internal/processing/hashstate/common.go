package hashablestate

import (
	"fmt"
	"github.com/kyma-project/api-gateway/internal/helpers"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	"strconv"
)

const (
	hashLabelName  string = "gateway.kyma-project.io/hash"
	indexLabelName string = "gateway.kyma-project.io/index"
)

func createHashKey(hashValue, indexValue string) string {
	return fmt.Sprintf("%s:%s", hashValue, indexValue)
}

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
