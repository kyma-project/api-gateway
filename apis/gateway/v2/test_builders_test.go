package v2_test

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type serviceBuilder struct {
	service *corev1.Service
}

func newServiceBuilder() *serviceBuilder {
	return &serviceBuilder{
		service: &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{},
		},
	}
}

func (b *serviceBuilder) withName(name string) *serviceBuilder {
	b.service.Name = name
	return b
}

func (b *serviceBuilder) withNamespace(namespace string) *serviceBuilder {
	b.service.Namespace = namespace
	return b
}

func (b *serviceBuilder) addSelector(key, value string) *serviceBuilder {
	if b.service.Spec.Selector == nil {
		b.service.Spec.Selector = map[string]string{}
	}

	b.service.Spec.Selector[key] = value
	return b
}

func (b *serviceBuilder) build() *corev1.Service {
	return b.service
}
