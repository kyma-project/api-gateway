package token

import (
	"strings"
)

type TokenType string

type Token struct {
	Type    TokenType
	Literal string
}

const (
	IDENT = "IDENT"
	EMPTY = "EMPTY"

	BRACED_DOUBLE_ASTERIX = "{**}"
	BRACED_ASTERIX        = "{*}"
	ASTERISK              = "*"

	SEPARATOR = "/"
)

type List []Token

func (tk List) String() string {
	var sb strings.Builder
	for _, t := range tk {
		sb.WriteString(SEPARATOR)
		sb.WriteString(t.Literal)
	}
	return sb.String()
}

func TokenizePath(apiPath string) []Token {
	var tokens []Token
	apiPath = strings.TrimLeft(apiPath, SEPARATOR)

	for _, tok := range strings.Split(apiPath, SEPARATOR) {
		switch {
		case tok == "":
			tokens = append(tokens, Token{Type: EMPTY, Literal: ""})
		case tok == BRACED_ASTERIX:
			tokens = append(tokens, Token{Type: BRACED_ASTERIX, Literal: BRACED_ASTERIX})
		case tok == BRACED_DOUBLE_ASTERIX:
			tokens = append(tokens, Token{Type: BRACED_DOUBLE_ASTERIX, Literal: BRACED_DOUBLE_ASTERIX})
		case tok == ASTERISK:
			tokens = append(tokens, Token{Type: ASTERISK, Literal: ASTERISK})
		default:
			tokens = append(tokens, Token{Type: IDENT, Literal: tok})
		}
	}
	return tokens
}
