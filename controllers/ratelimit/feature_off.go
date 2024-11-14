//go:build !ratelimit

package ratelimit

import (
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func Setup(_ manager.Manager, _ *runtime.Scheme) error {
	return nil
}
