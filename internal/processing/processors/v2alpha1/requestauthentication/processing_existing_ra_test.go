package requestauthentication_test

import (
	"context"
	"fmt"
	"github.com/kyma-project/api-gateway/internal/processing/processors/v2alpha1/requestauthentication"
	"net/http"

	"github.com/kyma-project/api-gateway/internal/processing"
	"istio.io/api/security/v1beta1"
	typev1beta1 "istio.io/api/type/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	gomegatypes "github.com/onsi/gomega/types"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
)

var _ = Describe("Processing with existing RequestAuthentication", func() {

	getRequestAuthentication := func(name string, serviceName string, jwksUri string, issuer string) securityv1beta1.RequestAuthentication {
		return securityv1beta1.RequestAuthentication{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: apiRuleNamespace,
				Labels: map[string]string{
					processing.OwnerLabel: fmt.Sprintf("%s.%s", apiRuleName, apiRuleNamespace),
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

	It("should delete RA when there is no rule configured in ApiRule", func() {
		// given: Cluster state
		existingRa := getRequestAuthentication("raName", serviceName, jwksUri, jwtIssuer)

		ctrlClient := getFakeClient(&existingRa)

		// given: New resources
		apiRule := newAPIRuleBuilderWithDummyData().build()
		processor := requestauthentication.NewProcessor(apiRule)

		// when
		result, err := processor.EvaluateReconciliation(context.Background(), ctrlClient)

		// then
		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(1))
		Expect(result[0].Action.String()).To(Equal("delete"))
	})

	When("one resource with JWT config exists", func() {

		It("should update RA when nothing changed", func() {
			// given: Cluster state
			existingRa := getRequestAuthentication("raName", "example-service", jwksUri, jwtIssuer)
			svc := newServiceBuilderWithDummyData().build()
			ctrlClient := getFakeClient(&existingRa, svc)

			// given: New resources
			jwtRule := newJwtRuleBuilderWithDummyData().build()
			apiRule := newAPIRuleBuilderWithDummyData().
				withRules(jwtRule).
				build()
			processor := requestauthentication.NewProcessor(apiRule)

			// when
			result, err := processor.EvaluateReconciliation(context.Background(), ctrlClient)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(1))
			Expect(result[0].Action.String()).To(Equal("update"))
		})

		It("should delete and create new RA when only service name in JWT Rule has changed", func() {
			// given: Cluster state
			existingRa := getRequestAuthentication("raName", "old-service", jwksUri, jwtIssuer)
			svcOld := newServiceBuilder().
				withName("old-service").
				withNamespace("example-namespace").
				addSelector("app", "old-service").
				build()
			svcUpdated := newServiceBuilder().
				withName("updated-service").
				withNamespace("example-namespace").
				addSelector("app", "updated-service").
				build()
			ctrlClient := getFakeClient(&existingRa, svcOld, svcUpdated)

			// given: New resources
			jwtRule := newJwtRuleBuilderWithDummyData().
				withServiceName("updated-service").
				build()
			apiRule := newAPIRuleBuilderWithDummyData().
				withRules(jwtRule).
				build()
			processor := requestauthentication.NewProcessor(apiRule)

			// when
			result, err := processor.EvaluateReconciliation(context.Background(), ctrlClient)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(2))

			deleteMatcher := getActionMatcher("delete", apiRuleNamespace, "old-service", jwksUri, jwtIssuer)
			createMatcher := getActionMatcher("create", apiRuleNamespace, "updated-service", jwksUri, jwtIssuer)
			Expect(result).To(ContainElements(deleteMatcher, createMatcher))
		})

		It("should create new RA when new service with new JWT config is added to ApiRule", func() {
			// given: Cluster state
			existingRa := getRequestAuthentication("raName", "existing-service", jwksUri, jwtIssuer)
			svcExisting := newServiceBuilder().
				withName("existing-service").
				withNamespace("example-namespace").
				addSelector("app", "existing-service").
				build()
			svcNew := newServiceBuilder().
				withName("new-service").
				withNamespace("example-namespace").
				addSelector("app", "new-service").
				build()
			ctrlClient := getFakeClient(&existingRa, svcExisting, svcNew)

			// given: New resources
			existingJwtRule := newJwtRuleBuilderWithDummyData().
				withServiceName("existing-service").
				build()

			newJwtRule := newRuleBuilder().
				withPath("/").
				addMethods(http.MethodGet).
				withServiceName("new-service").
				withServiceNamespace("example-namespace").
				withServicePort(8080).
				addJwtAuthentication(anotherJwtIssuer, anotherJwksUri).
				build()
			apiRule := newAPIRuleBuilderWithDummyData().
				withRules(existingJwtRule, newJwtRule).
				build()
			processor := requestauthentication.NewProcessor(apiRule)

			// when
			result, err := processor.EvaluateReconciliation(context.Background(), ctrlClient)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(2))

			updateResultMatcher := getActionMatcher("update", apiRuleNamespace, "existing-service", jwksUri, jwtIssuer)
			createResultMatcher := getActionMatcher("create", apiRuleNamespace, "new-service", anotherJwksUri, anotherJwtIssuer)
			Expect(result).To(ContainElements(createResultMatcher, updateResultMatcher))
		})

		It("should create new RA and delete old RA when JWT ApiRule has new JWKS URI", func() {
			// given: Cluster state
			existingRa := getRequestAuthentication("raName", "example-service", jwksUri, jwtIssuer)
			svc := newServiceBuilderWithDummyData().build()
			ctrlClient := getFakeClient(&existingRa, svc)

			// given: New resources
			jwtRule := newRuleBuilder().
				withPath("/").
				addMethods(http.MethodGet).
				withServiceName("example-service").
				withServiceNamespace("example-namespace").
				withServicePort(8080).
				addJwtAuthentication(jwtIssuer, anotherJwksUri).
				build()
			apiRule := newAPIRuleBuilderWithDummyData().
				withRules(jwtRule).
				build()
			processor := requestauthentication.NewProcessor(apiRule)

			// when
			result, err := processor.EvaluateReconciliation(context.Background(), ctrlClient)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(2))

			deleteResultMatcher := getActionMatcher("delete", apiRuleNamespace, "example-service", jwksUri, jwtIssuer)
			createResultMatcher := getActionMatcher("create", apiRuleNamespace, "example-service", anotherJwksUri, jwtIssuer)

			Expect(result).To(ContainElements(deleteResultMatcher, createResultMatcher))
		})
	})

	When("Two resources with same JWT config for different services exist", func() {

		It("should update RAs and create new RA for first-service and delete old RA when JWT issuer in JWT Rule for first-service has changed", func() {
			// given: Cluster state
			firstServiceRa := getRequestAuthentication("firstRa", "first-service", jwksUri, jwtIssuer)
			secondServiceRa := getRequestAuthentication("secondRa", "second-service", jwksUri, jwtIssuer)
			svcFirst := newServiceBuilder().
				withName("first-service").
				withNamespace("example-namespace").
				addSelector("app", "first-service").
				build()
			svcSecond := newServiceBuilder().
				withName("second-service").
				withNamespace("example-namespace").
				addSelector("app", "second-service").
				build()
			ctrlClient := getFakeClient(&firstServiceRa, &secondServiceRa, svcFirst, svcSecond)

			// given: New resources
			firstJwtRule := newRuleBuilder().
				withPath("/").
				addMethods(http.MethodGet).
				withServiceName("first-service").
				withServiceNamespace("example-namespace").
				withServicePort(8080).
				addJwtAuthentication(anotherJwtIssuer, jwksUri).
				build()
			secondJwtRule := newRuleBuilder().
				withPath("/").
				addMethods(http.MethodGet).
				withServiceName("second-service").
				withServiceNamespace("example-namespace").
				withServicePort(8080).
				addJwtAuthentication(jwtIssuer, jwksUri).
				build()

			apiRule := newAPIRuleBuilderWithDummyData().
				withRules(firstJwtRule, secondJwtRule).
				build()
			processor := requestauthentication.NewProcessor(apiRule)

			// when
			result, err := processor.EvaluateReconciliation(context.Background(), ctrlClient)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(3))

			deleteFirstServiceRaResultMatcher := getActionMatcher("delete", apiRuleNamespace, "first-service", jwksUri, jwtIssuer)
			createFirstServiceRaResultMatcher := getActionMatcher("create", apiRuleNamespace, "first-service", jwksUri, anotherJwtIssuer)
			secondRaResultMatcher := getActionMatcher("update", apiRuleNamespace, "second-service", jwksUri, jwtIssuer)
			Expect(result).To(ContainElements(deleteFirstServiceRaResultMatcher, createFirstServiceRaResultMatcher, secondRaResultMatcher))
		})

		It("should delete only first-service RA when it was removed from ApiRule", func() {
			// given: Cluster state
			firstServiceRa := getRequestAuthentication("firstRa", "first-service", jwksUri, jwtIssuer)
			secondServiceRa := getRequestAuthentication("secondRa", "second-service", jwksUri, jwtIssuer)
			svcFirst := newServiceBuilder().
				withName("first-service").
				withNamespace("example-namespace").
				addSelector("app", "first-service").
				build()
			svcSecond := newServiceBuilder().
				withName("second-service").
				withNamespace("example-namespace").
				addSelector("app", "second-service").
				build()
			ctrlClient := getFakeClient(&firstServiceRa, &secondServiceRa, svcFirst, svcSecond)

			// given: New resources
			secondJwtRule := newJwtRuleBuilderWithDummyData().
				withServiceName("second-service").
				build()
			apiRule := newAPIRuleBuilderWithDummyData().
				withRules(secondJwtRule).
				build()
			processor := requestauthentication.NewProcessor(apiRule)

			// when
			result, err := processor.EvaluateReconciliation(context.Background(), ctrlClient)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(2))

			deleteResultMatcher := getActionMatcher("delete", apiRuleNamespace, "first-service", jwksUri, jwtIssuer)
			updateResultMatcher := getActionMatcher("update", apiRuleNamespace, "second-service", jwksUri, jwtIssuer)
			Expect(result).To(ContainElements(deleteResultMatcher, updateResultMatcher))
		})

		It("should create new RA when it has different service", func() {
			// given: Cluster state
			firstServiceRa := getRequestAuthentication("firstRa", "first-service", jwksUri, jwtIssuer)
			secondServiceRa := getRequestAuthentication("secondRa", "second-service", jwksUri, jwtIssuer)
			svcFirst := newServiceBuilder().
				withName("first-service").
				withNamespace("example-namespace").
				addSelector("app", "first-service").
				build()
			svcSecond := newServiceBuilder().
				withName("second-service").
				withNamespace("example-namespace").
				addSelector("app", "second-service").
				build()
			svcNew := newServiceBuilder().
				withName("new-service").
				withNamespace("example-namespace").
				addSelector("app", "new-service").
				build()
			ctrlClient := getFakeClient(&firstServiceRa, &secondServiceRa, svcFirst, svcSecond, svcNew)

			// given: New resources
			firstJwtRule := newJwtRuleBuilderWithDummyData().
				withServiceName("first-service").
				build()
			secondJwtRule := newJwtRuleBuilderWithDummyData().
				withServiceName("second-service").
				build()
			newJwtRule := newJwtRuleBuilderWithDummyData().
				withServiceName("new-service").
				build()
			apiRule := newAPIRuleBuilderWithDummyData().
				withRules(firstJwtRule, secondJwtRule, newJwtRule).
				build()
			processor := requestauthentication.NewProcessor(apiRule)

			// when
			result, err := processor.EvaluateReconciliation(context.Background(), ctrlClient)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(3))

			firstRaMatcher := getActionMatcher("update", apiRuleNamespace, "first-service", jwksUri, jwtIssuer)
			secondRaMatcher := getActionMatcher("update", apiRuleNamespace, "second-service", jwksUri, jwtIssuer)
			newRaMatcher := getActionMatcher("create", apiRuleNamespace, "new-service", jwksUri, jwtIssuer)
			Expect(result).To(ContainElements(firstRaMatcher, secondRaMatcher, newRaMatcher))
		})

		It("should delete and create new RA when it has different namespace on spec level", func() {
			// given: Cluster state
			oldRa := getRequestAuthentication("firstRa", "old-service", jwksUri, jwtIssuer)
			svcOld := newServiceBuilder().
				withName("old-service").
				withNamespace("example-namespace").
				addSelector("app", "old-service").
				build()
			svcNewNS := newServiceBuilder().
				withName("old-service").
				withNamespace("new-namespace").
				addSelector("app", "old-service").
				build()
			ctrlClient := getFakeClient(&oldRa, svcOld, svcNewNS)

			// given: New resources
			jwtRule := newRuleBuilder().
				withPath("/").
				addMethods(http.MethodGet).
				withServiceName("old-service").
				withServicePort(8080).
				addJwtAuthentication(jwtIssuer, jwksUri).
				build()

			apiRule := newAPIRuleBuilder().
				withName(apiRuleName).
				withNamespace(apiRuleNamespace).
				withHost("example-host.example.com").
				withGateway("example-namespace/example-gateway").
				withServiceNamespace("new-namespace").
				withRules(jwtRule).
				build()
			processor := requestauthentication.NewProcessor(apiRule)

			// when
			result, err := processor.EvaluateReconciliation(context.Background(), ctrlClient)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(2))

			deletionMatcher := getActionMatcher("delete", apiRuleNamespace, "old-service", jwksUri, jwtIssuer)
			creationMatcher := getActionMatcher("create", "new-namespace", "old-service", jwksUri, jwtIssuer)
			Expect(result).To(ContainElements(deletionMatcher, creationMatcher))
		})

		It("should delete and create new RA when it has different namespace on rule level", func() {
			// given: Cluster state
			oldRa := getRequestAuthentication("firstRa", "old-service", jwksUri, jwtIssuer)
			svcOld := newServiceBuilder().
				withName("old-service").
				withNamespace("example-namespace").
				addSelector("app", "old-service").
				build()
			svcNewNS := newServiceBuilder().
				withName("old-service").
				withNamespace("new-namespace").
				addSelector("app", "old-service").
				build()
			ctrlClient := getFakeClient(&oldRa, svcOld, svcNewNS)

			// given: New resources
			jwtRule := newRuleBuilder().
				withPath("/").
				addMethods(http.MethodGet).
				withServiceName("old-service").
				withServiceNamespace("new-namespace").
				withServicePort(8080).
				addJwtAuthentication(jwtIssuer, jwksUri).
				build()
			apiRule := newAPIRuleBuilderWithDummyData().
				withRules(jwtRule).
				build()
			processor := requestauthentication.NewProcessor(apiRule)

			// when
			result, err := processor.EvaluateReconciliation(context.Background(), ctrlClient)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(2))

			deletionMatcher := getActionMatcher("delete", apiRuleNamespace, "old-service", jwksUri, jwtIssuer)
			creationMatcher := getActionMatcher("create", "new-namespace", "old-service", jwksUri, jwtIssuer)
			Expect(result).To(ContainElements(deletionMatcher, creationMatcher))
		})
	})

})
