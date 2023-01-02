package istio_test

import (
	"context"
	"fmt"

	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	"github.com/kyma-incubator/api-gateway/internal/processing"
	. "github.com/kyma-incubator/api-gateway/internal/processing/internal/test"
	"github.com/kyma-incubator/api-gateway/internal/processing/istio"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"golang.org/x/exp/slices"
	"istio.io/api/security/v1beta1"
	typev1beta1 "istio.io/api/type/v1beta1"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var _ = Describe("Authorization Policy Processor", func() {
	createIstioJwtAccessStrategy := func() *gatewayv1beta1.Authenticator {
		jwtConfigJSON := fmt.Sprintf(`{
			"authentications": [{"issuer": "%s", "jwksUri": "%s"}]}`, JwtIssuer, JwksUri)
		return &gatewayv1beta1.Authenticator{
			Handler: &gatewayv1beta1.Handler{
				Name: "jwt",
				Config: &runtime.RawExtension{
					Raw: []byte(jwtConfigJSON),
				},
			},
		}
	}

	It("should produce two APs for a rule with one issuer and two paths", func() {
		// given
		jwt := createIstioJwtAccessStrategy()
		service := &gatewayv1beta1.Service{
			Name: &ServiceName,
			Port: &ServicePort,
		}

		ruleJwt := GetRuleWithServiceFor(HeadersApiPath, ApiMethods, []*gatewayv1beta1.Mutator{}, []*gatewayv1beta1.Authenticator{jwt}, service)
		ruleJwt2 := GetRuleWithServiceFor(ImgApiPath, ApiMethods, []*gatewayv1beta1.Mutator{}, []*gatewayv1beta1.Authenticator{jwt}, service)
		apiRule := GetAPIRuleFor([]gatewayv1beta1.Rule{ruleJwt, ruleJwt2})
		client := GetFakeClient()
		processor := istio.NewAuthorizationPolicyProcessor(GetTestConfig())

		// when
		result, err := processor.EvaluateReconciliation(context.TODO(), client, apiRule)

		// then
		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(2))

		ap1 := result[0].Obj.(*securityv1beta1.AuthorizationPolicy)
		ap2 := result[1].Obj.(*securityv1beta1.AuthorizationPolicy)

		Expect(ap1).NotTo(BeNil())
		Expect(ap1.ObjectMeta.Name).To(BeEmpty())
		Expect(ap1.ObjectMeta.GenerateName).To(Equal(ApiName + "-"))
		Expect(ap1.ObjectMeta.Namespace).To(Equal(ApiNamespace))
		Expect(ap1.ObjectMeta.Labels[TestLabelKey]).To(Equal(TestLabelValue))

		Expect(ap1.Spec.Selector.MatchLabels[TestSelectorKey]).NotTo(BeNil())
		Expect(ap1.Spec.Selector.MatchLabels[TestSelectorKey]).To(Equal(ServiceName))
		Expect(len(ap1.Spec.Rules)).To(Equal(1))
		Expect(len(ap1.Spec.Rules[0].From)).To(Equal(1))
		Expect(len(ap1.Spec.Rules[0].From[0].Source.RequestPrincipals)).To(Equal(1))
		Expect(ap1.Spec.Rules[0].From[0].Source.RequestPrincipals[0]).To(Equal("*"))
		Expect(len(ap1.Spec.Rules[0].To)).To(Equal(1))
		Expect(len(ap1.Spec.Rules[0].To[0].Operation.Methods)).To(Equal(1))
		Expect(ap1.Spec.Rules[0].To[0].Operation.Methods).To(ContainElements(ApiMethods))
		Expect(len(ap1.Spec.Rules[0].To[0].Operation.Paths)).To(Equal(1))

		Expect(len(ap1.OwnerReferences)).To(Equal(1))
		Expect(ap1.OwnerReferences[0].APIVersion).To(Equal(ApiAPIVersion))
		Expect(ap1.OwnerReferences[0].Kind).To(Equal(ApiKind))
		Expect(ap1.OwnerReferences[0].Name).To(Equal(ApiName))
		Expect(ap1.OwnerReferences[0].UID).To(Equal(ApiUID))

		Expect(ap2).NotTo(BeNil())
		Expect(ap2.ObjectMeta.Name).To(BeEmpty())
		Expect(ap2.ObjectMeta.GenerateName).To(Equal(ApiName + "-"))
		Expect(ap2.ObjectMeta.Namespace).To(Equal(ApiNamespace))
		Expect(ap2.ObjectMeta.Labels[TestLabelKey]).To(Equal(TestLabelValue))

		Expect(ap2.Spec.Selector.MatchLabels[TestSelectorKey]).NotTo(BeNil())
		Expect(ap2.Spec.Selector.MatchLabels[TestSelectorKey]).To(Equal(ServiceName))
		Expect(len(ap2.Spec.Rules)).To(Equal(1))
		Expect(len(ap2.Spec.Rules[0].From)).To(Equal(1))
		Expect(len(ap2.Spec.Rules[0].From[0].Source.RequestPrincipals)).To(Equal(1))
		Expect(ap2.Spec.Rules[0].From[0].Source.RequestPrincipals[0]).To(Equal("*"))
		Expect(len(ap2.Spec.Rules[0].To)).To(Equal(1))
		Expect(len(ap2.Spec.Rules[0].To[0].Operation.Methods)).To(Equal(1))
		Expect(ap2.Spec.Rules[0].To[0].Operation.Methods).To(ContainElements(ApiMethods))
		Expect(len(ap2.Spec.Rules[0].To[0].Operation.Paths)).To(Equal(1))

		Expect(ap2.OwnerReferences[0].APIVersion).To(Equal(ApiAPIVersion))
		Expect(ap2.OwnerReferences[0].Kind).To(Equal(ApiKind))
		Expect(ap2.OwnerReferences[0].Name).To(Equal(ApiName))
		Expect(ap2.OwnerReferences[0].UID).To(Equal(ApiUID))
	})

	It("should produce one AP for a Rule without service, but service definition on ApiRule level", func() {
		// given
		jwt := createIstioJwtAccessStrategy()
		client := GetFakeClient()
		ruleJwt := GetRuleFor(HeadersApiPath, ApiMethods, []*gatewayv1beta1.Mutator{}, []*gatewayv1beta1.Authenticator{jwt})
		apiRule := GetAPIRuleFor([]gatewayv1beta1.Rule{ruleJwt})
		processor := istio.NewAuthorizationPolicyProcessor(GetTestConfig())

		// when
		result, err := processor.EvaluateReconciliation(context.TODO(), client, apiRule)

		// then
		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(1))

		ap := result[0].Obj.(*securityv1beta1.AuthorizationPolicy)
		Expect(ap).NotTo(BeNil())
		Expect(ap.Spec.Selector.MatchLabels[TestSelectorKey]).To(Equal(ServiceName))
	})

	It("should produce one AP for a Rule with service with configured namespace, in the configured namespace", func() {
		// given
		jwt := createIstioJwtAccessStrategy()
		ruleServiceName := "rule-scope-example-service"
		namespace := "other-namespace"
		service := &gatewayv1beta1.Service{
			Name:      &ruleServiceName,
			Port:      &ServicePort,
			Namespace: &namespace,
		}
		client := GetFakeClient()
		ruleJwt := GetRuleWithServiceFor(HeadersApiPath, ApiMethods, []*gatewayv1beta1.Mutator{}, []*gatewayv1beta1.Authenticator{jwt}, service)
		apiRule := GetAPIRuleFor([]gatewayv1beta1.Rule{ruleJwt})
		processor := istio.NewAuthorizationPolicyProcessor(GetTestConfig())

		// when
		result, err := processor.EvaluateReconciliation(context.TODO(), client, apiRule)

		// then
		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(1))

		ap := result[0].Obj.(*securityv1beta1.AuthorizationPolicy)
		Expect(ap).NotTo(BeNil())
		Expect(ap.Spec.Selector.MatchLabels[TestSelectorKey]).To(Equal(ruleServiceName))
		// The AP should be in .Service.Namespace
		Expect(ap.Namespace).To(Equal(namespace))
		// And the OwnerLabel should point to APIRule namespace
		Expect(ap.Labels[processing.OwnerLabel]).ToNot(BeEmpty())
		Expect(ap.Labels[processing.OwnerLabel]).To(Equal(fmt.Sprintf("%s.%s",apiRule.Name,apiRule.Namespace)))
	})

	It("should produce AP with service from Rule, when service is configured on Rule and ApiRule level", func() {
		// given
		jwt := createIstioJwtAccessStrategy()
		ruleServiceName := "rule-scope-example-service"
		service := &gatewayv1beta1.Service{
			Name: &ruleServiceName,
			Port: &ServicePort,
		}
		client := GetFakeClient()
		ruleJwt := GetRuleWithServiceFor(HeadersApiPath, ApiMethods, []*gatewayv1beta1.Mutator{}, []*gatewayv1beta1.Authenticator{jwt}, service)
		apiRule := GetAPIRuleFor([]gatewayv1beta1.Rule{ruleJwt})
		processor := istio.NewAuthorizationPolicyProcessor(GetTestConfig())

		// when
		result, err := processor.EvaluateReconciliation(context.TODO(), client, apiRule)

		// then
		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(1))

		ap := result[0].Obj.(*securityv1beta1.AuthorizationPolicy)
		Expect(ap).NotTo(BeNil())
		Expect(ap.Spec.Selector.MatchLabels[TestSelectorKey]).To(Equal(ruleServiceName))
	})

	It("should produce AP from a rule with two issuers and one path", func() {
		jwtConfigJSON := fmt.Sprintf(`{
			"authentications": [{"issuer": "%s", "jwksUri": "%s"}, {"issuer": "%s", "jwksUri": "%s"}]
			}`, JwtIssuer, JwksUri, JwtIssuer2, JwksUri2)
		jwt := &gatewayv1beta1.Authenticator{
			Handler: &gatewayv1beta1.Handler{
				Name: "jwt",
				Config: &runtime.RawExtension{
					Raw: []byte(jwtConfigJSON),
				},
			},
		}
		client := GetFakeClient()
		service := &gatewayv1beta1.Service{
			Name: &ServiceName,
			Port: &ServicePort,
		}
		ruleJwt := GetRuleWithServiceFor(HeadersApiPath, ApiMethods, []*gatewayv1beta1.Mutator{}, []*gatewayv1beta1.Authenticator{jwt}, service)
		apiRule := GetAPIRuleFor([]gatewayv1beta1.Rule{ruleJwt})
		processor := istio.NewAuthorizationPolicyProcessor(GetTestConfig())

		// when
		result, err := processor.EvaluateReconciliation(context.TODO(), client, apiRule)

		// then
		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(1))

		ap := result[0].Obj.(*securityv1beta1.AuthorizationPolicy)

		Expect(ap).NotTo(BeNil())
		Expect(ap.ObjectMeta.Name).To(BeEmpty())
		Expect(ap.ObjectMeta.GenerateName).To(Equal(ApiName + "-"))
		Expect(ap.ObjectMeta.Namespace).To(Equal(ApiNamespace))
		Expect(ap.ObjectMeta.Labels[TestLabelKey]).To(Equal(TestLabelValue))

		Expect(ap.Spec.Selector.MatchLabels[TestSelectorKey]).NotTo(BeNil())
		Expect(ap.Spec.Selector.MatchLabels[TestSelectorKey]).To(Equal(ServiceName))
		Expect(len(ap.Spec.Rules)).To(Equal(1))
		Expect(len(ap.Spec.Rules[0].From)).To(Equal(1))
		Expect(len(ap.Spec.Rules[0].From[0].Source.RequestPrincipals)).To(Equal(1))
		Expect(ap.Spec.Rules[0].From[0].Source.RequestPrincipals[0]).To(Equal("*"))
		Expect(len(ap.Spec.Rules[0].To)).To(Equal(1))
		Expect(len(ap.Spec.Rules[0].To[0].Operation.Methods)).To(Equal(1))
		Expect(ap.Spec.Rules[0].To[0].Operation.Methods).To(ContainElements(ApiMethods))
		Expect(len(ap.Spec.Rules[0].To[0].Operation.Paths)).To(Equal(1))
		Expect(ap.Spec.Rules[0].To[0].Operation.Paths).To(ContainElements(HeadersApiPath))

		Expect(ap.OwnerReferences[0].APIVersion).To(Equal(ApiAPIVersion))
		Expect(ap.OwnerReferences[0].Kind).To(Equal(ApiKind))
		Expect(ap.OwnerReferences[0].Name).To(Equal(ApiName))
		Expect(ap.OwnerReferences[0].UID).To(Equal(ApiUID))
	})

	When("single handler only", func() {

		It("should create AP with From in Rules Spec for jwt", func() {
			// given
			strategies := []*gatewayv1beta1.Authenticator{
				{
					Handler: &gatewayv1beta1.Handler{
						Name: "jwt",
					},
				},
			}

			rule := GetRuleFor(ApiPath, ApiMethods, []*gatewayv1beta1.Mutator{}, strategies)
			rules := []gatewayv1beta1.Rule{rule}

			apiRule := GetAPIRuleFor(rules)

			overrideServiceName := "testName"
			overrideServiceNamespace := "testName-namespace"
			overrideServicePort := uint32(8080)

			apiRule.Spec.Service = &gatewayv1beta1.Service{
				Name:      &overrideServiceName,
				Namespace: &overrideServiceNamespace,
				Port:      &overrideServicePort,
			}

			client := GetFakeClient()
			processor := istio.NewAuthorizationPolicyProcessor(GetTestConfig())

			// when
			result, err := processor.EvaluateReconciliation(context.TODO(), client, apiRule)

			// then
			Expect(err).To(BeNil())
			Expect(len(result)).To(Equal(1))
			ap := result[0].Obj.(*securityv1beta1.AuthorizationPolicy)
			Expect(len(result)).To(Equal(1))
			Expect(ap.Spec.Rules[0].From).NotTo(BeEmpty())
		})

		It("should not create AP for allow", func() {
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

			client := GetFakeClient()
			processor := istio.NewAuthorizationPolicyProcessor(GetTestConfig())

			// when
			result, err := processor.EvaluateReconciliation(context.TODO(), client, apiRule)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(BeEmpty())
		})

		It("should not create AP for noop", func() {
			// given
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

			overrideServiceName := "testName"
			overrideServiceNamespace := "testName-namespace"
			overrideServicePort := uint32(8080)

			apiRule.Spec.Service = &gatewayv1beta1.Service{
				Name:      &overrideServiceName,
				Namespace: &overrideServiceNamespace,
				Port:      &overrideServicePort,
			}

			client := GetFakeClient()
			processor := istio.NewAuthorizationPolicyProcessor(GetTestConfig())

			// when
			result, err := processor.EvaluateReconciliation(context.TODO(), client, apiRule)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(BeEmpty())
		})
	})

	When("additional handler to JWT", func() {
		It("should create AP for allow with From having Source.Principals == cluster.local/ns/istio-system/sa/istio-ingressgateway-service-account", func() {
			// given
			jwt := createIstioJwtAccessStrategy()
			allow := &gatewayv1beta1.Authenticator{
				Handler: &gatewayv1beta1.Handler{
					Name: "allow",
				},
			}

			service := &gatewayv1beta1.Service{
				Name: &ServiceName,
				Port: &ServicePort,
			}

			ruleAllow := GetRuleWithServiceFor(HeadersApiPath, ApiMethods, []*gatewayv1beta1.Mutator{}, []*gatewayv1beta1.Authenticator{allow}, service)
			ruleJwt := GetRuleWithServiceFor(ImgApiPath, ApiMethods, []*gatewayv1beta1.Mutator{}, []*gatewayv1beta1.Authenticator{jwt}, service)
			apiRule := GetAPIRuleFor([]gatewayv1beta1.Rule{ruleAllow, ruleJwt})
			client := GetFakeClient()
			processor := istio.NewAuthorizationPolicyProcessor(GetTestConfig())

			// when
			results, err := processor.EvaluateReconciliation(context.TODO(), client, apiRule)

			// then
			Expect(err).To(BeNil())
			Expect(results).To(HaveLen(2))

			for _, result := range results {
				ap := result.Obj.(*securityv1beta1.AuthorizationPolicy)

				Expect(ap).NotTo(BeNil())
				Expect(len(ap.Spec.Rules)).To(Equal(1))
				Expect(len(ap.Spec.Rules[0].To)).To(Equal(1))
				Expect(len(ap.Spec.Rules[0].To[0].Operation.Paths)).To(Equal(1))

				expectedHandlers := []string{HeadersApiPath, ImgApiPath}
				Expect(slices.Contains(expectedHandlers, ap.Spec.Rules[0].To[0].Operation.Paths[0])).To(BeTrue())

				switch ap.Spec.Rules[0].To[0].Operation.Paths[0] {
				case HeadersApiPath:
					Expect(len(ap.Spec.Rules[0].From)).To(Equal(1))
					Expect(ap.Spec.Rules[0].From[0].Source.Principals[0]).To(Equal("cluster.local/ns/istio-system/sa/istio-ingressgateway-service-account"))
				case ImgApiPath:
					Expect(len(ap.Spec.Rules[0].From)).To(Equal(1))
				}
			}
		})

		It("should create AP for noop with From spec having Source.Principals == cluster.local/ns/kyma-system/sa/oathkeeper-maester-account", func() {
			// given
			jwt := createIstioJwtAccessStrategy()
			noop := &gatewayv1beta1.Authenticator{
				Handler: &gatewayv1beta1.Handler{
					Name: "noop",
				},
			}

			service := &gatewayv1beta1.Service{
				Name: &ServiceName,
				Port: &ServicePort,
			}

			ruleNoop := GetRuleWithServiceFor(HeadersApiPath, ApiMethods, []*gatewayv1beta1.Mutator{}, []*gatewayv1beta1.Authenticator{noop}, service)
			ruleJwt := GetRuleWithServiceFor(ImgApiPath, ApiMethods, []*gatewayv1beta1.Mutator{}, []*gatewayv1beta1.Authenticator{jwt}, service)
			apiRule := GetAPIRuleFor([]gatewayv1beta1.Rule{ruleNoop, ruleJwt})
			client := GetFakeClient()
			processor := istio.NewAuthorizationPolicyProcessor(GetTestConfig())

			// when
			results, err := processor.EvaluateReconciliation(context.TODO(), client, apiRule)

			// then
			Expect(err).To(BeNil())
			Expect(results).To(HaveLen(2))

			for _, result := range results {
				ap := result.Obj.(*securityv1beta1.AuthorizationPolicy)

				Expect(ap).NotTo(BeNil())
				Expect(len(ap.Spec.Rules)).To(Equal(1))
				Expect(len(ap.Spec.Rules[0].To)).To(Equal(1))
				Expect(len(ap.Spec.Rules[0].To[0].Operation.Paths)).To(Equal(1))

				expectedHandlers := []string{HeadersApiPath, ImgApiPath}
				Expect(slices.Contains(expectedHandlers, ap.Spec.Rules[0].To[0].Operation.Paths[0])).To(BeTrue())

				switch ap.Spec.Rules[0].To[0].Operation.Paths[0] {
				case HeadersApiPath:
					Expect(len(ap.Spec.Rules[0].From)).To(Equal(1))
					Expect(ap.Spec.Rules[0].From[0].Source.Principals[0]).To(Equal("cluster.local/ns/kyma-system/sa/oathkeeper-maester-account"))
				case ImgApiPath:
					Expect(len(ap.Spec.Rules[0].From)).To(Equal(1))
				}
			}
		})
	})

	It("should create AP when no exists", func() {
		// given: New resources
		methods := []string{"GET"}
		path := "/"
		serviceName := "test-service"

		rule := getRuleForApTest(methods, path, serviceName)
		rules := []gatewayv1beta1.Rule{rule}

		apiRule := GetAPIRuleFor(rules)

		processor := istio.NewAuthorizationPolicyProcessor(GetTestConfig())

		// when
		result, err := processor.EvaluateReconciliation(context.TODO(), GetFakeClient(), apiRule)

		// then
		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(1))
		Expect(result[0].Action.String()).To(Equal("create"))
	})

	It("should update existing AP when path, methods and service name didn't change", func() {
		// given: Cluster state
		existingAp := securityv1beta1.AuthorizationPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					processing.OwnerLabelv1alpha1: fmt.Sprintf("%s.%s", ApiName, ApiNamespace),
				},
			},
			Spec: v1beta1.AuthorizationPolicy{
				Selector: &typev1beta1.WorkloadSelector{
					MatchLabels: map[string]string{
						"app": "test-service",
					},
				},
				Rules: []*v1beta1.Rule{
					{
						To: []*v1beta1.Rule_To{
							{
								Operation: &v1beta1.Operation{
									Methods: []string{"GET", "POST"},
									Paths:   []string{"/"},
								},
							},
						},
					},
				},
			},
		}

		ctrlClient := GetFakeClient(&existingAp)
		processor := istio.NewAuthorizationPolicyProcessor(GetTestConfig())

		// given: New resources
		methods := []string{"GET", "POST"}
		path := "/"
		serviceName := "test-service"

		rule := getRuleForApTest(methods, path, serviceName)
		rules := []gatewayv1beta1.Rule{rule}

		apiRule := GetAPIRuleFor(rules)

		// when
		result, err := processor.EvaluateReconciliation(context.TODO(), ctrlClient, apiRule)

		// then
		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(1))

		resultMatcher := PointTo(MatchFields(IgnoreExtras, Fields{
			"Action": WithTransform(ActionToString, Equal("update")),
			"Obj": PointTo(MatchFields(IgnoreExtras, Fields{
				"Spec": MatchFields(IgnoreExtras, Fields{
					"Selector": PointTo(MatchFields(IgnoreExtras, Fields{
						"MatchLabels": ContainElement("test-service"),
					})),
					"Rules": ContainElements(
						PointTo(MatchFields(IgnoreExtras, Fields{
							"To": ContainElements(
								PointTo(MatchFields(IgnoreExtras, Fields{
									"Operation": PointTo(MatchFields(IgnoreExtras, Fields{
										"Methods": ContainElements("GET", "POST"),
										"Paths":   ContainElements("/"),
									})),
								})),
							),
						})),
					),
				}),
			})),
		}))

		Expect(result).To(ContainElements(resultMatcher))
	})

	When("Two AP for different services with JWT handler exist", func() {
		It("should update existing AP when handler changed for one of the AP to noop", func() {
			// given: Cluster state
			beingUpdatedAp := securityv1beta1.AuthorizationPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name: "being-updated-ap",
					Labels: map[string]string{
						processing.OwnerLabelv1alpha1: fmt.Sprintf("%s.%s", ApiName, ApiNamespace),
					},
				},
				Spec: v1beta1.AuthorizationPolicy{
					Selector: &typev1beta1.WorkloadSelector{
						MatchLabels: map[string]string{
							"app": "test-service",
						},
					},
					Rules: []*v1beta1.Rule{
						{
							From: []*v1beta1.Rule_From{
								{
									Source: &v1beta1.Source{
										RequestPrincipals: []string{"*"},
									},
								},
							},
							To: []*v1beta1.Rule_To{
								{
									Operation: &v1beta1.Operation{
										Methods: []string{"GET", "POST"},
										Paths:   []string{"/"},
									},
								},
							},
						},
					},
				},
			}

			jwtSecuredAp := securityv1beta1.AuthorizationPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name: "jwt-secured-ap",
					Labels: map[string]string{
						processing.OwnerLabelv1alpha1: fmt.Sprintf("%s.%s", ApiName, ApiNamespace),
					},
				},
				Spec: v1beta1.AuthorizationPolicy{
					Selector: &typev1beta1.WorkloadSelector{
						MatchLabels: map[string]string{
							"app": "jwt-secured-service",
						},
					},
					Rules: []*v1beta1.Rule{
						{
							From: []*v1beta1.Rule_From{
								{
									Source: &v1beta1.Source{
										RequestPrincipals: []string{"*"},
									},
								},
							},
							To: []*v1beta1.Rule_To{
								{
									Operation: &v1beta1.Operation{
										Methods: []string{"GET", "POST"},
										Paths:   []string{"/"},
									},
								},
							},
						},
					},
				},
			}

			ctrlClient := GetFakeClient(&beingUpdatedAp, &jwtSecuredAp)
			processor := istio.NewAuthorizationPolicyProcessor(GetTestConfig())

			// given: New resources
			jwtRule := getRuleForApTest([]string{"GET", "POST"}, "/", "jwt-secured-service")

			strategies := []*gatewayv1beta1.Authenticator{
				{
					Handler: &gatewayv1beta1.Handler{
						Name: "noop",
					},
				},
			}

			serviceName := "test-service"
			port := uint32(8080)
			service := &gatewayv1beta1.Service{
				Name: &serviceName,
				Port: &port,
			}

			rule := GetRuleWithServiceFor("/", []string{"GET", "POST"}, []*gatewayv1beta1.Mutator{}, strategies, service)

			rules := []gatewayv1beta1.Rule{rule, jwtRule}

			apiRule := GetAPIRuleFor(rules)

			// when
			result, err := processor.EvaluateReconciliation(context.TODO(), ctrlClient, apiRule)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(2))

			updatedToNoopMatcher := PointTo(MatchFields(IgnoreExtras, Fields{
				"Action": WithTransform(ActionToString, Equal("update")),
				"Obj": PointTo(MatchFields(IgnoreExtras, Fields{
					"Spec": MatchFields(IgnoreExtras, Fields{
						"Selector": PointTo(MatchFields(IgnoreExtras, Fields{
							"MatchLabels": ContainElement("test-service"),
						})),
						"Rules": ContainElements(
							PointTo(MatchFields(IgnoreExtras, Fields{
								"From": ContainElement(
									PointTo(MatchFields(IgnoreExtras, Fields{
										"Source": PointTo(MatchFields(IgnoreExtras, Fields{
											"Principals": ContainElements("cluster.local/ns/kyma-system/sa/oathkeeper-maester-account"),
										})),
									})),
								),
								"To": ContainElements(
									PointTo(MatchFields(IgnoreExtras, Fields{
										"Operation": PointTo(MatchFields(IgnoreExtras, Fields{
											"Methods": ContainElements("GET", "POST"),
											"Paths":   ContainElements("/"),
										})),
									})),
								),
							})),
						),
					}),
				})),
			}))

			notChangedMatcher := PointTo(MatchFields(IgnoreExtras, Fields{
				"Action": WithTransform(ActionToString, Equal("update")),
				"Obj": PointTo(MatchFields(IgnoreExtras, Fields{
					"Spec": MatchFields(IgnoreExtras, Fields{
						"Selector": PointTo(MatchFields(IgnoreExtras, Fields{
							"MatchLabels": ContainElement("jwt-secured-service"),
						})),
						"Rules": ContainElements(
							PointTo(MatchFields(IgnoreExtras, Fields{
								"From": ContainElement(
									PointTo(MatchFields(IgnoreExtras, Fields{
										"Source": PointTo(MatchFields(IgnoreExtras, Fields{
											"RequestPrincipals": ContainElements("*"),
										})),
									})),
								),
								"To": ContainElements(
									PointTo(MatchFields(IgnoreExtras, Fields{
										"Operation": PointTo(MatchFields(IgnoreExtras, Fields{
											"Methods": ContainElements("GET", "POST"),
											"Paths":   ContainElements("/"),
										})),
									})),
								),
							})),
						),
					}),
				})),
			}))

			Expect(result).To(ContainElements(updatedToNoopMatcher, notChangedMatcher))
		})

	})
	It("should delete AP when there is no desired AP", func() {
		//given: Cluster state
		existingAp := securityv1beta1.AuthorizationPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					processing.OwnerLabelv1alpha1: fmt.Sprintf("%s.%s", ApiName, ApiNamespace),
				},
			},
			Spec: v1beta1.AuthorizationPolicy{
				Selector: &typev1beta1.WorkloadSelector{
					MatchLabels: map[string]string{
						"app": "test-service",
					},
				},
				Rules: []*v1beta1.Rule{
					{
						To: []*v1beta1.Rule_To{
							{
								Operation: &v1beta1.Operation{
									Methods: []string{"GET", "POST"},
									Paths:   []string{"/"},
								},
							},
						},
					},
				},
			},
		}

		ctrlClient := GetFakeClient(&existingAp)
		processor := istio.NewAuthorizationPolicyProcessor(GetTestConfig())

		// given: New resources
		apiRule := GetAPIRuleFor([]gatewayv1beta1.Rule{})

		// when
		result, err := processor.EvaluateReconciliation(context.TODO(), ctrlClient, apiRule)

		// then
		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(1))

		resultMatcher := PointTo(MatchFields(IgnoreExtras, Fields{
			"Action": WithTransform(ActionToString, Equal("delete")),
			"Obj": PointTo(MatchFields(IgnoreExtras, Fields{
				"Spec": MatchFields(IgnoreExtras, Fields{
					"Selector": PointTo(MatchFields(IgnoreExtras, Fields{
						"MatchLabels": ContainElement("test-service"),
					})),
					"Rules": ContainElements(
						PointTo(MatchFields(IgnoreExtras, Fields{
							"To": ContainElements(
								PointTo(MatchFields(IgnoreExtras, Fields{
									"Operation": PointTo(MatchFields(IgnoreExtras, Fields{
										"Methods": ContainElements("GET", "POST"),
										"Paths":   ContainElements("/"),
									})),
								})),
							),
						})),
					),
				}),
			})),
		}))

		Expect(result).To(ContainElements(resultMatcher))
	})

	When("AP with RuleTo exists", func() {
		It("should create new AP when new rule with same methods and service but different path is added to ApiRule", func() {
			// given: Cluster state
			existingAp := securityv1beta1.AuthorizationPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						processing.OwnerLabelv1alpha1: fmt.Sprintf("%s.%s", ApiName, ApiNamespace),
					},
				},
				Spec: v1beta1.AuthorizationPolicy{
					Selector: &typev1beta1.WorkloadSelector{
						MatchLabels: map[string]string{
							"app": "test-service",
						},
					},
					Rules: []*v1beta1.Rule{
						{
							To: []*v1beta1.Rule_To{
								{
									Operation: &v1beta1.Operation{
										Methods: []string{"GET", "POST"},
										Paths:   []string{"/"},
									},
								},
							},
						},
					},
				},
			}

			ctrlClient := GetFakeClient(&existingAp)
			processor := istio.NewAuthorizationPolicyProcessor(GetTestConfig())

			// given: New resources

			existingRule := getRuleForApTest([]string{"GET", "POST"}, "/", "test-service")
			newRule := getRuleForApTest([]string{"GET", "POST"}, "/new-path", "test-service")
			rules := []gatewayv1beta1.Rule{existingRule, newRule}

			apiRule := GetAPIRuleFor(rules)

			// when
			result, err := processor.EvaluateReconciliation(context.TODO(), ctrlClient, apiRule)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(2))

			existingApMatcher := PointTo(MatchFields(IgnoreExtras, Fields{
				"Action": WithTransform(ActionToString, Equal("update")),
				"Obj": PointTo(MatchFields(IgnoreExtras, Fields{
					"Spec": MatchFields(IgnoreExtras, Fields{
						"Selector": PointTo(MatchFields(IgnoreExtras, Fields{
							"MatchLabels": ContainElement("test-service"),
						})),
						"Rules": ContainElements(
							PointTo(MatchFields(IgnoreExtras, Fields{
								"To": ContainElements(
									PointTo(MatchFields(IgnoreExtras, Fields{
										"Operation": PointTo(MatchFields(IgnoreExtras, Fields{
											"Methods": ContainElements("GET", "POST"),
											"Paths":   ContainElements("/"),
										})),
									})),
								),
							})),
						),
					}),
				})),
			}))

			newApMatcher := PointTo(MatchFields(IgnoreExtras, Fields{
				"Action": WithTransform(ActionToString, Equal("create")),
				"Obj": PointTo(MatchFields(IgnoreExtras, Fields{
					"Spec": MatchFields(IgnoreExtras, Fields{
						"Selector": PointTo(MatchFields(IgnoreExtras, Fields{
							"MatchLabels": ContainElement("test-service"),
						})),
						"Rules": ContainElements(
							PointTo(MatchFields(IgnoreExtras, Fields{
								"To": ContainElements(
									PointTo(MatchFields(IgnoreExtras, Fields{
										"Operation": PointTo(MatchFields(IgnoreExtras, Fields{
											"Methods": ContainElements("GET", "POST"),
											"Paths":   ContainElements("/new-path"),
										})),
									})),
								),
							})),
						),
					}),
				})),
			}))

			Expect(result).To(ContainElements(existingApMatcher, newApMatcher))
		})

		It("should create new AP when new rule with same path and service but different methods is added to ApiRule", func() {
			// given: Cluster state
			existingAp := securityv1beta1.AuthorizationPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						processing.OwnerLabelv1alpha1: fmt.Sprintf("%s.%s", ApiName, ApiNamespace),
					},
				},
				Spec: v1beta1.AuthorizationPolicy{
					Selector: &typev1beta1.WorkloadSelector{
						MatchLabels: map[string]string{
							"app": "test-service",
						},
					},
					Rules: []*v1beta1.Rule{
						{
							To: []*v1beta1.Rule_To{
								{
									Operation: &v1beta1.Operation{
										Methods: []string{"GET", "POST"},
										Paths:   []string{"/"},
									},
								},
							},
						},
					},
				},
			}

			ctrlClient := GetFakeClient(&existingAp)
			processor := istio.NewAuthorizationPolicyProcessor(GetTestConfig())

			// given: New resources

			existingRule := getRuleForApTest([]string{"GET", "POST"}, "/", "test-service")
			newRule := getRuleForApTest([]string{"DELETE"}, "/", "test-service")
			rules := []gatewayv1beta1.Rule{existingRule, newRule}

			apiRule := GetAPIRuleFor(rules)

			// when
			result, err := processor.EvaluateReconciliation(context.TODO(), ctrlClient, apiRule)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(2))

			existingApMatcher := PointTo(MatchFields(IgnoreExtras, Fields{
				"Action": WithTransform(ActionToString, Equal("update")),
				"Obj": PointTo(MatchFields(IgnoreExtras, Fields{
					"Spec": MatchFields(IgnoreExtras, Fields{
						"Selector": PointTo(MatchFields(IgnoreExtras, Fields{
							"MatchLabels": ContainElement("test-service"),
						})),
						"Rules": ContainElements(
							PointTo(MatchFields(IgnoreExtras, Fields{
								"To": ContainElements(
									PointTo(MatchFields(IgnoreExtras, Fields{
										"Operation": PointTo(MatchFields(IgnoreExtras, Fields{
											"Methods": ContainElements("GET", "POST"),
											"Paths":   ContainElements("/"),
										})),
									})),
								),
							})),
						),
					}),
				})),
			}))

			newApMatcher := PointTo(MatchFields(IgnoreExtras, Fields{
				"Action": WithTransform(ActionToString, Equal("create")),
				"Obj": PointTo(MatchFields(IgnoreExtras, Fields{
					"Spec": MatchFields(IgnoreExtras, Fields{
						"Selector": PointTo(MatchFields(IgnoreExtras, Fields{
							"MatchLabels": ContainElement("test-service"),
						})),
						"Rules": ContainElements(
							PointTo(MatchFields(IgnoreExtras, Fields{
								"To": ContainElements(
									PointTo(MatchFields(IgnoreExtras, Fields{
										"Operation": PointTo(MatchFields(IgnoreExtras, Fields{
											"Methods": ContainElements("DELETE"),
											"Paths":   ContainElements("/"),
										})),
									})),
								),
							})),
						),
					}),
				})),
			}))

			Expect(result).To(ContainElements(existingApMatcher, newApMatcher))
		})

		It("should create new AP when new rule with same path and methods, but different service is added to ApiRule", func() {
			//given: Cluster state
			existingAp := securityv1beta1.AuthorizationPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						processing.OwnerLabelv1alpha1: fmt.Sprintf("%s.%s", ApiName, ApiNamespace),
					},
				},
				Spec: v1beta1.AuthorizationPolicy{
					Selector: &typev1beta1.WorkloadSelector{
						MatchLabels: map[string]string{
							"app": "test-service",
						},
					},
					Rules: []*v1beta1.Rule{
						{
							To: []*v1beta1.Rule_To{
								{
									Operation: &v1beta1.Operation{
										Methods: []string{"GET", "POST"},
										Paths:   []string{"/"},
									},
								},
							},
						},
					},
				},
			}

			// given: New resources
			existingRule := getRuleForApTest([]string{"GET", "POST"}, "/", "test-service")
			newRule := getRuleForApTest([]string{"GET", "POST"}, "/", "new-service")

			rules := []gatewayv1beta1.Rule{existingRule, newRule}

			apiRule := GetAPIRuleFor(rules)

			ctrlClient := GetFakeClient(&existingAp)
			processor := istio.NewAuthorizationPolicyProcessor(GetTestConfig())

			// when
			result, err := processor.EvaluateReconciliation(context.TODO(), ctrlClient, apiRule)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(2))

			existingApMatcher := PointTo(MatchFields(IgnoreExtras, Fields{
				"Action": WithTransform(ActionToString, Equal("update")),
				"Obj": PointTo(MatchFields(IgnoreExtras, Fields{
					"Spec": MatchFields(IgnoreExtras, Fields{
						"Selector": PointTo(MatchFields(IgnoreExtras, Fields{
							"MatchLabels": ContainElement("test-service"),
						})),
						"Rules": ContainElements(
							PointTo(MatchFields(IgnoreExtras, Fields{
								"To": ContainElements(
									PointTo(MatchFields(IgnoreExtras, Fields{
										"Operation": PointTo(MatchFields(IgnoreExtras, Fields{
											"Methods": ContainElements("GET", "POST"),
											"Paths":   ContainElements("/"),
										})),
									})),
								),
							})),
						),
					}),
				})),
			}))

			newApMatcher := PointTo(MatchFields(IgnoreExtras, Fields{
				"Action": WithTransform(ActionToString, Equal("create")),
				"Obj": PointTo(MatchFields(IgnoreExtras, Fields{
					"Spec": MatchFields(IgnoreExtras, Fields{
						"Selector": PointTo(MatchFields(IgnoreExtras, Fields{
							"MatchLabels": ContainElement("new-service"),
						})),
						"Rules": ContainElements(
							PointTo(MatchFields(IgnoreExtras, Fields{
								"To": ContainElements(
									PointTo(MatchFields(IgnoreExtras, Fields{
										"Operation": PointTo(MatchFields(IgnoreExtras, Fields{
											"Methods": ContainElements("GET", "POST"),
											"Paths":   ContainElements("/"),
										})),
									})),
								),
							})),
						),
					}),
				})),
			}))

			Expect(result).To(ContainElements(existingApMatcher, newApMatcher))
		})

		It("should create new AP and delete old AP when path in ApiRule changed", func() {
			// given: Cluster state
			existingAp := securityv1beta1.AuthorizationPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						processing.OwnerLabelv1alpha1: fmt.Sprintf("%s.%s", ApiName, ApiNamespace),
					},
				},
				Spec: v1beta1.AuthorizationPolicy{
					Selector: &typev1beta1.WorkloadSelector{
						MatchLabels: map[string]string{
							"app": "test-service",
						},
					},
					Rules: []*v1beta1.Rule{
						{
							To: []*v1beta1.Rule_To{
								{
									Operation: &v1beta1.Operation{
										Methods: []string{"GET", "POST"},
										Paths:   []string{"/"},
									},
								},
							},
						},
					},
				},
			}

			ctrlClient := GetFakeClient(&existingAp)
			processor := istio.NewAuthorizationPolicyProcessor(GetTestConfig())

			// given: New resources
			methods := []string{"GET", "POST"}
			path := "/new-path"
			serviceName := "test-service"

			rule := getRuleForApTest(methods, path, serviceName)
			rules := []gatewayv1beta1.Rule{rule}

			apiRule := GetAPIRuleFor(rules)

			// when
			result, err := processor.EvaluateReconciliation(context.TODO(), ctrlClient, apiRule)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(2))

			existingApMatcher := PointTo(MatchFields(IgnoreExtras, Fields{
				"Action": WithTransform(ActionToString, Equal("delete")),
				"Obj": PointTo(MatchFields(IgnoreExtras, Fields{
					"Spec": MatchFields(IgnoreExtras, Fields{
						"Selector": PointTo(MatchFields(IgnoreExtras, Fields{
							"MatchLabels": ContainElement("test-service"),
						})),
						"Rules": ContainElements(
							PointTo(MatchFields(IgnoreExtras, Fields{
								"To": ContainElements(
									PointTo(MatchFields(IgnoreExtras, Fields{
										"Operation": PointTo(MatchFields(IgnoreExtras, Fields{
											"Methods": ContainElements("GET", "POST"),
											"Paths":   ContainElements("/"),
										})),
									})),
								),
							})),
						),
					}),
				})),
			}))

			newApMatcher := PointTo(MatchFields(IgnoreExtras, Fields{
				"Action": WithTransform(ActionToString, Equal("create")),
				"Obj": PointTo(MatchFields(IgnoreExtras, Fields{
					"Spec": MatchFields(IgnoreExtras, Fields{
						"Selector": PointTo(MatchFields(IgnoreExtras, Fields{
							"MatchLabels": ContainElement("test-service"),
						})),
						"Rules": ContainElements(
							PointTo(MatchFields(IgnoreExtras, Fields{
								"To": ContainElements(
									PointTo(MatchFields(IgnoreExtras, Fields{
										"Operation": PointTo(MatchFields(IgnoreExtras, Fields{
											"Methods": ContainElements("GET", "POST"),
											"Paths":   ContainElements("/new-path"),
										})),
									})),
								),
							})),
						),
					}),
				})),
			}))

			Expect(result).To(ContainElements(existingApMatcher, newApMatcher))
		})

	})

	When("Two AP with different methods for same path and service exist", func() {
		It("should create new AP and delete old AP with matching method, when path has changed", func() {
			// given: Cluster state
			unchangedAp := securityv1beta1.AuthorizationPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name: "unchanged-ap",
					Labels: map[string]string{
						processing.OwnerLabelv1alpha1: fmt.Sprintf("%s.%s", ApiName, ApiNamespace),
					},
				},
				Spec: v1beta1.AuthorizationPolicy{
					Selector: &typev1beta1.WorkloadSelector{
						MatchLabels: map[string]string{
							"app": "test-service",
						},
					},
					Rules: []*v1beta1.Rule{
						{
							To: []*v1beta1.Rule_To{
								{
									Operation: &v1beta1.Operation{
										Methods: []string{"DELETE"},
										Paths:   []string{"/"},
									},
								},
							},
						},
					},
				},
			}

			toBeUpdateAp := securityv1beta1.AuthorizationPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name: "to-be-updated-ap",
					Labels: map[string]string{
						processing.OwnerLabelv1alpha1: fmt.Sprintf("%s.%s", ApiName, ApiNamespace),
					},
				},
				Spec: v1beta1.AuthorizationPolicy{
					Selector: &typev1beta1.WorkloadSelector{
						MatchLabels: map[string]string{
							"app": "test-service",
						},
					},
					Rules: []*v1beta1.Rule{
						{
							To: []*v1beta1.Rule_To{
								{
									Operation: &v1beta1.Operation{
										Methods: []string{"GET"},
										Paths:   []string{"/"},
									},
								},
							},
						},
					},
				},
			}

			ctrlClient := GetFakeClient(&toBeUpdateAp, &unchangedAp)
			processor := istio.NewAuthorizationPolicyProcessor(GetTestConfig())

			// given: New resources
			unchangedRule := getRuleForApTest([]string{"DELETE"}, "/", "test-service")
			updatedRule := getRuleForApTest([]string{"GET"}, "/new-path", "test-service")
			rules := []gatewayv1beta1.Rule{updatedRule, unchangedRule}

			apiRule := GetAPIRuleFor(rules)

			// when
			result, err := processor.EvaluateReconciliation(context.TODO(), ctrlClient, apiRule)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(3))

			unchangedApMatcher := PointTo(MatchFields(IgnoreExtras, Fields{
				"Action": WithTransform(ActionToString, Equal("update")),
				"Obj": PointTo(MatchFields(IgnoreExtras, Fields{
					"Spec": MatchFields(IgnoreExtras, Fields{
						"Selector": PointTo(MatchFields(IgnoreExtras, Fields{
							"MatchLabels": ContainElement("test-service"),
						})),
						"Rules": ContainElements(
							PointTo(MatchFields(IgnoreExtras, Fields{
								"To": ContainElements(
									PointTo(MatchFields(IgnoreExtras, Fields{
										"Operation": PointTo(MatchFields(IgnoreExtras, Fields{
											"Methods": ContainElements("DELETE"),
											"Paths":   ContainElements("/"),
										})),
									})),
								),
							})),
						),
					}),
				})),
			}))

			deleteApMatcher := PointTo(MatchFields(IgnoreExtras, Fields{
				"Action": WithTransform(ActionToString, Equal("delete")),
				"Obj": PointTo(MatchFields(IgnoreExtras, Fields{
					"Spec": MatchFields(IgnoreExtras, Fields{
						"Selector": PointTo(MatchFields(IgnoreExtras, Fields{
							"MatchLabels": ContainElement("test-service"),
						})),
						"Rules": ContainElements(
							PointTo(MatchFields(IgnoreExtras, Fields{
								"To": ContainElements(
									PointTo(MatchFields(IgnoreExtras, Fields{
										"Operation": PointTo(MatchFields(IgnoreExtras, Fields{
											"Methods": ContainElements("GET"),
											"Paths":   ContainElements("/"),
										})),
									})),
								),
							})),
						),
					}),
				})),
			}))

			updatedApMatcher := PointTo(MatchFields(IgnoreExtras, Fields{
				"Action": WithTransform(ActionToString, Equal("create")),
				"Obj": PointTo(MatchFields(IgnoreExtras, Fields{
					"Spec": MatchFields(IgnoreExtras, Fields{
						"Selector": PointTo(MatchFields(IgnoreExtras, Fields{
							"MatchLabels": ContainElement("test-service"),
						})),
						"Rules": ContainElements(
							PointTo(MatchFields(IgnoreExtras, Fields{
								"To": ContainElements(
									PointTo(MatchFields(IgnoreExtras, Fields{
										"Operation": PointTo(MatchFields(IgnoreExtras, Fields{
											"Methods": ContainElements("GET"),
											"Paths":   ContainElements("/new-path"),
										})),
									})),
								),
							})),
						),
					}),
				})),
			}))

			Expect(result).To(ContainElements(unchangedApMatcher, updatedApMatcher, deleteApMatcher))
		})
	})

	When("Two AP with same RuleTo for different services exist", func() {
		It("should create new AP and delete old AP with matching service, when path has changed", func() {
			// given: Cluster state
			unchangedAp := securityv1beta1.AuthorizationPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name: "unchanged-ap",
					Labels: map[string]string{
						processing.OwnerLabelv1alpha1: fmt.Sprintf("%s.%s", ApiName, ApiNamespace),
					},
				},
				Spec: v1beta1.AuthorizationPolicy{
					Selector: &typev1beta1.WorkloadSelector{
						MatchLabels: map[string]string{
							"app": "first-service",
						},
					},
					Rules: []*v1beta1.Rule{
						{
							To: []*v1beta1.Rule_To{
								{
									Operation: &v1beta1.Operation{
										Methods: []string{"GET"},
										Paths:   []string{"/"},
									},
								},
							},
						},
					},
				},
			}

			toBeUpdateAp := securityv1beta1.AuthorizationPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name: "to-be-updated-ap",
					Labels: map[string]string{
						processing.OwnerLabelv1alpha1: fmt.Sprintf("%s.%s", ApiName, ApiNamespace),
					},
				},
				Spec: v1beta1.AuthorizationPolicy{
					Selector: &typev1beta1.WorkloadSelector{
						MatchLabels: map[string]string{
							"app": "second-service",
						},
					},
					Rules: []*v1beta1.Rule{
						{
							To: []*v1beta1.Rule_To{
								{
									Operation: &v1beta1.Operation{
										Methods: []string{"GET"},
										Paths:   []string{"/"},
									},
								},
							},
						},
					},
				},
			}

			ctrlClient := GetFakeClient(&toBeUpdateAp, &unchangedAp)
			processor := istio.NewAuthorizationPolicyProcessor(GetTestConfig())

			// given: New resources
			unchangedRule := getRuleForApTest([]string{"GET"}, "/", "first-service")
			updatedRule := getRuleForApTest([]string{"GET"}, "/new-path", "second-service")
			rules := []gatewayv1beta1.Rule{updatedRule, unchangedRule}

			apiRule := GetAPIRuleFor(rules)

			// when
			result, err := processor.EvaluateReconciliation(context.TODO(), ctrlClient, apiRule)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(3))

			unchangedApMatcher := PointTo(MatchFields(IgnoreExtras, Fields{
				"Action": WithTransform(ActionToString, Equal("update")),
				"Obj": PointTo(MatchFields(IgnoreExtras, Fields{
					"Spec": MatchFields(IgnoreExtras, Fields{
						"Selector": PointTo(MatchFields(IgnoreExtras, Fields{
							"MatchLabels": ContainElement("first-service"),
						})),
						"Rules": ContainElements(
							PointTo(MatchFields(IgnoreExtras, Fields{
								"To": ContainElements(
									PointTo(MatchFields(IgnoreExtras, Fields{
										"Operation": PointTo(MatchFields(IgnoreExtras, Fields{
											"Methods": ContainElements("GET"),
											"Paths":   ContainElements("/"),
										})),
									})),
								),
							})),
						),
					}),
				})),
			}))

			deleteApMatcher := PointTo(MatchFields(IgnoreExtras, Fields{
				"Action": WithTransform(ActionToString, Equal("delete")),
				"Obj": PointTo(MatchFields(IgnoreExtras, Fields{
					"Spec": MatchFields(IgnoreExtras, Fields{
						"Selector": PointTo(MatchFields(IgnoreExtras, Fields{
							"MatchLabels": ContainElement("second-service"),
						})),
						"Rules": ContainElements(
							PointTo(MatchFields(IgnoreExtras, Fields{
								"To": ContainElements(
									PointTo(MatchFields(IgnoreExtras, Fields{
										"Operation": PointTo(MatchFields(IgnoreExtras, Fields{
											"Methods": ContainElements("GET"),
											"Paths":   ContainElements("/"),
										})),
									})),
								),
							})),
						),
					}),
				})),
			}))

			updatedApMatcher := PointTo(MatchFields(IgnoreExtras, Fields{
				"Action": WithTransform(ActionToString, Equal("create")),
				"Obj": PointTo(MatchFields(IgnoreExtras, Fields{
					"Spec": MatchFields(IgnoreExtras, Fields{
						"Selector": PointTo(MatchFields(IgnoreExtras, Fields{
							"MatchLabels": ContainElement("second-service"),
						})),
						"Rules": ContainElements(
							PointTo(MatchFields(IgnoreExtras, Fields{
								"To": ContainElements(
									PointTo(MatchFields(IgnoreExtras, Fields{
										"Operation": PointTo(MatchFields(IgnoreExtras, Fields{
											"Methods": ContainElements("GET"),
											"Paths":   ContainElements("/new-path"),
										})),
									})),
								),
							})),
						),
					}),
				})),
			}))

			Expect(result).To(ContainElements(unchangedApMatcher, updatedApMatcher, deleteApMatcher))
		})
	})
})

func getRuleForApTest(methods []string, path string, serviceName string) gatewayv1beta1.Rule {
	jwtConfigJSON := fmt.Sprintf(`{"authentications": [{"issuer": "%s", "jwksUri": "%s"}]}`, JwtIssuer, JwksUri)
	strategies := []*gatewayv1beta1.Authenticator{
		{
			Handler: &gatewayv1beta1.Handler{
				Name: "jwt",
				Config: &runtime.RawExtension{
					Raw: []byte(jwtConfigJSON),
				},
			},
		},
	}

	port := uint32(8080)
	service := &gatewayv1beta1.Service{
		Name: &serviceName,
		Port: &port,
	}

	return GetRuleWithServiceFor(path, methods, []*gatewayv1beta1.Mutator{}, strategies, service)
}
