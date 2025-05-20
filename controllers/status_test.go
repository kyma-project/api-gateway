package controllers

import (
	"context"
	"fmt"
	"reflect"

	operatorv1alpha1 "github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("status", func() {

	Context("UpdateApiGatewayStatus", func() {

		It("Should Update APIGateway CR state and set description", func() {
			// given
			cr := operatorv1alpha1.APIGateway{
				ObjectMeta: metav1.ObjectMeta{Name: "test"},
			}

			newStatus := ErrorStatus(fmt.Errorf("test error"), "test description", nil)
			k8sClient := createFakeClient(&cr)
			// when
			err := UpdateApiGatewayStatus(context.Background(), k8sClient, &cr, newStatus)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(k8sClient.Get(context.Background(), types.NamespacedName{Name: "test"}, &cr)).To(Succeed())
			Expect(cr.Status.State).To(Equal(operatorv1alpha1.Error))
			Expect(cr.Status.Description).To(Equal("test description"))
		})

		It("Should return error if update status fails", func() {
			// given
			cr := operatorv1alpha1.APIGateway{
				ObjectMeta: metav1.ObjectMeta{Name: "test"},
			}

			newStatus := ReadyStatus(nil)
			k8sClient := fake.NewClientBuilder().Build()
			// when
			err := UpdateApiGatewayStatus(context.Background(), k8sClient, &cr, newStatus)

			// then
			Expect(err).To(HaveOccurred())
		})

		It("Should contain condition that is not nil and with expected value", func() {
			// given
			status := ErrorStatus(fmt.Errorf("asd"), "", &metav1.Condition{Type: "Ready", Status: "False"})
			expected := &metav1.Condition{Type: "Ready", Status: "False"}

			cr := operatorv1alpha1.APIGateway{
				ObjectMeta: metav1.ObjectMeta{Name: "test"},
			}
			k8sClient := createFakeClient(&cr)

			// when
			err := UpdateApiGatewayStatus(context.Background(), k8sClient, &cr, status)
			result := status.Condition()
			ok := reflect.DeepEqual(result, expected)

			// then
			Expect(err).To(BeNil())
			Expect(result).ToNot(BeNil())
			Expect(ok).To(BeTrue())
		})
	})

	Context("ToAPIGatewayStatus", func() {

		It("Should return Error with description set", func() {
			// given
			status := ErrorStatus(fmt.Errorf("test error"), "test description", nil)

			// when
			apiGatewayStatus, err := status.ToAPIGatewayStatus()

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(apiGatewayStatus.State).To(Equal(operatorv1alpha1.Error))
			Expect(apiGatewayStatus.Description).To(Equal("test description"))
		})

		It("Should return Warning with description set", func() {
			// given
			status := WarningStatus(fmt.Errorf("test error"), "test description", nil)

			// when
			apiGatewayStatus, err := status.ToAPIGatewayStatus()

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(apiGatewayStatus.State).To(Equal(operatorv1alpha1.Warning))
			Expect(apiGatewayStatus.Description).To(Equal("test description"))
		})

		It("Should return Ready with default description", func() {
			// given
			status := ReadyStatus(nil)

			// when
			apiGatewayStatus, err := status.ToAPIGatewayStatus()

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(apiGatewayStatus.State).To(Equal(operatorv1alpha1.Ready))
			Expect(apiGatewayStatus.Description).To(Equal("Successfully reconciled"))
		})

	})

	Context("IsError", func() {
		It("Should return true if status is Error", func() {
			// given
			status := ErrorStatus(fmt.Errorf("test error"), "test description", nil)

			// when
			isError := status.IsError()

			// then
			Expect(isError).To(BeTrue())
		})
		It("Should return false if status is not Error", func() {
			// given
			status := WarningStatus(fmt.Errorf("test error"), "test description", nil)

			// when
			isError := status.IsError()

			// then
			Expect(isError).To(BeFalse())
		})
	})

	Context("IsWarning", func() {
		It("Should return true if status is Warning", func() {
			// given
			status := WarningStatus(fmt.Errorf("test error"), "test description", nil)

			// when
			isWarning := status.IsWarning()

			// then
			Expect(isWarning).To(BeTrue())
		})
		It("Should return false if status is not Warning", func() {
			// given
			status := ErrorStatus(fmt.Errorf("test error"), "test description", nil)

			// when
			isWarning := status.IsWarning()

			// then
			Expect(isWarning).To(BeFalse())
		})
	})

	Context("IsReady", func() {
		It("Should return true if status is Ready", func() {
			// given
			status := ReadyStatus(nil)

			// when
			result := status.IsReady()

			// then
			Expect(result).To(BeTrue())
		})
		It("Should return false if status is not Ready", func() {
			// given
			status := ErrorStatus(fmt.Errorf("test error"), "test description", nil)

			// when
			result := status.IsReady()

			// then
			Expect(result).To(BeFalse())
		})
	})
})
