package authorizationpolicy_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"

	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	"github.com/kyma-project/api-gateway/internal/processing/processors/v2alpha1/authorizationpolicy"
)

// apForPath returns the AuthorizationPolicy whose operation targets the given path.
func apForPath(results []*securityv1beta1.AuthorizationPolicy, path string) *securityv1beta1.AuthorizationPolicy {
	for _, ap := range results {
		if len(ap.Spec.Rules) == 0 {
			continue
		}
		if ap.Spec.Rules[0].To[0].Operation.Paths[0] == path {
			return ap
		}
	}
	return nil
}

var _ = Describe("Generating notPaths for rule ordering", func() {
	reconcile := func(apiRule *gatewayv2alpha1.APIRule) []*securityv1beta1.AuthorizationPolicy {
		svc := newServiceBuilderWithDummyData().build()
		gateway := newGatewayBuilderWithDummyData().build()
		client := getFakeClient(svc)
		processor := authorizationpolicy.NewProcessor(&testLogger, apiRule, gateway, client)

		results, err := processor.EvaluateReconciliation(context.Background(), client)
		Expect(err).To(BeNil())

		aps := make([]*securityv1beta1.AuthorizationPolicy, 0, len(results))
		for _, result := range results {
			aps = append(aps, result.Obj.(*securityv1beta1.AuthorizationPolicy))
		}
		return aps
	}

	It("carves a secured exact `/` out of a later noAuth `/{**}` catch-all", func() {
		// given
		securedRoot := newJwtRuleBuilderWithDummyData().
			withPath("/").
			withMethods(gatewayv2alpha1.HttpMethod("GET")).
			build()
		catchAll := newNoAuthRuleBuilderWithDummyData().
			withPath("/{**}").
			withMethods(gatewayv2alpha1.HttpMethod("GET")).
			build()
		apiRule := newAPIRuleBuilderWithDummyData().
			withRules(securedRoot, catchAll).
			build()

		// when
		aps := reconcile(apiRule)

		// then
		catchAllAP := apForPath(aps, "/{**}")
		Expect(catchAllAP).NotTo(BeNil())
		Expect(catchAllAP.Spec.Rules[0].To[0].Operation.NotPaths).To(ConsistOf("/"))
	})

	It("carves a secured prefixed `/x/{**}` out of a later noAuth `/{**}` catch-all", func() {
		// given
		securedPrefix := newJwtRuleBuilderWithDummyData().
			withPath("/x/{**}").
			withMethods(gatewayv2alpha1.HttpMethod("GET")).
			build()
		catchAll := newNoAuthRuleBuilderWithDummyData().
			withPath("/{**}").
			withMethods(gatewayv2alpha1.HttpMethod("GET")).
			build()
		apiRule := newAPIRuleBuilderWithDummyData().
			withRules(securedPrefix, catchAll).
			build()

		// when
		aps := reconcile(apiRule)

		// then
		catchAllAP := apForPath(aps, "/{**}")
		Expect(catchAllAP).NotTo(BeNil())
		Expect(catchAllAP.Spec.Rules[0].To[0].Operation.NotPaths).To(ConsistOf("/x/{**}"))
	})

	It("does not emit a bare catch-all as a notPaths source", func() {
		// given: two catch-alls on the same method; the later one cannot carve out
		// the earlier bare catch-all (that case is rejected at admission instead).
		securedCatchAll := newJwtRuleBuilderWithDummyData().
			withPath("/{**}").
			withMethods(gatewayv2alpha1.HttpMethod("POST")).
			build()
		laterRule := newJwtRuleBuilderWithDummyData().
			withPath("/x").
			withMethods(gatewayv2alpha1.HttpMethod("POST")).
			build()
		apiRule := newAPIRuleBuilderWithDummyData().
			withRules(securedCatchAll, laterRule).
			build()

		// when
		aps := reconcile(apiRule)

		// then: the /x rule comes after the catch-all, so nothing is carved out of it
		laterAP := apForPath(aps, "/x")
		Expect(laterAP).NotTo(BeNil())
		Expect(laterAP.Spec.Rules[0].To[0].Operation.NotPaths).To(BeEmpty())
	})

	It("carves a secured exact path out of a later noAuth `/{*}` single-segment catch-all", func() {
		// given: secured /secret GET, then noAuth /{*} GET
		// /{*} matches any single-segment path including /secret; it must be carved out.
		securedExact := newJwtRuleBuilderWithDummyData().
			withPath("/secret").
			withMethods(gatewayv2alpha1.HttpMethod("GET")).
			build()
		singleSegment := newNoAuthRuleBuilderWithDummyData().
			withPath("/{*}").
			withMethods(gatewayv2alpha1.HttpMethod("GET")).
			build()
		apiRule := newAPIRuleBuilderWithDummyData().
			withRules(securedExact, singleSegment).
			build()

		// when
		aps := reconcile(apiRule)

		// then
		singleSegAP := apForPath(aps, "/{*}")
		Expect(singleSegAP).NotTo(BeNil())
		Expect(singleSegAP.Spec.Rules[0].To[0].Operation.NotPaths).To(ConsistOf("/secret"))
	})

	It("carves a secured `/a/{*}` out of a later noAuth `/{**}` catch-all", func() {
		// given: secured /a/{*} GET, then noAuth /{**} GET
		// /{**} covers /a/x and must exclude /a/{*}.
		securedPrefixed := newJwtRuleBuilderWithDummyData().
			withPath("/a/{*}").
			withMethods(gatewayv2alpha1.HttpMethod("GET")).
			build()
		catchAll := newNoAuthRuleBuilderWithDummyData().
			withPath("/{**}").
			withMethods(gatewayv2alpha1.HttpMethod("GET")).
			build()
		apiRule := newAPIRuleBuilderWithDummyData().
			withRules(securedPrefixed, catchAll).
			build()

		// when
		aps := reconcile(apiRule)

		// then
		catchAllAP := apForPath(aps, "/{**}")
		Expect(catchAllAP).NotTo(BeNil())
		Expect(catchAllAP.Spec.Rules[0].To[0].Operation.NotPaths).To(ConsistOf("/a/{*}"))
	})

	It("does not carve an earlier path when methods do not overlap", func() {
		// given: secured /x/{**} GET, then noAuth /{**} POST — different methods
		securedPrefix := newJwtRuleBuilderWithDummyData().
			withPath("/x/{**}").
			withMethods(gatewayv2alpha1.HttpMethod("GET")).
			build()
		catchAll := newNoAuthRuleBuilderWithDummyData().
			withPath("/{**}").
			withMethods(gatewayv2alpha1.HttpMethod("POST")).
			build()
		apiRule := newAPIRuleBuilderWithDummyData().
			withRules(securedPrefix, catchAll).
			build()

		// when
		aps := reconcile(apiRule)

		// then: POST catch-all AP must not carve out the GET /x/{**} path
		catchAllAP := apForPath(aps, "/{**}")
		Expect(catchAllAP).NotTo(BeNil())
		Expect(catchAllAP.Spec.Rules[0].To[0].Operation.NotPaths).To(BeEmpty())
	})

	It("does not carve an earlier path when the catch-all does not cover it", func() {
		// given: secured /a/b GET, then noAuth /x/{**} GET
		// pathCovers("/x/{**}", "/a/b") = false, so /a/b must not appear in notPaths.
		securedExact := newJwtRuleBuilderWithDummyData().
			withPath("/a/b").
			withMethods(gatewayv2alpha1.HttpMethod("GET")).
			build()
		prefixCatchAll := newNoAuthRuleBuilderWithDummyData().
			withPath("/x/{**}").
			withMethods(gatewayv2alpha1.HttpMethod("GET")).
			build()
		apiRule := newAPIRuleBuilderWithDummyData().
			withRules(securedExact, prefixCatchAll).
			build()

		// when
		aps := reconcile(apiRule)

		// then
		prefixAP := apForPath(aps, "/x/{**}")
		Expect(prefixAP).NotTo(BeNil())
		Expect(prefixAP.Spec.Rules[0].To[0].Operation.NotPaths).To(BeEmpty())
	})

	It("normalizes /* to /{**} and carves an earlier secured path", func() {
		// given: secured /foo GET, then noAuth /* GET
		// /* is normalized to /{**}; /foo must be carved out.
		securedExact := newJwtRuleBuilderWithDummyData().
			withPath("/foo").
			withMethods(gatewayv2alpha1.HttpMethod("GET")).
			build()
		catchAll := newNoAuthRuleBuilderWithDummyData().
			withPath("/*").
			withMethods(gatewayv2alpha1.HttpMethod("GET")).
			build()
		apiRule := newAPIRuleBuilderWithDummyData().
			withRules(securedExact, catchAll).
			build()

		// when
		aps := reconcile(apiRule)

		// then: the AP is stored with the normalized path /{**}
		catchAllAP := apForPath(aps, "/{**}")
		Expect(catchAllAP).NotTo(BeNil())
		Expect(catchAllAP.Spec.Rules[0].To[0].Operation.NotPaths).To(ConsistOf("/foo"))
	})

	It("carves multiple earlier secured paths out of a single catch-all", func() {
		// given: secured /a GET, secured /b/{**} GET, then noAuth /{**} GET
		securedA := newJwtRuleBuilderWithDummyData().
			withPath("/a").
			withMethods(gatewayv2alpha1.HttpMethod("GET")).
			build()
		securedB := newJwtRuleBuilderWithDummyData().
			withPath("/b/{**}").
			withMethods(gatewayv2alpha1.HttpMethod("GET")).
			build()
		catchAll := newNoAuthRuleBuilderWithDummyData().
			withPath("/{**}").
			withMethods(gatewayv2alpha1.HttpMethod("GET")).
			build()
		apiRule := newAPIRuleBuilderWithDummyData().
			withRules(securedA, securedB, catchAll).
			build()

		// when
		aps := reconcile(apiRule)

		// then
		catchAllAP := apForPath(aps, "/{**}")
		Expect(catchAllAP).NotTo(BeNil())
		Expect(catchAllAP.Spec.Rules[0].To[0].Operation.NotPaths).To(ConsistOf("/a", "/b/{**}"))
	})
})
