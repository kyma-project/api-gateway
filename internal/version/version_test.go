package version

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("GetModuleVersion", func() {
	It("Returns default version if not set", func() {
		Expect(GetModuleVersion()).To(Equal(defaultModuleVersion))
	})
})
