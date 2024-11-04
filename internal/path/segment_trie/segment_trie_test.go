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
		Entry("Conflict: exact with single asterisk", []string{
			"/abc/def/ghi",
			"/abc/{*}/ghi",
		}, 1),
		Entry("Conflict: exact with exact", []string{
			"/abc/def/ghi",
			"/abc/def/ghi",
		}, 1),
		Entry("Conflict: exact with double asterisk", []string{
			"/abc/def/ghi",
			"/abc/{**}/ghi",
			"/abc/{**}",
			"/{**}/ghi",
			"/{**}",
			"/{**}/def/ghi",
		}, 5),
		Entry("Conflict: double asterisk with single asterisk", []string{
			"/abc/{**}/ghi",
			"/abc/{*}/ghi",
			"/{*}/{*}/ghi",
		}, 2),
		Entry("Conflict: double asterisk with double asterisk", []string{
			"/abc/{**}/def/ghi",
			"/abc/{**}/ghi",
			"/abc/{**}",
		}, 2),
		Entry("No conflict: exact paths", []string{
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
