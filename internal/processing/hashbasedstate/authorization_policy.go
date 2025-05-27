package hashbasedstate

import (
	"fmt"
	"strconv"

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
	hashService, err := hashstructure.Hash(ap.Spec.GetSelector(), hashstructure.FormatV2, &hashstructure.HashOptions{SlicesAsSets: true})
	if err != nil {
		return "", err
	}

	var hashTo uint64
	if len(ap.Spec.GetRules()) > 0 && ap.Spec.Rules[0].To != nil {
		hash, err := hashstructure.Hash(ap.Spec.GetRules()[0].GetTo(), hashstructure.FormatV2, &hashstructure.HashOptions{SlicesAsSets: true})
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

func (a *AuthorizationPolicyHashable) ToObject() client.Object {
	return a.ap
}

func (a *AuthorizationPolicyHashable) hash() (string, bool) {
	val, ok := a.ap.Labels[hashLabelName]
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
