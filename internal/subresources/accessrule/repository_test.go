package accessrule_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/kyma-project/api-gateway/internal/processing"
	"github.com/kyma-project/api-gateway/internal/subresources/accessrule"
	rulev1alpha1 "github.com/kyma-project/api-gateway/internal/types/ory/oathkeeper-maester/api/v1alpha1"
)

var _ = Describe("AccessRule Repository", func() {
	var (
		repo      accessrule.Repository
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
		Expect(rulev1alpha1.AddToScheme(scheme)).To(Succeed())

		k8sClient = fake.NewClientBuilder().WithScheme(scheme).Build()
		repo = accessrule.NewRepository(k8sClient)
	})

	Describe("GetAll", func() {
		Context("when no AccessRules exist", func() {
			It("should return an empty list", func() {
				// When
				result, err := repo.GetAll(ctx, labeler)

				// Then
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(BeEmpty())
			})
		})

		Context("when AccessRules with legacy owner labels exist", func() {
			BeforeEach(func() {
				rule1 := &rulev1alpha1.Rule{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "rule-1",
						Namespace:  "test-namespace",
						Generation: 1,
						Labels: map[string]string{
							"apirule.gateway.kyma-project.io/v1beta1": "test-apirule.test-namespace",
						},
					},
				}
				rule2 := &rulev1alpha1.Rule{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "rule-2",
						Namespace:  "test-namespace",
						Generation: 1,
						Labels: map[string]string{
							"apirule.gateway.kyma-project.io/v1beta1": "test-apirule.test-namespace",
						},
					},
				}

				Expect(k8sClient.Create(ctx, rule1)).To(Succeed())
				Expect(k8sClient.Create(ctx, rule2)).To(Succeed())
			})

			It("should return all AccessRules with legacy labels", func() {
				// When
				result, err := repo.GetAll(ctx, labeler)

				// Then
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(HaveLen(2))

				names := make([]string, len(result))
				for i, rule := range result {
					names[i] = rule.Name
				}
				Expect(names).To(ConsistOf("rule-1", "rule-2"))
			})
		})

		Context("when AccessRules with new owner labels exist", func() {
			BeforeEach(func() {
				rule1 := &rulev1alpha1.Rule{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "rule-1",
						Namespace:  "test-namespace",
						Generation: 1,
						Labels: map[string]string{
							"apirule.gateway.kyma-project.io/name":      "test-apirule",
							"apirule.gateway.kyma-project.io/namespace": "test-namespace",
						},
					},
				}
				rule2 := &rulev1alpha1.Rule{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "rule-2",
						Namespace:  "test-namespace",
						Generation: 1,
						Labels: map[string]string{
							"apirule.gateway.kyma-project.io/name":      "test-apirule",
							"apirule.gateway.kyma-project.io/namespace": "test-namespace",
						},
					},
				}
				rule3 := &rulev1alpha1.Rule{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "rule-3",
						Namespace:  "test-namespace",
						Generation: 1,
						Labels: map[string]string{
							"apirule.gateway.kyma-project.io/name":      "other-apirule",
							"apirule.gateway.kyma-project.io/namespace": "test-namespace",
						},
					},
				}
				Expect(k8sClient.Create(ctx, rule1)).To(Succeed())
				Expect(k8sClient.Create(ctx, rule2)).To(Succeed())
				Expect(k8sClient.Create(ctx, rule3)).To(Succeed())
			})

			It("should return all AccessRules with new labels", func() {
				// When
				result, err := repo.GetAll(ctx, labeler)

				// Then
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(HaveLen(2))

				names := make([]string, len(result))
				for i, rule := range result {
					names[i] = rule.Name
				}
				Expect(names).To(ConsistOf("rule-1", "rule-2"))
			})
		})

		Context("when AccessRules with both legacy and new owner labels exist", func() {
			BeforeEach(func() {
				// AccessRule with legacy labels only
				rule1 := &rulev1alpha1.Rule{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "rule-legacy",
						Namespace:  "test-namespace",
						Generation: 1,
						Labels: map[string]string{
							"apirule.gateway.kyma-project.io/v1beta1": "test-apirule.test-namespace",
						},
					},
				}

				// AccessRule with new labels only
				rule2 := &rulev1alpha1.Rule{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "rule-new",
						Namespace:  "test-namespace",
						Generation: 1,
						Labels: map[string]string{
							"apirule.gateway.kyma-project.io/name":      "test-apirule",
							"apirule.gateway.kyma-project.io/namespace": "test-namespace",
						},
					},
				}
				// other AccessRule with legacy labels only but different APIRule
				rule3 := &rulev1alpha1.Rule{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "rule-other-legacy",
						Namespace:  "test-namespace",
						Generation: 1,
						Labels: map[string]string{
							"apirule.gateway.kyma-project.io/v1beta1": "other-apirule.test-namespace",
						},
					},
				}
				// other AccessRule with legacy labels only but different APIRule
				rule4 := &rulev1alpha1.Rule{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "rule-other-new",
						Namespace:  "test-namespace",
						Generation: 1,
						Labels: map[string]string{
							"apirule.gateway.kyma-project.io/name":      "other-apirule",
							"apirule.gateway.kyma-project.io/namespace": "test-namespace",
						},
					},
				}

				Expect(k8sClient.Create(ctx, rule1)).To(Succeed())
				Expect(k8sClient.Create(ctx, rule2)).To(Succeed())
				Expect(k8sClient.Create(ctx, rule3)).To(Succeed())
				Expect(k8sClient.Create(ctx, rule4)).To(Succeed())
			})

			It("should return all AccessRules from both label sets", func() {
				// When
				result, err := repo.GetAll(ctx, labeler)

				// Then
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(HaveLen(2))

				names := make([]string, len(result))
				for i, rule := range result {
					names[i] = rule.Name
				}
				Expect(names).To(ConsistOf("rule-new", "rule-legacy"))
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
