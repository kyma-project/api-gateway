package reconciliations_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/api-gateway/internal/reconciliations"
)

var _ = Describe("Gardener", func() {
	Context("GetGardenerDomain", func() {
		It("should return the domain name from the Gardener shoot-info config", func() {
			// given
			cm := corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "shoot-info",
					Namespace: "kube-system",
				},
				Data: map[string]string{
					"domain": "some.gardener.domain",
				},
			}

			k8sClient := createFakeClient(&cm)

			// when
			domain, err := reconciliations.GetGardenerDomain(context.Background(), k8sClient)

			// then
			Expect(err).ShouldNot(HaveOccurred())
			Expect(domain).To(Equal("some.gardener.domain"))
		})

		It("should return an error if no Gardener shoot-info is available", func() {
			// given
			k8sClient := createFakeClient()

			// when
			_, err := reconciliations.GetGardenerDomain(context.Background(), k8sClient)

			// then
			Expect(err).Should(HaveOccurred())
		})

		It("should return an empty string and no error if shoot-info does not have a domain", func() {
			// given
			cm := corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "shoot-info",
					Namespace: "kube-system",
				},
			}

			k8sClient := createFakeClient(&cm)

			// when
			_, err := reconciliations.GetGardenerDomain(context.Background(), k8sClient)

			// then
			Expect(err).Should(Succeed())
			Expect(cm.Data["domain"]).Should(BeEmpty())
		})
	})
})
