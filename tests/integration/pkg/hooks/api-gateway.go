package hooks

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"

	"github.com/kyma-project/api-gateway/tests/integration/pkg/patch"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	"github.com/avast/retry-go/v4"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	ratelimit "github.com/kyma-project/api-gateway/apis/gateway/ratelimit/v1alpha1"
	v2 "github.com/kyma-project/api-gateway/apis/gateway/v2"
	"github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	oryv1alpha1 "github.com/kyma-project/api-gateway/internal/types/ory/oathkeeper-maester/api/v1alpha1"
	k8sclient "github.com/kyma-project/api-gateway/tests/integration/pkg/client"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/testcontext"
)

const templateFileName string = "pkg/hooks/manifests/apigateway.yaml"
const ApiGatewayCRName string = "default"

func applyAndVerifyApiGateway(scaleDownOathkeeper bool) error {
	log.Printf("Creating APIGateway CR %s", ApiGatewayCRName)
	k8sClient := k8sclient.GetK8sClient()

	if err := createAllowAPIRuleV1Signatures(context.Background(), k8sClient); err != nil {
		return err
	}

	apiGateway, err := createApiGatewayCRObjectFromTemplate(ApiGatewayCRName)
	if err != nil {
		return err
	}

	var existingGateway v1alpha1.APIGateway
	err = k8sClient.Get(context.Background(), client.ObjectKey{
		Namespace: apiGateway.GetNamespace(),
		Name:      apiGateway.GetName(),
	}, &existingGateway)
	if err == nil {
		return fmt.Errorf("apigateway with name '%s' already exists", existingGateway.Name)
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

	if scaleDownOathkeeper {
		// scale down oathkeeper if needed -> this saves up time if the test does not depend on oathkeeper, as APIGateway will become Ready faster
		err = retry.Do(func() error {
			oathkeeperDeployment := &appsv1.Deployment{}
			err := k8sClient.Get(context.Background(), client.ObjectKey{
				Namespace: "kyma-system",
				Name:      "ory-oathkeeper",
			}, oathkeeperDeployment)

			if k8serrors.IsNotFound(err) {
				return nil
			}
			if err != nil {
				return err
			}

			return k8sClient.Patch(context.Background(), oathkeeperDeployment, client.RawPatch(
				client.Merge.Type(),
				[]byte(`{"spec":{"replicas":0}}`),
			))
		}, testcontext.GetRetryOpts()...)
		if err != nil {
			return err
		}
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

var DisableV2Alpha1RequiredFieldsHook = func() error {
	err := patch.Removev2alpha1VersionRequiredFields(k8sclient.GetK8sClient())
	if err != nil {
		log.Printf("Failed to disable v2alpha1 required fields: %v", err)
	}
	return err
}

var ApplyAndVerifyApiGatewayCrSuiteHook = func() error {
	return applyAndVerifyApiGateway(false)
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

	apiRuleList := v2.APIRuleList{}
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
			if client.IgnoreNotFound(err) != nil {
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
			if client.IgnoreNotFound(err) != nil {
				return fmt.Errorf("failed to delete VirtualService %s", vs.GetName())
			}
			return nil
		}, testcontext.GetRetryOpts()...)
		if err != nil {
			return err
		}
	}

	rateLimitList := ratelimit.RateLimitList{}
	err = k8sClient.List(ctx, &rateLimitList)
	if err != nil {
		return err
	}

	for _, rateLimit := range rateLimitList.Items {
		err = retry.Do(func() error {
			err := k8sClient.Delete(ctx, &rateLimit)
			if client.IgnoreNotFound(err) != nil {
				return fmt.Errorf("failed to delete RateLimit %s", rateLimit.GetName())
			}
			return nil
		}, testcontext.GetRetryOpts()...)
		if err != nil {
			return err
		}
	}

	var oryCRD apiextensionsv1.CustomResourceDefinition
	err = k8sClient.Get(ctx, client.ObjectKey{Name: "rules.oathkeeper.ory.sh"}, &oryCRD)
	if err == nil {
		oryRuleList := oryv1alpha1.RuleList{}
		err = k8sClient.List(ctx, &oryRuleList)
		if err == nil {
			for _, oryRule := range oryRuleList.Items {
				err = retry.Do(func() error {
					err := k8sClient.Delete(ctx, &oryRule)
					if client.IgnoreNotFound(err) != nil {
						return fmt.Errorf("failed to delete ORY Oathkeeper Rule %s", oryRule.GetName())
					}
					return nil
				}, testcontext.GetRetryOpts()...)
				if err != nil {
					return err
				}
			}
		}
	} else if !k8serrors.IsNotFound(err) {
		return err
	}
	return nil
}
