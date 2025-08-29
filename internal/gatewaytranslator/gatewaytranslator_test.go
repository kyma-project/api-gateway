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

		It("should return false for incorrect old gateway name format with no namespace", func() {
			gatewayName := "test-gateway.svc.cluster.local"
			Expect(gatewaytranslator.IsOldGatewayNameFormat(gatewayName)).To(BeFalse())
		})
	})

	Describe("TranslateGatewayNameToNewFormat", func() {
		It("should correctly translate a gateway name in old format", func() {
			gatewayName := "test-gateway.default.svc.cluster.local"
			expectedNewName := "default/test-gateway"

			newGatewayName, err := gatewaytranslator.TranslateGatewayNameToNewFormat(gatewayName)

			Expect(err).ToNot(HaveOccurred())
			Expect(newGatewayName).To(Equal(expectedNewName))
		})

		It("should return an error for incorrect old gateway name format with no DNS suffix", func() {
			gatewayName := "test-gateway.default"
			expectedNewName := "default/test-gateway"
			newGatewayName, err := gatewaytranslator.TranslateGatewayNameToNewFormat(gatewayName)

			Expect(err).ToNot(HaveOccurred())
			Expect(newGatewayName).To(Equal(expectedNewName))
		})

		It("should return an error for incorrect old gateway name format with no namespace", func() {
			gatewayName := "test-gateway.svc.cluster.local"
			expectedError := errors.New("gateway name (test-gateway.svc.cluster.local) is not in old gateway format")

			_, err := gatewaytranslator.TranslateGatewayNameToNewFormat(gatewayName)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(expectedError.Error()))
		})

		It("should return an error for incorrect old gateway name format that is too long", func() {
			gatewayName := "to-long.test-gateway.default.svc.cluster.local"
			expectedError := errors.New("gateway name (to-long.test-gateway.default.svc.cluster.local) is not in old gateway format")

			_, err := gatewaytranslator.TranslateGatewayNameToNewFormat(gatewayName)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(expectedError.Error()))
		})
		It("should return an error for incorrect old gateway name format that do not have specified namespace of gateway", func() {
			gatewayName := "test-gateway.svc.cluster.local"
			expectedError := errors.New("gateway name (test-gateway.svc.cluster.local) is not in old gateway format")

			_, err := gatewaytranslator.TranslateGatewayNameToNewFormat(gatewayName)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(expectedError.Error()))
		})
		It("should return an error for incorrect old gateway name format that do not have specified namespace of gateway", func() {
			gatewayName := "test-gateway.svc.cluster.local"
			expectedError := errors.New("gateway name (test-gateway.svc.cluster.local) is not in old gateway format")

			_, err := gatewaytranslator.TranslateGatewayNameToNewFormat(gatewayName)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(expectedError.Error()))
		})

		It("should correctly translate a gateway name in old format with shorter name without .local ", func() {
			gatewayName := "test-gateway.default.svc.cluster"
			expectedNewName := "default/test-gateway"

			newGatewayName, err := gatewaytranslator.TranslateGatewayNameToNewFormat(gatewayName)

			Expect(err).ToNot(HaveOccurred())
			Expect(newGatewayName).To(Equal(expectedNewName))
		})

		It("should correctly translate a gateway name in old format with shorter name without .cluster.local ", func() {
			gatewayName := "test-gateway.default.svc"
			expectedNewName := "default/test-gateway"

			newGatewayName, err := gatewaytranslator.TranslateGatewayNameToNewFormat(gatewayName)

			Expect(err).ToNot(HaveOccurred())
			Expect(newGatewayName).To(Equal(expectedNewName))
		})

		It("should correctly translate a gateway name in old format with shorter name without .svc.cluster.local ", func() {
			gatewayName := "test-gateway.default"
			expectedNewName := "default/test-gateway"

			newGatewayName, err := gatewaytranslator.TranslateGatewayNameToNewFormat(gatewayName)

			Expect(err).ToNot(HaveOccurred())
			Expect(newGatewayName).To(Equal(expectedNewName))
		})
		It("should correctly translate a gateway name in old format without namespace specified", func() {
			gatewayName := "test-gateway"
			expectedNewName := "default/test-gateway"

			newGatewayName, err := gatewaytranslator.TranslateGatewayNameToNewFormat(gatewayName)

			Expect(err).ToNot(HaveOccurred())
			Expect(newGatewayName).To(Equal(expectedNewName))
		})
	})
})
