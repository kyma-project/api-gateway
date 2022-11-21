package ory_test

import (
	"fmt"
	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	"github.com/kyma-incubator/api-gateway/internal/processing"
	. "github.com/kyma-incubator/api-gateway/internal/processing/internal/test"
	"github.com/kyma-incubator/api-gateway/internal/processing/ory"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	rulev1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/strings/slices"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("Access Rule Processor", func() {
	When("handler is allow", func() {

		It("should not create access rules", func() {
			// given
			strategies := []*gatewayv1beta1.Authenticator{
				{
					Handler: &gatewayv1beta1.Handler{
						Name: "allow",
					},
				},
			}

			allowRule := GetRuleFor(ApiPath, ApiMethods, []*gatewayv1beta1.Mutator{}, strategies)
			rules := []gatewayv1beta1.Rule{allowRule}

			apiRule := GetAPIRuleFor(rules)

			overrideServiceName := "testName"
			overrideServiceNamespace := "testName-namespace"
			overrideServicePort := uint32(8080)

			apiRule.Spec.Service = &gatewayv1beta1.Service{
				Name:      &overrideServiceName,
				Namespace: &overrideServiceNamespace,
				Port:      &overrideServicePort,
			}

			// when
			processor := ory.NewAccessRuleProcessor(GetConfigWithEmptyFakeClient())

			// then
			result, err := processor.EvaluateReconciliation(apiRule)

			Expect(err).To(BeNil())
			Expect(result).To(BeEmpty())
		})

	})

	When("handler is noop", func() {

		It("should override rule with meta data", func() {
			// given
			strategies := []*gatewayv1beta1.Authenticator{
				{
					Handler: &gatewayv1beta1.Handler{
						Name: "noop",
					},
				},
			}

			allowRule := GetRuleWithServiceFor(ApiPath, ApiMethods, []*gatewayv1beta1.Mutator{}, strategies, nil)
			rules := []gatewayv1beta1.Rule{allowRule}

			apiRule := GetAPIRuleFor(rules)

			// when
			processor := ory.NewAccessRuleProcessor(GetConfigWithEmptyFakeClient())

			// then
			result, err := processor.EvaluateReconciliation(apiRule)

			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(1))

			accessRule := result[0].Obj.(*rulev1alpha1.Rule)

			Expect(accessRule.ObjectMeta.Name).To(BeEmpty())
			Expect(accessRule.ObjectMeta.GenerateName).To(Equal(ApiName + "-"))
			Expect(accessRule.ObjectMeta.Namespace).To(Equal(ApiNamespace))
			Expect(accessRule.ObjectMeta.Labels[TestLabelKey]).To(Equal(TestLabelValue))

			Expect(accessRule.ObjectMeta.OwnerReferences[0].APIVersion).To(Equal(ApiAPIVersion))
			Expect(accessRule.ObjectMeta.OwnerReferences[0].Kind).To(Equal(ApiKind))
			Expect(accessRule.ObjectMeta.OwnerReferences[0].Name).To(Equal(ApiName))
			Expect(accessRule.ObjectMeta.OwnerReferences[0].UID).To(Equal(ApiUID))
		})

		It("should override rule upstream with rule level service", func() {
			// given
			strategies := []*gatewayv1beta1.Authenticator{
				{
					Handler: &gatewayv1beta1.Handler{
						Name: "noop",
					},
				},
			}

			overrideServiceName := "testName"
			overrideServicePort := uint32(8080)

			service := &gatewayv1beta1.Service{
				Name: &overrideServiceName,
				Port: &overrideServicePort,
			}

			allowRule := GetRuleWithServiceFor(ApiPath, ApiMethods, []*gatewayv1beta1.Mutator{}, strategies, service)
			rules := []gatewayv1beta1.Rule{allowRule}

			apiRule := GetAPIRuleFor(rules)

			// when
			processor := ory.NewAccessRuleProcessor(GetConfigWithEmptyFakeClient())

			// then
			result, err := processor.EvaluateReconciliation(apiRule)

			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(1))

			accessRule := result[0].Obj.(*rulev1alpha1.Rule)
			expectedRuleUpstreamURL := fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", overrideServiceName, ApiNamespace, overrideServicePort)
			Expect(accessRule.Spec.Upstream.URL).To(Equal(expectedRuleUpstreamURL))
		})

		It("should override rule upstream with rule level service for specified namespace", func() {
			// given
			strategies := []*gatewayv1beta1.Authenticator{
				{
					Handler: &gatewayv1beta1.Handler{
						Name: "noop",
					},
				},
			}

			overrideServiceName := "testName"
			overrideServiceNamespace := "testName-namespace"
			overrideServicePort := uint32(8080)

			service := &gatewayv1beta1.Service{
				Name:      &overrideServiceName,
				Namespace: &overrideServiceNamespace,
				Port:      &overrideServicePort,
			}

			allowRule := GetRuleWithServiceFor(ApiPath, ApiMethods, []*gatewayv1beta1.Mutator{}, strategies, service)
			rules := []gatewayv1beta1.Rule{allowRule}

			apiRule := GetAPIRuleFor(rules)

			// when
			processor := ory.NewAccessRuleProcessor(GetConfigWithEmptyFakeClient())

			// then
			result, err := processor.EvaluateReconciliation(apiRule)

			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(1))

			accessRule := result[0].Obj.(*rulev1alpha1.Rule)
			expectedRuleUpstreamURL := fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", overrideServiceName, overrideServiceNamespace, overrideServicePort)
			Expect(accessRule.Spec.Upstream.URL).To(Equal(expectedRuleUpstreamURL))
		})

		It("should return rule with default domain name when the hostname does not contain domain name", func() {
			strategies := []*gatewayv1beta1.Authenticator{
				{
					Handler: &gatewayv1beta1.Handler{
						Name: "noop",
					},
				},
			}

			allowRule := GetRuleFor(ApiPath, ApiMethods, []*gatewayv1beta1.Mutator{}, strategies)
			rules := []gatewayv1beta1.Rule{allowRule}

			apiRule := GetAPIRuleFor(rules)
			apiRule.Spec.Host = &ServiceHostWithNoDomain

			processor := ory.NewAccessRuleProcessor(GetConfigWithEmptyFakeClient())

			// when
			result, err := processor.EvaluateReconciliation(apiRule)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(1))

			expectedRuleMatchURL := fmt.Sprintf("<http|https>://%s<%s>", ServiceHost, ApiPath)

			accessRule := result[0].Obj.(*rulev1alpha1.Rule)
			Expect(accessRule.Spec.Match.URL).To(Equal(expectedRuleMatchURL))
		})

		Context("when existing rule has owner v1alpha1 owner label", func() {
			It("should get and update match methods of rule", func() {
				// given
				noop := []*gatewayv1beta1.Authenticator{
					{
						Handler: &gatewayv1beta1.Handler{
							Name: "noop",
						},
					},
				}

				noopRule := GetRuleFor(ApiPath, ApiMethods, []*gatewayv1beta1.Mutator{}, noop)
				rules := []gatewayv1beta1.Rule{noopRule}

				apiRule := GetAPIRuleFor(rules)

				rule := rulev1alpha1.Rule{

					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							processing.OwnerLabelv1alpha1: fmt.Sprintf("%s.%s", apiRule.ObjectMeta.Name, apiRule.ObjectMeta.Namespace),
						},
					},
					Spec: rulev1alpha1.RuleSpec{
						Match: &rulev1alpha1.Match{
							URL:     fmt.Sprintf("<http|https>://%s<%s>", ServiceHost, ApiPath),
							Methods: []string{"DELETE"},
						},
					},
				}

				vs := networkingv1beta1.VirtualService{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							processing.OwnerLabelv1alpha1: fmt.Sprintf("%s.%s", apiRule.ObjectMeta.Name, apiRule.ObjectMeta.Namespace),
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

				processor := ory.NewAccessRuleProcessor(GetConfigWithClient(client))

				// when
				result, err := processor.EvaluateReconciliation(apiRule)

				// then
				Expect(err).To(BeNil())
				Expect(result).To(HaveLen(1))
				Expect(result[0].Action).To(Equal("update"))

				accessRule := result[0].Obj.(*rulev1alpha1.Rule)
				Expect(accessRule.Spec.Match.Methods).To(Equal([]string{"GET"}))
			})
		})

		When("rule exists and and rule path is different", func() {
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
							processing.OwnerLabelv1alpha1: fmt.Sprintf("%s.%s", apiRule.ObjectMeta.Name, apiRule.ObjectMeta.Namespace),
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
							processing.OwnerLabelv1alpha1: fmt.Sprintf("%s.%s", apiRule.ObjectMeta.Name, apiRule.ObjectMeta.Namespace),
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

				processor := ory.NewAccessRuleProcessor(GetConfigWithClient(client))

				// when
				result, err := processor.EvaluateReconciliation(apiRule)

				// then
				Expect(err).To(BeNil())
				Expect(result).To(HaveLen(2))
				Expect(result[0].Action).To(Equal("create"))
				createdRule := result[0].Obj.(*rulev1alpha1.Rule)
				Expect(createdRule.Spec.Match.URL).To(Equal("<http|https>://myService.myDomain.com<newPath>"))

				Expect(result[1].Action).To(Equal("delete"))
				deletedRule := result[1].Obj.(*rulev1alpha1.Rule)
				Expect(deletedRule.Spec.Match.URL).To(Equal("<http|https>://myService.myDomain.com<oldPath>"))
			})
		})

	})

	When("multiple handler", func() {

		It("should return two rules for given paths", func() {
			// given
			noop := []*gatewayv1beta1.Authenticator{
				{
					Handler: &gatewayv1beta1.Handler{
						Name: "noop",
					},
				},
			}

			jwtConfigJSON := fmt.Sprintf(`
						{
							"trusted_issuers": ["%s"],
							"jwks": [],
							"required_scope": [%s]
					}`, JwtIssuer, ToCSVList(ApiScopes))

			jwt := []*gatewayv1beta1.Authenticator{
				{
					Handler: &gatewayv1beta1.Handler{
						Name: "jwt",
						Config: &runtime.RawExtension{
							Raw: []byte(jwtConfigJSON),
						},
					},
				},
			}

			testMutators := []*gatewayv1beta1.Mutator{
				{
					Handler: &gatewayv1beta1.Handler{
						Name: "noop",
					},
				},
				{
					Handler: &gatewayv1beta1.Handler{
						Name: "idtoken",
					},
				},
			}

			noopRule := GetRuleFor(ApiPath, ApiMethods, []*gatewayv1beta1.Mutator{}, noop)
			jwtRule := GetRuleFor(HeadersApiPath, ApiMethods, testMutators, jwt)
			rules := []gatewayv1beta1.Rule{noopRule, jwtRule}

			apiRule := GetAPIRuleFor(rules)

			// when
			processor := ory.NewAccessRuleProcessor(GetConfigWithEmptyFakeClient())

			// then
			result, err := processor.EvaluateReconciliation(apiRule)

			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(2))

			expectedNoopRuleMatchURL := fmt.Sprintf("<http|https>://%s<%s>", ServiceHost, ApiPath)
			expectedJwtRuleMatchURL := fmt.Sprintf("<http|https>://%s<%s>", ServiceHost, HeadersApiPath)
			expectedRuleUpstreamURL := fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", ServiceName, ApiNamespace, ServicePort)

			var jwtAccessRule *rulev1alpha1.Rule
			var noopAccessRule *rulev1alpha1.Rule

			for _, change := range result {
				rule := change.Obj.(*rulev1alpha1.Rule)
				switch rule.Spec.Authenticators[0].Handler.Name {
				case "noop":
					noopAccessRule = rule
				case "jwt":
					jwtAccessRule = rule
				default:
					Fail("Rule is not expected.")
				}
			}

			Expect(noopAccessRule).NotTo(BeNil())
			Expect(len(noopAccessRule.Spec.Authenticators)).To(Equal(1))
			Expect(noopAccessRule.Spec.Authorizer.Name).To(Equal("allow"))
			Expect(noopAccessRule.Spec.Authorizer.Config).To(BeNil())
			Expect(noopAccessRule.Spec.Authenticators[0].Handler.Name).To(Equal("noop"))
			Expect(noopAccessRule.Spec.Authenticators[0].Handler.Config).To(BeNil())
			Expect(len(noopAccessRule.Spec.Match.Methods)).To(Equal(len(ApiMethods)))
			Expect(noopAccessRule.Spec.Match.Methods).To(Equal(ApiMethods))
			Expect(noopAccessRule.Spec.Match.URL).To(Equal(expectedNoopRuleMatchURL))
			Expect(noopAccessRule.Spec.Upstream.URL).To(Equal(expectedRuleUpstreamURL))

			Expect(jwtAccessRule).NotTo(BeNil())
			Expect(len(jwtAccessRule.Spec.Authenticators)).To(Equal(1))
			Expect(jwtAccessRule.Spec.Authorizer.Name).To(Equal("allow"))
			Expect(jwtAccessRule.Spec.Authorizer.Config).To(BeNil())
			Expect(jwtAccessRule.Spec.Authenticators[0].Handler.Name).To(Equal("jwt"))
			Expect(jwtAccessRule.Spec.Authenticators[0].Handler.Config).NotTo(BeNil())
			Expect(string(jwtAccessRule.Spec.Authenticators[0].Handler.Config.Raw)).To(Equal(jwtConfigJSON))
			Expect(len(jwtAccessRule.Spec.Match.Methods)).To(Equal(len(ApiMethods)))
			Expect(jwtAccessRule.Spec.Match.Methods).To(Equal(ApiMethods))
			Expect(jwtAccessRule.Spec.Match.URL).To(Equal(expectedJwtRuleMatchURL))
			Expect(jwtAccessRule.Spec.Upstream.URL).To(Equal(expectedRuleUpstreamURL))
			Expect(jwtAccessRule.Spec.Mutators).NotTo(BeNil())
			Expect(len(jwtAccessRule.Spec.Mutators)).To(Equal(len(testMutators)))
			Expect(jwtAccessRule.Spec.Mutators[0].Handler.Name).To(Equal(testMutators[0].Name))
			Expect(jwtAccessRule.Spec.Mutators[1].Handler.Name).To(Equal(testMutators[1].Name))
		})

		It("should return two rules for two same paths and different methods", func() {
			// given
			noop := []*gatewayv1beta1.Authenticator{
				{
					Handler: &gatewayv1beta1.Handler{
						Name: "noop",
					},
				},
			}

			jwtConfigJSON := fmt.Sprintf(`
						{
							"trusted_issuers": ["%s"],
							"jwks": [],
							"required_scope": [%s]
					}`, JwtIssuer, ToCSVList(ApiScopes))

			jwt := []*gatewayv1beta1.Authenticator{
				{
					Handler: &gatewayv1beta1.Handler{
						Name: "jwt",
						Config: &runtime.RawExtension{
							Raw: []byte(jwtConfigJSON),
						},
					},
				},
			}

			testMutators := []*gatewayv1beta1.Mutator{
				{
					Handler: &gatewayv1beta1.Handler{
						Name: "noop",
					},
				},
				{
					Handler: &gatewayv1beta1.Handler{
						Name: "idtoken",
					},
				},
			}
			getMethod := []string{"GET"}
			postMethod := []string{"POST"}
			noopRule := GetRuleFor(ApiPath, getMethod, []*gatewayv1beta1.Mutator{}, noop)
			jwtRule := GetRuleFor(ApiPath, postMethod, testMutators, jwt)
			rules := []gatewayv1beta1.Rule{noopRule, jwtRule}

			apiRule := GetAPIRuleFor(rules)

			// when
			processor := ory.NewAccessRuleProcessor(GetConfigWithEmptyFakeClient())

			// then
			result, err := processor.EvaluateReconciliation(apiRule)

			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(2))

			expectedNoopRuleMatchURL := fmt.Sprintf("<http|https>://%s<%s>", ServiceHost, ApiPath)
			expectedJwtRuleMatchURL := fmt.Sprintf("<http|https>://%s<%s>", ServiceHost, ApiPath)
			expectedRuleUpstreamURL := fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", ServiceName, ApiNamespace, ServicePort)

			var jwtAccessRule *rulev1alpha1.Rule
			var noopAccessRule *rulev1alpha1.Rule

			for _, change := range result {
				rule := change.Obj.(*rulev1alpha1.Rule)
				switch rule.Spec.Authenticators[0].Handler.Name {
				case "noop":
					noopAccessRule = rule
				case "jwt":
					jwtAccessRule = rule
				default:
					Fail("Rule is not expected.")
				}
			}

			Expect(noopAccessRule).NotTo(BeNil())
			Expect(len(noopAccessRule.Spec.Authenticators)).To(Equal(1))
			Expect(noopAccessRule.Spec.Authorizer.Name).To(Equal("allow"))
			Expect(noopAccessRule.Spec.Authorizer.Config).To(BeNil())
			Expect(noopAccessRule.Spec.Authenticators[0].Handler.Name).To(Equal("noop"))
			Expect(noopAccessRule.Spec.Authenticators[0].Handler.Config).To(BeNil())
			Expect(len(noopAccessRule.Spec.Match.Methods)).To(Equal(len(getMethod)))
			Expect(noopAccessRule.Spec.Match.Methods).To(Equal(getMethod))
			Expect(noopAccessRule.Spec.Match.URL).To(Equal(expectedNoopRuleMatchURL))
			Expect(noopAccessRule.Spec.Upstream.URL).To(Equal(expectedRuleUpstreamURL))

			Expect(jwtAccessRule).NotTo(BeNil())
			Expect(len(jwtAccessRule.Spec.Authenticators)).To(Equal(1))
			Expect(jwtAccessRule.Spec.Authorizer.Name).To(Equal("allow"))
			Expect(jwtAccessRule.Spec.Authorizer.Config).To(BeNil())
			Expect(jwtAccessRule.Spec.Authenticators[0].Handler.Name).To(Equal("jwt"))
			Expect(jwtAccessRule.Spec.Authenticators[0].Handler.Config).NotTo(BeNil())
			Expect(string(jwtAccessRule.Spec.Authenticators[0].Handler.Config.Raw)).To(Equal(jwtConfigJSON))
			Expect(len(jwtAccessRule.Spec.Match.Methods)).To(Equal(len(postMethod)))
			Expect(jwtAccessRule.Spec.Match.Methods).To(Equal(postMethod))
			Expect(jwtAccessRule.Spec.Match.URL).To(Equal(expectedJwtRuleMatchURL))
			Expect(jwtAccessRule.Spec.Upstream.URL).To(Equal(expectedRuleUpstreamURL))
			Expect(jwtAccessRule.Spec.Mutators).NotTo(BeNil())
			Expect(len(jwtAccessRule.Spec.Mutators)).To(Equal(len(testMutators)))
			Expect(jwtAccessRule.Spec.Mutators[0].Handler.Name).To(Equal(testMutators[0].Name))
			Expect(jwtAccessRule.Spec.Mutators[1].Handler.Name).To(Equal(testMutators[1].Name))

		})

		It("should return two rules for two same paths and one different", func() {
			// given
			noop := []*gatewayv1beta1.Authenticator{
				{
					Handler: &gatewayv1beta1.Handler{
						Name: "noop",
					},
				},
			}

			jwtConfigJSON := fmt.Sprintf(`
						{
							"trusted_issuers": ["%s"],
							"jwks": [],
							"required_scope": [%s]
					}`, JwtIssuer, ToCSVList(ApiScopes))

			jwt := []*gatewayv1beta1.Authenticator{
				{
					Handler: &gatewayv1beta1.Handler{
						Name: "jwt",
						Config: &runtime.RawExtension{
							Raw: []byte(jwtConfigJSON),
						},
					},
				},
			}

			testMutators := []*gatewayv1beta1.Mutator{
				{
					Handler: &gatewayv1beta1.Handler{
						Name: "noop",
					},
				},
				{
					Handler: &gatewayv1beta1.Handler{
						Name: "idtoken",
					},
				},
			}
			getMethod := []string{"GET"}
			postMethod := []string{"POST"}
			noopGetRule := GetRuleFor(ApiPath, getMethod, []*gatewayv1beta1.Mutator{}, noop)
			noopPostRule := GetRuleFor(ApiPath, postMethod, []*gatewayv1beta1.Mutator{}, noop)
			jwtRule := GetRuleFor(HeadersApiPath, ApiMethods, testMutators, jwt)
			rules := []gatewayv1beta1.Rule{noopGetRule, noopPostRule, jwtRule}

			apiRule := GetAPIRuleFor(rules)

			// when
			processor := ory.NewAccessRuleProcessor(GetConfigWithEmptyFakeClient())

			// then
			result, err := processor.EvaluateReconciliation(apiRule)

			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(3))

			expectedNoopGetRuleMatchURL := fmt.Sprintf("<http|https>://%s<%s>", ServiceHost, ApiPath)
			expectedNoopPostRuleMatchURL := fmt.Sprintf("<http|https>://%s<%s>", ServiceHost, ApiPath)
			expectedJwtRuleMatchURL := fmt.Sprintf("<http|https>://%s<%s>", ServiceHost, HeadersApiPath)
			expectedRuleUpstreamURL := fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", ServiceName, ApiNamespace, ServicePort)

			var jwtAccessRule *rulev1alpha1.Rule
			var noopPostAccessRule *rulev1alpha1.Rule
			var noopGetAccessRule *rulev1alpha1.Rule

			for _, change := range result {
				rule := change.Obj.(*rulev1alpha1.Rule)
				switch rule.Spec.Authenticators[0].Handler.Name {
				case "noop":
					if slices.Contains(rule.Spec.Match.Methods, "GET") {
						noopGetAccessRule = rule
					} else {
						noopPostAccessRule = rule
					}
				case "jwt":
					jwtAccessRule = rule

				default:
					Fail("Rule is not expected.")
				}
			}

			Expect(noopGetAccessRule).NotTo(BeNil())
			Expect(noopGetAccessRule.Spec.Authorizer.Name).To(Equal("allow"))
			Expect(noopGetAccessRule.Spec.Authorizer.Config).To(BeNil())
			Expect(noopGetAccessRule.Spec.Authenticators[0].Handler.Name).To(Equal("noop"))
			Expect(noopGetAccessRule.Spec.Authenticators[0].Handler.Config).To(BeNil())
			Expect(len(noopGetAccessRule.Spec.Match.Methods)).To(Equal(len(getMethod)))
			Expect(noopGetAccessRule.Spec.Match.Methods).To(Equal(getMethod))
			Expect(noopGetAccessRule.Spec.Match.URL).To(Equal(expectedNoopGetRuleMatchURL))
			Expect(noopGetAccessRule.Spec.Upstream.URL).To(Equal(expectedRuleUpstreamURL))

			Expect(noopPostAccessRule).NotTo(BeNil())
			Expect(len(noopPostAccessRule.Spec.Authenticators)).To(Equal(1))
			Expect(noopPostAccessRule.Spec.Authorizer.Name).To(Equal("allow"))
			Expect(noopPostAccessRule.Spec.Authorizer.Config).To(BeNil())
			Expect(noopPostAccessRule.Spec.Authenticators[0].Handler.Name).To(Equal("noop"))
			Expect(noopPostAccessRule.Spec.Authenticators[0].Handler.Config).To(BeNil())
			Expect(len(noopPostAccessRule.Spec.Match.Methods)).To(Equal(len(postMethod)))
			Expect(noopPostAccessRule.Spec.Match.Methods).To(Equal(postMethod))
			Expect(noopPostAccessRule.Spec.Match.URL).To(Equal(expectedNoopPostRuleMatchURL))
			Expect(noopPostAccessRule.Spec.Upstream.URL).To(Equal(expectedRuleUpstreamURL))

			Expect(jwtAccessRule).NotTo(BeNil())
			Expect(len(jwtAccessRule.Spec.Authenticators)).To(Equal(1))
			Expect(jwtAccessRule.Spec.Authorizer.Name).To(Equal("allow"))
			Expect(jwtAccessRule.Spec.Authorizer.Config).To(BeNil())
			Expect(jwtAccessRule.Spec.Authenticators[0].Handler.Name).To(Equal("jwt"))
			Expect(jwtAccessRule.Spec.Authenticators[0].Handler.Config).NotTo(BeNil())
			Expect(string(jwtAccessRule.Spec.Authenticators[0].Handler.Config.Raw)).To(Equal(jwtConfigJSON))
			Expect(len(jwtAccessRule.Spec.Match.Methods)).To(Equal(len(ApiMethods)))
			Expect(jwtAccessRule.Spec.Match.Methods).To(Equal(ApiMethods))
			Expect(jwtAccessRule.Spec.Match.URL).To(Equal(expectedJwtRuleMatchURL))
			Expect(jwtAccessRule.Spec.Upstream.URL).To(Equal(expectedRuleUpstreamURL))
			Expect(jwtAccessRule.Spec.Mutators).NotTo(BeNil())
			Expect(len(jwtAccessRule.Spec.Mutators)).To(Equal(len(testMutators)))
			Expect(jwtAccessRule.Spec.Mutators[0].Handler.Name).To(Equal(testMutators[0].Name))
			Expect(jwtAccessRule.Spec.Mutators[1].Handler.Name).To(Equal(testMutators[1].Name))

		})

		It("should return rule for jwt & oauth authenticators for given path", func() {
			// given
			oauthConfigJSON := fmt.Sprintf(`{"required_scope": [%s]}`, ToCSVList(ApiScopes))

			jwtConfigJSON := fmt.Sprintf(`
						{
							"trusted_issuers": ["%s"],
							"jwks": [],
							"required_scope": [%s]
					}`, JwtIssuer, ToCSVList(ApiScopes))

			jwt := &gatewayv1beta1.Authenticator{
				Handler: &gatewayv1beta1.Handler{
					Name: "jwt",
					Config: &runtime.RawExtension{
						Raw: []byte(jwtConfigJSON),
					},
				},
			}
			oauth := &gatewayv1beta1.Authenticator{
				Handler: &gatewayv1beta1.Handler{
					Name: "oauth2_introspection",
					Config: &runtime.RawExtension{
						Raw: []byte(oauthConfigJSON),
					},
				},
			}

			strategies := []*gatewayv1beta1.Authenticator{jwt, oauth}

			allowRule := GetRuleFor(ApiPath, ApiMethods, []*gatewayv1beta1.Mutator{}, strategies)
			rules := []gatewayv1beta1.Rule{allowRule}

			apiRule := GetAPIRuleFor(rules)

			// when
			processor := ory.NewAccessRuleProcessor(GetConfigWithEmptyFakeClient())

			// then
			result, err := processor.EvaluateReconciliation(apiRule)

			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(1))

			rule := result[0].Obj.(*rulev1alpha1.Rule)

			Expect(len(rule.Spec.Authenticators)).To(Equal(2))

			Expect(rule.Spec.Authorizer.Name).To(Equal("allow"))
			Expect(rule.Spec.Authorizer.Config).To(BeNil())

			Expect(rule.Spec.Authenticators[0].Handler.Name).To(Equal("jwt"))
			Expect(rule.Spec.Authenticators[0].Handler.Config).NotTo(BeNil())
			Expect(string(rule.Spec.Authenticators[0].Handler.Config.Raw)).To(Equal(jwtConfigJSON))

			Expect(rule.Spec.Authenticators[1].Handler.Name).To(Equal("oauth2_introspection"))
			Expect(rule.Spec.Authenticators[1].Handler.Config).NotTo(BeNil())
			Expect(string(rule.Spec.Authenticators[1].Handler.Config.Raw)).To(Equal(oauthConfigJSON))

			expectedRuleMatchURL := fmt.Sprintf("<http|https>://%s<%s>", ServiceHost, ApiPath)
			expectedRuleUpstreamURL := fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", ServiceName, ApiNamespace, ServicePort)

			Expect(len(rule.Spec.Match.Methods)).To(Equal(len(ApiMethods)))
			Expect(rule.Spec.Match.Methods).To(Equal(ApiMethods))
			Expect(rule.Spec.Match.URL).To(Equal(expectedRuleMatchURL))

			Expect(rule.Spec.Upstream.URL).To(Equal(expectedRuleUpstreamURL))
		})
	})
})
