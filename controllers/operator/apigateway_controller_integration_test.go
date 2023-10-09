package operator

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"github.com/kyma-project/api-gateway/internal/gateway"
	"github.com/kyma-project/api-gateway/internal/operator/reconciliations/api_gateway"
	. "github.com/onsi/gomega"
	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/utils/ptr"

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
	Context("APIGateway CR", func() {
		It("Should set ready state on APIGateway CR when reconciliation succeeds", func() {
			// given
			apiGateway := v1alpha1.APIGateway{
				ObjectMeta: metav1.ObjectMeta{
					Name: generateName(),
				},
			}

			// when
			Expect(k8sClient.Create(ctx, &apiGateway)).Should(Succeed())
			defer func() {
				apiGatewayTeardown(&apiGateway)
			}()

			// then
			Eventually(func(g Gomega) {
				created := v1alpha1.APIGateway{}
				g.Expect(k8sClient.Get(ctx, client.ObjectKey{Name: apiGateway.Name}, &created)).Should(Succeed())
				g.Expect(created.ObjectMeta.Finalizers).To(HaveLen(1))
				g.Expect(created.ObjectMeta.Finalizers[0]).To(Equal(api_gateway.ApiGatewayFinalizer))
				g.Expect(created.Status.State).To(Equal(v1alpha1.Ready))
			}, eventuallyTimeout).Should(Succeed())

			Expect(k8sClient.Delete(ctx, &apiGateway)).Should(Succeed())
		})

		It("Should delete APIGateway CR when there are no blocking resources", func() {
			// given
			apiGateway := v1alpha1.APIGateway{
				ObjectMeta: metav1.ObjectMeta{
					Name: generateName(),
				},
			}

			// when
			Expect(k8sClient.Create(ctx, &apiGateway)).Should(Succeed())
			defer func() {
				apiGatewayTeardown(&apiGateway)
			}()

			Eventually(func(g Gomega) {
				created := v1alpha1.APIGateway{}
				g.Expect(k8sClient.Get(ctx, client.ObjectKey{Name: apiGateway.Name}, &created)).Should(Succeed())
				g.Expect(created.ObjectMeta.Finalizers).To(HaveLen(1))
				g.Expect(created.ObjectMeta.Finalizers[0]).To(Equal(api_gateway.ApiGatewayFinalizer))
				g.Expect(created.Status.State).To(Equal(v1alpha1.Ready))
			}, eventuallyTimeout).Should(Succeed())

			Expect(k8sClient.Delete(ctx, &apiGateway)).Should(Succeed())

			// then
			Eventually(func(g Gomega) {
				deleted := v1alpha1.APIGateway{}
				defaultGateway := networkingv1alpha3.Gateway{}
				g.Expect(errors.IsNotFound(k8sClient.Get(ctx, client.ObjectKey{Name: apiGateway.Name}, &deleted))).Should(BeTrue())
				g.Expect(errors.IsNotFound(k8sClient.Get(ctx, client.ObjectKey{Name: kymaGatewayName, Namespace: kymaNamespace}, &defaultGateway))).Should(BeTrue())
			}, eventuallyTimeout).Should(Succeed())
		})

		It("Should set APIGateway in warning state on deletion when APIRule exist", func() {
			// given
			apiGateway := v1alpha1.APIGateway{
				ObjectMeta: metav1.ObjectMeta{
					Name: generateName(),
				},
			}
			apiRule := getApiRule()
			By("Creating APIRule")
			Expect(k8sClient.Create(ctx, &apiRule)).Should(Succeed())
			By("Creating APIGateway")
			Expect(k8sClient.Create(ctx, &apiGateway)).Should(Succeed())
			defer func() {
				apiRuleTeardown(&apiRule)
				apiGatewayTeardown(&apiGateway)
			}()

			By("Verifying that APIGateway CR reconciliation was successful")
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, client.ObjectKey{Name: apiGateway.Name}, &apiGateway)).Should(Succeed())
				g.Expect(apiGateway.Status.State).To(Equal(v1alpha1.Ready))
			}, eventuallyTimeout).Should(Succeed())

			// when
			By("Deleting APIGateway")
			Expect(k8sClient.Delete(ctx, &apiGateway)).Should(Succeed())

			// then
			By("Verifying that APIGateway CR has Warning state")
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, client.ObjectKey{Name: apiGateway.Name}, &apiGateway)).Should(Succeed())
				g.Expect(apiGateway.Status.State).To(Equal(v1alpha1.Warning))
				g.Expect(apiGateway.Status.Description).To(Equal("There are APIRule(s) that block the deletion. Please take a look at kyma-system/api-gateway-controller-manager logs to see more information about the warning"))
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
			Expect(k8sClient.Create(ctx, &apiGateway)).Should(Succeed())
			defer func() {
				apiGatewayTeardown(&apiGateway)
			}()
			// then
			Eventually(func(g Gomega) {
				created := v1alpha1.APIGateway{}
				g.Expect(k8sClient.Get(ctx, client.ObjectKey{Name: apiGateway.Name}, &created)).Should(Succeed())
				g.Expect(created.Status.State).To(Equal(v1alpha1.Ready))

				kymaGw := v1alpha3.Gateway{}
				g.Expect(k8sClient.Get(ctx, client.ObjectKey{Name: "kyma-gateway", Namespace: "kyma-system"}, &kymaGw)).Should(Succeed())
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
			Expect(k8sClient.Create(ctx, &apiGateway)).Should(Succeed())
			defer func() {
				apiGatewayTeardown(&apiGateway)
			}()

			// then
			Eventually(func(g Gomega) {
				created := v1alpha1.APIGateway{}
				g.Expect(k8sClient.Get(ctx, client.ObjectKey{Name: apiGateway.Name}, &created)).Should(Succeed())
				g.Expect(created.Status.State).To(Equal(v1alpha1.Ready))

				kymaGw := v1alpha3.Gateway{}
				err := k8sClient.Get(ctx, client.ObjectKey{Name: "kyma-gateway", Namespace: "kyma-system"}, &kymaGw)
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
			Expect(k8sClient.Create(ctx, &apiGateway)).Should(Succeed())
			defer func() {
				apiGatewayTeardown(&apiGateway)
			}()

			By("Verifying that APIGateway CR reconciliation was successful and Kyma Gateway was created")
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, client.ObjectKey{Name: apiGateway.Name}, &apiGateway)).Should(Succeed())
				g.Expect(apiGateway.Status.State).To(Equal(v1alpha1.Ready))

				kymaGw := v1alpha3.Gateway{}
				g.Expect(k8sClient.Get(ctx, client.ObjectKey{Name: "kyma-gateway", Namespace: "kyma-system"}, &kymaGw)).Should(Succeed())
			}, eventuallyTimeout).Should(Succeed())

			By("Updating APIGateway CR with Kyma Gateway disabled")
			// when
			Eventually(func(g Gomega) {
				apiGateway = fetchLatestApiGateway(apiGateway)
				apiGateway.Spec.EnableKymaGateway = ptr.To(false)
				Expect(k8sClient.Update(ctx, &apiGateway)).Should(Succeed())
			}, eventuallyTimeout).Should(Succeed())

			// then
			By("Verifying that APIGateway CR reconciliation was successful and Kyma gateway was deleted")
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, client.ObjectKey{Name: apiGateway.Name}, &apiGateway)).Should(Succeed())
				g.Expect(apiGateway.Status.State).To(Equal(v1alpha1.Ready))

				kymaGw := v1alpha3.Gateway{}
				err := k8sClient.Get(ctx, client.ObjectKey{Name: "kyma-gateway", Namespace: "kyma-system"}, &kymaGw)
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
			Expect(k8sClient.Create(ctx, &apiRule)).Should(Succeed())
			By("Creating APIGateway")
			Expect(k8sClient.Create(ctx, &apiGateway)).Should(Succeed())
			defer func() {
				apiRuleTeardown(&apiRule)
				apiGatewayTeardown(&apiGateway)
			}()

			By("Validating that APIGateway CR is in ready state")
			Eventually(func(g Gomega) {
				created := v1alpha1.APIGateway{}
				g.Expect(k8sClient.Get(ctx, client.ObjectKey{Name: apiGateway.Name}, &created)).Should(Succeed())
				g.Expect(created.ObjectMeta.Finalizers).To(HaveLen(2))
				g.Expect(created.ObjectMeta.Finalizers).To(ContainElement(api_gateway.ApiGatewayFinalizer))
				g.Expect(created.ObjectMeta.Finalizers).To(ContainElement(gateway.KymaGatewayFinalizer))
				g.Expect(created.Status.State).To(Equal(v1alpha1.Ready))
			}, eventuallyTimeout).Should(Succeed())

			By("Deleting APIGateway")
			Expect(k8sClient.Delete(ctx, &apiGateway)).Should(Succeed())

			// then
			By("Validating APIGateway is in warning state")
			Eventually(func(g Gomega) {
				deleted := v1alpha1.APIGateway{}
				g.Expect(k8sClient.Get(ctx, client.ObjectKey{Name: apiGateway.Name}, &deleted)).Should(Succeed())
				g.Expect(deleted.ObjectMeta.Finalizers).To(ContainElement(api_gateway.ApiGatewayFinalizer))
				g.Expect(deleted.ObjectMeta.Finalizers).To(ContainElement(gateway.KymaGatewayFinalizer))
				g.Expect(deleted.Status.State).To(Equal(v1alpha1.Warning))
				g.Expect(deleted.Status.Description).To(Equal("There are custom resources that block the deletion. Please take a look at kyma-system/api-gateway-controller-manager logs to see more information about the warning"))
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
			Expect(k8sClient.Create(ctx, &vs)).Should(Succeed())
			By("Creating APIGateway")
			Expect(k8sClient.Create(ctx, &apiGateway)).Should(Succeed())
			defer func() {
				virtualServiceTeardown(&vs)
				apiGatewayTeardown(&apiGateway)
			}()

			By("Validating that APIGateway CR is in ready state")
			Eventually(func(g Gomega) {
				created := v1alpha1.APIGateway{}
				g.Expect(k8sClient.Get(ctx, client.ObjectKey{Name: apiGateway.Name}, &created)).Should(Succeed())
				g.Expect(created.ObjectMeta.Finalizers).To(HaveLen(2))
				g.Expect(created.ObjectMeta.Finalizers).To(ContainElement(api_gateway.ApiGatewayFinalizer))
				g.Expect(created.ObjectMeta.Finalizers).To(ContainElement(gateway.KymaGatewayFinalizer))
				g.Expect(created.Status.State).To(Equal(v1alpha1.Ready))
			}, eventuallyTimeout).Should(Succeed())

			By("Deleting APIGateway CR")
			Expect(k8sClient.Delete(ctx, &apiGateway)).Should(Succeed())

			// then
			By("Validating that APIGateway CR is in warning state")
			Eventually(func(g Gomega) {
				deleted := v1alpha1.APIGateway{}
				g.Expect(k8sClient.Get(ctx, client.ObjectKey{Name: apiGateway.Name}, &deleted)).Should(Succeed())
				g.Expect(deleted.ObjectMeta.Finalizers).To(HaveLen(2))
				g.Expect(deleted.ObjectMeta.Finalizers).To(ContainElement(api_gateway.ApiGatewayFinalizer))
				g.Expect(deleted.ObjectMeta.Finalizers).To(ContainElement(gateway.KymaGatewayFinalizer))
				g.Expect(deleted.Status.State).To(Equal(v1alpha1.Warning))
				g.Expect(deleted.Status.Description).To(Equal("There are custom resources that block the deletion. Please take a look at kyma-system/api-gateway-controller-manager logs to see more information about the warning"))
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

func apiGatewayTeardown(apiGateway *v1alpha1.APIGateway) {
	By(fmt.Sprintf("Deleting APIGateway %s as part of teardown", apiGateway.Name))
	Eventually(func(g Gomega) {
		err := k8sClient.Delete(context.TODO(), apiGateway)

		if err != nil {
			Expect(errors.IsNotFound(err)).To(BeTrue())
		}

		a := v1alpha1.APIGateway{}
		err = k8sClient.Get(context.TODO(), client.ObjectKey{Name: apiGateway.Name}, &a)
		g.Expect(errors.IsNotFound(err)).To(BeTrue())
	}, eventuallyTimeout).Should(Succeed())
}

func fetchLatestApiGateway(apiGateway v1alpha1.APIGateway) v1alpha1.APIGateway {
	a := v1alpha1.APIGateway{}
	Eventually(func(g Gomega) {

		err := k8sClient.Get(context.TODO(), client.ObjectKey{Name: apiGateway.Name}, &a)
		g.Expect(err).To(Not(HaveOccurred()))
	}, eventuallyTimeout).Should(Succeed())

	return a
}

func virtualServiceTeardown(vs *networkingv1beta1.VirtualService) {
	By(fmt.Sprintf("Deleting Virtual Service %s as part of teardown", vs.Name))
	Eventually(func(g Gomega) {
		err := k8sClient.Delete(context.TODO(), vs)

		if err != nil {
			Expect(errors.IsNotFound(err)).To(BeTrue())
		}

		v := networkingv1beta1.VirtualService{}
		err = k8sClient.Get(context.TODO(), client.ObjectKey{Name: vs.Name, Namespace: vs.Namespace}, &v)
		g.Expect(errors.IsNotFound(err)).To(BeTrue())
	}, eventuallyTimeout).Should(Succeed())
}

func apiRuleTeardown(apiRule *v1beta1.APIRule) {
	By(fmt.Sprintf("Deleting ApiRule %s as part of teardown", apiRule.Name))
	err := k8sClient.Delete(context.TODO(), apiRule)

	if err != nil {
		Expect(errors.IsNotFound(err)).To(BeTrue())
	}

	Eventually(func(g Gomega) {
		a := v1beta1.APIRule{}
		err := k8sClient.Get(context.TODO(), client.ObjectKey{Name: apiRule.Name, Namespace: apiRule.Namespace}, &a)
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
					Methods: []string{"GET"},
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
