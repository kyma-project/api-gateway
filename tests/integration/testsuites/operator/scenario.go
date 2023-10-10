package operator

import (
	"context"
	"errors"
	"fmt"
	"log"
	"path"

	"github.com/avast/retry-go/v4"
	"github.com/cucumber/godog"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/helpers"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/manifestprocessor"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/resource"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/testcontext"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

const manifestsPath = "testsuites/operator/manifests/"

type scenario struct {
	testID          string
	namespace       string
	k8sClient       dynamic.Interface
	resourceManager *resource.Manager
	config          testcontext.Config
}

func initScenario(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario, err := createScenario(ts)

	if err != nil {
		log.Fatalf("could not initialize scenario err=%s", err)
	}

	ctx.Step(`APIGateway CR is applied`, scenario.applyApiGatewayCR)
	ctx.Step(`^APIGateway CR is in "([^"]*)" state with description "([^"]*)"$`, scenario.apiGatewayCRinState)
	ctx.Step(`^APIGateway CR is deleted$`, scenario.deleteApiGatewayCR)
	ctx.Step(`^APIGateway CR is "([^"]*)" on cluster$`, scenario.findApiGatewayCR)
	ctx.Step(`^APIRule "([^"]*)" is applied$`, scenario.applyAnAPIRule)
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

func (c *scenario) applyApiGatewayCR() error {
	manifestDirectory := path.Dir(manifestsPath)
	apiGatewayCR, err := manifestprocessor.ParseFromFileWithTemplate("api-gateway-cr.yaml", manifestDirectory, struct {
		Name string
	}{
		Name: c.config.GatewayCRName,
	})
	if err != nil {
		return fmt.Errorf("failed to process api-gateway-cr.yaml, details %s", err.Error())
	}
	_, err = c.resourceManager.CreateOrUpdateResourcesWithoutNS(c.k8sClient, apiGatewayCR...)
	if err != nil {
		return err
	}

	return nil
}

func (c *scenario) apiGatewayCRinState(state, description string) error {
	return retry.Do(func() error {
		res := schema.GroupVersionResource{Group: "operator.kyma-project.io", Version: "v1alpha1", Resource: "apigateways"}
		apiGatewayCR, err := c.k8sClient.Resource(res).Get(context.Background(), c.config.GatewayCRName, v1.GetOptions{})
		if err != nil {
			return fmt.Errorf("API-Gateway CR could not be found")
		}

		apiGatewayCRState, found, err := unstructured.NestedString(apiGatewayCR.Object, "status", "state")
		if err != nil || !found {
			return err
		}
		if !found {
			return errors.New("unable to find API-Gateway CR status state")
		} else if apiGatewayCRState != state {
			return fmt.Errorf("API-Gateway CR status state %s, is not as expected %s", apiGatewayCRState, state)
		}

		if description != "" {
			apiGatewayCRDesc, found, err := unstructured.NestedString(apiGatewayCR.Object, "status", "description")
			if err != nil {
				return err
			}

			if !found {
				return errors.New("unable to find API-Gateway CR state description")
			} else if apiGatewayCRDesc != description {
				return fmt.Errorf("API-Gateway CR state description %s, is not as expected %s", apiGatewayCRDesc, description)
			}
		}

		return nil
	}, testcontext.GetRetryOpts(c.config)...)
}

func (c *scenario) deleteApiGatewayCR() error {
	manifestDirectory := path.Dir(manifestsPath)
	apiGatewayCR, err := manifestprocessor.ParseFromFileWithTemplate("api-gateway-cr.yaml", manifestDirectory, struct {
		Name string
	}{
		Name: c.config.GatewayCRName,
	})
	if err != nil {
		return fmt.Errorf("failed to process api-gateway-cr.yaml, details %s", err.Error())
	}

	nsResourceSchema, ns, name := c.resourceManager.GetResourceSchemaAndNamespace(apiGatewayCR[0])
	err = c.resourceManager.DeleteResource(c.k8sClient, nsResourceSchema, ns, name)
	if err != nil {
		return err
	}

	return nil
}

func (c *scenario) findApiGatewayCR(expected string) error {
	return retry.Do(func() error {
		res := schema.GroupVersionResource{Group: "operator.kyma-project.io", Version: "v1alpha1", Resource: "apigateways"}
		gateway, err := c.k8sClient.Resource(res).Get(context.Background(), c.config.GatewayCRName, v1.GetOptions{})
		if err != nil {
			return fmt.Errorf("gateway could not be found")
		}

		switch expected {
		case "present":
			if gateway == nil {
				return fmt.Errorf("API-Gateway CR is not found, but expected to be %s", expected)
			}
		case "not present":
			if gateway != nil {
				return fmt.Errorf("API-Gateway CR is found, but expected to be %s", expected)
			}
		}

		return nil
	}, testcontext.GetRetryOpts(c.config)...)
}

func (c *scenario) applyAnAPIRule(name string) error {
	customDomainManifestDirectory := path.Dir(manifestsPath)
	apiRule, err := manifestprocessor.ParseFromFileWithTemplate("apirule.yaml", customDomainManifestDirectory, struct {
		Name      string
		Namespace string
	}{
		Name:      name,
		Namespace: c.namespace,
	})
	if err != nil {
		return fmt.Errorf("failed to process apirule.yaml, details %s", err.Error())
	}
	_, err = c.resourceManager.CreateResources(c.k8sClient, apiRule...)
	if err != nil {
		return err
	}

	return nil
}
