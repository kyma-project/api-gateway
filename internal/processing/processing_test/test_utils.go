package processing_test

import (
	"encoding/json"
	"fmt"
	apirulev1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"net/http"

	"github.com/kyma-project/api-gateway/internal/processing"
	rulev1alpha1 "github.com/kyma-project/api-gateway/internal/types/ory/oathkeeper-maester/api/v1alpha1"
	"github.com/onsi/gomega"
	"istio.io/api/networking/v1beta1"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	ApiName                     = "test-apirule"
	ApiUID            types.UID = "eab0f1c8-c417-11e9-bf11-4ac644044351"
	ApiNamespace                = "some-namespace"
	ApiAPIVersion               = "gateway.kyma-project.io/v1alpha1"
	ApiKind                     = "ApiRule"
	ApiPath                     = "/.*"
	HeadersApiPath              = "/headers"
	ImgApiPath                  = "/img"
	JwtIssuer                   = "https://oauth2.example.com/"
	JwksUri                     = "https://oauth2.example.com/.well-known/jwks.json"
	JwtIssuer2                  = "https://oauth2.another.example.com/"
	JwksUri2                    = "https://oauth2.another.example.com/.well-known/jwks.json"
	OathkeeperSvc               = "fake.oathkeeper"
	OathkeeperSvcPort uint32    = 1234
	TestLabelKey                = "key"
	TestLabelValue              = "value"
	DefaultDomain               = "myDomain.com"
	TestSelectorKey             = "app"
)

var (
	ApiMethods                     = []apirulev1beta1.HttpMethod{http.MethodGet}
	ApiScopes                      = []string{"write", "read"}
	ServicePort             uint32 = 8080
	ApiGateway                     = "some-gateway"
	ServiceName                    = "example-service"
	ServiceHostWithNoDomain        = "myService"
	ServiceHost                    = ServiceHostWithNoDomain + "." + DefaultDomain

	TestAllowOrigin  = []*v1beta1.StringMatch{{MatchType: &v1beta1.StringMatch_Regex{Regex: ".*"}}}
	TestAllowMethods = []apirulev1beta1.HttpMethod{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete}
	TestAllowHeaders = []string{"header1", "header2"}
	TestCors         = &processing.CorsConfig{
		AllowOrigins: []*v1beta1.StringMatch{{MatchType: &v1beta1.StringMatch_Regex{Regex: ".*"}}},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
		AllowHeaders: []string{"header1", "header2"},
	}
)

func GetTestConfig() processing.ReconciliationConfig {
	return processing.ReconciliationConfig{
		OathkeeperSvc:     OathkeeperSvc,
		OathkeeperSvcPort: OathkeeperSvcPort,
		CorsConfig:        TestCors,
		DefaultDomainName: DefaultDomain,
	}
}

func GetFakeClient(objs ...client.Object) client.Client {
	scheme := runtime.NewScheme()
	err := networkingv1beta1.AddToScheme(scheme)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	err = rulev1alpha1.AddToScheme(scheme)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	err = securityv1beta1.AddToScheme(scheme)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	err = apirulev1beta1.AddToScheme(scheme)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	err = corev1.AddToScheme(scheme)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	err = apiextensionsv1.AddToScheme(scheme)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	crds := []string{
		"rules.oathkeeper.ory.sh",
		"authorizationpolicies.security.istio.io",
		"requestauthentications.security.istio.io",
		"virtualservices.networking.istio.io",
	}
	objs = append(objs, getCrds(crds...)...)
	k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(objs...).Build()
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	return k8sClient
}

func getCrds(names ...string) []client.Object {
	var crds []client.Object
	for _, name := range names {
		crd := &apiextensionsv1.CustomResourceDefinition{}
		crd.Name = name
		crds = append(crds, crd)
	}
	return crds
}

func GetRuleFor(path string, methods []apirulev1beta1.HttpMethod, mutators []*apirulev1beta1.Mutator, accessStrategies []*apirulev1beta1.Authenticator) apirulev1beta1.Rule {
	return apirulev1beta1.Rule{
		Path:             path,
		Methods:          methods,
		Mutators:         mutators,
		AccessStrategies: accessStrategies,
	}
}

func GetRuleWithServiceFor(path string, methods []apirulev1beta1.HttpMethod, mutators []*apirulev1beta1.Mutator, accessStrategies []*apirulev1beta1.Authenticator, service *apirulev1beta1.Service) apirulev1beta1.Rule {
	return apirulev1beta1.Rule{
		Path:             path,
		Methods:          methods,
		Mutators:         mutators,
		AccessStrategies: accessStrategies,
		Service:          service,
	}
}

func GetJwtRuleWithService(jwtIssuer, jwksUri, serviceName string, namespace ...string) apirulev1beta1.Rule {
	jwtConfigJSON := fmt.Sprintf(`{"authentications": [{"issuer": "%s", "jwksUri": "%s"}]}`, jwtIssuer, jwksUri)
	jwt := []*apirulev1beta1.Authenticator{
		{
			Handler: &apirulev1beta1.Handler{
				Name: "jwt",
				Config: &runtime.RawExtension{
					Raw: []byte(jwtConfigJSON),
				},
			},
		},
	}

	port := uint32(8080)
	jwtRuleService := &apirulev1beta1.Service{
		Name: &serviceName,
		Port: &port,
	}
	if len(namespace) > 0 {
		jwtRuleService.Namespace = &namespace[0]
	}

	return GetRuleWithServiceFor("path", ApiMethods, []*apirulev1beta1.Mutator{}, jwt, jwtRuleService)
}

func GetAPIRuleFor(rules []apirulev1beta1.Rule, namespace ...string) *apirulev1beta1.APIRule {
	apiRule := apirulev1beta1.APIRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ApiName,
			UID:       ApiUID,
			Namespace: ApiNamespace,
		},
		TypeMeta: metav1.TypeMeta{
			APIVersion: ApiAPIVersion,
			Kind:       ApiKind,
		},
		Spec: apirulev1beta1.APIRuleSpec{
			Gateway: &ApiGateway,
			Service: &apirulev1beta1.Service{
				Name: &ServiceName,
				Port: &ServicePort,
			},
			Host:  &ServiceHost,
			Rules: rules,
		},
	}
	if len(namespace) > 0 {
		apiRule.Namespace = namespace[0]
	}
	return &apiRule
}

func GetService(name string, namespace ...string) *corev1.Service {
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ApiNamespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": name,
			},
		},
	}
	if len(namespace) > 0 {
		svc.Namespace = namespace[0]
	}
	return svc
}

func ToCSVList(input []string) string {
	if len(input) == 0 {
		return ""
	}

	res := `"` + input[0] + `"`

	for i := 1; i < len(input); i++ {
		res = res + "," + `"` + input[i] + `"`
	}

	return res
}

var ActionToString = func(a processing.Action) string { return a.String() }

func GetRawConfig(config any) *runtime.RawExtension {
	bytes, err := json.Marshal(config)
	gomega.Expect(err).To(gomega.BeNil())
	return &runtime.RawExtension{
		Raw: bytes,
	}
}
