package dependencies

import (
	"context"
	"github.com/kyma-project/api-gateway/controllers"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Dependencies interface {
	AreAvailable(context.Context, client.Client) controllers.Status
}

func ApiGateway() *dependecies {
	return &dependecies{
		CRDNames: []string{
			"gateways.networking.istio.io",
			"virtualservices.networking.istio.io",
		},
	}
}

func GardenerAPIGateway() *dependecies {
	return &dependecies{
		CRDNames: []string{
			"gateways.networking.istio.io",
			"virtualservices.networking.istio.io",
			"dnsentries.dns.gardener.cloud",
			"certificates.cert.gardener.cloud",
		},
	}
}

func APIRule() *dependecies {
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

// AreAvailable checks whether pre-requisite CRDs are present on the cluster. It returns the name of the not found CRD and an error if it is not found.
func (d dependecies) AreAvailable(ctx context.Context, k8sClient client.Client) (string, error) {
	for _, name := range d.CRDNames {
		err := k8sClient.Get(ctx, types.NamespacedName{Name: name}, &apiextensionsv1.CustomResourceDefinition{})
		if err != nil {
			return name, err
		}
	}
	return "", nil
}
