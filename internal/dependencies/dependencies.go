package dependencies

import (
	"context"
	"fmt"
	"github.com/kyma-project/api-gateway/controllers"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Dependencies interface {
	Check(context.Context, client.Client) controllers.Status
}

func NewAPIGateway() *dependecies {
	return &dependecies{
		CRDNames: []string{
			"gateways.networking.istio.io",
			"virtualservices.networking.istio.io",
		},
	}
}

func NewGardenerAPIGateway() *dependecies {
	return &dependecies{
		CRDNames: []string{
			"gateways.networking.istio.io",
			"virtualservices.networking.istio.io",
			"dnsentries.dns.gardener.cloud",
			"certificates.cert.gardener.cloud",
		},
	}
}

func NewAPIRule() *dependecies {
	return &dependecies{
		CRDNames: []string{
			"virtualservices.networking.istio.io",
			"rules.oathkeeper.ory.sh",
			"authorizationpolicies.security.istio.io",
			"requestauthentications.security.istio.io",
		},
	}
}

type dependecies struct {
	CRDNames []string
}

func (d dependecies) Check(ctx context.Context, k8sClient client.Client) controllers.Status {
	for _, name := range d.CRDNames {
		err := k8sClient.Get(ctx, types.NamespacedName{Name: name}, &apiextensionsv1.CustomResourceDefinition{})
		if err != nil {
			if k8serrors.IsNotFound(err) {
				return controllers.WarningStatus(err, fmt.Sprintf("CRD %s is not present. Make sure to install required dependencies for the component", name))
			} else {
				return controllers.ErrorStatus(err, "Error happened during discovering dependencies")
			}
		}
	}
	return controllers.ReadyStatus()
}
