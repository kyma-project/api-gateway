package oathkeeper_test

import (
	"context"
	"errors"
	"testing"

	"github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	"github.com/kyma-project/api-gateway/tests"
	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"
	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

func TestOathkeeper(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Oathkeeper Suite")
}

var _ = ReportAfterSuite("custom reporter", func(report types.Report) {
	tests.GenerateGinkgoJunitReport("oathkeeper-suite", report)
})

func createFakeClient(objects ...client.Object) client.Client {
	return createFakeBuilderWithScheme().WithObjects(objects...).Build()
}

func createFakeClientThatFailsOnCreate() client.Client {
	interceptor := interceptor.Funcs{
		Create: func(_ context.Context, _ client.WithWatch, _ client.Object, _ ...client.CreateOption) error {
			return errors.New("faked create failed")
		},
	}
	return createFakeBuilderWithScheme().WithInterceptorFuncs(interceptor).Build()
}

func createFakeBuilderWithScheme() *fake.ClientBuilder {
	Expect(v1alpha1.AddToScheme(scheme.Scheme)).Should(Succeed())
	Expect(corev1.AddToScheme(scheme.Scheme)).Should(Succeed())
	Expect(v1alpha3.AddToScheme(scheme.Scheme)).Should(Succeed())
	Expect(v1beta1.AddToScheme(scheme.Scheme)).Should(Succeed())
	Expect(apiextensionsv1.AddToScheme(scheme.Scheme)).Should(Succeed())
	Expect(securityv1beta1.AddToScheme(scheme.Scheme)).Should(Succeed())

	return fake.NewClientBuilder().WithScheme(scheme.Scheme)
}
