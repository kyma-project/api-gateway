package helpers

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("IsFqdnOrWildcardHostName", func() {
	It("Should be false if host is empty", func() {
		Expect(IsFqdnOrWildcardHostName("")).To(BeFalse())
	})

	It("Should be true if host is a valid FQDN", func() {
		Expect(IsFqdnOrWildcardHostName("host.example.com")).To(BeTrue())
	})

	It("Should be true if host is a valid FQDN with 1 char labels", func() {
		Expect(IsFqdnOrWildcardHostName("a.b.ca")).To(BeTrue())
	})

	It("Should be false if host name has uppercase letters", func() {
		Expect(IsFqdnOrWildcardHostName("host.exaMple.com")).To(BeFalse())
	})

	It("Should be true if host name has 255 chars", func() {
		Expect(IsFqdnOrWildcardHostName(fmt.Sprintf("%s.%s.%s.%s.eee", strings.Repeat("a", 63), strings.Repeat("b", 63),
			strings.Repeat("c", 63), strings.Repeat("d", 59)))).To(BeTrue())
	})

	It("Should be false if host name is longer than 255 chars", func() {
		Expect(IsFqdnOrWildcardHostName(fmt.Sprintf("%s.%s.%s.%s.eee", strings.Repeat("a", 63), strings.Repeat("b", 63),
			strings.Repeat("c", 63), strings.Repeat("d", 60)))).To(BeFalse())
	})

	It("Should be false if any domain label is longer than 63 chars", func() {
		Expect(IsFqdnOrWildcardHostName(fmt.Sprintf("%s.com", strings.Repeat("a", 64)))).To(BeFalse())
		Expect(IsFqdnOrWildcardHostName(fmt.Sprintf("host.%s.com", strings.Repeat("a", 64)))).To(BeFalse())
		Expect(IsFqdnOrWildcardHostName(fmt.Sprintf("host.example.%s", strings.Repeat("a", 64)))).To(BeFalse())
	})

	It("Should be false if any domain label is empty", func() {
		Expect(IsFqdnOrWildcardHostName("host.")).To(BeFalse())
		Expect(IsFqdnOrWildcardHostName(".com")).To(BeFalse())
		Expect(IsFqdnOrWildcardHostName(".example.com")).To(BeFalse())
		Expect(IsFqdnOrWildcardHostName("host..com")).To(BeFalse())
		Expect(IsFqdnOrWildcardHostName("host.example.")).To(BeFalse())
	})

	It("Should be false if top level domain is too short", func() {
		Expect(IsFqdnOrWildcardHostName("host.example.c")).To(BeFalse())
	})

	It("Should be false if any domain label contain wrong characters", func() {
		Expect(IsFqdnOrWildcardHostName("*host.example.com")).To(BeFalse())
		Expect(IsFqdnOrWildcardHostName("ho*st.example.com")).To(BeFalse())
		Expect(IsFqdnOrWildcardHostName("host*.example.com")).To(BeFalse())
		Expect(IsFqdnOrWildcardHostName("host.*example.com")).To(BeFalse())
		Expect(IsFqdnOrWildcardHostName("host.exam*ple.com")).To(BeFalse())
		Expect(IsFqdnOrWildcardHostName("host.example*.com")).To(BeFalse())
		Expect(IsFqdnOrWildcardHostName("host.example.*com")).To(BeFalse())
		Expect(IsFqdnOrWildcardHostName("host.example.co*m")).To(BeFalse())
		Expect(IsFqdnOrWildcardHostName("host.example.com*")).To(BeFalse())
	})

	It("Should be true if host is a valid wildcard domain", func() {
		Expect(IsFqdnOrWildcardHostName("*.example.com")).To(BeTrue())
		Expect(IsFqdnOrWildcardHostName("*.sub.example.com")).To(BeTrue())
		Expect(IsFqdnOrWildcardHostName("*.local.kyma.dev")).To(BeTrue())
	})

	It("Should be false if wildcard domain is invalid", func() {
		Expect(IsFqdnOrWildcardHostName("*.com")).To(BeFalse())
		Expect(IsFqdnOrWildcardHostName("*.c")).To(BeFalse())
		Expect(IsFqdnOrWildcardHostName("*example.com")).To(BeFalse())
		Expect(IsFqdnOrWildcardHostName("*.")).To(BeFalse())
		Expect(IsFqdnOrWildcardHostName("*.example.")).To(BeFalse())
		Expect(IsFqdnOrWildcardHostName("**.example.com")).To(BeFalse())
	})

	It("Should be false if any segment in host name starts or ends with hyphen", func() {
		Expect(IsFqdnOrWildcardHostName("-host.example.com")).To(BeFalse())
		Expect(IsFqdnOrWildcardHostName("host-.example.com")).To(BeFalse())
		Expect(IsFqdnOrWildcardHostName("host.-example.com")).To(BeFalse())
		Expect(IsFqdnOrWildcardHostName("host.example-.com")).To(BeFalse())
		Expect(IsFqdnOrWildcardHostName("host.example.-com")).To(BeFalse())
		Expect(IsFqdnOrWildcardHostName("host.example.com-")).To(BeFalse())
	})

	It("Should be false if top level domain is too short", func() {
		Expect(IsFqdnOrWildcardHostName("example.c")).To(BeFalse())
	})
})

var _ = Describe("IsShortHostName", func() {
	It("Should be false if host is empty", func() {
		Expect(IsShortHostName("")).To(BeFalse())
	})

	It("Should be true if host is a valid short host name", func() {
		Expect(IsShortHostName("short-host--name")).To(BeTrue())
	})

	It("Should be false if short host name has uppercase letters", func() {
		Expect(IsShortHostName("sHort")).To(BeFalse())
	})

	It("Should be true if short host name has 1 char", func() {
		Expect(IsShortHostName("a")).To(BeTrue())
	})

	It("Should be true if short host name has 63 chars", func() {
		Expect(IsShortHostName(strings.Repeat("a", 63))).To(BeTrue())
	})

	It("Should be false if short host name is longer than 63 chars", func() {
		Expect(IsShortHostName(strings.Repeat("a", 64))).To(BeFalse())
	})

	It("Should be false if short host name contains not allowed char", func() {
		Expect(IsShortHostName("short-host.")).To(BeFalse())
		Expect(IsShortHostName(".short-host")).To(BeFalse())
	})
})
