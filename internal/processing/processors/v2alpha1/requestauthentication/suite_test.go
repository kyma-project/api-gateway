package requestauthentication_test

import (
	apirulev1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"github.com/kyma-project/api-gateway/internal/processing"
	rulev1alpha1 "github.com/kyma-project/api-gateway/internal/types/ory/oathkeeper-maester/api/v1alpha1"
	"github.com/kyma-project/api-gateway/tests"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	ginkgotypes "github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"
)

func Test(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "requestauthentication v2alpha1 Suite")
}

var _ = ReportAfterSuite("custom reporter", func(report ginkgotypes.Report) {
	tests.GenerateGinkgoJunitReport("requestauthentication-v2alpha1-suite", report)
})

var apiRuleName = "test-apirule"
var apiRuleNamespace = "example-namespace"
var serviceName = "example-service"
var jwtIssuer = "https://oauth2.example.com/"
var jwksUri = "https://oauth2.example.com/.well-known/jwks.json"
var anotherJwtIssuer = "https://oauth2.another.example.com/"
var anotherJwksUri = "https://oauth2.another.example.com/.well-known/jwks.json"

var ActionToString = func(a processing.Action) string { return a.String() }

func getFakeClient(objs ...client.Object) client.Client {
	scheme := runtime.NewScheme()
	err := networkingv1beta1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())
	err = rulev1alpha1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())
	err = securityv1beta1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())
	err = apirulev1beta1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())
	err = corev1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())
	err = apiextensionsv1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	return fake.NewClientBuilder().WithScheme(scheme).WithObjects(objs...).Build()
}
