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

		It("should return false for incorrect old gateway name format with wrong DNS suffix", func() {
			gatewayName := "test-gateway.default.svc.cluster"
			Expect(gatewaytranslator.IsOldGatewayNameFormat(gatewayName)).To(BeFalse())
		})

		It("should return false for incorrect old gateway name format with no DNS suffix", func() {
			gatewayName := "test-gateway.default"
			Expect(gatewaytranslator.IsOldGatewayNameFormat(gatewayName)).To(BeFalse())
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

		It("should return an error for incorrect old gateway name format with wrong DNS suffix", func() {
			gatewayName := "test-gateway.default.svc.cluster"
			expectedError := errors.New("gateway name (test-gateway.default.svc.cluster) is not in old gateway format")

			_, err := gatewaytranslator.TranslateGatewayNameToNewFormat(gatewayName)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(expectedError.Error()))
		})

		It("should return an error for incorrect old gateway name format with no DNS suffix", func() {
			gatewayName := "test-gateway.default"
			expectedError := errors.New("gateway name (test-gateway.default) is not in old gateway format")

			_, err := gatewaytranslator.TranslateGatewayNameToNewFormat(gatewayName)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(expectedError.Error()))
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
	})
})
