package gateway

import (
	"context"
	"fmt"
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
	"log"
	"path"
)

const manifestsPath = "testsuites/gateway/manifests/"

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
		log.Fatalf("could not initialize custom domain endpoint err=%s", err)
	}

	ctx.Step(`APIGateway CR is applied`, scenario.applyAPIGatewayCR)
	ctx.Step(`^APIGateway CR is in "([^"]*)" state$`, scenario.thereIsAnAPIGatewayCR)
	ctx.Step(`^there is a "([^"]*)" gateway in "([^"]*)" namespace$`, scenario.thereIsAGateway)
	ctx.Step(`^there is a "([^"]*)" secret in "([^"]*)" namespace$`, scenario.thereIsACertificate)
	ctx.Step(`^there is an "([^"]*)" APIRule$`, scenario.thereIsAnAPIRule)
	ctx.Step(`^APIRule "([^"]*)" is removed$`, scenario.deleteAPIRule)
	ctx.Step(`^disabling kyma gateway will result in "([^"]*)" state$`, scenario.disableKymaGatewayAndCheckStatus)
	ctx.Step(`^gateway "([^"]*)" is removed$`, scenario.deleteGateway)
	ctx.Step(`^gateway "([^"]*)" in "([^"]*)" namespace does not exist$`, scenario.thereIsNoGateway)
	ctx.Step(`^there is a "([^"]*)" Gardener Certificate CR in "([^"]*)" namespace$`, scenario.thereIsACertificateCR)
	ctx.Step(`^there is a "([^"]*)" Gardener DNSEntry CR in "([^"]*)" namespace$`, scenario.thereIsADNSEntryCR)
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

func (c *scenario) applyAPIGatewayCR() error {
	customDomainManifestDirectory := path.Dir(manifestsPath)
	kymaGatewayEnabled, err := manifestprocessor.ParseFromFileWithTemplate("kyma-gateway-enabled.yaml", customDomainManifestDirectory, struct {
		NamePrefix string
	}{
		NamePrefix: c.config.GatewayCRName,
	})
	if err != nil {
		return fmt.Errorf("failed to process kyma-gateway-enabled.yaml, details %s", err.Error())
	}
	_, err = c.resourceManager.CreateOrUpdateResourcesWithoutNS(c.k8sClient, kymaGatewayEnabled...)
	if err != nil {
		return err
	}

	return nil
}

func (c *scenario) thereIsAnAPIGatewayCR(state string) error {
	return retry.Do(func() error {
		res := schema.GroupVersionResource{Group: "operator.kyma-project.io", Version: "v1alpha1", Resource: "apigateways"}
		gateway, err := c.k8sClient.Resource(res).Get(context.Background(), c.config.GatewayCRName, v1.GetOptions{})
		if err != nil {
			return fmt.Errorf("gateway could not be found")
		}

		gatewayState, found, err := unstructured.NestedString(gateway.Object, "status", "state")
		if err != nil || !found {
			return err
		} else if gatewayState != state {
			return fmt.Errorf("gateway state %s, is not in the expected state %s", gatewayState, state)
		}

		return nil
	}, testcontext.GetRetryOpts(c.config)...)
}

func (c *scenario) thereIsAGateway(name string, namespace string) error {
	return retry.Do(func() error {
		res := schema.GroupVersionResource{Group: "networking.istio.io", Version: "v1beta1", Resource: "gateways"}
		_, err := c.k8sClient.Resource(res).Namespace(namespace).Get(context.Background(), name, v1.GetOptions{})
		if err != nil {
			return fmt.Errorf("%s could not be found", name)
		}

		return nil
	}, testcontext.GetRetryOpts(c.config)...)
}

func (c *scenario) thereIsACertificate(name string, namespace string) error {
	return retry.Do(func() error {
		res := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "secrets"}
		_, err := c.k8sClient.Resource(res).Namespace(namespace).Get(context.Background(), name, v1.GetOptions{})
		if err != nil {
			return fmt.Errorf("%s could not be found", name)
		}

		return nil
	}, testcontext.GetRetryOpts(c.config)...)
}

func (c *scenario) thereIsAnAPIRule(name string) error {
	customDomainManifestDirectory := path.Dir(manifestsPath)
	apiRule, err := manifestprocessor.ParseFromFileWithTemplate("apirule.yaml", customDomainManifestDirectory, struct {
		Namespace string
	}{
		Namespace: c.namespace,
	})
	if err != nil {
		return fmt.Errorf("failed to process apirule.yaml, details %s", err.Error())
	}
	_, err = c.resourceManager.CreateResources(c.k8sClient, apiRule...)
	if err != nil {
		return err
	}

	res := schema.GroupVersionResource{Group: "gateway.kyma-project.io", Version: "v1beta1", Resource: "apirules"}
	_, err = c.k8sClient.Resource(res).Namespace(c.namespace).Get(context.Background(), name, v1.GetOptions{})
	if err != nil {
		return fmt.Errorf("%s could not be found", name)
	}

	return nil
}

func (c *scenario) deleteAPIRule(name string) error {
	customDomainManifestDirectory := path.Dir(manifestsPath)
	apiRule, err := manifestprocessor.ParseFromFileWithTemplate("apirule.yaml", customDomainManifestDirectory, struct {
		Namespace string
	}{
		Namespace: c.namespace,
	})
	if err != nil {
		return fmt.Errorf("failed to process apirule.yaml, details %s", err.Error())
	}

	err = c.resourceManager.DeleteResources(c.k8sClient, apiRule...)
	if err != nil {
		return err
	}

	return retry.Do(func() error {
		res := schema.GroupVersionResource{Group: "gateway.kyma-project.io", Version: "v1beta1", Resource: "apirules"}
		_, err = c.k8sClient.Resource(res).Namespace(c.namespace).Get(context.Background(), name, v1.GetOptions{})
		if err != nil {
			return nil
		}

		return fmt.Errorf("%s stil exists", name)
	}, testcontext.GetRetryOpts(c.config)...)
}

func (c *scenario) deleteGateway(name string) error {
	customDomainManifestDirectory := path.Dir(manifestsPath)
	gateway, err := manifestprocessor.ParseFromFileWithTemplate("kyma-gateway-enabled.yaml", customDomainManifestDirectory, struct {
		NamePrefix string
	}{
		NamePrefix: c.config.GatewayCRName,
	})
	if err != nil {
		return fmt.Errorf("failed to process kyma-gateway-enabled.yaml, details %s", err.Error())
	}

	err = c.resourceManager.DeleteResourcesWithoutNS(c.k8sClient, gateway...)
	if err != nil {
		return err
	}

	return retry.Do(func() error {
		res := schema.GroupVersionResource{Group: "operator.kyma-project.io", Version: "v1alpha1", Resource: "apigateways"}
		_, err = c.k8sClient.Resource(res).Get(context.Background(), name, v1.GetOptions{})
		if err != nil {
			return nil
		}

		return fmt.Errorf("%s stil exists", name)
	}, testcontext.GetRetryOpts(c.config)...)
}

func (c *scenario) disableKymaGatewayAndCheckStatus(state string) error {
	customDomainManifestDirectory := path.Dir(manifestsPath)
	kymaGatewayDisabled, err := manifestprocessor.ParseFromFileWithTemplate("kyma-gateway-disabled.yaml", customDomainManifestDirectory, struct {
		NamePrefix string
	}{
		NamePrefix: c.config.GatewayCRName,
	})
	if err != nil {
		return fmt.Errorf("failed to process kyma-gateway-disabled.yaml, details %s", err.Error())
	}
	_, err = c.resourceManager.UpdateResourcesWithoutNS(c.k8sClient, kymaGatewayDisabled...)
	if err != nil {
		return err
	}

	return retry.Do(func() error {
		res := schema.GroupVersionResource{Group: "operator.kyma-project.io", Version: "v1alpha1", Resource: "apigateways"}
		gateway, err := c.k8sClient.Resource(res).Get(context.Background(), c.config.GatewayCRName, v1.GetOptions{})
		if err != nil {
			return fmt.Errorf("gateway could not be found")
		}

		gatewayState, found, err := unstructured.NestedString(gateway.Object, "status", "state")
		if err != nil || !found {
			return fmt.Errorf("could not get gateway status: %s", err)
		} else if gatewayState != state {
			return fmt.Errorf("gateway state %s, is not in the expected state %s", gatewayState, state)
		}

		return nil
	}, testcontext.GetRetryOpts(c.config)...)
}

func (c *scenario) thereIsNoGateway(name, namespace string) error {
	return retry.Do(func() error {
		res := schema.GroupVersionResource{Group: "networking.istio.io", Version: "v1beta1", Resource: "gateways"}
		_, err := c.k8sClient.Resource(res).Namespace(namespace).Get(context.Background(), name, v1.GetOptions{})
		if err != nil {
			return nil
		}

		return fmt.Errorf("%s stil exists", name)
	}, testcontext.GetRetryOpts(c.config)...)
}

func (c *scenario) thereIsACertificateCR(name, namespace string) error {
	res := schema.GroupVersionResource{Group: "cert.gardener.cloud", Version: "v1alpha1", Resource: "certificates"}
	_, err := c.k8sClient.Resource(res).Namespace(namespace).Get(context.Background(), name, v1.GetOptions{})
	if err != nil {
		return fmt.Errorf("%s could not be found", name)
	}

	return nil
}

func (c *scenario) thereIsADNSEntryCR(name, namespace string) error {
	res := schema.GroupVersionResource{Group: "dns.gardener.cloud", Version: "v1alpha1", Resource: "dnsentries"}
	_, err := c.k8sClient.Resource(res).Namespace(namespace).Get(context.Background(), name, v1.GetOptions{})
	if err != nil {
		return fmt.Errorf("%s could not be found", name)
	}

	return nil
}
