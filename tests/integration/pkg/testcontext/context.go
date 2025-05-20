package testcontext

import (
	"context"
	"testing"

	"github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// apiGatewayCrCtxKey is the key used to store the ApiGatewayCR used by a scenario in the context.Context.
type apiGatewayCrCtxKey struct{}

func GetApiGatewayCRsFromContext(ctx context.Context) ([]*v1alpha1.APIGateway, bool) {
	v, ok := ctx.Value(apiGatewayCrCtxKey{}).([]*v1alpha1.APIGateway)
	return v, ok
}

func AddApiGatewayCRIntoContext(ctx context.Context, apiGateway *v1alpha1.APIGateway) context.Context {
	apiGateways, ok := GetApiGatewayCRsFromContext(ctx)
	if !ok {
		apiGateways = []*v1alpha1.APIGateway{}
	}
	apiGateways = append(apiGateways, apiGateway)
	return context.WithValue(ctx, apiGatewayCrCtxKey{}, apiGateways)
}

// createdTestObjectsCtxKey is the key used to store the test resources created during tests in the context.Context.
type createdTestObjectsCtxKey struct{}

func GetCreatedTestObjectsFromContext(ctx context.Context) ([]client.Object, bool) {
	v, ok := ctx.Value(createdTestObjectsCtxKey{}).([]client.Object)
	return v, ok
}

func AddCreatedTestObjectInContext(ctx context.Context, object client.Object) context.Context {
	objects, ok := GetCreatedTestObjectsFromContext(ctx)
	if !ok {
		objects = []client.Object{}
	}

	objects = append(objects, object)
	return context.WithValue(ctx, createdTestObjectsCtxKey{}, objects)
}

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

func GetTestingFromContext(ctx context.Context) (*testing.T, error) {
	v, ok := ctx.Value(testingContextKey{}).(*testing.T)
	if !ok {
		return v, errors.New("testing.T not found in context")
	}
	return v, nil

}

func SetTestingInContext(ctx context.Context, testing *testing.T) context.Context {
	return context.WithValue(ctx, testingContextKey{}, testing)
}
