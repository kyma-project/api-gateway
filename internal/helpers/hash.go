package helpers

import (
	"fmt"
	"strconv"

	"github.com/mitchellh/hashstructure/v2"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
)

func GetAuthorizationPolicyHash(ap securityv1beta1.AuthorizationPolicy) (string, error) {
	hashTo, err := hashstructure.Hash(ap.Spec.Selector, hashstructure.FormatV2, &hashstructure.HashOptions{SlicesAsSets: true})
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s.%s", ap.Namespace, strconv.FormatUint(hashTo, 32)), nil
}
