package migration

import (
	"context"
	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"github.com/kyma-project/api-gateway/internal/processing"
	rulev1alpha1 "github.com/kyma-project/api-gateway/internal/types/ory/oathkeeper-maester/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("RuleDeletionProcessor", func() {
	It("should delete Ory Rules owned by the APIRule by ownership label", func() {
		// given
		apiRule := &gatewayv1beta1.APIRule{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-api-rule",
				Namespace: "test-namespace",
			},
		}

		rules := []*rulev1alpha1.Rule{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-rule-1",
					Namespace: "test-namespace",

					Labels: map[string]string{
						"apirule.gateway.kyma-project.io/v1beta1": "test-api-rule.test-namespace",
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-rule-2",
					Namespace: "test-namespace",

					Labels: map[string]string{
						"apirule.gateway.kyma-project.io/v1beta1": "test-api-rule.test-namespace",
					},
				},
			},
		}

		scheme := runtime.NewScheme()
		err := rulev1alpha1.AddToScheme(scheme)
		Expect(err).To(BeNil())

		k8sClientBuilder := fake.NewClientBuilder().WithScheme(scheme)
		for _, rule := range rules {
			k8sClientBuilder.WithObjects(rule)
		}
		k8sClient := k8sClientBuilder.Build()

		processor := NewAccessRuleDeletionProcessor(processing.ReconciliationConfig{}, apiRule)

		// when
		changes, err := processor.EvaluateReconciliation(context.Background(), k8sClient)

		// then
		Expect(err).To(BeNil())
		Expect(changes).To(HaveLen(2))
		Expect(changes[0].Action).To(Equal(processing.NewObjectDeleteAction(rules[0]).Action))
		Expect(changes[1].Action).To(Equal(processing.NewObjectDeleteAction(rules[1]).Action))
	})

	It("should not delete Ory Rules not owned by the APIRule", func() {
		// given
		apiRule := &gatewayv1beta1.APIRule{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-api-rule",
				Namespace: "test-namespace",
			},
		}

		rules := []*rulev1alpha1.Rule{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-rule-1",
					Namespace: "test-namespace",

					Labels: map[string]string{
						"apirule.gateway.kyma-project.io/v1beta1": "other-api-rule.test-namespace",
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-rule-2",
					Namespace: "test-namespace",

					Labels: map[string]string{
						"apirule.gateway.kyma-project.io/v1beta1": "other-api-rule.test-namespace",
					},
				},
			},
		}

		scheme := runtime.NewScheme()
		err := rulev1alpha1.AddToScheme(scheme)
		Expect(err).To(BeNil())

		k8sClientBuilder := fake.NewClientBuilder().WithScheme(scheme)
		for _, rule := range rules {
			k8sClientBuilder.WithObjects(rule)
		}
		k8sClient := k8sClientBuilder.Build()

		processor := NewAccessRuleDeletionProcessor(processing.ReconciliationConfig{}, apiRule)

		// when
		changes, err := processor.EvaluateReconciliation(context.Background(), k8sClient)

		// then
		Expect(err).To(BeNil())
		Expect(changes).To(HaveLen(0))
	})
})
