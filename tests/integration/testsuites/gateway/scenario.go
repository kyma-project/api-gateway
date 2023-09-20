package gateway

import (
	"context"
	"fmt"
	"github.com/cucumber/godog"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/helpers"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/resource"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/testcontext"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"log"
)

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

	ctx.Step(`^there is a "([^"]*)" Gateway in "([^"]*)" namespace on k3d cluster$`, scenario.thereIsAKymaGatewayK3D)
	ctx.Step(`^there is a "([^"]*)" Gateway in "([^"]*)" namespace on gardener cluster$`, scenario.thereIsAKymaGatewayGardener)
}

func createScenario(t *testsuite) (*scenario, error) {
	ns := t.namespace
	testID := helpers.GenerateRandomTestId()

	return &scenario{
		testID:          testID,
		namespace:       ns,
		k8sClient:       t.k8sClient,
		resourceManager: t.resourceManager,
		config:          t.config,
	}, nil
}

func (c *scenario) thereIsAKymaGatewayK3D(name string, namespace string) error {
	res := schema.GroupVersionResource{Group: "networking.istio.io", Version: "v1alpha3", Resource: "gateways"}
	_, err := c.k8sClient.Resource(res).Namespace(namespace).Get(context.Background(), name, v1.GetOptions{})

	if err != nil {
		return fmt.Errorf("gateway could not be found")
	}

	return nil
}

func (c *scenario) thereIsAKymaGatewayGardener(name string, namespace string) error {
	res := schema.GroupVersionResource{Group: "networking.istio.io", Version: "v1alpha3", Resource: "gateways"}
	_, err := c.k8sClient.Resource(res).Namespace(namespace).Get(context.Background(), name, v1.GetOptions{})

	if err != nil {
		return fmt.Errorf("gateway could not be found")
	}

	return nil
}
