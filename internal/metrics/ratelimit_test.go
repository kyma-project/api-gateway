package metrics

import (
	"testing"

	ratelimitv1alpha1 "github.com/kyma-project/api-gateway/apis/gateway/ratelimit/v1alpha1"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	ctrlmetrics "sigs.k8s.io/controller-runtime/pkg/metrics"
)

func TestRateLimitCollector_Collect(t *testing.T) {
	tc := []struct {
		name                             string
		rateLimits                       []runtime.Object
		expectNumOfGatewayRateLimits     float64
		expectNumOfWorkloadRateLimits    float64
		expectNumOfCustomRateLimitBucket float64
	}{
		{
			name: "single gateway RateLimit with custom buckets increments every gauge",
			rateLimits: []runtime.Object{
				&ratelimitv1alpha1.RateLimit{
					ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "bar"},
					Spec: ratelimitv1alpha1.RateLimitSpec{
						SelectorLabels: map[string]string{"app": "istio-ingressgateway"},
						Local: ratelimitv1alpha1.LocalConfig{
							Buckets: []ratelimitv1alpha1.BucketConfig{{}},
						},
					},
				},
			},
			expectNumOfGatewayRateLimits:     1,
			expectNumOfWorkloadRateLimits:    0,
			expectNumOfCustomRateLimitBucket: 1,
		},
		{
			name: "multiple RateLimits with different configurations set expected gauges",
			rateLimits: []runtime.Object{
				&ratelimitv1alpha1.RateLimit{
					ObjectMeta: metav1.ObjectMeta{Name: "gateway-with-buckets", Namespace: "bar"},
					Spec: ratelimitv1alpha1.RateLimitSpec{
						SelectorLabels: map[string]string{"app": "istio-ingressgateway"},
						Local: ratelimitv1alpha1.LocalConfig{
							Buckets: []ratelimitv1alpha1.BucketConfig{{}, {}},
						},
					},
				},
				&ratelimitv1alpha1.RateLimit{
					ObjectMeta: metav1.ObjectMeta{Name: "workload-with-buckets", Namespace: "bar"},
					Spec: ratelimitv1alpha1.RateLimitSpec{
						SelectorLabels: map[string]string{"app": "my-workload"},
						Local: ratelimitv1alpha1.LocalConfig{
							Buckets: []ratelimitv1alpha1.BucketConfig{{}},
						},
					},
				},
				&ratelimitv1alpha1.RateLimit{
					ObjectMeta: metav1.ObjectMeta{Name: "workload-no-buckets", Namespace: "bar"},
					Spec: ratelimitv1alpha1.RateLimitSpec{
						SelectorLabels: map[string]string{"app": "another-workload"},
					},
				},
			},
			expectNumOfGatewayRateLimits:     1,
			expectNumOfWorkloadRateLimits:    2,
			expectNumOfCustomRateLimitBucket: 2,
		},
		{
			name: "no RateLimits leaves every gauge at zero",
		},
	}

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			scheme := runtime.NewScheme()
			assert.NoError(t, ratelimitv1alpha1.AddToScheme(scheme))
			c := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(tt.rateLimits...).Build()

			cl := NewRateLimitCollector(c)
			t.Cleanup(func() { ctrlmetrics.Registry.Unregister(cl) })

			// Drain the channel so Collect can emit all metrics without blocking.
			ch := make(chan prometheus.Metric, 16)
			cl.Collect(ch)
			close(ch)
			for range ch {
			}

			assert.Equal(t, tt.expectNumOfGatewayRateLimits, testutil.ToFloat64(cl.NumOfGatewayRateLimits))
			assert.Equal(t, tt.expectNumOfWorkloadRateLimits, testutil.ToFloat64(cl.NumOfWorkloadRateLimits))
			assert.Equal(t, tt.expectNumOfCustomRateLimitBucket, testutil.ToFloat64(cl.NumOfCustomRateLimitBuckets))
		})
	}
}
