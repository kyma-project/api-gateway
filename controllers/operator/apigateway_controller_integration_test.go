package operator

import (
	"context"
	"fmt"
	dnsv1alpha1 "github.com/gardener/external-dns-management/pkg/apis/dns/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"math/rand"
	"time"

	"github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"github.com/kyma-project/api-gateway/internal/reconciliations/gateway"
	. "github.com/onsi/gomega"
	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/utils/ptr"

	ratelimitv1alpha1 "github.com/kyma-project/api-gateway/apis/gateway/ratelimit/v1alpha1"
	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	apinetworkingv1beta1 "istio.io/api/networking/v1beta1"
	networkingv1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Tests needs to be executed serially because of the shared cluster-wide resources like the APIGateway CR.
var _ = Describe("API Gateway Controller", Serial, func() {
	AfterEach(func() {
		deleteApiRules()
		deleteVirtualServices()
		deleteRateLimitRules()
		deleteApiGateways()
	})

	Context("APIGateway CR", func() {
		It("Should set ready state on APIGateway CR when reconciliation succeeds", func() {
			// given
			apiGateway := v1alpha1.APIGateway{
				ObjectMeta: metav1.ObjectMeta{
					Name: generateName(),
				},
			}

			// when
			Expect(k8sClient.Create(context.Background(), &apiGateway)).Should(Succeed())

			// then
			Eventually(func(g Gomega) {
				created := v1alpha1.APIGateway{}
				g.Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: apiGateway.Name}, &created)).Should(Succeed())
				g.Expect(created.ObjectMeta.Finalizers).To(HaveLen(1))
				g.Expect(created.ObjectMeta.Finalizers[0]).To(Equal(ApiGatewayFinalizer))
				g.Expect(created.Status.State).To(Equal(v1alpha1.Ready))
			}, eventuallyTimeout).Should(Succeed())

			Expect(k8sClient.Delete(context.Background(), &apiGateway)).Should(Succeed())
		})

		It("Should set DNSEntry according to istio-ingressgateway svc", func() {
			// given
			apiGateway := v1alpha1.APIGateway{
				ObjectMeta: metav1.ObjectMeta{
					Name: generateName(),
				},
			}

			// when
			Expect(k8sClient.Create(context.Background(), &apiGateway)).Should(Succeed())
			Eventually(func(g Gomega) {
				created := v1alpha1.APIGateway{}
				g.Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: apiGateway.Name}, &created)).Should(Succeed())
				g.Expect(created.ObjectMeta.Finalizers).To(HaveLen(1))
				g.Expect(created.ObjectMeta.Finalizers[0]).To(Equal(ApiGatewayFinalizer))
				g.Expect(created.Status.State).To(Equal(v1alpha1.Ready))
			}, eventuallyTimeout).Should(Succeed())

			ingressGatewaySvc := corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "istio-ingressgateway",
					Namespace: "istio-system",
				},
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{
						{
							Name: "http",
							Port: 80,
						},
					},
				},
				Status: corev1.ServiceStatus{
					LoadBalancer: corev1.LoadBalancerStatus{
						Ingress: []corev1.LoadBalancerIngress{
							{
								Hostname: "example1.com",
							},
						},
					},
				},
			}
			Expect(k8sClient.Create(context.Background(), &ingressGatewaySvc)).Should(Succeed())
			Eventually(func(g Gomega) {
				dnsEntry := dnsv1alpha1.DNSEntry{}
				Expect(dnsEntry.Spec.Targets).To(HaveLen(1))
				Expect(dnsEntry.Spec.Targets[0]).To(Equal("example1.com"))
			})

			updatedIngressGatewaySvc := corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "istio-ingressgateway",
					Namespace: "istio-system",
				},
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{
						{
							Name: "http",
							Port: 80,
						},
					},
				},
				Status: corev1.ServiceStatus{
					LoadBalancer: corev1.LoadBalancerStatus{
						Ingress: []corev1.LoadBalancerIngress{
							{
								Hostname: "example2.com",
							},
						},
					},
				},
			}
			Expect(k8sClient.Update(context.Background(), &updatedIngressGatewaySvc)).Should(Succeed())

			Eventually(func(g Gomega) {
				dnsEntry := dnsv1alpha1.DNSEntry{}
				Expect(dnsEntry.Spec.Targets).To(HaveLen(1))
				Expect(dnsEntry.Spec.Targets[0]).To(Equal("example2.com"))
			})

			Expect(k8sClient.Delete(context.Background(), &apiGateway)).Should(Succeed())
		})

		It("Should set ready state on first APIGateway CR and warning on second APIGateway CR when reconciliation succeeds", func() {
			// given
			apiGateway := v1alpha1.APIGateway{
				ObjectMeta: metav1.ObjectMeta{
					Name: generateName(),
				},
			}
			apiGateway2 := v1alpha1.APIGateway{
				ObjectMeta: metav1.ObjectMeta{
					Name: generateName(),
				},
			}

			// when
			Expect(k8sClient.Create(context.Background(), &apiGateway)).Should(Succeed())
			Eventually(func(g Gomega) {
				created := v1alpha1.APIGateway{}
				g.Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: apiGateway.Name}, &created)).Should(Succeed())
				g.Expect(created.ObjectMeta.Finalizers).To(HaveLen(1))
				g.Expect(created.ObjectMeta.Finalizers[0]).To(Equal(ApiGatewayFinalizer))
				g.Expect(created.Status.State).To(Equal(v1alpha1.Ready))
			}, eventuallyTimeout).Should(Succeed())

			Expect(k8sClient.Create(context.Background(), &apiGateway2)).Should(Succeed())

			// then
			Eventually(func(g Gomega) {
				created := v1alpha1.APIGateway{}
				g.Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: apiGateway2.Name}, &created)).Should(Succeed())
				g.Expect(created.ObjectMeta.Finalizers).To(HaveLen(0))
				g.Expect(created.Status.State).To(Equal(v1alpha1.Warning))
				g.Expect(created.Status.Description).To(Equal(fmt.Sprintf("stopped APIGateway CR reconciliation: only APIGateway CR %s reconciles the module", apiGateway.Name)))
			}, eventuallyTimeout).Should(Succeed())

			Expect(k8sClient.Delete(context.Background(), &apiGateway)).Should(Succeed())
			Expect(k8sClient.Delete(context.Background(), &apiGateway2)).Should(Succeed())
		})

		It("Should delete APIGateway CR when there are no blocking resources", func() {
			// given
			apiGateway := v1alpha1.APIGateway{
				ObjectMeta: metav1.ObjectMeta{
					Name: generateName(),
				},
			}

			// when
			Expect(k8sClient.Create(context.Background(), &apiGateway)).Should(Succeed())

			Eventually(func(g Gomega) {
				created := v1alpha1.APIGateway{}
				g.Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: apiGateway.Name}, &created)).Should(Succeed())
				g.Expect(created.ObjectMeta.Finalizers).To(HaveLen(1))
				g.Expect(created.ObjectMeta.Finalizers[0]).To(Equal(ApiGatewayFinalizer))
				g.Expect(created.Status.State).To(Equal(v1alpha1.Ready))
			}, eventuallyTimeout).Should(Succeed())

			Expect(k8sClient.Delete(context.Background(), &apiGateway)).Should(Succeed())

			// then
			Eventually(func(g Gomega) {
				deleted := v1alpha1.APIGateway{}
				defaultGateway := networkingv1alpha3.Gateway{}
				g.Expect(errors.IsNotFound(k8sClient.Get(context.Background(), client.ObjectKey{Name: apiGateway.Name}, &deleted))).Should(BeTrue())
				g.Expect(errors.IsNotFound(k8sClient.Get(context.Background(), client.ObjectKey{Name: kymaGatewayName, Namespace: kymaNamespace}, &defaultGateway))).Should(BeTrue())
			}, eventuallyTimeout).Should(Succeed())
		})

		It("Should set APIGateway CR in Warning state on deletion when APIRule exist", func() {
			// given
			apiGateway := v1alpha1.APIGateway{
				ObjectMeta: metav1.ObjectMeta{
					Name: generateName(),
				},
			}
			apiRule := getApiRule()
			By("Creating APIRule")
			Expect(k8sClient.Create(context.Background(), &apiRule)).Should(Succeed())
			By("Creating APIGateway")
			Expect(k8sClient.Create(context.Background(), &apiGateway)).Should(Succeed())

			By("Verifying that APIGateway CR reconciliation was successful")
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: apiGateway.Name}, &apiGateway)).Should(Succeed())
				g.Expect(apiGateway.Status.State).To(Equal(v1alpha1.Ready))
			}, eventuallyTimeout).Should(Succeed())

			// when
			By("Deleting APIGateway")
			Expect(k8sClient.Delete(context.Background(), &apiGateway)).Should(Succeed())

			// then
			By("Verifying that APIGateway CR has Warning state")
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: apiGateway.Name}, &apiGateway)).Should(Succeed())
				g.Expect(apiGateway.Status.State).To(Equal(v1alpha1.Warning))
				g.Expect(apiGateway.Status.Description).To(Equal("There are APIRule(s) that block the deletion of API-Gateway CR. Please take a look at kyma-system/api-gateway-controller-manager logs to see more information about the warning"))
			}, eventuallyTimeout).Should(Succeed())
		})
		It("Should set APIGateway CR in Warning state on deletion when RateLimit(s) exist", func() {
			// given
			apiGateway := v1alpha1.APIGateway{
				ObjectMeta: metav1.ObjectMeta{
					Name: generateName(),
				},
			}
			rateLimit := getRateLimit()

			By("Creating RateLimit")
			Expect(k8sClient.Create(context.Background(), &rateLimit)).Should(Succeed())
			By("Creating APIGateway")
			Expect(k8sClient.Create(context.Background(), &apiGateway)).Should(Succeed())

			By("Verifying that APIGateway CR reconciliation was successful")
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: apiGateway.Name}, &apiGateway)).Should(Succeed())
				g.Expect(apiGateway.Status.State).To(Equal(v1alpha1.Ready))
			}, eventuallyTimeout).Should(Succeed())

			// when
			By("Deleting APIGateway")
			Expect(k8sClient.Delete(context.Background(), &apiGateway)).Should(Succeed())

			// then
			By("Verifying that APIGateway CR has Warning state")
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: apiGateway.Name}, &apiGateway)).Should(Succeed())
				g.Expect(apiGateway.Status.State).To(Equal(v1alpha1.Warning))
				g.Expect(apiGateway.Status.Description).To(Equal("There are RateLimit(s) that block the deletion of API-Gateway CR. Please take a look at kyma-system/api-gateway-controller-manager logs to see more information about the warning"))
			}, eventuallyTimeout).Should(Succeed())
		})
		It("should update lastTransitionTime of Ready condition when only reason or message changed", func() {
			// given
			blockingApiRule := getApiRule()
			blockingApiRule.Name = "blocking-api-rule"
			blockingVs := getVirtualService()
			blockingVs.Name = "blocking-vs"

			apiGateway := v1alpha1.APIGateway{
				ObjectMeta: metav1.ObjectMeta{
					Name: generateName(),
				},
				Spec: v1alpha1.APIGatewaySpec{
					EnableKymaGateway: ptr.To(true),
				},
			}

			By("Creating VirtualService that references default gateway")
			Expect(k8sClient.Create(context.Background(), &blockingVs)).Should(Succeed())

			By("Creating APIRule that references default gateway")
			Expect(k8sClient.Create(context.Background(), &blockingApiRule)).Should(Succeed())

			By("Creating APIGateway")
			Expect(k8sClient.Create(context.Background(), &apiGateway)).Should(Succeed())

			By("Validating that APIGateway CR is in ready state")
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: apiGateway.Name}, &apiGateway)).Should(Succeed())
				g.Expect(apiGateway.Status.State).To(Equal(v1alpha1.Ready))
			}, eventuallyTimeout).Should(Succeed())

			By("Disabling default gateway in APIGateway")
			apiGateway.Spec.EnableKymaGateway = ptr.To(false)
			Expect(k8sClient.Update(context.Background(), &apiGateway)).Should(Succeed())

			// then
			By("Validating APIGateway is in warning state")
			var firstNotReadyTransitionTime metav1.Time
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: apiGateway.Name}, &apiGateway)).Should(Succeed())
				g.Expect(apiGateway.Status.State).To(Equal(v1alpha1.Warning))
				for _, condition := range apiGateway.Status.Conditions {
					if condition.Type == "Ready" {
						g.Expect(condition.Message).To(Equal("Kyma Gateway deletion blocked because of the existing custom resources: blocking-api-rule, blocking-vs"))
						firstNotReadyTransitionTime = condition.LastTransitionTime
					}
				}
				g.Expect(firstNotReadyTransitionTime).ToNot(BeZero())
			}, eventuallyTimeout).Should(Succeed())

			By("Deleting APIRule that referenced APIGateway")
			Expect(k8sClient.Delete(context.Background(), &blockingApiRule)).Should(Succeed())

			By("Verifying that the condition lastTransitionTime is also updated")
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: apiGateway.Name}, &apiGateway)).Should(Succeed())
				g.Expect(apiGateway.Status.State).To(Equal(v1alpha1.Warning))
				for _, condition := range apiGateway.Status.Conditions {
					if condition.Type == "Ready" {
						g.Expect(condition.Message).To(Equal("Kyma Gateway deletion blocked because of the existing custom resources: blocking-vs"))
						g.Expect(condition.LastTransitionTime.Compare(firstNotReadyTransitionTime.Time) >= 0).To(BeTrue())
					}
				}
				g.Expect(firstNotReadyTransitionTime).ToNot(BeZero())
			}, eventuallyTimeout).Should(Succeed())

		})
	})

	Context("Kyma Gateway reconciliation", func() {
		It("Should create Kyma Gateway when it's enabled", func() {
			// given
			apiGateway := v1alpha1.APIGateway{
				ObjectMeta: metav1.ObjectMeta{
					Name: generateName(),
				},
				Spec: v1alpha1.APIGatewaySpec{
					EnableKymaGateway: ptr.To(true),
				},
			}

			// when
			Expect(k8sClient.Create(context.Background(), &apiGateway)).Should(Succeed())

			// then
			Eventually(func(g Gomega) {
				created := v1alpha1.APIGateway{}
				g.Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: apiGateway.Name}, &created)).Should(Succeed())
				g.Expect(created.Status.State).To(Equal(v1alpha1.Ready))

				kymaGw := v1alpha3.Gateway{}
				g.Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: "kyma-gateway", Namespace: "kyma-system"}, &kymaGw)).Should(Succeed())
			}, eventuallyTimeout).Should(Succeed())
		})

		It("Should create no Kyma Gateway when it's not enabled", func() {
			// given
			apiGateway := v1alpha1.APIGateway{
				ObjectMeta: metav1.ObjectMeta{
					Name: generateName(),
				},
			}

			// when
			Expect(k8sClient.Create(context.Background(), &apiGateway)).Should(Succeed())

			// then
			Eventually(func(g Gomega) {
				created := v1alpha1.APIGateway{}
				g.Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: apiGateway.Name}, &created)).Should(Succeed())
				g.Expect(created.Status.State).To(Equal(v1alpha1.Ready))

				kymaGw := v1alpha3.Gateway{}
				err := k8sClient.Get(context.Background(), client.ObjectKey{Name: "kyma-gateway", Namespace: "kyma-system"}, &kymaGw)
				g.Expect(err).To(HaveOccurred())
				g.Expect(errors.IsNotFound(err)).To(BeTrue())
			}, eventuallyTimeout).Should(Succeed())
		})

		It("Should delete Kyma Gateway when it's disabled after it was enabled", func() {
			// given
			apiGateway := v1alpha1.APIGateway{
				ObjectMeta: metav1.ObjectMeta{
					Name: generateName(),
				},
				Spec: v1alpha1.APIGatewaySpec{
					EnableKymaGateway: ptr.To(true),
				},
			}

			By("Creating APIGateway with Kyma Gateway enabled")
			Expect(k8sClient.Create(context.Background(), &apiGateway)).Should(Succeed())

			By("Verifying that APIGateway CR reconciliation was successful and Kyma Gateway was created")
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: apiGateway.Name}, &apiGateway)).Should(Succeed())
				g.Expect(apiGateway.Status.State).To(Equal(v1alpha1.Ready))

				kymaGw := v1alpha3.Gateway{}
				g.Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: "kyma-gateway", Namespace: "kyma-system"}, &kymaGw)).Should(Succeed())
			}, eventuallyTimeout).Should(Succeed())

			By("Updating APIGateway CR with Kyma Gateway disabled")
			// when
			Eventually(func(g Gomega) {
				apiGateway = fetchLatestApiGateway(apiGateway)
				apiGateway.Spec.EnableKymaGateway = ptr.To(false)
				Expect(k8sClient.Update(context.Background(), &apiGateway)).Should(Succeed())
			}, eventuallyTimeout).Should(Succeed())

			// then
			By("Verifying that APIGateway CR reconciliation was successful and Kyma gateway was deleted")
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: apiGateway.Name}, &apiGateway)).Should(Succeed())
				g.Expect(apiGateway.Status.State).To(Equal(v1alpha1.Ready))

				kymaGw := v1alpha3.Gateway{}
				err := k8sClient.Get(context.Background(), client.ObjectKey{Name: "kyma-gateway", Namespace: "kyma-system"}, &kymaGw)
				g.Expect(err).To(HaveOccurred())
				g.Expect(errors.IsNotFound(err)).To(BeTrue())
			}, eventuallyTimeout).Should(Succeed())

		})

		It("Should block deletion of APIGateway CR with Kyma Gateway enabled when there is APIRule referring it", func() {
			// given
			apiRule := getApiRule()
			apiGateway := v1alpha1.APIGateway{
				ObjectMeta: metav1.ObjectMeta{
					Name: generateName(),
				},
				Spec: v1alpha1.APIGatewaySpec{
					EnableKymaGateway: ptr.To(true),
				},
			}

			// when
			By("Creating APIRule")
			Expect(k8sClient.Create(context.Background(), &apiRule)).Should(Succeed())
			By("Creating APIGateway")
			Expect(k8sClient.Create(context.Background(), &apiGateway)).Should(Succeed())

			By("Validating that APIGateway CR is in ready state")
			Eventually(func(g Gomega) {
				created := v1alpha1.APIGateway{}
				g.Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: apiGateway.Name}, &created)).Should(Succeed())
				g.Expect(created.ObjectMeta.Finalizers).To(HaveLen(2))
				g.Expect(created.ObjectMeta.Finalizers).To(ContainElement(ApiGatewayFinalizer))
				g.Expect(created.ObjectMeta.Finalizers).To(ContainElement(gateway.KymaGatewayFinalizer))
				g.Expect(created.Status.State).To(Equal(v1alpha1.Ready))
			}, eventuallyTimeout).Should(Succeed())

			By("Deleting APIGateway")
			Expect(k8sClient.Delete(context.Background(), &apiGateway)).Should(Succeed())

			// then
			By("Validating APIGateway is in warning state")
			Eventually(func(g Gomega) {
				deleted := v1alpha1.APIGateway{}
				g.Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: apiGateway.Name}, &deleted)).Should(Succeed())
				g.Expect(deleted.ObjectMeta.Finalizers).To(ContainElement(ApiGatewayFinalizer))
				g.Expect(deleted.ObjectMeta.Finalizers).To(ContainElement(gateway.KymaGatewayFinalizer))
				g.Expect(deleted.Status.State).To(Equal(v1alpha1.Warning))
				g.Expect(deleted.Status.Description).To(Equal("There are APIRule(s) that block the deletion of API-Gateway CR. Please take a look at kyma-system/api-gateway-controller-manager logs to see more information about the warning"))
			}, eventuallyTimeout).Should(Succeed())
		})

		It("Should block deletion of APIGateway CR with Kyma Gateway enabled when there is VirtualService referring it", func() {
			// given
			vs := getVirtualService()
			apiGateway := v1alpha1.APIGateway{
				ObjectMeta: metav1.ObjectMeta{
					Name: generateName(),
				},
				Spec: v1alpha1.APIGatewaySpec{
					EnableKymaGateway: ptr.To(true),
				},
			}

			// when
			By("Creating VirtualService")
			Expect(k8sClient.Create(context.Background(), &vs)).Should(Succeed())
			By("Creating APIGateway")
			Expect(k8sClient.Create(context.Background(), &apiGateway)).Should(Succeed())

			By("Validating that APIGateway CR is in ready state")
			Eventually(func(g Gomega) {
				created := v1alpha1.APIGateway{}
				g.Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: apiGateway.Name}, &created)).Should(Succeed())
				g.Expect(created.ObjectMeta.Finalizers).To(HaveLen(2))
				g.Expect(created.ObjectMeta.Finalizers).To(ContainElement(ApiGatewayFinalizer))
				g.Expect(created.ObjectMeta.Finalizers).To(ContainElement(gateway.KymaGatewayFinalizer))
				g.Expect(created.Status.State).To(Equal(v1alpha1.Ready))
			}, eventuallyTimeout).Should(Succeed())

			By("Deleting APIGateway CR")
			Expect(k8sClient.Delete(context.Background(), &apiGateway)).Should(Succeed())

			// then
			By("Validating that APIGateway CR is in warning state")
			Eventually(func(g Gomega) {
				deleted := v1alpha1.APIGateway{}
				g.Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: apiGateway.Name}, &deleted)).Should(Succeed())
				g.Expect(deleted.ObjectMeta.Finalizers).To(HaveLen(1))
				g.Expect(deleted.ObjectMeta.Finalizers).To(ContainElement(gateway.KymaGatewayFinalizer))
				g.Expect(deleted.Status.State).To(Equal(v1alpha1.Warning))
				g.Expect(deleted.Status.Description).To(Equal("There are custom resources that block the deletion of Kyma Gateway. Please take a look at kyma-system/api-gateway-controller-manager logs to see more information about the warning"))
			}, eventuallyTimeout).Should(Succeed())
		})
	})
})

func generateName() string {
	rand.NewSource(time.Now().UnixNano())

	letterRunes := []rune("abcdefghijklmnopqrstuvwxyz")

	b := make([]rune, 5)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return fmt.Sprintf("test-%s", string(b))
}

func deleteApiGateways() {
	Eventually(func(g Gomega) {
		By("Checking if APIGateway exists as part of teardown")
		list := v1alpha1.APIGatewayList{}
		Expect(k8sClient.List(context.Background(), &list)).Should(Succeed())

		for _, item := range list.Items {
			apiGatewayTeardown(&item)
		}
	}, eventuallyTimeout).Should(Succeed())
}

func apiGatewayTeardown(apiGateway *v1alpha1.APIGateway) {
	By(fmt.Sprintf("Deleting APIGateway %s as part of teardown", apiGateway.Name))
	Eventually(func(g Gomega) {
		err := k8sClient.Delete(context.Background(), apiGateway)

		if err != nil {
			Expect(errors.IsNotFound(err)).To(BeTrue())
		}

		a := v1alpha1.APIGateway{}
		err = k8sClient.Get(context.Background(), client.ObjectKey{Name: apiGateway.Name}, &a)
		g.Expect(errors.IsNotFound(err)).To(BeTrue())
	}, eventuallyTimeout).Should(Succeed())
}

func deleteApiRules() {
	Eventually(func(g Gomega) {
		By("Checking if APIRules exists as part of teardown")
		list := v1beta1.APIRuleList{}
		Expect(k8sClient.List(context.Background(), &list)).Should(Succeed())

		for _, item := range list.Items {
			apiRuleTeardown(&item)
		}
	}, eventuallyTimeout).Should(Succeed())
}

func fetchLatestApiGateway(apiGateway v1alpha1.APIGateway) v1alpha1.APIGateway {
	a := v1alpha1.APIGateway{}
	Eventually(func(g Gomega) {

		err := k8sClient.Get(context.Background(), client.ObjectKey{Name: apiGateway.Name}, &a)
		g.Expect(err).To(Not(HaveOccurred()))
	}, eventuallyTimeout).Should(Succeed())

	return a
}

func virtualServiceTeardown(vs *networkingv1beta1.VirtualService) {
	By(fmt.Sprintf("Deleting Virtual Service %s as part of teardown", vs.Name))
	Eventually(func(g Gomega) {
		err := k8sClient.Delete(context.Background(), vs)

		if err != nil {
			Expect(errors.IsNotFound(err)).To(BeTrue())
		}

		v := networkingv1beta1.VirtualService{}
		err = k8sClient.Get(context.Background(), client.ObjectKey{Name: vs.Name, Namespace: vs.Namespace}, &v)
		g.Expect(errors.IsNotFound(err)).To(BeTrue())
	}, eventuallyTimeout).Should(Succeed())
}

func apiRuleTeardown(apiRule *v1beta1.APIRule) {
	By(fmt.Sprintf("Deleting APIRule %s as part of teardown", apiRule.Name))
	err := k8sClient.Delete(context.Background(), apiRule)

	if err != nil {
		Expect(errors.IsNotFound(err)).To(BeTrue())
	}

	Eventually(func(g Gomega) {
		a := v1beta1.APIRule{}
		err := k8sClient.Get(context.Background(), client.ObjectKey{Name: apiRule.Name, Namespace: apiRule.Namespace}, &a)
		g.Expect(errors.IsNotFound(err)).To(BeTrue())
	}, eventuallyTimeout).Should(Succeed())
}

func getApiRule() gatewayv1beta1.APIRule {
	var servicePort uint32 = 8080

	return gatewayv1beta1.APIRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test",
			Namespace:  "default",
			Generation: 1,
		},
		Spec: gatewayv1beta1.APIRuleSpec{
			Host: ptr.To("test-host"),
			Service: &gatewayv1beta1.Service{
				Name: ptr.To("test-service"),
				Port: &servicePort,
			},
			Gateway: ptr.To(gateway.KymaGatewayFullName),
			Rules: []gatewayv1beta1.Rule{
				{
					Path:    "/.*",
					Methods: []gatewayv1beta1.HttpMethod{"GET"},
					AccessStrategies: []*gatewayv1beta1.Authenticator{
						{
							Handler: &gatewayv1beta1.Handler{
								Name: "noop",
							},
						},
					},
				},
			},
		},
	}
}

func getVirtualService() networkingv1beta1.VirtualService {
	var (
		host    = "foo.bar"
		gateway = gateway.KymaGatewayFullName
	)

	return networkingv1beta1.VirtualService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: apinetworkingv1beta1.VirtualService{
			Hosts:    []string{host},
			Gateways: []string{gateway},
		},
	}
}

func deleteVirtualServices() {
	Eventually(func(g Gomega) {
		By("Checking if VirtualServices exists as part of teardown")
		list := networkingv1beta1.VirtualServiceList{}
		Expect(k8sClient.List(context.Background(), &list)).Should(Succeed())

		for _, item := range list.Items {
			virtualServiceTeardown(item)
		}
	}, eventuallyTimeout).Should(Succeed())
}

func getRateLimit() ratelimitv1alpha1.RateLimit {
	return ratelimitv1alpha1.RateLimit{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: ratelimitv1alpha1.RateLimitSpec{
			SelectorLabels: map[string]string{"app": "test"},
			Local: ratelimitv1alpha1.LocalConfig{
				DefaultBucket: ratelimitv1alpha1.BucketSpec{
					MaxTokens:     1,
					TokensPerFill: 1,
					FillInterval: &metav1.Duration{
						Duration: 60 * time.Second,
					},
				},
			},
			EnableResponseHeaders: false,
			Enforce:               false,
		},
	}
}

func deleteRateLimitRules() {
	Eventually(func(g Gomega) {
		By("Checking if RateLimit exists as part of teardown")
		list := ratelimitv1alpha1.RateLimitList{}
		Expect(k8sClient.List(context.Background(), &list)).Should(Succeed())

		for _, rateLimit := range list.Items {
			rateLimitTeardown(&rateLimit)
		}
	}, eventuallyTimeout).Should(Succeed())
}

func rateLimitTeardown(rateLimit *ratelimitv1alpha1.RateLimit) {
	By(fmt.Sprintf("Deleting RateLimit %s as part of teardown", rateLimit.GetName()))
	Eventually(func(g Gomega) {
		err := k8sClient.Delete(context.Background(), rateLimit)

		if err != nil {
			Expect(errors.IsNotFound(err)).To(BeTrue())
		}

		r := ratelimitv1alpha1.RateLimit{}

		err = k8sClient.Get(context.Background(), client.ObjectKey{Name: rateLimit.Name, Namespace: rateLimit.Namespace}, &r)
		g.Expect(errors.IsNotFound(err)).To(BeTrue())

	}, eventuallyTimeout).Should(Succeed())
}
