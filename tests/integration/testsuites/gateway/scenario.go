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
	v12 "k8s.io/api/apps/v1"
	v2 "k8s.io/api/autoscaling/v2"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"log"
	"path"
)

const manifestsPath = "testsuites/gateway/manifests/"

type godogResourceMapping int

func (k godogResourceMapping) String() string {
	switch k {
	case Deployment:
		return "Deployment"
	case Service:
		return "Service"
	case HorizontalPodAutoscaler:
		return "HorizontalPodAutoscaler"
	case ConfigMap:
		return "ConfigMap"
	case Secret:
		return "Secret"
	case CustomResourceDefinition:
		return "CustomResourceDefinition"
	case ServiceAccount:
		return "ServiceAccount"
	case ClusterRole:
		return "ClusterRole"
	case ClusterRoleBinding:
		return "ClusterRoleBinding"
	case PeerAuthentication:
		return "PeerAuthentication"
	}
	panic(fmt.Errorf("%#v has unimplemented String() method", k))
}

const (
	Deployment godogResourceMapping = iota
	Service
	HorizontalPodAutoscaler
	ConfigMap
	Secret
	CustomResourceDefinition
	ServiceAccount
	ClusterRole
	ClusterRoleBinding
	PeerAuthentication
)

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
	ctx.Step(`there is a "([^"]*)" "([^"]*)" in the cluster`, scenario.resourceIsPresent)
	ctx.Step(`there is a "([^"]*)" "([^"]*)" in namespace "([^"]*)"`, scenario.namespacedResourceIsPresent)
	ctx.Step(`"([^"]*)" "([^"]*)" in namespace "([^"]*)" has status "([^"]*)"`, scenario.namespacedResourceHasStatusReady)
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

func (c *scenario) resourceIsPresent(kind, name string) error {
	return retry.Do(func() error {
		gvr := getResourceGvr(kind, name)
		_, err := c.k8sClient.Resource(gvr).Get(context.Background(), name, v1.GetOptions{})
		if err != nil {
			if errors.IsNotFound(err) {
				return fmt.Errorf("could not find non namespaced resource %s, named: %s", kind, name)
			}
			return err
		}
		return nil
	}, testcontext.GetRetryOpts(c.config)...)
}

func (c *scenario) namespacedResourceIsPresent(kind, name, namespace string) error {
	return retry.Do(func() error {
		gvr := getResourceGvr(kind, name)
		_, err := c.k8sClient.Resource(gvr).Namespace(namespace).Get(context.Background(), name, v1.GetOptions{})
		if err != nil {
			if errors.IsNotFound(err) {
				return fmt.Errorf("could not find resource: %s, named: %s, in namespace %s", kind, name, namespace)
			}
			return err
		}
		return nil
	}, testcontext.GetRetryOpts(c.config)...)
}

func (c *scenario) namespacedResourceHasStatusReady(kind, name, namespace string) error {
	return retry.Do(func() error {
		gvr := getResourceGvr(kind, name)
		unstr, err := c.k8sClient.Resource(gvr).Namespace(namespace).Get(context.Background(), name, v1.GetOptions{})
		if err != nil {
			return fmt.Errorf("could not get resource: %s, named: %s, in namespace %s", kind, name, namespace)
		}
		switch kind {
		case Deployment.String():
			var dep v12.Deployment
			err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstr.UnstructuredContent(), &dep)
			if err != nil {
				return fmt.Errorf("cannot convert unstructured to structured kind: %s, name: %s, namespace: %s", kind, name, namespace)
			}
			if dep.Status.UnavailableReplicas != 0 {
				return fmt.Errorf("kind: %s, name %s, namespace %s, is not Ready", kind, name, namespace)
			}
		case HorizontalPodAutoscaler.String():
			var hpa v2.HorizontalPodAutoscaler
			err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstr.UnstructuredContent(), &hpa)
			if err != nil {
				return fmt.Errorf("cannot convert unstructured to structured kind: %s, name: %s, namespace: %s", kind, name, namespace)
			}
			if hpa.Status.CurrentReplicas != hpa.Status.DesiredReplicas {
				return fmt.Errorf("kind: %s, name %s, namespace %s, is not Ready", kind, name, namespace)
			}
		default:
			panic(fmt.Errorf("not implemented yet for kind: %s", kind))
		}
		return nil
	}, testcontext.GetRetryOpts(c.config)...)
}

func getResourceGvr(kind, name string) schema.GroupVersionResource {
	var gvr schema.GroupVersionResource
	switch kind {
	case Deployment.String():
		gvr = schema.GroupVersionResource{
			Group:    "apps",
			Version:  "v1",
			Resource: "deployments",
		}
	case Service.String():
		gvr = schema.GroupVersionResource{
			Group:    "",
			Version:  "v1",
			Resource: "services",
		}
	case HorizontalPodAutoscaler.String():
		gvr = schema.GroupVersionResource{
			Group:    "autoscaling",
			Version:  "v2",
			Resource: "horizontalpodautoscalers",
		}
	case ConfigMap.String():
		gvr = schema.GroupVersionResource{
			Group:    "",
			Version:  "v1",
			Resource: "configmaps",
		}
	case Secret.String():
		gvr = schema.GroupVersionResource{
			Group:    "",
			Version:  "v1",
			Resource: "secrets",
		}
	case CustomResourceDefinition.String():
		gvr = schema.GroupVersionResource{
			Group:    "apiextensions.k8s.io",
			Version:  "v1",
			Resource: "customresourcedefinitions",
		}
	case ServiceAccount.String():
		gvr = schema.GroupVersionResource{
			Group:    "",
			Version:  "v1",
			Resource: "serviceaccounts",
		}
	case ClusterRole.String():
		gvr = schema.GroupVersionResource{
			Group:    "rbac.authorization.k8s.io",
			Version:  "v1",
			Resource: "clusterroles",
		}
	case ClusterRoleBinding.String():
		gvr = schema.GroupVersionResource{
			Group:    "rbac.authorization.k8s.io",
			Version:  "v1",
			Resource: "clusterrolebindings",
		}
	case PeerAuthentication.String():
		gvr = schema.GroupVersionResource{
			Group:    "security.istio.io",
			Version:  "v1beta1",
			Resource: "peerauthentications",
		}
	default:
		panic(fmt.Errorf("cannot get gvr for kind: %s, name: %s", kind, name))
	}
	return gvr
}
