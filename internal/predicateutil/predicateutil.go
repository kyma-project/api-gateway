package predicateutil

import (
	"slices"

	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const (
	CreateEvent EventType = iota
	UpdateEvent
	DeleteEvent
	GenericEvent
)

// EventType represents a controller-runtime event kind (Create, Update, Delete, or Generic).
type EventType int

// ForEventTypes returns a predicate that only allows the specified event types to pass through.
func ForEventTypes(eventType ...EventType) predicate.Predicate {
	has := func(t EventType) bool {
		return slices.Contains(eventType, t)
	}
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return has(CreateEvent)
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			return has(UpdateEvent)
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return has(DeleteEvent)
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return has(GenericEvent)
		}}
}
