package hashbasedstate

import (
	"fmt"
	"github.com/mitchellh/hashstructure/v2"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
)

// AddLabelsToAuthorizationPolicy adds hashing labels.
func AddLabelsToAuthorizationPolicy(ap *securityv1beta1.AuthorizationPolicy, indexInYaml int) error {

	hash, err := GetAuthorizationPolicyHash(ap)
	if err != nil {
		return err
	}

	addHashingLabels(ap, hash, indexInYaml)

	return nil
}

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

type AuthorizationPolicyHashable struct {
	ap *securityv1beta1.AuthorizationPolicy
}

func (a *AuthorizationPolicyHashable) value() interface{ client.Object } {
	return a.ap
}

func (a *AuthorizationPolicyHashable) updateWith(o Hashable) {
	ap := o.value().(*securityv1beta1.AuthorizationPolicy)
	a.ap.Spec = *ap.Spec.DeepCopy()
	a.ap.Labels = ap.Labels
}

func NewAuthorizationPolicy(ap *securityv1beta1.AuthorizationPolicy) AuthorizationPolicyHashable {
	return AuthorizationPolicyHashable{ap}
}
