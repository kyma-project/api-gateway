package gateway

import (
	"context"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Gateway", func() {

	Context("reconcileGateway", func() {

		It("should create gateway with domain", func() {
			// given
			k8sClient := createFakeClient()

			// when
			err := reconcileGateway(context.TODO(), k8sClient, "test", "test-ns", "test-domain.com")

			// then
			Expect(err).ShouldNot(HaveOccurred())

			createdGateway := v1alpha3.Gateway{}
			Expect(k8sClient.Get(context.TODO(), client.ObjectKey{Name: "test", Namespace: "test-ns"}, &createdGateway)).Should(Succeed())

			for _, server := range createdGateway.Spec.GetServers() {
				Expect(server.Hosts).To(ContainElement("*.test-domain.com"))
			}
		})

		It("should apply disclaimer annotation on gateway when it was removed", func() {
			// given
			k8sClient := createFakeClient()
			Expect(reconcileGateway(context.TODO(), k8sClient, "test", "test-ns", "test-domain.com")).Should(Succeed())

			By("removing disclaimer annotation from gateway")
			gateway := v1alpha3.Gateway{}
			Expect(k8sClient.Get(context.TODO(), client.ObjectKey{Name: "test", Namespace: "test-ns"}, &gateway)).Should(Succeed())
			gateway.Annotations = nil
			Expect(k8sClient.Update(context.TODO(), &gateway)).Should(Succeed())

			// when
			Expect(reconcileGateway(context.TODO(), k8sClient, "test", "test-ns", "test-domain.com")).Should(Succeed())

			// then
			Expect(k8sClient.Get(context.TODO(), client.ObjectKey{Name: "test", Namespace: "test-ns"}, &gateway)).Should(Succeed())

			Expect(gateway.Annotations).To(HaveKeyWithValue("apigateways.operator.kyma-project.io/managed-by-disclaimer",
				"DO NOT EDIT - This resource is managed by Kyma.\nAny modifications are discarded and the resource is reverted to the original state."))
		})

	})
})
