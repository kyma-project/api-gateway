package hooks

import (
	"bytes"
	"context"
	"fmt"
	"k8s.io/client-go/dynamic"
	"os"

	"github.com/avast/retry-go/v4"
	"github.com/cucumber/godog"
	"github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	k8sclient "github.com/kyma-project/api-gateway/tests/integration/pkg/client"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/testcontext"
	"github.com/pkg/errors"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

const manifestsPath = "testsuites/gateway/manifests/"
const templateFileName string = "manifests/apigateway_cr_template.yaml"
const namePrefix string = "api-gateway"

var ApplyApiGatewayCr = func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
	k8sClient, err := testcontext.GetK8sClientFromContext(ctx)
	if err != nil {
		return ctx, err
	}

	apiGateway, err := createApiGatewayCRFromTemplate(namePrefix)

	err = retry.Do(func() error {
		err := k8sClient.Create(ctx, &apiGateway)
		if err != nil {
			return err
		}
		ctx = testcontext.AddApiGatewayCRIntoContext(ctx, &apiGateway)
		return nil
	}, testcontext.GetRetryOpts()...)

	return ctx, nil
}

var ApplyAndVerifyApiGatewayCr = func(d dynamic.Interface) error {
	k8sClient := k8sclient.GetK8sClient()

	apiGateway, err := createApiGatewayCRFromTemplate(namePrefix)
	if err != nil {
		return err
	}

	err = retry.Do(func() error {
		err := k8sClient.Create(context.Background(), &apiGateway)
		if err != nil {
			return err
		}
		return nil
	}, testcontext.GetRetryOpts()...)

	if err != nil {
		return err
	}

	// TODO add verify apigateway cr
	return nil
}

var ApiGatewayCrTearDown = func(d dynamic.Interface) error {
	// TODO
	_ = k8sclient.GetK8sClient()

	//if apiGateways, ok := testcontext.GetApiGatewayCRsFromContext(ctx); ok {
	//	// We can ignore a failed removal of the ApiGateway CR, because we need to run force remove in any case to make sure no resource is left before the next scenario
	//	for _, apiGateway := range apiGateways {
	//		_ = retry.Do(func() error {
	//			err := removeObjectFromCluster(ctx, apiGateway)
	//			if err != nil {
	//				//t.Logf("Failed to delete ApiGateway CR %s", apiGateway.GetName())
	//				return err
	//			}
	//			//t.Logf("Deleted ApiGateway CR %s", apiGateway.GetName())
	//			return nil
	//		}, testcontext.GetRetryOpts()...)
	//		err := forceApiGatewayCrRemoval(ctx, apiGateway)
	//		if err != nil {
	//			return  err
	//		}
	//	}
	//}
	return nil
}

func forceApiGatewayCrRemoval(ctx context.Context, apiGateway *v1alpha1.APIGateway) error {
	c, err := testcontext.GetK8sClientFromContext(ctx)
	if err != nil {
		return err
	}

	t, err := testcontext.GetTestingFromContext(ctx)
	if err != nil {
		return err
	}

	return retry.Do(func() error {
		err = c.Get(ctx, client.ObjectKey{Namespace: apiGateway.GetNamespace(), Name: apiGateway.GetName()}, apiGateway)

		if k8serrors.IsNotFound(err) {
			return nil
		}

		if err != nil {
			return err
		}

		if apiGateway.Status.State == v1alpha1.Error {
			t.Logf("ApiGateway CR in error state (%s), force removal", apiGateway.Status.Description)
			apiGateway.Finalizers = nil
			err = c.Update(ctx, apiGateway)
			if err != nil {
				return err
			}

			return nil
		}

		return errors.New(fmt.Sprintf("apiGateway CR in status %s found (%s), skipping force removal", apiGateway.Status.State, apiGateway.Status.Description))
	}, testcontext.GetRetryOpts()...)
}

func removeObjectFromCluster(ctx context.Context, object client.Object) error {
	k8sClient, err := testcontext.GetK8sClientFromContext(ctx)
	if err != nil {
		return err
	}

	deletePolicy := metav1.DeletePropagationForeground
	err = k8sClient.Delete(context.TODO(), object, &client.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})
	if err != nil && !k8serrors.IsNotFound(err) {
		return err
	}

	return nil
}

func createApiGatewayCRFromTemplate(name string) (v1alpha1.APIGateway, error) {
	apiGatewayCRYaml, err := os.ReadFile(templateFileName)
	if err != nil {
		return v1alpha1.APIGateway{}, err
	}

	resource := bytes.NewBuffer(apiGatewayCRYaml)

	var apiGateway v1alpha1.APIGateway
	err = yaml.Unmarshal(resource.Bytes(), &apiGateway)
	if err != nil {
		return v1alpha1.APIGateway{}, err
	}

	apiGateway.Name = name
	return apiGateway, nil
}
