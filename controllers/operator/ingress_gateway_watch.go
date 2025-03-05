package operator

import (
	"context"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type IngressGatewayEventHandler struct{}

const (
	name = "istio-ingressgateway"
	ns   = "istio-system"
)

func (i IngressGatewayEventHandler) Create(_ context.Context, e event.TypedCreateEvent[client.Object], w workqueue.TypedRateLimitingInterface[reconcile.Request]) {
	if e.Object.GetName() != name || e.Object.GetNamespace() != ns {
		return
	}

	w.Add(reconcile.Request{NamespacedName: types.NamespacedName{Namespace: ns, Name: name}})
}

func (i IngressGatewayEventHandler) Update(_ context.Context, e event.TypedUpdateEvent[client.Object], w workqueue.TypedRateLimitingInterface[reconcile.Request]) {
	if e.ObjectOld.GetName() != name || e.ObjectOld.GetNamespace() != ns {
		return
	}

	w.Add(reconcile.Request{NamespacedName: types.NamespacedName{Namespace: ns, Name: name}})
}

func (i IngressGatewayEventHandler) Delete(_ context.Context, e event.TypedDeleteEvent[client.Object], w workqueue.TypedRateLimitingInterface[reconcile.Request]) {
	if e.Object.GetName() != name || e.Object.GetNamespace() != ns {
		return
	}

	w.Add(reconcile.Request{NamespacedName: types.NamespacedName{Namespace: ns, Name: name}})
}

func (i IngressGatewayEventHandler) Generic(_ context.Context, e event.TypedGenericEvent[client.Object], w workqueue.TypedRateLimitingInterface[reconcile.Request]) {
	if e.Object.GetName() != name || e.Object.GetNamespace() != ns {
		return
	}

	w.Add(reconcile.Request{NamespacedName: types.NamespacedName{Namespace: ns, Name: name}})
}
