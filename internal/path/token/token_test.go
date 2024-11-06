package token_test

import (
	. "github.com/kyma-project/api-gateway/internal/path/token"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"testing"
)

func TestToken(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Token Suite")
}

var _ = Describe("Token", func() {
	DescribeTable("ParsePath",
		func(path string, expectedTokens []Token) {
			tokens := TokenizePath(path)
			Expect(tokens).To(HaveLen(len(expectedTokens)))
			for i, tok := range tokens {
				Expect(tok.Type).To(Equal(expectedTokens[i].Type))
				Expect(tok.Literal).To(Equal(expectedTokens[i].Literal))
			}
		},

		Entry(
			"ident",
			"/path",
			[]Token{{Type: Ident, Literal: "path"}},
		),
		Entry(
			"ident / empty",
			"/path/",
			[]Token{{Type: Ident, Literal: "path"}, {Type: Empty, Literal: ""}},
		),
		Entry(
			"asterisk",
			"/*",
			[]Token{{Type: Asterix, Literal: "*"}},
		),
		Entry(
			"braced double asterisk",
			"/{**}",
			[]Token{{Type: BracedDoubleAsterix, Literal: "{**}"}},
		),
		Entry(
			"braced asterisk",
			"/{*}",
			[]Token{{Type: BracedAsterix, Literal: "{*}"}},
		),
		Entry(
			"ident / braced double asterisk",
			"/path/{**}",
			[]Token{{Type: Ident, Literal: "path"}, {Type: BracedDoubleAsterix, Literal: "{**}"}},
		),
		Entry(
			"ident / braced asterisk",
			"/path/{*}",
			[]Token{{Type: Ident, Literal: "path"}, {Type: BracedAsterix, Literal: "{*}"}},
		),
		Entry(
			"ident / braced double asterisk / ident",
			"/path/{**}/path2",
			[]Token{{Type: Ident, Literal: "path"}, {Type: BracedDoubleAsterix, Literal: "{**}"}, {Type: Ident, Literal: "path2"}},
		),
		Entry(
			"ident / braced asterisk / ident",
			"/path/{*}/path2",
			[]Token{{Type: Ident, Literal: "path"}, {Type: BracedAsterix, Literal: "{*}"}, {Type: Ident, Literal: "path2"}},
		),
	)

	Describe("TokenListString", func() {
		It("should convert token list to string", func() {
			tokens := List{
				{Type: Ident, Literal: "path"},
				{Type: BracedDoubleAsterix, Literal: "{**}"},
				{Type: Ident, Literal: "path2"},
			}
			expected := "/path/{**}/path2"
			Expect(tokens.String()).To(Equal(expected))
		})
	})
})
