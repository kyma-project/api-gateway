package dependencies

import (
	"context"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Dependencies interface {
	AreAvailable(ctx context.Context, k8sClient client.Client) (string, error)
}

func ApiGateway() Dependencies {
	return &dependencies{
		CRDNames: []string{
			"gateways.networking.istio.io",
			"virtualservices.networking.istio.io",
		},
	}
}

func Gardener() Dependencies {
	return &dependencies{
		CRDNames: []string{
			"dnsentries.dns.gardener.cloud",
			"certificates.cert.gardener.cloud",
		},
	}
}

func RateLimit() Dependencies {
	return &dependencies{
		CRDNames: []string{
			"envoyfilters.networking.istio.io",
		},
	}
}

func APIRule() Dependencies {
	return &dependencies{
		CRDNames: []string{
			"virtualservices.networking.istio.io",
			"rules.oathkeeper.ory.sh",
			"authorizationpolicies.security.istio.io",
			"requestauthentications.security.istio.io",
		},
	}
}

type dependencies struct {
	CRDNames []string
}

// AreAvailable checks whether pre-requisite CRDs are present on the cluster. It returns the name of the not found CRD and an error if it is not found.
func (d dependencies) AreAvailable(ctx context.Context, k8sClient client.Client) (string, error) {
	for _, name := range d.CRDNames {
		err := k8sClient.Get(ctx, types.NamespacedName{Name: name}, &apiextensionsv1.CustomResourceDefinition{})
		if err != nil {
			return name, err
		}
	}
	return "", nil
}
