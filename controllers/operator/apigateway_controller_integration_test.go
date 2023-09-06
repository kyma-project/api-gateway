package operator

import (
	"fmt"
	. "github.com/onsi/gomega"
	"math/rand"
	"time"

	"github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("API Gateway Controller", func() {

	It("should reconcile APIGateway CR sucessfully", func() {
		// given
		apiGateway := v1alpha1.APIGateway{
			ObjectMeta: metav1.ObjectMeta{
				Name: generateApiGatewayName(),
			},
		}

		// when
		Expect(k8sClient.Create(ctx, &apiGateway)).Should(Succeed())

		// then
		Eventually(func(g Gomega) {
			created := v1alpha1.APIGateway{}
			g.Expect(k8sClient.Get(ctx, client.ObjectKey{Name: apiGateway.Name}, &created)).Should(Succeed())
			g.Expect(created.Status.State).To(Equal(v1alpha1.Ready))
		}, eventuallyTimeout).Should(Succeed())
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
