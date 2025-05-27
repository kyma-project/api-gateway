package token

import (
	"strings"
)

type Type string

type Token struct {
	Type    Type
	Literal string
}

const (
	Ident = "IDENT"
	Empty = "EMPTY"

	BracedDoubleAsterix = "{**}"
	BracedAsterix       = "{*}"
	Asterix             = "*"

	Separator = "/"

	EmptyLiteral = ""
)

type List []Token

func (tk List) String() string {
	var sb strings.Builder
	for _, t := range tk {
		sb.WriteString(Separator)
		sb.WriteString(t.Literal)
	}
	return sb.String()
}

func TokenizePath(apiPath string) []Token {
	var tokens []Token
	apiPath = strings.TrimLeft(apiPath, Separator)

	for _, tok := range strings.Split(apiPath, Separator) {
		switch tok {
		case "":
			tokens = append(tokens, Token{Type: Empty, Literal: EmptyLiteral})
		case BracedAsterix:
			tokens = append(tokens, Token{Type: BracedAsterix, Literal: BracedAsterix})
		case BracedDoubleAsterix:
			tokens = append(tokens, Token{Type: BracedDoubleAsterix, Literal: BracedDoubleAsterix})
		case Asterix:
			tokens = append(tokens, Token{Type: Asterix, Literal: Asterix})
		default:
			tokens = append(tokens, Token{Type: Ident, Literal: tok})
		}
	}

	return tokens
}
