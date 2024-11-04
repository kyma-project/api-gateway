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
			[]Token{{Type: IDENT, Literal: "path"}},
		),
		Entry(
			"ident / empty",
			"/path/",
			[]Token{{Type: IDENT, Literal: "path"}, {Type: EMPTY, Literal: ""}},
		),
		Entry(
			"asterisk",
			"/*",
			[]Token{{Type: ASTERISK, Literal: "*"}},
		),
		Entry(
			"braced double asterisk",
			"/{**}",
			[]Token{{Type: BRACED_DOUBLE_ASTERIX, Literal: "{**}"}},
		),
		Entry(
			"braced asterisk",
			"/{*}",
			[]Token{{Type: BRACED_ASTERIX, Literal: "{*}"}},
		),
		Entry(
			"ident / braced double asterisk",
			"/path/{**}",
			[]Token{{Type: IDENT, Literal: "path"}, {Type: BRACED_DOUBLE_ASTERIX, Literal: "{**}"}},
		),
		Entry(
			"ident / braced asterisk",
			"/path/{*}",
			[]Token{{Type: IDENT, Literal: "path"}, {Type: BRACED_ASTERIX, Literal: "{*}"}},
		),
		Entry(
			"ident / braced double asterisk / ident",
			"/path/{**}/path2",
			[]Token{{Type: IDENT, Literal: "path"}, {Type: BRACED_DOUBLE_ASTERIX, Literal: "{**}"}, {Type: IDENT, Literal: "path2"}},
		),
		Entry(
			"ident / braced asterisk / ident",
			"/path/{*}/path2",
			[]Token{{Type: IDENT, Literal: "path"}, {Type: BRACED_ASTERIX, Literal: "{*}"}, {Type: IDENT, Literal: "path2"}},
		),
	)

	Describe("TokenListString", func() {
		It("should convert token list to string", func() {
			tokens := List{
				{Type: IDENT, Literal: "path"},
				{Type: BRACED_DOUBLE_ASTERIX, Literal: "{**}"},
				{Type: IDENT, Literal: "path2"},
			}
			expected := "/path/{**}/path2"
			Expect(tokens.String()).To(Equal(expected))
		})
	})
})
