package controllers

import (
	"golang.org/x/time/rate"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/ratelimiter"
	"time"
)

type RateLimiterConfig struct {
	Burst            int
	Frequency        int
	FailureBaseDelay time.Duration
	FailureMaxDelay  time.Duration
}

// NewRateLimiter returns a rate limiter for a client-go.workqueue.  It has both an overall (token bucket) and per-item (exponential) rate limiting.
func NewRateLimiter(c RateLimiterConfig) ratelimiter.RateLimiter {
	return workqueue.NewMaxOfRateLimiter(
		workqueue.NewItemExponentialFailureRateLimiter(c.FailureBaseDelay, c.FailureMaxDelay),
		&workqueue.BucketRateLimiter{Limiter: rate.NewLimiter(rate.Limit(c.Frequency), c.Burst)})
}
