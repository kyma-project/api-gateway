//go:build !ratelimit

package ratelimit

import "sigs.k8s.io/controller-runtime/pkg/manager"

func Setup(_ manager.Manager) error {
	return nil
}
