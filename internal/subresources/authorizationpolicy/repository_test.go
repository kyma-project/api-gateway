package authorizationpolicy_test

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
	"github.com/kyma-project/api-gateway/internal/subresources/authorizationpolicy"
)

var _ = Describe("AuthorizationPolicy Repository", func() {
	var (
		repo      authorizationpolicy.Repository
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
		repo = authorizationpolicy.NewRepository(k8sClient)
	})

	Describe("GetAll", func() {
		Context("when no AuthorizationPolicies exist", func() {
			It("should return an empty list", func() {
				// When
				result, err := repo.GetAll(ctx, labeler)

				// Then
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(BeEmpty())
			})
		})

		Context("when AuthorizationPolicies with legacy owner labels exist", func() {
			BeforeEach(func() {
				ap1 := &securityv1beta1.AuthorizationPolicy{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "ap-1",
						Namespace:  "test-namespace",
						Generation: 1,
						Labels: map[string]string{
							"apirule.gateway.kyma-project.io/v1beta1": "test-apirule.test-namespace",
						},
					},
				}
				ap2 := &securityv1beta1.AuthorizationPolicy{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "ap-2",
						Namespace:  "test-namespace",
						Generation: 1,
						Labels: map[string]string{
							"apirule.gateway.kyma-project.io/v1beta1": "test-apirule.test-namespace",
						},
					},
				}

				Expect(k8sClient.Create(ctx, ap1)).To(Succeed())
				Expect(k8sClient.Create(ctx, ap2)).To(Succeed())
			})

			It("should return all AuthorizationPolicies with legacy labels", func() {
				// When
				result, err := repo.GetAll(ctx, labeler)

				// Then
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(HaveLen(2))

				names := make([]string, len(result))
				for i, ap := range result {
					names[i] = ap.Name
				}
				Expect(names).To(ConsistOf("ap-1", "ap-2"))
			})
		})

		Context("when AuthorizationPolicies with new owner labels exist", func() {
			BeforeEach(func() {
				ap1 := &securityv1beta1.AuthorizationPolicy{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "ap-1",
						Namespace:  "test-namespace",
						Generation: 1,
						Labels: map[string]string{
							"apirule.gateway.kyma-project.io/name":      "test-apirule",
							"apirule.gateway.kyma-project.io/namespace": "test-namespace",
						},
					},
				}
				ap2 := &securityv1beta1.AuthorizationPolicy{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "ap-2",
						Namespace:  "test-namespace",
						Generation: 1,
						Labels: map[string]string{
							"apirule.gateway.kyma-project.io/name":      "test-apirule",
							"apirule.gateway.kyma-project.io/namespace": "test-namespace",
						},
					},
				}
				ap3 := &securityv1beta1.AuthorizationPolicy{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "ap-3",
						Namespace:  "test-namespace",
						Generation: 1,
						Labels: map[string]string{
							"apirule.gateway.kyma-project.io/name":      "other-apirule",
							"apirule.gateway.kyma-project.io/namespace": "test-namespace",
						},
					},
				}
				Expect(k8sClient.Create(ctx, ap1)).To(Succeed())
				Expect(k8sClient.Create(ctx, ap2)).To(Succeed())
				Expect(k8sClient.Create(ctx, ap3)).To(Succeed())
			})

			It("should return all AuthorizationPolicies with new labels", func() {
				// When
				result, err := repo.GetAll(ctx, labeler)

				// Then
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(HaveLen(2))

				names := make([]string, len(result))
				for i, ap := range result {
					names[i] = ap.Name
				}
				Expect(names).To(ConsistOf("ap-1", "ap-2"))
			})
		})

		Context("when AuthorizationPolicies with both legacy and new owner labels exist", func() {
			BeforeEach(func() {
				// AuthorizationPolicy with legacy labels only
				ap1 := &securityv1beta1.AuthorizationPolicy{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "ap-legacy",
						Namespace:  "test-namespace",
						Generation: 1,
						Labels: map[string]string{
							"apirule.gateway.kyma-project.io/v1beta1": "test-apirule.test-namespace",
						},
					},
				}

				// AuthorizationPolicy with new labels only
				ap2 := &securityv1beta1.AuthorizationPolicy{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "ap-new",
						Namespace:  "test-namespace",
						Generation: 1,
						Labels: map[string]string{
							"apirule.gateway.kyma-project.io/name":      "test-apirule",
							"apirule.gateway.kyma-project.io/namespace": "test-namespace",
						},
					},
				}
				// other AuthorizationPolicy with legacy labels only but different APIRule
				ap3 := &securityv1beta1.AuthorizationPolicy{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "ap-other-legacy",
						Namespace:  "test-namespace",
						Generation: 1,
						Labels: map[string]string{
							"apirule.gateway.kyma-project.io/v1beta1": "other-apirule.test-namespace",
						},
					},
				}
				// other AuthorizationPolicy with new labels only but different APIRule
				ap4 := &securityv1beta1.AuthorizationPolicy{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "ap-other-new",
						Namespace:  "test-namespace",
						Generation: 1,
						Labels: map[string]string{
							"apirule.gateway.kyma-project.io/name":      "other-apirule",
							"apirule.gateway.kyma-project.io/namespace": "test-namespace",
						},
					},
				}

				Expect(k8sClient.Create(ctx, ap1)).To(Succeed())
				Expect(k8sClient.Create(ctx, ap2)).To(Succeed())
				Expect(k8sClient.Create(ctx, ap3)).To(Succeed())
				Expect(k8sClient.Create(ctx, ap4)).To(Succeed())
			})

			It("should return all AuthorizationPolicies from both label sets", func() {
				// When
				result, err := repo.GetAll(ctx, labeler)

				// Then
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(HaveLen(2))

				names := make([]string, len(result))
				for i, ap := range result {
					names[i] = ap.Name
				}
				Expect(names).To(ConsistOf("ap-new", "ap-legacy"))
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
