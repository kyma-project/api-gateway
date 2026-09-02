package gateway

import (
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type isApiGatewayConfigMapPredicate struct {
	Log logr.Logger
	predicate.Funcs
}

func (p isApiGatewayConfigMapPredicate) Create(e event.CreateEvent) bool {
	return p.Generic(event.GenericEvent{
		Object: e.Object,
	})
}

func (p isApiGatewayConfigMapPredicate) Delete(e event.DeleteEvent) bool {
	return p.Generic(event.GenericEvent{
		Object: e.Object,
	})
}

func (p isApiGatewayConfigMapPredicate) Update(e event.UpdateEvent) bool {
	return p.Generic(event.GenericEvent{
		Object: e.ObjectNew,
	})
}

func (p isApiGatewayConfigMapPredicate) Generic(e event.GenericEvent) bool {
	if e.Object == nil {
		p.Log.Error(nil, "Generic event has no object", "event", e)
		return false
	}
	configMap, okCM := e.Object.(*corev1.ConfigMap)
	return okCM && configMap.GetNamespace() == configMapNamespace && configMap.GetName() == configMapName
}
