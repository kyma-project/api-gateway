package hooks

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"

	"github.com/avast/retry-go/v4"
	"github.com/cucumber/godog"
	"github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	k8sclient "github.com/kyma-project/api-gateway/tests/integration/pkg/client"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/testcontext"
	oryv1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
	"github.com/pkg/errors"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
)

const templateFileName string = "pkg/hooks/manifests/apigateway.yaml"
const ApiGatewayCRName string = "default"

var ApplyApiGatewayCrScenarioHook = func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
	k8sClient, err := testcontext.GetK8sClientFromContext(ctx)
	if err != nil {
		return ctx, err
	}
	apiGateway, err := createApiGatewayCRObjectFromTemplate(ApiGatewayCRName)
	if err != nil {
		return ctx, err
	}
	err = retry.Do(func() error {
		err := k8sClient.Create(ctx, &apiGateway)
		if err != nil {
			return err
		}
		ctx = testcontext.AddApiGatewayCRIntoContext(ctx, &apiGateway)
		return nil
	}, testcontext.GetRetryOpts()...)
	return ctx, err
}

var DeleteBlockingResourcesScenarioHook = func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
	return ctx, deleteBlockingResources(ctx)
}

var ApiGatewayCrTearDownScenarioHook = func(ctx context.Context, sc *godog.Scenario, _ error) (context.Context, error) {
	if apiGateways, ok := testcontext.GetApiGatewayCRsFromContext(ctx); ok {
		// We can ignore a failed removal of the ApiGateway CR, because we need to run force remove in any case to make sure no resource is left before the next scenario
		for _, apiGateway := range apiGateways {
			_ = retry.Do(func() error {
				err := removeObjectFromCluster(ctx, apiGateway)
				if err != nil {
					return fmt.Errorf("Failed to delete ApiGateway CR %s", apiGateway.GetName())
				}
				return nil
			}, testcontext.GetRetryOpts()...)
			err := forceApiGatewayCrRemoval(ctx, apiGateway)
			if err != nil {
				return ctx, err
			}
		}
	}
	return ctx, nil
}

var ApplyAndVerifyApiGatewayCrSuiteHook = func() error {
	log.Printf("Creating APIGateway CR %s", ApiGatewayCRName)
	k8sClient := k8sclient.GetK8sClient()

	apiGateway, err := createApiGatewayCRObjectFromTemplate(ApiGatewayCRName)
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

	err = retry.Do(func() error {
		err := k8sClient.Get(context.Background(), client.ObjectKey{
			Namespace: apiGateway.GetNamespace(),
			Name:      apiGateway.GetName(),
		}, &apiGateway)

		if err != nil {
			return err
		}

		if apiGateway.Status.State != "Ready" {
			return fmt.Errorf("apigateway cr should be in Ready state, but is in %s", apiGateway.Status.State)
		}

		return nil
	}, testcontext.GetRetryOpts()...)

	if err != nil {
		return err
	}

	log.Printf("APIGateway CR %s in state %s", ApiGatewayCRName, apiGateway.Status.State)

	return nil
}

var DeleteBlockingResourcesSuiteHook = func() error {
	return deleteBlockingResources(context.Background())
}

var ApiGatewayCrTearDownSuiteHook = func() error {
	k8sClient := k8sclient.GetK8sClient()

	apiGateway, err := createApiGatewayCRObjectFromTemplate(ApiGatewayCRName)
	if err != nil {
		return err
	}

	err = retry.Do(func() error {
		err := k8sClient.Delete(context.Background(), &apiGateway)
		if err != nil {
			return err
		}
		return nil
	}, testcontext.GetRetryOpts()...)
	if err != nil {
		return err
	}

	err = retry.Do(func() error {
		err := k8sClient.Get(context.Background(), client.ObjectKey{
			Namespace: apiGateway.GetNamespace(),
			Name:      apiGateway.GetName(),
		}, &apiGateway)

		if err == nil {
			return fmt.Errorf("ApiGatewayCrTearDownSuiteHook did not delete APIGateway CR, state: %s description: %s", apiGateway.Status.State, apiGateway.Status.Description)
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func forceApiGatewayCrRemoval(ctx context.Context, apiGateway *v1alpha1.APIGateway) error {
	c, err := testcontext.GetK8sClientFromContext(ctx)
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

	err = k8sClient.Delete(context.Background(), object, &client.DeleteOptions{})
	if err != nil && !k8serrors.IsNotFound(err) {
		return err
	}

	return nil
}

func createApiGatewayCRObjectFromTemplate(name string) (v1alpha1.APIGateway, error) {
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

func deleteBlockingResources(ctx context.Context) error {
	k8sClient, err := testcontext.GetK8sClientFromContext(ctx)
	if err != nil {
		k8sClient = k8sclient.GetK8sClient()
	}

	apiRuleList := v1beta1.APIRuleList{}
	err = k8sClient.List(ctx, &apiRuleList)
	if err != nil {
		return err
	}

	for _, apiRule := range apiRuleList.Items {
		err = retry.Do(func() error {
			if apiRule.Finalizers != nil {
				apiRule.Finalizers = nil
				err = k8sClient.Update(ctx, &apiRule)
				if err != nil {
					return err
				}
			}
			err := k8sClient.Delete(ctx, &apiRule)
			if err != nil {
				return fmt.Errorf("failed to delete APIRule %s", apiRule.GetName())
			}
			return nil
		}, testcontext.GetRetryOpts()...)
		if err != nil {
			return err
		}
	}

	vsList := networkingv1beta1.VirtualServiceList{}
	err = k8sClient.List(ctx, &vsList)
	if err != nil {
		return err
	}

	for _, vs := range vsList.Items {
		err = retry.Do(func() error {
			err := k8sClient.Delete(ctx, vs)
			if err != nil {
				return fmt.Errorf("failed to delete VirtualService %s", vs.GetName())
			}
			return nil
		}, testcontext.GetRetryOpts()...)
		if err != nil {
			return err
		}
	}

	oryRuleList := oryv1alpha1.RuleList{}
	err = k8sClient.List(ctx, &oryRuleList)
	if err == nil {
		for _, oryRule := range oryRuleList.Items {
			err = retry.Do(func() error {
				err := k8sClient.Delete(ctx, &oryRule)
				if err != nil {
					return fmt.Errorf("failed to delete ORY Oathkeeper Rule %s", oryRule.GetName())
				}
				return nil
			}, testcontext.GetRetryOpts()...)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
