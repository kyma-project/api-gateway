package migration

import (
	"github.com/go-logr/logr"
	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	"github.com/kyma-project/api-gateway/internal/processing"
	processors2 "github.com/kyma-project/api-gateway/internal/processing/processors"
	v2alpha1VirtualService "github.com/kyma-project/api-gateway/internal/processing/processors/v2alpha1/virtualservice"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("NewMigrationProcessors", func() {
	var (
		config = processing.ReconciliationConfig{}
		log    = logr.Discard()
	)
	It("should return applyIstioAuthorizationMigrationStep when annotation is not set", func() {
		apirule := &gatewayv2alpha1.APIRule{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{},
			},
		}

		apiruleBeta := &gatewayv1beta1.APIRule{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{},
			},
		}

		processors := NewMigrationProcessors(apirule, apiruleBeta, config, &log)
		Expect(processors).To(HaveLen(2))
		Expect(processors[0]).To(BeAssignableToTypeOf(processors2.AuthorizationPolicyProcessor{}))
		Expect(processors[1]).To(BeAssignableToTypeOf(processors2.RequestAuthenticationProcessor{}))
	})

	DescribeTable("should return processors according to migration step", func(annotation string, expectedProcessors []processing.ReconciliationProcessor) {
		// when
		apirule := &gatewayv2alpha1.APIRule{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					"gateway.kyma-project.io/migration-step": annotation,
				},
			},
		}
		apiruleBeta := &gatewayv1beta1.APIRule{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					"gateway.kyma-project.io/migration-step": annotation,
				},
			},
		}
		processors := NewMigrationProcessors(apirule, apiruleBeta, config, &log)
		// then
		Expect(len(processors)).To(Equal(len(expectedProcessors)))
		for i, processor := range expectedProcessors {
			Expect(processors[i]).To(BeAssignableToTypeOf(processor))
		}

	},
		Entry("should return AP and RA processors when annotation is not set",
			"",
			[]processing.ReconciliationProcessor{
				processors2.AuthorizationPolicyProcessor{},
				processors2.RequestAuthenticationProcessor{},
			},
		),
		Entry("should return VS, AP and RA processors when current step is switchVsToService",
			string(applyIstioAuthorizationMigrationStep),
			[]processing.ReconciliationProcessor{
				v2alpha1VirtualService.VirtualServiceProcessor{},
				processors2.AuthorizationPolicyProcessor{},
				processors2.RequestAuthenticationProcessor{},
			},
		),
		Entry("should return AccessRule deletion, VS, AP and RA processors when current step is removeOryRule",
			string(switchVsToService),
			[]processing.ReconciliationProcessor{
				accessRuleDeletionProcessor{},
				v2alpha1VirtualService.VirtualServiceProcessor{},
				processors2.AuthorizationPolicyProcessor{},
				processors2.RequestAuthenticationProcessor{},
			},
		),
	)
})

var _ = Describe("nextMigrationStep", func() {
	DescribeTable("should return correct next step", func(annotation string, expectedStep string) {
		// when
		step := nextMigrationStep(&gatewayv1beta1.APIRule{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					annotationName: annotation,
				},
			},
		})
		// then
		Expect(string(step)).To(Equal(expectedStep))
	},
		Entry("should return applyIstioAuthorizationMigrationStep when annotation is set to something not expected", "another", "apply-istio-authorization"),
		Entry("should return switchVsToService when current step is applyIstioAuthorizationMigrationStep", "apply-istio-authorization", "vs-switch-to-service"),
		Entry("should return removeOryRule when current step is switchVsToService", "vs-switch-to-service", "remove-ory-rule"),
	)
})
