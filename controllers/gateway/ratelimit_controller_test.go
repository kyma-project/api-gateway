package gateway_test

import (
	"context"
	ratelimitv1alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v1alpha1"
	"github.com/kyma-project/api-gateway/controllers/gateway"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ = Describe("Rate Limit Controller", func() {
	It("Finish reconciliation if there is no RateLimit CR in the cluster", func() {

		fakeClient := fake.NewClientBuilder().WithScheme(getTestScheme()).WithObjects().Build()

		r := gateway.RateLimitReconciler{
			Scheme: getTestScheme(),
			Client: fakeClient,
		}

		result, err := r.Reconcile(context.Background(), reconcile.Request{NamespacedName: types.NamespacedName{Namespace: "test", Name: "test"}})

		Expect(err).ShouldNot(HaveOccurred())
		Expect(result).Should(Equal(reconcile.Result{}))
	})

})

func getTestScheme() *runtime.Scheme {
	s := runtime.NewScheme()
	utilruntime.Must(ratelimitv1alpha1.AddToScheme(s))
	Expect(corev1.AddToScheme(s)).Should(Succeed())
	Expect(apiextensionsv1.AddToScheme(s)).Should(Succeed())

	return s
}
