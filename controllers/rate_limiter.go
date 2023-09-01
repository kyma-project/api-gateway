package controllers

import (
	"golang.org/x/time/rate"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/ratelimiter"
	"time"
)

const (
	rateLimiterBurst     = 200
	rateLimiterFrequency = 30
	failureBaseDelay     = 1 * time.Second
	failureMaxDelay      = 1000 * time.Second
)

// NewRateLimiter returns a rate limiter for a client-go.workqueue.  It has both an overall (token bucket) and per-item (exponential) rate limiting.
func NewRateLimiter() ratelimiter.RateLimiter {
	return workqueue.NewMaxOfRateLimiter(
		workqueue.NewItemExponentialFailureRateLimiter(failureBaseDelay, failureMaxDelay),
		&workqueue.BucketRateLimiter{Limiter: rate.NewLimiter(rate.Limit(rateLimiterFrequency), rateLimiterBurst)})
}
