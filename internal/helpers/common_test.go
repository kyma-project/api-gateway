package helpers

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/mod/semver"
)

var _ = Describe("GetModuleVersion", func() {
	It("Should read VERSION from file", func() {
		version := GetModuleVersion()
		Expect(semver.IsValid(fmt.Sprintf("v%s", version))).To(BeTrue())
	})
})
