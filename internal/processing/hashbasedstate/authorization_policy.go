package hashbasedstate

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/mitchellh/hashstructure/v2"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

	hashNamespace, err := hashstructure.Hash(ap.Namespace, hashstructure.FormatV2, nil)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s.%s.%s", strconv.FormatUint(hashNamespace, 36), strconv.FormatUint(hashService, 32), strconv.FormatUint(hashTo, 32)), nil
}

type AuthorizationPolicyHashable struct {
	ap *securityv1beta1.AuthorizationPolicy
}

func (a *AuthorizationPolicyHashable) ToObject() client.Object {
	return a.ap
}

func (a *AuthorizationPolicyHashable) hash() (string, bool) {
	val, ok := a.ap.Labels[hashLabelName]
	if !ok {
		return "", false
	}

	parts := strings.Split(val, ".")
	if len(parts) != 3 {
		return "", false
	}

	if parts[0] == a.ap.Namespace {
		hashNamespace, err := hashstructure.Hash(a.ap.Namespace, hashstructure.FormatV2, nil)
		if err != nil {
			return "", false
		}

		return fmt.Sprintf("%s.%s.%s", strconv.FormatUint(hashNamespace, 36), parts[1], parts[2]), true
	}
	return val, ok
}

func (a *AuthorizationPolicyHashable) index() (string, bool) {
	val, ok := a.ap.Labels[indexLabelName]
	return val, ok
}

func (a *AuthorizationPolicyHashable) updateSpec(h Hashable) client.Object {
	obj := h.ToObject()
	obj.SetResourceVersion(a.ap.ResourceVersion)
	obj.SetAnnotations(a.ap.Annotations)
	obj.SetNamespace(a.ap.Namespace)
	obj.SetName(a.ap.Name)
	return obj
}

func NewAuthorizationPolicy(ap *securityv1beta1.AuthorizationPolicy) AuthorizationPolicyHashable {
	return AuthorizationPolicyHashable{ap}
}
