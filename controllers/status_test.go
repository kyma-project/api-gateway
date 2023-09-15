package controllers

import (
	"context"
	operatorv1alpha1 "github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("status", func() {

	Context("UpdateToReady", func() {

		It("Should update APIGateway CR status to ready", func() {
			// given
			cr := operatorv1alpha1.APIGateway{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
			}

			k8sClient := createFakeClient(&cr)
			handler := NewStatusHandler(k8sClient)

			// when
			err := handler.UpdateToReady(context.TODO(), &cr)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Name: "test", Namespace: "default"}, &cr)).To(Succeed())
			Expect(cr.Status.State).To(Equal(operatorv1alpha1.Ready))
		})
	})

	Context("UpdateToError", func() {
		It("Should update CR status to error with description", func() {
			// given
			cr := operatorv1alpha1.APIGateway{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
			}

			k8sClient := createFakeClient(&cr)
			handler := NewStatusHandler(k8sClient)

			// when
			err := handler.UpdateToError(context.TODO(), &cr, "Something bad happened")

			// then
			Expect(err).ToNot(HaveOccurred())

			err = k8sClient.Get(context.TODO(), types.NamespacedName{Name: "test", Namespace: "default"}, &cr)
			Expect(err).ToNot(HaveOccurred())
			Expect(cr.Status.State).To(Equal(operatorv1alpha1.Error))
			Expect(cr.Status.Description).To(Equal("Something bad happened"))
		})

	})
})
