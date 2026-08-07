package rules_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/reporters"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	"github.com/kyma-project/api-gateway/internal/processing/processors/v2alpha1/rules"
	rulev1alpha1 "github.com/kyma-project/api-gateway/internal/types/ory/oathkeeper-maester/api/v1alpha1"
)

func TestRulesProcessor(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Rules Processor Suite")
}

var _ = ReportAfterSuite("custom reporter", func(report types.Report) {
	if key, ok := os.LookupEnv("ARTIFACTS"); ok {
		reportsFilename := fmt.Sprintf("%s/%s", key, "junit-rules-processor.xml")
		err := reporters.GenerateJUnitReport(report, reportsFilename)
		Expect(err).NotTo(HaveOccurred())
	}
})

var testLogger = ctrl.Log.WithName("test")

var _ = Describe("DeletionProcessor", func() {
	var (
		ctx     context.Context
		apiRule *gatewayv2alpha1.APIRule
	)

	BeforeEach(func() {
		ctx = context.Background()
		host := gatewayv2alpha1.Host("myservice.test.com")
		apiRule = &gatewayv2alpha1.APIRule{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-apirule",
				Namespace: "test-namespace",
			},
			Spec: gatewayv2alpha1.APIRuleSpec{
				Hosts: []*gatewayv2alpha1.Host{&host},
			},
		}
	})

	Context("when the Ory Rule CRD is not installed", func() {
		It("should return empty without error", func() {
			scheme := runtime.NewScheme()
			// rulev1alpha1 intentionally NOT added to scheme to simulate missing CRD
			fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()

			processor := rules.NewDeletionProcessor(&testLogger, apiRule, fakeClient)
			changes, err := processor.EvaluateReconciliation(ctx, fakeClient)

			Expect(err).NotTo(HaveOccurred())
			Expect(changes).To(BeEmpty())
		})
	})

	Context("when no Ory Rules exist for the APIRule", func() {
		It("should return an empty slice", func() {
			fakeClient := fakeClientWithRules()

			processor := rules.NewDeletionProcessor(&testLogger, apiRule, fakeClient)
			changes, err := processor.EvaluateReconciliation(ctx, fakeClient)

			Expect(err).NotTo(HaveOccurred())
			Expect(changes).To(BeEmpty())
		})
	})

	Context("when orphaned Ory Rules exist for the APIRule", func() {
		It("should return a Delete action for each rule", func() {
			oryRule1 := oryRuleWithNewLabels("rule-1", "test-namespace", "test-apirule", "test-namespace")
			oryRule2 := oryRuleWithNewLabels("rule-2", "test-namespace", "test-apirule", "test-namespace")
			fakeClient := fakeClientWithRules(oryRule1, oryRule2)

			processor := rules.NewDeletionProcessor(&testLogger, apiRule, fakeClient)
			changes, err := processor.EvaluateReconciliation(ctx, fakeClient)

			Expect(err).NotTo(HaveOccurred())
			Expect(changes).To(HaveLen(2))
			for _, change := range changes {
				Expect(change.Action.String()).To(Equal("delete"))
			}
		})

		It("should not return rules belonging to a different APIRule", func() {
			oryRule := oryRuleWithNewLabels("rule-1", "test-namespace", "other-apirule", "test-namespace")
			fakeClient := fakeClientWithRules(oryRule)

			processor := rules.NewDeletionProcessor(&testLogger, apiRule, fakeClient)
			changes, err := processor.EvaluateReconciliation(ctx, fakeClient)

			Expect(err).NotTo(HaveOccurred())
			Expect(changes).To(BeEmpty())
		})
	})
})

func fakeClientWithRules(objs ...client.Object) client.Client {
	scheme := runtime.NewScheme()
	Expect(rulev1alpha1.AddToScheme(scheme)).To(Succeed())
	return fake.NewClientBuilder().WithScheme(scheme).WithObjects(objs...).Build()
}

func oryRuleWithNewLabels(name, namespace, ownerName, ownerNamespace string) *rulev1alpha1.Rule {
	return &rulev1alpha1.Rule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"apirule.gateway.kyma-project.io/name":      ownerName,
				"apirule.gateway.kyma-project.io/namespace": ownerNamespace,
			},
		},
	}
}
