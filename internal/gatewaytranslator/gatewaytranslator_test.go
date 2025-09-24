package gatewaytranslator_test

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kyma-project/api-gateway/internal/gatewaytranslator"
)

var _ = Describe("Gateway translator ", func() {
	Describe("IsOldGatewayNameFormat", func() {
		It("should return true for correct old gateway name format", func() {
			gatewayName := "test-gateway.default.svc.cluster.local"
			Expect(gatewaytranslator.IsOldGatewayNameFormat(gatewayName)).To(BeTrue())
		})

		It("should return true for correct shorter old gateway name format ", func() {
			gatewayName := "test-gateway.default.svc.cluster"
			Expect(gatewaytranslator.IsOldGatewayNameFormat(gatewayName)).To(BeTrue())
		})

		It("should return true for correct shorter old gateway name format without full DNS suffix", func() {
			gatewayName := "test-gateway.default"
			Expect(gatewaytranslator.IsOldGatewayNameFormat(gatewayName)).To(BeTrue())
		})
	})

	Describe("TranslateGatewayNameToNewFormat", func() {
		It("should correctly translate a gateway name in old format", func() {
			gatewayName := "test-gateway.default.svc.cluster.local"
			namespace := "default"
			expectedNewName := "default/test-gateway"

			newGatewayName, err := gatewaytranslator.TranslateGatewayNameToNewFormat(gatewayName, namespace)

			Expect(err).ToNot(HaveOccurred())
			Expect(newGatewayName).To(Equal(expectedNewName))
		})

		It("should return converted value for old gateway name format with no DNS suffix", func() {
			gatewayName := "test-gateway.default"
			namespace := "default"
			expectedNewName := "default/test-gateway"

			newGatewayName, err := gatewaytranslator.TranslateGatewayNameToNewFormat(gatewayName, namespace)

			Expect(err).ToNot(HaveOccurred())
			Expect(newGatewayName).To(Equal(expectedNewName))
		})

		It("should return an error for incorrect old gateway name format with no namespace", func() {
			gatewayName := "test-gateway.svc.cluster.local"
			namespace := "default"
			expectedError := errors.New("gateway name (test-gateway.svc.cluster.local) is not in old gateway format")

			_, err := gatewaytranslator.TranslateGatewayNameToNewFormat(gatewayName, namespace)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(expectedError.Error()))
		})

		It("should return an error for incorrect old gateway name format that is too long", func() {
			gatewayName := "to-long.test-gateway.default.svc.cluster.local"
			namespace := "default"
			expectedError := errors.New("gateway name (to-long.test-gateway.default.svc.cluster.local) is not in old gateway format")

			_, err := gatewaytranslator.TranslateGatewayNameToNewFormat(gatewayName, namespace)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(expectedError.Error()))
		})
		It("should correctly translate a gateway name in old format with shorter name without .local ", func() {
			gatewayName := "test-gateway.default.svc.cluster"
			namespace := "default"
			expectedNewName := "default/test-gateway"

			newGatewayName, err := gatewaytranslator.TranslateGatewayNameToNewFormat(gatewayName, namespace)

			Expect(err).ToNot(HaveOccurred())
			Expect(newGatewayName).To(Equal(expectedNewName))
		})

		It("should correctly translate a gateway name in old format with shorter name without .cluster.local ", func() {
			gatewayName := "test-gateway.default.svc"
			namespace := "default"
			expectedNewName := "default/test-gateway"

			newGatewayName, err := gatewaytranslator.TranslateGatewayNameToNewFormat(gatewayName, namespace)

			Expect(err).ToNot(HaveOccurred())
			Expect(newGatewayName).To(Equal(expectedNewName))
		})

		It("should correctly translate a gateway name in old format with shorter name without .svc.cluster.local ", func() {
			gatewayName := "test-gateway.default"
			namespace := "example-namespace"
			expectedNewName := "default/test-gateway"

			newGatewayName, err := gatewaytranslator.TranslateGatewayNameToNewFormat(gatewayName, namespace)

			Expect(err).ToNot(HaveOccurred())
			Expect(newGatewayName).To(Equal(expectedNewName))
		})
		It("should correctly translate a gateway name in old format without namespace specified", func() {
			gatewayName := "test-gateway"
			namespace := "example-namespace"
			expectedNewName := "example-namespace/test-gateway"

			newGatewayName, err := gatewaytranslator.TranslateGatewayNameToNewFormat(gatewayName, namespace)

			Expect(err).ToNot(HaveOccurred())
			Expect(newGatewayName).To(Equal(expectedNewName))
		})
	})
	Describe("IsCorrectNewGatewayNameFormat", func() {
		It("should be correct new gateway format", func() {

			Expect(gatewaytranslator.IsCorrectNewGatewayNameFormat("test-1/test-gateway")).To(BeTrue())
		})

		It("should not be correct new gateway format - missing namespace", func() {
			Expect(gatewaytranslator.IsCorrectNewGatewayNameFormat("test-gateway")).To(BeFalse())
		})

		It("should not be correct new gateway format - invalid characters", func() {
			Expect(gatewaytranslator.IsCorrectNewGatewayNameFormat("test-1/test_gateway")).To(BeFalse())
		})

		It("should not be correct new gateway format - too long", func() {
			Expect(gatewaytranslator.IsCorrectNewGatewayNameFormat("this-is-a-very-long-namespace-name-that-exceeds-the-maximum-length/test-gateway")).To(BeFalse())
		})
	})

})
