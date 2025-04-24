package authorizationpolicy_test

import (
	"context"
	"fmt"
	"net/http"

	"github.com/kyma-project/api-gateway/internal/processing/processors/v2alpha1/authorizationpolicy"

	"github.com/kyma-project/api-gateway/internal/processing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"istio.io/api/security/v1beta1"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
)

var _ = Describe("Processing", func() {
	serviceName := "example-service"

	It("should set authorization policy path to `/*` when the Rule applies to all paths", func() {
		// given
		ruleJwt := newNoAuthRuleBuilderWithDummyData().
			withPath("/*").
			build()

		apiRule := newAPIRuleBuilderWithDummyData().
			withRules(ruleJwt).
			build()
		svc := newServiceBuilderWithDummyData().build()
		gateway := newGatewayBuilderWithDummyData().build()
		client := getFakeClient(svc)
		processor := authorizationpolicy.NewProcessor(&testLogger, apiRule, gateway)

		// when
		result, err := processor.EvaluateReconciliation(context.Background(), client)

		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(1))

		ap := result[0].Obj.(*securityv1beta1.AuthorizationPolicy)

		Expect(len(ap.Spec.Rules[0].To[0].Operation.Paths)).To(Equal(1))
		Expect(ap.Spec.Rules[0].To[0].Operation.Paths).To(ContainElement("/{**}"))
	})

	It("should produce one AP for a Rule without service, but service definition on ApiRule level", func() {
		// given
		ruleJwt := newNoAuthRuleBuilderWithDummyData().
			withPath("/headers").
			build()

		apiRule := newAPIRuleBuilderWithDummyData().
			withRules(ruleJwt).
			build()
		svc := newServiceBuilderWithDummyData().build()
		gateway := newGatewayBuilderWithDummyData().build()
		client := getFakeClient(svc)
		processor := authorizationpolicy.NewProcessor(&testLogger, apiRule, gateway)

		// when
		result, err := processor.EvaluateReconciliation(context.Background(), client)

		// then
		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(1))

		ap := result[0].Obj.(*securityv1beta1.AuthorizationPolicy)
		Expect(ap).NotTo(BeNil())
		// The AP should be in .Spec.Service.Namespace
		Expect(ap.Namespace).To(Equal(apiRuleNamespace))
		Expect(ap.Spec.Selector.MatchLabels["app"]).To(Equal(serviceName))
	})

	It("should produce AP with service from Rule, when service is configured on Rule and ApiRule level", func() {
		// given
		ruleServiceName := "rule-scope-example-service"
		specServiceNamespace := "spec-service-namespace"

		ruleJwt := newRuleBuilder().
			withPath("/").
			addMethods(http.MethodGet).
			withServiceName(ruleServiceName).
			withServicePort(8080).
			withNoAuth().
			build()

		apiRule := newAPIRuleBuilderWithDummyData().
			withRules(ruleJwt).
			withServiceNamespace(specServiceNamespace).
			build()
		svc := newServiceBuilder().
			withName(ruleServiceName).
			withNamespace(specServiceNamespace).
			addSelector("app", ruleServiceName).
			build()
		gateway := newGatewayBuilderWithDummyData().build()
		client := getFakeClient(svc)
		processor := authorizationpolicy.NewProcessor(&testLogger, apiRule, gateway)

		// when
		result, err := processor.EvaluateReconciliation(context.Background(), client)

		// then
		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(1))

		ap := result[0].Obj.(*securityv1beta1.AuthorizationPolicy)
		Expect(ap).NotTo(BeNil())
		// The RA should be in .Spec.Service.Namespace
		Expect(ap.Namespace).To(Equal(specServiceNamespace))
		Expect(ap.Spec.Selector.MatchLabels["app"]).To(Equal(ruleServiceName))
	})

	It("should produce one AP for a Rule with service with namespace, in the configured namespace", func() {
		// given
		ruleServiceName := "rule-scope-example-service"
		ruleServiceNamespace := "rule-service-namespace"

		ruleJwt := newNoAuthRuleBuilderWithDummyData().
			withPath("/headers").
			withServiceName(ruleServiceName).
			withServiceNamespace(ruleServiceNamespace).
			build()

		apiRule := newAPIRuleBuilderWithDummyData().
			withRules(ruleJwt).
			build()
		svc := newServiceBuilder().
			withName(ruleServiceName).
			withNamespace(ruleServiceNamespace).
			addSelector("app", ruleServiceName).
			build()
		gateway := newGatewayBuilderWithDummyData().build()
		client := getFakeClient(svc)
		processor := authorizationpolicy.NewProcessor(&testLogger, apiRule, gateway)

		// when
		result, err := processor.EvaluateReconciliation(context.Background(), client)

		// then
		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(1))

		ap := result[0].Obj.(*securityv1beta1.AuthorizationPolicy)
		Expect(ap).NotTo(BeNil())
		Expect(ap.Spec.Selector.MatchLabels["app"]).To(Equal(ruleServiceName))
		// The AP should be in .Service.Namespace
		Expect(ap.Namespace).To(Equal(ruleServiceNamespace))
		// And the OwnerLabel should point to APIRule namespace
		Expect(ap.Labels[processing.OwnerLabel]).ToNot(BeEmpty())
		Expect(ap.Labels[processing.OwnerLabel]).To(Equal(fmt.Sprintf("%s.%s", apiRule.Name, apiRule.Namespace)))
	})

	It("should create AP when no exists", func() {
		// given: New resources
		rule := newNoAuthRuleBuilderWithDummyData().build()
		apiRule := newAPIRuleBuilderWithDummyData().
			withRules(rule).
			build()
		svc := newServiceBuilderWithDummyData().build()
		gateway := newGatewayBuilderWithDummyData().build()
		client := getFakeClient(svc)

		processor := authorizationpolicy.NewProcessor(&testLogger, apiRule, gateway)

		// when
		result, err := processor.EvaluateReconciliation(context.Background(), client)

		// then
		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(1))
		Expect(result[0].Action.String()).To(Equal("create"))
	})

	It("should update AP when path, methods and service name didn't change", func() {
		// given: Cluster state
		existingAp := getAuthorizationPolicy("ap1", apiRuleNamespace, serviceName, []string{"example-host.example.com"}, []string{http.MethodGet, http.MethodPost})

		// given: New resources
		rule := newJwtRuleBuilderWithDummyData().
			withMethods(http.MethodGet, http.MethodPost).
			build()
		apiRule := newAPIRuleBuilderWithDummyData().
			withRules(rule).
			build()
		svc := newServiceBuilderWithDummyData().build()
		gateway := newGatewayBuilderWithDummyData().build()
		client := getFakeClient(existingAp, svc)

		processor := authorizationpolicy.NewProcessor(&testLogger, apiRule, gateway)

		// when
		result, err := processor.EvaluateReconciliation(context.Background(), client)

		// then
		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(1))

		updateMatcher := getActionMatcher("update", apiRuleNamespace, serviceName, "RequestPrincipals", ContainElements("https://oauth2.example.com//*"), ContainElements(http.MethodGet, http.MethodPost), ContainElements("/"), BeNil())
		Expect(result).To(ContainElements(updateMatcher))
	})

	It("should delete AP when there is no desired AP", func() {
		//given: Cluster state
		existingAp := getAuthorizationPolicy("ap1", apiRuleNamespace, serviceName, []string{"example-host.example.com"}, []string{http.MethodGet, http.MethodPost})

		// given: New resources
		apiRule := newAPIRuleBuilderWithDummyData().build()
		svc := newServiceBuilderWithDummyData().build()
		ctrlClient := getFakeClient(existingAp, svc)
		gateway := newGatewayBuilderWithDummyData().build()
		processor := authorizationpolicy.NewProcessor(&testLogger, apiRule, gateway)

		// when
		result, err := processor.EvaluateReconciliation(context.Background(), ctrlClient)

		// then
		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(1))

		resultMatcher := getActionMatcher("delete", apiRuleNamespace, serviceName, "RequestPrincipals", ContainElements("*"), ContainElements(http.MethodGet, http.MethodPost), ContainElements("/"), BeNil())
		Expect(result).To(ContainElements(resultMatcher))
	})

	When("AP with RuleTo exists", func() {
		It("should create new AP and delete old AP when new rule with same methods and service but different path is added to ApiRule", func() {
			// given: Cluster state
			existingAp := getAuthorizationPolicy("ap1", apiRuleNamespace, serviceName, []string{"example-host.example.com"}, []string{http.MethodGet, http.MethodPost})
			svc := newServiceBuilderWithDummyData().build()
			ctrlClient := getFakeClient(existingAp, svc)

			// given: New resources
			existingRule := newJwtRuleBuilderWithDummyData().
				withMethods(http.MethodGet, http.MethodPost).
				build()
			newRule := newJwtRuleBuilderWithDummyData().
				withPath("/new-path").
				withMethods(http.MethodGet, http.MethodPost).
				build()
			apiRule := newAPIRuleBuilderWithDummyData().
				withRules(existingRule, newRule).
				build()
			gateway := newGatewayBuilderWithDummyData().build()
			processor := authorizationpolicy.NewProcessor(&testLogger, apiRule, gateway)

			// when
			result, err := processor.EvaluateReconciliation(context.Background(), ctrlClient)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(3))

			newApMatcher := getActionMatcher("create", apiRuleNamespace, serviceName, "RequestPrincipals", ContainElements("https://oauth2.example.com//*"), ContainElements(http.MethodGet, http.MethodPost), ContainElements("/"), ContainElements("/new-path"))
			newApMatcher2 := getActionMatcher("create", apiRuleNamespace, serviceName, "RequestPrincipals", ContainElements("https://oauth2.example.com//*"), ContainElements(http.MethodGet, http.MethodPost), ContainElements("/new-path"), BeNil())
			deleteApMatcher := getActionMatcher("delete", apiRuleNamespace, serviceName, "RequestPrincipals", ContainElements("*"), ContainElements(http.MethodGet, http.MethodPost), ContainElements("/"), BeNil())
			Expect(result).To(ContainElements(newApMatcher, newApMatcher2, deleteApMatcher))
		})

		It("should create new AP and update existing AP when new rule with same path and service but different methods is added to ApiRule", func() {
			// given: Cluster state
			existingAp := getAuthorizationPolicy("ap1", apiRuleNamespace, serviceName, []string{"example-host.example.com"}, []string{http.MethodGet, http.MethodPost})
			svc := newServiceBuilderWithDummyData().build()
			ctrlClient := getFakeClient(existingAp, svc)

			// given: New resources
			existingRule := newJwtRuleBuilderWithDummyData().
				withMethods(http.MethodGet, http.MethodPost).
				build()
			newRule := newJwtRuleBuilderWithDummyData().
				withMethods(http.MethodDelete).
				build()
			apiRule := newAPIRuleBuilderWithDummyData().
				withRules(existingRule, newRule).
				build()
			gateway := newGatewayBuilderWithDummyData().build()
			processor := authorizationpolicy.NewProcessor(&testLogger, apiRule, gateway)

			// when
			result, err := processor.EvaluateReconciliation(context.Background(), ctrlClient)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(2))

			updateExistingApMatcher := getActionMatcher("update", apiRuleNamespace, serviceName, "RequestPrincipals", ContainElements("https://oauth2.example.com//*"), ContainElements(http.MethodGet, http.MethodPost), ContainElements("/"), BeNil())
			newApMatcher := getActionMatcher("create", apiRuleNamespace, serviceName, "RequestPrincipals", ContainElements("https://oauth2.example.com//*"), ContainElements(http.MethodDelete), ContainElements("/"), BeNil())
			Expect(result).To(ContainElements(updateExistingApMatcher, newApMatcher))
		})

		It("should create new AP and update existing AP when new rule with same path and methods, but different service is added to ApiRule", func() {
			//given: Cluster state
			existingAp := getAuthorizationPolicy("ap1", apiRuleNamespace, serviceName, []string{"example-host.example.com"}, []string{http.MethodGet, http.MethodPost})
			// given: New resources
			existingRule := newJwtRuleBuilderWithDummyData().
				withMethods(http.MethodGet, http.MethodPost).
				build()
			newRule := newJwtRuleBuilderWithDummyData().
				withMethods(http.MethodGet, http.MethodPost).
				withServiceName("new-service").
				build()
			apiRule := newAPIRuleBuilderWithDummyData().
				withRules(existingRule, newRule).
				build()
			svc1 := newServiceBuilderWithDummyData().build()
			svc2 := newServiceBuilder().
				withName("new-service").
				withNamespace(apiRuleNamespace).
				addSelector("app", "new-service").
				build()
			ctrlClient := getFakeClient(existingAp, svc1, svc2)
			gateway := newGatewayBuilderWithDummyData().build()
			processor := authorizationpolicy.NewProcessor(&testLogger, apiRule, gateway)

			// when
			result, err := processor.EvaluateReconciliation(context.Background(), ctrlClient)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(2))

			updateExistingApMatcher := getActionMatcher("update", apiRuleNamespace, serviceName, "RequestPrincipals", ContainElements("https://oauth2.example.com//*"), ContainElements(http.MethodGet, http.MethodPost), ContainElements("/"), BeNil())
			newApMatcher := getActionMatcher("create", apiRuleNamespace, "new-service", "RequestPrincipals", ContainElements("https://oauth2.example.com//*"), ContainElements(http.MethodGet, http.MethodPost), ContainElements("/"), BeNil())
			Expect(result).To(ContainElements(updateExistingApMatcher, newApMatcher))
		})

		It("should recreate AP when path in ApiRule changed", func() {
			// given: Cluster state
			existingAp := getAuthorizationPolicy("ap1", apiRuleNamespace, serviceName, []string{"example-host.example.com"}, []string{http.MethodGet, http.MethodPost})
			svc := newServiceBuilderWithDummyData().build()
			ctrlClient := getFakeClient(existingAp, svc)

			// given: New resources
			rule := newJwtRuleBuilderWithDummyData().
				withMethods(http.MethodGet, http.MethodPost).
				withPath("/new-path").
				build()
			apiRule := newAPIRuleBuilderWithDummyData().
				withRules(rule).
				build()
			gateway := newGatewayBuilderWithDummyData().build()
			processor := authorizationpolicy.NewProcessor(&testLogger, apiRule, gateway)

			// when
			result, err := processor.EvaluateReconciliation(context.Background(), ctrlClient)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(2))

			existingApMatcher := getActionMatcher("delete", apiRuleNamespace, serviceName, "RequestPrincipals", ContainElements("*"), ContainElements(http.MethodGet, http.MethodPost), ContainElements("/"), BeNil())
			newApMatcher := getActionMatcher("create", apiRuleNamespace, serviceName, "RequestPrincipals", ContainElements("https://oauth2.example.com//*"), ContainElements(http.MethodGet, http.MethodPost), ContainElements("/new-path"), BeNil())
			Expect(result).To(ContainElements(existingApMatcher, newApMatcher))
		})

	})

	When("Two AP with different methods for same path and service exist", func() {
		It("should delete and create AP, when path has changed", func() {
			// given: Cluster state
			unchangedAp := getAuthorizationPolicy("unchanged-ap", apiRuleNamespace, serviceName, []string{"example-host.example.com"}, []string{http.MethodDelete})
			toBeUpdateAp := getAuthorizationPolicy("to-be-updated-ap", apiRuleNamespace, serviceName, []string{"example-host.example.com"}, []string{http.MethodGet})
			svc := newServiceBuilderWithDummyData().build()
			ctrlClient := getFakeClient(toBeUpdateAp, unchangedAp, svc)

			// given: New resources
			unchangedRule := newJwtRuleBuilderWithDummyData().
				withMethods(http.MethodDelete).
				build()
			updatedRule := newJwtRuleBuilderWithDummyData().
				withPath("/new-path").
				build()

			apiRule := newAPIRuleBuilderWithDummyData().
				withRules(unchangedRule, updatedRule).
				build()
			gateway := newGatewayBuilderWithDummyData().build()
			processor := authorizationpolicy.NewProcessor(&testLogger, apiRule, gateway)

			// when
			result, err := processor.EvaluateReconciliation(context.Background(), ctrlClient)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(4))

			createMatcher := getActionMatcher("create", apiRuleNamespace, serviceName, "RequestPrincipals", ContainElements("https://oauth2.example.com//*"), ContainElements(http.MethodDelete), ContainElements("/"), ContainElements("/new-path"))
			createMatcher2 := getActionMatcher("create", apiRuleNamespace, serviceName, "RequestPrincipals", ContainElements("https://oauth2.example.com//*"), ContainElements(http.MethodGet), ContainElements("/new-path"), BeNil())
			deleteMatcher := getActionMatcher("delete", apiRuleNamespace, serviceName, "RequestPrincipals", ContainElements("*"), ContainElements(http.MethodGet), ContainElements("/"), BeNil())
			deleteMatcher2 := getActionMatcher("delete", apiRuleNamespace, serviceName, "RequestPrincipals", ContainElements("*"), ContainElements(http.MethodDelete), ContainElements("/"), BeNil())
			Expect(result).To(ContainElements(createMatcher, createMatcher2, deleteMatcher, deleteMatcher2))
		})
	})

	When("Namespace changes", func() {
		It("should create new AP in new namespace and delete old AP when namespace is on APIRule spec level", func() {
			// given: Cluster state
			oldAP := getAuthorizationPolicy("unchanged-ap", apiRuleNamespace, serviceName, []string{"example-host.example.com"}, []string{http.MethodDelete})

			svc := newServiceBuilderWithDummyData().build()
			specNewServiceNamespace := "new-namespace"
			svcNewNS := newServiceBuilderWithDummyData().
				withNamespace(specNewServiceNamespace).
				build()
			ctrlClient := getFakeClient(oldAP, svc, svcNewNS)

			// given: New resources
			movedRule := newRuleBuilder().
				withPath("/").
				addMethods(http.MethodDelete).
				addJwtAuthentication("https://oauth2.example.com/", "https://oauth2.example.com/.well-known/jwks.json").
				build()

			apiRule := newAPIRuleBuilderWithDummyData().
				withRules(movedRule).
				withServiceNamespace(specNewServiceNamespace).
				build()
			gateway := newGatewayBuilderWithDummyData().build()
			processor := authorizationpolicy.NewProcessor(&testLogger, apiRule, gateway)

			// when
			result, err := processor.EvaluateReconciliation(context.Background(), ctrlClient)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(2))

			deleteMatcher := getActionMatcher("delete", apiRuleNamespace, serviceName, "RequestPrincipals", ContainElements("*"), ContainElements(http.MethodDelete), ContainElements("/"), BeNil())
			createMatcher := getActionMatcher("create", specNewServiceNamespace, serviceName, "RequestPrincipals", ContainElements("https://oauth2.example.com//*"), ContainElements(http.MethodDelete), ContainElements("/"), BeNil())
			Expect(result).To(ContainElements(deleteMatcher, createMatcher))
		})

		It("should create new AP in new namespace and delete old AP when namespace on rule level", func() {
			// given: Cluster state
			oldAP := getAuthorizationPolicy("unchanged-ap", apiRuleNamespace, serviceName, []string{"example-host.example.com"}, []string{http.MethodDelete})
			ruleServiceNamespace := "new-namespace"
			svc := newServiceBuilderWithDummyData().
				withNamespace(ruleServiceNamespace).
				build()
			ctrlClient := getFakeClient(oldAP, svc)

			// given: New resources
			movedRule := newJwtRuleBuilderWithDummyData().
				withMethods(http.MethodDelete).
				withServiceNamespace(ruleServiceNamespace).
				build()

			apiRule := newAPIRuleBuilderWithDummyData().
				withRules(movedRule).
				build()
			gateway := newGatewayBuilderWithDummyData().build()
			processor := authorizationpolicy.NewProcessor(&testLogger, apiRule, gateway)

			// when
			result, err := processor.EvaluateReconciliation(context.Background(), ctrlClient)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(2))

			deleteMatcher := getActionMatcher("delete", apiRuleNamespace, serviceName, "RequestPrincipals", ContainElements("*"), ContainElements(http.MethodDelete), ContainElements("/"), BeNil())
			createMatcher := getActionMatcher("create", ruleServiceNamespace, serviceName, "RequestPrincipals", ContainElements("https://oauth2.example.com//*"), ContainElements(http.MethodDelete), ContainElements("/"), BeNil())
			Expect(result).To(ContainElements(deleteMatcher, createMatcher))
		})
	})

	When("Two AP with same RuleTo for different services exist", func() {
		It("should delete and create AP, when path has changed", func() {
			// given: Cluster state
			unchangedAp := getAuthorizationPolicy("unchanged-ap", apiRuleNamespace, "first-service", []string{"example-host.example.com"}, []string{http.MethodGet})
			toBeUpdateAp := getAuthorizationPolicy("to-be-updated-ap", apiRuleNamespace, "second-service", []string{"example-host.example.com"}, []string{http.MethodGet})
			svc1 := newServiceBuilder().
				withName("first-service").
				withNamespace(apiRuleNamespace).
				addSelector("app", "first-service").
				build()
			svc2 := newServiceBuilder().
				withName("second-service").
				withNamespace(apiRuleNamespace).
				addSelector("app", "second-service").
				build()
			ctrlClient := getFakeClient(toBeUpdateAp, unchangedAp, svc1, svc2)

			// given: New resources
			unchangedRule := newJwtRuleBuilderWithDummyData().
				withServiceName("first-service").
				build()
			updatedRule := newJwtRuleBuilderWithDummyData().
				withServiceName("second-service").
				withPath("/new-path").
				withServiceName("second-service").
				build()
			apiRule := newAPIRuleBuilderWithDummyData().
				withRules(unchangedRule, updatedRule).
				build()
			gateway := newGatewayBuilderWithDummyData().build()
			processor := authorizationpolicy.NewProcessor(&testLogger, apiRule, gateway)

			// when
			result, err := processor.EvaluateReconciliation(context.Background(), ctrlClient)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(4))

			createdApMatcher := getActionMatcher("create", apiRuleNamespace, "second-service", "RequestPrincipals", ContainElements("https://oauth2.example.com//*"), ContainElements(http.MethodGet), ContainElements("/new-path"), BeNil())
			createdApMatcher2 := getActionMatcher("create", apiRuleNamespace, "first-service", "RequestPrincipals", ContainElements("https://oauth2.example.com//*"), ContainElements(http.MethodGet), ContainElements("/"), ContainElements("/new-path"))
			deleteMatcher := getActionMatcher("delete", apiRuleNamespace, "second-service", "RequestPrincipals", ContainElements("*"), ContainElements(http.MethodGet), ContainElements("/"), BeNil())
			deleteMatcher2 := getActionMatcher("delete", apiRuleNamespace, "first-service", "RequestPrincipals", ContainElements("*"), ContainElements(http.MethodGet), ContainElements("/"), BeNil())
			Expect(result).To(ContainElements(createdApMatcher, createdApMatcher2, deleteMatcher, deleteMatcher2))
		})
	})

	When("Service has custom selector spec", func() {
		It("should create AP with selector from service", func() {
			// given: New resources
			rule := newJwtRuleBuilderWithDummyData().build()
			apiRule := newAPIRuleBuilderWithDummyData().
				withRules(rule).
				build()
			svc := newServiceBuilder().
				withName("example-service").
				withNamespace("example-namespace").
				addSelector("custom", "example-service").
				build()
			client := getFakeClient(svc)
			gateway := newGatewayBuilderWithDummyData().build()

			processor := authorizationpolicy.NewProcessor(&testLogger, apiRule, gateway)

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
			differentNamespace := "different-namespace"

			rule := newJwtRuleBuilderWithDummyData().
				withServiceNamespace(differentNamespace).
				build()
			apiRule := newAPIRuleBuilderWithDummyData().
				withRules(rule).
				build()
			svc := newServiceBuilder().
				withName("example-service").
				withNamespace(differentNamespace).
				addSelector("custom", "example-service").
				build()
			client := getFakeClient(svc)
			gateway := newGatewayBuilderWithDummyData().build()

			processor := authorizationpolicy.NewProcessor(&testLogger, apiRule, gateway)

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
			rule := newJwtRuleBuilderWithDummyData().build()
			apiRule := newAPIRuleBuilderWithDummyData().
				withRules(rule).
				build()
			svc := newServiceBuilder().
				withName("example-service").
				withNamespace("example-namespace").
				addSelector("custom", "example-service").
				addSelector("second-custom", "blah").
				build()
			client := getFakeClient(svc)
			gateway := newGatewayBuilderWithDummyData().build()

			processor := authorizationpolicy.NewProcessor(&testLogger, apiRule, gateway)

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
			serviceName := serviceName

			ap := getAuthorizationPolicy("ap", apiRuleNamespace, serviceName, []string{"example-host.example.com"}, []string{http.MethodGet})
			ap.Spec.Rules[0].When = []*v1beta1.Condition{
				{
					Key:    "request.auth.claims[aud]",
					Values: []string{"audience1"},
				},
			}

			// We need to store the hash for comparison later
			expectedHash := ap.Labels["gateway.kyma-project.io/hash"]
			delete(ap.Labels, missingLabel)

			svc := newServiceBuilderWithDummyData().build()
			ctrlClient := getFakeClient(ap, svc)

			// given: ApiRule with updated audiences in jwt authorizations
			rule := newJwtRuleBuilderWithDummyData().
				addJwtAuthorizationAudiences("audience1").
				build()
			apiRule := newAPIRuleBuilderWithDummyData().
				withRules(rule).
				build()
			gateway := newGatewayBuilderWithDummyData().build()
			processor := authorizationpolicy.NewProcessor(&testLogger, apiRule, gateway)

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
