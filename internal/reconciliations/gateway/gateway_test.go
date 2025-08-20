package gateway

import (
	"context"
	"github.com/kyma-project/api-gateway/internal/conditions"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	certv1alpha1 "github.com/gardener/cert-management/pkg/apis/cert/v1alpha1"
	dnsv1alpha1 "github.com/gardener/external-dns-management/pkg/apis/dns/v1alpha1"
	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	"github.com/kyma-project/api-gateway/controllers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	istioapiv1beta1 "istio.io/api/networking/v1beta1"
	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	testIstioIngressGatewayLoadBalancerIp = "172.0.0.1"
	resourceListPath                      = "../../../manifests/controlled_resources_list.yaml"
)

type mockAPIRuleReconciliationStarter struct {
	setupError error
	stopError  error
}

func (m mockAPIRuleReconciliationStarter) SetupAndStartManager() error {
	return m.setupError
}

func (m mockAPIRuleReconciliationStarter) StopManager() error {
	return m.stopError
}

var _ = Describe("Kyma Gateway reconciliation", func() {
	It("Should add finalizer when EnableKymaGateway is true", func() {
		// given
		apiGateway := getApiGateway(true)

		k8sClient := createFakeClient(&apiGateway)

		// when
		status := ReconcileKymaGateway(context.Background(), k8sClient, &apiGateway, resourceListPath, mockAPIRuleReconciliationStarter{})

		// then
		Expect(status.IsReady()).To(BeTrue())

		Expect(k8sClient.Get(context.Background(), client.ObjectKeyFromObject(&apiGateway), &apiGateway)).Should(Succeed())
		Expect(apiGateway.GetFinalizers()).To(ContainElement(KymaGatewayFinalizer))
	})

	It("Should add condition on successful reconcile ", func() {
		// given
		apiGateway := getApiGateway(true)

		k8sClient := createFakeClient(&apiGateway)

		// when
		status := ReconcileKymaGateway(context.Background(), k8sClient, &apiGateway, resourceListPath, mockAPIRuleReconciliationStarter{})

		// then
		Expect(status.IsReady()).To(BeTrue())
		Expect(status.Condition()).To(Not(BeNil()))
		Expect(status.Condition().Type).To(Equal(conditions.KymaGatewayReconcileSucceeded.Condition().Type))
		Expect(status.Condition().Reason).To(Equal(conditions.KymaGatewayReconcileSucceeded.Condition().Reason))
		Expect(status.Condition().Status).To(Equal(metav1.ConditionFalse))
	})

	It("Should name of blocking resource in condition for warning", func() {
		blockingVs := getVirtualService(KymaGatewayFullName)
		status := testShouldDeleteKymaGatewayNonGardenerResources(func(gw v1alpha1.APIGateway) v1alpha1.APIGateway {
			gw.Spec.EnableKymaGateway = ptr.To(false)
			return gw
		}, controllers.Warning, BeFalse(), ContainElement(KymaGatewayFinalizer), &blockingVs)

		Expect(status.Condition()).To(Not(BeNil()))
		Expect(status.Condition().Type).To(Equal(conditions.KymaGatewayDeletionBlocked.Condition().Type))
		Expect(status.Condition().Reason).To(Equal(conditions.KymaGatewayDeletionBlocked.Condition().Reason))
		Expect(status.Condition().Status).To(Equal(metav1.ConditionFalse))
		Expect(status.Condition().Message).To(ContainSubstring(blockingVs.Name))
	})

	Context("Non-Gardener cluster", func() {
		It("Should not create gateway when Spec doesn't contain EnableKymaGateway flag", func() {
			// given
			apiGateway := v1alpha1.APIGateway{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
			}

			k8sClient := createFakeClient(&apiGateway)

			// when
			status := ReconcileKymaGateway(context.Background(), k8sClient, &apiGateway, resourceListPath, mockAPIRuleReconciliationStarter{})

			// then
			Expect(status.IsReady()).To(BeTrue())

			created := v1alpha3.Gateway{}
			err := k8sClient.Get(context.Background(), client.ObjectKey{Name: KymaGatewayName, Namespace: KymaGatewayNamespace}, &created)
			Expect(errors.IsNotFound(err)).To(BeTrue())
		})

		It("Should not create gateway when EnableKymaGateway is false", func() {
			// given
			apiGateway := getApiGateway(false)

			k8sClient := createFakeClient(&apiGateway)

			// when
			status := ReconcileKymaGateway(context.Background(), k8sClient, &apiGateway, resourceListPath, mockAPIRuleReconciliationStarter{})

			// then
			Expect(status.IsReady()).To(BeTrue())

			created := v1alpha3.Gateway{}
			err := k8sClient.Get(context.Background(), client.ObjectKey{Name: KymaGatewayName, Namespace: KymaGatewayNamespace}, &created)
			Expect(errors.IsNotFound(err)).To(BeTrue())
		})

		It("Should create gateway with *.local.kyma.dev hosts when EnableKymaGateway is true and no Gardener shoot-info exists", func() {
			// given
			apiGateway := getApiGateway(true)

			k8sClient := createFakeClient(&apiGateway)

			// when
			status := ReconcileKymaGateway(context.Background(), k8sClient, &apiGateway, resourceListPath, mockAPIRuleReconciliationStarter{})

			// then
			Expect(status.IsReady()).To(BeTrue())

			created := v1alpha3.Gateway{}
			Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: KymaGatewayName, Namespace: KymaGatewayNamespace}, &created)).Should(Succeed())

			for _, server := range created.Spec.GetServers() {
				Expect(server.Hosts).To(ContainElement("*.local.kyma.dev"))
			}
		})

		It("Should create secret with default certificate when EnableKymaGateway is true and no Gardener shoot-info exists", func() {
			// given
			apiGateway := getApiGateway(true)

			k8sClient := createFakeClient(&apiGateway)

			// when
			status := ReconcileKymaGateway(context.Background(), k8sClient, &apiGateway, resourceListPath, mockAPIRuleReconciliationStarter{})

			// then
			Expect(status.IsReady()).To(BeTrue())

			secret := corev1.Secret{}
			Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: kymaGatewayCertSecretName, Namespace: certificateDefaultNamespace}, &secret)).Should(Succeed())
			Expect(secret.Data).To(HaveKey("tls.key"))
			Expect(secret.Data).To(HaveKey("tls.crt"))
		})

		It("Should create secret with default certificate and no Gardener resources "+
			"when Kyma Gateway is enabled and no domain is defined in Gardener shoot-info", func() {
			// given
			apiGateway := getApiGateway(true)

			cm := corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "shoot-info",
					Namespace: "kube-system",
				},
			}

			k8sClient := createFakeClient(&apiGateway, &cm)

			// when
			status := ReconcileKymaGateway(context.Background(), k8sClient, &apiGateway, resourceListPath, mockAPIRuleReconciliationStarter{})

			// then
			Expect(status.IsReady()).To(BeTrue())

			secret := corev1.Secret{}
			Expect(k8sClient.Get(context.Background(),
				client.ObjectKey{Name: kymaGatewayCertSecretName, Namespace: certificateDefaultNamespace}, &secret)).
				Should(Succeed())

			Expect(secret.Data).To(HaveKey("tls.key"))
			Expect(secret.Data).To(HaveKey("tls.crt"))

			cert := certv1alpha1.Certificate{}
			err := k8sClient.Get(context.Background(),
				client.ObjectKey{Name: kymaGatewayCertificateName, Namespace: certificateDefaultNamespace}, &cert)
			Expect(errors.IsNotFound(err)).To(BeTrue())

			dnsEntry := dnsv1alpha1.DNSEntry{}
			err = k8sClient.Get(context.Background(),
				client.ObjectKey{Name: kymaGatewayDnsEntryName, Namespace: kymaGatewayDnsEntryNamespace}, &dnsEntry)
			Expect(errors.IsNotFound(err)).To(BeTrue())
		})

		It("Should create secret with istio-healthz virtual service when EnableKymaGateway is true and no Gardener shoot-info exists", func() {
			// given
			apiGateway := getApiGateway(true)

			k8sClient := createFakeClient(&apiGateway)

			// when
			status := ReconcileKymaGateway(context.Background(), k8sClient, &apiGateway, resourceListPath, mockAPIRuleReconciliationStarter{})

			// then
			Expect(status.IsReady()).To(BeTrue())

			createdVs := networkingv1beta1.VirtualService{}
			Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: kymaGatewayVirtualServiceName, Namespace: kymaGatewayVirtualServiceNamespace}, &createdVs)).Should(Succeed())
		})

		It("Should not create Certificate and DNSEntry when EnableKymaGateway is true and no Gardener shoot-info exists", func() {
			// given
			apiGateway := getApiGateway(true)

			k8sClient := createFakeClient(&apiGateway)

			// when
			status := ReconcileKymaGateway(context.Background(), k8sClient, &apiGateway, resourceListPath, mockAPIRuleReconciliationStarter{})

			// then
			Expect(status.IsReady()).To(BeTrue())

			cert := certv1alpha1.Certificate{}
			err := k8sClient.Get(context.Background(), client.ObjectKey{Name: kymaGatewayCertificateName, Namespace: certificateDefaultNamespace}, &cert)
			Expect(errors.IsNotFound(err)).To(BeTrue())

			dnsEntry := dnsv1alpha1.DNSEntry{}
			err = k8sClient.Get(context.Background(), client.ObjectKey{Name: kymaGatewayDnsEntryName, Namespace: kymaGatewayDnsEntryNamespace}, &dnsEntry)
			Expect(errors.IsNotFound(err)).To(BeTrue())
		})

		It("Should delete Kyma gateway, certificate secret, virtual service and remove finalizer when EnableKymaGateway is updated to false and finalizer is set", func() {
			testShouldDeleteKymaGatewayNonGardenerResources(func(gw v1alpha1.APIGateway) v1alpha1.APIGateway {
				gw.Spec.EnableKymaGateway = ptr.To(false)
				return gw
			}, controllers.Ready, BeTrue(), Not(ContainElement(KymaGatewayFinalizer)))
		})

		It("Should delete Kyma gateway, certificate secret, virtual service  and remove finalizer when EnableKymaGateway is removed and finalizer is set in updated APIGateway", func() {
			testShouldDeleteKymaGatewayNonGardenerResources(func(gw v1alpha1.APIGateway) v1alpha1.APIGateway {
				gw.Spec.EnableKymaGateway = nil
				return gw
			}, controllers.Ready, BeTrue(), Not(ContainElement(KymaGatewayFinalizer)))
		})

		It("Should not delete Kyma Gateway, certificate secret, virtual service and finalizer when EnableKymaGateway is updated to false but there is blocking APIRule", func() {
			apiRule := getApiRule(KymaGatewayFullName)
			status := testShouldDeleteKymaGatewayNonGardenerResources(func(gw v1alpha1.APIGateway) v1alpha1.APIGateway {
				gw.Spec.EnableKymaGateway = ptr.To(false)
				return gw
			}, controllers.Warning, BeFalse(), ContainElement(KymaGatewayFinalizer), &apiRule)

			Expect(status.NestedError().Error()).To(Equal("could not delete Kyma Gateway since there are 1 custom resource(s) present that block its deletion"))
			Expect(status.Description()).To(Equal("There are custom resources that block the deletion of Kyma Gateway. Please take a look at kyma-system/api-gateway-controller-manager logs to see more information about the warning"))
			Expect(status.Condition().Status).To(Equal(metav1.ConditionFalse))
			Expect(status.Condition().Reason).To(Equal("KymaGatewayDeletionBlocked"))
			Expect(status.Condition().Message).To(Equal("Kyma Gateway deletion blocked because of the existing custom resources: api-rule"))
		})

		It("Should not delete Kyma Gateway, certificate secret, virtual service and finalizer when EnableKymaGateway is updated to false but there is blocking VirtualService", func() {
			vs := getVirtualService(KymaGatewayFullName)
			status := testShouldDeleteKymaGatewayNonGardenerResources(func(gw v1alpha1.APIGateway) v1alpha1.APIGateway {
				gw.Spec.EnableKymaGateway = ptr.To(false)
				return gw
			}, controllers.Warning, BeFalse(), ContainElement(KymaGatewayFinalizer), &vs)

			Expect(status.NestedError().Error()).To(Equal("could not delete Kyma Gateway since there are 1 custom resource(s) present that block its deletion"))
			Expect(status.Description()).To(Equal("There are custom resources that block the deletion of Kyma Gateway. Please take a look at kyma-system/api-gateway-controller-manager logs to see more information about the warning"))
			Expect(status.Condition().Status).To(Equal(metav1.ConditionFalse))
			Expect(status.Condition().Reason).To(Equal("KymaGatewayDeletionBlocked"))
			Expect(status.Condition().Message).To(Equal("Kyma Gateway deletion blocked because of the existing custom resources: virtual-service"))
		})
	})

	Context("Gardener cluster", func() {
		It("Should create gateway, Virtual Service, DNSEntry and Certificate with shoot-info domain when EnableKymaGateway is true and Gardener shoot-info exists", func() {
			// given
			apiGateway := getApiGateway(true)
			cm := getTestShootInfo()
			igwService := getTestIstioIngressGatewayIpBasedService()

			k8sClient := createFakeClient(&apiGateway, &cm, &igwService,
				&v1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "dnsentries.dns.gardener.cloud"}},
				&v1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "certificates.cert.gardener.cloud"}},
			)

			// when
			status := ReconcileKymaGateway(context.Background(), k8sClient, &apiGateway, resourceListPath, mockAPIRuleReconciliationStarter{})

			// then
			Expect(status.IsReady()).To(BeTrue())

			By("Validating Kyma Gateway")
			createdGateway := v1alpha3.Gateway{}
			Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: KymaGatewayName, Namespace: KymaGatewayNamespace}, &createdGateway)).Should(Succeed())

			for _, server := range createdGateway.Spec.GetServers() {
				Expect(server.Hosts).To(ContainElement("*.some.gardener.domain"))
			}

			By("Validating istio-healthz Virtual Service")
			createdVs := networkingv1beta1.VirtualService{}
			Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: kymaGatewayVirtualServiceName, Namespace: kymaGatewayVirtualServiceNamespace}, &createdVs)).Should(Succeed())

			By("Validating DNSEntry")
			createdDnsEntry := dnsv1alpha1.DNSEntry{}
			Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: kymaGatewayDnsEntryName, Namespace: kymaGatewayDnsEntryNamespace}, &createdDnsEntry)).Should(Succeed())
			Expect(createdDnsEntry.Spec.DNSName).To(Equal("*.some.gardener.domain"))
			Expect(createdDnsEntry.Spec.Targets).To(ContainElement(testIstioIngressGatewayLoadBalancerIp))

			By("Validating Certificate")
			createdCert := certv1alpha1.Certificate{}
			Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: kymaGatewayCertificateName, Namespace: certificateDefaultNamespace}, &createdCert)).Should(Succeed())
			Expect(*createdCert.Spec.SecretName).To(Equal(kymaGatewayCertSecretName))
			Expect(*createdCert.Spec.CommonName).To(Equal("*.some.gardener.domain"))
		})

		It("Should not create gateway when Spec doesn't contain EnableKymaGateway flag", func() {
			// given
			apiGateway := v1alpha1.APIGateway{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
			}
			cm := getTestShootInfo()

			k8sClient := createFakeClient(&apiGateway, &cm)

			// when
			status := ReconcileKymaGateway(context.Background(), k8sClient, &apiGateway, resourceListPath, mockAPIRuleReconciliationStarter{})

			// then
			Expect(status.IsReady()).To(BeTrue())

			created := v1alpha3.Gateway{}
			err := k8sClient.Get(context.Background(), client.ObjectKey{Name: KymaGatewayName, Namespace: KymaGatewayNamespace}, &created)
			Expect(errors.IsNotFound(err)).To(BeTrue())
		})

		It("Should not create gateway when EnableKymaGateway is false", func() {
			// given
			apiGateway := getApiGateway(false)
			cm := getTestShootInfo()

			k8sClient := createFakeClient(&apiGateway, &cm)

			// when
			status := ReconcileKymaGateway(context.Background(), k8sClient, &apiGateway, resourceListPath, mockAPIRuleReconciliationStarter{})

			// then
			Expect(status.IsReady()).To(BeTrue())

			created := v1alpha3.Gateway{}
			err := k8sClient.Get(context.Background(), client.ObjectKey{Name: KymaGatewayName, Namespace: KymaGatewayNamespace}, &created)
			Expect(errors.IsNotFound(err)).To(BeTrue())
		})

		It("Should delete Kyma Gateway, Virtual Service, DNSEntry and Certificate and finalizer when shoot-info exists and EnableKymaGateway is updated to false and finalizer is set", func() {
			testShouldDeleteKymaGatewayResources(func(gw v1alpha1.APIGateway) v1alpha1.APIGateway {
				gw.Spec.EnableKymaGateway = ptr.To(false)
				return gw
			}, controllers.Ready, BeTrue(), Not(ContainElement(KymaGatewayFinalizer)))
		})

		It("Should delete Kyma Gateway, Virtual Service, DNSEntry and Certificate and finalizer when shoot-info exists and EnableKymaGateway is removed and finalizer is set in updated APIGateway", func() {
			testShouldDeleteKymaGatewayResources(func(gw v1alpha1.APIGateway) v1alpha1.APIGateway {
				gw.Spec.EnableKymaGateway = nil
				return gw
			}, controllers.Ready, BeTrue(), Not(ContainElement(KymaGatewayFinalizer)))
		})

		It("Should not delete Kyma Gateway, Virtual Service, DNSEntry and Certificate and finalizer when EnableKymaGateway is updated to false but there is blocking APIRule", func() {
			apiRule := getApiRule(KymaGatewayFullName)
			status := testShouldDeleteKymaGatewayResources(func(gw v1alpha1.APIGateway) v1alpha1.APIGateway {
				gw.Spec.EnableKymaGateway = ptr.To(false)
				return gw
			}, controllers.Warning, BeFalse(), ContainElement(KymaGatewayFinalizer), &apiRule)

			Expect(status.NestedError().Error()).To(Equal("could not delete Kyma Gateway since there are 1 custom resource(s) present that block its deletion"))
			Expect(status.Description()).To(Equal("There are custom resources that block the deletion of Kyma Gateway. Please take a look at kyma-system/api-gateway-controller-manager logs to see more information about the warning"))
			Expect(status.Condition().Status).To(Equal(metav1.ConditionFalse))
			Expect(status.Condition().Reason).To(Equal("KymaGatewayDeletionBlocked"))
			Expect(status.Condition().Message).To(Equal("Kyma Gateway deletion blocked because of the existing custom resources: api-rule"))
		})

		It("Should not delete Kyma Gateway, Virtual Service, DNSEntry and Certificate and finalizer when EnableKymaGateway is updated to false but there is blocking VirtualService", func() {
			vs := getVirtualService(KymaGatewayFullName)
			status := testShouldDeleteKymaGatewayResources(func(gw v1alpha1.APIGateway) v1alpha1.APIGateway {
				gw.Spec.EnableKymaGateway = ptr.To(false)
				return gw
			}, controllers.Warning, BeFalse(), ContainElement(KymaGatewayFinalizer), &vs)

			Expect(status.NestedError().Error()).To(Equal("could not delete Kyma Gateway since there are 1 custom resource(s) present that block its deletion"))
			Expect(status.Description()).To(Equal("There are custom resources that block the deletion of Kyma Gateway. Please take a look at kyma-system/api-gateway-controller-manager logs to see more information about the warning"))
			Expect(status.Condition().Status).To(Equal(metav1.ConditionFalse))
			Expect(status.Condition().Reason).To(Equal("KymaGatewayDeletionBlocked"))
			Expect(status.Condition().Message).To(Equal("Kyma Gateway deletion blocked because of the existing custom resources: virtual-service"))
		})
	})
})

func testShouldDeleteKymaGatewayNonGardenerResources(updateApiGateway func(gw v1alpha1.APIGateway) v1alpha1.APIGateway, state controllers.State, nfMatcher types.GomegaMatcher, fMatcher types.GomegaMatcher, objs ...client.Object) controllers.Status {
	// given
	apiGateway := getApiGateway(true, KymaGatewayFinalizer)
	objs = append(objs, &apiGateway)

	k8sClient := createFakeClient(objs...)
	status := ReconcileKymaGateway(context.Background(), k8sClient, &apiGateway, resourceListPath, mockAPIRuleReconciliationStarter{})
	Expect(status.IsReady()).To(BeTrue())
	kymaGateway := v1alpha3.Gateway{}
	Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: KymaGatewayName, Namespace: KymaGatewayNamespace}, &kymaGateway)).Should(Succeed())

	apiGateway = updateApiGateway(apiGateway)

	// when
	status = ReconcileKymaGateway(context.Background(), k8sClient, &apiGateway, resourceListPath, mockAPIRuleReconciliationStarter{})

	// then
	Expect(status.State()).To(Equal(state))

	By("Validating that Gateway is deleted")
	err := k8sClient.Get(context.Background(), client.ObjectKey{Name: KymaGatewayName, Namespace: KymaGatewayNamespace}, &kymaGateway)
	Expect(errors.IsNotFound(err)).To(nfMatcher)

	By("Validating that Certificate Secret is deleted")
	s := corev1.Secret{}
	err = k8sClient.Get(context.Background(), client.ObjectKey{Name: kymaGatewayCertSecretName, Namespace: certificateDefaultNamespace}, &s)
	Expect(errors.IsNotFound(err)).To(nfMatcher)

	By("Validating that istio-healthz Virtual Service is deleted")
	vs := networkingv1beta1.VirtualService{}
	err = k8sClient.Get(context.Background(), client.ObjectKey{Name: kymaGatewayVirtualServiceName, Namespace: kymaGatewayVirtualServiceNamespace}, &vs)
	Expect(errors.IsNotFound(err)).To(nfMatcher)

	By("Validating that finalizer is removed")
	Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: apiGateway.Name}, &apiGateway)).To(Succeed())
	Expect(apiGateway.GetFinalizers()).To(fMatcher)

	return status
}

func testShouldDeleteKymaGatewayResources(updateApiGateway func(gw v1alpha1.APIGateway) v1alpha1.APIGateway, state controllers.State, nfMatcher types.GomegaMatcher, fMatcher types.GomegaMatcher, objs ...client.Object) controllers.Status {
	// given
	apiGateway := getApiGateway(true, KymaGatewayFinalizer)
	objs = append(objs, &apiGateway)

	cm := getTestShootInfo()
	igwService := getTestIstioIngressGatewayIpBasedService()
	objs = append(objs, &cm, &igwService,
		&v1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "dnsentries.dns.gardener.cloud"}},
		&v1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "certificates.cert.gardener.cloud"}})

	k8sClient := createFakeClient(objs...)
	status := ReconcileKymaGateway(context.Background(), k8sClient, &apiGateway, resourceListPath, mockAPIRuleReconciliationStarter{})
	Expect(status.IsReady()).To(BeTrue())
	kymaGateway := v1alpha3.Gateway{}
	Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: KymaGatewayName, Namespace: KymaGatewayNamespace}, &kymaGateway)).Should(Succeed())

	apiGateway = updateApiGateway(apiGateway)

	// when
	status = ReconcileKymaGateway(context.Background(), k8sClient, &apiGateway, resourceListPath, mockAPIRuleReconciliationStarter{})

	// then
	Expect(status.State()).To(Equal(state))

	By("Validating that Gateway is deleted")
	err := k8sClient.Get(context.Background(), client.ObjectKey{Name: KymaGatewayName, Namespace: KymaGatewayNamespace}, &kymaGateway)
	Expect(errors.IsNotFound(err)).To(nfMatcher)

	By("Validating that DNSEntry is deleted")
	dnsEntry := dnsv1alpha1.DNSEntry{}
	err = k8sClient.Get(context.Background(), client.ObjectKey{Name: kymaGatewayDnsEntryName, Namespace: kymaGatewayDnsEntryNamespace}, &dnsEntry)
	Expect(errors.IsNotFound(err)).To(nfMatcher)

	By("Validating that Certificate is deleted")
	cert := certv1alpha1.Certificate{}
	err = k8sClient.Get(context.Background(), client.ObjectKey{Name: kymaGatewayCertificateName, Namespace: certificateDefaultNamespace}, &cert)
	Expect(errors.IsNotFound(err)).To(nfMatcher)

	By("Validating that istio-healthz Virtual Service is deleted")
	vs := networkingv1beta1.VirtualService{}
	err = k8sClient.Get(context.Background(), client.ObjectKey{Name: kymaGatewayVirtualServiceName, Namespace: kymaGatewayVirtualServiceNamespace}, &vs)
	Expect(errors.IsNotFound(err)).To(nfMatcher)

	By("Validating that finalizer is removed")
	Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: apiGateway.Name}, &apiGateway)).To(Succeed())
	Expect(apiGateway.GetFinalizers()).To(fMatcher)

	return status
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

func getApiRule(gateway string) gatewayv1beta1.APIRule {
	return gatewayv1beta1.APIRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "api-rule",
			Namespace: "default",
		},
		Spec: gatewayv1beta1.APIRuleSpec{
			Gateway: &gateway,
		},
	}
}

func getVirtualService(gateway string) networkingv1beta1.VirtualService {
	return networkingv1beta1.VirtualService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "virtual-service",
			Namespace: "default",
		},
		Spec: istioapiv1beta1.VirtualService{
			Gateways: []string{gateway},
		},
	}
}
