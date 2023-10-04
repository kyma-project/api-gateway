package operator

import (
	"context"
	"fmt"
	"github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	. "github.com/onsi/gomega"
	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/utils/ptr"
	"math/rand"
	"time"

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
				g.Expect(created.ObjectMeta.Finalizers[0]).To(Equal("apigateways.operator.kyma-project.io/api-gateway-reconciliation"))
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
				g.Expect(created.ObjectMeta.Finalizers[0]).To(Equal("apigateways.operator.kyma-project.io/api-gateway-reconciliation"))
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

		It("Should block deletion of APIGateway CR when there is APIRule blocking", func() {
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

			By("Validating APIGateway is ready")
			Eventually(func(g Gomega) {
				created := v1alpha1.APIGateway{}
				g.Expect(k8sClient.Get(ctx, client.ObjectKey{Name: apiGateway.Name}, &created)).Should(Succeed())
				g.Expect(created.ObjectMeta.Finalizers).To(ContainElement("apigateways.operator.kyma-project.io/api-gateway-reconciliation"))
				g.Expect(created.Status.State).To(Equal(v1alpha1.Ready))
			}, eventuallyTimeout).Should(Succeed())

			By("Deleting APIGateway")
			Expect(k8sClient.Delete(ctx, &apiGateway)).Should(Succeed())

			// then
			By("Validating APIGateway is in warning state")
			Eventually(func(g Gomega) {
				deleted := v1alpha1.APIGateway{}
				g.Expect(k8sClient.Get(ctx, client.ObjectKey{Name: apiGateway.Name}, &deleted)).Should(Succeed())
				g.Expect(deleted.ObjectMeta.Finalizers).To(ContainElement("apigateways.operator.kyma-project.io/api-gateway-reconciliation"))
				g.Expect(deleted.Status.State).To(Equal(v1alpha1.Warning))
				g.Expect(deleted.Status.Description).To(Equal("There are custom resource(s) that block the deletion. Please take a look at kyma-system/api-gateway-controller-manager logs to see more information about the warning"))
			}, eventuallyTimeout).Should(Succeed())
		})

		It("Should block deletion of APIGateway CR when there is VirtualService blocking", func() {
			// TODO: Check why logs report that two Virtual Services are blocking deletion
			// given
			vs := getVirtualService()
			apiGateway := v1alpha1.APIGateway{
				ObjectMeta: metav1.ObjectMeta{
					Name: generateName(),
				},
			}

			// when
			Expect(k8sClient.Create(ctx, &vs)).Should(Succeed())
			Expect(k8sClient.Create(ctx, &apiGateway)).Should(Succeed())
			defer func() {
				virtualServiceTeardown(&vs)
				apiGatewayTeardown(&apiGateway)
			}()

			By("Validating that APIGateway CR is in ready state")
			Eventually(func(g Gomega) {
				created := v1alpha1.APIGateway{}
				g.Expect(k8sClient.Get(ctx, client.ObjectKey{Name: apiGateway.Name}, &created)).Should(Succeed())
				g.Expect(created.ObjectMeta.Finalizers).To(HaveLen(1))
				g.Expect(created.ObjectMeta.Finalizers[0]).To(Equal("apigateways.operator.kyma-project.io/api-gateway-reconciliation"))
				g.Expect(created.Status.State).To(Equal(v1alpha1.Ready))
			}, eventuallyTimeout).Should(Succeed())

			By("Deleting APIGateway CR")
			Expect(k8sClient.Delete(ctx, &apiGateway)).Should(Succeed())

			// then
			By("Validating that APIGateway CR is in warning state")
			Eventually(func(g Gomega) {
				deleted := v1alpha1.APIGateway{}
				g.Expect(k8sClient.Get(ctx, client.ObjectKey{Name: apiGateway.Name}, &deleted)).Should(Succeed())
				g.Expect(deleted.ObjectMeta.Finalizers).To(HaveLen(1))
				g.Expect(deleted.ObjectMeta.Finalizers[0]).To(Equal("apigateways.operator.kyma-project.io/api-gateway-reconciliation"))
				g.Expect(deleted.Status.State).To(Equal(v1alpha1.Warning))
				g.Expect(deleted.Status.Description).To(Equal("There are custom resource(s) that block the deletion. Please take a look at kyma-system/api-gateway-controller-manager logs to see more information about the warning"))
			}, eventuallyTimeout).Should(Succeed())
		})

	})

	Context("Kyma Gateway reconciliation", func() {

		It("should create Kyma Gateway when it's enabled", func() {
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

		It("should create no Kyma Gateway when it's not enabled", func() {
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

		It("should delete Kyma Gateway when it's disabled after it was enabled", func() {
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

			By("Verifying that APIGateway CR reconciliation was successful and Kyma gateway was created")
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

		It("should set APIGateway in warning state when Kyma Gateway is disabled and APIRules exist", func() {
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

			By("Creating APIRule")
			apiRule := testApiRule()
			Expect(k8sClient.Create(ctx, &apiRule)).Should(Succeed())
			defer func() {
				apiRuleTeardown(&apiRule)
			}()

			By("Verifying that APIGateway CR reconciliation was successful and Kyma gateway was created")
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, client.ObjectKey{Name: apiGateway.Name}, &apiGateway)).Should(Succeed())
				g.Expect(apiGateway.Status.State).To(Equal(v1alpha1.Ready))

				kymaGw := v1alpha3.Gateway{}
				g.Expect(k8sClient.Get(ctx, client.ObjectKey{Name: "kyma-gateway", Namespace: "kyma-system"}, &kymaGw)).Should(Succeed())
			}, eventuallyTimeout).Should(Succeed())

			// when
			By("Disabling Kyma Gateway")
			Eventually(func(g Gomega) {
				apiGateway = fetchLatestApiGateway(apiGateway)
				apiGateway.Spec.EnableKymaGateway = ptr.To(false)

				g.Expect(k8sClient.Update(ctx, &apiGateway)).Should(Succeed())
			}, eventuallyTimeout).Should(Succeed())

			// then
			By("Verifying that APIGateway CR has Warning state and Kyma gateway was not deleted")
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, client.ObjectKey{Name: apiGateway.Name}, &apiGateway)).Should(Succeed())
				g.Expect(apiGateway.Status.State).To(Equal(v1alpha1.Warning))
				g.Expect(apiGateway.Status.Description).To(Equal("Kyma Gateway cannot be disabled because APIRules exist."))

				kymaGw := v1alpha3.Gateway{}
				err := k8sClient.Get(ctx, client.ObjectKey{Name: "kyma-gateway", Namespace: "kyma-system"}, &kymaGw)
				g.Expect(err).To(Not(HaveOccurred()))
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

func testApiRule() v1beta1.APIRule {
	rule := v1beta1.Rule{
		Path: "/.*",
		Methods: []string{
			"GET",
		},
		AccessStrategies: []*v1beta1.Authenticator{
			{
				Handler: &v1beta1.Handler{
					Name: "allow",
				},
			},
		},
	}

	var port uint32 = 8080
	return v1beta1.APIRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      generateName(),
			Namespace: testNamespace,
		},
		Spec: v1beta1.APIRuleSpec{
			Host:    ptr.To("test-host"),
			Gateway: ptr.To("kyma-system/kyma-gateway"),
			Service: &v1beta1.Service{
				Name:      ptr.To("test-service"),
				Namespace: ptr.To(testNamespace),
				Port:      ptr.To(port),
			},
			Rules: []v1beta1.Rule{rule},
		},
	}
}

func getApiRule() gatewayv1beta1.APIRule {
	var (
		serviceName        = "test"
		servicePort uint32 = 8000
		host               = "foo.bar"
		isExternal         = false
		gateway            = "kyma-system/kyma-gateway"
	)

	return gatewayv1beta1.APIRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test",
			Namespace:  "default",
			Generation: 1,
		},
		Spec: gatewayv1beta1.APIRuleSpec{
			Host: &host,
			Service: &gatewayv1beta1.Service{
				Name:       &serviceName,
				Port:       &servicePort,
				IsExternal: &isExternal,
			},
			Gateway: &gateway,
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
		gateway = "kyma-system/kyma-gateway"
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
