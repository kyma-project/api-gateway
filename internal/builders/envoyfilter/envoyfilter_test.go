package envoyfilter

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	networkingv1alpha3 "istio.io/api/networking/v1alpha3"
)

var _ = Describe("EnvoyFilterBuilder", func() {
	Context("Build", func() {
		patch := networkingv1alpha3.EnvoyFilter_EnvoyConfigObjectPatch{
			ApplyTo: networkingv1alpha3.EnvoyFilter_HTTP_ROUTE,
			Patch: &networkingv1alpha3.EnvoyFilter_Patch{
				Operation: networkingv1alpha3.EnvoyFilter_Patch_INSERT_BEFORE,
			},
		}
		f := NewEnvoyFilterBuilder().
			WithName("test").
			WithNamespace("test").
			WithAnnotation("test", "test").
			WithAnnotation("foo", "bar").
			WithWorkloadSelector("app", "test").
			WithWorkloadSelector("foo", "bar").
			WithConfigPatch(&patch).
			Build()
		It("has added name and namespace", func() {
			Expect(f.Name).To(Equal("test"))
			Expect(f.Namespace).To(Equal("test"))
		})
		It("has annotations added", func() {
			Expect(f.Annotations).To(HaveLen(2))
			Expect(f.Annotations).To(HaveKeyWithValue("test", "test"))
			Expect(f.Annotations).To(HaveKeyWithValue("foo", "bar"))
		})
		It("has workloadSelector added", func() {
			Expect(f.Spec.WorkloadSelector).ToNot(BeNil())
			Expect(f.Spec.WorkloadSelector.GetLabels()).To(HaveLen(2))
			Expect(f.Spec.WorkloadSelector.GetLabels()).To(HaveKeyWithValue("app", "test"))
			Expect(f.Spec.WorkloadSelector.GetLabels()).To(HaveKeyWithValue("foo", "bar"))
		})
		It("has ConfigPatch added", func() {
			Expect(f.Spec.ConfigPatches).To(HaveLen(1))
		})
	})
})

func TestEnvoyFilter(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "EnvoyFilter Package")
}
