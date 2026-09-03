package metrics

import (
	"testing"

	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	ctrlmetrics "sigs.k8s.io/controller-runtime/pkg/metrics"
)

func TestAPIRuleCollector_Collect(t *testing.T) {
	tc := []struct {
		name                                string
		apiRules                            []runtime.Object
		expectNumOfFeatureJWTProviderUsed   float64
		expectNumOfFeatureExtAuthUsed       float64
		expectNumOfFeatureCustomCORSUsed    float64
		expectNumOfFeatureCustomHeadersUsed float64
		expectNumOfFeatureNoAuthUsed        float64
	}{
		{
			name: "single APIRule with each feature configured increments every gauge",
			apiRules: []runtime.Object{
				&gatewayv2alpha1.APIRule{
					ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "bar"},
					Spec: gatewayv2alpha1.APIRuleSpec{
						CorsPolicy: &gatewayv2alpha1.CorsPolicy{AllowOrigins: gatewayv2alpha1.StringMatch{{"exact": "https://example.com"}}},
						Rules: []gatewayv2alpha1.Rule{
							{
								Jwt:     &gatewayv2alpha1.JwtConfig{},
								ExtAuth: &gatewayv2alpha1.ExtAuth{ExternalAuthorizers: []string{"authorizer"}},
								Request: &gatewayv2alpha1.Request{Headers: map[string]string{"X-Custom-Header": "value"}},
							},
							{
								NoAuth: new(true),
							},
							{
								NoAuth: new(true),
							},
						},
					},
				},
			},
			expectNumOfFeatureJWTProviderUsed:   1,
			expectNumOfFeatureExtAuthUsed:       1,
			expectNumOfFeatureCustomCORSUsed:    1,
			expectNumOfFeatureCustomHeadersUsed: 1,
			expectNumOfFeatureNoAuthUsed:        2,
		},
		{
			name: "multiple APIRules with different configurations set expected gauges",
			apiRules: []runtime.Object{
				&gatewayv2alpha1.APIRule{
					ObjectMeta: metav1.ObjectMeta{Name: "jwt-and-cors", Namespace: "bar"},
					Spec: gatewayv2alpha1.APIRuleSpec{
						CorsPolicy: &gatewayv2alpha1.CorsPolicy{AllowOrigins: gatewayv2alpha1.StringMatch{{"exact": "https://example.com"}}},
						Rules: []gatewayv2alpha1.Rule{
							{Jwt: &gatewayv2alpha1.JwtConfig{}},
							{Jwt: &gatewayv2alpha1.JwtConfig{}},
						},
					},
				},
				&gatewayv2alpha1.APIRule{
					ObjectMeta: metav1.ObjectMeta{Name: "extauth-and-headers", Namespace: "bar"},
					Spec: gatewayv2alpha1.APIRuleSpec{
						Rules: []gatewayv2alpha1.Rule{
							{ExtAuth: &gatewayv2alpha1.ExtAuth{ExternalAuthorizers: []string{"authorizer"}}},
							{Request: &gatewayv2alpha1.Request{Headers: map[string]string{"X-Custom-Header": "value"}}},
						},
					},
				},
				&gatewayv2alpha1.APIRule{
					ObjectMeta: metav1.ObjectMeta{Name: "cors-only", Namespace: "bar"},
					Spec: gatewayv2alpha1.APIRuleSpec{
						CorsPolicy: &gatewayv2alpha1.CorsPolicy{AllowOrigins: gatewayv2alpha1.StringMatch{{"exact": "https://other.com"}}},
						Rules:      []gatewayv2alpha1.Rule{{}},
					},
				},
			},
			expectNumOfFeatureJWTProviderUsed:   2,
			expectNumOfFeatureExtAuthUsed:       1,
			expectNumOfFeatureCustomCORSUsed:    2,
			expectNumOfFeatureCustomHeadersUsed: 1,
		},
		{
			name: "no APIRules leaves every gauge at zero",
		},
	}

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			scheme := runtime.NewScheme()
			assert.NoError(t, gatewayv2alpha1.AddToScheme(scheme))
			c := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(tt.apiRules...).Build()

			cl := NewAPIRuleCollector(c)
			t.Cleanup(func() { ctrlmetrics.Registry.Unregister(cl) })

			// Drain the channel so Collect can emit all metrics without blocking.
			ch := make(chan prometheus.Metric, 16)
			cl.Collect(ch)
			close(ch)
			for range ch {
			}

			assert.Equal(t, tt.expectNumOfFeatureJWTProviderUsed, testutil.ToFloat64(cl.featureJWTProviderUsed))
			assert.Equal(t, tt.expectNumOfFeatureExtAuthUsed, testutil.ToFloat64(cl.featureExtAuthUsed))
			assert.Equal(t, tt.expectNumOfFeatureCustomCORSUsed, testutil.ToFloat64(cl.featureCustomCORSUsed))
			assert.Equal(t, tt.expectNumOfFeatureCustomHeadersUsed, testutil.ToFloat64(cl.featureCustomHeadersUsed))
			assert.Equal(t, tt.expectNumOfFeatureNoAuthUsed, testutil.ToFloat64(cl.featureNoAuthUsed))
		})
	}
}
