package processing

import (
	v1beta1 "istio.io/api/networking/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Action int

const (
	create Action = iota
	update
	delete
)

func (s Action) String() string {
	switch s {
	case create:
		return "create"
	case update:
		return "update"
	case delete:
		return "delete"
	}
	return "unknown"
}

type ObjectChange struct {
	Action Action
	Obj    client.Object
}

func NewObjectCreateAction(obj client.Object) *ObjectChange {
	return &ObjectChange{
		Action: create,
		Obj:    obj,
	}
}

func NewObjectUpdateAction(obj client.Object) *ObjectChange {
	return &ObjectChange{
		Action: update,
		Obj:    obj,
	}
}

func NewObjectDeleteAction(obj client.Object) *ObjectChange {
	return &ObjectChange{
		Action: delete,
		Obj:    obj,
	}
}

// CorsConfig is an internal representation of v1alpha3.CorsPolicy object.
type CorsConfig struct {
	AllowOrigins []*v1beta1.StringMatch
	AllowMethods []string
	AllowHeaders []string
}

type ReconciliationConfig struct {
	OathkeeperSvc     string
	OathkeeperSvcPort uint32
	CorsConfig        *CorsConfig
	DefaultDomainName string
}
