package gateway

import (
	"context"
	"fmt"
	"istio.io/client-go/pkg/apis/networking/v1beta1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("VirtualService", func() {

	Context("reconcileVirtualService", func() {

		It("should create Virtual Service with gateway and port", func() {
			// given
			k8sClient := createFakeClient()

			// when
			err := reconcileVirtualService(context.Background(), k8sClient, "test", "test-ns", "test-domain.com", "15021", "istio-ingressgateway.istio-system.svc.cluster.local")

			// then
			Expect(err).ShouldNot(HaveOccurred())

			createdVirtualService := v1beta1.VirtualService{}
			Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: "test", Namespace: "test-ns"}, &createdVirtualService)).Should(Succeed())
			Expect(createdVirtualService.Spec.Hosts).To(ContainElement("*.test-domain.com"))
			Expect(createdVirtualService.Spec.Gateways).To(ContainElement(fmt.Sprintf("%s/%s", kymaGatewayName, kymaGatewayNamespace)))
			Expect(createdVirtualService.Spec.Http).To(HaveLen(1))
			Expect(createdVirtualService.Spec.Http[0].Match).To(HaveLen(1))
			Expect(createdVirtualService.Spec.Http[0].Match[0].Uri).To(ContainSubstring("/healthz/ready"))
			Expect(createdVirtualService.Spec.Http[0].Route).To(HaveLen(1))
			Expect(createdVirtualService.Spec.Http[0].Route[0].Destination.Host).To(Equal("istio-ingressgateway.istio-system.svc.cluster.local"))
			Expect(createdVirtualService.Spec.Http[0].Route[0].Destination.Port.Number).To(Equal(uint32(15021)))
		})
	})
})
