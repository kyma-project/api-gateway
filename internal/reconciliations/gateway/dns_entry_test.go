package gateway

import (
	"context"

	dnsv1alpha1 "github.com/gardener/external-dns-management/pkg/apis/dns/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("DNSEntry", func() {
	Context("reconcileDNSEntry", func() {
		It("should create dnsEntry with domain and IP", func() {
			// given
			k8sClient := createFakeClient()

			// when
			err := reconcileDNSEntry(context.Background(), k8sClient, "test", "test-ns", "test-domain.com", "10.0.0.1")

			// then
			Expect(err).ShouldNot(HaveOccurred())

			createdDnsEntry := dnsv1alpha1.DNSEntry{}
			Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: "test", Namespace: "test-ns"}, &createdDnsEntry)).Should(Succeed())
			Expect(createdDnsEntry.Spec.DNSName).To(Equal("*.test-domain.com"))
			Expect(createdDnsEntry.Spec.Targets).To(ContainElement("10.0.0.1"))
			Expect(createdDnsEntry.Annotations).To(HaveKeyWithValue("dns.gardener.cloud/class", "garden"))
		})

		It("should reapply disclaimer annotation on DNSEntry when it was removed", func() {
			// given
			k8sClient := createFakeClient()
			Expect(reconcileDNSEntry(context.Background(), k8sClient, "test", "test-ns", "test-domain.com", "10.0.0.1")).Should(Succeed())

			By("removing disclaimer annotation from DNSEntry")
			dnsEntry := dnsv1alpha1.DNSEntry{}
			Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: "test", Namespace: "test-ns"}, &dnsEntry)).Should(Succeed())
			dnsEntry.Annotations = nil
			Expect(k8sClient.Update(context.Background(), &dnsEntry)).Should(Succeed())

			// when
			Expect(reconcileDNSEntry(context.Background(), k8sClient, "test", "test-ns", "test-domain.com", "10.0.0.1")).Should(Succeed())

			// then
			Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: "test", Namespace: "test-ns"}, &dnsEntry)).Should(Succeed())
		})
	})

	Context("fetchIstioIngressGatewayIP", func() {
		It("should return IP from IP-based istio-ingressgateway load-balancer service", func() {
			// given
			svc := getTestIstioIngressGatewayIpBasedService()
			k8sClient := createFakeClient(&svc)

			// when
			ip, err := fetchIstioIngressGatewayIP(context.Background(), k8sClient)

			// then
			Expect(err).ShouldNot(HaveOccurred())
			Expect(ip).To(Equal(testIstioIngressGatewayLoadBalancerIp))
		})

		It("should return hostname from DNS-based istio-ingressgateway load-balancer service", func() {
			// given
			svc := getTestIstioIngressGatewayDnsBasedService()
			k8sClient := createFakeClient(&svc)

			// when
			ip, err := fetchIstioIngressGatewayIP(context.Background(), k8sClient)

			// then
			Expect(err).ShouldNot(HaveOccurred())
			Expect(ip).To(Equal("some.host.name"))
		})
	})
})

func getTestIstioIngressGatewayDnsBasedService() corev1.Service {
	return corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "istio-ingressgateway",
			Namespace: "istio-system",
		},
		Spec: corev1.ServiceSpec{
			ClusterIP: "10.43.158.160",
		}, Status: corev1.ServiceStatus{
			LoadBalancer: corev1.LoadBalancerStatus{
				Ingress: []corev1.LoadBalancerIngress{
					{
						Hostname: "some.host.name",
					},
				},
			},
		},
	}
}
