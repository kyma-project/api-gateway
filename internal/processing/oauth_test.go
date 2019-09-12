package processing

import (
	"k8s.io/apimachinery/pkg/runtime"
	"testing"

	gatewayv2alpha1 "github.com/kyma-incubator/api-gateway/api/v2alpha1"
	rulev1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func TestOauthGenerateVirtualService(t *testing.T) {
	assert := assert.New(t)

	gate := getGate()
	vs := generateVirtualService(gate, "test-oathkeeper", 4455, gate.Spec.Paths[0].Path)

	assert.Equal(len(vs.Spec.Gateways), 1)
	assert.Equal(vs.Spec.Gateways[0], apiGateway)

	assert.Equal(len(vs.Spec.Hosts), 1)
	assert.Equal(vs.Spec.Hosts[0], serviceHost)

	assert.Equal(len(vs.Spec.HTTP), 1)
	assert.Equal(len(vs.Spec.HTTP[0].Route), 1)
	assert.Equal(len(vs.Spec.HTTP[0].Match), 1)
	assert.Equal(vs.Spec.HTTP[0].Route[0].Destination.Host, "test-oathkeeper")
	assert.Equal(int(vs.Spec.HTTP[0].Route[0].Destination.Port.Number), 4455)
	assert.Equal(vs.Spec.HTTP[0].Match[0].URI.Regex, "/foo")

	assert.Equal(vs.ObjectMeta.Name, "test-gate-test-service")
	assert.Equal(vs.ObjectMeta.Namespace, "test-namespace")

	assert.Equal(vs.ObjectMeta.OwnerReferences[0].APIVersion, "gateway.kyma-project.io/v2alpha1")
	assert.Equal(vs.ObjectMeta.OwnerReferences[0].Kind, "Gate")
	assert.Equal(vs.ObjectMeta.OwnerReferences[0].Name, "test-gate")
	assert.Equal(vs.ObjectMeta.OwnerReferences[0].UID, types.UID("eab0f1c8-c417-11e9-bf11-4ac644044351"))

}

func TestOauthPrepareVirtualService(t *testing.T) {
	assert := assert.New(t)

	gate := getGate()

	oldVS := generateVirtualService(gate, "test-oathkeeper", 4455, gate.Spec.Paths[0].Path)

	oldVS.ObjectMeta.Generation = int64(15)
	oldVS.ObjectMeta.Name = "mst"

	newVS := prepareVirtualService(gate, oldVS, "test-oathkeeper", 4455, gate.Spec.Paths[0].Path)

	assert.Equal(newVS.ObjectMeta.Generation, int64(15))

	assert.Equal(len(newVS.Spec.Gateways), 1)
	assert.Equal(newVS.Spec.Gateways[0], apiGateway)

	assert.Equal(len(newVS.Spec.Hosts), 1)
	assert.Equal(newVS.Spec.Hosts[0], serviceHost)

	assert.Equal(len(newVS.Spec.HTTP), 1)
	assert.Equal(len(newVS.Spec.HTTP[0].Route), 1)
	assert.Equal(len(newVS.Spec.HTTP[0].Match), 1)
	assert.Equal(newVS.Spec.HTTP[0].Route[0].Destination.Host, "test-oathkeeper")
	assert.Equal(int(newVS.Spec.HTTP[0].Route[0].Destination.Port.Number), 4455)
	assert.Equal(newVS.Spec.HTTP[0].Match[0].URI.Regex, "/foo")

	assert.Equal(newVS.ObjectMeta.Name, "test-gate-test-service")
	assert.Equal(newVS.ObjectMeta.Namespace, "test-namespace")

	assert.Equal(newVS.ObjectMeta.OwnerReferences[0].APIVersion, "gateway.kyma-project.io/v2alpha1")
	assert.Equal(newVS.ObjectMeta.OwnerReferences[0].Kind, "Gate")
	assert.Equal(newVS.ObjectMeta.OwnerReferences[0].Name, "test-gate")
	assert.Equal(newVS.ObjectMeta.OwnerReferences[0].UID, types.UID("eab0f1c8-c417-11e9-bf11-4ac644044351"))

}

func TestOauthGenerateAccessRule(t *testing.T) {
	assert := assert.New(t)

	gate := getGate()
	requiredScopes := []byte(`required_scopes: ["write", "read"]`)

	accessStrategy := &rulev1alpha1.Authenticator{
		Handler: &rulev1alpha1.Handler{
			Name: "oauth2_introspection",
			Config: &runtime.RawExtension{
				Raw: requiredScopes,
			},
		},
	}

	ar := generateAccessRule(gate, gate.Spec.Paths[0], []*rulev1alpha1.Authenticator{accessStrategy})

	assert.Equal(len(ar.Spec.Authenticators), 1)
	assert.NotEmpty(ar.Spec.Authenticators[0].Config)
	assert.Equal(ar.Spec.Authenticators[0].Name, "oauth2_introspection")
	assert.Equal(string(ar.Spec.Authenticators[0].Config.Raw), string(requiredScopes))

	assert.Equal(len(ar.Spec.Match.Methods), 1)
	assert.Equal(ar.Spec.Match.Methods[0], "GET")
	assert.Equal(ar.Spec.Match.URL, "<http|https>://myService.myDomain.com</foo>")

	assert.Equal(ar.Spec.Authorizer.Name, "allow")
	assert.Empty(ar.Spec.Authorizer.Config)

	assert.Equal(ar.Spec.Upstream.URL, "http://test-service.test-namespace.svc.cluster.local:8080")

	assert.Equal(ar.ObjectMeta.Name, "test-gate-test-service")
	assert.Equal(ar.ObjectMeta.Namespace, "test-namespace")

	assert.Equal(ar.ObjectMeta.OwnerReferences[0].APIVersion, "gateway.kyma-project.io/v2alpha1")
	assert.Equal(ar.ObjectMeta.OwnerReferences[0].Kind, "Gate")
	assert.Equal(ar.ObjectMeta.OwnerReferences[0].Name, "test-gate")
	assert.Equal(ar.ObjectMeta.OwnerReferences[0].UID, types.UID("eab0f1c8-c417-11e9-bf11-4ac644044351"))

}

func TestOauthPrepareAccessRule(t *testing.T) {
	assert := assert.New(t)

	gate := getGate()
	requiredScopes := []byte(`required_scopes: ["write", "read"]`)

	accessStrategy := &rulev1alpha1.Authenticator{
		Handler: &rulev1alpha1.Handler{
			Name: "oauth2_introspection",
			Config: &runtime.RawExtension{
				Raw: requiredScopes,
			},
		},
	}

	oldAR := generateAccessRule(gate, gate.Spec.Paths[0], []*rulev1alpha1.Authenticator{accessStrategy})

	oldAR.ObjectMeta.Generation = int64(15)
	oldAR.ObjectMeta.Name = "mst"

	newAR := prepareAccessRule(gate, oldAR, gate.Spec.Paths[0], []*rulev1alpha1.Authenticator{accessStrategy})

	assert.Equal(newAR.ObjectMeta.Generation, int64(15))

	assert.Equal(len(newAR.Spec.Authenticators), 1)
	assert.NotEmpty(newAR.Spec.Authenticators[0].Config)
	assert.Equal(newAR.Spec.Authenticators[0].Name, "oauth2_introspection")
	assert.Equal(string(newAR.Spec.Authenticators[0].Config.Raw), string(requiredScopes))

	assert.Equal(len(newAR.Spec.Match.Methods), 1)
	assert.Equal(newAR.Spec.Match.Methods[0], "GET")
	assert.Equal(newAR.Spec.Match.URL, "<http|https>://myService.myDomain.com</foo>")

	assert.Equal(newAR.Spec.Authorizer.Name, "allow")
	assert.Empty(newAR.Spec.Authorizer.Config)

	assert.Equal(newAR.Spec.Upstream.URL, "http://test-service.test-namespace.svc.cluster.local:8080")

	assert.Equal(newAR.ObjectMeta.Name, "test-gate-test-service")
	assert.Equal(newAR.ObjectMeta.Namespace, "test-namespace")

	assert.Equal(newAR.ObjectMeta.OwnerReferences[0].APIVersion, "gateway.kyma-project.io/v2alpha1")
	assert.Equal(newAR.ObjectMeta.OwnerReferences[0].Kind, "Gate")
	assert.Equal(newAR.ObjectMeta.OwnerReferences[0].Name, "test-gate")
	assert.Equal(newAR.ObjectMeta.OwnerReferences[0].UID, types.UID("eab0f1c8-c417-11e9-bf11-4ac644044351"))

}

func getGate() *gatewayv2alpha1.Gate {
	var apiUID types.UID = "eab0f1c8-c417-11e9-bf11-4ac644044351"
	var apiGateway = "some-gateway"
	var serviceName = "test-service"
	var serviceHost = "myService.myDomain.com"
	var servicePort uint32 = 8080

	return &gatewayv2alpha1.Gate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-gate",
			UID:       apiUID,
			Namespace: "test-namespace",
		},
		TypeMeta: metav1.TypeMeta{
			APIVersion: "gateway.kyma-project.io/v2alpha1",
			Kind:       "Gate",
		},
		Spec: gatewayv2alpha1.GateSpec{
			Gateway: &apiGateway,
			Service: &gatewayv2alpha1.Service{
				Name: &serviceName,
				Host: &serviceHost,
				Port: &servicePort,
			},
			Paths: []gatewayv2alpha1.Path{
				{
					Path:    "/foo",
					Scopes:  []string{"write", "read"},
					Methods: []string{"GET"},
				},
			},
			Mutators: []*rulev1alpha1.Mutator{},
		},
	}
}
