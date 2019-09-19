package processing

import (
	"fmt"
	"testing"

	gatewayv1alpha1 "github.com/kyma-incubator/api-gateway/api/v1alpha1"
	rulev1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

var (
	apiName                 = "test-apirule"
	apiUID        types.UID = "eab0f1c8-c417-11e9-bf11-4ac644044351"
	apiNamespace            = "some-namespace"
	apiAPIVersion           = "gateway.kyma-project.io/v1alpha1"
	apiKind                 = "ApiRule"
	apiGateway              = "some-gateway"
	apiPath                 = "/.*"
	apiMethods              = []string{"GET"}
	serviceName             = "example-service"
	serviceHost             = "myService.myDomain.com"
	servicePort   uint32    = 8080
	authStrategy            = "ALLOW"
	apiScopes               = []string{"write", "read"}
	jwtIssuer               = "https://oauth2.example.com/"
)

func getAPIRuleFor(strategies []*rulev1alpha1.Authenticator, mutators []*rulev1alpha1.Mutator) *gatewayv1alpha1.APIRule {
	return &gatewayv1alpha1.APIRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      apiName,
			UID:       apiUID,
			Namespace: apiNamespace,
		},
		TypeMeta: metav1.TypeMeta{
			APIVersion: apiAPIVersion,
			Kind:       apiKind,
		},
		Spec: gatewayv1alpha1.APIRuleSpec{
			Gateway: &apiGateway,
			Service: &gatewayv1alpha1.Service{
				Name: &serviceName,
				Host: &serviceHost,
				Port: &servicePort,
			},
			Rules: []gatewayv1alpha1.Rule{
				{
					Path:             apiPath,
					Methods:          apiMethods,
					Mutators:         mutators,
					AccessStrategies: strategies,
				},
			},
		},
	}
}

func toCSVList(input []string) string {
	if len(input) == 0 {
		return ""
	}

	res := `"` + input[0] + `"`

	for i := 1; i < len(input); i++ {
		res = res + "," + `"` + input[i] + `"`
	}

	return res
}

func TestCreateVS_NoOp(t *testing.T) {
	assert := assert.New(t)

	strategies := []*rulev1alpha1.Authenticator{
		{
			Handler: &rulev1alpha1.Handler{
				Name: "allow",
			},
		},
	}

	apiRule := getAPIRuleFor(strategies, []*rulev1alpha1.Mutator{})
	f := &Factory{oathkeeperSvcPort: 1234, oathkeeperSvc: "fake.oathkeeper"}

	vs := f.generateVirtualService(apiRule)

	assert.Equal(len(vs.Spec.Gateways), 1)
	assert.Equal(vs.Spec.Gateways[0], apiGateway)

	assert.Equal(len(vs.Spec.Hosts), 1)
	assert.Equal(vs.Spec.Hosts[0], serviceHost)

	assert.Equal(len(vs.Spec.HTTP), 1)
	assert.Equal(len(vs.Spec.HTTP[0].Route), 1)
	assert.Equal(len(vs.Spec.HTTP[0].Match), 1)
	assert.Equal(vs.Spec.HTTP[0].Route[0].Destination.Host, serviceName+"."+apiNamespace+".svc.cluster.local")
	assert.Equal(vs.Spec.HTTP[0].Route[0].Destination.Port.Number, servicePort)
	assert.Equal(vs.Spec.HTTP[0].Match[0].URI.Regex, apiRule.Spec.Rules[0].Path)

	assert.Equal(vs.ObjectMeta.Name, apiName+"-"+serviceName)
	assert.Equal(vs.ObjectMeta.Namespace, apiNamespace)

	assert.Equal(vs.ObjectMeta.OwnerReferences[0].APIVersion, apiAPIVersion)
	assert.Equal(vs.ObjectMeta.OwnerReferences[0].Kind, apiKind)
	assert.Equal(vs.ObjectMeta.OwnerReferences[0].Name, apiName)
	assert.Equal(vs.ObjectMeta.OwnerReferences[0].UID, apiUID)

}

func TestCreateVS_JWT(t *testing.T) {
	assert := assert.New(t)

	configJSON := fmt.Sprintf(`
		{
			"trusted_issuers": ["%s"],
			"jwks": [],
			"required_scope": [%s]
	}`, jwtIssuer, toCSVList(apiScopes))

	strategies := []*rulev1alpha1.Authenticator{
		{
			Handler: &rulev1alpha1.Handler{
				Name: "jwt",
				Config: &runtime.RawExtension{
					Raw: []byte(configJSON),
				},
			},
		},
	}

	apiRule := getAPIRuleFor(strategies, []*rulev1alpha1.Mutator{})
	f := &Factory{oathkeeperSvcPort: 4455, oathkeeperSvc: "test-oathkeeper"}

	vs := f.generateVirtualService(apiRule)

	assert.Equal(len(vs.Spec.Gateways), 1)
	assert.Equal(vs.Spec.Gateways[0], apiGateway)

	assert.Equal(len(vs.Spec.Hosts), 1)
	assert.Equal(vs.Spec.Hosts[0], serviceHost)

	assert.Equal(len(vs.Spec.HTTP), 1)
	assert.Equal(len(vs.Spec.HTTP[0].Route), 1)
	assert.Equal(len(vs.Spec.HTTP[0].Match), 1)
	assert.Equal(vs.Spec.HTTP[0].Route[0].Destination.Host, "test-oathkeeper")
	assert.Equal(vs.Spec.HTTP[0].Route[0].Destination.Port.Number, uint32(4455))
	assert.Equal(vs.Spec.HTTP[0].Match[0].URI.Regex, apiRule.Spec.Rules[0].Path)

	assert.Equal(vs.ObjectMeta.Name, apiName+"-"+serviceName)
	assert.Equal(vs.ObjectMeta.Namespace, apiNamespace)

	assert.Equal(vs.ObjectMeta.OwnerReferences[0].APIVersion, apiAPIVersion)
	assert.Equal(vs.ObjectMeta.OwnerReferences[0].Kind, apiKind)
	assert.Equal(vs.ObjectMeta.OwnerReferences[0].Name, apiName)
	assert.Equal(vs.ObjectMeta.OwnerReferences[0].UID, apiUID)

}

func TestPrepareAR_JWT(t *testing.T) {
	assert := assert.New(t)

	configJSON := fmt.Sprintf(`
		{
			"trusted_issuers": ["%s"],
			"jwks": [],
			"required_scope": [%s]
	}`, jwtIssuer, toCSVList(apiScopes))

	strategies := []*rulev1alpha1.Authenticator{
		{
			Handler: &rulev1alpha1.Handler{
				Name: "jwt",
				Config: &runtime.RawExtension{
					Raw: []byte(configJSON),
				},
			},
		},
	}

	apiRule := getAPIRuleFor(strategies, []*rulev1alpha1.Mutator{})

	oldAR := generateAccessRule(apiRule, apiRule.Spec.Rules[0], 0, []*rulev1alpha1.Authenticator{strategies[0]})
	oldAROnwerRef := oldAR.OwnerReferences[0]
	newAR := prepareAccessRule(apiRule, oldAR, apiRule.Spec.Rules[0], 0, []*rulev1alpha1.Authenticator{strategies[0]})

	assert.Equal(len(newAR.Spec.Authenticators), 1)
	assert.Equal(newAR.Spec.Authenticators[0].Name, "jwt")
	assert.NotEmpty(newAR.Spec.Authenticators[0].Config)
	assert.Equal(string(newAR.Spec.Authenticators[0].Config.Raw), configJSON)

	assert.Equal(len(newAR.Spec.Match.Methods), 1)
	assert.Equal(newAR.Spec.Match.Methods, []string{"GET"})
	assert.Equal(newAR.Spec.Match.URL, "<http|https>://myService.myDomain.com</.*>")

	assert.Equal(newAR.Spec.Authorizer.Name, "allow")
	assert.Empty(newAR.Spec.Authorizer.Config)

	assert.Equal(newAR.Spec.Upstream.URL, "http://example-service.some-namespace.svc.cluster.local:8080")

	assert.Equal(newAR.ObjectMeta.Name, apiName+"-"+serviceName+"-0")
	assert.Equal(newAR.ObjectMeta.Namespace, apiNamespace)

	assert.Equal(len(newAR.ObjectMeta.OwnerReferences), 1)

	assert.Equal(newAR.ObjectMeta.OwnerReferences[0].APIVersion, oldAROnwerRef.APIVersion)
	assert.Equal(newAR.ObjectMeta.OwnerReferences[0].Kind, oldAROnwerRef.Kind)
	assert.Equal(newAR.ObjectMeta.OwnerReferences[0].Name, oldAROnwerRef.Name)
	assert.Equal(newAR.ObjectMeta.OwnerReferences[0].UID, oldAROnwerRef.UID)
}

func TestGenerateAR_JWT(t *testing.T) {
	assert := assert.New(t)

	configJSON := fmt.Sprintf(`
		{
			"trusted_issuers": ["%s"],
			"jwks": [],
			"required_scope": [%s]
	}`, jwtIssuer, toCSVList(apiScopes))

	strategies := []*rulev1alpha1.Authenticator{
		{
			Handler: &rulev1alpha1.Handler{
				Name: "jwt",
				Config: &runtime.RawExtension{
					Raw: []byte(configJSON),
				},
			},
		},
	}

	apiRule := getAPIRuleFor(strategies, []*rulev1alpha1.Mutator{})

	ar := generateAccessRule(apiRule, apiRule.Spec.Rules[0], 0, []*rulev1alpha1.Authenticator{strategies[0]})

	assert.Equal(len(ar.Spec.Authenticators), 1)
	assert.NotEmpty(ar.Spec.Authenticators[0].Config)
	assert.Equal(ar.Spec.Authenticators[0].Name, "jwt")
	assert.Equal(string(ar.Spec.Authenticators[0].Config.Raw), configJSON)

	assert.Equal(len(ar.Spec.Match.Methods), 1)
	assert.Equal(ar.Spec.Match.Methods, []string{"GET"})
	assert.Equal(ar.Spec.Match.URL, "<http|https>://myService.myDomain.com</.*>")

	assert.Equal(ar.Spec.Authorizer.Name, "allow")
	assert.Empty(ar.Spec.Authorizer.Config)

	assert.Equal(ar.Spec.Upstream.URL, "http://example-service.some-namespace.svc.cluster.local:8080")

	assert.Equal(ar.ObjectMeta.Name, apiName+"-"+serviceName+"-0")
	assert.Equal(ar.ObjectMeta.Namespace, apiNamespace)

	assert.Equal(ar.ObjectMeta.OwnerReferences[0].APIVersion, apiAPIVersion)
	assert.Equal(ar.ObjectMeta.OwnerReferences[0].Kind, apiKind)
	assert.Equal(ar.ObjectMeta.OwnerReferences[0].Name, apiName)
	assert.Equal(ar.ObjectMeta.OwnerReferences[0].UID, apiUID)

}

func TestGenerateVS_OAUTH(t *testing.T) {
	assert := assert.New(t)

	configJSON := fmt.Sprintf(`
		{
			"required_scope": [%s]
	}`, toCSVList(apiScopes))

	strategies := []*rulev1alpha1.Authenticator{
		{
			Handler: &rulev1alpha1.Handler{
				Name: "oauth2_introspection",
				Config: &runtime.RawExtension{
					Raw: []byte(configJSON),
				},
			},
		},
	}

	apiRule := getAPIRuleFor(strategies, []*rulev1alpha1.Mutator{})
	f := &Factory{oathkeeperSvcPort: 4455, oathkeeperSvc: "test-oathkeeper"}
	vs := f.generateVirtualService(apiRule)

	assert.Equal(len(vs.Spec.Gateways), 1)
	assert.Equal(vs.Spec.Gateways[0], apiGateway)

	assert.Equal(len(vs.Spec.Hosts), 1)
	assert.Equal(vs.Spec.Hosts[0], serviceHost)

	assert.Equal(len(vs.Spec.HTTP), 1)
	assert.Equal(len(vs.Spec.HTTP[0].Route), 1)
	assert.Equal(len(vs.Spec.HTTP[0].Match), 1)
	assert.Equal(vs.Spec.HTTP[0].Route[0].Destination.Host, "test-oathkeeper")
	assert.Equal(vs.Spec.HTTP[0].Route[0].Destination.Port.Number, uint32(4455))
	assert.Equal(vs.Spec.HTTP[0].Match[0].URI.Regex, apiPath)

	assert.Equal(vs.ObjectMeta.Name, apiName+"-"+serviceName)
	assert.Equal(vs.ObjectMeta.Namespace, apiNamespace)

	assert.Equal(vs.ObjectMeta.OwnerReferences[0].APIVersion, apiAPIVersion)
	assert.Equal(vs.ObjectMeta.OwnerReferences[0].Kind, apiKind)
	assert.Equal(vs.ObjectMeta.OwnerReferences[0].Name, apiName)
	assert.Equal(vs.ObjectMeta.OwnerReferences[0].UID, apiUID)

}

func TestPrepareVS_OAUTH(t *testing.T) {
	assert := assert.New(t)

	configJSON := fmt.Sprintf(`
		{
			"required_scope": [%s]
	}`, toCSVList(apiScopes))

	strategies := []*rulev1alpha1.Authenticator{
		{
			Handler: &rulev1alpha1.Handler{
				Name: "oauth2_introspection",
				Config: &runtime.RawExtension{
					Raw: []byte(configJSON),
				},
			},
		},
	}

	apiRule := getAPIRuleFor(strategies, []*rulev1alpha1.Mutator{})
	f := &Factory{oathkeeperSvcPort: 4455, oathkeeperSvc: "test-oathkeeper"}

	oldVS := f.generateVirtualService(apiRule)
	oldVSOnwerRef := oldVS.OwnerReferences[0]
	newVS := f.prepareVirtualService(apiRule, oldVS)

	assert.Equal(len(newVS.Spec.Gateways), 1)
	assert.Equal(newVS.Spec.Gateways[0], apiGateway)

	assert.Equal(len(newVS.Spec.Hosts), 1)
	assert.Equal(newVS.Spec.Hosts[0], serviceHost)

	assert.Equal(len(newVS.Spec.HTTP), 1)
	assert.Equal(len(newVS.Spec.HTTP[0].Route), 1)
	assert.Equal(len(newVS.Spec.HTTP[0].Match), 1)
	assert.Equal(newVS.Spec.HTTP[0].Route[0].Destination.Host, "test-oathkeeper")
	assert.Equal(int(newVS.Spec.HTTP[0].Route[0].Destination.Port.Number), 4455)
	assert.Equal(newVS.Spec.HTTP[0].Match[0].URI.Regex, apiPath)

	assert.Equal(len(newVS.ObjectMeta.OwnerReferences), 1)
	assert.Equal(newVS.ObjectMeta.OwnerReferences[0].APIVersion, oldVSOnwerRef.APIVersion)
	assert.Equal(newVS.ObjectMeta.OwnerReferences[0].Kind, oldVSOnwerRef.Kind)
	assert.Equal(newVS.ObjectMeta.OwnerReferences[0].Name, oldVSOnwerRef.Name)
	assert.Equal(newVS.ObjectMeta.OwnerReferences[0].UID, oldVSOnwerRef.UID)
}

func TestGenerateAR_OAUTH(t *testing.T) {
	assert := assert.New(t)

	configJSON := fmt.Sprintf(`
		{
			"required_scope": [%s]
	}`, toCSVList(apiScopes))

	strategies := []*rulev1alpha1.Authenticator{
		{
			Handler: &rulev1alpha1.Handler{
				Name: "oauth2_introspection",
				Config: &runtime.RawExtension{
					Raw: []byte(configJSON),
				},
			},
		},
	}

	apiRule := getAPIRuleFor(strategies, []*rulev1alpha1.Mutator{})

	ar := generateAccessRule(apiRule, apiRule.Spec.Rules[0], 0, []*rulev1alpha1.Authenticator{strategies[0]})

	assert.Equal(len(ar.Spec.Authenticators), 1)
	assert.NotEmpty(ar.Spec.Authenticators[0].Config)
	assert.Equal(ar.Spec.Authenticators[0].Name, "oauth2_introspection")
	assert.Equal(string(ar.Spec.Authenticators[0].Config.Raw), configJSON)

	assert.Equal(len(ar.Spec.Match.Methods), 1)
	assert.Equal(ar.Spec.Match.Methods[0], "GET")
	assert.Equal(ar.Spec.Match.URL, "<http|https>://myService.myDomain.com</.*>")

	assert.Equal(ar.Spec.Authorizer.Name, "allow")
	assert.Empty(ar.Spec.Authorizer.Config)

	assert.Equal(ar.Spec.Upstream.URL, "http://example-service.some-namespace.svc.cluster.local:8080")

	assert.Equal(ar.ObjectMeta.Name, apiName+"-"+serviceName+"-0")
	assert.Equal(ar.ObjectMeta.Namespace, apiNamespace)

	assert.Equal(ar.ObjectMeta.OwnerReferences[0].APIVersion, apiAPIVersion)
	assert.Equal(ar.ObjectMeta.OwnerReferences[0].Kind, apiKind)
	assert.Equal(ar.ObjectMeta.OwnerReferences[0].Name, apiName)
	assert.Equal(ar.ObjectMeta.OwnerReferences[0].UID, apiUID)

}

func TestPreapreAR_OAUTH(t *testing.T) {
	assert := assert.New(t)

	configJSON := fmt.Sprintf(`
		{
			"required_scope": [%s]
	}`, toCSVList(apiScopes))

	strategies := []*rulev1alpha1.Authenticator{
		{
			Handler: &rulev1alpha1.Handler{
				Name: "oauth2_introspection",
				Config: &runtime.RawExtension{
					Raw: []byte(configJSON),
				},
			},
		},
	}

	apiRule := getAPIRuleFor(strategies, []*rulev1alpha1.Mutator{})

	oldAR := generateAccessRule(apiRule, apiRule.Spec.Rules[0], 0, strategies)
	oldAROnwerRef := oldAR.OwnerReferences[0]
	newAR := prepareAccessRule(apiRule, oldAR, apiRule.Spec.Rules[0], 0, strategies)

	assert.Equal(len(newAR.Spec.Authenticators), 1)
	assert.NotEmpty(newAR.Spec.Authenticators[0].Config)
	assert.Equal(newAR.Spec.Authenticators[0].Name, "oauth2_introspection")
	assert.Equal(string(newAR.Spec.Authenticators[0].Config.Raw), configJSON)

	assert.Equal(len(newAR.Spec.Match.Methods), 1)
	assert.Equal(newAR.Spec.Match.Methods[0], "GET")
	assert.Equal(newAR.Spec.Match.URL, "<http|https>://myService.myDomain.com</.*>")

	assert.Equal(newAR.Spec.Authorizer.Name, "allow")
	assert.Empty(newAR.Spec.Authorizer.Config)

	assert.Equal(newAR.Spec.Upstream.URL, "http://example-service.some-namespace.svc.cluster.local:8080")

	assert.Equal(len(newAR.ObjectMeta.OwnerReferences), 1)
	assert.Equal(newAR.ObjectMeta.OwnerReferences[0].APIVersion, oldAROnwerRef.APIVersion)
	assert.Equal(newAR.ObjectMeta.OwnerReferences[0].Kind, oldAROnwerRef.Kind)
	assert.Equal(newAR.ObjectMeta.OwnerReferences[0].Name, oldAROnwerRef.Name)
	assert.Equal(newAR.ObjectMeta.OwnerReferences[0].UID, oldAROnwerRef.UID)
}
