package authorizationpolicy

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// pathCovers reports template-superset: does `container`'s match set contain
// all paths matched by `contained`? Both arguments are Istio path templates.
var _ = DescribeTable("pathCovers",
	func(container, contained string, want bool) {
		Expect(pathCovers(container, contained)).To(Equal(want))
	},

	// Reflexive: a template always covers itself.
	Entry("/foo covers /foo", "/foo", "/foo", true),
	Entry("/{**} covers /{**}", "/{**}", "/{**}", true),

	// /{**} — bare catch-all covers every template.
	Entry("/{**} covers /", "/{**}", "/", true),
	Entry("/{**} covers /foo", "/{**}", "/foo", true),
	Entry("/{**} covers /a/b/c", "/{**}", "/a/b/c", true),
	Entry("/{**} covers /x/{**}", "/{**}", "/x/{**}", true),
	Entry("/{**} covers /{*}", "/{**}", "/{*}", true),
	Entry("/{**} covers /a/{*}/c", "/{**}", "/a/{*}/c", true),

	// /x/{**} — prefixed double-star.
	Entry("/x/{**} covers /x/foo", "/x/{**}", "/x/foo", true),
	Entry("/x/{**} covers /x/foo/bar", "/x/{**}", "/x/foo/bar", true),
	Entry("/x/{**} covers /x/{*}", "/x/{**}", "/x/{*}", true),
	Entry("/x/{**} covers /x/{**}", "/x/{**}", "/x/{**}", true),
	Entry("/x/{**} does not cover /y (different prefix)", "/x/{**}", "/y", false),
	Entry("/x/{**} does not cover / (root)", "/x/{**}", "/", false),
	Entry("/x/{**} does not cover /{**} (broader)", "/x/{**}", "/{**}", false),

	// /{*} — single non-empty segment.
	Entry("/{*} covers /foo", "/{*}", "/foo", true),
	Entry("/{*} covers /{*}", "/{*}", "/{*}", true),
	Entry("/{*} does not cover /a/b (two segments)", "/{*}", "/a/b", false),
	Entry("/{*} does not cover / (root, empty)", "/{*}", "/", false),
	Entry("/{*} does not cover /{**} (broader)", "/{*}", "/{**}", false),

	// /x/{*} — prefixed single-segment.
	Entry("/x/{*} covers /x/foo", "/x/{*}", "/x/foo", true),
	Entry("/x/{*} covers /x/{*}", "/x/{*}", "/x/{*}", true),
	Entry("/x/{*} does not cover /x/foo/bar", "/x/{*}", "/x/foo/bar", false),
	Entry("/x/{*} does not cover /x (no segment)", "/x/{*}", "/x", false),
	Entry("/x/{*} does not cover /y/foo (wrong prefix)", "/x/{*}", "/y/foo", false),

	// Mid-segment {*}.
	Entry("/a/{*}/c covers /a/b/c", "/a/{*}/c", "/a/b/c", true),
	Entry("/a/{*}/c covers /a/{*}/c", "/a/{*}/c", "/a/{*}/c", true),
	Entry("/a/{*}/c does not cover /a/b/c/d", "/a/{*}/c", "/a/b/c/d", false),
	Entry("/a/{*}/c does not cover /a/b/{*} (differs in last)", "/a/{*}/c", "/a/b/{*}", false),

	// Exact/literal templates cover nothing broader.
	Entry("exact /foo does not cover /bar", "/foo", "/bar", false),
	Entry("exact /foo/bar does not cover /foo/bar/baz", "/foo/bar", "/foo/bar/baz", false),
	Entry("exact /a/b does not cover /a/{*}", "/a/b", "/a/{*}", false),
)

// pathsOverlap reports whether two Istio path templates can both match at least
// one concrete request path (match-set intersection). It is symmetric.
var _ = DescribeTable("pathsOverlap",
	func(a, b string, want bool) {
		Expect(pathsOverlap(a, b)).To(Equal(want))
		Expect(pathsOverlap(b, a)).To(Equal(want)) // symmetry
	},

	// Bare catch-all overlaps everything.
	Entry("/{**} vs /", "/{**}", "/", true),
	Entry("/{**} vs /a/b/c", "/{**}", "/a/b/c", true),
	Entry("/{**} vs /a/{*}/c", "/{**}", "/a/{*}/c", true),

	// Disjoint by segment count / literal.
	Entry("/{*} vs / are disjoint", "/{*}", "/", false),
	Entry("/a vs /a/b are disjoint", "/a", "/a/b", false),
	Entry("/a/{*} vs /a are disjoint", "/a/{*}", "/a", false),

	// pathsOverlap is deliberately conservative about the trailing-slash
	// distinction that `{**}` introduces (`/a/{**}` matches `/a/`, not bare
	// `/a`): it treats `{**}` as able to match zero trailing segments, so it
	// reports overlap here even though a strict Envoy match set would not.
	// This only ever causes extra (harmless) notPaths carve-outs — carving is
	// fail-closed — never a missed one.
	Entry("/a/{**} vs /{*} conservatively overlap", "/a/{**}", "/{*}", true),

	// The regression case: neither covers the other, but they overlap on /a/b/c.
	Entry("/a/b/{*} vs /a/{*}/c overlap on /a/b/c", "/a/b/{*}", "/a/{*}/c", true),
	Entry("/a/{*}/c vs /a/b/{**} overlap on /a/b/c", "/a/{*}/c", "/a/b/{**}", true),

	// Single-segment templates.
	Entry("/{*} vs /foo overlap", "/{*}", "/foo", true),
	Entry("/{*} vs /{*} overlap", "/{*}", "/{*}", true),
	Entry("/{*} vs /a/b disjoint (segment count)", "/{*}", "/a/b", false),

	// Prefixed stars.
	Entry("/a/{**} vs /a/b/{*} overlap", "/a/{**}", "/a/b/{*}", true),
	Entry("/a/{**} vs /b/{**} disjoint (prefix)", "/a/{**}", "/b/{**}", false),
	Entry("/a/{*} vs /a/{**} overlap on /a/x", "/a/{*}", "/a/{**}", true),

	// Literal mismatch.
	Entry("/a/b vs /a/c disjoint", "/a/b", "/a/c", false),
	Entry("/a/{*}/c vs /a/{*}/d disjoint (last literal)", "/a/{*}/c", "/a/{*}/d", false),
)
