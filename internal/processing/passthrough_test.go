package processing

import (
	"testing"

	gatewayv2alpha1 "github.com/kyma-incubator/api-gateway/api/v2alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var (
	apiName                 = "some-api"
	apiUID        types.UID = "eab0f1c8-c417-11e9-bf11-4ac644044351"
	apiNamespace            = "some-namespace"
	apiAPIVersion           = "gateway.kyma-project.io/v2alpha1"
	apiKind                 = "Gate"
	apiGateway              = "some-gateway"
	serviceName             = "example-service"
	serviceHost             = "myService.myDomain.com"
	servicePort   uint32    = 8080
)

func TestGenerateVirtualService(t *testing.T) {
	assert := assert.New(t)

	exampleAPI := &gatewayv2alpha1.Gate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      apiName,
			UID:       apiUID,
			Namespace: apiNamespace,
		},
		TypeMeta: metav1.TypeMeta{
			APIVersion: apiAPIVersion,
			Kind:       apiKind,
		},
		Spec: gatewayv2alpha1.GateSpec{
			Gateway: &apiGateway,
			Service: &gatewayv2alpha1.Service{
				Name: &serviceName,
				Host: &serviceHost,
				Port: &servicePort,
			},
		},
	}

	vs := generateVirtualService(exampleAPI, serviceName+"."+apiNamespace+".svc.cluster.local", servicePort, "/.*")

	assert.Equal(len(vs.Spec.Gateways), 1)
	assert.Equal(vs.Spec.Gateways[0], apiGateway)

	assert.Equal(len(vs.Spec.Hosts), 1)
	assert.Equal(vs.Spec.Hosts[0], serviceHost)

	assert.Equal(len(vs.Spec.HTTP), 1)
	assert.Equal(len(vs.Spec.HTTP[0].Route), 1)
	assert.Equal(len(vs.Spec.HTTP[0].Match), 1)
	assert.Equal(vs.Spec.HTTP[0].Route[0].Destination.Host, serviceName+"."+apiNamespace+".svc.cluster.local")
	assert.Equal(vs.Spec.HTTP[0].Route[0].Destination.Port.Number, servicePort)
	assert.Equal(vs.Spec.HTTP[0].Match[0].URI.Regex, "/.*")

	assert.Equal(vs.ObjectMeta.Name, apiName+"-"+serviceName)
	assert.Equal(vs.ObjectMeta.Namespace, apiNamespace)

	assert.Equal(vs.ObjectMeta.OwnerReferences[0].APIVersion, apiAPIVersion)
	assert.Equal(vs.ObjectMeta.OwnerReferences[0].Kind, apiKind)
	assert.Equal(vs.ObjectMeta.OwnerReferences[0].Name, apiName)
	assert.Equal(vs.ObjectMeta.OwnerReferences[0].UID, apiUID)

}
