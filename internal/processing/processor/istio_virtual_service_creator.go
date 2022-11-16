package processor

import (
	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	"github.com/kyma-incubator/api-gateway/internal/builders"
	"github.com/kyma-incubator/api-gateway/internal/processing"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
)

type IstioVirtualServiceCreator struct {
	oathkeeperSvc     string
	oathkeeperSvcPort uint32
	corsConfig        *processing.CorsConfig
	defaultDomainName string
	additionalLabels  map[string]string
}

func (r IstioVirtualServiceCreator) create(_ *gatewayv1beta1.APIRule) *networkingv1beta1.VirtualService {
	return builders.VirtualService().Get()
}
