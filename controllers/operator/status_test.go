package operator

import (
	"context"
	operatorv1alpha1 "github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("status", func() {
	It("Should update APIGateway CR status to ready", func() {
		cr := operatorv1alpha1.APIGateway{
			ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: testNamespace},
		}
		Expect(k8sClient.Create(context.TODO(), &cr)).To(Succeed())

		handler := newStatusHandler(k8sClient)

		Expect(handler.updateToReady(context.TODO(), &cr)).To(Succeed())

		Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Name: "test", Namespace: "default"}, &cr)).To(Succeed())
		Expect(cr.Status.State).To(Equal(operatorv1alpha1.Ready))
	})
})
