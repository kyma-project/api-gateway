package istio

import (
	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	"github.com/kyma-incubator/api-gateway/internal/builders"
	"github.com/kyma-incubator/api-gateway/internal/processing"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
)

// NewVirtualServiceProcessor returns a VirtualServiceProcessor with the desired state handling specific for the Istio handler.
func NewVirtualServiceProcessor(config processing.ReconciliationConfig) processing.VirtualServiceProcessor {
	return processing.VirtualServiceProcessor{
		Creator: virtualServiceCreator{
			oathkeeperSvc:     config.OathkeeperSvc,
			oathkeeperSvcPort: config.OathkeeperSvcPort,
			corsConfig:        config.CorsConfig,
			additionalLabels:  config.AdditionalLabels,
			defaultDomainName: config.DefaultDomainName,
		},
	}
}

type virtualServiceCreator struct {
	oathkeeperSvc     string
	oathkeeperSvcPort uint32
	corsConfig        *processing.CorsConfig
	defaultDomainName string
	additionalLabels  map[string]string
}

func (r virtualServiceCreator) Create(_ *gatewayv1beta1.APIRule) *networkingv1beta1.VirtualService {
	return builders.VirtualService().Get()
}
