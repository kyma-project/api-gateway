package virtualservice_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/kyma-project/api-gateway/internal/processing"
	"github.com/kyma-project/api-gateway/internal/subresources/virtualservice"
)

var _ = Describe("VirtualService Repository", func() {
	var (
		repo      virtualservice.Repository
		k8sClient client.Client
		ctx       context.Context
		labeler   *mockLabeler
	)

	BeforeEach(func() {
		ctx = context.Background()
		labeler = &mockLabeler{
			name:      "test-apirule",
			namespace: "test-namespace",
		}

		scheme := runtime.NewScheme()
		Expect(networkingv1beta1.AddToScheme(scheme)).To(Succeed())

		k8sClient = fake.NewClientBuilder().WithScheme(scheme).Build()
		repo = virtualservice.NewRepository(k8sClient)
	})

	Describe("GetAll", func() {
		Context("when no VirtualServices exist", func() {
			It("should return an empty list", func() {
				// When
				result, err := repo.GetAll(ctx, labeler)

				// Then
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(BeEmpty())
			})
		})

		Context("when VirtualServices with legacy owner labels exist", func() {
			BeforeEach(func() {
				vs1 := &networkingv1beta1.VirtualService{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "vs-1",
						Namespace:  "test-namespace",
						Generation: 1,
						Labels: map[string]string{
							"apirule.gateway.kyma-project.io/v1beta1": "test-apirule.test-namespace",
						},
					},
				}
				vs2 := &networkingv1beta1.VirtualService{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "vs-2",
						Namespace:  "test-namespace",
						Generation: 1,
						Labels: map[string]string{
							"apirule.gateway.kyma-project.io/v1beta1": "test-apirule.test-namespace",
						},
					},
				}

				Expect(k8sClient.Create(ctx, vs1)).To(Succeed())
				Expect(k8sClient.Create(ctx, vs2)).To(Succeed())
			})

			It("should return all VirtualServices with legacy labels", func() {
				// When
				result, err := repo.GetAll(ctx, labeler)

				// Then
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(HaveLen(2))

				names := make([]string, len(result))
				for i, vs := range result {
					names[i] = vs.Name
				}
				Expect(names).To(ConsistOf("vs-1", "vs-2"))
			})
		})

		Context("when VirtualServices with new owner labels exist", func() {
			BeforeEach(func() {
				vs1 := &networkingv1beta1.VirtualService{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "vs-1",
						Namespace:  "test-namespace",
						Generation: 1,
						Labels: map[string]string{
							"apirule.gateway.kyma-project.io/name":      "test-apirule",
							"apirule.gateway.kyma-project.io/namespace": "test-namespace",
						},
					},
				}
				vs2 := &networkingv1beta1.VirtualService{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "vs-2",
						Namespace:  "test-namespace",
						Generation: 1,
						Labels: map[string]string{
							"apirule.gateway.kyma-project.io/name":      "test-apirule",
							"apirule.gateway.kyma-project.io/namespace": "test-namespace",
						},
					},
				}
				vs3 := &networkingv1beta1.VirtualService{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "vs-3",
						Namespace:  "test-namespace",
						Generation: 1,
						Labels: map[string]string{
							"apirule.gateway.kyma-project.io/name":      "other-apirule",
							"apirule.gateway.kyma-project.io/namespace": "test-namespace",
						},
					},
				}

				Expect(k8sClient.Create(ctx, vs1)).To(Succeed())
				Expect(k8sClient.Create(ctx, vs2)).To(Succeed())
				Expect(k8sClient.Create(ctx, vs3)).To(Succeed())
			})

			It("should return all VirtualServices with new labels", func() {
				// When
				result, err := repo.GetAll(ctx, labeler)

				// Then
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(HaveLen(2))

				names := make([]string, len(result))
				for i, vs := range result {
					names[i] = vs.Name
				}
				Expect(names).To(ConsistOf("vs-1", "vs-2"))
			})
		})

		Context("when VirtualServices with both legacy and new owner labels exist", func() {
			BeforeEach(func() {
				// VirtualService with legacy labels only
				vs1 := &networkingv1beta1.VirtualService{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "vs-legacy",
						Namespace:  "test-namespace",
						Generation: 1,
						Labels: map[string]string{
							"apirule.gateway.kyma-project.io/v1beta1": "test-apirule.test-namespace",
						},
					},
				}

				// VirtualService with new labels only
				vs2 := &networkingv1beta1.VirtualService{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "vs-new",
						Namespace:  "test-namespace",
						Generation: 1,
						Labels: map[string]string{
							"apirule.gateway.kyma-project.io/name":      "test-apirule",
							"apirule.gateway.kyma-project.io/namespace": "test-namespace",
						},
					},
				}
				// other VirtualService with legacy labels only but different APIRule
				vs3 := &networkingv1beta1.VirtualService{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "vs-other-legacy",
						Namespace:  "test-namespace",
						Generation: 1,
						Labels: map[string]string{
							"apirule.gateway.kyma-project.io/v1beta1": "other-apirule.test-namespace",
						},
					},
				}
				// other VirtualService with new labels only but different APIRule
				vs4 := &networkingv1beta1.VirtualService{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "vs-other-new",
						Namespace:  "test-namespace",
						Generation: 1,
						Labels: map[string]string{
							"apirule.gateway.kyma-project.io/name":      "other-apirule",
							"apirule.gateway.kyma-project.io/namespace": "test-namespace",
						},
					},
				}

				Expect(k8sClient.Create(ctx, vs1)).To(Succeed())
				Expect(k8sClient.Create(ctx, vs2)).To(Succeed())
				Expect(k8sClient.Create(ctx, vs3)).To(Succeed())
				Expect(k8sClient.Create(ctx, vs4)).To(Succeed())
			})

			It("should return all VirtualServices from both label sets", func() {
				// When
				result, err := repo.GetAll(ctx, labeler)

				// Then
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(HaveLen(2))

				names := make([]string, len(result))
				for i, vs := range result {
					names[i] = vs.Name
				}
				Expect(names).To(ConsistOf("vs-new", "vs-legacy"))
			})
		})
	})
})

type mockLabeler struct {
	name      string
	namespace string
}

func (m *mockLabeler) GetName() string {
	return m.name
}

func (m *mockLabeler) GetNamespace() string {
	return m.namespace
}

var _ processing.Labeler = &mockLabeler{}
