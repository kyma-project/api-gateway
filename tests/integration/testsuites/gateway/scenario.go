package gateway

import (
	"context"
	"fmt"
	"github.com/cucumber/godog"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/helpers"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/manifestprocessor"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/resource"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/testcontext"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"log"
	"path"
	"time"
)

const manifestsPath = "testsuites/gateway/manifests/"

type scenario struct {
	testID          string
	namespace       string
	k8sClient       dynamic.Interface
	httpClient      *helpers.RetryableHttpClient
	resourceManager *resource.Manager
	config          testcontext.Config
}

func initScenario(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario, err := createScenario(ts)

	if err != nil {
		log.Fatalf("could not initialize custom domain endpoint err=%s", err)
	}

	ctx.Step(`^there is an APIGateway operator in "([^"]*)" state$`, scenario.thereIsAnAPIGatewayOperator)
	ctx.Step(`^there is a "([^"]*)" gateway in "([^"]*)" namespace$`, scenario.thereIsAGateway)
	ctx.Step(`^there is a "([^"]*)" secret in "([^"]*)" namespace$`, scenario.thereIsACertificate)
	ctx.Step(`^there is an "([^"]*)" APIRule$`, scenario.thereIsAnAPIRule)
	ctx.Step(`^disabling kyma gateway will result in "([^"]*)" due to existing APIRule$`, scenario.gatewayErrorWhenKymaGatewayDisabled)
}

func createScenario(t *testsuite) (*scenario, error) {
	ns := t.namespace
	testID := helpers.GenerateRandomTestId()
	customDomainManifestDirectory := path.Dir(manifestsPath)

	// Create APIRule
	commonResources, err := manifestprocessor.ParseFromFileWithTemplate("apirule.yaml", customDomainManifestDirectory, struct {
		Namespace string
	}{
		Namespace: t.namespace,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to process common manifest files, details %s", err.Error())
	}
	_, err = t.resourceManager.CreateResources(t.k8sClient, commonResources...)

	if err != nil {
		return nil, err
	}

	return &scenario{
		testID:          testID,
		namespace:       ns,
		k8sClient:       t.k8sClient,
		resourceManager: t.resourceManager,
		config:          t.config,
	}, nil
}

func (c *scenario) thereIsAnAPIGatewayOperator(state string) error {
	err := wait.ExponentialBackoff(wait.Backoff{
		Duration: time.Second,
		Factor:   2,
		Steps:    10,
	}, func() (done bool, err error) {
		res := schema.GroupVersionResource{Group: "operator.kyma-project.io", Version: "v1alpha1", Resource: "apigateways"}
		gateway, err := c.k8sClient.Resource(res).Get(context.Background(), resource.TestGatewayOperatorName, v1.GetOptions{})
		if err != nil {
			return false, fmt.Errorf("gateway could not be found")
		}

		gatewayState, found, err := unstructured.NestedString(gateway.Object, "status", "state")
		if err != nil || !found {
			return false, err
		} else if gatewayState != state {
			return false, fmt.Errorf("gateway state %s, is not in the expected state %s", gatewayState, state)
		}
		return true, nil
	})
	if err != nil {
		return fmt.Errorf("could not get gateway status: %s", err)
	}
	return nil
}

func (c *scenario) thereIsAGateway(name string, namespace string) error {
	res := schema.GroupVersionResource{Group: "networking.istio.io", Version: "v1beta1", Resource: "gateways"}
	_, err := c.k8sClient.Resource(res).Namespace(namespace).Get(context.Background(), name, v1.GetOptions{})
	if err != nil {
		return fmt.Errorf("%s could not be found", name)
	}

	return nil
}

func (c *scenario) thereIsACertificate(name string, namespace string) error {
	res := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "secrets"}
	_, err := c.k8sClient.Resource(res).Namespace(namespace).Get(context.Background(), name, v1.GetOptions{})
	if err != nil {
		return fmt.Errorf("%s could not be found", name)
	}

	return nil
}

func (c *scenario) thereIsAnAPIRule(name string) error {
	res := schema.GroupVersionResource{Group: "gateway.kyma-project.io", Version: "v1beta1", Resource: "apirules"}
	_, err := c.k8sClient.Resource(res).Namespace(c.namespace).Get(context.Background(), name, v1.GetOptions{})
	if err != nil {
		return fmt.Errorf("%s could not be found", name)
	}

	return nil
}

func (c *scenario) gatewayErrorWhenKymaGatewayDisabled(state string) error {
	customDomainManifestDirectory := path.Dir(manifestsPath)

	kymaGatewayDisabled, err := manifestprocessor.ParseFromFileWithTemplate("kyma-gateway-disabled.yaml", customDomainManifestDirectory, struct {
		NamePrefix string
	}{
		NamePrefix: resource.TestGatewayOperatorName,
	})
	if err != nil {
		return fmt.Errorf("failed to process kyma-gateway-disabled.yaml, details %s", err.Error())
	}
	_, err = c.resourceManager.UpdateGateway(c.k8sClient, kymaGatewayDisabled...)
	if err != nil {
		return err
	}

	err = wait.ExponentialBackoff(wait.Backoff{
		Duration: time.Second,
		Factor:   2,
		Steps:    10,
	}, func() (done bool, err error) {
		res := schema.GroupVersionResource{Group: "operator.kyma-project.io", Version: "v1alpha1", Resource: "apigateways"}
		gateway, err := c.k8sClient.Resource(res).Get(context.Background(), resource.TestGatewayOperatorName, v1.GetOptions{})
		if err != nil {
			return false, fmt.Errorf("gateway could not be found")
		}

		gatewayState, found, err := unstructured.NestedString(gateway.Object, "status", "state")
		if err != nil || !found {
			return false, err
		} else if gatewayState != state {
			return false, fmt.Errorf("gateway state %s, is not in the expected state %s", gatewayState, state)
		}
		return true, nil
	})
	if err != nil {
		return fmt.Errorf("could not get gateway status: %s", err)
	}
	return nil
}
