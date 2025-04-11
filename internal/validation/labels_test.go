package validation

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func entry(value string, valid bool) TableEntry {
	return Entry(value, value, valid)
}

var _ = Describe("Label validation function", func() {

	var commonEntries = []TableEntry{
		entry(" we-trim-whitespaces   ", true),
		entry("a1b", true),
		entry("aaa", true),
		entry("1aa", true),
		entry("2.a-a", true),
		entry("a.3_a", true),
		entry("s-CAPITAL-L3TT3RS-ARE-0K", true),
		entry("2.a=a", false), //Invalid character: "="
		entry("2.a_ENDS-WITH-DOT.", false),
		entry("2.a_ENDS-WITH-DASH-", false),
		entry("2.a_ENDS-WITH-UNDERSCORE_", false),
		entry("LEN_64_IS_TOO_LONG-.bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", false),
		entry("len_63_is_ok-.bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", true),
		entry("ENDS-WITH-DASH-------------------------------------------------", false),
		entry("many-dashes-are-allowed---------------------------------------b", true),
		entry("many-dots-are-allowed-also---..............-------------------b", true),
	}

	Describe("validateLabelValue() should correctly", func() {

		entries := []TableEntry{
			entry("", true), //Empty is OK
		}
		entries = append(entries, commonEntries...)

		DescribeTable("validate label values",

			func(labelValue string, shouldBeValid bool) {
				err := validateLabelValue(labelValue)
				if err != nil {
					fmt.Println(err)
				}
				Expect(err == nil).To(Equal(shouldBeValid))
			},
			entries,
		)
	})

	Describe("validateLabelKey() should correctly", func() {
		DescribeTable("recognize generally invalid labels",
			func(labelKey string, shouldBeValid bool) {
				err := validateLabelKey(labelKey)
				Expect(err == nil).To(Equal(shouldBeValid))
			},
			entry("", false),
			entry("  ", false),
			entry("/a", false),
			entry("a/", false),
			entry("a/1/b", false),
			entry("//", false),
			entry("a /b", false),
			entry("a/ b", false),
			entry("a / b", false),
			entry("label-with-318-characters-is-to-long-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa-e", false),
			entry("label-with-317-characters-is-ok-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa-x/s_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa-e", true),
		)

		DescribeTable("validate label keys with only \"name\" part",
			func(labelKey string, shouldBeValid bool) {
				err := validateLabelKey(labelKey)
				Expect(err == nil).To(Equal(shouldBeValid))
			}, commonEntries)

		DescribeTable("validate label keys with \"prefix\" part",
			func(labelKey string, shouldBeValid bool) {
				err := validateLabelKey(labelKey)
				Expect(err == nil).To(Equal(shouldBeValid))
			},
			entry("a/b", true),
			entry(" a/b ", true),
			entry("a1/a1b", true),
			entry("abc.def/aaa", true),
			entry("a.b/a", true),
			entry("a.b/a-THIS-IS-OK", true),
			entry("a.b.c/a", true),
			entry("a-b.c/a", true),
			entry("velero.io/exclude-from-backup", true),
			entry("a..b/a-DOUBLE-DOT-IN-PREFIX", false),
			entry("a-b-.c/a-DASH-DOT-SEQUENCE-IN-PREFIX", false),
			entry("a.b.c-/a-DASH-AS-LAST-CHARACTER-IN-PREFIX", false),
			entry("a_b.c/a-UNDERSCORE-IN-PREFIX", false),
			entry("a.B.c-/a-CAPITAL-LETTER-IN-PREFIX", false),
		)
	})
})
