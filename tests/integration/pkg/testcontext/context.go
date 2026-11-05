package testcontext

import (
	"context"
	"testing"

	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// k8sClientCtxKey is the key used to store the k8sClient used during tests in the context.Context.
type k8sClientCtxKey struct{}

func GetK8sClientFromContext(ctx context.Context) (client.Client, error) {
	v, ok := ctx.Value(k8sClientCtxKey{}).(client.Client)
	if !ok {
		return v, errors.New("k8sClient not found in context")
	}
	return v, nil
}

func SetK8sClientInContext(ctx context.Context, k8sClient client.Client) context.Context {
	return context.WithValue(ctx, k8sClientCtxKey{}, k8sClient)
}

type testingContextKey struct{}

func SetTestingInContext(ctx context.Context, testing *testing.T) context.Context {
	return context.WithValue(ctx, testingContextKey{}, testing)
}
