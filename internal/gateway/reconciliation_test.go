package gateway

import (
	"context"
	certv1alpha1 "github.com/gardener/cert-management/pkg/apis/cert/v1alpha1"
	dnsv1alpha1 "github.com/gardener/external-dns-management/pkg/apis/dns/v1alpha1"
	"github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const testIstioIngressGatewayLoadBalancerIp = "172.0.0.1"

var _ = Describe("Kyma Gateway reconciliation", func() {

	It("should not create gateway when Spec doesn't contain EnableKymaGateway flag", func() {
		// given
		apiGateway := v1alpha1.APIGateway{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test",
			},
		}

		k8sClient := createFakeClient(&apiGateway)

		// when
		status := Reconcile(context.TODO(), k8sClient, apiGateway)

		// then
		Expect(status.IsSuccessful()).To(BeTrue())

		created := v1alpha3.Gateway{}
		err := k8sClient.Get(context.TODO(), client.ObjectKey{Name: kymaGatewayName, Namespace: kymaGatewayNamespace}, &created)
		Expect(errors.IsNotFound(err)).To(BeTrue())
	})

	It("should not create gateway when EnableKymaGateway is false", func() {
		// given
		apiGateway := getApiGateway(false)

		k8sClient := createFakeClient(&apiGateway)

		// when
		status := Reconcile(context.TODO(), k8sClient, apiGateway)

		// then
		Expect(status.IsSuccessful()).To(BeTrue())

		created := v1alpha3.Gateway{}
		err := k8sClient.Get(context.TODO(), client.ObjectKey{Name: kymaGatewayName, Namespace: kymaGatewayNamespace}, &created)
		Expect(errors.IsNotFound(err)).To(BeTrue())
	})

	It("should create gateway with *.local.kyma.dev hosts when EnableKymaGateway is true and no Gardener shoot-info exists", func() {
		// given
		apiGateway := getApiGateway(true)

		k8sClient := createFakeClient(&apiGateway)

		// when
		status := Reconcile(context.TODO(), k8sClient, apiGateway)

		// then
		Expect(status.IsSuccessful()).To(BeTrue())

		created := v1alpha3.Gateway{}
		Expect(k8sClient.Get(context.TODO(), client.ObjectKey{Name: kymaGatewayName, Namespace: kymaGatewayNamespace}, &created)).Should(Succeed())

		for _, server := range created.Spec.GetServers() {
			Expect(server.Hosts).To(ContainElement("*.local.kyma.dev"))
		}
	})

	It("should create secret with default certificate when EnableKymaGateway is true and no Gardener shoot-info exists", func() {
		// given
		apiGateway := getApiGateway(true)

		k8sClient := createFakeClient(&apiGateway)

		// when
		status := Reconcile(context.TODO(), k8sClient, apiGateway)

		// then
		Expect(status.IsSuccessful()).To(BeTrue())

		secret := corev1.Secret{}
		Expect(k8sClient.Get(context.TODO(), client.ObjectKey{Name: "kyma-gateway-certs", Namespace: "istio-system"}, &secret)).Should(Succeed())
		Expect(secret.Data).To(HaveKey("tls.key"))
		Expect(secret.Data).To(HaveKey("tls.crt"))
	})

	It("should not create DNSEntry when EnableKymaGateway is true and no Gardener shoot-info exists", func() {
		// given
		apiGateway := getApiGateway(true)

		k8sClient := createFakeClient(&apiGateway)

		// when
		status := Reconcile(context.TODO(), k8sClient, apiGateway)

		// then
		Expect(status.IsSuccessful()).To(BeTrue())

		created := dnsv1alpha1.DNSEntry{}
		err := k8sClient.Get(context.TODO(), client.ObjectKey{Name: kymaGatewayDnsEntryName, Namespace: kymaGatewayDnsEntryNamespace}, &created)
		Expect(errors.IsNotFound(err)).To(BeTrue())
	})

	It("should not create Certificate when EnableKymaGateway is true and no Gardener shoot-info exists", func() {
		// given
		apiGateway := getApiGateway(true)

		k8sClient := createFakeClient(&apiGateway)

		// when
		status := Reconcile(context.TODO(), k8sClient, apiGateway)

		// then
		Expect(status.IsSuccessful()).To(BeTrue())

		created := certv1alpha1.Certificate{}
		err := k8sClient.Get(context.TODO(), client.ObjectKey{Name: kymaGatewayCertificateName, Namespace: certificateDefaultNamespace}, &created)
		Expect(errors.IsNotFound(err)).To(BeTrue())
	})

	It("should create gateway, DNSEntry and Certificate with shoot-info domain when EnableKymaGateway is true and Gardener shoot-info exists", func() {
		// given
		apiGateway := getApiGateway(true)
		cm := getTestShootInfo()
		igwService := getTestIstioIngressGatewayService()

		k8sClient := createFakeClient(&apiGateway, &cm, &igwService)

		// when
		status := Reconcile(context.TODO(), k8sClient, apiGateway)

		// then
		Expect(status.IsSuccessful()).To(BeTrue())

		By("Validating Kyma Gateway")
		createdGateway := v1alpha3.Gateway{}
		Expect(k8sClient.Get(context.TODO(), client.ObjectKey{Name: kymaGatewayName, Namespace: kymaGatewayNamespace}, &createdGateway)).Should(Succeed())

		for _, server := range createdGateway.Spec.GetServers() {
			Expect(server.Hosts).To(ContainElement("*.some.gardener.domain"))
		}

		By("Validating DNSEntry")
		createdDnsEntry := dnsv1alpha1.DNSEntry{}
		Expect(k8sClient.Get(context.TODO(), client.ObjectKey{Name: kymaGatewayDnsEntryName, Namespace: kymaGatewayDnsEntryNamespace}, &createdDnsEntry)).Should(Succeed())
		Expect(createdDnsEntry.Spec.DNSName).To(Equal("*.some.gardener.domain"))
		Expect(createdDnsEntry.Spec.Targets).To(ContainElement(testIstioIngressGatewayLoadBalancerIp))

		By("Validating Certificate")
		createdCert := certv1alpha1.Certificate{}
		Expect(k8sClient.Get(context.TODO(), client.ObjectKey{Name: kymaGatewayCertificateName, Namespace: "istio-system"}, &createdCert)).Should(Succeed())
		Expect(*createdCert.Spec.SecretName).To(Equal(kymaGatewayCertSecretName))
		Expect(*createdCert.Spec.CommonName).To(Equal("*.some.gardener.domain"))
	})

	It("should delete Kyma gateway when EnableKymaGateway is updated to false", func() {
		updatedApiGateway := getApiGateway(false)

		testShouldDeleteKymaGateway(updatedApiGateway)
	})

	It("should delete Kyma gateway when EnableKymaGateway is removed in updated APIGateway", func() {
		updatedApiGateway := v1alpha1.APIGateway{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test",
			},
		}
		testShouldDeleteKymaGateway(updatedApiGateway)
	})

	It("should delete Kyma Gateway, DNSEntry and Certificate when shoot-info exists and EnableKymaGateway is updated to false", func() {
		updatedApiGateway := getApiGateway(false)
		testShouldDeleteKymaGatewayResources(updatedApiGateway)
	})

	It("should delete Kyma Gateway, DNSEntry and Certificate when shoot-info exists and EnableKymaGateway is removed in updated APIGateway", func() {
		updatedApiGateway := v1alpha1.APIGateway{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test",
			},
		}
		testShouldDeleteKymaGatewayResources(updatedApiGateway)
	})

	It("should not delete Kyma Gateway when EnableKymaGateway is updated to false, but any APIRule exists", func() {
		// given
		apiGateway := getApiGateway(true)

		apiRule := v1beta1.APIRule{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test",
			},
		}

		k8sClient := createFakeClient(&apiGateway, &apiRule)
		status := Reconcile(context.TODO(), k8sClient, apiGateway)
		Expect(status.IsSuccessful()).To(BeTrue())
		kymaGateway := v1alpha3.Gateway{}
		Expect(k8sClient.Get(context.TODO(), client.ObjectKey{Name: kymaGatewayName, Namespace: kymaGatewayNamespace}, &kymaGateway)).Should(Succeed())

		updatedApiGateway := v1alpha1.APIGateway{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test",
			},
		}

		// when
		status = Reconcile(context.TODO(), k8sClient, updatedApiGateway)

		// then
		Expect(status.IsWarning()).To(BeTrue())
		err := k8sClient.Get(context.TODO(), client.ObjectKey{Name: kymaGatewayName, Namespace: kymaGatewayNamespace}, &kymaGateway)
		Expect(err).ShouldNot(HaveOccurred())
	})

})

func testShouldDeleteKymaGateway(updatedApiGateway v1alpha1.APIGateway) {
	// given
	apiGateway := v1alpha1.APIGateway{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: v1alpha1.APIGatewaySpec{
			EnableKymaGateway: ptr.To(true),
		},
	}

	k8sClient := createFakeClient(&apiGateway)
	status := Reconcile(context.TODO(), k8sClient, apiGateway)
	Expect(status.IsSuccessful()).To(BeTrue())
	kymaGateway := v1alpha3.Gateway{}
	Expect(k8sClient.Get(context.TODO(), client.ObjectKey{Name: kymaGatewayName, Namespace: kymaGatewayNamespace}, &kymaGateway)).Should(Succeed())

	// when
	status = Reconcile(context.TODO(), k8sClient, updatedApiGateway)

	// then
	Expect(status.IsSuccessful()).To(BeTrue())
	err := k8sClient.Get(context.TODO(), client.ObjectKey{Name: kymaGatewayName, Namespace: kymaGatewayNamespace}, &kymaGateway)
	Expect(errors.IsNotFound(err)).To(BeTrue())
}

func testShouldDeleteKymaGatewayResources(updatedApiGateway v1alpha1.APIGateway) {
	// given
	apiGateway := v1alpha1.APIGateway{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: v1alpha1.APIGatewaySpec{
			EnableKymaGateway: ptr.To(true),
		},
	}
	cm := getTestShootInfo()
	igwService := getTestIstioIngressGatewayService()

	k8sClient := createFakeClient(&apiGateway, &cm, &igwService)
	status := Reconcile(context.TODO(), k8sClient, apiGateway)
	Expect(status.IsSuccessful()).To(BeTrue())
	kymaGateway := v1alpha3.Gateway{}
	Expect(k8sClient.Get(context.TODO(), client.ObjectKey{Name: kymaGatewayName, Namespace: kymaGatewayNamespace}, &kymaGateway)).Should(Succeed())

	// when
	status = Reconcile(context.TODO(), k8sClient, updatedApiGateway)

	// then
	Expect(status.IsSuccessful()).To(BeTrue())
	By("Validating that Gateway is deleted")
	err := k8sClient.Get(context.TODO(), client.ObjectKey{Name: kymaGatewayName, Namespace: kymaGatewayNamespace}, &kymaGateway)
	Expect(errors.IsNotFound(err)).To(BeTrue())

	By("Validating that DNSEntry is deleted")
	dnsEntry := dnsv1alpha1.DNSEntry{}
	err = k8sClient.Get(context.TODO(), client.ObjectKey{Name: kymaGatewayDnsEntryName, Namespace: kymaGatewayDnsEntryNamespace}, &dnsEntry)
	Expect(errors.IsNotFound(err)).To(BeTrue())

	By("Validating that Certificate is deleted")
	cert := certv1alpha1.Certificate{}
	err = k8sClient.Get(context.TODO(), client.ObjectKey{Name: kymaGatewayCertificateName, Namespace: certificateDefaultNamespace}, &cert)
	Expect(errors.IsNotFound(err)).To(BeTrue())
}

func getApiGateway(enableKymaGateway bool) v1alpha1.APIGateway {
	return v1alpha1.APIGateway{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: v1alpha1.APIGatewaySpec{
			EnableKymaGateway: ptr.To(enableKymaGateway),
		},
	}
}

func getTestIstioIngressGatewayService() corev1.Service {
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
						IP: testIstioIngressGatewayLoadBalancerIp,
					},
				},
			},
		},
	}
}

func getTestShootInfo() corev1.ConfigMap {
	return corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "shoot-info",
			Namespace: "kube-system",
		},
		Data: map[string]string{
			"domain": "some.gardener.domain",
		},
	}
}
