package authorizationpolicy_test

import (
	"context"
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"istio.io/api/security/v1beta1"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"

	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	"github.com/kyma-project/api-gateway/internal/processing/processors/v2alpha1/authorizationpolicy"
)

var _ = Describe("Processing JWT rules", func() {
	imgPath := "/img"
	requiredScopeA := "scope-a"
	requiredScopeB := "scope-b"
	jwtIssuer := "https://oauth2.example.com/"

	It("should produce two APs for a rule with one issuer and two paths", func() {
		// given
		ruleJwt := newJwtRuleBuilderWithDummyData().
			addJwtAuthorizationRequiredScopes("scope-a", "scope-b").
			build()
		ruleJwt2 := newJwtRuleBuilderWithDummyData().
			withPath(imgPath).
			addJwtAuthorizationRequiredScopes("scope-a", "scope-b").
			build()

		apiRule := newAPIRuleBuilderWithDummyData().
			withRules(ruleJwt, ruleJwt2).
			build()
		svc := newServiceBuilderWithDummyData().build()
		gateway := newGatewayBuilderWithDummyData().build()
		client := getFakeClient(svc)
		processor := authorizationpolicy.NewProcessor(&testLogger, apiRule, gateway, client)

		// when
		result, err := processor.EvaluateReconciliation(context.Background(), client)

		// then
		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(2))

		ap1 := result[0].Obj.(*securityv1beta1.AuthorizationPolicy)
		ap2 := result[1].Obj.(*securityv1beta1.AuthorizationPolicy)

		Expect(ap1).NotTo(BeNil())
		Expect(ap1.ObjectMeta.Name).To(BeEmpty())
		Expect(ap1.ObjectMeta.GenerateName).To(Equal(apiRuleName + "-"))

		Expect(ap1.Spec.Selector.MatchLabels["app"]).NotTo(BeNil())
		Expect(ap1.Spec.Selector.MatchLabels["app"]).To(Equal(serviceName))
		Expect(len(ap1.Spec.Rules)).To(Equal(3))
		Expect(len(ap1.Spec.Rules[0].From)).To(Equal(1))
		Expect(len(ap1.Spec.Rules[0].From[0].Source.RequestPrincipals)).To(Equal(1))
		Expect(ap1.Spec.Rules[0].From[0].Source.RequestPrincipals[0]).To(Equal(fmt.Sprintf("%s/*", jwtIssuer)))
		Expect(len(ap1.Spec.Rules[0].To)).To(Equal(1))
		Expect(len(ap1.Spec.Rules[0].To[0].Operation.Methods)).To(Equal(1))
		Expect(ap1.Spec.Rules[0].To[0].Operation.Methods).To(ContainElements([]string{http.MethodGet}))
		Expect(len(ap1.Spec.Rules[0].To[0].Operation.Paths)).To(Equal(1))

		for i := 0; i < 3; i++ {
			Expect(ap1.Spec.Rules[i].When[0].Key).To(BeElementOf(testExpectedScopeKeys))
			Expect(ap1.Spec.Rules[i].When).To(HaveLen(2))
			Expect(ap1.Spec.Rules[i].When[0].Key).To(BeElementOf(testExpectedScopeKeys))
			Expect(ap1.Spec.Rules[i].When[0].Values[0]).To(BeElementOf(requiredScopeA, requiredScopeB))
			Expect(ap1.Spec.Rules[i].When[1].Key).To(BeElementOf(testExpectedScopeKeys))
			Expect(ap1.Spec.Rules[i].When[1].Values[0]).To(BeElementOf(requiredScopeA, requiredScopeB))
		}

		Expect(ap2).NotTo(BeNil())
		Expect(ap2.ObjectMeta.Name).To(BeEmpty())
		Expect(ap2.ObjectMeta.GenerateName).To(Equal(apiRuleName + "-"))
		Expect(ap2.ObjectMeta.Namespace).To(Equal(apiRuleNamespace))

		Expect(ap2.Spec.Selector.MatchLabels["app"]).NotTo(BeNil())
		Expect(ap2.Spec.Selector.MatchLabels["app"]).To(Equal(serviceName))
		Expect(len(ap2.Spec.Rules)).To(Equal(3))
		Expect(len(ap2.Spec.Rules[0].From)).To(Equal(1))
		Expect(len(ap2.Spec.Rules[0].From[0].Source.RequestPrincipals)).To(Equal(1))
		Expect(ap2.Spec.Rules[0].From[0].Source.RequestPrincipals[0]).To(Equal(fmt.Sprintf("%s/*", jwtIssuer)))
		Expect(len(ap2.Spec.Rules[0].To)).To(Equal(1))
		Expect(len(ap2.Spec.Rules[0].To[0].Operation.Methods)).To(Equal(1))
		Expect(ap2.Spec.Rules[0].To[0].Operation.Methods).To(ContainElements([]string{http.MethodGet}))
		Expect(len(ap2.Spec.Rules[0].To[0].Operation.Paths)).To(Equal(1))

		for i := 0; i < 3; i++ {
			Expect(ap2.Spec.Rules[i].When[0].Key).To(BeElementOf(testExpectedScopeKeys))
			Expect(ap2.Spec.Rules[i].When).To(HaveLen(2))
			Expect(ap2.Spec.Rules[i].When[0].Key).To(BeElementOf(testExpectedScopeKeys))
			Expect(ap2.Spec.Rules[i].When[0].Values[0]).To(BeElementOf(requiredScopeA, requiredScopeB))
			Expect(ap2.Spec.Rules[i].When[1].Key).To(BeElementOf(testExpectedScopeKeys))
			Expect(ap2.Spec.Rules[i].When[1].Values[0]).To(BeElementOf(requiredScopeA, requiredScopeB))
		}
	})

	It("should produce two APs for a rule with two authorizations", func() {
		// given
		jwtRule := newJwtRuleBuilderWithDummyData().
			addJwtAuthorizationRequiredScopes("scope-a").
			addJwtAuthorizationRequiredScopes("scope-b").
			build()

		apiRule := newAPIRuleBuilderWithDummyData().withRules(jwtRule).build()
		svc := newServiceBuilderWithDummyData().build()
		gateway := newGatewayBuilderWithDummyData().build()
		client := getFakeClient(svc)
		processor := authorizationpolicy.NewProcessor(&testLogger, apiRule, gateway, client)

		// when
		result, err := processor.EvaluateReconciliation(context.Background(), client)

		// then
		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(2))

		ap1 := result[0].Obj.(*securityv1beta1.AuthorizationPolicy)
		ap2 := result[1].Obj.(*securityv1beta1.AuthorizationPolicy)

		Expect(ap1).NotTo(BeNil())
		Expect(ap1.ObjectMeta.Name).To(BeEmpty())
		Expect(ap1.ObjectMeta.GenerateName).To(Equal(apiRuleName + "-"))
		Expect(ap1.ObjectMeta.Namespace).To(Equal(apiRuleNamespace))

		Expect(ap1.Spec.Selector.MatchLabels["app"]).NotTo(BeNil())
		Expect(ap1.Spec.Selector.MatchLabels["app"]).To(Equal(serviceName))
		Expect(len(ap1.Spec.Rules)).To(Equal(3))
		Expect(len(ap1.Spec.Rules[0].From)).To(Equal(1))
		Expect(len(ap1.Spec.Rules[0].From[0].Source.RequestPrincipals)).To(Equal(1))
		Expect(ap1.Spec.Rules[0].From[0].Source.RequestPrincipals[0]).To(Equal(fmt.Sprintf("%s/*", jwtIssuer)))
		Expect(len(ap1.Spec.Rules[0].To)).To(Equal(1))
		Expect(len(ap1.Spec.Rules[0].To[0].Operation.Methods)).To(Equal(1))
		Expect(ap1.Spec.Rules[0].To[0].Operation.Methods).To(ContainElements([]string{http.MethodGet}))
		Expect(len(ap1.Spec.Rules[0].To[0].Operation.Paths)).To(Equal(1))

		for i := 0; i < 3; i++ {
			Expect(ap1.Spec.Rules[i].When[0].Key).To(BeElementOf(testExpectedScopeKeys))
			Expect(ap1.Spec.Rules[i].When[0].Values).To(HaveLen(1))
			Expect(ap1.Spec.Rules[i].When[0].Values[0]).To(BeElementOf(requiredScopeA, requiredScopeB))
		}

		Expect(ap2).NotTo(BeNil())
		Expect(ap2.ObjectMeta.Name).To(BeEmpty())
		Expect(ap2.ObjectMeta.GenerateName).To(Equal(apiRuleName + "-"))
		Expect(ap2.ObjectMeta.Namespace).To(Equal(apiRuleNamespace))

		Expect(ap2.Spec.Selector.MatchLabels["app"]).NotTo(BeNil())
		Expect(ap2.Spec.Selector.MatchLabels["app"]).To(Equal(serviceName))
		Expect(len(ap2.Spec.Rules)).To(Equal(3))
		Expect(len(ap2.Spec.Rules[0].From)).To(Equal(1))
		Expect(len(ap2.Spec.Rules[0].From[0].Source.RequestPrincipals)).To(Equal(1))
		Expect(ap2.Spec.Rules[0].From[0].Source.RequestPrincipals[0]).To(Equal(fmt.Sprintf("%s/*", jwtIssuer)))
		Expect(len(ap2.Spec.Rules[0].To)).To(Equal(1))
		Expect(len(ap2.Spec.Rules[0].To[0].Operation.Methods)).To(Equal(1))
		Expect(ap2.Spec.Rules[0].To[0].Operation.Methods).To(ContainElements([]string{http.MethodGet}))
		Expect(len(ap2.Spec.Rules[0].To[0].Operation.Paths)).To(Equal(1))

		for i := 0; i < 3; i++ {
			Expect(ap2.Spec.Rules[i].When[0].Key).To(BeElementOf(testExpectedScopeKeys))
			Expect(ap2.Spec.Rules[i].When[0].Values).To(HaveLen(1))
			Expect(ap2.Spec.Rules[i].When[0].Values[0]).To(BeElementOf(requiredScopeA, requiredScopeB))
		}
	})

	It("should produce AP from a rule with two issuers and one path", func() {
		// given
		testExpectedScopeKeys := []string{"request.auth.claims[scp]", "request.auth.claims[scope]", "request.auth.claims[scopes]"}
		ruleJwt := newJwtRuleBuilderWithDummyData().
			withPath("/headers").
			addJwtAuthorizationRequiredScopes("scope-a", "scope-b").
			addJwtAuthentication("https://oauth2.another.example.com/", "https://oauth2.another.example.com/.well-known/jwks.json").
			build()

		apiRule := newAPIRuleBuilderWithDummyData().withRules(ruleJwt).build()
		svc := newServiceBuilderWithDummyData().build()
		gateway := newGatewayBuilderWithDummyData().build()
		client := getFakeClient(svc)
		processor := authorizationpolicy.NewProcessor(&testLogger, apiRule, gateway, client)

		// when
		result, err := processor.EvaluateReconciliation(context.Background(), client)

		// then
		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(1))

		ap := result[0].Obj.(*securityv1beta1.AuthorizationPolicy)

		Expect(ap).NotTo(BeNil())
		Expect(ap.ObjectMeta.Name).To(BeEmpty())
		Expect(ap.ObjectMeta.GenerateName).To(Equal(apiRuleName + "-"))
		Expect(ap.ObjectMeta.Namespace).To(Equal(apiRuleNamespace))

		Expect(ap.Spec.Selector.MatchLabels["app"]).NotTo(BeNil())
		Expect(ap.Spec.Selector.MatchLabels["app"]).To(Equal(serviceName))
		Expect(len(ap.Spec.Rules)).To(Equal(3))
		Expect(len(ap.Spec.Rules[0].From)).To(Equal(1))
		Expect(len(ap.Spec.Rules[0].From[0].Source.RequestPrincipals)).To(Equal(2))
		Expect(ap.Spec.Rules[0].From[0].Source.RequestPrincipals[0]).To(Equal(fmt.Sprintf("%s/*", jwtIssuer)))
		Expect(ap.Spec.Rules[0].From[0].Source.RequestPrincipals[1]).To(Equal(fmt.Sprintf("%s/*", "https://oauth2.another.example.com/")))
		Expect(len(ap.Spec.Rules[0].To)).To(Equal(1))
		Expect(len(ap.Spec.Rules[0].To[0].Operation.Methods)).To(Equal(1))
		Expect(ap.Spec.Rules[0].To[0].Operation.Methods).To(ContainElements([]string{http.MethodGet}))
		Expect(len(ap.Spec.Rules[0].To[0].Operation.Paths)).To(Equal(1))
		Expect(ap.Spec.Rules[0].To[0].Operation.Paths).To(ContainElements("/headers"))

		for i := 0; i < 3; i++ {
			Expect(ap.Spec.Rules[i].When[0].Key).To(BeElementOf(testExpectedScopeKeys))
			Expect(ap.Spec.Rules[i].When).To(HaveLen(2))
			Expect(ap.Spec.Rules[i].When[0].Key).To(BeElementOf(testExpectedScopeKeys))
			Expect(ap.Spec.Rules[i].When[0].Values[0]).To(BeElementOf(requiredScopeA, requiredScopeB))
			Expect(ap.Spec.Rules[i].When[1].Key).To(BeElementOf(testExpectedScopeKeys))
			Expect(ap.Spec.Rules[i].When[1].Values[0]).To(BeElementOf(requiredScopeA, requiredScopeB))
		}

	})

	When("Two AP for different services with JWT handler exist", func() {
		It("should update APs and update principal when handler changed for one of the AP to noAuth", func() {
			// given: Cluster state
			beingUpdatedAp := getAuthorizationPolicy("being-updated-ap", apiRuleNamespace, "test-service", []string{"example-host.example.com"}, []string{http.MethodGet, http.MethodPost})
			jwtSecuredAp := getAuthorizationPolicy("jwt-secured-ap", apiRuleNamespace, "jwt-secured-service", []string{"example-host.example.com"}, []string{http.MethodGet, http.MethodPost})
			svc1 := newServiceBuilder().
				withName("test-service").
				withNamespace("example-namespace").
				addSelector("app", "test-service").
				build()

			svc2 := newServiceBuilder().
				withName("jwt-secured-service").
				withNamespace("example-namespace").
				addSelector("app", "jwt-secured-service").
				build()

			ctrlClient := getFakeClient(beingUpdatedAp, jwtSecuredAp, svc1, svc2)

			// given: New resources
			jwtRule := newJwtRuleBuilderWithDummyData().
				withPath("/").
				withMethods(http.MethodGet, http.MethodPost).
				withServiceName("jwt-secured-service").
				build()

			noAuthRule := newNoAuthRuleBuilderWithDummyData().
				withMethods(http.MethodGet, http.MethodPost).
				withServiceName("test-service").
				build()

			rules := []*gatewayv2alpha1.Rule{noAuthRule, jwtRule}

			apiRule := newAPIRuleBuilder().
				withName("test-apirule").
				withNamespace("example-namespace").
				withHost("example-host.example.com").
				withGateway("example-namespace/example-gateway").
				withRules(rules...).
				build()

			gateway := newGatewayBuilderWithDummyData().build()
			processor := authorizationpolicy.NewProcessor(&testLogger, apiRule, gateway, ctrlClient)

			// when
			result, err := processor.EvaluateReconciliation(context.Background(), ctrlClient)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(2))

			updatedNoopMatcher := getActionMatcher("update", apiRuleNamespace, "test-service", "Principals", ContainElements("cluster.local/ns/istio-system/sa/istio-ingressgateway-service-account"), ContainElements("GET", "POST"), ContainElements("/"), BeNil())
			updatedNotChangedMatcher := getActionMatcher("update", apiRuleNamespace, "jwt-secured-service", "RequestPrincipals", ContainElements("https://oauth2.example.com//*"), ContainElements("GET", "POST"), ContainElements("/"), BeNil())
			Expect(result).To(ContainElements(updatedNoopMatcher, updatedNotChangedMatcher))
		})

	})

	When("Rule with two JWT authorizations resulting in two APs exists", func() {
		It("should update both APs when audience is updated for both authorizations", func() {
			// given: Cluster state
			serviceName := "test-service"

			ap1 := getAuthorizationPolicy("ap1", apiRuleNamespace, serviceName, []string{"example-host.example.com"}, []string{"GET"})
			ap1.Spec.Rules[0].When = []*v1beta1.Condition{
				{
					Key:    "request.auth.claims[aud]",
					Values: []string{"audience1", "audience2"},
				},
			}

			ap2 := getAuthorizationPolicy("ap2", apiRuleNamespace, serviceName, []string{"example-host.example.com"}, []string{"GET"})
			ap2.Spec.Rules[0].When = []*v1beta1.Condition{
				{
					Key:    "request.auth.claims[aud]",
					Values: []string{"audience3"},
				},
			}
			// We need to set the index to 1 as this is expected to be the second authorization configured in the rule.
			ap2.Labels["gateway.kyma-project.io/index"] = "1"

			svc := newServiceBuilder().
				withName(serviceName).
				withNamespace("example-namespace").
				addSelector("app", serviceName).
				build()

			ctrlClient := getFakeClient(ap1, ap2, svc)

			// given: ApiRule with updated audiences in jwt authorizations
			jwtRule := newJwtRuleBuilderWithDummyData().
				withServiceName(serviceName).
				addJwtAuthorizationAudiences("audience1", "audience3").
				addJwtAuthorizationAudiences("audience5", "audience6").
				build()

			apiRule := newAPIRuleBuilderWithDummyData().
				withServiceName(serviceName).
				withRules(jwtRule).build()
			gateway := newGatewayBuilderWithDummyData().build()
			processor := authorizationpolicy.NewProcessor(&testLogger, apiRule, gateway, ctrlClient)

			// when
			result, err := processor.EvaluateReconciliation(context.Background(), ctrlClient)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(2))

			// It's expected that the hash is the same for all the objects, as the fields that were updated are not part of the hash.
			expectedHash := ap1.Labels["gateway.kyma-project.io/hash"]

			ap1Matcher := getAudienceMatcher("update", expectedHash, "0", []string{"audience1", "audience3"})
			ap2Matcher := getAudienceMatcher("update", expectedHash, "1", []string{"audience5", "audience6"})
			Expect(result).To(ContainElements(ap1Matcher, ap2Matcher))
		})

		It("should create new AP and update existing APs without changes when new authorization is added", func() {
			// given: Cluster state
			serviceName := "test-service"

			ap1 := getAuthorizationPolicy("ap1", apiRuleNamespace, serviceName, []string{"example-host.example.com"}, []string{"GET"})
			ap1.Spec.Rules[0].When = []*v1beta1.Condition{
				{
					Key:    "request.auth.claims[aud]",
					Values: []string{"audience1", "audience2"},
				},
			}

			ap2 := getAuthorizationPolicy("ap2", apiRuleNamespace, serviceName, []string{"example-host.example.com"}, []string{"GET"})
			ap2.Spec.Rules[0].When = []*v1beta1.Condition{
				{
					Key:    "request.auth.claims[aud]",
					Values: []string{"audience3"},
				},
			}
			// We need to set the index to 1 as this is expected to be the second authorization configured in the rule.
			ap2.Labels["gateway.kyma-project.io/index"] = "1"

			svc := newServiceBuilder().
				withName(serviceName).
				withNamespace("example-namespace").
				addSelector("app", serviceName).
				build()

			ctrlClient := getFakeClient(ap1, ap2, svc)

			// given: ApiRule with updated audiences in jwt authorizations
			jwtRule := newJwtRuleBuilderWithDummyData().
				withServiceName(serviceName).
				addJwtAuthorizationAudiences("audience1", "audience2").
				addJwtAuthorizationAudiences("audience3").
				addJwtAuthorizationAudiences("audience4").
				build()

			apiRule := newAPIRuleBuilderWithDummyData().withRules(jwtRule).build()
			gateway := newGatewayBuilderWithDummyData().build()
			processor := authorizationpolicy.NewProcessor(&testLogger, apiRule, gateway, ctrlClient)

			// when
			result, err := processor.EvaluateReconciliation(context.Background(), ctrlClient)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(3))

			// It's expected that the hash is the same for all the objects, as the fields that were updated are not part of the hash.
			expectedHash := ap1.Labels["gateway.kyma-project.io/hash"]

			ap1Matcher := getAudienceMatcher("update", expectedHash, "0", []string{"audience1", "audience2"})
			ap2Matcher := getAudienceMatcher("update", expectedHash, "1", []string{"audience3"})
			newApMatcher := getAudienceMatcher("create", expectedHash, "2", []string{"audience4"})
			Expect(result).To(ContainElements(ap1Matcher, ap2Matcher, newApMatcher))
		})

		It("should delete existing AP and update existing AP without changes when authorization is removed", func() {
			// given: Cluster state
			serviceName := "test-service"

			ap1 := getAuthorizationPolicy("ap1", apiRuleNamespace, serviceName, []string{"example-host.example.com"}, []string{"GET"})
			ap1.Spec.Rules[0].When = []*v1beta1.Condition{
				{
					Key:    "request.auth.claims[aud]",
					Values: []string{"audience1", "audience2"},
				},
			}

			ap2 := getAuthorizationPolicy("ap2", apiRuleNamespace, serviceName, []string{"example-host.example.com"}, []string{"GET"})
			ap2.Spec.Rules[0].When = []*v1beta1.Condition{
				{
					Key:    "request.auth.claims[aud]",
					Values: []string{"audience3"},
				},
			}
			// We need to set the index to 1 as this is expected to be the second authorization configured in the rule.
			ap2.Labels["gateway.kyma-project.io/index"] = "1"

			svc := newServiceBuilder().
				withName(serviceName).
				withNamespace("example-namespace").
				addSelector("app", serviceName).
				build()

			ctrlClient := getFakeClient(ap1, ap2, svc)

			// given: ApiRule with updated audiences in jwt authorizations
			jwtRule := newJwtRuleBuilderWithDummyData().
				withServiceName(serviceName).
				addJwtAuthorizationAudiences("audience1", "audience2").
				build()

			apiRule := newAPIRuleBuilderWithDummyData().withRules(jwtRule).build()
			gateway := newGatewayBuilderWithDummyData().build()
			processor := authorizationpolicy.NewProcessor(&testLogger, apiRule, gateway, ctrlClient)

			// when
			result, err := processor.EvaluateReconciliation(context.Background(), ctrlClient)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(2))

			// It's expected that the hash is the same for all the objects, as the fields that were updated are not part of the hash.
			expectedHash := ap1.Labels["gateway.kyma-project.io/hash"]

			ap1Matcher := getAudienceMatcher("update", expectedHash, "0", []string{"audience1", "audience2"})
			ap2Matcher := getAudienceMatcher("delete", expectedHash, "1", []string{"audience3"})
			Expect(result).To(ContainElements(ap1Matcher, ap2Matcher))
		})
	})

	When("Rule with three JWT authorizations resulting in three APs exists", func() {
		It("should update first two APs and delete third AP when first authorization is removed", func() {
			// given: Cluster state
			serviceName := "test-service"

			ap1 := getAuthorizationPolicy("ap1", apiRuleNamespace, serviceName, []string{"example-host.example.com"}, []string{"GET"})
			ap1.Spec.Rules[0].When = []*v1beta1.Condition{
				{
					Key:    "request.auth.claims[aud]",
					Values: []string{"audience1", "audience2"},
				},
			}

			ap2 := getAuthorizationPolicy("ap2", apiRuleNamespace, serviceName, []string{"example-host.example.com"}, []string{"GET"})
			ap2.Spec.Rules[0].When = []*v1beta1.Condition{
				{
					Key:    "request.auth.claims[aud]",
					Values: []string{"audience3"},
				},
			}
			// We need to set the index to 1 as this is expected to be the second authorization configured in the rule.
			ap2.Labels["gateway.kyma-project.io/index"] = "1"

			ap3 := getAuthorizationPolicy("ap3", apiRuleNamespace, serviceName, []string{"example-host.example.com"}, []string{"GET"})
			ap3.Spec.Rules[0].When = []*v1beta1.Condition{
				{
					Key:    "request.auth.claims[aud]",
					Values: []string{"audience4"},
				},
			}
			// We need to set the index to 1 as this is expected to be the second authorization configured in the rule.
			ap3.Labels["gateway.kyma-project.io/index"] = "2"

			svc := newServiceBuilder().
				withName(serviceName).
				withNamespace("example-namespace").
				addSelector("app", serviceName).
				build()

			ctrlClient := getFakeClient(ap1, ap2, ap3, svc)

			// given: ApiRule with updated audiences in jwt authorizations
			jwtRule := newJwtRuleBuilderWithDummyData().
				withServiceName(serviceName).
				addJwtAuthorizationAudiences("audience3").
				addJwtAuthorizationAudiences("audience4").
				build()

			apiRule := newAPIRuleBuilderWithDummyData().withRules(jwtRule).build()
			gateway := newGatewayBuilderWithDummyData().build()
			processor := authorizationpolicy.NewProcessor(&testLogger, apiRule, gateway, ctrlClient)

			// when
			result, err := processor.EvaluateReconciliation(context.Background(), ctrlClient)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(3))

			// It's expected that the hash is the same for all the objects, as the fields that were updated are not part of the hash.
			expectedHash := ap1.Labels["gateway.kyma-project.io/hash"]

			ap2Matcher := getAudienceMatcher("update", expectedHash, "0", []string{"audience3"})
			ap3Matcher := getAudienceMatcher("update", expectedHash, "1", []string{"audience4"})
			deletedMatcher := getAudienceMatcher("delete", expectedHash, "2", []string{"audience4"})
			Expect(result).To(ContainElements(ap2Matcher, ap3Matcher, deletedMatcher))
		})
	})

})
