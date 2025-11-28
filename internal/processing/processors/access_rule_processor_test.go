package processors_test

import (
	"context"
	"fmt"
	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"

	"github.com/kyma-project/api-gateway/internal/builders"
	"github.com/kyma-project/api-gateway/internal/processing"
	. "github.com/kyma-project/api-gateway/internal/processing/processing_test"
	"github.com/kyma-project/api-gateway/internal/processing/processors"
	"github.com/kyma-project/api-gateway/internal/subresources/accessrule"
	rulev1alpha1 "github.com/kyma-project/api-gateway/internal/types/ory/oathkeeper-maester/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("Access Rule Processor", func() {
	It("should create access rule when no exists", func() {
		// given
		client := GetFakeClient()
		processor := processors.AccessRuleProcessor{
			ApiRule: &gatewayv1beta1.APIRule{},
			Creator: mockCreator{
				createMock: func() map[string]*rulev1alpha1.Rule {
					return map[string]*rulev1alpha1.Rule{
						"<http|https>://myService.myDomain.com<path>": builders.AccessRule().Spec(
							builders.AccessRuleSpec().Match(
								builders.Match().URL("<http|https>://myService.myDomain.com<path>"))).Get(),
					}
				},
			},
			Repository: accessrule.NewRepository(client),
		}

		// when
		result, err := processor.EvaluateReconciliation(context.Background(), client)

		// then
		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(1))
		Expect(result[0].Action.String()).To(Equal("create"))
	})

	It("should update access rule when path exists", func() {
		// given
		noop := []*gatewayv1beta1.Authenticator{
			{
				Handler: &gatewayv1beta1.Handler{
					Name: "noop",
				},
			},
		}

		noopRule := GetRuleFor("path", ApiMethods, []*gatewayv1beta1.Mutator{}, noop)
		rules := []gatewayv1beta1.Rule{noopRule}

		apiRule := GetAPIRuleFor(rules)

		rule := rulev1alpha1.Rule{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					processing.LegacyOwnerLabel: fmt.Sprintf("%s.%s", apiRule.Name, apiRule.Namespace),
				},
			},
			Spec: rulev1alpha1.RuleSpec{
				Match: &rulev1alpha1.Match{
					URL: fmt.Sprintf("<http|https>://%s<%s>", ServiceHost, "path"),
				},
			},
		}

		vs := networkingv1beta1.VirtualService{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					processing.LegacyOwnerLabel: fmt.Sprintf("%s.%s", apiRule.Name, apiRule.Namespace),
				},
			},
		}

		scheme := runtime.NewScheme()
		err := rulev1alpha1.AddToScheme(scheme)
		Expect(err).NotTo(HaveOccurred())
		err = networkingv1beta1.AddToScheme(scheme)
		Expect(err).NotTo(HaveOccurred())
		err = gatewayv1beta1.AddToScheme(scheme)
		Expect(err).NotTo(HaveOccurred())

		client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(&rule, &vs).Build()
		processor := processors.AccessRuleProcessor{
			ApiRule: apiRule,
			Creator: mockCreator{
				createMock: func() map[string]*rulev1alpha1.Rule {
					return map[string]*rulev1alpha1.Rule{
						"<http|https>://myService.myDomain.com<path>": builders.AccessRule().Spec(
							builders.AccessRuleSpec().Match(
								builders.Match().URL("<http|https>://myService.myDomain.com<path>"))).Get(),
					}
				},
			},
			Repository: accessrule.NewRepository(client),
		}

		// when
		result, err := processor.EvaluateReconciliation(context.Background(), client)

		// then
		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(1))
		Expect(result[0].Action.String()).To(Equal("update"))
	})

	It("should delete access rule", func() {
		// given
		noop := []*gatewayv1beta1.Authenticator{
			{
				Handler: &gatewayv1beta1.Handler{
					Name: "noop",
				},
			},
		}

		noopRule := GetRuleFor("same", ApiMethods, []*gatewayv1beta1.Mutator{}, noop)
		rules := []gatewayv1beta1.Rule{noopRule}

		apiRule := GetAPIRuleFor(rules)

		rule := rulev1alpha1.Rule{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					processing.LegacyOwnerLabel: fmt.Sprintf("%s.%s", apiRule.Name, apiRule.Namespace),
				},
			},
			Spec: rulev1alpha1.RuleSpec{
				Match: &rulev1alpha1.Match{
					URL: fmt.Sprintf("<http|https>://%s<%s>", ServiceHost, "path"),
				},
			},
		}

		vs := networkingv1beta1.VirtualService{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					processing.LegacyOwnerLabel: fmt.Sprintf("%s.%s", apiRule.Name, apiRule.Namespace),
				},
			},
		}

		scheme := runtime.NewScheme()
		err := rulev1alpha1.AddToScheme(scheme)
		Expect(err).NotTo(HaveOccurred())
		err = networkingv1beta1.AddToScheme(scheme)
		Expect(err).NotTo(HaveOccurred())
		err = gatewayv1beta1.AddToScheme(scheme)
		Expect(err).NotTo(HaveOccurred())

		client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(&rule, &vs).Build()
		processor := processors.AccessRuleProcessor{
			ApiRule: apiRule,
			Creator: mockCreator{
				createMock: func() map[string]*rulev1alpha1.Rule {
					return map[string]*rulev1alpha1.Rule{}
				},
			},
			Repository: accessrule.NewRepository(client),
		}

		// when
		result, err := processor.EvaluateReconciliation(context.Background(), client)

		// then
		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(1))
		Expect(result[0].Action.String()).To(Equal("delete"))
	})

	When("rule exists and rule path is different", func() {
		It("should create new rule and delete old rule", func() {
			// given
			noop := []*gatewayv1beta1.Authenticator{
				{
					Handler: &gatewayv1beta1.Handler{
						Name: "noop",
					},
				},
			}

			noopRule := GetRuleFor("newPath", ApiMethods, []*gatewayv1beta1.Mutator{}, noop)
			rules := []gatewayv1beta1.Rule{noopRule}

			apiRule := GetAPIRuleFor(rules)

			rule := rulev1alpha1.Rule{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						processing.LegacyOwnerLabel: fmt.Sprintf("%s.%s", apiRule.Name, apiRule.Namespace),
					},
				},
				Spec: rulev1alpha1.RuleSpec{
					Match: &rulev1alpha1.Match{
						URL: fmt.Sprintf("<http|https>://%s<%s>", ServiceHost, "oldPath"),
					},
				},
			}

			vs := networkingv1beta1.VirtualService{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						processing.LegacyOwnerLabel: fmt.Sprintf("%s.%s", apiRule.Name, apiRule.Namespace),
					},
				},
			}

			scheme := runtime.NewScheme()
			err := rulev1alpha1.AddToScheme(scheme)
			Expect(err).NotTo(HaveOccurred())
			err = networkingv1beta1.AddToScheme(scheme)
			Expect(err).NotTo(HaveOccurred())
			err = gatewayv1beta1.AddToScheme(scheme)
			Expect(err).NotTo(HaveOccurred())

			client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(&rule, &vs).Build()
			processor := processors.AccessRuleProcessor{
				ApiRule: apiRule,
				Creator: mockCreator{
					createMock: func() map[string]*rulev1alpha1.Rule {
						return map[string]*rulev1alpha1.Rule{
							"<http|https>://myService.myDomain.com<newPath>": builders.AccessRule().Spec(
								builders.AccessRuleSpec().Match(
									builders.Match().URL("<http|https>://myService.myDomain.com<newPath>"))).Get(),
						}
					},
				},
				Repository: accessrule.NewRepository(client),
			}

			// when
			result, err := processor.EvaluateReconciliation(context.Background(), client)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(2))

			createResultMatcher := PointTo(MatchFields(IgnoreExtras, Fields{
				"Action": WithTransform(actionToString, Equal("create")),
				"Obj": PointTo(MatchFields(IgnoreExtras, Fields{
					"Spec": MatchFields(IgnoreExtras, Fields{
						"Match": PointTo(MatchFields(IgnoreExtras, Fields{
							"URL": Equal("<http|https>://myService.myDomain.com<newPath>"),
						})),
					}),
				})),
			}))

			deleteResultMatcher := PointTo(MatchFields(IgnoreExtras, Fields{
				"Action": WithTransform(actionToString, Equal("delete")),
				"Obj": PointTo(MatchFields(IgnoreExtras, Fields{
					"Spec": MatchFields(IgnoreExtras, Fields{
						"Match": PointTo(MatchFields(IgnoreExtras, Fields{
							"URL": Equal("<http|https>://myService.myDomain.com<oldPath>"),
						})),
					}),
				})),
			}))

			Expect(result).To(ContainElements(createResultMatcher, deleteResultMatcher))
		})
	})
})

type mockCreator struct {
	createMock func() map[string]*rulev1alpha1.Rule
}

func (r mockCreator) Create(_ *gatewayv1beta1.APIRule) map[string]*rulev1alpha1.Rule {
	return r.createMock()
}

var actionToString = func(a processing.Action) string { return a.String() }
