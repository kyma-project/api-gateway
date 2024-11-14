package ratelimit_test

import (
	"context"
	"github.com/kyma-project/api-gateway/controllers/ratelimit"
	"k8s.io/apimachinery/pkg/runtime"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Rate Limit Controller", func() {
	It("Dummy test", func() {

		r := ratelimit.Reconciler{
			Scheme: runtime.NewScheme(),
		}

		result, err := r.Reconcile(context.Background(), reconcile.Request{NamespacedName: types.NamespacedName{Namespace: "test", Name: "test"}})

		Expect(err).ShouldNot(HaveOccurred())
		Expect(result).Should(Equal(reconcile.Result{
			Requeue: false,
		}))
	})

})
