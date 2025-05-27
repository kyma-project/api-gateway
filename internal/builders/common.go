package builders

import (
	k8sMeta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ObjectMeta returns builder for k8s.io/apimachinery/pkg/apis/meta/v1/ObjectMeta type.
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

func (om *objectMeta) Get() *k8sMeta.ObjectMeta {
	return om.value
}
