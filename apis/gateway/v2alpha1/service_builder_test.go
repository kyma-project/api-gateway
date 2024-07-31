package v2alpha1_test

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ServiceBuilder struct {
	service *corev1.Service
}

func NewServiceBuilder() *ServiceBuilder {
	return &ServiceBuilder{
		service: &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{},
		},
	}
}

func (b *ServiceBuilder) SetName(name string) *ServiceBuilder {
	b.service.ObjectMeta.Name = name
	return b
}

func (b *ServiceBuilder) SetNamespace(namespace string) *ServiceBuilder {
	b.service.ObjectMeta.Namespace = namespace
	return b
}

func (b *ServiceBuilder) AddSelector(key, value string) *ServiceBuilder {
	if b.service.Spec.Selector == nil {
		b.service.Spec.Selector = map[string]string{}
	}

	b.service.Spec.Selector[key] = value
	return b
}

func (b *ServiceBuilder) Build() *corev1.Service {
	return b.service
}
