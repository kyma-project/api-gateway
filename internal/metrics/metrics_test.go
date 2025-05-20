package metrics

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	dto "github.com/prometheus/client_model/go"
)

var _ = Describe("ApiGateway metrics", func() {
	It("ApiRule object modified errors should increase the counter", func() {
		metrics := NewApiGatewayMetrics()
		metrics.IncreaseApiRuleObjectModifiedErrorsCounter()

		metric := metrics.apiRuleObjectModifiedErrorsCounter

		pb := &dto.Metric{}
		err := metric.Write(pb)

		Expect(err).To(BeNil())
		Expect(pb.GetCounter().GetValue()).To(Equal(float64(1)))
	})
})
