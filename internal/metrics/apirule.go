package metrics

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlmetrics "sigs.k8s.io/controller-runtime/pkg/metrics"

	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
)

// APIRuleCollector implements prometheus.Collector so the feature-usage gauges
// are recomputed from the APIRule set every time Prometheus scrapes, rather than
// pushed on each reconcile. This keeps them drift-free and avoids listing
// APIRules in the reconciliation loop.\
type APIRuleCollector struct {
	client                           client.Reader
	APIRuleObjectModifiedErrorsCount prometheus.Counter
	JWTHandlerIstioUsed              prometheus.Gauge

	featureJWTProviderUsed   prometheus.Gauge
	featureExtAuthUsed       prometheus.Gauge
	featureCustomCORSUsed    prometheus.Gauge
	featureCustomHeadersUsed prometheus.Gauge
	featureNoAuthUsed        prometheus.Gauge
}

// Describe implements prometheus.Collector.
func (m *APIRuleCollector) Describe(ch chan<- *prometheus.Desc) {
	m.JWTHandlerIstioUsed.Describe(ch)
	m.APIRuleObjectModifiedErrorsCount.Describe(ch)
	m.featureJWTProviderUsed.Describe(ch)
	m.featureExtAuthUsed.Describe(ch)
	m.featureCustomCORSUsed.Describe(ch)
	m.featureCustomHeadersUsed.Describe(ch)
	m.featureNoAuthUsed.Describe(ch)
}

// Collect implements prometheus.Collector. It is called by Prometheus at scrape
// time: it lists APIRules from the cache, recomputes the feature flags, and
// emits the gauges. A list failure leaves the gauges unreported for this scrape.
func (m *APIRuleCollector) Collect(ch chan<- prometheus.Metric) {
	var apiRuleList gatewayv2alpha1.APIRuleList
	if err := m.client.List(context.Background(), &apiRuleList); err != nil {
		return
	}

	jwt, extAuth, cors, headers, noauth := 0.0, 0.0, 0.0, 0.0, 0.0
	for _, apiRule := range apiRuleList.Items {
		if apiRule.Spec.CorsPolicy != nil {
			cors += 1.0
		}
		for _, rule := range apiRule.Spec.Rules {
			if rule.Jwt != nil {
				jwt += 1.0
			}
			if rule.ExtAuth != nil {
				extAuth += 1.0
			}
			if rule.Request != nil && len(rule.Request.Headers) > 0 {
				headers += 1.0
			}
			if rule.NoAuth != nil && *rule.NoAuth {
				noauth += 1.0
			}
		}
	}

	m.featureJWTProviderUsed.Set(jwt)
	m.featureExtAuthUsed.Set(extAuth)
	m.featureCustomCORSUsed.Set(cors)
	m.featureCustomHeadersUsed.Set(headers)
	m.featureNoAuthUsed.Set(noauth)

	m.APIRuleObjectModifiedErrorsCount.Collect(ch)
	m.JWTHandlerIstioUsed.Collect(ch)
	m.featureJWTProviderUsed.Collect(ch)
	m.featureExtAuthUsed.Collect(ch)
	m.featureCustomCORSUsed.Collect(ch)
	m.featureCustomHeadersUsed.Collect(ch)
	m.featureNoAuthUsed.Collect(ch)
}

func NewAPIRuleCollector(reader client.Reader) *APIRuleCollector {
	collector := &APIRuleCollector{
		client: reader,
		APIRuleObjectModifiedErrorsCount: prometheus.NewCounter(prometheus.CounterOpts{
			Name:      "api_rule_object_modified_errors_total",
			Namespace: "api_gateway",
			Help:      "The total number of errors that occurred while modifying the APIRule object",
		}),
		JWTHandlerIstioUsed: prometheus.NewGauge(prometheus.GaugeOpts{
			Name:      "jwt_handler_istio_used",
			Namespace: "api_gateway",
			Help:      "Whether the Istio JWT handler is currently configured (1) or not (0)",
		}),
		featureJWTProviderUsed: prometheus.NewGauge(prometheus.GaugeOpts{
			Name:      "num_feature_jwt_provider_used",
			Namespace: "api_rule",
			Help:      "number of APIRules that have a JWT provider configured",
		}),
		featureExtAuthUsed: prometheus.NewGauge(prometheus.GaugeOpts{
			Name:      "num_feature_ext_auth_used",
			Namespace: "api_rule",
			Help:      "number of APIRules that have an external authorization configured",
		}),
		featureCustomCORSUsed: prometheus.NewGauge(prometheus.GaugeOpts{
			Name:      "num_feature_custom_cors_used",
			Namespace: "api_rule",
			Help:      "number of APIRules that have a custom CORS policy configured",
		}),
		featureCustomHeadersUsed: prometheus.NewGauge(prometheus.GaugeOpts{
			Name:      "num_feature_custom_headers_used",
			Namespace: "api_rule",
			Help:      "number of APIRules that use custom request headers in their rules",
		}),
		featureNoAuthUsed: prometheus.NewGauge(prometheus.GaugeOpts{
			Name:      "num_feature_no_auth_used",
			Namespace: "api_rule",
			Help:      "number of APIRules that have NoAuth configured",
		}),
	}
	ctrlmetrics.Registry.MustRegister(collector)
	return collector
}
