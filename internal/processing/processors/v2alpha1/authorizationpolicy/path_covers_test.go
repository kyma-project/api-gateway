package authorizationpolicy

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = DescribeTable("pathCovers",
	func(container, contained string, want bool) {
		Expect(pathCovers(container, contained)).To(Equal(want))
	},

	// /{**} — bare catch-all covers every path
	Entry("/{**} covers /", "/{**}", "/", true),
	Entry("/{**} covers single-segment /foo", "/{**}", "/foo", true),
	Entry("/{**} covers multi-segment /a/b/c", "/{**}", "/a/b/c", true),

	// /x/{**} — prefixed double-star catch-all
	Entry("/x/{**} covers exact prefix /x", "/x/{**}", "/x", true),
	Entry("/x/{**} covers /x/foo", "/x/{**}", "/x/foo", true),
	Entry("/x/{**} covers deep /x/foo/bar", "/x/{**}", "/x/foo/bar", true),
	Entry("/x/{**} does not cover /xy (substring trap)", "/x/{**}", "/xy", false),
	Entry("/x/{**} does not cover /y (different prefix)", "/x/{**}", "/y", false),
	Entry("/x/{**} does not cover /", "/x/{**}", "/", false),

	// /{*} — bare single-segment catch-all
	Entry("/{*} covers /foo", "/{*}", "/foo", true),
	Entry("/{*} covers /secret", "/{*}", "/secret", true),
	Entry("/{*} does not cover /a/b (two segments)", "/{*}", "/a/b", false),
	Entry("/{*} does not cover / (empty segment)", "/{*}", "/", false),

	// /x/{*} — prefixed single-segment catch-all
	Entry("/x/{*} covers /x/foo", "/x/{*}", "/x/foo", true),
	Entry("/x/{*} does not cover /x/foo/bar (two segments under /x/)", "/x/{*}", "/x/foo/bar", false),
	Entry("/x/{*} does not cover /x (no segment after prefix)", "/x/{*}", "/x", false),
	Entry("/x/{*} does not cover /y/foo (wrong prefix)", "/x/{*}", "/y/foo", false),
	Entry("/x/{*} does not cover /x/ (empty segment)", "/x/{*}", "/x/", false),

	// Exact paths — never cover anything
	Entry("exact /foo does not cover /foo", "/foo", "/foo", false),
	Entry("exact /foo does not cover /bar", "/foo", "/bar", false),
	Entry("exact /foo/bar does not cover /foo/bar/baz", "/foo/bar", "/foo/bar/baz", false),
	Entry("empty container does not cover /foo", "", "/foo", false),
)
