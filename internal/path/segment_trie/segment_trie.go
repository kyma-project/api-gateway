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
	"strings"
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
			if _, ok := node.Children["{**}"]; !ok {
				node.Children["{**}"] = &Node{
					EndNode:  i == len(tokens)-1,
					Children: make(map[string]*Node),
					Suffixes: []string{token.List(tokens[i+1:]).String()},
				}
			} else {
				suffixString := token.List(tokens[i:]).String()
				node.Children["{**}"].Suffixes = append(node.Children["{**}"].Suffixes, suffixString)
			}
		} else {
			if _, ok := node.Children[tok.Literal]; !ok {
				node.Children[tok.Literal] = &Node{
					EndNode:  i == len(tokens)-1,
					Children: make(map[string]*Node),
				}
			}
		}

		if _, ok := node.Children[tok.Literal]; !ok {
			return errors.New("path collision detected")
		}
		node = node.Children[tok.Literal]
	}
	return nil
}

func suffixExist(node *Node, suffix []token.Token, cur int) bool {
	if cur >= len(suffix) {
		return true
	}

	if cnode, ok := node.Children["{**}"]; ok {
		tokensString := token.List(suffix).String()
		for _, suffix := range cnode.Suffixes {
			if strings.HasSuffix(tokensString, suffix) || strings.HasSuffix(suffix, tokensString) {
				return true
			}
		}
	}

	if _, ok := node.Children[suffix[cur].Literal]; ok {
		if suffixExist(node.Children[suffix[cur].Literal], suffix, cur+1) {
			return true
		}
	}

	for k, v := range node.Children {
		if k != "{**}" && suffixExist(v, suffix, cur) {
			return true
		}
	}
	return false
}

func findExistingPath(node *Node, tokens []token.Token, cur int) bool {
	if len(tokens) == 0 {
		return node.EndNode
	}

	if cur >= len(tokens) {
		return node.EndNode
	}

	if len(node.Suffixes) > 0 {
		return hasAnySuffix(tokens, node.Suffixes)
	}

	switch tokens[cur].Type {
	case token.Ident:
		if _, ok := node.Children[tokens[cur].Literal]; ok {
			if findExistingPath(node.Children[tokens[cur].Literal], tokens, cur+1) {
				return true
			}
		}

		if _, ok := node.Children["{*}"]; ok {
			if findExistingPath(node.Children["{*}"], tokens, cur+1) {
				return true
			}
		}

		if _, ok := node.Children["{**}"]; ok {
			if findExistingPath(node.Children["{**}"], tokens, cur+1) {
				return true
			}
		}
	case token.BracedAsterix:
		for _, node := range node.Children {
			if findExistingPath(node, tokens, cur+1) {
				return true
			}
		}
	case token.BracedDoubleAsterix:
		bracedAsterixSuffix := tokens[cur+1:]
		return suffixExist(node, bracedAsterixSuffix, 0)
	}
	return false
}

func hasAnySuffix(tokens token.List, suffixes []string) bool {
	if len(suffixes) == 0 {
		return false
	}
	tokensString := tokens.String()

	for _, suffix := range suffixes {
		if strings.HasSuffix(tokensString, suffix) {
			return true
		}
	}
	return false
}
