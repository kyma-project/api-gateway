package gateway

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kyma-project/api-gateway/internal/reconciliations"
)

type unstructuredManifest struct {
	GVK      schema.GroupVersionKind
	Manifest []byte
}

var _ = Describe("Resource", func() {
	Context("ApplyResource", func() {
		It("should reapply disclaimer annotation and module labels on resource when they were removed", func() {
			// given
			k8sClient := createFakeClient()

			templateValues := make(map[string]string)
			templateValues["Name"] = "test"
			templateValues["Namespace"] = "istio-system"
			templateValues["Domain"] = "test-domain.com"
			templateValues["SecretName"] = "cert-secret"
			templateValues["Version"] = "1.0.0"
			templateValues["IngressGatewayServiceIp"] = "1.1.1.1"
			templateValues["CertificateSecretName"] = "test"
			templateValues["Gateway"] = "test-gateway"

			resources := []unstructuredManifest{
				{schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Secret"}, nonGardenerCertificateSecretManifest},
				{schema.GroupVersionKind{Group: "cert.gardener.cloud", Version: "v1alpha1", Kind: "Certificate"}, certificateManifest},
				{schema.GroupVersionKind{Group: "dns.gardener.cloud", Version: "v1alpha1", Kind: "DNSEntry"}, dnsEntryManifest},
				{schema.GroupVersionKind{Group: "networking.istio.io", Version: "v1alpha3", Kind: "Gateway"}, kymaGatewayManifest},
				{schema.GroupVersionKind{Group: "networking.istio.io", Version: "v1beta1", Kind: "VirtualService"}, virtualServiceManifest},
			}

			for _, resource := range resources {
				Expect(reconciliations.ApplyResource(context.Background(), k8sClient, resource.Manifest, templateValues)).Should(Succeed())

				By("Removing disclaimer annotation from resource")
				obj := unstructured.Unstructured{}
				obj.SetGroupVersionKind(resource.GVK)

				Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: "test", Namespace: "istio-system"}, &obj)).Should(Succeed())
				obj.SetAnnotations(nil)
				obj.SetLabels(nil)
				Expect(k8sClient.Update(context.Background(), &obj)).Should(Succeed())

				// when
				Expect(reconciliations.ApplyResource(context.Background(), k8sClient, resource.Manifest, templateValues)).Should(Succeed())

				// then
				Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: "test", Namespace: "istio-system"}, &obj)).Should(Succeed())
				Expect(obj.GetAnnotations()).To(HaveKeyWithValue("apigateways.operator.kyma-project.io/managed-by-disclaimer",
					"DO NOT EDIT - This resource is managed by Kyma.\nAny modifications are discarded and the resource is reverted to the original state."))

				Expect(obj.GetLabels()).To(HaveKeyWithValue("kyma-project.io/module", "api-gateway"))

				Expect(obj.GetLabels()).To(HaveKeyWithValue("app.kubernetes.io/name", "api-gateway-operator"))
				Expect(obj.GetLabels()).To(HaveKeyWithValue("app.kubernetes.io/instance", "api-gateway-operator-default"))
				Expect(obj.GetLabels()).To(HaveKeyWithValue("app.kubernetes.io/version", "1.0.0"))
				Expect(obj.GetLabels()).To(HaveKeyWithValue("app.kubernetes.io/component", "operator"))
				Expect(obj.GetLabels()).To(HaveKeyWithValue("app.kubernetes.io/part-of", "api-gateway"))
			}
		})
	})
})
