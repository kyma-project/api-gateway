package builders

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	k8sTypes "k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Builder for", func() {

	path := "/headers"
	methods := []string{"GET", "POST"}

	Describe("AccessRule", func() {
		It("should build the object", func() {
			name := "testName"
			namespace := "testNs"

			refName := "refName"
			refVersion := "v1alpha1"
			refKind := "APIRule"
			var refUID k8sTypes.UID = "123"

			testMatchLabelsKey := "app"
			testMatchLabelsValue := "httpbin"
			testMatchLabels := map[string]string{testMatchLabelsKey: testMatchLabelsValue}
			testRulesSourceRequestPrincipals := "*"

			ap := AuthorizationPolicyBuilder().GenerateName(name).Namespace(namespace).
				Owner(OwnerReference().Name(refName).APIVersion(refVersion).Kind(refKind).UID(refUID).Controller(true)).
				Spec(AuthorizationPolicySpecBuilder().
					Selector(SelectorBuilder().
						MatchLabels(testMatchLabelsKey, testMatchLabelsValue)).
					Rule(RuleBuilder().
						RuleFrom(RuleFromBuilder().
							Source()).
						RuleTo(RuleToBuilder().
							Operation(OperationBuilder().
								Path(path).
								Methods(methods))))).
				Get()

			Expect(ap.Name).To(BeEmpty())
			Expect(ap.GenerateName).To(Equal(name))
			Expect(ap.Namespace).To(Equal(namespace))
			Expect(ap.OwnerReferences).To(HaveLen(1))
			Expect(ap.OwnerReferences[0].Name).To(Equal(refName))
			Expect(ap.OwnerReferences[0].APIVersion).To(Equal(refVersion))
			Expect(ap.OwnerReferences[0].Kind).To(Equal(refKind))
			Expect(ap.OwnerReferences[0].UID).To(BeEquivalentTo(refUID))
			Expect(ap.Spec.Selector.MatchLabels).To(Equal(testMatchLabels))
			Expect(ap.Spec.Rules[0].From[0].Source.RequestPrincipals).To(Equal(testRulesSourceRequestPrincipals))
			Expect(ap.Spec.Rules[0].To[0].Operation.Paths[0]).To(Equal(path))
			Expect(ap.Spec.Rules[0].To[0].Operation.Methods).To(BeEquivalentTo(methods))
		})
	})
})
