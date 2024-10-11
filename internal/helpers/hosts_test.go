package helpers

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("IsFqdnHostName", func() {
	It("Should be false if host is empty", func() {
		Expect(IsFqdnHostName("")).To(BeFalse())
	})

	It("Should be true if host is a valid FQDN", func() {
		Expect(IsFqdnHostName("host.example.com")).To(BeTrue())
	})

	It("Should be true if host is a valid FQDN with 1 char labels", func() {
		Expect(IsFqdnHostName("a.b.ca")).To(BeTrue())
	})

	It("Should be false if host name has uppercase letters", func() {
		Expect(IsFqdnHostName("host.exaMple.com")).To(BeFalse())
	})

	It("Should be true if host name has 255 chars", func() {
		Expect(IsFqdnHostName(fmt.Sprintf("%s.%s.%s.%s.eee", strings.Repeat("a", 63), strings.Repeat("b", 63),
			strings.Repeat("c", 63), strings.Repeat("d", 59)))).To(BeTrue())
	})

	It("Should be false if host name is longer than 255 chars", func() {
		Expect(IsFqdnHostName(fmt.Sprintf("%s.%s.%s.%s.eee", strings.Repeat("a", 63), strings.Repeat("b", 63),
			strings.Repeat("c", 63), strings.Repeat("d", 60)))).To(BeFalse())
	})

	It("Should be false if any domain label is longer than 63 chars", func() {
		Expect(IsFqdnHostName(fmt.Sprintf("%s.com", strings.Repeat("a", 64)))).To(BeFalse())
		Expect(IsFqdnHostName(fmt.Sprintf("host.%s.com", strings.Repeat("a", 64)))).To(BeFalse())
		Expect(IsFqdnHostName(fmt.Sprintf("host.example.%s", strings.Repeat("a", 64)))).To(BeFalse())
	})

	It("Should be false if any domain label is empty", func() {
		Expect(IsFqdnHostName("host.")).To(BeFalse())
		Expect(IsFqdnHostName(".com")).To(BeFalse())
		Expect(IsFqdnHostName(".example.com")).To(BeFalse())
		Expect(IsFqdnHostName("host..com")).To(BeFalse())
		Expect(IsFqdnHostName("host.example.")).To(BeFalse())
	})

	It("Should be false if top level domain is too short", func() {
		Expect(IsFqdnHostName("host.example.c")).To(BeFalse())
	})

	It("Should be false if any domain label contain wrong characters", func() {
		Expect(IsFqdnHostName("*host.example.com")).To(BeFalse())
		Expect(IsFqdnHostName("ho*st.example.com")).To(BeFalse())
		Expect(IsFqdnHostName("host*.example.com")).To(BeFalse())
		Expect(IsFqdnHostName("host.*example.com")).To(BeFalse())
		Expect(IsFqdnHostName("host.exam*ple.com")).To(BeFalse())
		Expect(IsFqdnHostName("host.example*.com")).To(BeFalse())
		Expect(IsFqdnHostName("host.example.*com")).To(BeFalse())
		Expect(IsFqdnHostName("host.example.co*m")).To(BeFalse())
		Expect(IsFqdnHostName("host.example.com*")).To(BeFalse())
	})

	It("Should be false if any segment in host name starts or ends with hyphen", func() {
		Expect(IsFqdnHostName("-host.example.com")).To(BeFalse())
		Expect(IsFqdnHostName("host-.example.com")).To(BeFalse())
		Expect(IsFqdnHostName("host.-example.com")).To(BeFalse())
		Expect(IsFqdnHostName("host.example-.com")).To(BeFalse())
		Expect(IsFqdnHostName("host.example.-com")).To(BeFalse())
		Expect(IsFqdnHostName("host.example.com-")).To(BeFalse())
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
