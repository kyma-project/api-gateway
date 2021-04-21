package validation

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ValidateDomainName function", func() {

	It("Should return true for kyma.local domain", func() {
		//given
		testDomain := "kyma.local"

		//when
		valid := ValidateDomainName(testDomain)

		//then
		Expect(valid).To(BeTrue())
	})

	It("Should return true for valid complicated domain", func() {
		//given
		testDomain := "gke-upgrade-pr-5776-47nlgu1ch0.a.build.kyma-project.io"

		//when
		valid := ValidateDomainName(testDomain)

		//then
		Expect(valid).To(BeTrue())
	})

	It("Should return true for foo domain", func() {
		//given
		testDomain := "foo"

		//when
		valid := ValidateDomainName(testDomain)

		//then
		Expect(valid).To(BeTrue())
	})

	It("Should return false for subdomain starting with -", func() {
		//given
		testDomain := "subdomain.-example.com"

		//when
		valid := ValidateDomainName(testDomain)

		//then
		Expect(valid).To(BeFalse())
	})

	It("Should return false for domain containing with /", func() {
		//given
		testDomain := "example.com/parameter"

		//when
		valid := ValidateDomainName(testDomain)

		//then
		Expect(valid).To(BeFalse())
	})

})
var _ = Describe("ValidateServiceName function", func() {

	It("Should return true for kubernetes.default service", func() {
		//given
		testService := "kubernetes.default"

		//when
		valid := ValidateServiceName(testService)

		//then
		Expect(valid).To(BeTrue())
	})

	It("Should return false for service containing more than one .", func() {
		//given
		testService := "kubernetes.default.example"

		//when
		valid := ValidateServiceName(testService)

		//then
		Expect(valid).To(BeFalse())
	})

	It("Should return false for service name ending with -", func() {
		//given
		testService := "kubernetes-.default"

		//when
		valid := ValidateServiceName(testService)

		//then
		Expect(valid).To(BeFalse())
	})

	It("Should return false for service containing forbidden character /", func() {
		//given
		testService := "kubernetes.de/fault"

		//when
		valid := ValidateServiceName(testService)

		//then
		Expect(valid).To(BeFalse())
	})
})
