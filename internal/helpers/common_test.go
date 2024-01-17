package helpers

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/mod/semver"
)

var _ = Describe("GetModuleVersion", func() {
	It("Should verify and return version if value is valid", func() {
		ReadVersionFileHandle = func(name string) ([]byte, error) {
			return []byte("1.0.0"), nil
		}
		version := GetModuleVersion()
		Expect(semver.IsValid(fmt.Sprintf("v%s", version))).To(BeTrue())
	})

	It("Should verify and return unknown if value is invalid", func() {
		ReadVersionFileHandle = func(name string) ([]byte, error) {
			return []byte("wrong"), nil
		}
		version := GetModuleVersion()
		Expect(semver.IsValid(fmt.Sprintf("v%s", version))).To(BeFalse())
	})
})
