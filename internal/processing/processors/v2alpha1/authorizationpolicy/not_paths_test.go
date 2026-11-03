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

	It("carves a secured mid-segment `/a/b/{*}` out of a later noAuth `/a/{*}/c` they overlap on", func() {
		// given: secured /a/b/{*} GET, then noAuth /a/{*}/c GET.
		// Neither path is a superset of the other, but both match the concrete
		// request /a/b/c. A superset-only carve would miss this and expose
		// /a/b/c unauthenticated; the intersection test must carve it.
		securedMid := newJwtRuleBuilderWithDummyData().
			withPath("/a/b/{*}").
			withMethods(gatewayv2alpha1.HttpMethod("GET")).
			build()
		noAuthMid := newNoAuthRuleBuilderWithDummyData().
			withPath("/a/{*}/c").
			withMethods(gatewayv2alpha1.HttpMethod("GET")).
			build()
		apiRule := newAPIRuleBuilderWithDummyData().
			withRules(securedMid, noAuthMid).
			build()

		// when
		aps := reconcile(apiRule)

		// then
		noAuthAP := apForPath(aps, "/a/{*}/c")
		Expect(noAuthAP).NotTo(BeNil())
		Expect(noAuthAP.Spec.Rules[0].To[0].Operation.NotPaths).To(ConsistOf("/a/b/{*}"))
	})

	It("carves a secured mid-segment `/a/{*}/c` out of a later noAuth `/a/b/{**}` they overlap on", func() {
		// given: secured /a/{*}/c GET, then noAuth /a/b/{**} GET.
		// They overlap on /a/b/c; neither covers the other. Must be carved.
		securedMid := newJwtRuleBuilderWithDummyData().
			withPath("/a/{*}/c").
			withMethods(gatewayv2alpha1.HttpMethod("GET")).
			build()
		noAuthPrefix := newNoAuthRuleBuilderWithDummyData().
			withPath("/a/b/{**}").
			withMethods(gatewayv2alpha1.HttpMethod("GET")).
			build()
		apiRule := newAPIRuleBuilderWithDummyData().
			withRules(securedMid, noAuthPrefix).
			build()

		// when
		aps := reconcile(apiRule)

		// then
		noAuthAP := apForPath(aps, "/a/b/{**}")
		Expect(noAuthAP).NotTo(BeNil())
		Expect(noAuthAP.Spec.Rules[0].To[0].Operation.NotPaths).To(ConsistOf("/a/{*}/c"))
	})

	It("does not carve mid-segment paths that only differ in a literal and never overlap", func() {
		// given: secured /a/{*}/c GET, then noAuth /a/{*}/d GET.
		// The trailing literal differs (c vs d), so no concrete path matches both.
		securedC := newJwtRuleBuilderWithDummyData().
			withPath("/a/{*}/c").
			withMethods(gatewayv2alpha1.HttpMethod("GET")).
			build()
		noAuthD := newNoAuthRuleBuilderWithDummyData().
			withPath("/a/{*}/d").
			withMethods(gatewayv2alpha1.HttpMethod("GET")).
			build()
		apiRule := newAPIRuleBuilderWithDummyData().
			withRules(securedC, noAuthD).
			build()

		// when
		aps := reconcile(apiRule)

		// then
		noAuthAP := apForPath(aps, "/a/{*}/d")
		Expect(noAuthAP).NotTo(BeNil())
		Expect(noAuthAP.Spec.Rules[0].To[0].Operation.NotPaths).To(BeEmpty())
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
