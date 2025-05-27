package istio_test

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	gomegatypes "github.com/onsi/gomega/types"
	"istio.io/api/security/v1beta1"
	typev1beta1 "istio.io/api/type/v1beta1"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"github.com/kyma-project/api-gateway/internal/processing"
	. "github.com/kyma-project/api-gateway/internal/processing/processing_test"
	"github.com/kyma-project/api-gateway/internal/processing/processors/istio"
)

var _ = Describe("Request Authentication Processor", func() {
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

	getRequestAuthentication := func(name string, serviceName string, jwksUri string, issuer string) securityv1beta1.RequestAuthentication {
		return securityv1beta1.RequestAuthentication{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: ApiNamespace,
				Labels: map[string]string{
					processing.OwnerLabel: fmt.Sprintf("%s.%s", ApiName, ApiNamespace),
				},
			},
			Spec: v1beta1.RequestAuthentication{
				Selector: &typev1beta1.WorkloadSelector{
					MatchLabels: map[string]string{
						"app": serviceName,
					},
				},
				JwtRules: []*v1beta1.JWTRule{
					{
						JwksUri: jwksUri,
						Issuer:  issuer,
					},
				},
			},
		}
	}

	getActionMatcher := func(action string, namespace string, serviceName string, jwksUri string, issuer string) gomegatypes.GomegaMatcher {
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
					"JwtRules": ContainElements(
						PointTo(MatchFields(IgnoreExtras, Fields{
							"JwksUri": Equal(jwksUri),
							"Issuer":  Equal(issuer),
						})),
					),
				}),
			})),
		}))
	}

	It("should produce one RA for a rule with one issuer and two paths", func() {
		// given
		jwt := createIstioJwtAccessStrategy()
		service := &gatewayv1beta1.Service{
			Name: &ServiceName,
			Port: &ServicePort,
		}

		ruleJwt := GetRuleWithServiceFor(HeadersApiPath, ApiMethods, []*gatewayv1beta1.Mutator{}, []*gatewayv1beta1.Authenticator{jwt}, service)
		ruleJwt2 := GetRuleWithServiceFor(ImgApiPath, ApiMethods, []*gatewayv1beta1.Mutator{}, []*gatewayv1beta1.Authenticator{jwt}, service)
		apiRule := GetAPIRuleFor([]gatewayv1beta1.Rule{ruleJwt, ruleJwt2})
		svc := GetService(ServiceName)
		client := GetFakeClient(svc)
		processor := istio.Newv1beta1RequestAuthenticationProcessor(GetTestConfig(), apiRule)

		// when
		result, err := processor.EvaluateReconciliation(context.Background(), client)

		// then
		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(1))
		ra := result[0].Obj.(*securityv1beta1.RequestAuthentication)
		Expect(ra).NotTo(BeNil())
		Expect(ra.ObjectMeta.Name).To(BeEmpty())
		Expect(ra.ObjectMeta.GenerateName).To(Equal(ApiName + "-"))
		Expect(ra.ObjectMeta.Namespace).To(Equal(ApiNamespace))

		Expect(ra.Spec.Selector.MatchLabels[TestSelectorKey]).NotTo(BeNil())
		Expect(ra.Spec.Selector.MatchLabels[TestSelectorKey]).To(Equal(ServiceName))
		Expect(len(ra.Spec.JwtRules)).To(Equal(1))
		Expect(ra.Spec.JwtRules[0].Issuer).To(Equal(JwtIssuer))
		Expect(ra.Spec.JwtRules[0].JwksUri).To(Equal(JwksUri))
	})

	It("should produce RA for a Rule without service, but service definition on ApiRule level", func() {
		// given
		jwt := createIstioJwtAccessStrategy()
		svc := GetService(ServiceName)
		client := GetFakeClient(svc)
		ruleJwt := GetRuleFor(HeadersApiPath, ApiMethods, []*gatewayv1beta1.Mutator{}, []*gatewayv1beta1.Authenticator{jwt})
		apiRule := GetAPIRuleFor([]gatewayv1beta1.Rule{ruleJwt})
		processor := istio.Newv1beta1RequestAuthenticationProcessor(GetTestConfig(), apiRule)

		// when
		result, err := processor.EvaluateReconciliation(context.Background(), client)

		// then
		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(1))

		ra := result[0].Obj.(*securityv1beta1.RequestAuthentication)
		Expect(ra).NotTo(BeNil())
		// The RA should be in .Spec.Service.Namespace
		Expect(ra.Namespace).To(Equal(ApiNamespace))
		Expect(ra.Spec.Selector.MatchLabels[TestSelectorKey]).To(Equal(ServiceName))
	})

	It("should produce RA with service from Rule, when service is configured on Rule and ApiRule level", func() {
		// given
		jwt := createIstioJwtAccessStrategy()
		ruleServiceName := "rule-scope-example-service"
		specServiceNamespace := "spec-service-namespace"
		service := &gatewayv1beta1.Service{
			Name: &ruleServiceName,
			Port: &ServicePort,
		}
		svc := GetService(ruleServiceName, specServiceNamespace)
		client := GetFakeClient(svc)
		ruleJwt := GetRuleWithServiceFor(HeadersApiPath, ApiMethods, []*gatewayv1beta1.Mutator{}, []*gatewayv1beta1.Authenticator{jwt}, service)
		apiRule := GetAPIRuleFor([]gatewayv1beta1.Rule{ruleJwt}, specServiceNamespace)
		processor := istio.Newv1beta1RequestAuthenticationProcessor(GetTestConfig(), apiRule)

		// when
		result, err := processor.EvaluateReconciliation(context.Background(), client)

		// then
		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(1))

		ra := result[0].Obj.(*securityv1beta1.RequestAuthentication)
		Expect(ra).NotTo(BeNil())
		// The RA should be in .Spec.Service.Namespace
		Expect(ra.Namespace).To(Equal(specServiceNamespace))
		Expect(ra.Spec.Selector.MatchLabels[TestSelectorKey]).To(Equal(ruleServiceName))
	})

	It("should produce RA for a Rule with service with configured namespace, in the configured namespace", func() {
		// given
		jwt := createIstioJwtAccessStrategy()
		ruleServiceName := "rule-scope-example-service"
		ruleServiceNamespace := "rule-service-namespace"
		service := &gatewayv1beta1.Service{
			Name:      &ruleServiceName,
			Port:      &ServicePort,
			Namespace: &ruleServiceNamespace,
		}
		svc := GetService(ruleServiceName, ruleServiceNamespace)
		client := GetFakeClient(svc)
		ruleJwt := GetRuleWithServiceFor(HeadersApiPath, ApiMethods, []*gatewayv1beta1.Mutator{}, []*gatewayv1beta1.Authenticator{jwt}, service)
		apiRule := GetAPIRuleFor([]gatewayv1beta1.Rule{ruleJwt})
		processor := istio.Newv1beta1RequestAuthenticationProcessor(GetTestConfig(), apiRule)

		// when
		result, err := processor.EvaluateReconciliation(context.Background(), client)

		// then
		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(1))

		ra := result[0].Obj.(*securityv1beta1.RequestAuthentication)
		Expect(ra).NotTo(BeNil())
		Expect(ra.Spec.Selector.MatchLabels[TestSelectorKey]).To(Equal(ruleServiceName))
		// The RA should be in .Service.Namespace
		Expect(ra.Namespace).To(Equal(ruleServiceNamespace))
		// And the OwnerLabel should point to APIRule namespace
		Expect(ra.Labels[processing.OwnerLabel]).ToNot(BeEmpty())
		Expect(ra.Labels[processing.OwnerLabel]).To(Equal(fmt.Sprintf("%s.%s", apiRule.Name, apiRule.Namespace)))
	})

	It("should produce RA from a rule with two issuers and one path", func() {
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
		svc := GetService(ServiceName)
		client := GetFakeClient(svc)
		service := &gatewayv1beta1.Service{
			Name: &ServiceName,
			Port: &ServicePort,
		}
		ruleJwt := GetRuleWithServiceFor(HeadersApiPath, ApiMethods, []*gatewayv1beta1.Mutator{}, []*gatewayv1beta1.Authenticator{jwt}, service)
		apiRule := GetAPIRuleFor([]gatewayv1beta1.Rule{ruleJwt})
		processor := istio.Newv1beta1RequestAuthenticationProcessor(GetTestConfig(), apiRule)

		// when
		result, err := processor.EvaluateReconciliation(context.Background(), client)

		// then
		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(1))
		ra := result[0].Obj.(*securityv1beta1.RequestAuthentication)

		Expect(ra).NotTo(BeNil())
		Expect(ra.ObjectMeta.Name).To(BeEmpty())
		Expect(ra.ObjectMeta.GenerateName).To(Equal(ApiName + "-"))
		Expect(ra.ObjectMeta.Namespace).To(Equal(ApiNamespace))

		Expect(ra.Spec.Selector.MatchLabels[TestSelectorKey]).NotTo(BeNil())
		Expect(ra.Spec.Selector.MatchLabels[TestSelectorKey]).To(Equal(ServiceName))
		Expect(len(ra.Spec.JwtRules)).To(Equal(2))
		Expect(ra.Spec.JwtRules[0].Issuer).To(Equal(JwtIssuer))
		Expect(ra.Spec.JwtRules[0].JwksUri).To(Equal(JwksUri))
		Expect(ra.Spec.JwtRules[1].Issuer).To(Equal(JwtIssuer2))
		Expect(ra.Spec.JwtRules[1].JwksUri).To(Equal(JwksUri2))
	})

	DescribeTable("should not create RA if access strategy is", func(accessStrategy string) {
		// given
		strategies := []*gatewayv1beta1.Authenticator{
			{
				Handler: &gatewayv1beta1.Handler{
					Name: accessStrategy,
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
		processor := istio.Newv1beta1RequestAuthenticationProcessor(GetTestConfig(), apiRule)

		// when
		result, err := processor.EvaluateReconciliation(context.Background(), client)

		// then
		Expect(err).To(BeNil())
		Expect(result).To(BeEmpty())
	},
		Entry(nil, gatewayv1beta1.AccessStrategyNoAuth),
		Entry(nil, gatewayv1beta1.AccessStrategyAllow),
		Entry(nil, gatewayv1beta1.AccessStrategyNoop),
	)

	It("should create RA when no exists", func() {
		// given: New resources
		jwtRule := GetJwtRuleWithService(JwtIssuer, JwksUri, "test-service")
		rules := []gatewayv1beta1.Rule{jwtRule}
		svc := GetService("test-service")
		client := GetFakeClient(svc)
		apiRule := GetAPIRuleFor(rules)
		processor := istio.Newv1beta1RequestAuthenticationProcessor(GetTestConfig(), apiRule)

		// when
		result, err := processor.EvaluateReconciliation(context.Background(), client)

		// then
		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(1))
		Expect(result[0].Action.String()).To(Equal("create"))
	})

	It("should delete RA when there is no rule configured in ApiRule", func() {
		// given: Cluster state
		existingRa := getRequestAuthentication("raName", "test-service", JwksUri, JwtIssuer)

		ctrlClient := GetFakeClient(&existingRa)

		// given: New resources
		apiRule := GetAPIRuleFor([]gatewayv1beta1.Rule{})
		processor := istio.Newv1beta1RequestAuthenticationProcessor(GetTestConfig(), apiRule)

		// when
		result, err := processor.EvaluateReconciliation(context.Background(), ctrlClient)

		// then
		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(1))
		Expect(result[0].Action.String()).To(Equal("delete"))
	})

	When("RA with JWT config exists", func() {

		It("should update RA when nothing changed", func() {
			// given: Cluster state
			existingRa := getRequestAuthentication("raName", "test-service", JwksUri, JwtIssuer)
			svc := GetService("test-service")
			ctrlClient := GetFakeClient(&existingRa, svc)

			// given: New resources
			jwtRule := GetJwtRuleWithService(JwtIssuer, JwksUri, "test-service")
			rules := []gatewayv1beta1.Rule{jwtRule}
			apiRule := GetAPIRuleFor(rules)
			processor := istio.Newv1beta1RequestAuthenticationProcessor(GetTestConfig(), apiRule)

			// when
			result, err := processor.EvaluateReconciliation(context.Background(), ctrlClient)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(1))
			Expect(result[0].Action.String()).To(Equal("update"))
		})

		It("should delete and create new RA when only service name in JWT Rule has changed", func() {
			// given: Cluster state
			existingRa := getRequestAuthentication("raName", "old-service", JwksUri, JwtIssuer)
			svcOld := GetService("old-service")
			svcUpdated := GetService("updated-service")
			ctrlClient := GetFakeClient(&existingRa, svcOld, svcUpdated)

			// given: New resources
			jwtRule := GetJwtRuleWithService(JwtIssuer, JwksUri, "updated-service")
			rules := []gatewayv1beta1.Rule{jwtRule}
			apiRule := GetAPIRuleFor(rules)
			processor := istio.Newv1beta1RequestAuthenticationProcessor(GetTestConfig(), apiRule)

			// when
			result, err := processor.EvaluateReconciliation(context.Background(), ctrlClient)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(2))

			deleteMatcher := getActionMatcher("delete", ApiNamespace, "old-service", JwksUri, JwtIssuer)
			createMatcher := getActionMatcher("create", ApiNamespace, "updated-service", JwksUri, JwtIssuer)
			Expect(result).To(ContainElements(deleteMatcher, createMatcher))
		})

		It("should create new RA when new service with new JWT config is added to ApiRule", func() {
			// given: Cluster state
			existingRa := getRequestAuthentication("raName", "existing-service", JwksUri, JwtIssuer)
			svcExisting := GetService("existing-service")
			svcNew := GetService("new-service")
			ctrlClient := GetFakeClient(&existingRa, svcExisting, svcNew)

			// given: New resources
			existingJwtRule := GetJwtRuleWithService(JwtIssuer, JwksUri, "existing-service")
			newJwtRule := GetJwtRuleWithService("https://new.issuer.com/", JwksUri, "new-service")
			rules := []gatewayv1beta1.Rule{existingJwtRule, newJwtRule}
			apiRule := GetAPIRuleFor(rules)
			processor := istio.Newv1beta1RequestAuthenticationProcessor(GetTestConfig(), apiRule)

			// when
			result, err := processor.EvaluateReconciliation(context.Background(), ctrlClient)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(2))

			updateResultMatcher := getActionMatcher("update", ApiNamespace, "existing-service", JwksUri, JwtIssuer)
			createResultMatcher := getActionMatcher("create", ApiNamespace, "new-service", JwksUri, "https://new.issuer.com/")
			Expect(result).To(ContainElements(createResultMatcher, updateResultMatcher))
		})

		It("should create new RA and delete old RA when JWT ApiRule has new JWKS URI", func() {
			// given: Cluster state
			existingRa := getRequestAuthentication("raName", "test-service", JwksUri, JwtIssuer)
			svc := GetService("test-service")
			ctrlClient := GetFakeClient(&existingRa, svc)

			// given: New resources
			jwtRule := GetJwtRuleWithService(JwtIssuer, JwksUri2, "test-service")
			rules := []gatewayv1beta1.Rule{jwtRule}
			apiRule := GetAPIRuleFor(rules)
			processor := istio.Newv1beta1RequestAuthenticationProcessor(GetTestConfig(), apiRule)

			// when
			result, err := processor.EvaluateReconciliation(context.Background(), ctrlClient)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(2))

			deleteResultMatcher := getActionMatcher("delete", ApiNamespace, "test-service", JwksUri, JwtIssuer)
			createResultMatcher := getActionMatcher("create", ApiNamespace, "test-service", JwksUri2, JwtIssuer)

			Expect(result).To(ContainElements(deleteResultMatcher, createResultMatcher))
		})
	})

	When("Two RA with same JWT config for different services exist", func() {

		It("should update RAs and create new RA for first-service and delete old RA when JWT issuer in JWT Rule for first-service has changed", func() {
			// given: Cluster state
			firstServiceRa := getRequestAuthentication("firstRa", "first-service", JwksUri, JwtIssuer)
			secondServiceRa := getRequestAuthentication("secondRa", "second-service", JwksUri, JwtIssuer)
			svcFirst := GetService("first-service")
			svcSecond := GetService("second-service")
			ctrlClient := GetFakeClient(&firstServiceRa, &secondServiceRa, svcFirst, svcSecond)

			// given: New resources
			firstJwtRule := GetJwtRuleWithService("https://new.issuer.com/", JwksUri, "first-service")
			secondJwtRule := GetJwtRuleWithService(JwtIssuer, JwksUri, "second-service")
			rules := []gatewayv1beta1.Rule{firstJwtRule, secondJwtRule}
			apiRule := GetAPIRuleFor(rules)
			processor := istio.Newv1beta1RequestAuthenticationProcessor(GetTestConfig(), apiRule)

			// when
			result, err := processor.EvaluateReconciliation(context.Background(), ctrlClient)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(3))

			deleteFirstServiceRaResultMatcher := getActionMatcher("delete", ApiNamespace, "first-service", JwksUri, JwtIssuer)
			createFirstServiceRaResultMatcher := getActionMatcher("create", ApiNamespace, "first-service", JwksUri, "https://new.issuer.com/")
			secondRaResultMatcher := getActionMatcher("update", ApiNamespace, "second-service", JwksUri, JwtIssuer)
			Expect(result).To(ContainElements(deleteFirstServiceRaResultMatcher, createFirstServiceRaResultMatcher, secondRaResultMatcher))
		})

		It("should delete only first-service RA when it was removed from ApiRule", func() {
			// given: Cluster state
			firstServiceRa := getRequestAuthentication("firstRa", "first-service", JwksUri, JwtIssuer)
			secondServiceRa := getRequestAuthentication("secondRa", "second-service", JwksUri, JwtIssuer)
			svcFirst := GetService("first-service")
			svcSecond := GetService("second-service")
			ctrlClient := GetFakeClient(&firstServiceRa, &secondServiceRa, svcFirst, svcSecond)

			// given: New resources
			secondJwtRule := GetJwtRuleWithService(JwtIssuer, JwksUri, "second-service")
			rules := []gatewayv1beta1.Rule{secondJwtRule}
			apiRule := GetAPIRuleFor(rules)
			processor := istio.Newv1beta1RequestAuthenticationProcessor(GetTestConfig(), apiRule)

			// when
			result, err := processor.EvaluateReconciliation(context.Background(), ctrlClient)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(2))

			deleteResultMatcher := getActionMatcher("delete", ApiNamespace, "first-service", JwksUri, JwtIssuer)
			updateResultMatcher := getActionMatcher("update", ApiNamespace, "second-service", JwksUri, JwtIssuer)
			Expect(result).To(ContainElements(deleteResultMatcher, updateResultMatcher))
		})

		It("should create new RA when it has different service", func() {
			// given: Cluster state
			firstServiceRa := getRequestAuthentication("firstRa", "first-service", JwksUri, JwtIssuer)
			secondServiceRa := getRequestAuthentication("secondRa", "second-service", JwksUri, JwtIssuer)
			svcFirst := GetService("first-service")
			svcSecond := GetService("second-service")
			svcNew := GetService("new-service")
			ctrlClient := GetFakeClient(&firstServiceRa, &secondServiceRa, svcFirst, svcSecond, svcNew)

			// given: New resources
			firstJwtRule := GetJwtRuleWithService(JwtIssuer, JwksUri, "first-service")
			secondJwtRule := GetJwtRuleWithService(JwtIssuer, JwksUri, "second-service")
			newJwtRule := GetJwtRuleWithService(JwtIssuer, JwksUri, "new-service")
			rules := []gatewayv1beta1.Rule{firstJwtRule, secondJwtRule, newJwtRule}
			apiRule := GetAPIRuleFor(rules)
			processor := istio.Newv1beta1RequestAuthenticationProcessor(GetTestConfig(), apiRule)

			// when
			result, err := processor.EvaluateReconciliation(context.Background(), ctrlClient)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(3))

			firstRaMatcher := getActionMatcher("update", ApiNamespace, "first-service", JwksUri, JwtIssuer)
			secondRaMatcher := getActionMatcher("update", ApiNamespace, "second-service", JwksUri, JwtIssuer)
			newRaMatcher := getActionMatcher("create", ApiNamespace, "new-service", JwksUri, JwtIssuer)
			Expect(result).To(ContainElements(firstRaMatcher, secondRaMatcher, newRaMatcher))
		})

		It("should delete and create new RA when it has different namespace on spec level", func() {
			// given: Cluster state
			oldRa := getRequestAuthentication("firstRa", "old-service", JwksUri, JwtIssuer)
			svcOld := GetService("old-service")
			svcNewNS := GetService("old-service", "new-namespace")
			ctrlClient := GetFakeClient(&oldRa, svcOld, svcNewNS)

			// given: New resources
			jwtRule := GetJwtRuleWithService(JwtIssuer, JwksUri, "old-service", "new-namespace")
			rules := []gatewayv1beta1.Rule{jwtRule}
			apiRule := GetAPIRuleFor(rules)
			specServiceNamespace := "new-namespace"
			apiRule.Spec.Service.Namespace = &specServiceNamespace
			processor := istio.Newv1beta1RequestAuthenticationProcessor(GetTestConfig(), apiRule)

			// when
			result, err := processor.EvaluateReconciliation(context.Background(), ctrlClient)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(2))

			deletionMatcher := getActionMatcher("delete", ApiNamespace, "old-service", JwksUri, JwtIssuer)
			creationMatcher := getActionMatcher("create", "new-namespace", "old-service", JwksUri, JwtIssuer)
			Expect(result).To(ContainElements(deletionMatcher, creationMatcher))
		})

		It("should delete and create new RA when it has different namespace on rule level", func() {
			// given: Cluster state
			oldRa := getRequestAuthentication("firstRa", "old-service", JwksUri, JwtIssuer)
			svcOld := GetService("old-service")
			svcNewNS := GetService("old-service", "new-namespace")
			ctrlClient := GetFakeClient(&oldRa, svcOld, svcNewNS)

			// given: New resources
			jwtRule := GetJwtRuleWithService(JwtIssuer, JwksUri, "old-service", "new-namespace")
			rules := []gatewayv1beta1.Rule{jwtRule}
			apiRule := GetAPIRuleFor(rules)
			processor := istio.Newv1beta1RequestAuthenticationProcessor(GetTestConfig(), apiRule)

			// when
			result, err := processor.EvaluateReconciliation(context.Background(), ctrlClient)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(2))

			deletionMatcher := getActionMatcher("delete", ApiNamespace, "old-service", JwksUri, JwtIssuer)
			creationMatcher := getActionMatcher("create", "new-namespace", "old-service", JwksUri, JwtIssuer)
			Expect(result).To(ContainElements(deletionMatcher, creationMatcher))
		})
	})

	When("Service has custom selector spec", func() {
		It("should create RA with selector from service", func() {
			// given: New resources
			path := "/"
			serviceName := "test-service"

			rule := getRuleForApTest(methodsGet, path, serviceName)
			rules := []gatewayv1beta1.Rule{rule}
			apiRule := GetAPIRuleFor(rules)
			svc := GetService(serviceName)
			delete(svc.Spec.Selector, "app")
			svc.Spec.Selector["custom"] = serviceName
			client := GetFakeClient(svc)

			processor := istio.Newv1beta1RequestAuthenticationProcessor(GetTestConfig(), apiRule)

			// when
			result, err := processor.EvaluateReconciliation(context.Background(), client)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(1))

			ra := result[0].Obj.(*securityv1beta1.RequestAuthentication)

			Expect(ra).NotTo(BeNil())
			Expect(ra.Spec.Selector.MatchLabels).To(HaveLen(1))
			Expect(ra.Spec.Selector.MatchLabels["custom"]).To(Equal(serviceName))
		})

		It("should create RA with selector from service in different namespace", func() {
			// given: New resources
			path := "/"
			serviceName := "test-service"
			differentNamespace := "different-namespace"

			rule := getRuleForApTest(methodsGet, path, serviceName)
			rule.Service.Namespace = &differentNamespace
			rules := []gatewayv1beta1.Rule{rule}
			apiRule := GetAPIRuleFor(rules)
			svc := GetService(serviceName, differentNamespace)
			delete(svc.Spec.Selector, "app")
			svc.Spec.Selector["custom"] = serviceName
			client := GetFakeClient(svc)

			processor := istio.Newv1beta1RequestAuthenticationProcessor(GetTestConfig(), apiRule)

			// when
			result, err := processor.EvaluateReconciliation(context.Background(), client)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(1))

			ra := result[0].Obj.(*securityv1beta1.RequestAuthentication)

			Expect(ra).NotTo(BeNil())
			Expect(ra.Spec.Selector.MatchLabels).To(HaveLen(1))
			Expect(ra.Spec.Selector.MatchLabels["custom"]).To(Equal(serviceName))
		})

		It("should create RA with selector from service with multiple selector labels", func() {
			// given: New resources
			path := "/"
			serviceName := "test-service"

			rule := getRuleForApTest(methodsGet, path, serviceName)
			rules := []gatewayv1beta1.Rule{rule}
			apiRule := GetAPIRuleFor(rules)
			svc := GetService(serviceName)
			delete(svc.Spec.Selector, "app")
			svc.Spec.Selector["custom"] = serviceName
			svc.Spec.Selector["second-custom"] = "blah"
			client := GetFakeClient(svc)

			processor := istio.Newv1beta1RequestAuthenticationProcessor(GetTestConfig(), apiRule)

			// when
			result, err := processor.EvaluateReconciliation(context.Background(), client)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(1))

			ra := result[0].Obj.(*securityv1beta1.RequestAuthentication)

			Expect(ra).NotTo(BeNil())
			Expect(ra.Spec.Selector.MatchLabels).To(HaveLen(2))
			Expect(ra.Spec.Selector.MatchLabels).To(HaveKeyWithValue("custom", serviceName))
			Expect(ra.Spec.Selector.MatchLabels).To(HaveKeyWithValue("second-custom", "blah"))
		})
	})
})
