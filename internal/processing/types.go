package processing

import (
	"context"
	"github.com/go-logr/logr"
	v1beta12 "istio.io/api/networking/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ObjectChange struct {
	Action string
	Obj    client.Object
}

func NewObjectCreateAction(obj client.Object) *ObjectChange {
	return &ObjectChange{
		Action: "create",
		Obj:    obj,
	}
}

func NewObjectUpdateAction(obj client.Object) *ObjectChange {
	return &ObjectChange{
		Action: "update",
		Obj:    obj,
	}
}

func NewObjectDeleteAction(obj client.Object) *ObjectChange {
	return &ObjectChange{
		Action: "delete",
		Obj:    obj,
	}
}

// CorsConfig is an internal representation of v1alpha3.CorsPolicy object
type CorsConfig struct {
	AllowOrigins []*v1beta12.StringMatch
	AllowMethods []string
	AllowHeaders []string
}

type ReconciliationConfig struct {
	Client            client.Client
	Ctx               context.Context
	Logger            logr.Logger
	OathkeeperSvc     string
	OathkeeperSvcPort uint32
	CorsConfig        *CorsConfig
	AdditionalLabels  map[string]string
	DefaultDomainName string
	ServiceBlockList  map[string][]string
	DomainAllowList   []string
	HostBlockList     []string
}
