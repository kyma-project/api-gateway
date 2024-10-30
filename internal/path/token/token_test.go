package token_test

import (
	. "github.com/kyma-project/api-gateway/internal/path/token"

	"testing"
)

func TestTableParsePath(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		tokens []Token
	}{
		{
			name: "ident",
			path: "/path",
			tokens: []Token{
				{Type: IDENT, Literal: "path"},
			},
		},
		{
			name: "ident / empty",
			path: "/path/",
			tokens: []Token{
				{Type: IDENT, Literal: "path"},
				{Type: EMPTY, Literal: ""},
			},
		},
		{
			name: "asterisk",
			path: "/*",
			tokens: []Token{
				{Type: ASTERISK, Literal: "*"},
			},
		},
		{
			name: "braced double asterisk",
			path: "/{**}",
			tokens: []Token{
				{Type: BRACED_DOUBLE_ASTERIX, Literal: "{**}"},
			},
		},
		{
			name: "braced asterisk",
			path: "/{*}",
			tokens: []Token{
				{Type: BRACED_ASTERIX, Literal: "{*}"},
			},
		},
		{
			name: "ident / braced double asterisk",
			path: "/path/{**}",
			tokens: []Token{
				{Type: IDENT, Literal: "path"},
				{Type: BRACED_DOUBLE_ASTERIX, Literal: "{**}"},
			},
		},
		{
			name: "ident / braced asterisk",
			path: "/path/{*}",
			tokens: []Token{
				{Type: IDENT, Literal: "path"},
				{Type: BRACED_ASTERIX, Literal: "{*}"},
			},
		},
		{
			name: "ident / braced double asterisk / ident",
			path: "/path/{**}/path2",
			tokens: []Token{
				{Type: IDENT, Literal: "path"},
				{Type: BRACED_DOUBLE_ASTERIX, Literal: "{**}"},
				{Type: IDENT, Literal: "path2"},
			},
		},
		{
			name: "ident / braced asterisk / ident",
			path: "/path/{*}/path2",
			tokens: []Token{
				{Type: IDENT, Literal: "path"},
				{Type: BRACED_ASTERIX, Literal: "{*}"},
				{Type: IDENT, Literal: "path2"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := TokenizePath(tt.path)
			if len(tokens) != len(tt.tokens) {
				t.Fatalf("expected %d tokens, got %d", len(tt.tokens), len(tokens))
			}
			for i, tok := range tokens {
				if tok.Type != tt.tokens[i].Type {
					t.Fatalf("expected token type %s, got %s", tt.tokens[i].Type, tok.Type)
				}
				if tok.Literal != tt.tokens[i].Literal {
					t.Fatalf("expected token literal %s, got %s", tt.tokens[i].Literal, tok.Literal)
				}
			}
		})
	}
}

func TestTokenListString(t *testing.T) {
	tokens := List{
		{Type: IDENT, Literal: "path"},
		{Type: BRACED_DOUBLE_ASTERIX, Literal: "{**}"},
		{Type: IDENT, Literal: "path2"},
	}
	expected := "/path/{**}/path2"
	if tokens.String() != expected {
		t.Fatalf("expected %s, got %s", expected, tokens.String())
	}
}
