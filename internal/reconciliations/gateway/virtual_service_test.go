package gateway

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"istio.io/client-go/pkg/apis/networking/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("VirtualService", func() {

	Context("reconcileVirtualService", func() {

		It("should create Virtual Service with gateway and port", func() {
			// given
			k8sClient := createFakeClient()

			// when
			err := reconcileVirtualService(context.Background(), k8sClient, "test", "test-ns", "test-domain.com")

			// then
			Expect(err).ShouldNot(HaveOccurred())

			createdVirtualService := v1beta1.VirtualService{}
			Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: "test", Namespace: "test-ns"}, &createdVirtualService)).Should(Succeed())
			Expect(createdVirtualService.Spec.Hosts).To(ContainElement("healthz.test-domain.com"))
			Expect(createdVirtualService.Spec.Gateways).To(ContainElement("kyma-system/kyma-gateway"))
			Expect(createdVirtualService.Spec.Http).To(HaveLen(1))
			Expect(createdVirtualService.Spec.Http[0].Match).To(HaveLen(1))
			Expect(createdVirtualService.Spec.Http[0].Match[0].Uri).To(ContainSubstring("/healthz/ready"))
			Expect(createdVirtualService.Spec.Http[0].Route).To(HaveLen(1))
			Expect(createdVirtualService.Spec.Http[0].Route[0].Destination.Host).To(Equal("istio-ingressgateway.istio-system.svc.cluster.local"))
			Expect(createdVirtualService.Spec.Http[0].Route[0].Destination.Port.Number).To(Equal(uint32(15021)))
		})
	})
})
