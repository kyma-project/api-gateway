package helpers

import (
	"errors"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/mod/semver"
)

var _ = Describe("GetModuleVersion", func() {
	It("Should return a valid version if succesfully read", func() {
		ReadVersionFileHandle = func(name string) ([]byte, error) {
			return []byte("1.0.0"), nil
		}
		version := GetModuleVersion()
		Expect(semver.IsValid(fmt.Sprintf("v%s", version))).To(BeTrue())
	})

	It("Should return unknown version if could not read it", func() {
		ReadVersionFileHandle = func(name string) ([]byte, error) {
			return []byte{}, errors.New("could not read")
		}
		version := GetModuleVersion()
		Expect(version).To(Equal("unknown"))
	})
})
