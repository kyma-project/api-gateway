package subresources_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/api-gateway/internal/subresources"
)

var _ = Describe("MergeResourceSlices", func() {

	Context("with empty slices", func() {
		It("should return empty result when both slices are empty", func() {
			// When
			result := subresources.MergeResourceSlices(
				[]*securityv1beta1.RequestAuthentication{},
				[]*securityv1beta1.RequestAuthentication{},
			)

			// Then
			Expect(result).To(BeEmpty())
		})

		It("should return slice2 items when slice1 is empty", func() {
			// Given
			slice2 := []*securityv1beta1.RequestAuthentication{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ra-1",
						Namespace: "default",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ra-2",
						Namespace: "default",
					},
				},
			}

			// When
			result := subresources.MergeResourceSlices([]*securityv1beta1.RequestAuthentication{}, slice2)

			// Then
			Expect(result).To(HaveLen(2))
			// Verify both expected items are present (order not guaranteed)
			var names []string
			for _, item := range result {
				names = append(names, item.Name)
				Expect(item.Namespace).To(Equal("default"))
			}
			Expect(names).To(ContainElements([]string{"ra-1", "ra-2"}))
		})

		It("should return slice1 items when slice2 is empty", func() {
			// Given
			slice1 := []*securityv1beta1.RequestAuthentication{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ra-1",
						Namespace: "default",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ra-2",
						Namespace: "default",
					},
				},
			}

			// When
			result := subresources.MergeResourceSlices(slice1, []*securityv1beta1.RequestAuthentication{})

			// Then
			Expect(result).To(HaveLen(2))
			// Verify both expected items are present (order not guaranteed)
			var names []string
			for _, item := range result {
				names = append(names, item.Name)
				Expect(item.Namespace).To(Equal("default"))
			}
			Expect(names).To(ContainElements([]string{"ra-1", "ra-2"}))
		})
	})

	Context("with no duplicate items", func() {
		It("should merge all items from both slices", func() {
			// Given
			slice1 := []*securityv1beta1.RequestAuthentication{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ra-1",
						Namespace: "default",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ra-2",
						Namespace: "default",
					},
				},
			}

			slice2 := []*securityv1beta1.RequestAuthentication{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ra-3",
						Namespace: "default",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ra-4",
						Namespace: "default",
					},
				},
			}

			// When
			result := subresources.MergeResourceSlices(slice1, slice2)

			// Then
			Expect(result).To(HaveLen(4))
			// Verify all expected items are present (order not guaranteed)
			var names []string
			for _, item := range result {
				names = append(names, item.Name)
			}
			Expect(names).To(ContainElements([]string{"ra-1", "ra-2", "ra-3", "ra-4"}))
		})
	})

	Context("with duplicate items", func() {
		It("should deduplicate items with same name and namespace", func() {
			// Given
			slice1 := []*securityv1beta1.RequestAuthentication{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ra-1",
						Namespace: "default",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ra-2",
						Namespace: "default",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ra-3",
						Namespace: "default",
					},
				},
			}

			slice2 := []*securityv1beta1.RequestAuthentication{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ra-2", // Duplicate
						Namespace: "default",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ra-4",
						Namespace: "default",
					},
				},
			}

			// When
			result := subresources.MergeResourceSlices(slice1, slice2)

			// Then
			Expect(result).To(HaveLen(4))

			// Verify no duplicates and all expected items are present
			var names []string
			for _, item := range result {
				names = append(names, item.Name)
				Expect(item.Namespace).To(Equal("default"))
			}
			// Verify all expected items are present
			Expect(names).To(ContainElements([]string{"ra-1", "ra-2", "ra-3", "ra-4"}))
			// Verify no duplicates (length check already done above)
		})

		It("should keep all items when all are duplicates", func() {
			// Given
			slice1 := []*securityv1beta1.RequestAuthentication{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ra-1",
						Namespace: "default",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ra-2",
						Namespace: "default",
					},
				},
			}

			slice2 := []*securityv1beta1.RequestAuthentication{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ra-1",
						Namespace: "default",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ra-2",
						Namespace: "default",
					},
				},
			}

			// When
			result := subresources.MergeResourceSlices(slice1, slice2)

			// Then
			Expect(result).To(HaveLen(2))

			// Verify both expected items are present
			var names []string
			for _, item := range result {
				names = append(names, item.Name)
				Expect(item.Namespace).To(Equal("default"))
			}
			Expect(names).To(ContainElements([]string{"ra-1", "ra-2"}))
		})

		It("should preserve item from slice1 when generations are equal", func() {
			// Given - both have same generation (0), slice1 should be preserved
			slice1 := []*securityv1beta1.RequestAuthentication{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ra-1",
						Namespace: "default",
						Labels: map[string]string{
							"source": "slice1",
						},
					},
				},
			}

			slice2 := []*securityv1beta1.RequestAuthentication{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ra-1",
						Namespace: "default",
						Labels: map[string]string{
							"source": "slice2",
						},
					},
				},
			}

			// When
			result := subresources.MergeResourceSlices(slice1, slice2)

			// Then
			Expect(result).To(HaveLen(1))
			Expect(result[0].Name).To(Equal("ra-1"))
			Expect(result[0].Namespace).To(Equal("default"))
			// When generations are equal, slice1 is preserved (processed first)
			Expect(result[0].Labels["source"]).To(Equal("slice1"))
		})

		It("should preserve item with higher generation when duplicate exists", func() {
			// Given - slice1 has generation 1, slice2 has generation 2
			slice1 := []*securityv1beta1.RequestAuthentication{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "ra-1",
						Namespace:  "default",
						Generation: 1,
						Labels: map[string]string{
							"source": "slice1",
						},
					},
				},
			}

			slice2 := []*securityv1beta1.RequestAuthentication{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "ra-1",
						Namespace:  "default",
						Generation: 2,
						Labels: map[string]string{
							"source": "slice2",
						},
					},
				},
			}

			// When
			result := subresources.MergeResourceSlices(slice1, slice2)

			// Then
			Expect(result).To(HaveLen(1))
			Expect(result[0].Name).To(Equal("ra-1"))
			Expect(result[0].Namespace).To(Equal("default"))
			Expect(result[0].Labels["source"]).To(Equal("slice2"))
			Expect(result[0].Generation).To(Equal(int64(2)))
		})

		It("should preserve item from slice1 when it has higher generation", func() {
			// Given - slice1 has generation 3, slice2 has generation 1
			slice1 := []*securityv1beta1.RequestAuthentication{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "ra-1",
						Namespace:  "default",
						Generation: 3,
						Labels: map[string]string{
							"source": "slice1",
						},
					},
				},
			}

			slice2 := []*securityv1beta1.RequestAuthentication{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "ra-1",
						Namespace:  "default",
						Generation: 1,
						Labels: map[string]string{
							"source": "slice2",
						},
					},
				},
			}

			// When
			result := subresources.MergeResourceSlices(slice1, slice2)

			// Then
			Expect(result).To(HaveLen(1))
			Expect(result[0].Name).To(Equal("ra-1"))
			Expect(result[0].Namespace).To(Equal("default"))
			Expect(result[0].Labels["source"]).To(Equal("slice1"))
			Expect(result[0].Generation).To(Equal(int64(3)))
		})
	})

	Context("with different namespaces", func() {
		It("should not deduplicate items with same name but different namespaces", func() {
			// Given
			slice1 := []*securityv1beta1.RequestAuthentication{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ra-1",
						Namespace: "namespace-a",
					},
				},
			}

			slice2 := []*securityv1beta1.RequestAuthentication{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ra-1",
						Namespace: "namespace-b",
					},
				},
			}

			// When
			result := subresources.MergeResourceSlices(slice1, slice2)

			// Then
			Expect(result).To(HaveLen(2))
			// Verify both namespaces are present (order not guaranteed)
			var namespaces []string
			for _, item := range result {
				namespaces = append(namespaces, item.Namespace)
			}
			Expect(namespaces).To(ContainElements([]string{"namespace-a", "namespace-b"}))
		})
	})

	Context("with different resource types", func() {
		It("should work with VirtualService resource type", func() {
			// Given
			slice1 := []*networkingv1beta1.VirtualService{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "vs-1",
						Namespace: "default",
					},
				},
			}

			slice2 := []*networkingv1beta1.VirtualService{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "vs-2",
						Namespace: "default",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "vs-1", // Duplicate
						Namespace: "default",
					},
				},
			}

			// When
			result := subresources.MergeResourceSlices(slice1, slice2)

			// Then
			Expect(result).To(HaveLen(2))
			var exnames []string
			// Verify the correct items are present
			for _, item := range result {
				exnames = append(exnames, item.Name)
			}
			Expect(exnames).To(ContainElements([]string{"vs-1", "vs-2"}))
		})
	})

	Context("with order preservation", func() {
		It("should merge items from both slices", func() {
			// Given
			slice1 := []*securityv1beta1.RequestAuthentication{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ra-a",
						Namespace: "default",
					},
				},
			}

			slice2 := []*securityv1beta1.RequestAuthentication{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ra-b",
						Namespace: "default",
					},
				},
			}

			// When
			result := subresources.MergeResourceSlices(slice1, slice2)

			// Then
			Expect(result).To(HaveLen(2))
			// Verify both items are present (order not guaranteed due to map iteration)
			var names []string
			for _, item := range result {
				names = append(names, item.Name)
			}
			Expect(names).To(ContainElements([]string{"ra-a", "ra-b"}))
		})
	})
})
