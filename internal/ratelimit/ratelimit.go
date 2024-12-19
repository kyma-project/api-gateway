package ratelimit

import (
	"github.com/kyma-project/api-gateway/internal/builders/envoyfilter"
	"google.golang.org/protobuf/types/known/structpb"
	"istio.io/api/networking/v1alpha3"
	networkingv1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	"slices"
	"time"
)

const (
	LocalRateLimitFilterUrl  = "type.googleapis.com/envoy.extensions.filters.http.local_ratelimit.v3.LocalRateLimit"
	TypedStruct              = "type.googleapis.com/udpa.type.v1.TypedStruct"
	LocalRateLimitFilterName = "envoy.filters.http.local_ratelimit"
)

// RateLimit contains configuration for Rate Limiting service, exposing functions to manage Envoy's settings.
type RateLimit struct {
	limitType             string
	limityTypeUrl         string
	enforce               bool
	enableResponseHeaders bool
	actions               []Action
	descriptors           []Descriptor
	defaultBucket         Bucket
}

// Action implements Envoy's Action API fields needed for Rate Limit configuration.
// See: https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/route/v3/route_components.proto#config-route-v3-ratelimit-action
type Action struct {
	RequestHeaders RequestHeader
}

func (a Action) Value() *structpb.Value {
	return structpb.NewStructValue(&structpb.Struct{Fields: map[string]*structpb.Value{
		"request_headers": a.RequestHeaders.Value(),
	}})
}

// RequestHeader implements Envoy's RequestHeader's API fields.
// See: https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/route/v3/route_components.proto#envoy-v3-api-msg-config-route-v3-ratelimit-action-requestheaders
type RequestHeader struct {
	Name          string
	DescriptorKey string
}

func (rh RequestHeader) Value() *structpb.Value {
	return structpb.NewStructValue(&structpb.Struct{Fields: map[string]*structpb.Value{
		"header_name":    structpb.NewStringValue(rh.Name),
		"descriptor_key": structpb.NewStringValue(rh.DescriptorKey),
	}})
}

// Descriptor describes each entry for rate limit.
type Descriptor struct {
	Entries DescriptorEntries
	Bucket  Bucket
}

// DescriptorEntries wraps a list of DescriptorEntry and implements PBValue interface
type DescriptorEntries []DescriptorEntry

func (de DescriptorEntries) Value() *structpb.Value {
	var values []*structpb.Value
	for _, entry := range de {
		values = append(values, entry.Value())
	}
	return structpb.NewListValue(&structpb.ListValue{Values: values})
}

func (d Descriptor) Value() *structpb.Value {
	return structpb.NewStructValue(&structpb.Struct{
		Fields: map[string]*structpb.Value{
			"entries":      d.Entries.Value(),
			"token_bucket": d.Bucket.Value(),
		},
	})
}

// DescriptorEntry contains routes to which local rate limits apply in scope of each Descriptor
type DescriptorEntry struct {
	Key string
	Val string
}

func (e DescriptorEntry) Value() *structpb.Value {
	return structpb.NewStructValue(&structpb.Struct{
		Fields: map[string]*structpb.Value{
			"key":   structpb.NewStringValue(e.Key),
			"value": structpb.NewStringValue(e.Val),
		},
	})
}

// Bucket implements token_bucket fields from Envoy API.
// See: https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/http/local_ratelimit/v3/local_rate_limit.proto#envoy-v3-api-msg-extensions-filters-http-local-ratelimit-v3-localratelimit
type Bucket struct {
	MaxTokens     int64
	TokensPerFill int64
	FillInterval  time.Duration
}

func (b Bucket) Value() *structpb.Value {
	return structpb.NewStructValue(&structpb.Struct{
		Fields: map[string]*structpb.Value{
			"fill_interval":   structpb.NewStringValue(b.FillInterval.String()),
			"max_tokens":      structpb.NewNumberValue(float64(b.MaxTokens)),
			"tokens_per_fill": structpb.NewNumberValue(float64(b.TokensPerFill))},
	})
}

// it returns generic http filter needed for applying local rate limit
func localHttpFilterPatch() *envoyfilter.ConfigPatch {
	return &v1alpha3.EnvoyFilter_EnvoyConfigObjectPatch{
		ApplyTo: v1alpha3.EnvoyFilter_HTTP_FILTER,
		Match: &v1alpha3.EnvoyFilter_EnvoyConfigObjectMatch{
			Context: v1alpha3.EnvoyFilter_SIDECAR_INBOUND,
			ObjectTypes: &v1alpha3.EnvoyFilter_EnvoyConfigObjectMatch_Listener{
				Listener: &v1alpha3.EnvoyFilter_ListenerMatch{
					FilterChain: &v1alpha3.EnvoyFilter_ListenerMatch_FilterChainMatch{
						Filter: &v1alpha3.EnvoyFilter_ListenerMatch_FilterMatch{
							Name: "envoy.filters.network.http_connection_manager",
						},
					},
				},
			},
		},
		Patch: &v1alpha3.EnvoyFilter_Patch{
			Operation: v1alpha3.EnvoyFilter_Patch_INSERT_BEFORE,
			Value: &structpb.Struct{Fields: map[string]*structpb.Value{
				"name": structpb.NewStringValue(LocalRateLimitFilterName),
				"typed_config": structpb.NewStructValue(&structpb.Struct{Fields: map[string]*structpb.Value{
					"@type":    structpb.NewStringValue(TypedStruct),
					"type_url": structpb.NewStringValue(LocalRateLimitFilterUrl),
					"value": structpb.NewStructValue(&structpb.Struct{Fields: map[string]*structpb.Value{
						"stat_prefix": structpb.NewStringValue("http_local_rate_limiter"),
					}}),
				}}),
			}}},
	}
}

// RateLimitConfigPatch generates Istio-compatible ConfigPatch containing local rate limit configuration
func (rl *RateLimit) RateLimitConfigPatch() *envoyfilter.ConfigPatch {
	return &envoyfilter.ConfigPatch{
		ApplyTo: v1alpha3.EnvoyFilter_HTTP_ROUTE,
		Match: &v1alpha3.EnvoyFilter_EnvoyConfigObjectMatch{
			Context: v1alpha3.EnvoyFilter_SIDECAR_INBOUND,
		},
		Patch: &v1alpha3.EnvoyFilter_Patch{
			Operation: v1alpha3.EnvoyFilter_Patch_MERGE,
			Value: &structpb.Struct{Fields: map[string]*structpb.Value{
				"route": structpb.NewStructValue(&structpb.Struct{Fields: map[string]*structpb.Value{
					"rate_limits": structpb.NewListValue(&structpb.ListValue{Values: []*structpb.Value{
						structpb.NewStructValue(&structpb.Struct{Fields: map[string]*structpb.Value{
							"actions": func() *structpb.Value {
								var actVal []*structpb.Value
								for _, a := range rl.actions {
									actVal = append(actVal, a.Value())
								}
								return structpb.NewListValue(&structpb.ListValue{Values: actVal})
							}(),
						}}),
					}}),
				}}),
				"typed_per_filter_config": structpb.NewStructValue(&structpb.Struct{Fields: map[string]*structpb.Value{
					LocalRateLimitFilterName: structpb.NewStructValue(&structpb.Struct{Fields: map[string]*structpb.Value{
						"@type":    structpb.NewStringValue(rl.limitType),
						"type_url": structpb.NewStringValue(rl.limityTypeUrl),
						"value": structpb.NewStructValue(&structpb.Struct{Fields: map[string]*structpb.Value{
							"stat_prefix": structpb.NewStringValue("rate_limit"),
							"enable_x_ratelimit_headers": func() *structpb.Value {
								if rl.enableResponseHeaders {
									return structpb.NewStringValue("DRAFT_VERSION_03")
								}
								return structpb.NewStringValue("OFF")
							}(),
							"filter_enabled": structpb.NewStructValue(&structpb.Struct{Fields: map[string]*structpb.Value{
								"runtime_key": structpb.NewStringValue("local_rate_limit_enabled"),
								"default_value": structpb.NewStructValue(&structpb.Struct{Fields: map[string]*structpb.Value{
									"numerator":   structpb.NewNumberValue(float64(100)),
									"denominator": structpb.NewStringValue("HUNDRED"),
								}}),
							}}),
							"filter_enforced": structpb.NewStructValue(&structpb.Struct{Fields: map[string]*structpb.Value{
								"runtime_key": structpb.NewStringValue("local_rate_limit_enforced"),
								"default_value": structpb.NewStructValue(&structpb.Struct{Fields: map[string]*structpb.Value{
									"numerator": func() *structpb.Value {
										if rl.enforce {
											return structpb.NewNumberValue(float64(100))
										}
										return structpb.NewNumberValue(float64(0))
									}(),
									"denominator": structpb.NewStringValue("HUNDRED"),
								}}),
							}}),
							"always_consume_default_token_bucket": structpb.NewBoolValue(false),
							"token_bucket":                        rl.defaultBucket.Value(),
							"descriptors": func() *structpb.Value {
								var vals []*structpb.Value
								for _, a := range rl.descriptors {
									vals = append(vals, a.Value())
								}
								return structpb.NewListValue(&structpb.ListValue{Values: vals})
							}(),
						}}),
					}}),
				}}),
			}},
		},
	}
}

func hasHeader(header RequestHeader, actions []Action) bool {
	for _, a := range actions {
		if a.RequestHeaders.Name == header.Name {
			return true
		}
	}
	return false
}

// Enforce sets if the RateLimit configuration shouild be enforced or not.
func (rl *RateLimit) Enforce(enforce bool) *RateLimit {
	rl.enforce = enforce
	return rl
}

// EnableResponseHeaders enables sending `X-Rate-Limit` headers in the response.
// See: https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/http/ratelimit/v3/rate_limit.proto#rate-limit-proto
func (rl *RateLimit) EnableResponseHeaders(enable bool) *RateLimit {
	rl.enableResponseHeaders = enable
	return rl
}

// For adds rate limit for specific descriptor. From this descriptor a RequestHeader is created, that is then added
// to the Action list.
func (rl *RateLimit) For(descriptor Descriptor) *RateLimit {
	for _, d := range descriptor.Entries {
		name := d.Key
		// These particular values are pseudo headers defined in Envoy API.
		// They need to be prefixed with ':'
		if slices.Contains([]string{"path", "method", "scheme"}, name) {
			name = ":" + name
		}
		rh := RequestHeader{Name: name, DescriptorKey: d.Key}
		if !hasHeader(rh, rl.actions) {
			rl.actions = append(rl.actions, Action{RequestHeaders: rh})
		}
	}
	rl.descriptors = append(rl.descriptors, descriptor)
	return rl
}

// WithDefaultBucket adds configuration for the token bucked used by default.
func (rl *RateLimit) WithDefaultBucket(bucket Bucket) *RateLimit {
	rl.defaultBucket = bucket
	return rl
}

// SetConfigPatches parses RateLimit configuration, then applies the parsed ConfigPatches directly into the
// networkingv1alpha3.EnvoyFilter struct, replacing previous configuration.
func (rl *RateLimit) SetConfigPatches(filter *networkingv1alpha3.EnvoyFilter) {
	filter.Spec.ConfigPatches = []*envoyfilter.ConfigPatch{
		localHttpFilterPatch(),
		rl.RateLimitConfigPatch(),
	}
}

// NewLocalRateLimit returns RateLimit struct for configuring local rate limits
func NewLocalRateLimit() *RateLimit {
	return &RateLimit{
		limitType:     TypedStruct,
		limityTypeUrl: LocalRateLimitFilterUrl,
	}
}
