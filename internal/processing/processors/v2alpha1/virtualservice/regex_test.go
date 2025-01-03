package virtualservice

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"regexp"
)

var _ = Describe("Envoy templates regex matching", func() {
	DescribeTable(`{\*} template`,
		func(input string, shouldMatch bool) {
			// when
			matched, err := regexp.MatchString(prepareRegexPath("/{*}"), input)

			// then
			Expect(err).To(Not(HaveOccurred()))
			Expect(matched).To(Equal(shouldMatch))
		},
		Entry("should not match empty path", "/", false),
		Entry("should match path with one segment", "/segment", true),
		Entry("should match special characters", "/segment-._~!$&'()*+,;=:@", true),
		Entry("should match with correct % encoding", "/segment%20", true),
		Entry("should not match with incorrect % encoding", "/segment%2", false),
		Entry("should not match path with multiple segments", "/segment1/segment2/segment3", false),
	)

	DescribeTable(`{\*\*} template`, func(input string, shouldMatch bool) {
		// when
		matched, err := regexp.MatchString(prepareRegexPath("/{**}"), input)

		// then
		Expect(err).To(Not(HaveOccurred()))
		Expect(matched).To(Equal(shouldMatch))
	},
		Entry("should match empty path", "/", true),
		Entry("should match path with one segment", "/segment", true),
		Entry("should match special characters", "/segment-._~!$&'()*+,;=:@", true),
		Entry("should match with correct % encoding", "/segment%20", true),
		Entry("should not match with incorrect % encoding", "/segment%2", false),
		Entry("should match path with multiple segments", "/segment1/segment2/segment3", true),
	)
})
