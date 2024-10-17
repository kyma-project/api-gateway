package controllers

import (
	"golang.org/x/time/rate"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"time"
)

const (
	RateLimiterBurst            = 200
	RateLimiterFrequency        = 30
	RateLimiterFailureBaseDelay = 1 * time.Second
	RateLimiterFailureMaxDelay  = 1000 * time.Second
)

type RateLimiterConfig struct {
	Burst            int
	Frequency        int
	FailureBaseDelay time.Duration
	FailureMaxDelay  time.Duration
}

// NewRateLimiter returns a rate limiter for a client-go workqueue.
// It has both an overall (token bucket) and per-item (exponential) rate limiting.
func NewRateLimiter(c RateLimiterConfig) workqueue.TypedRateLimiter[ctrl.Request] {
	return workqueue.NewTypedMaxOfRateLimiter(
		workqueue.NewTypedItemExponentialFailureRateLimiter[ctrl.Request](c.FailureBaseDelay, c.FailureMaxDelay),
		&workqueue.TypedBucketRateLimiter[ctrl.Request]{Limiter: rate.NewLimiter(rate.Limit(c.Frequency), c.Burst)},
	)
}
