//go:build !dev_features

package ratelimit

import (
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func SetupRateLimit(_ manager.Manager, _ *runtime.Scheme) error {
	return nil
}
