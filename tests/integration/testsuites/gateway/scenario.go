package gateway

import (
	"context"
	"fmt"
	"github.com/cucumber/godog"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/helpers"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/resource"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/testcontext"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"log"
	"time"
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

	ctx.Step(`^there is an APIGateway operator in "([^"]*)" state$`, scenario.thereIsAnAPIGatewayOperator)
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
