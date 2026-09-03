package metrics

import (
	"context"

	ratelimitv1alpha1 "github.com/kyma-project/api-gateway/apis/gateway/ratelimit/v1alpha1"
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlmetrics "sigs.k8s.io/controller-runtime/pkg/metrics"
)

type RateLimitCollector struct {
	client                      client.Reader
	NumOfGatewayRateLimits      prometheus.Gauge
	NumOfWorkloadRateLimits     prometheus.Gauge
	NumOfCustomRateLimitBuckets prometheus.Gauge
}

func (c *RateLimitCollector) Describe(ch chan<- *prometheus.Desc) {
	c.NumOfGatewayRateLimits.Describe(ch)
	c.NumOfWorkloadRateLimits.Describe(ch)
	c.NumOfCustomRateLimitBuckets.Describe(ch)
}

func (c *RateLimitCollector) Collect(ch chan<- prometheus.Metric) {
	rl := ratelimitv1alpha1.RateLimitList{}
	if err := c.client.List(context.Background(), &rl); err != nil {
		return
	}
	workload, gateway, customBuckets := 0.0, 0.0, 0.0
	for _, r := range rl.Items {
		if r.Spec.SelectorLabels["app"] == "istio-ingressgateway" {
			gateway += 1.0
		} else {
			workload += 1.0
		}
		if len(r.Spec.Local.Buckets) > 0 {
			customBuckets += 1.0
		}
	}

	c.NumOfCustomRateLimitBuckets.Set(customBuckets)
	c.NumOfWorkloadRateLimits.Set(workload)
	c.NumOfGatewayRateLimits.Set(gateway)

	c.NumOfGatewayRateLimits.Collect(ch)
	c.NumOfWorkloadRateLimits.Collect(ch)
	c.NumOfCustomRateLimitBuckets.Collect(ch)
}

func NewRateLimitCollector(client client.Reader) *RateLimitCollector {
	c := &RateLimitCollector{
		client: client,
		NumOfGatewayRateLimits: prometheus.NewGauge(prometheus.GaugeOpts{
			Name:      "num_of_gateway_rate_limits",
			Namespace: "ratelimit",
			Help:      "Number of RateLimit CRs targeting gateway",
		}),
		NumOfWorkloadRateLimits: prometheus.NewGauge(prometheus.GaugeOpts{
			Name:      "num_of_workload_rate_limits",
			Namespace: "ratelimit",
			Help:      "Number of RateLimit CRs targeting workloads",
		}),
		NumOfCustomRateLimitBuckets: prometheus.NewGauge(prometheus.GaugeOpts{
			Name:      "num_of_custom_rate_limit_buckets",
			Namespace: "ratelimit",
			Help:      "Number of RateLimit CRs with custom buckets configured",
		}),
	}
	ctrlmetrics.Registry.MustRegister(c)
	return c
}
