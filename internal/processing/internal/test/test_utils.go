package processing_test

import (
	v1beta12 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	"github.com/kyma-incubator/api-gateway/internal/processing"
	"github.com/onsi/gomega"
	rulev1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
	"istio.io/api/networking/v1beta1"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
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
	JwtIssuer                   = "https://oauth2.example.com/"
	OathkeeperSvc               = "fake.oathkeeper"
	OathkeeperSvcPort uint32    = 1234
	TestLabelKey                = "key"
	TestLabelValue              = "value"
	DefaultDomain               = "myDomain.com"
)

var (
	ApiMethods                     = []string{"GET"}
	ApiScopes                      = []string{"write", "read"}
	ServicePort             uint32 = 8080
	ApiGateway                     = "some-gateway"
	ServiceName                    = "example-service"
	ServiceHostWithNoDomain        = "myService"
	ServiceHost                    = ServiceHostWithNoDomain + "." + DefaultDomain

	TestAllowOrigin  = []*v1beta1.StringMatch{{MatchType: &v1beta1.StringMatch_Regex{Regex: ".*"}}}
	TestAllowMethods = []string{"GET", "POST", "PUT", "DELETE"}
	TestAllowHeaders = []string{"header1", "header2"}

	TestCors = &processing.CorsConfig{
		AllowOrigins: TestAllowOrigin,
		AllowMethods: TestAllowMethods,
		AllowHeaders: TestAllowHeaders,
	}

	TestAdditionalLabels = map[string]string{TestLabelKey: TestLabelValue}
)

func GetTestConfig() processing.ReconciliationConfig {
	return processing.ReconciliationConfig{
		OathkeeperSvc:     OathkeeperSvc,
		OathkeeperSvcPort: OathkeeperSvcPort,
		CorsConfig:        TestCors,
		AdditionalLabels:  TestAdditionalLabels,
		DefaultDomainName: DefaultDomain,
	}
}

func GetEmptyFakeClient() client.Client {
	scheme := runtime.NewScheme()
	err := networkingv1beta1.AddToScheme(scheme)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	err = rulev1alpha1.AddToScheme(scheme)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	return fake.NewClientBuilder().WithScheme(scheme).WithObjects().Build()
}

func GetRuleFor(path string, methods []string, mutators []*v1beta12.Mutator, accessStrategies []*v1beta12.Authenticator) v1beta12.Rule {
	return v1beta12.Rule{
		Path:             path,
		Methods:          methods,
		Mutators:         mutators,
		AccessStrategies: accessStrategies,
	}
}

func GetRuleWithServiceFor(path string, methods []string, mutators []*v1beta12.Mutator, accessStrategies []*v1beta12.Authenticator, service *v1beta12.Service) v1beta12.Rule {
	return v1beta12.Rule{
		Path:             path,
		Methods:          methods,
		Mutators:         mutators,
		AccessStrategies: accessStrategies,
		Service:          service,
	}
}

func GetAPIRuleFor(rules []v1beta12.Rule) *v1beta12.APIRule {
	return &v1beta12.APIRule{
		ObjectMeta: v1.ObjectMeta{
			Name:      ApiName,
			UID:       ApiUID,
			Namespace: ApiNamespace,
		},
		TypeMeta: v1.TypeMeta{
			APIVersion: ApiAPIVersion,
			Kind:       ApiKind,
		},
		Spec: v1beta12.APIRuleSpec{
			Gateway: &ApiGateway,
			Service: &v1beta12.Service{
				Name: &ServiceName,
				Port: &ServicePort,
			},
			Host:  &ServiceHost,
			Rules: rules,
		},
	}
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
