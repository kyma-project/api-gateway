package reconciliations_test

import (
	"context"

	"github.com/kyma-project/api-gateway/internal/reconciliations"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

		It("should return an error if the Gardener shoot-info does not have a domain", func() {
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
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("domain not found in Gardener shoot-info"))
		})
	})
})
