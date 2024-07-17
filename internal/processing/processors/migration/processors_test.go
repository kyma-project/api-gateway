package migration

import (
	"github.com/go-logr/logr"
	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	"github.com/kyma-project/api-gateway/internal/processing"
	processors2 "github.com/kyma-project/api-gateway/internal/processing/processors"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("NewMigrationProcessors", func() {
	var (
		config      = processing.ReconciliationConfig{}
		log         = logr.Discard()
		apiruleBeta = &gatewayv1beta1.APIRule{}
	)
	DescribeTable("should return processors according to migration step", func(annotation string, expectedProcessors []processing.ReconciliationProcessor) {
		// when
		apirule := &gatewayv2alpha1.APIRule{
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
				annotationProcessor{},
			},
		),
		Entry("should return AP and RA processors when current step is switchVsToService",
			string(applyIstioAuthorizationMigrationStep),
			[]processing.ReconciliationProcessor{
				processors2.VirtualServiceProcessor{},
				processors2.AuthorizationPolicyProcessor{},
				processors2.RequestAuthenticationProcessor{},
				annotationProcessor{},
			},
		),
		Entry("should return AP and RA processors when current step is removeOryRule",
			string(switchVsToService),
			[]processing.ReconciliationProcessor{
				accessRuleDeletionProcessor{},
				processors2.VirtualServiceProcessor{},
				processors2.AuthorizationPolicyProcessor{},
				processors2.RequestAuthenticationProcessor{},
				annotationProcessor{},
			},
		),
	)

})
