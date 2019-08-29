package clients

import (
	istioClient "github.com/kyma-incubator/api-gateway/internal/clients/istio"
	oryClient "github.com/kyma-incubator/api-gateway/internal/clients/ory"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//New .
func New(crClient client.Client) *ExternalCRClients {
	return &ExternalCRClients{
		virtualService:       istioClient.ForVirtualService(crClient),
		authenticationPolicy: istioClient.ForAuthenticationPolicy(crClient),
		accessRule:           oryClient.ForAccessRule(crClient),
	}
}

//ExternalCRClients exposes clients for external Custom Resources (e.g. Istio VirtualService)
type ExternalCRClients struct {
	virtualService       *istioClient.VirtualService
	authenticationPolicy *istioClient.AuthenticationPolicy
	accessRule           *oryClient.AccessRule
}

//ForVirtualService .
func (c *ExternalCRClients) ForVirtualService() *istioClient.VirtualService {
	return c.virtualService
}

//ForAuthenticationPolicy .
func (c *ExternalCRClients) ForAuthenticationPolicy() *istioClient.AuthenticationPolicy {
	return c.authenticationPolicy
}

//ForAccessRule .
func (c *ExternalCRClients) ForAccessRule() *oryClient.AccessRule {
	return c.accessRule
}
