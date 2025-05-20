package requestauthentication_test

import (
	"context"
	"fmt"
	"net/http"

	"github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	"github.com/kyma-project/api-gateway/internal/builders/builders_test/v2alpha1_test"
	"github.com/kyma-project/api-gateway/internal/processing"
	"github.com/kyma-project/api-gateway/internal/processing/processors/v2alpha1/requestauthentication"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
)

var _ = Describe("Processing", func() {

	appSelector := "app"

	It("should produce one RA for a rule with one issuer and two paths", func() {
		// given
		ruleJwt := newJwtRuleBuilderWithDummyData().
			build()
		ruleJwt2 := newJwtRuleBuilderWithDummyData().
			withPath("/img").
			build()

		apiRule := newAPIRuleBuilderWithDummyData().
			withRules(ruleJwt, ruleJwt2).
			build()
		svc := newServiceBuilderWithDummyData().build()
		client := getFakeClient(svc)
		processor := requestauthentication.NewProcessor(apiRule)

		// when
		result, err := processor.EvaluateReconciliation(context.Background(), client)

		// then
		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(1))
		ra := result[0].Obj.(*securityv1beta1.RequestAuthentication)
		Expect(ra).NotTo(BeNil())
		Expect(ra.ObjectMeta.Name).To(BeEmpty())
		Expect(ra.ObjectMeta.GenerateName).To(Equal(apiRuleName + "-"))
		Expect(ra.ObjectMeta.Namespace).To(Equal(apiRuleNamespace))

		Expect(ra.Spec.Selector.MatchLabels[appSelector]).NotTo(BeNil())
		Expect(ra.Spec.Selector.MatchLabels[appSelector]).To(Equal(serviceName))
		Expect(len(ra.Spec.JwtRules)).To(Equal(1))
		Expect(ra.Spec.JwtRules[0].Issuer).To(Equal(jwtIssuer))
		Expect(ra.Spec.JwtRules[0].JwksUri).To(Equal(jwksUri))
	})

	It("should produce RA for a Rule without service, but service definition on ApiRule level", func() {
		// given
		ruleJwt := newJwtRuleBuilderWithDummyData().
			build()
		apiRule := newAPIRuleBuilderWithDummyData().
			withRules(ruleJwt).
			build()
		svc := newServiceBuilderWithDummyData().build()
		client := getFakeClient(svc)
		processor := requestauthentication.NewProcessor(apiRule)

		// when
		result, err := processor.EvaluateReconciliation(context.Background(), client)

		// then
		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(1))

		ra := result[0].Obj.(*securityv1beta1.RequestAuthentication)
		Expect(ra).NotTo(BeNil())
		// The RA should be in .Spec.Service.Namespace
		Expect(ra.Namespace).To(Equal(apiRuleNamespace))
		Expect(ra.Spec.Selector.MatchLabels[appSelector]).To(Equal(serviceName))
	})

	It("should produce RA with service from Rule, when service is configured on Rule and ApiRule level", func() {
		// given
		ruleServiceName := "rule-scope-example-service"
		specServiceNamespace := "spec-service-namespace"

		ruleJwt := newRuleBuilder().
			withPath("/").
			addMethods(http.MethodGet).
			withServiceName(ruleServiceName).
			withServicePort(8080).
			addJwtAuthentication(jwtIssuer, jwksUri).
			build()
		apiRule := newAPIRuleBuilderWithDummyData().
			withServiceNamespace(specServiceNamespace).
			withRules(ruleJwt).
			build()

		svc := newServiceBuilder().
			withName(ruleServiceName).
			withNamespace(specServiceNamespace).
			addSelector(appSelector, ruleServiceName).
			build()

		client := getFakeClient(svc)
		processor := requestauthentication.NewProcessor(apiRule)

		// when
		result, err := processor.EvaluateReconciliation(context.Background(), client)

		// then
		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(1))

		ra := result[0].Obj.(*securityv1beta1.RequestAuthentication)
		Expect(ra).NotTo(BeNil())
		// The RA should be in .Spec.Service.Namespace
		Expect(ra.Namespace).To(Equal(specServiceNamespace))
		Expect(ra.Spec.Selector.MatchLabels[appSelector]).To(Equal(ruleServiceName))
	})

	It("should produce RA for a Rule with service with configured namespace, in the configured namespace", func() {
		// given
		ruleServiceName := "rule-scope-example-service"
		ruleServiceNamespace := "rule-service-namespace"

		jwtRule := newJwtRuleBuilderWithDummyData().
			withServiceName(ruleServiceName).
			withServiceNamespace(ruleServiceNamespace).
			build()

		apiRule := newAPIRuleBuilderWithDummyData().
			withRules(jwtRule).
			build()

		svc := newServiceBuilder().
			withName(ruleServiceName).
			withNamespace(ruleServiceNamespace).
			addSelector(appSelector, ruleServiceName).
			build()

		client := getFakeClient(svc)
		processor := requestauthentication.NewProcessor(apiRule)

		// when
		result, err := processor.EvaluateReconciliation(context.Background(), client)

		// then
		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(1))

		ra := result[0].Obj.(*securityv1beta1.RequestAuthentication)
		Expect(ra).NotTo(BeNil())
		Expect(ra.Spec.Selector.MatchLabels[appSelector]).To(Equal(ruleServiceName))
		// The RA should be in .Service.Namespace
		Expect(ra.Namespace).To(Equal(ruleServiceNamespace))
		// And the OwnerLabel should point to APIRule namespace
		Expect(ra.Labels[processing.OwnerLabel]).ToNot(BeEmpty())
		Expect(ra.Labels[processing.OwnerLabel]).To(Equal(fmt.Sprintf("%s.%s", apiRule.Name, apiRule.Namespace)))
	})

	It("should produce RA from a rule with two issuers and one path", func() {
		jwtRule := newJwtRuleBuilderWithDummyData().
			addJwtAuthentication(anotherJwtIssuer, anotherJwksUri).
			build()
		apiRule := newAPIRuleBuilderWithDummyData().
			withRules(jwtRule).
			build()
		svc := newServiceBuilderWithDummyData().build()
		client := getFakeClient(svc)
		processor := requestauthentication.NewProcessor(apiRule)

		// when
		result, err := processor.EvaluateReconciliation(context.Background(), client)

		// then
		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(1))
		ra := result[0].Obj.(*securityv1beta1.RequestAuthentication)

		Expect(ra).NotTo(BeNil())
		Expect(ra.ObjectMeta.Name).To(BeEmpty())
		Expect(ra.ObjectMeta.GenerateName).To(Equal(apiRuleName + "-"))
		Expect(ra.ObjectMeta.Namespace).To(Equal(apiRuleNamespace))

		Expect(ra.Spec.Selector.MatchLabels[appSelector]).NotTo(BeNil())
		Expect(ra.Spec.Selector.MatchLabels[appSelector]).To(Equal(serviceName))
		Expect(len(ra.Spec.JwtRules)).To(Equal(2))
		Expect(ra.Spec.JwtRules[0].Issuer).To(Equal(jwtIssuer))
		Expect(ra.Spec.JwtRules[0].JwksUri).To(Equal(jwksUri))
		Expect(ra.Spec.JwtRules[1].Issuer).To(Equal(anotherJwtIssuer))
		Expect(ra.Spec.JwtRules[1].JwksUri).To(Equal(anotherJwksUri))
	})

	It("should not create RA when NoAuth is true", func() {
		// given
		rule := newNoAuthRuleBuilderWithDummyData().build()
		apiRule := newAPIRuleBuilderWithDummyData().
			withRules(rule).
			build()

		client := getFakeClient()
		processor := requestauthentication.NewProcessor(apiRule)

		// when
		result, err := processor.EvaluateReconciliation(context.Background(), client)

		// then
		Expect(err).To(BeNil())
		Expect(result).To(BeEmpty())
	})

	It("should create RA when no exists", func() {
		// given: New resources
		jwtRule := newJwtRuleBuilderWithDummyData().build()
		apiRule := newAPIRuleBuilderWithDummyData().
			withRules(jwtRule).
			build()
		svc := newServiceBuilderWithDummyData().build()
		client := getFakeClient(svc)
		processor := requestauthentication.NewProcessor(apiRule)

		// when
		result, err := processor.EvaluateReconciliation(context.Background(), client)

		// then
		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(1))
		Expect(result[0].Action.String()).To(Equal("create"))
	})

	Context("extAuth", func() {
		It("should not create RA when ExtAuth has no restrictions configured", func() {
			// given
			rule := v2alpha1_test.NewRuleBuilder().WithExtAuth(v2alpha1_test.NewExtAuthBuilder().WithAuthorizers("abc").Build()).Build()
			apiRule := newAPIRuleBuilderWithDummyData().
				withRules(rule).
				build()
			client := getFakeClient()
			processor := requestauthentication.NewProcessor(apiRule)

			// when
			result, err := processor.EvaluateReconciliation(context.Background(), client)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(BeEmpty())
		})

		It("should create RA when ExtAuth has restrictions with authenticators configured", func() {
			// given
			rule := v2alpha1_test.
				NewRuleBuilder().
				WithExtAuth(
					v2alpha1_test.NewExtAuthBuilder().
						WithRestriction(&v2alpha1.JwtConfig{
							Authentications: []*v2alpha1.JwtAuthentication{
								{
									Issuer:  jwtIssuer,
									JwksUri: jwksUri,
								},
							},
						}).
						Build()).
				Build()

			apiRule := newAPIRuleBuilderWithDummyData().
				withRules(rule).
				build()

			svc := newServiceBuilderWithDummyData().build()
			client := getFakeClient(svc)
			processor := requestauthentication.NewProcessor(apiRule)

			// when
			result, err := processor.EvaluateReconciliation(context.Background(), client)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(1))
			Expect(result[0].Action.String()).To(Equal("create"))
		})
	})
})
