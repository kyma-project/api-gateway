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

	It("should add finalizer when EnableKymaGateway is true", func() {
		// given
		apiGateway := getApiGateway(true)

		k8sClient := createFakeClient(&apiGateway)

		// when
		status := ReconcileKymaGateway(context.Background(), k8sClient, &apiGateway)

		// then
		Expect(status.IsReady()).To(BeTrue())

		Expect(k8sClient.Get(context.Background(), client.ObjectKeyFromObject(&apiGateway), &apiGateway)).Should(Succeed())
		Expect(apiGateway.GetFinalizers()).To(ContainElement(kymaGatewayFinalizer))
	})

	Context("Non-Gardener cluster", func() {

		It("should not create gateway when Spec doesn't contain EnableKymaGateway flag", func() {
			// given
			apiGateway := v1alpha1.APIGateway{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
			}

			k8sClient := createFakeClient(&apiGateway)

			// when
			status := ReconcileKymaGateway(context.Background(), k8sClient, &apiGateway)

			// then
			Expect(status.IsReady()).To(BeTrue())

			created := v1alpha3.Gateway{}
			err := k8sClient.Get(context.Background(), client.ObjectKey{Name: kymaGatewayName, Namespace: kymaGatewayNamespace}, &created)
			Expect(errors.IsNotFound(err)).To(BeTrue())
		})

		It("should not create gateway when EnableKymaGateway is false", func() {
			// given
			apiGateway := getApiGateway(false)

			k8sClient := createFakeClient(&apiGateway)

			// when
			status := ReconcileKymaGateway(context.Background(), k8sClient, &apiGateway)

			// then
			Expect(status.IsReady()).To(BeTrue())

			created := v1alpha3.Gateway{}
			err := k8sClient.Get(context.Background(), client.ObjectKey{Name: kymaGatewayName, Namespace: kymaGatewayNamespace}, &created)
			Expect(errors.IsNotFound(err)).To(BeTrue())
		})

		It("should create gateway with *.local.kyma.dev hosts when EnableKymaGateway is true and no Gardener shoot-info exists", func() {
			// given
			apiGateway := getApiGateway(true)

			k8sClient := createFakeClient(&apiGateway)

			// when
			status := ReconcileKymaGateway(context.Background(), k8sClient, &apiGateway)

			// then
			Expect(status.IsReady()).To(BeTrue())

			created := v1alpha3.Gateway{}
			Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: kymaGatewayName, Namespace: kymaGatewayNamespace}, &created)).Should(Succeed())

			for _, server := range created.Spec.GetServers() {
				Expect(server.Hosts).To(ContainElement("*.local.kyma.dev"))
			}
		})

		It("should create secret with default certificate when EnableKymaGateway is true and no Gardener shoot-info exists", func() {
			// given
			apiGateway := getApiGateway(true)

			k8sClient := createFakeClient(&apiGateway)

			// when
			status := ReconcileKymaGateway(context.Background(), k8sClient, &apiGateway)

			// then
			Expect(status.IsReady()).To(BeTrue())

			secret := corev1.Secret{}
			Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: "kyma-gateway-certs", Namespace: "istio-system"}, &secret)).Should(Succeed())
			Expect(secret.Data).To(HaveKey("tls.key"))
			Expect(secret.Data).To(HaveKey("tls.crt"))
		})

		It("should not create Certificate and DNSEntry when EnableKymaGateway is true and no Gardener shoot-info exists", func() {
			// given
			apiGateway := getApiGateway(true)

			k8sClient := createFakeClient(&apiGateway)

			// when
			status := ReconcileKymaGateway(context.Background(), k8sClient, &apiGateway)

			// then
			Expect(status.IsReady()).To(BeTrue())

			cert := certv1alpha1.Certificate{}
			err := k8sClient.Get(context.Background(), client.ObjectKey{Name: kymaGatewayCertificateName, Namespace: certificateDefaultNamespace}, &cert)
			Expect(errors.IsNotFound(err)).To(BeTrue())

			dnsEntry := dnsv1alpha1.DNSEntry{}
			err = k8sClient.Get(context.Background(), client.ObjectKey{Name: kymaGatewayDnsEntryName, Namespace: kymaGatewayDnsEntryNamespace}, &dnsEntry)
			Expect(errors.IsNotFound(err)).To(BeTrue())
		})

		It("should delete Kyma gateway and certificate secret and remove finalizer when EnableKymaGateway is updated to false and finalizer is set", func() {
			testShouldDeleteKymaGatewayNonGardenerResources(func(gw v1alpha1.APIGateway) v1alpha1.APIGateway {
				gw.Spec.EnableKymaGateway = ptr.To(false)
				return gw
			})
		})

		It("should delete Kyma gateway and certificate secret and remove finalizer when EnableKymaGateway is removed and finalizer is set in updated APIGateway", func() {
			testShouldDeleteKymaGatewayNonGardenerResources(func(gw v1alpha1.APIGateway) v1alpha1.APIGateway {
				gw.Spec.EnableKymaGateway = nil
				return gw
			})
		})
	})

	Context("Gardener cluster", func() {

		It("should create gateway, DNSEntry and Certificate with shoot-info domain when EnableKymaGateway is true and Gardener shoot-info exists", func() {
			// given
			apiGateway := getApiGateway(true)
			cm := getTestShootInfo()
			igwService := getTestIstioIngressGatewayIpBasedService()

			k8sClient := createFakeClient(&apiGateway, &cm, &igwService)

			// when
			status := ReconcileKymaGateway(context.Background(), k8sClient, &apiGateway)

			// then
			Expect(status.IsReady()).To(BeTrue())

			By("Validating Kyma Gateway")
			createdGateway := v1alpha3.Gateway{}
			Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: kymaGatewayName, Namespace: kymaGatewayNamespace}, &createdGateway)).Should(Succeed())

			for _, server := range createdGateway.Spec.GetServers() {
				Expect(server.Hosts).To(ContainElement("*.some.gardener.domain"))
			}

			By("Validating DNSEntry")
			createdDnsEntry := dnsv1alpha1.DNSEntry{}
			Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: kymaGatewayDnsEntryName, Namespace: kymaGatewayDnsEntryNamespace}, &createdDnsEntry)).Should(Succeed())
			Expect(createdDnsEntry.Spec.DNSName).To(Equal("*.some.gardener.domain"))
			Expect(createdDnsEntry.Spec.Targets).To(ContainElement(testIstioIngressGatewayLoadBalancerIp))

			By("Validating Certificate")
			createdCert := certv1alpha1.Certificate{}
			Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: kymaGatewayCertificateName, Namespace: "istio-system"}, &createdCert)).Should(Succeed())
			Expect(*createdCert.Spec.SecretName).To(Equal(kymaGatewayCertSecretName))
			Expect(*createdCert.Spec.CommonName).To(Equal("*.some.gardener.domain"))
		})

		It("should not create gateway when Spec doesn't contain EnableKymaGateway flag", func() {
			// given
			apiGateway := v1alpha1.APIGateway{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
			}
			cm := getTestShootInfo()

			k8sClient := createFakeClient(&apiGateway, &cm)

			// when
			status := ReconcileKymaGateway(context.Background(), k8sClient, &apiGateway)

			// then
			Expect(status.IsReady()).To(BeTrue())

			created := v1alpha3.Gateway{}
			err := k8sClient.Get(context.Background(), client.ObjectKey{Name: kymaGatewayName, Namespace: kymaGatewayNamespace}, &created)
			Expect(errors.IsNotFound(err)).To(BeTrue())
		})

		It("should not create gateway when EnableKymaGateway is false", func() {
			// given
			apiGateway := getApiGateway(false)
			cm := getTestShootInfo()

			k8sClient := createFakeClient(&apiGateway, &cm)

			// when
			status := ReconcileKymaGateway(context.Background(), k8sClient, &apiGateway)

			// then
			Expect(status.IsReady()).To(BeTrue())

			created := v1alpha3.Gateway{}
			err := k8sClient.Get(context.Background(), client.ObjectKey{Name: kymaGatewayName, Namespace: kymaGatewayNamespace}, &created)
			Expect(errors.IsNotFound(err)).To(BeTrue())
		})

		It("should delete Kyma Gateway, DNSEntry and Certificate and finalizer when shoot-info exists and EnableKymaGateway is updated to false and finalizer is set", func() {
			testShouldDeleteKymaGatewayResources(func(gw v1alpha1.APIGateway) v1alpha1.APIGateway {
				gw.Spec.EnableKymaGateway = ptr.To(false)
				return gw
			})
		})

		It("should delete Kyma Gateway, DNSEntry and Certificate and finalizer when shoot-info exists and EnableKymaGateway is removed and finalizer is set in updated APIGateway", func() {
			testShouldDeleteKymaGatewayResources(func(gw v1alpha1.APIGateway) v1alpha1.APIGateway {
				gw.Spec.EnableKymaGateway = nil
				return gw
			})
		})

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
		status := ReconcileKymaGateway(context.Background(), k8sClient, &apiGateway)
		Expect(status.IsReady()).To(BeTrue())
		kymaGateway := v1alpha3.Gateway{}
		Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: kymaGatewayName, Namespace: kymaGatewayNamespace}, &kymaGateway)).Should(Succeed())

		updatedApiGateway := v1alpha1.APIGateway{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test",
			},
		}

		// when
		status = ReconcileKymaGateway(context.Background(), k8sClient, &updatedApiGateway)

		// then
		Expect(status.IsWarning()).To(BeTrue())
		err := k8sClient.Get(context.Background(), client.ObjectKey{Name: kymaGatewayName, Namespace: kymaGatewayNamespace}, &kymaGateway)
		Expect(err).ShouldNot(HaveOccurred())
	})

})

func testShouldDeleteKymaGatewayNonGardenerResources(updateApiGateway func(gw v1alpha1.APIGateway) v1alpha1.APIGateway) {
	// given
	apiGateway := getApiGateway(true, kymaGatewayFinalizer)

	k8sClient := createFakeClient(&apiGateway)
	status := ReconcileKymaGateway(context.Background(), k8sClient, &apiGateway)
	Expect(status.IsReady()).To(BeTrue())
	kymaGateway := v1alpha3.Gateway{}
	Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: kymaGatewayName, Namespace: kymaGatewayNamespace}, &kymaGateway)).Should(Succeed())

	apiGateway = updateApiGateway(apiGateway)

	// when
	status = ReconcileKymaGateway(context.Background(), k8sClient, &apiGateway)

	// then
	Expect(status.IsReady()).To(BeTrue())

	By("Validating that Gateway is deleted")
	err := k8sClient.Get(context.Background(), client.ObjectKey{Name: kymaGatewayName, Namespace: kymaGatewayNamespace}, &kymaGateway)
	Expect(errors.IsNotFound(err)).To(BeTrue())

	By("Validating that Certificate Secret is deleted")
	s := corev1.Secret{}
	err = k8sClient.Get(context.Background(), client.ObjectKey{Name: kymaGatewayCertSecretName, Namespace: certificateDefaultNamespace}, &s)
	Expect(errors.IsNotFound(err)).To(BeTrue())

	By("Validating that finalizer is removed")
	Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: apiGateway.Name}, &apiGateway)).To(Succeed())
	Expect(apiGateway.GetFinalizers()).ToNot(ContainElement(kymaGatewayFinalizer))
}

func testShouldDeleteKymaGatewayResources(updateApiGateway func(gw v1alpha1.APIGateway) v1alpha1.APIGateway) {
	// given
	apiGateway := getApiGateway(true, kymaGatewayFinalizer)
	cm := getTestShootInfo()
	igwService := getTestIstioIngressGatewayIpBasedService()

	k8sClient := createFakeClient(&apiGateway, &cm, &igwService)
	status := ReconcileKymaGateway(context.Background(), k8sClient, &apiGateway)
	Expect(status.IsReady()).To(BeTrue())
	kymaGateway := v1alpha3.Gateway{}
	Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: kymaGatewayName, Namespace: kymaGatewayNamespace}, &kymaGateway)).Should(Succeed())

	apiGateway = updateApiGateway(apiGateway)

	// when
	status = ReconcileKymaGateway(context.Background(), k8sClient, &apiGateway)

	// then
	Expect(status.IsReady()).To(BeTrue())
	By("Validating that Gateway is deleted")
	err := k8sClient.Get(context.Background(), client.ObjectKey{Name: kymaGatewayName, Namespace: kymaGatewayNamespace}, &kymaGateway)
	Expect(errors.IsNotFound(err)).To(BeTrue())

	By("Validating that DNSEntry is deleted")
	dnsEntry := dnsv1alpha1.DNSEntry{}
	err = k8sClient.Get(context.Background(), client.ObjectKey{Name: kymaGatewayDnsEntryName, Namespace: kymaGatewayDnsEntryNamespace}, &dnsEntry)
	Expect(errors.IsNotFound(err)).To(BeTrue())

	By("Validating that Certificate is deleted")
	cert := certv1alpha1.Certificate{}
	err = k8sClient.Get(context.Background(), client.ObjectKey{Name: kymaGatewayCertificateName, Namespace: certificateDefaultNamespace}, &cert)
	Expect(errors.IsNotFound(err)).To(BeTrue())

	By("Validating that finalizer is removed")
	Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: apiGateway.Name}, &apiGateway)).To(Succeed())
	Expect(apiGateway.GetFinalizers()).ToNot(ContainElement(kymaGatewayFinalizer))
}

func getApiGateway(enableKymaGateway bool, finalizers ...string) v1alpha1.APIGateway {
	return v1alpha1.APIGateway{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test",
			Finalizers: finalizers,
		},
		Spec: v1alpha1.APIGatewaySpec{
			EnableKymaGateway: ptr.To(enableKymaGateway),
		},
	}
}

func getTestIstioIngressGatewayIpBasedService() corev1.Service {
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
