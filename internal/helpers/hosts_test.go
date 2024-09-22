package helpers

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("IsHostFqdn", func() {
	It("Should be false if host is empty", func() {
		Expect(IsHostFqdn("")).To(BeFalse())
	})

	It("Should be true if host is a valid FQDN", func() {
		Expect(IsHostFqdn("host.example.com")).To(BeTrue())
	})

	It("Should be false if host name has uppercase letters", func() {
		Expect(IsHostFqdn("host.exaMple.com")).To(BeFalse())
	})

	It("Should be true if host name has 255 chars", func() {
		Expect(IsHostFqdn(fmt.Sprintf("%s.%s.%s.%s.eee", strings.Repeat("a", 63), strings.Repeat("b", 63),
			strings.Repeat("c", 63), strings.Repeat("d", 59)))).To(BeTrue())
	})

	It("Should be false if host name is longer than 255 chars", func() {
		Expect(IsHostFqdn(fmt.Sprintf("%s.%s.%s.%s.eee", strings.Repeat("a", 63), strings.Repeat("b", 63),
			strings.Repeat("c", 63), strings.Repeat("d", 60)))).To(BeFalse())
	})

	It("Should be false if any domain label is longer than 63 chars", func() {
		Expect(IsHostFqdn(fmt.Sprintf("%s.com", strings.Repeat("a", 64)))).To(BeFalse())
		Expect(IsHostFqdn(fmt.Sprintf("host.%s.com", strings.Repeat("a", 64)))).To(BeFalse())
		Expect(IsHostFqdn(fmt.Sprintf("host.example.%s", strings.Repeat("a", 64)))).To(BeFalse())
	})

	It("Should be false if any domain label is empty", func() {
		Expect(IsHostFqdn("host.")).To(BeFalse())
		Expect(IsHostFqdn(".com")).To(BeFalse())
		Expect(IsHostFqdn(".example.com")).To(BeFalse())
		Expect(IsHostFqdn("host..com")).To(BeFalse())
		Expect(IsHostFqdn("host.example.")).To(BeFalse())
	})

	It("Should be false if top level domain is too short", func() {
		Expect(IsHostFqdn("host.example.c")).To(BeFalse())
	})

	It("Should be false if any domain label contain wrong characters", func() {
		Expect(IsHostFqdn("*host.example.com")).To(BeFalse())
		Expect(IsHostFqdn("ho*st.example.com")).To(BeFalse())
		Expect(IsHostFqdn("host*.example.com")).To(BeFalse())
		Expect(IsHostFqdn("host.*example.com")).To(BeFalse())
		Expect(IsHostFqdn("host.exam*ple.com")).To(BeFalse())
		Expect(IsHostFqdn("host.example*.com")).To(BeFalse())
		Expect(IsHostFqdn("host.example.*com")).To(BeFalse())
		Expect(IsHostFqdn("host.example.co*m")).To(BeFalse())
		Expect(IsHostFqdn("host.example.com*")).To(BeFalse())
	})

	It("Should be false if any segment in host name starts or ends with hyphen", func() {
		Expect(IsHostFqdn("-host.example.com")).To(BeFalse())
		Expect(IsHostFqdn("host-.example.com")).To(BeFalse())
		Expect(IsHostFqdn("host.-example.com")).To(BeFalse())
		Expect(IsHostFqdn("host.example-.com")).To(BeFalse())
		Expect(IsHostFqdn("host.example.-com")).To(BeFalse())
		Expect(IsHostFqdn("host.example.com-")).To(BeFalse())
	})
})

var _ = Describe("IsHostFqdn", func() {
	It("Should be false if host is empty", func() {
		Expect(IsHostShortName("")).To(BeFalse())
	})

	It("Should be true if host is a valid short name", func() {
		Expect(IsHostShortName("short-host--name")).To(BeTrue())
	})

	It("Should be false if short host name has uppercase letters", func() {
		Expect(IsHostShortName("sHort")).To(BeFalse())
	})

	It("Should be true if short host name has 1 char", func() {
		Expect(IsHostShortName("a")).To(BeTrue())
	})

	It("Should be true if short host name has 63 chars", func() {
		Expect(IsHostShortName(strings.Repeat("a", 63))).To(BeTrue())
	})

	It("Should be false if short host name is longer than 63 chars", func() {
		Expect(IsHostShortName(strings.Repeat("a", 64))).To(BeFalse())
	})

	It("Should be false if short host name contains not allowed char", func() {
		Expect(IsHostShortName("short-host.")).To(BeFalse())
		Expect(IsHostShortName(".short-host")).To(BeFalse())
	})
})
