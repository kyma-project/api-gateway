package clients

import (
	istioClient "github.com/kyma-incubator/api-gateway/internal/clients/istio"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func New(crClient client.Client) *ExternalCRClients {
	return &ExternalCRClients{
		virtualService: istioClient.ForVirtualService(crClient),
	}
}

//Exposes clients for external Custom Resources (e.g. Istio VirtualService)
type ExternalCRClients struct {
	virtualService *istioClient.VirtualService
}

func (c *ExternalCRClients) ForVirtualService() *istioClient.VirtualService {
	return c.virtualService
}
