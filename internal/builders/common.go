package builders

import (
	k8sMeta "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sTypes "k8s.io/apimachinery/pkg/types"
)

// OwnerReference returns builder for k8s.io/apimachinery/pkg/apis/meta/v1/OwnerReference type
func OwnerReference() *ownerReference {
	return &ownerReference{
		value: &k8sMeta.OwnerReference{},
	}
}

type ownerReference struct {
	value *k8sMeta.OwnerReference
}

func (or *ownerReference) From(val *k8sMeta.OwnerReference) *ownerReference {
	or.value = val
	return or
}

func (or *ownerReference) Name(val string) *ownerReference {
	or.value.Name = val
	return or
}

func (or *ownerReference) APIVersion(val string) *ownerReference {
	or.value.APIVersion = val
	return or
}

func (or *ownerReference) Kind(val string) *ownerReference {
	or.value.Kind = val
	return or
}

func (or *ownerReference) UID(val k8sTypes.UID) *ownerReference {
	or.value.UID = val
	return or
}

func (or *ownerReference) Controller(val bool) *ownerReference {
	or.value.Controller = &val
	return or
}

func (or *ownerReference) Get() *k8sMeta.OwnerReference {
	return or.value
}

// ObjectMeta returns builder for k8s.io/apimachinery/pkg/apis/meta/v1/ObjectMeta type
func ObjectMeta() *objectMeta {
	return &objectMeta{
		value: &k8sMeta.ObjectMeta{},
	}
}

type objectMeta struct {
	value *k8sMeta.ObjectMeta
}

func (om *objectMeta) Name(val string) *objectMeta {
	om.value.Name = val
	return om
}

func (om *objectMeta) Namespace(val string) *objectMeta {
	om.value.Namespace = val
	return om
}

func (om *objectMeta) OwnerReference(val *ownerReference) *objectMeta {
	om.value.OwnerReferences = append(om.value.OwnerReferences, *val.Get())
	return om
}

func (om *objectMeta) Get() *k8sMeta.ObjectMeta {
	return om.value
}
