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

	"github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Tests needs to be executed serially because of the shared cluster-wide resources like the APIGateway CR.
var _ = Describe("API Gateway Controller", Serial, func() {

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
				kymaGatewayTeardown()
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
				kymaGatewayTeardown()
			}()

			By("Verifying that APIGateway CR reconciliation was successful and Kyma gateway was created")
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, client.ObjectKey{Name: apiGateway.Name}, &apiGateway)).Should(Succeed())
				g.Expect(apiGateway.Status.State).To(Equal(v1alpha1.Ready))

				kymaGw := v1alpha3.Gateway{}
				g.Expect(k8sClient.Get(ctx, client.ObjectKey{Name: "kyma-gateway", Namespace: "kyma-system"}, &kymaGw)).Should(Succeed())
			}, eventuallyTimeout).Should(Succeed())

			By("Updating APIGateway CR with Kyma Gateway disabled")
			apiGateway.Spec.EnableKymaGateway = ptr.To(false)

			// when
			Expect(k8sClient.Update(ctx, &apiGateway)).Should(Succeed())

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

		It("should set APIGateway in error state when Kyma Gateway is disabled and APIRules exist", func() {
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
				kymaGatewayTeardown()
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

			By("Disabling Kyma Gateway")
			apiGateway.Spec.EnableKymaGateway = ptr.To(false)

			// when
			Expect(k8sClient.Update(ctx, &apiGateway)).Should(Succeed())

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

// TODO: Remove this since this is a workaround until it's clarified how we handle ownership issue of Kyma Gateway
func kymaGatewayTeardown() {
	By("Deleting Gateway kyma-gateway as part of teardown")
	Eventually(func(g Gomega) {
		gw := v1alpha3.Gateway{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "kyma-gateway",
				Namespace: "kyma-system",
			},
		}
		err := k8sClient.Delete(context.TODO(), &gw)

		if err != nil {
			Expect(errors.IsNotFound(err)).To(BeTrue())
		}

		a := v1alpha3.Gateway{}
		err = k8sClient.Get(context.TODO(), client.ObjectKey{Name: gw.Name, Namespace: gw.Namespace}, &a)
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
