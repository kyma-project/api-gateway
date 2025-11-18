package cleaner_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"github.com/kyma-project/api-gateway/internal/processing/cleaner"
	rulev1alpha1 "github.com/kyma-project/api-gateway/internal/types/ory/oathkeeper-maester/api/v1alpha1"
)

var _ = Describe("Cleaner", func() {
	var (
		k8sClient client.Client
		ctx       context.Context
		apiRule   *gatewayv1beta1.APIRule
	)

	BeforeEach(func() {
		ctx = context.Background()

		apiRule = &gatewayv1beta1.APIRule{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-apirule",
				Namespace: "test-namespace",
			},
			Spec: gatewayv1beta1.APIRuleSpec{
				Host: ptr("example.com"),
			},
		}

		scheme := runtime.NewScheme()
		Expect(networkingv1beta1.AddToScheme(scheme)).To(Succeed())
		Expect(rulev1alpha1.AddToScheme(scheme)).To(Succeed())
		Expect(securityv1beta1.AddToScheme(scheme)).To(Succeed())
		Expect(gatewayv1beta1.AddToScheme(scheme)).To(Succeed())
		Expect(apiextensionsv1.AddToScheme(scheme)).To(Succeed())

		k8sClient = fake.NewClientBuilder().WithScheme(scheme).Build()

		// Create Ory CRD to enable AccessRule deletion
		oryCrd := &apiextensionsv1.CustomResourceDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name: "rules.oathkeeper.ory.sh",
			},
		}
		Expect(k8sClient.Create(ctx, oryCrd)).To(Succeed())
	})

	Describe("DeleteAPIRuleSubresources", func() {
		Context("when APIRule has subresources with legacy labels", func() {
			BeforeEach(func() {
				// Create subresources with legacy labels
				vs := &networkingv1beta1.VirtualService{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-vs-legacy",
						Namespace: "test-namespace",
						Labels: map[string]string{
							"apirule.gateway.kyma-project.io/v1beta1": "test-apirule.test-namespace",
						},
					},
				}
				ap := &securityv1beta1.AuthorizationPolicy{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-ap-legacy",
						Namespace: "test-namespace",
						Labels: map[string]string{
							"apirule.gateway.kyma-project.io/v1beta1": "test-apirule.test-namespace",
						},
					},
				}
				ra := &securityv1beta1.RequestAuthentication{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-ra-legacy",
						Namespace: "test-namespace",
						Labels: map[string]string{
							"apirule.gateway.kyma-project.io/v1beta1": "test-apirule.test-namespace",
						},
					},
				}
				rule := &rulev1alpha1.Rule{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-rule-legacy",
						Namespace: "test-namespace",
						Labels: map[string]string{
							"apirule.gateway.kyma-project.io/v1beta1": "test-apirule.test-namespace",
						},
					},
				}

				Expect(k8sClient.Create(ctx, vs)).To(Succeed())
				Expect(k8sClient.Create(ctx, ap)).To(Succeed())
				Expect(k8sClient.Create(ctx, ra)).To(Succeed())
				Expect(k8sClient.Create(ctx, rule)).To(Succeed())
			})

			It("should delete all subresources with legacy labels", func() {
				// When
				err := cleaner.DeleteAPIRuleSubresources(k8sClient, ctx, apiRule)

				// Then
				Expect(err).NotTo(HaveOccurred())

				// Verify all resources are deleted
				var vsList networkingv1beta1.VirtualServiceList
				Expect(k8sClient.List(ctx, &vsList)).To(Succeed())
				Expect(vsList.Items).To(BeEmpty())

				var apList securityv1beta1.AuthorizationPolicyList
				Expect(k8sClient.List(ctx, &apList)).To(Succeed())
				Expect(apList.Items).To(BeEmpty())

				var raList securityv1beta1.RequestAuthenticationList
				Expect(k8sClient.List(ctx, &raList)).To(Succeed())
				Expect(raList.Items).To(BeEmpty())

				var ruleList rulev1alpha1.RuleList
				Expect(k8sClient.List(ctx, &ruleList)).To(Succeed())
				Expect(ruleList.Items).To(BeEmpty())
			})
		})

		Context("when APIRule has subresources with new labels", func() {
			BeforeEach(func() {
				// Create subresources with new labels
				vs := &networkingv1beta1.VirtualService{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-vs-new",
						Namespace: "test-namespace",
						Labels: map[string]string{
							"apirule.gateway.kyma-project.io/name":      "test-apirule",
							"apirule.gateway.kyma-project.io/namespace": "test-namespace",
						},
					},
				}
				ap := &securityv1beta1.AuthorizationPolicy{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-ap-new",
						Namespace: "test-namespace",
						Labels: map[string]string{
							"apirule.gateway.kyma-project.io/name":      "test-apirule",
							"apirule.gateway.kyma-project.io/namespace": "test-namespace",
						},
					},
				}
				ra := &securityv1beta1.RequestAuthentication{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-ra-new",
						Namespace: "test-namespace",
						Labels: map[string]string{
							"apirule.gateway.kyma-project.io/name":      "test-apirule",
							"apirule.gateway.kyma-project.io/namespace": "test-namespace",
						},
					},
				}
				rule := &rulev1alpha1.Rule{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-rule-new",
						Namespace: "test-namespace",
						Labels: map[string]string{
							"apirule.gateway.kyma-project.io/name":      "test-apirule",
							"apirule.gateway.kyma-project.io/namespace": "test-namespace",
						},
					},
				}

				Expect(k8sClient.Create(ctx, vs)).To(Succeed())
				Expect(k8sClient.Create(ctx, ap)).To(Succeed())
				Expect(k8sClient.Create(ctx, ra)).To(Succeed())
				Expect(k8sClient.Create(ctx, rule)).To(Succeed())
			})

			It("should delete all subresources with new labels", func() {
				// When
				err := cleaner.DeleteAPIRuleSubresources(k8sClient, ctx, apiRule)

				// Then
				Expect(err).NotTo(HaveOccurred())

				// Verify all resources are deleted
				var vsList networkingv1beta1.VirtualServiceList
				Expect(k8sClient.List(ctx, &vsList)).To(Succeed())
				Expect(vsList.Items).To(BeEmpty())

				var apList securityv1beta1.AuthorizationPolicyList
				Expect(k8sClient.List(ctx, &apList)).To(Succeed())
				Expect(apList.Items).To(BeEmpty())

				var raList securityv1beta1.RequestAuthenticationList
				Expect(k8sClient.List(ctx, &raList)).To(Succeed())
				Expect(raList.Items).To(BeEmpty())

				var ruleList rulev1alpha1.RuleList
				Expect(k8sClient.List(ctx, &ruleList)).To(Succeed())
				Expect(ruleList.Items).To(BeEmpty())
			})
		})

		Context("when APIRule has subresources but other APIRules also have subresources", func() {
			BeforeEach(func() {
				// Create subresources for test-apirule with legacy labels
				vs1 := &networkingv1beta1.VirtualService{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-apirule-vs",
						Namespace: "test-namespace",
						Labels: map[string]string{
							"apirule.gateway.kyma-project.io/v1beta1": "test-apirule.test-namespace",
						},
					},
				}
				ap1 := &securityv1beta1.AuthorizationPolicy{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-apirule-ap",
						Namespace: "test-namespace",
						Labels: map[string]string{
							"apirule.gateway.kyma-project.io/v1beta1": "test-apirule.test-namespace",
						},
					},
				}
				ra1 := &securityv1beta1.RequestAuthentication{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-apirule-ra",
						Namespace: "test-namespace",
						Labels: map[string]string{
							"apirule.gateway.kyma-project.io/v1beta1": "test-apirule.test-namespace",
						},
					},
				}
				rule1 := &rulev1alpha1.Rule{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-apirule-rule",
						Namespace: "test-namespace",
						Labels: map[string]string{
							"apirule.gateway.kyma-project.io/v1beta1": "test-apirule.test-namespace",
						},
					},
				}

				// Create subresources for other-apirule
				vs2 := &networkingv1beta1.VirtualService{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "other-apirule-vs",
						Namespace: "test-namespace",
						Labels: map[string]string{
							"apirule.gateway.kyma-project.io/v1beta1": "other-apirule.test-namespace",
						},
					},
				}
				ap2 := &securityv1beta1.AuthorizationPolicy{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "other-apirule-ap",
						Namespace: "test-namespace",
						Labels: map[string]string{
							"apirule.gateway.kyma-project.io/v1beta1": "other-apirule.test-namespace",
						},
					},
				}
				ra2 := &securityv1beta1.RequestAuthentication{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "other-apirule-ra",
						Namespace: "test-namespace",
						Labels: map[string]string{
							"apirule.gateway.kyma-project.io/v1beta1": "other-apirule.test-namespace",
						},
					},
				}
				rule2 := &rulev1alpha1.Rule{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "other-apirule-rule",
						Namespace: "test-namespace",
						Labels: map[string]string{
							"apirule.gateway.kyma-project.io/v1beta1": "other-apirule.test-namespace",
						},
					},
				}

				Expect(k8sClient.Create(ctx, vs1)).To(Succeed())
				Expect(k8sClient.Create(ctx, ap1)).To(Succeed())
				Expect(k8sClient.Create(ctx, ra1)).To(Succeed())
				Expect(k8sClient.Create(ctx, rule1)).To(Succeed())

				Expect(k8sClient.Create(ctx, vs2)).To(Succeed())
				Expect(k8sClient.Create(ctx, ap2)).To(Succeed())
				Expect(k8sClient.Create(ctx, ra2)).To(Succeed())
				Expect(k8sClient.Create(ctx, rule2)).To(Succeed())
			})

			It("should delete only the subresources belonging to the specified APIRule", func() {
				// When
				err := cleaner.DeleteAPIRuleSubresources(k8sClient, ctx, apiRule)

				// Then
				Expect(err).NotTo(HaveOccurred())

				// Verify only test-apirule resources are deleted, other-apirule resources remain
				var vsList networkingv1beta1.VirtualServiceList
				Expect(k8sClient.List(ctx, &vsList)).To(Succeed())
				Expect(vsList.Items).To(HaveLen(1))
				Expect(vsList.Items[0].Name).To(Equal("other-apirule-vs"))

				var apList securityv1beta1.AuthorizationPolicyList
				Expect(k8sClient.List(ctx, &apList)).To(Succeed())
				Expect(apList.Items).To(HaveLen(1))
				Expect(apList.Items[0].Name).To(Equal("other-apirule-ap"))

				var raList securityv1beta1.RequestAuthenticationList
				Expect(k8sClient.List(ctx, &raList)).To(Succeed())
				Expect(raList.Items).To(HaveLen(1))
				Expect(raList.Items[0].Name).To(Equal("other-apirule-ra"))

				var ruleList rulev1alpha1.RuleList
				Expect(k8sClient.List(ctx, &ruleList)).To(Succeed())
				Expect(ruleList.Items).To(HaveLen(1))
				Expect(ruleList.Items[0].Name).To(Equal("other-apirule-rule"))
			})
		})

		Context("when Ory CRD does not exist", func() {
			BeforeEach(func() {
				// Delete the Ory CRD
				oryCrd := &apiextensionsv1.CustomResourceDefinition{
					ObjectMeta: metav1.ObjectMeta{
						Name: "rules.oathkeeper.ory.sh",
					},
				}
				Expect(k8sClient.Delete(ctx, oryCrd)).To(Succeed())

				// Create other subresources
				vs := &networkingv1beta1.VirtualService{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-vs",
						Namespace: "test-namespace",
						Labels: map[string]string{
							"apirule.gateway.kyma-project.io/name":      "test-apirule",
							"apirule.gateway.kyma-project.io/namespace": "test-namespace",
						},
					},
				}
				Expect(k8sClient.Create(ctx, vs)).To(Succeed())
			})

			It("should delete other subresources but skip AccessRules without error", func() {
				// When
				err := cleaner.DeleteAPIRuleSubresources(k8sClient, ctx, apiRule)

				// Then
				Expect(err).NotTo(HaveOccurred())

				// Verify VirtualService is deleted
				var vsList networkingv1beta1.VirtualServiceList
				Expect(k8sClient.List(ctx, &vsList)).To(Succeed())
				Expect(vsList.Items).To(BeEmpty())
			})
		})

		Context("when APIRule has no subresources", func() {
			It("should complete without error", func() {
				// When
				err := cleaner.DeleteAPIRuleSubresources(k8sClient, ctx, apiRule)

				// Then
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})

func ptr(s string) *string {
	return &s
}
