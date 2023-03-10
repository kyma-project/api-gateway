package hashbasedstate_test

import (
	"github.com/kyma-project/api-gateway/internal/processing/hashbasedstate"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Desired state", func() {
	Context("Add", func() {
		It("should return an error when index label does not exist", func() {
			// given
			ap := securityv1beta1.AuthorizationPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"gateway.kyma-project.io/hash": "56fggdf4",
					},
				},
			}

			hashbasedAp := hashbasedstate.NewAuthorizationPolicy(&ap)
			sut := hashbasedstate.NewDesired()

			// when
			err := sut.Add(&hashbasedAp)

			// then
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).To(Equal("label gateway.kyma-project.io/index not found on hashable"))
		})

		It("should return an error when hash label does not exist", func() {
			// given
			ap := securityv1beta1.AuthorizationPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"gateway.kyma-project.io/index": "0",
					},
				},
			}
			hashbasedAp := hashbasedstate.NewAuthorizationPolicy(&ap)
			sut := hashbasedstate.NewDesired()

			// when
			err := sut.Add(&hashbasedAp)

			// then
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).To(Equal("label gateway.kyma-project.io/hash not found on hashable"))
		})
	})
})
