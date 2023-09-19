package gateway

import (
	"context"
	"github.com/gardener/cert-management/pkg/apis/cert/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Resource", func() {

	Context("applyResource", func() {

		It("should reapply disclaimer annotation on resource when it was removed", func() {
			// given
			k8sClient := createFakeClient()

			templateValues := make(map[string]string)
			templateValues["Name"] = "test"
			templateValues["Namespace"] = "istio-system"
			templateValues["Domain"] = "test-domain.com"
			templateValues["SecretName"] = "cert-secret"

			Expect(applyResource(context.TODO(), k8sClient, certificateManifest, templateValues)).Should(Succeed())

			By("removing disclaimer annotation from certificate")
			cert := v1alpha1.Certificate{}
			Expect(k8sClient.Get(context.TODO(), client.ObjectKey{Name: "test", Namespace: "istio-system"}, &cert)).Should(Succeed())
			cert.Annotations = nil
			Expect(k8sClient.Update(context.TODO(), &cert)).Should(Succeed())

			// when
			Expect(applyResource(context.TODO(), k8sClient, certificateManifest, templateValues)).Should(Succeed())

			// then
			Expect(k8sClient.Get(context.TODO(), client.ObjectKey{Name: "test", Namespace: "istio-system"}, &cert)).Should(Succeed())

			Expect(cert.Annotations).To(HaveKeyWithValue("apigateways.operator.kyma-project.io/managed-by-disclaimer",
				"DO NOT EDIT - This resource is managed by Kyma.\nAny modifications are discarded and the resource is reverted to the original state."))
		})
	})
})
