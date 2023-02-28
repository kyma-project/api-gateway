package helpers

import (
	"fmt"
	"strconv"

	"github.com/mitchellh/hashstructure/v2"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
)

func GetAuthorizationPolicyHash(ap *securityv1beta1.AuthorizationPolicy) (string, error) {
	hashService, err := hashstructure.Hash(ap.Spec.Selector, hashstructure.FormatV2, &hashstructure.HashOptions{SlicesAsSets: true})
	if err != nil {
		return "", err
	}

	var hashTo uint64
	if len(ap.Spec.Rules) > 0 && ap.Spec.Rules[0].To != nil {
		hash, err := hashstructure.Hash(ap.Spec.Rules[0].To, hashstructure.FormatV2, &hashstructure.HashOptions{SlicesAsSets: true})
		if err != nil {
			return "", err
		}
		hashTo = hash
	}

	return fmt.Sprintf("%s.%s.%s", ap.Namespace, strconv.FormatUint(hashService, 32), strconv.FormatUint(hashTo, 32)), nil
}
