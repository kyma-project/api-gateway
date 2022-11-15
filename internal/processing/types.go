package processing

import (
	"github.com/ory/oathkeeper-maester/api/v1alpha1"
	v1beta12 "istio.io/api/networking/v1beta1"
	"istio.io/client-go/pkg/apis/networking/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ObjToPatch struct {
	action string
	obj    client.Object
}

func NewCreateObjectAction(obj client.Object) *ObjToPatch {
	return &ObjToPatch{
		action: "create",
		obj:    obj,
	}
}

func NewUpdateObjectAction(obj client.Object) *ObjToPatch {
	return &ObjToPatch{
		action: "update",
		obj:    obj,
	}
}

func NewDeleteObjectAction(obj client.Object) *ObjToPatch {
	return &ObjToPatch{
		action: "delete",
		obj:    obj,
	}
}

// State represents desired or actual state of Istio Virtual Services and Oathkeeper Rules
type State struct {
	virtualService *v1beta1.VirtualService
	accessRules    map[string]*v1alpha1.Rule
}

// CorsConfig is an internal representation of v1alpha3.CorsPolicy object
type CorsConfig struct {
	AllowOrigins []*v1beta12.StringMatch
	AllowMethods []string
	AllowHeaders []string
}
