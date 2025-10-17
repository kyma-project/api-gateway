package requestauthentication_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/kyma-project/api-gateway/internal/processing"
	"github.com/kyma-project/api-gateway/internal/subresources/requestauthentication"
)

var _ = Describe("RequestAuthentication Repository", func() {
	var (
		repo      requestauthentication.Repository
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
		Expect(securityv1beta1.AddToScheme(scheme)).To(Succeed())

		k8sClient = fake.NewClientBuilder().WithScheme(scheme).Build()
		repo = requestauthentication.NewRepository(k8sClient)
	})

	Describe("GetAll", func() {
		Context("when no RequestAuthentications exist", func() {
			It("should return an empty list", func() {
				// When
				result, err := repo.GetAll(ctx, labeler)

				// Then
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(BeEmpty())
			})
		})

		Context("when RequestAuthentications with legacy owner labels exist", func() {
			BeforeEach(func() {
				ra1 := &securityv1beta1.RequestAuthentication{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "ra-1",
						Namespace:  "test-namespace",
						Generation: 1,
						Labels: map[string]string{
							"apirule.gateway.kyma-project.io/v1beta1": "test-apirule.test-namespace",
						},
					},
				}
				ra2 := &securityv1beta1.RequestAuthentication{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "ra-2",
						Namespace:  "test-namespace",
						Generation: 1,
						Labels: map[string]string{
							"apirule.gateway.kyma-project.io/v1beta1": "test-apirule.test-namespace",
						},
					},
				}

				Expect(k8sClient.Create(ctx, ra1)).To(Succeed())
				Expect(k8sClient.Create(ctx, ra2)).To(Succeed())
			})

			It("should return all RequestAuthentications with legacy labels", func() {
				// When
				result, err := repo.GetAll(ctx, labeler)

				// Then
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(HaveLen(2))

				names := make([]string, len(result))
				for i, ra := range result {
					names[i] = ra.Name
				}
				Expect(names).To(ConsistOf("ra-1", "ra-2"))
			})
		})

		Context("when RequestAuthentications with new owner labels exist", func() {
			BeforeEach(func() {
				ra1 := &securityv1beta1.RequestAuthentication{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "ra-1",
						Namespace:  "test-namespace",
						Generation: 1,
						Labels: map[string]string{
							"apirule.gateway.kyma-project.io/name":      "test-apirule",
							"apirule.gateway.kyma-project.io/namespace": "test-namespace",
						},
					},
				}
				ra2 := &securityv1beta1.RequestAuthentication{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "ra-2",
						Namespace:  "test-namespace",
						Generation: 1,
						Labels: map[string]string{
							"apirule.gateway.kyma-project.io/name":      "test-apirule",
							"apirule.gateway.kyma-project.io/namespace": "test-namespace",
						},
					},
				}
				ra3 := &securityv1beta1.RequestAuthentication{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "ra-3",
						Namespace:  "test-namespace",
						Generation: 1,
						Labels: map[string]string{
							"apirule.gateway.kyma-project.io/name":      "other-apirule",
							"apirule.gateway.kyma-project.io/namespace": "test-namespace",
						},
					},
				}
				Expect(k8sClient.Create(ctx, ra1)).To(Succeed())
				Expect(k8sClient.Create(ctx, ra2)).To(Succeed())
				Expect(k8sClient.Create(ctx, ra3)).To(Succeed())
			})

			It("should return all RequestAuthentications with new labels", func() {
				// When
				result, err := repo.GetAll(ctx, labeler)

				// Then
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(HaveLen(2))

				names := make([]string, len(result))
				for i, ra := range result {
					names[i] = ra.Name
				}
				Expect(names).To(ConsistOf("ra-1", "ra-2"))
			})
		})

		Context("when RequestAuthentications with both legacy and new owner labels exist", func() {
			BeforeEach(func() {
				// RequestAuthentication with legacy labels only
				ra1 := &securityv1beta1.RequestAuthentication{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "ra-legacy",
						Namespace:  "test-namespace",
						Generation: 1,
						Labels: map[string]string{
							"apirule.gateway.kyma-project.io/v1beta1": "test-apirule.test-namespace",
						},
					},
				}

				// RequestAuthentication with new labels only
				ra2 := &securityv1beta1.RequestAuthentication{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "ra-new",
						Namespace:  "test-namespace",
						Generation: 1,
						Labels: map[string]string{
							"apirule.gateway.kyma-project.io/name":      "test-apirule",
							"apirule.gateway.kyma-project.io/namespace": "test-namespace",
						},
					},
				}
				// other RequestAuthentication with legacy labels only but different APIRule
				ra3 := &securityv1beta1.RequestAuthentication{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "ra-other-legacy",
						Namespace:  "test-namespace",
						Generation: 1,
						Labels: map[string]string{
							"apirule.gateway.kyma-project.io/v1beta1": "other-apirule.test-namespace",
						},
					},
				}
				// other RequestAuthentication with new labels only but different APIRule
				ra4 := &securityv1beta1.RequestAuthentication{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "ra-other-new",
						Namespace:  "test-namespace",
						Generation: 1,
						Labels: map[string]string{
							"apirule.gateway.kyma-project.io/name":      "other-apirule",
							"apirule.gateway.kyma-project.io/namespace": "test-namespace",
						},
					},
				}

				Expect(k8sClient.Create(ctx, ra1)).To(Succeed())
				Expect(k8sClient.Create(ctx, ra2)).To(Succeed())
				Expect(k8sClient.Create(ctx, ra3)).To(Succeed())
				Expect(k8sClient.Create(ctx, ra4)).To(Succeed())
			})

			It("should return all RequestAuthentications from both label sets", func() {
				// When
				result, err := repo.GetAll(ctx, labeler)

				// Then
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(HaveLen(2))

				names := make([]string, len(result))
				for i, ra := range result {
					names[i] = ra.Name
				}
				Expect(names).To(ConsistOf("ra-new", "ra-legacy"))
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
