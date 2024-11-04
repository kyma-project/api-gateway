package segment_trie_test

import (
	"github.com/kyma-project/api-gateway/internal/path/token"
	"github.com/thoas/go-funk"
	"strings"
	"testing"

	. "github.com/kyma-project/api-gateway/internal/path/segment_trie"
)

func TestTableConflictChecking(t *testing.T) {
	tt := []struct {
		name           string
		paths          []string
		conflictNumber int
	}{
		{
			name: "Conflict: exact with single asterisk",
			paths: []string{
				"/abc/def/ghi",
				"/abc/{*}/ghi",
			},
			conflictNumber: 1,
		},
		{
			name: "Conflict: exact with exact",
			paths: []string{
				"/abc/def/ghi",
				"/abc/def/ghi",
			},
			conflictNumber: 1,
		},
		{
			name: "Conflict: exact with double asterisk",
			paths: []string{
				"/abc/def/ghi",
				"/abc/{**}/ghi",
				"/abc/{**}",
				"/{**}/ghi",
				"/{**}",
				"/{**}/def/ghi",
			},
			conflictNumber: 5,
		},
		{
			name: "Conflict: double asterisk with single asterisk",
			paths: []string{
				"/abc/{**}/ghi",
				"/abc/{*}/ghi",
				"/{*}/{*}/ghi",
			},
			conflictNumber: 2,
		},
		{
			name: "Conflict: double asterisk with double asterisk",
			paths: []string{
				"/abc/{**}/def/ghi",
				"/abc/{**}/ghi",
				"/abc/{**}",
			},
			conflictNumber: 2,
		},
		{
			name: "No conflict: exact paths",
			paths: []string{
				"/abc/def/ghi",
				"/abc/def/foo",
				"/def/ghi",
			},
			conflictNumber: 0,
		},
		{
			name: "No conflict: paths with single asterisk, but different suffix",
			paths: []string{
				"/abc/{*}/def",
				"/abc/{*}/ghi",
				"/abc/{*}/foo/bar",
				"/abc/{*}",
			},
			conflictNumber: 0,
		},
		{
			name: "No conflict: paths with double asterisk, but different suffix",
			paths: []string{
				"/abc/{**}/def",
				"/abc/{**}/ghi",
				"/abc/{**}/foo/bar",
			},
			conflictNumber: 0,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			trie := New()
			errNumber := 0
			for _, path := range tc.paths {
				tokenizedPath := token.TokenizePath(path)
				err := trie.InsertAndCheckCollisions(tokenizedPath)
				if err != nil {
					errNumber++
				}
			}
			if errNumber != tc.conflictNumber {
				t.Logf("Tree: %s", trie.String())
				t.Errorf("Expected %d conflicts, got %d", tc.conflictNumber, errNumber)
			}
		})
	}
}

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
