// Package segment_trie provides a trie data structure for storing URL paths.
// It has the main purpose of checking for path collisions.
//
// The SegmentTrie is a trie data structure that segments paths into tokens.
// Stored paths can contain the `{*}` and `{**}` operators:
//   - operator `{*}` is used to match a single segment in a path, and may include a prefix and/or suffix.
//   - operator `{**}` is used any number of segments in a path, and may include a prefix and/or suffix.
//     It must be the last operator in the stored path (this is not validated but is assumed to be true).
//
// It does two things that are not done by a regular trie:
//   - Nodes that are pointed to by `{**}` don't store their children,
//     instead they store the path suffix that exists after `{**}`.
//     Conflicts are checked by comparing the suffixes of paths with the same segments that precede `{**}`.
//     This can be done, since the `{**}` operator must be the last in the path.
//   - `{*}` nodes are stored just like any other exact segment node, but are always included in conflict search.
package segment_trie

import (
	"errors"
	"github.com/kyma-project/api-gateway/internal/path/token"
)

type SegmentTrie struct {
	Root *Node
}

type Node struct {
	EndNode  bool             `json:"-"`
	Children map[string]*Node `json:"children"`
	Suffixes []string         `json:"suffixes"`
}

func New() *SegmentTrie {
	return &SegmentTrie{
		Root: &Node{
			EndNode:  false,
			Children: map[string]*Node{},
		},
	}
}

func (t *SegmentTrie) InsertAndCheckCollisions(tokens []token.Token) error {
	if len(tokens) == 0 {
		return nil
	}

	node := t.Root
	if len(t.Root.Children) != 0 {
		pathExist := findExistingPath(node, tokens, 0)
		if pathExist {
			return errors.New("path collision detected")
		}
	}

	for i, tok := range tokens {
		if tok.Type == token.BracedDoubleAsterix {
			if n, ok := node.Children["{**}"]; !ok {
				node.Children["{**}"] = &Node{
					EndNode:  i == len(tokens)-1,
					Children: make(map[string]*Node),
					Suffixes: []string{token.List(tokens[i+1:]).String()},
				}
			} else {
				suffixString := token.List(tokens[i+1:]).String()
				n.Suffixes = append(n.Suffixes, suffixString)
			}
		} else {
			if _, ok := node.Children[tok.Literal]; !ok {
				node.Children[tok.Literal] = &Node{
					EndNode:  i == len(tokens)-1,
					Children: make(map[string]*Node),
				}
			}
		}

		node = node.Children[tok.Literal]
	}

	return nil
}

func findExistingPath(node *Node, tokens []token.Token, cur int) bool {
	if cur >= len(tokens) {
		return node.EndNode
	}

	tok := tokens[cur]

	if next, ok := node.Children[tok.Literal]; ok {
		if findExistingPath(next, tokens, cur+1) {
			return true
		}
	}

	if next, ok := node.Children["{*}"]; ok {
		if findExistingPath(next, tokens, cur+1) {
			return true
		}
	}

	if next, ok := node.Children["{**}"]; ok {
		suffix := token.List(tokens[cur+1:]).String()
		for _, s := range next.Suffixes {
			if s == suffix {
				return true
			}
			if s == "" && suffix != "" {
				return true
			}
		}
	}

	return false
}
