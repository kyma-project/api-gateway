package metrics

import "github.com/prometheus/client_golang/prometheus"

type ApiGatewayMetrics struct {
	apiRuleObjectModifiedErrorsCounter prometheus.Counter
}

func NewApiGatewayMetrics() *ApiGatewayMetrics {
	return &ApiGatewayMetrics{
		apiRuleObjectModifiedErrorsCounter: prometheus.NewCounter(prometheus.CounterOpts{
			Name:      "api_rule_object_modified_errors_total",
			Namespace: "api_gateway",
			Help:      "The total number of errors that occurred while modifying the APIRule object",
		}),
	}
}

func (m *ApiGatewayMetrics) IncreaseApiRuleObjectModifiedErrorsCounter() {
	m.apiRuleObjectModifiedErrorsCounter.Inc()
}
