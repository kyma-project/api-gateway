package helpers

import (
	"strconv"

	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	"github.com/mitchellh/hashstructure/v2"
)

func GetAuthorizationPolicyHash(ap securityv1beta1.AuthorizationPolicy) (string, error) {
	hash, err := hashstructure.Hash(ap.Spec, hashstructure.FormatV2, &hashstructure.HashOptions{SlicesAsSets: true})
	if err != nil {
		return "", err
	}
	return strconv.FormatUint(hash, 32), nil
}
