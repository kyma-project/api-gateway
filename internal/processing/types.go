package processing

import (
	"context"
	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	"github.com/ory/oathkeeper-maester/api/v1alpha1"
	v1beta12 "istio.io/api/networking/v1beta1"
	"istio.io/client-go/pkg/apis/networking/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ReconciliationCommand struct {
	action string
	obj    client.Object
}

func NewCreateCommand(obj client.Object) *ReconciliationCommand {
	return &ReconciliationCommand{
		action: "create",
		obj:    obj,
	}
}

func NewUpdateCommand(obj client.Object) *ReconciliationCommand {
	return &ReconciliationCommand{
		action: "update",
		obj:    obj,
	}
}

func NewDeleteCommand(obj client.Object) *ReconciliationCommand {
	return &ReconciliationCommand{
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
	client            client.Client
	ctx               context.Context
	oathkeeperSvc     string
	oathkeeperSvcPort uint32
	corsConfig        *CorsConfig
	additionalLabels  map[string]string
	defaultDomainName string
}

func NewReconciliationConfig(client client.Client,
	ctx context.Context,
	oathkeeperSvc string,
	oathkeeperSvcPort uint32,
	corsConfig *CorsConfig,
	additionalLabels map[string]string,
	defaultDomainName string) ReconciliationConfig {
	return ReconciliationConfig{
		client:            client,
		ctx:               ctx,
		oathkeeperSvc:     oathkeeperSvc,
		oathkeeperSvcPort: oathkeeperSvcPort,
		corsConfig:        corsConfig,
		additionalLabels:  additionalLabels,
		defaultDomainName: defaultDomainName,
	}
}

type ReconciliationProcessor interface {
	evaluateReconciliation(*gatewayv1beta1.APIRule) ([]*ReconciliationCommand, gatewayv1beta1.StatusCode, error)
}
