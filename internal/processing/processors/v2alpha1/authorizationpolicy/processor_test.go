package authorizationpolicy_test

/*
import (
	"context"
	"fmt"
	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	"net/http"

	"github.com/kyma-project/api-gateway/internal/processing/hashbasedstate"

	"github.com/kyma-project/api-gateway/internal/processing"
	. "github.com/kyma-project/api-gateway/internal/processing/processing_test"
	"github.com/kyma-project/api-gateway/internal/processing/processors/istio"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"github.com/onsi/gomega/types"
	"golang.org/x/exp/slices"
	"istio.io/api/security/v1beta1"
	typev1beta1 "istio.io/api/type/v1beta1"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	requiredScopeA = "scope-a"
	requiredScopeB = "scope-b"
)

var _ = Describe("JwtAuthorization Policy Processor", func() {
	testExpectedScopeKeys := []string{"request.auth.claims[scp]", "request.auth.claims[scope]", "request.auth.claims[scopes]"}

	createIstioJwtAccessStrategy := func() *gatewayv2alpha1.Authenticator {
		jwtConfigJSON := fmt.Sprintf(`{
			"authentications": [{"issuer": "%s", "jwksUri": "%s"}],
			"authorizations": [{"requiredScopes": ["%s", "%s"]}]}`, JwtIssuer, JwksUri, requiredScopeA, requiredScopeB)
		return &gatewayv2alpha1.Authenticator{
			Handler: &gatewayv2alpha1.Handler{
				Name: "jwt",
				Config: &runtime.RawExtension{
					Raw: []byte(jwtConfigJSON),
				},
			},
		}
	}

	createIstioJwtAccessStrategyTwoAuthorizations := func() *gatewayv2alpha1.Authenticator {
		jwtConfigJSON := fmt.Sprintf(`{
			"authentications": [{"issuer": "%s", "jwksUri": "%s"}],
			"authorizations": [{"requiredScopes": ["%s"]}, {"requiredScopes": ["%s"]}]}`, JwtIssuer, JwksUri, requiredScopeA, requiredScopeB)
		return &gatewayv2alpha1.Authenticator{
			Handler: &gatewayv2alpha1.Handler{
				Name: "jwt",
				Config: &runtime.RawExtension{
					Raw: []byte(jwtConfigJSON),
				},
			},
		}
	}

	getAuthorizationPolicy := func(name string, namespace string, serviceName string, methods []string) *securityv1beta1.AuthorizationPolicy {
		ap := securityv1beta1.AuthorizationPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
				Labels: map[string]string{
					processing.OwnerLabel: fmt.Sprintf("%s.%s", ApiName, ApiNamespace),
				},
			},
			Spec: v1beta1.AuthorizationPolicy{
				Selector: &typev1beta1.WorkloadSelector{
					MatchLabels: map[string]string{
						"app": serviceName,
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
									Methods: methods,
									Paths:   []string{"/"},
								},
							},
						},
					},
				},
			},
		}

		apHash, err := hashbasedstate.GetAuthorizationPolicyHash(&ap)
		Expect(err).ShouldNot(HaveOccurred())
		ap.Labels["gateway.kyma-project.io/hash"] = apHash
		ap.Labels["gateway.kyma-project.io/index"] = "0"

		return &ap
	}

	getActionMatcher := func(action string, namespace string, serviceName string, principalsName string, principals types.GomegaMatcher, methods types.GomegaMatcher, paths types.GomegaMatcher) types.GomegaMatcher {
		return PointTo(MatchFields(IgnoreExtras, Fields{
			"Action": WithTransform(ActionToString, Equal(action)),
			"Obj": PointTo(MatchFields(IgnoreExtras, Fields{
				"ObjectMeta": MatchFields(IgnoreExtras, Fields{
					"Namespace": Equal(namespace),
				}),
				"Spec": MatchFields(IgnoreExtras, Fields{
					"Selector": PointTo(MatchFields(IgnoreExtras, Fields{
						"MatchLabels": ContainElement(serviceName),
					})),
					"Rules": ContainElements(
						PointTo(MatchFields(IgnoreExtras, Fields{
							"From": ContainElement(
								PointTo(MatchFields(IgnoreExtras, Fields{
									"Source": PointTo(MatchFields(IgnoreExtras, Fields{
										principalsName: principals,
									})),
								})),
							),
							"To": ContainElements(
								PointTo(MatchFields(IgnoreExtras, Fields{
									"Operation": PointTo(MatchFields(IgnoreExtras, Fields{
										"Methods": methods,
										"Paths":   paths,
									})),
								})),
							),
						})),
					),
				}),
			})),
		}))
	}

	getAudienceMatcher := func(action string, hashLabelValue string, indexLabelValue string, audiences []string) types.GomegaMatcher {
		var audiencesMatchers []types.GomegaMatcher

		for _, audience := range audiences {
			m := PointTo(MatchFields(IgnoreExtras, Fields{
				"Key":    Equal("request.auth.claims[aud]"),
				"Values": ContainElement(audience),
			}))
			audiencesMatchers = append(audiencesMatchers, m)
		}

		return PointTo(MatchFields(IgnoreExtras, Fields{
			"Action": WithTransform(ActionToString, Equal(action)),
			"Obj": PointTo(MatchFields(IgnoreExtras, Fields{
				"ObjectMeta": MatchFields(IgnoreExtras, Fields{
					"Labels": And(
						HaveKeyWithValue("gateway.kyma-project.io/index", indexLabelValue),
						HaveKeyWithValue("gateway.kyma-project.io/hash", hashLabelValue),
					),
				}),
				"Spec": MatchFields(IgnoreExtras, Fields{
					"Rules": ContainElements(
						PointTo(MatchFields(IgnoreExtras, Fields{
							"When": ContainElements(audiencesMatchers),
						})),
					),
				}),
			})),
		}))
	}

	It("should set path to `/*` when the Rule path is `/.*`", func() {
		// given
		jwt := createIstioJwtAccessStrategy()
		service := &gatewayv2alpha1.Service{
			Name: &ServiceName,
			Port: &ServicePort,
		}

		ruleJwt := GetRuleWithServiceFor("/.*", ApiMethods, []*gatewayv2alpha1.Mutator{}, []*gatewayv2alpha1.Authenticator{jwt}, service)
		apiRule := GetAPIRuleFor([]gatewayv2alpha1.Rule{ruleJwt})
		svc := GetService(*apiRule.Spec.Service.Name)
		client := GetFakeClient(svc)
		processor := istio.Newv1beta1AuthorizationPolicyProcessor(GetTestConfig(), &testLogger, apiRule)

		// when
		result, err := processor.EvaluateReconciliation(context.Background(), client)

		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(1))

		ap := result[0].Obj.(*securityv1beta1.AuthorizationPolicy)

		Expect(len(ap.Spec.Rules[0].To[0].Operation.Paths)).To(Equal(1))
		Expect(ap.Spec.Rules[0].To[0].Operation.Paths).To(ContainElement("/*"))
	})



	It("should produce one AP for a Rule without service, but service definition on ApiRule level", func() {
		// given
		jwt := createIstioJwtAccessStrategy()
		ruleJwt := GetRuleFor(HeadersApiPath, ApiMethods, []*gatewayv2alpha1.Mutator{}, []*gatewayv2alpha1.Authenticator{jwt})
		apiRule := GetAPIRuleFor([]gatewayv2alpha1.Rule{ruleJwt})
		svc := GetService(*apiRule.Spec.Service.Name)
		client := GetFakeClient(svc)
		processor := istio.Newv1beta1AuthorizationPolicyProcessor(GetTestConfig(), &testLogger, apiRule)

		// when
		result, err := processor.EvaluateReconciliation(context.Background(), client)

		// then
		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(1))

		ap := result[0].Obj.(*securityv1beta1.AuthorizationPolicy)
		Expect(ap).NotTo(BeNil())
		// The AP should be in .Spec.Service.Namespace
		Expect(ap.Namespace).To(Equal(ApiNamespace))
		Expect(ap.Spec.Selector.MatchLabels[TestSelectorKey]).To(Equal(ServiceName))
	})

	It("should produce AP with service from Rule, when service is configured on Rule and ApiRule level", func() {
		// given
		jwt := createIstioJwtAccessStrategy()
		ruleServiceName := "rule-scope-example-service"
		specServiceNamespace := "spec-service-namespace"
		service := &gatewayv2alpha1.Service{
			Name: &ruleServiceName,
			Port: &ServicePort,
		}
		ruleJwt := GetRuleWithServiceFor(HeadersApiPath, ApiMethods, []*gatewayv2alpha1.Mutator{}, []*gatewayv2alpha1.Authenticator{jwt}, service)
		apiRule := GetAPIRuleFor([]gatewayv2alpha1.Rule{ruleJwt}, specServiceNamespace)
		svc := GetService(ruleServiceName, specServiceNamespace)
		client := GetFakeClient(svc)
		processor := istio.Newv1beta1AuthorizationPolicyProcessor(GetTestConfig(), &testLogger, apiRule)

		// when
		result, err := processor.EvaluateReconciliation(context.Background(), client)

		// then
		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(1))

		ap := result[0].Obj.(*securityv1beta1.AuthorizationPolicy)
		Expect(ap).NotTo(BeNil())
		// The RA should be in .Spec.Service.Namespace
		Expect(ap.Namespace).To(Equal(specServiceNamespace))
		Expect(ap.Spec.Selector.MatchLabels[TestSelectorKey]).To(Equal(ruleServiceName))
	})

	It("should produce one AP for a Rule with service with configured namespace, in the configured namespace", func() {
		// given
		jwt := createIstioJwtAccessStrategy()
		ruleServiceName := "rule-scope-example-service"
		ruleServiceNamespace := "rule-service-namespace"
		service := &gatewayv2alpha1.Service{
			Name:      &ruleServiceName,
			Port:      &ServicePort,
			Namespace: &ruleServiceNamespace,
		}
		ruleJwt := GetRuleWithServiceFor(HeadersApiPath, ApiMethods, []*gatewayv2alpha1.Mutator{}, []*gatewayv2alpha1.Authenticator{jwt}, service)
		apiRule := GetAPIRuleFor([]gatewayv2alpha1.Rule{ruleJwt})
		svc := GetService(ruleServiceName, ruleServiceNamespace)
		client := GetFakeClient(svc)
		processor := istio.Newv1beta1AuthorizationPolicyProcessor(GetTestConfig(), &testLogger, apiRule)

		// when
		result, err := processor.EvaluateReconciliation(context.Background(), client)

		// then
		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(1))

		ap := result[0].Obj.(*securityv1beta1.AuthorizationPolicy)
		Expect(ap).NotTo(BeNil())
		Expect(ap.Spec.Selector.MatchLabels[TestSelectorKey]).To(Equal(ruleServiceName))
		// The AP should be in .Service.Namespace
		Expect(ap.Namespace).To(Equal(ruleServiceNamespace))
		// And the OwnerLabel should point to APIRule namespace
		Expect(ap.Labels[processing.OwnerLabel]).ToNot(BeEmpty())
		Expect(ap.Labels[processing.OwnerLabel]).To(Equal(fmt.Sprintf("%s.%s", apiRule.Name, apiRule.Namespace)))
	})

	When("single handler only", func() {

		It("should create AP with From in Rules Spec for jwt", func() {
			// given
			strategies := []*gatewayv2alpha1.Authenticator{
				{
					Handler: &gatewayv2alpha1.Handler{
						Name: "jwt",
					},
				},
			}

			rule := GetRuleFor(ApiPath, ApiMethods, []*gatewayv2alpha1.Mutator{}, strategies)
			rules := []gatewayv2alpha1.Rule{rule}

			apiRule := GetAPIRuleFor(rules)

			overrideServiceName := "testName"
			overrideServiceNamespace := "testName-namespace"
			overrideServicePort := uint32(8080)

			apiRule.Spec.Service = &gatewayv2alpha1.Service{
				Name:      &overrideServiceName,
				Namespace: &overrideServiceNamespace,
				Port:      &overrideServicePort,
			}

			svc := GetService(overrideServiceName, overrideServiceNamespace)
			client := GetFakeClient(svc)
			processor := istio.Newv1beta1AuthorizationPolicyProcessor(GetTestConfig(), &testLogger, apiRule)

			// when
			result, err := processor.EvaluateReconciliation(context.Background(), client)

			// then
			Expect(err).To(BeNil())
			Expect(len(result)).To(Equal(1))
			ap := result[0].Obj.(*securityv1beta1.AuthorizationPolicy)
			Expect(len(result)).To(Equal(1))
			Expect(ap.Spec.Rules[0].From).NotTo(BeEmpty())
		})

		DescribeTable("should not create AP for handler", func(handler string) {
			// given
			strategies := []*gatewayv2alpha1.Authenticator{
				{
					Handler: &gatewayv2alpha1.Handler{
						Name: handler,
					},
				},
			}

			rule := GetRuleFor(ApiPath, ApiMethods, []*gatewayv2alpha1.Mutator{}, strategies)
			rules := []gatewayv2alpha1.Rule{rule}

			apiRule := GetAPIRuleFor(rules)

			overrideServiceName := "testName"
			overrideServiceNamespace := "testName-namespace"
			overrideServicePort := uint32(8080)

			apiRule.Spec.Service = &gatewayv2alpha1.Service{
				Name:      &overrideServiceName,
				Namespace: &overrideServiceNamespace,
				Port:      &overrideServicePort,
			}

			client := GetFakeClient()
			processor := istio.Newv1beta1AuthorizationPolicyProcessor(GetTestConfig(), &testLogger, apiRule)

			// when
			result, err := processor.EvaluateReconciliation(context.Background(), client)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(BeEmpty())
		},
			Entry(nil, gatewayv2alpha1.AccessStrategyAllow),
			Entry(nil, gatewayv2alpha1.AccessStrategyNoop),
		)

	})

	When("additional handler to JWT", func() {

		DescribeTable("should create AP with From having Source.Principals == cluster.local/ns/istio-system/sa/istio-ingressgateway-service-account for handler", func(handler string) {
			// given
			jwt := createIstioJwtAccessStrategy()
			allow := &gatewayv2alpha1.Authenticator{
				Handler: &gatewayv2alpha1.Handler{
					Name: handler,
				},
			}

			service := &gatewayv2alpha1.Service{
				Name: &ServiceName,
				Port: &ServicePort,
			}

			ruleAllow := GetRuleWithServiceFor(HeadersApiPath, ApiMethods, []*gatewayv2alpha1.Mutator{}, []*gatewayv2alpha1.Authenticator{allow}, service)
			ruleJwt := GetRuleWithServiceFor(ImgApiPath, ApiMethods, []*gatewayv2alpha1.Mutator{}, []*gatewayv2alpha1.Authenticator{jwt}, service)
			apiRule := GetAPIRuleFor([]gatewayv2alpha1.Rule{ruleAllow, ruleJwt})
			svc := GetService(*apiRule.Spec.Service.Name)
			client := GetFakeClient(svc)
			processor := istio.Newv1beta1AuthorizationPolicyProcessor(GetTestConfig(), &testLogger, apiRule)

			// when
			results, err := processor.EvaluateReconciliation(context.Background(), client)

			// then
			Expect(err).To(BeNil())
			Expect(results).To(HaveLen(2))

			for _, result := range results {
				ap := result.Obj.(*securityv1beta1.AuthorizationPolicy)

				Expect(ap).NotTo(BeNil())
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

				Expect(len(ap.Spec.Rules)).To(BeElementOf([]int{1, 3}))
				if len(ap.Spec.Rules) == 3 {
					for i := 0; i < 3; i++ {
						Expect(ap.Spec.Rules[i].When[0].Key).To(BeElementOf(testExpectedScopeKeys))
						Expect(ap.Spec.Rules[i].When).To(HaveLen(2))
						Expect(ap.Spec.Rules[i].When[0].Key).To(BeElementOf(testExpectedScopeKeys))
						Expect(ap.Spec.Rules[i].When[0].Values[0]).To(BeElementOf(requiredScopeA, requiredScopeB))
						Expect(ap.Spec.Rules[i].When[1].Key).To(BeElementOf(testExpectedScopeKeys))
						Expect(ap.Spec.Rules[i].When[1].Values[0]).To(BeElementOf(requiredScopeA, requiredScopeB))
					}
				} else {
					Expect(len(ap.Spec.Rules)).To(Equal(1))
				}
			}
		},
			Entry(nil, gatewayv2alpha1.AccessStrategyNoAuth),
			Entry(nil, gatewayv2alpha1.AccessStrategyAllow),
		)

		It("should create AP for noAuth with From spec having Source.Principals == cluster.local/ns/istio-system/sa/istio-ingressgateway-service-account", func() {
			// given
			jwt := createIstioJwtAccessStrategy()
			noAuth := &gatewayv2alpha1.Authenticator{
				Handler: &gatewayv2alpha1.Handler{
					Name: "no_auth",
				},
			}

			service := &gatewayv2alpha1.Service{
				Name: &ServiceName,
				Port: &ServicePort,
			}

			ruleNoAuth := GetRuleWithServiceFor(HeadersApiPath, ApiMethods, []*gatewayv2alpha1.Mutator{}, []*gatewayv2alpha1.Authenticator{noAuth}, service)
			ruleJwt := GetRuleWithServiceFor(ImgApiPath, ApiMethods, []*gatewayv2alpha1.Mutator{}, []*gatewayv2alpha1.Authenticator{jwt}, service)
			apiRule := GetAPIRuleFor([]gatewayv2alpha1.Rule{ruleNoAuth, ruleJwt})
			svc := GetService(*apiRule.Spec.Service.Name)
			client := GetFakeClient(svc)
			processor := istio.Newv1beta1AuthorizationPolicyProcessor(GetTestConfig(), &testLogger, apiRule)

			// when
			results, err := processor.EvaluateReconciliation(context.Background(), client)

			// then
			Expect(err).To(BeNil())
			Expect(results).To(HaveLen(2))

			for _, result := range results {
				ap := result.Obj.(*securityv1beta1.AuthorizationPolicy)

				Expect(ap).NotTo(BeNil())
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

				Expect(len(ap.Spec.Rules)).To(BeElementOf([]int{1, 3}))
				if len(ap.Spec.Rules) == 3 {
					for i := 0; i < 3; i++ {
						Expect(ap.Spec.Rules[i].When[0].Key).To(BeElementOf(testExpectedScopeKeys))
						Expect(ap.Spec.Rules[i].When).To(HaveLen(2))
						Expect(ap.Spec.Rules[i].When[0].Key).To(BeElementOf(testExpectedScopeKeys))
						Expect(ap.Spec.Rules[i].When[0].Values[0]).To(BeElementOf(requiredScopeA, requiredScopeB))
						Expect(ap.Spec.Rules[i].When[1].Key).To(BeElementOf(testExpectedScopeKeys))
						Expect(ap.Spec.Rules[i].When[1].Values[0]).To(BeElementOf(requiredScopeA, requiredScopeB))
					}
				} else {
					Expect(len(ap.Spec.Rules)).To(Equal(1))
				}
			}
		})

		It("should create AP for noop with From spec having Source.Principals == cluster.local/ns/kyma-system/sa/oathkeeper-maester-account", func() {
			// given
			jwt := createIstioJwtAccessStrategy()
			noop := &gatewayv2alpha1.Authenticator{
				Handler: &gatewayv2alpha1.Handler{
					Name: "noop",
				},
			}

			service := &gatewayv2alpha1.Service{
				Name: &ServiceName,
				Port: &ServicePort,
			}

			ruleNoop := GetRuleWithServiceFor(HeadersApiPath, ApiMethods, []*gatewayv2alpha1.Mutator{}, []*gatewayv2alpha1.Authenticator{noop}, service)
			ruleJwt := GetRuleWithServiceFor(ImgApiPath, ApiMethods, []*gatewayv2alpha1.Mutator{}, []*gatewayv2alpha1.Authenticator{jwt}, service)
			apiRule := GetAPIRuleFor([]gatewayv2alpha1.Rule{ruleNoop, ruleJwt})
			svc := GetService(*apiRule.Spec.Service.Name)
			client := GetFakeClient(svc)
			processor := istio.Newv1beta1AuthorizationPolicyProcessor(GetTestConfig(), &testLogger, apiRule)

			// when
			results, err := processor.EvaluateReconciliation(context.Background(), client)

			// then
			Expect(err).To(BeNil())
			Expect(results).To(HaveLen(2))

			for _, result := range results {
				ap := result.Obj.(*securityv1beta1.AuthorizationPolicy)

				Expect(ap).NotTo(BeNil())
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

				Expect(len(ap.Spec.Rules)).To(BeElementOf([]int{1, 3}))
				if len(ap.Spec.Rules) == 3 {
					for i := 0; i < 3; i++ {
						Expect(ap.Spec.Rules[i].When[0].Key).To(BeElementOf(testExpectedScopeKeys))
						Expect(ap.Spec.Rules[i].When).To(HaveLen(2))
						Expect(ap.Spec.Rules[i].When[0].Key).To(BeElementOf(testExpectedScopeKeys))
						Expect(ap.Spec.Rules[i].When[0].Values[0]).To(BeElementOf(requiredScopeA, requiredScopeB))
						Expect(ap.Spec.Rules[i].When[1].Key).To(BeElementOf(testExpectedScopeKeys))
						Expect(ap.Spec.Rules[i].When[1].Values[0]).To(BeElementOf(requiredScopeA, requiredScopeB))
					}
				} else {
					Expect(len(ap.Spec.Rules)).To(Equal(1))
				}
			}
		})
	})

	It("should create AP when no exists", func() {
		// given: New resources
		path := "/"
		serviceName := "test-service"

		rule := getRuleForApTest(methodsGet, path, serviceName)
		rules := []gatewayv2alpha1.Rule{rule}
		apiRule := GetAPIRuleFor(rules)
		svc := GetService(serviceName)
		client := GetFakeClient(svc)

		processor := istio.Newv1beta1AuthorizationPolicyProcessor(GetTestConfig(), &testLogger, apiRule)

		// when
		result, err := processor.EvaluateReconciliation(context.Background(), client)

		// then
		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(1))
		Expect(result[0].Action.String()).To(Equal("create"))
	})

	It("should update AP when path, methods and service name didn't change", func() {
		// given: Cluster state
		existingAp := getAuthorizationPolicy("raName", ApiNamespace, "test-service", []string{"GET", "POST"})

		// given: New resources
		path := "/"
		serviceName := "test-service"

		rule := getRuleForApTest(methodsGetPost, path, serviceName)
		rules := []gatewayv2alpha1.Rule{rule}

		apiRule := GetAPIRuleFor(rules)
		svc := GetService(serviceName)
		client := GetFakeClient(existingAp, svc)

		processor := istio.Newv1beta1AuthorizationPolicyProcessor(GetTestConfig(), &testLogger, apiRule)

		// when
		result, err := processor.EvaluateReconciliation(context.Background(), client)

		// then
		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(1))

		updateMatcher := getActionMatcher("update", ApiNamespace, "test-service", "RequestPrincipals", ContainElements("https://oauth2.example.com//*"), ContainElements("GET", "POST"), ContainElements("/"))
		Expect(result).To(ContainElements(updateMatcher))
	})

	It("should delete AP when there is no desired AP", func() {
		//given: Cluster state
		existingAp := getAuthorizationPolicy("raName", ApiNamespace, "test-service", []string{"GET", "POST"})
		svc := GetService("test-service")
		ctrlClient := GetFakeClient(existingAp, svc)

		// given: New resources
		apiRule := GetAPIRuleFor([]gatewayv2alpha1.Rule{})
		processor := istio.Newv1beta1AuthorizationPolicyProcessor(GetTestConfig(), &testLogger, apiRule)

		// when
		result, err := processor.EvaluateReconciliation(context.Background(), ctrlClient)

		// then
		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(1))

		resultMatcher := getActionMatcher("delete", ApiNamespace, "test-service", "RequestPrincipals", ContainElements("*"), ContainElements("GET", "POST"), ContainElements("/"))
		Expect(result).To(ContainElements(resultMatcher))
	})

	When("AP with RuleTo exists", func() {
		It("should create new AP and update existing AP when new rule with same methods and service but different path is added to ApiRule", func() {
			// given: Cluster state
			existingAp := getAuthorizationPolicy("raName", ApiNamespace, "test-service", []string{"GET", "POST"})
			svc := GetService("test-service")
			ctrlClient := GetFakeClient(existingAp, svc)

			// given: New resources

			existingRule := getRuleForApTest(methodsGetPost, "/", "test-service")
			newRule := getRuleForApTest(methodsGetPost, "/new-path", "test-service")
			rules := []gatewayv2alpha1.Rule{existingRule, newRule}

			apiRule := GetAPIRuleFor(rules)
			processor := istio.Newv1beta1AuthorizationPolicyProcessor(GetTestConfig(), &testLogger, apiRule)

			// when
			result, err := processor.EvaluateReconciliation(context.Background(), ctrlClient)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(2))

			updateExistingApMatcher := getActionMatcher("update", ApiNamespace, "test-service", "RequestPrincipals", ContainElements("https://oauth2.example.com//*"), ContainElements("GET", "POST"), ContainElements("/"))
			newApMatcher := getActionMatcher("create", ApiNamespace, "test-service", "RequestPrincipals", ContainElements("https://oauth2.example.com//*"), ContainElements("GET", "POST"), ContainElements("/new-path"))
			Expect(result).To(ContainElements(updateExistingApMatcher, newApMatcher))
		})

		It("should create new AP and update existing AP when new rule with same path and service but different methods is added to ApiRule", func() {
			// given: Cluster state
			existingAp := getAuthorizationPolicy("raName", ApiNamespace, "test-service", []string{"GET", "POST"})
			svc := GetService("test-service")
			ctrlClient := GetFakeClient(existingAp, svc)

			// given: New resources

			existingRule := getRuleForApTest(methodsGetPost, "/", "test-service")
			newRule := getRuleForApTest(methodsDelete, "/", "test-service")
			rules := []gatewayv2alpha1.Rule{existingRule, newRule}

			apiRule := GetAPIRuleFor(rules)
			processor := istio.Newv1beta1AuthorizationPolicyProcessor(GetTestConfig(), &testLogger, apiRule)

			// when
			result, err := processor.EvaluateReconciliation(context.Background(), ctrlClient)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(2))

			updateExistingApMatcher := getActionMatcher("update", ApiNamespace, "test-service", "RequestPrincipals", ContainElements("https://oauth2.example.com//*"), ContainElements("GET", "POST"), ContainElements("/"))
			newApMatcher := getActionMatcher("create", ApiNamespace, "test-service", "RequestPrincipals", ContainElements("https://oauth2.example.com//*"), ContainElements("DELETE"), ContainElements("/"))
			Expect(result).To(ContainElements(updateExistingApMatcher, newApMatcher))
		})

		It("should create new AP and update existing AP when new rule with same path and methods, but different service is added to ApiRule", func() {
			//given: Cluster state
			existingAp := getAuthorizationPolicy("raName", ApiNamespace, "test-service", []string{"GET", "POST"})
			// given: New resources
			existingRule := getRuleForApTest(methodsGetPost, "/", "test-service")
			newRule := getRuleForApTest(methodsGetPost, "/", "new-service")

			rules := []gatewayv2alpha1.Rule{existingRule, newRule}
			apiRule := GetAPIRuleFor(rules)
			svc1 := GetService("test-service")
			svc2 := GetService("new-service")
			ctrlClient := GetFakeClient(existingAp, svc1, svc2)
			processor := istio.Newv1beta1AuthorizationPolicyProcessor(GetTestConfig(), &testLogger, apiRule)

			// when
			result, err := processor.EvaluateReconciliation(context.Background(), ctrlClient)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(2))

			updateExistingApMatcher := getActionMatcher("update", ApiNamespace, "test-service", "RequestPrincipals", ContainElements("https://oauth2.example.com//*"), ContainElements("GET", "POST"), ContainElements("/"))
			newApMatcher := getActionMatcher("create", ApiNamespace, "new-service", "RequestPrincipals", ContainElements("https://oauth2.example.com//*"), ContainElements("GET", "POST"), ContainElements("/"))
			Expect(result).To(ContainElements(updateExistingApMatcher, newApMatcher))
		})

		It("should recreate AP when path in ApiRule changed", func() {
			// given: Cluster state
			existingAp := getAuthorizationPolicy("raName", ApiNamespace, "test-service", []string{"GET", "POST"})
			svc := GetService("test-service")
			ctrlClient := GetFakeClient(existingAp, svc)

			// given: New resources
			path := "/new-path"
			serviceName := "test-service"

			rule := getRuleForApTest(methodsGetPost, path, serviceName)
			rules := []gatewayv2alpha1.Rule{rule}

			apiRule := GetAPIRuleFor(rules)
			processor := istio.Newv1beta1AuthorizationPolicyProcessor(GetTestConfig(), &testLogger, apiRule)

			// when
			result, err := processor.EvaluateReconciliation(context.Background(), ctrlClient)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(2))

			existingApMatcher := getActionMatcher("delete", ApiNamespace, "test-service", "RequestPrincipals", ContainElements("*"), ContainElements("GET", "POST"), ContainElements("/"))
			newApMatcher := getActionMatcher("create", ApiNamespace, "test-service", "RequestPrincipals", ContainElements("https://oauth2.example.com//*"), ContainElements("GET", "POST"), ContainElements("/new-path"))
			Expect(result).To(ContainElements(existingApMatcher, newApMatcher))
		})

	})

	When("Two AP with different methods for same path and service exist", func() {
		It("should create new AP, delete old AP and update unchanged AP with matching method, when path has changed", func() {
			// given: Cluster state
			unchangedAp := getAuthorizationPolicy("unchanged-ap", ApiNamespace, "test-service", []string{"DELETE"})
			toBeUpdateAp := getAuthorizationPolicy("to-be-updated-ap", ApiNamespace, "test-service", []string{"GET"})
			svc := GetService("test-service")
			ctrlClient := GetFakeClient(toBeUpdateAp, unchangedAp, svc)

			// given: New resources
			unchangedRule := getRuleForApTest(methodsDelete, "/", "test-service")
			updatedRule := getRuleForApTest(methodsGet, "/new-path", "test-service")
			rules := []gatewayv2alpha1.Rule{updatedRule, unchangedRule}

			apiRule := GetAPIRuleFor(rules)
			processor := istio.Newv1beta1AuthorizationPolicyProcessor(GetTestConfig(), &testLogger, apiRule)

			// when
			result, err := processor.EvaluateReconciliation(context.Background(), ctrlClient)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(3))

			updateUnchangedApMatcher := getActionMatcher("update", ApiNamespace, "test-service", "RequestPrincipals", ContainElements("https://oauth2.example.com//*"), ContainElements("DELETE"), ContainElements("/"))
			deleteMatcher := getActionMatcher("delete", ApiNamespace, "test-service", "RequestPrincipals", ContainElements("*"), ContainElements("GET"), ContainElements("/"))
			createdMatcher := getActionMatcher("create", ApiNamespace, "test-service", "RequestPrincipals", ContainElements("https://oauth2.example.com//*"), ContainElements("GET"), ContainElements("/new-path"))
			Expect(result).To(ContainElements(updateUnchangedApMatcher, deleteMatcher, createdMatcher))
		})
	})

	When("Namespace changes", func() {
		It("should create new AP in new namespace and delete old AP, namespace on spec level", func() {
			// given: Cluster state
			oldAP := getAuthorizationPolicy("unchanged-ap", ApiNamespace, "test-service", []string{"DELETE"})
			specNewServiceNamespace := "new-namespace"
			svc := GetService("test-service")
			svcNewNS := GetService("test-service", specNewServiceNamespace)
			ctrlClient := GetFakeClient(oldAP, svc, svcNewNS)

			// given: New resources
			movedRule := getRuleForApTest(methodsDelete, "/", "test-service")
			rules := []gatewayv2alpha1.Rule{movedRule}

			apiRule := GetAPIRuleFor(rules)
			apiRule.Spec.Service.Namespace = &specNewServiceNamespace
			processor := istio.Newv1beta1AuthorizationPolicyProcessor(GetTestConfig(), &testLogger, apiRule)

			// when
			result, err := processor.EvaluateReconciliation(context.Background(), ctrlClient)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(2))

			deleteMatcher := getActionMatcher("delete", ApiNamespace, "test-service", "RequestPrincipals", ContainElements("*"), ContainElements("DELETE"), ContainElements("/"))
			createMatcher := getActionMatcher("create", "new-namespace", "test-service", "RequestPrincipals", ContainElements("https://oauth2.example.com//*"), ContainElements("DELETE"), ContainElements("/"))
			Expect(result).To(ContainElements(deleteMatcher, createMatcher))
		})

		It("should create new AP in new namespace and delete old AP, namespace on rule level", func() {
			// given: Cluster state
			oldAP := getAuthorizationPolicy("unchanged-ap", ApiNamespace, "test-service", []string{"DELETE"})
			svc := GetService("test-service", "new-namespace")
			ctrlClient := GetFakeClient(oldAP, svc)

			// given: New resources
			movedRule := getRuleForApTest(methodsDelete, "/", "test-service", "new-namespace")
			rules := []gatewayv2alpha1.Rule{movedRule}

			apiRule := GetAPIRuleFor(rules)
			processor := istio.Newv1beta1AuthorizationPolicyProcessor(GetTestConfig(), &testLogger, apiRule)

			// when
			result, err := processor.EvaluateReconciliation(context.Background(), ctrlClient)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(2))

			deleteMatcher := getActionMatcher("delete", ApiNamespace, "test-service", "RequestPrincipals", ContainElements("*"), ContainElements("DELETE"), ContainElements("/"))
			createMatcher := getActionMatcher("create", "new-namespace", "test-service", "RequestPrincipals", ContainElements("https://oauth2.example.com//*"), ContainElements("DELETE"), ContainElements("/"))
			Expect(result).To(ContainElements(deleteMatcher, createMatcher))
		})
	})

	When("Two AP with same RuleTo for different services exist", func() {
		It("should update unchanged AP and update AP with matching service, when path has changed", func() {
			// given: Cluster state
			unchangedAp := getAuthorizationPolicy("unchanged-ap", ApiNamespace, "first-service", []string{"GET"})
			toBeUpdateAp := getAuthorizationPolicy("to-be-updated-ap", ApiNamespace, "second-service", []string{"GET"})
			svc1 := GetService("first-service")
			svc2 := GetService("second-service")
			ctrlClient := GetFakeClient(toBeUpdateAp, unchangedAp, svc1, svc2)

			// given: New resources
			unchangedRule := getRuleForApTest(methodsGet, "/", "first-service")
			updatedRule := getRuleForApTest(methodsGet, "/new-path", "second-service")
			rules := []gatewayv2alpha1.Rule{updatedRule, unchangedRule}

			apiRule := GetAPIRuleFor(rules)
			processor := istio.Newv1beta1AuthorizationPolicyProcessor(GetTestConfig(), &testLogger, apiRule)

			// when
			result, err := processor.EvaluateReconciliation(context.Background(), ctrlClient)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(3))

			updateUnchangedApMatcher := getActionMatcher("update", ApiNamespace, "first-service", "RequestPrincipals", ContainElements("https://oauth2.example.com//*"), ContainElements("GET"), ContainElements("/"))
			deleteMatcher := getActionMatcher("delete", ApiNamespace, "second-service", "RequestPrincipals", ContainElements("*"), ContainElements("GET"), ContainElements("/"))
			createdApMatcher := getActionMatcher("create", ApiNamespace, "second-service", "RequestPrincipals", ContainElements("https://oauth2.example.com//*"), ContainElements("GET"), ContainElements("/new-path"))
			Expect(result).To(ContainElements(updateUnchangedApMatcher, deleteMatcher, createdApMatcher))
		})
	})

	When("Service has custom selector spec", func() {
		It("should create AP with selector from service", func() {
			// given: New resources
			path := "/"
			serviceName := "test-service"

			rule := getRuleForApTest(methodsGet, path, serviceName)
			rules := []gatewayv2alpha1.Rule{rule}
			apiRule := GetAPIRuleFor(rules)
			svc := GetService(serviceName)
			delete(svc.Spec.Selector, "app")
			svc.Spec.Selector["custom"] = serviceName
			client := GetFakeClient(svc)

			processor := istio.Newv1beta1AuthorizationPolicyProcessor(GetTestConfig(), &testLogger, apiRule)

			// when
			result, err := processor.EvaluateReconciliation(context.Background(), client)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(1))

			ap := result[0].Obj.(*securityv1beta1.AuthorizationPolicy)

			Expect(ap).NotTo(BeNil())
			Expect(ap.Spec.Selector.MatchLabels).To(HaveLen(1))
			Expect(ap.Spec.Selector.MatchLabels["custom"]).To(Equal(serviceName))
		})

		It("should create AP with selector from service in different namespace", func() {
			// given: New resources
			path := "/"
			serviceName := "test-service"
			differentNamespace := "different-namespace"

			rule := getRuleForApTest(methodsGet, path, serviceName)
			rule.Service.Namespace = &differentNamespace
			rules := []gatewayv2alpha1.Rule{rule}
			apiRule := GetAPIRuleFor(rules)
			svc := GetService(serviceName, differentNamespace)
			delete(svc.Spec.Selector, "app")
			svc.Spec.Selector["custom"] = serviceName
			client := GetFakeClient(svc)

			processor := istio.Newv1beta1AuthorizationPolicyProcessor(GetTestConfig(), &testLogger, apiRule)

			// when
			result, err := processor.EvaluateReconciliation(context.Background(), client)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(1))

			ap := result[0].Obj.(*securityv1beta1.AuthorizationPolicy)

			Expect(ap).NotTo(BeNil())
			Expect(ap.Spec.Selector.MatchLabels).To(HaveLen(1))
			Expect(ap.Spec.Selector.MatchLabels["custom"]).To(Equal(serviceName))
		})

		It("should create AP with selector from service with multiple selector labels", func() {
			// given: New resources
			path := "/"
			serviceName := "test-service"

			rule := getRuleForApTest(methodsGet, path, serviceName)
			rules := []gatewayv2alpha1.Rule{rule}
			apiRule := GetAPIRuleFor(rules)
			svc := GetService(serviceName)
			delete(svc.Spec.Selector, "app")
			svc.Spec.Selector["custom"] = serviceName
			svc.Spec.Selector["second-custom"] = "blah"
			client := GetFakeClient(svc)

			processor := istio.Newv1beta1AuthorizationPolicyProcessor(GetTestConfig(), &testLogger, apiRule)

			// when
			result, err := processor.EvaluateReconciliation(context.Background(), client)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(1))

			ap := result[0].Obj.(*securityv1beta1.AuthorizationPolicy)

			Expect(ap).NotTo(BeNil())
			Expect(ap.Spec.Selector.MatchLabels).To(HaveLen(2))
			Expect(ap.Spec.Selector.MatchLabels).To(HaveKeyWithValue("custom", serviceName))
			Expect(ap.Spec.Selector.MatchLabels).To(HaveKeyWithValue("second-custom", "blah"))
		})
	})

	for _, missingLabel := range []string{"gateway.kyma-project.io/hash", "gateway.kyma-project.io/index"} {

		It(fmt.Sprintf("should delete existing AP without hashing label %s and create new AP for same authorization in Rule", missingLabel), func() {
			// given: Cluster state
			serviceName := "test-service"

			ap := getAuthorizationPolicy("ap", ApiNamespace, serviceName, []string{"GET"})
			ap.Spec.Rules[0].When = []*v1beta1.Condition{
				{
					Key:    "request.auth.claims[aud]",
					Values: []string{"audience1"},
				},
			}

			// We need to store the hash for comparison later
			expectedHash := ap.Labels["gateway.kyma-project.io/hash"]

			delete(ap.Labels, missingLabel)

			svc := GetService(serviceName)
			ctrlClient := GetFakeClient(ap, svc)

			// given: ApiRule with updated audiences in jwt authorizations
			authorization := `{"audiences": ["audience1"]}`
			jwtConfigJSON := fmt.Sprintf(`{"authentications": [{"issuer": "%s", "jwksUri": "%s"}], "authorizations": [%s]}`,
				JwtIssuer, JwksUri, authorization)
			jwtAuth := &gatewayv2alpha1.Authenticator{
				Handler: &gatewayv2alpha1.Handler{
					Name: "jwt",
					Config: &runtime.RawExtension{
						Raw: []byte(jwtConfigJSON),
					},
				},
			}

			service := &gatewayv2alpha1.Service{
				Name: &serviceName,
				Port: &ServicePort,
			}

			rule := GetRuleWithServiceFor("/", []gatewayv2alpha1.HttpMethod{http.MethodGet}, []*gatewayv2alpha1.Mutator{}, []*gatewayv2alpha1.Authenticator{jwtAuth}, service)
			rules := []gatewayv2alpha1.Rule{rule}

			apiRule := GetAPIRuleFor(rules)
			processor := istio.Newv1beta1AuthorizationPolicyProcessor(GetTestConfig(), &testLogger, apiRule)

			// when
			result, err := processor.EvaluateReconciliation(context.Background(), ctrlClient)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(2))

			newMatcher := getAudienceMatcher("create", expectedHash, "0", []string{"audience1"})
			deletedMatcher := PointTo(MatchFields(IgnoreExtras, Fields{
				"Action": WithTransform(ActionToString, Equal("delete")),
				"Obj": PointTo(MatchFields(IgnoreExtras, Fields{
					"Spec": MatchFields(IgnoreExtras, Fields{
						"Rules": ContainElements(
							PointTo(MatchFields(IgnoreExtras, Fields{
								"When": ContainElements(
									PointTo(MatchFields(IgnoreExtras, Fields{
										"Key":    Equal("request.auth.claims[aud]"),
										"Values": ContainElement("audience1"),
									}))),
							})),
						),
					}),
				})),
			}))

			Expect(result).To(ContainElements(newMatcher, deletedMatcher))
		})
	}
})

func getRuleForApTest(methods []gatewayv2alpha1.HttpMethod, path string, serviceName string, namespace ...string) gatewayv2alpha1.Rule {
	jwtConfigJSON := fmt.Sprintf(`{"authentications": [{"issuer": "%s", "jwksUri": "%s"}]}`, JwtIssuer, JwksUri)
	strategies := []*gatewayv2alpha1.Authenticator{
		{
			Handler: &gatewayv2alpha1.Handler{
				Name: "jwt",
				Config: &runtime.RawExtension{
					Raw: []byte(jwtConfigJSON),
				},
			},
		},
	}

	port := uint32(8080)
	service := &gatewayv2alpha1.Service{
		Name: &serviceName,
		Port: &port,
	}
	if len(namespace) > 0 {
		service.Namespace = &namespace[0]
	}

	return GetRuleWithServiceFor(path, methods, []*gatewayv2alpha1.Mutator{}, strategies, service)
}

*/
