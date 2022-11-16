package processing

import (
	"context"
	"github.com/go-logr/logr"
	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	"github.com/kyma-incubator/api-gateway/controllers"
	"github.com/ory/oathkeeper-maester/api/v1alpha1"
	v1beta12 "istio.io/api/networking/v1beta1"
	"istio.io/client-go/pkg/apis/networking/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ObjectChange struct {
	action string
	obj    client.Object
}

func NewObjectCreateAction(obj client.Object) *ObjectChange {
	return &ObjectChange{
		action: "create",
		obj:    obj,
	}
}

func NewObjectUpdateAction(obj client.Object) *ObjectChange {
	return &ObjectChange{
		action: "update",
		obj:    obj,
	}
}

func NewObjectDeleteAction(obj client.Object) *ObjectChange {
	return &ObjectChange{
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

func NewReconciliationConfig(ctx context.Context, r *controllers.APIRuleReconciler) ReconciliationConfig {
	return ReconciliationConfig{
		Client:            r.Client,
		Ctx:               ctx,
		Logger:            r.Log,
		OathkeeperSvc:     r.OathkeeperSvc,
		OathkeeperSvcPort: r.OathkeeperSvcPort,
		CorsConfig:        r.CorsConfig,
		AdditionalLabels:  r.GeneratedObjectsLabels,
		DefaultDomainName: r.DefaultDomainName,
		ServiceBlockList:  r.ServiceBlockList,
		DomainAllowList:   r.DomainAllowList,
		HostBlockList:     r.HostBlockList,
	}
}

type ReconciliationProcessor interface {
	EvaluateReconciliation(*gatewayv1beta1.APIRule) ([]*ObjectChange, gatewayv1beta1.StatusCode, error)
}
