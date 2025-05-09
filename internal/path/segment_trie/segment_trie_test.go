package segment_trie_test

import (
	"github.com/kyma-project/api-gateway/internal/path/token"
	"github.com/thoas/go-funk"
	"strings"
	"testing"

	. "github.com/kyma-project/api-gateway/internal/path/segment_trie"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSegmentTrie(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SegmentTrie Suite")
}

var _ = Describe("SegmentTrie", func() {
	It("should not panic on a multiple operator path", func() {
		trie := New()
		err := trie.InsertAndCheckCollisions(token.TokenizePath("/abc/{**}/def/{**}"))
		Expect(err).To(BeNil())
	})

	It("should return nil for empty path", func() {
		trie := New()
		err := trie.InsertAndCheckCollisions([]token.Token{})
		Expect(err).To(BeNil())
	})

	DescribeTable("Conflict Checking",
		func(paths []string, conflictNumber int) {
			trie := New()
			errNumber := 0
			for _, path := range paths {
				tokenizedPath := token.TokenizePath(path)
				err := trie.InsertAndCheckCollisions(tokenizedPath)
				if err != nil {
					errNumber++
				}
			}
			Expect(errNumber).To(Equal(conflictNumber))
		},
		Entry("No conflict: only one double asterisk", []string{
			"/{**}",
		}, 0),
		Entry("No conflict: prefix and double asterisk", []string{
			"/abc/{**}",
		}, 0),
		Entry("No conflict: path without / at end and same path with double asterisk", []string{
			"/abc",
			"/abc/{**}",
		}, 0),
		Entry("No conflict: path without / at end and same path with single asterisk", []string{
			"/abc",
			"/abc/{*}",
		}, 0),
		Entry("No conflict: path with / at end and same path with single asterisk", []string{
			"/abc/",
			"/abc/{*}",
		}, 0),
		Entry("No conflict: path with / at end and same path with double asterisk", []string{
			"/abc/",
			"/abc/{**}",
		}, 0),
		Entry("No conflict: exact with single asterisk", []string{
			"/abc/def/ghi",
			"/abc/{*}/ghi",
		}, 0),
		Entry("Conflict: exact with exact", []string{
			"/abc/def",
			"/abc/def/ghi",
			"/abc/def/ghi",
			"/abc/def",
		}, 2),
		Entry("Conflict: exact with double asterisk", []string{
			"/abc/def",
			"/abc/def/ghi",
			"/abc/{**}/ghi",
			"/abc/{**}",
			"/{**}/ghi",
			"/{**}",
			"/{**}/def/ghi",
		}, 1),
		Entry("Conflict: exact with double asterisk", []string{
			"/{**}",
			"/{**}/def/ghi",
		}, 1),
		Entry("No conflict: double asterisk with single asterisk", []string{
			"/abc/{**}/ghi",
			"/{*}/{*}/ghi",
		}, 0),
		Entry("Conflict: double asterisk with single asterisk", []string{
			"/abc/{**}/ghi",
			"/abc/{*}/ghi",
		}, 1),
		Entry("No conflict: double asterisk with double asterisk", []string{
			"/abc/{**}/def/ghi",
			"/abc/{**}/ghi",
			"/abc/{**}",
		}, 0),
		Entry("Multiple conflicts: double asterisk covering another double asterisk", []string{
			"/abc/{**}",
			"/abc/{**}/ghi",
			"/abc/{**}/def/ghi",
		}, 2),
		Entry("No conflict: exact paths", []string{
			"/abc/def",
			"/abc/def/ghi",
			"/abc/def/foo",
			"/def/ghi",
		}, 0),
		Entry("No conflict: paths with single asterisk, but different suffix", []string{
			"/abc/{*}/def",
			"/abc/{*}/ghi",
			"/abc/{*}/foo/bar",
			"/abc/{*}",
		}, 0),
		Entry("No conflict: paths with double asterisk, but different suffix", []string{
			"/abc/{**}/def",
			"/abc/{**}/ghi",
			"/abc/{**}/foo/bar",
		}, 0),
		Entry("Conflict: exact path with single asterisk path", []string{
			"/abc/{*}",
			"/abc/def",
		}, 1),
		Entry("Conflict: exact path with double asterisk path", []string{
			"/abc/{**}",
			"/abc/def",
		}, 1),
		Entry("No conflict: single asterisk path with exact path", []string{
			"/abc/def",
			"/abc/{*}",
		}, 0),
		Entry("No conflict: single asterisk path with exact path", []string{
			"/abc/def",
			"/abc/{*}",
		}, 0),
		Entry("No conflict: double asterisk path different prefixes", []string{
			"/abc/{**}",
			"/bca/{**}/abc",
		}, 0),
		Entry("Conflict: double asterisk with a different containing the first one", []string{
			"/abc/{**}/def",
			"/abc/def/{**}/def",
		}, 1),
		Entry("No conflict: single double asterisk with single path", []string{
			"/a/{**}/abc",
			"/a/b/{**}/abc/abcd",
		}, 0))
})

func BenchmarkSegmentTrieInsertion(b *testing.B) {
	b.StopTimer()
	pathNumber := b.N

	paths := make([]string, pathNumber)

	for i := 0; i < pathNumber; i++ {
		numSegments := funk.RandomInt(1, 10)
		pathBuilder := strings.Builder{}
		for j := 0; j < numSegments; j++ {
			pathBuilder.WriteString("/")
			if r := funk.RandomInt(0, 6); r == 0 {
				pathBuilder.WriteString("{**}")
			} else if r == 1 {
				pathBuilder.WriteString("{*}")
			} else {
				pathBuilder.WriteString(funk.RandomString(1))
			}
		}
		path := pathBuilder.String()
		paths[i] = path
	}

	errNr := 0
	trie := New()
	b.StartTimer()
	for i := 0; i < pathNumber; i++ {
		err := trie.InsertAndCheckCollisions(token.TokenizePath(paths[i]))
		if err != nil {
			errNr++
		}
	}
	b.StopTimer()
	b.Logf("Number of conflicts: %d, Number of paths: %d, Benchmark time: %s", errNr, pathNumber, b.Elapsed().String())
}
