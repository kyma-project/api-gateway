package operator

import (
	"context"
	"fmt"
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
					Name: generateApiGatewayName(),
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
					Name: generateApiGatewayName(),
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
					Name: generateApiGatewayName(),
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
	})
})

func generateApiGatewayName() string {

	rand.NewSource(time.Now().UnixNano())

	letterRunes := []rune("abcdefghijklmnopqrstuvwxyz")

	b := make([]rune, 5)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return fmt.Sprintf("int-test-%s", string(b))
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
