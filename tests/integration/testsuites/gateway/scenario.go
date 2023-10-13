package gateway

import (
	"context"
	"errors"
	"fmt"
	"log"
	"path"

	"github.com/avast/retry-go/v4"
	"github.com/cucumber/godog"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/helpers"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/hooks"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/manifestprocessor"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/resource"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/testcontext"
	v12 "k8s.io/api/apps/v1"
	v2 "k8s.io/api/autoscaling/v2"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
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

	ctx.Before(hooks.DeleteBlockingResourcesScenarioHook)
	ctx.Before(hooks.ApplyApiGatewayCrScenarioHook)
	ctx.After(hooks.ApiGatewayCrTearDownScenarioHook)

	ctx.Step(`^APIGateway CR "([^"]*)" "([^"]*)" present$`, scenario.thereIsAnAPIGatewayCR)
	ctx.Step(`^APIGateway CR is in "([^"]*)" state with description "([^"]*)"$`, scenario.checkAPIGatewayCRState)
	ctx.Step(`^there is Istio Gateway "([^"]*)" in "([^"]*)" namespace$`, scenario.thereIsAGateway)
	ctx.Step(`^there is a "([^"]*)" secret in "([^"]*)" namespace$`, scenario.thereIsACertificate)
	ctx.Step(`^there is an "([^"]*)" APIRule with Gateway "([^"]*)"$`, scenario.thereIsAnAPIRule)
	ctx.Step(`^there is an "([^"]*)" VirtualService with Gateway "([^"]*)"$`, scenario.thereIsAVirtualService)
	ctx.Step(`^APIRule "([^"]*)" is removed$`, scenario.deleteAPIRule)
	ctx.Step(`^VirtualService "([^"]*)" is removed$`, scenario.deleteVirtualService)
	ctx.Step(`^disabling Kyma gateway$`, scenario.disableKymaGateway)
	ctx.Step(`^APIGateway CR "([^"]*)" is removed$`, scenario.deleteAPIGatewayCR)
	ctx.Step(`^gateway "([^"]*)" in "([^"]*)" namespace does not exist$`, scenario.thereIsNoGateway)
	ctx.Step(`^there is a "([^"]*)" Gardener Certificate CR in "([^"]*)" namespace$`, scenario.thereIsACertificateCR)
	ctx.Step(`^there is a "([^"]*)" Gardener DNSEntry CR in "([^"]*)" namespace$`, scenario.thereIsADNSEntryCR)
	ctx.Step(`there "([^"]*)" "([^"]*)" "([^"]*)" in the cluster`, scenario.resourceIsPresent)
	ctx.Step(`there "([^"]*)" "([^"]*)" "([^"]*)" in namespace "([^"]*)"`, scenario.namespacedResourceIsPresent)
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

func (c *scenario) thereIsAnAPIGatewayCR(name, isPresent string) error {
	const (
		is    = "is"
		isNot = "is not"
	)

	return retry.Do(func() error {
		res := schema.GroupVersionResource{Group: "operator.kyma-project.io", Version: "v1alpha1", Resource: "apigateways"}
		_, err := c.k8sClient.Resource(res).Get(context.Background(), name, v1.GetOptions{})
		if isPresent == is {
			if err != nil {
				if k8serrors.IsNotFound(err) {
					return fmt.Errorf("apigateway cr should be present but is not")
				}
				return err
			}
			return nil
		}

		if isPresent == isNot {
			if err == nil {
				return fmt.Errorf("apigateway cr, should not be present but is")
			}
			if k8serrors.IsNotFound(err) {
				return nil
			}
			return err
		}
		return fmt.Errorf("choose between %s and %s", is, isNot)
	}, testcontext.GetRetryOpts()...)
}

func (c *scenario) checkAPIGatewayCRState(state, description string) error {
	return retry.Do(func() error {
		res := schema.GroupVersionResource{Group: "operator.kyma-project.io", Version: "v1alpha1", Resource: "apigateways"}
		gateway, err := c.k8sClient.Resource(res).Get(context.Background(), hooks.ApiGatewayCRName, v1.GetOptions{})
		if err != nil {
			return err
		}

		gatewayState, found, err := unstructured.NestedString(gateway.Object, "status", "state")
		if err != nil {
			return err
		}

		if !found {
			return errors.New("status state not found")
		} else if gatewayState != state {
			return fmt.Errorf("gateway state %s, is not in the expected state %s", gatewayState, state)
		}

		if len(description) > 0 {
			gatewayDesc, found, err := unstructured.NestedString(gateway.Object, "status", "description")
			if err != nil {
				return err
			}

			if !found {
				return errors.New("status description not found")
			} else if gatewayDesc != description {
				return fmt.Errorf("gateway description %s, is not as the expected description %s", gatewayDesc, description)
			}
		}
		return nil
	}, testcontext.GetRetryOpts()...)
}

func (c *scenario) thereIsAGateway(name string, namespace string) error {
	return retry.Do(func() error {
		res := schema.GroupVersionResource{Group: "networking.istio.io", Version: "v1beta1", Resource: "gateways"}
		_, err := c.k8sClient.Resource(res).Namespace(namespace).Get(context.Background(), name, v1.GetOptions{})
		if err != nil {
			return fmt.Errorf("%s could not be found", name)
		}

		return nil
	}, testcontext.GetRetryOpts()...)
}

func (c *scenario) thereIsACertificate(name string, namespace string) error {
	return retry.Do(func() error {
		res := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "secrets"}
		_, err := c.k8sClient.Resource(res).Namespace(namespace).Get(context.Background(), name, v1.GetOptions{})
		if err != nil {
			return fmt.Errorf("%s could not be found", name)
		}

		return nil
	}, testcontext.GetRetryOpts()...)
}

func (c *scenario) thereIsAnAPIRule(name, gateway string) error {
	customDomainManifestDirectory := path.Dir(manifestsPath)
	apiRule, err := manifestprocessor.ParseFromFileWithTemplate("apirule.yaml", customDomainManifestDirectory, struct {
		Namespace string
		Gateway   string
	}{
		Namespace: c.namespace,
		Gateway:   gateway,
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

func (c *scenario) thereIsAVirtualService(name, gateway string) error {
	customDomainManifestDirectory := path.Dir(manifestsPath)
	vs, err := manifestprocessor.ParseFromFileWithTemplate("virtual-service.yaml", customDomainManifestDirectory, struct {
		Namespace string
		Gateway   string
	}{
		Namespace: c.namespace,
		Gateway:   gateway,
	})
	if err != nil {
		return fmt.Errorf("failed to process virtual-service.yaml, details %s", err.Error())
	}
	_, err = c.resourceManager.CreateResources(c.k8sClient, vs...)
	if err != nil {
		return err
	}

	res := schema.GroupVersionResource{Group: "networking.istio.io", Version: "v1alpha3", Resource: "virtualservices"}
	_, err = c.k8sClient.Resource(res).Namespace(c.namespace).Get(context.Background(), name, v1.GetOptions{})
	if err != nil {
		return fmt.Errorf("%s could not be found", name)
	}

	return nil
}

func (c *scenario) deleteAPIRule(name string) error {
	// res := schema.GroupVersionResource{Group: "gateway.kyma-project.io", Version: "v1beta1", Resource: "apirules"}
	// apiRule, err := c.k8sClient.Resource(res).Namespace(c.namespace).Get(context.Background(), name, v1.GetOptions{})
	// if err != nil {
	// 	return err
	// }

	// if apiRule.Object["metadata"].(map[string]interface{})["finalizers"] != nil {
	// 	apiRule.Object["metadata"].(map[string]interface{})["finalizers"] = nil
	// 	_, err = c.k8sClient.Resource(res).Namespace(c.namespace).Update(context.Background(), apiRule, v1.UpdateOptions{})
	// 	if err != nil {
	// 		return err
	// 	}
	// }
	// res := schema.GroupVersionResource{Group: "networking.istio.io", Version: "v1alpha3", Resource: "virtualservices"}
	// vsList, err := c.k8sClient.Resource(res).Namespace(c.namespace).List(context.Background(), v1.ListOptions{})
	// if err != nil {
	// 	return err
	// }
	// for _, vs := range vsList.Items {
	// 	if strings.HasPrefix(vs.GetName(), fmt.Sprintf("%s-", name)) {
	// 		err = c.deleteVirtualService(vs.GetName())
	// 		if err != nil {
	// 			return err
	// 		}
	// 	}
	// }
	res := schema.GroupVersionResource{Group: "gateway.kyma-project.io", Version: "v1beta1", Resource: "apirules"}
	err := c.k8sClient.Resource(res).Namespace(c.namespace).Delete(context.Background(), name, v1.DeleteOptions{})
	if err != nil {
		return err
	}

	return retry.Do(func() error {
		_, err = c.k8sClient.Resource(res).Namespace(c.namespace).Get(context.Background(), name, v1.GetOptions{})
		if err != nil {
			return nil
		}

		return fmt.Errorf("%s still exists", name)
	}, testcontext.GetRetryOpts()...)
}

func (c *scenario) deleteVirtualService(name string) error {
	res := schema.GroupVersionResource{Group: "networking.istio.io", Version: "v1alpha3", Resource: "virtualservices"}
	err := c.k8sClient.Resource(res).Namespace(c.namespace).Delete(context.Background(), name, v1.DeleteOptions{})
	if err != nil {
		return err
	}

	return retry.Do(func() error {
		_, err = c.k8sClient.Resource(res).Namespace(c.namespace).Get(context.Background(), name, v1.GetOptions{})
		if err != nil {
			return nil
		}

		return fmt.Errorf("%s still exists", name)
	}, testcontext.GetRetryOpts()...)
}

func (c *scenario) deleteAPIGatewayCR(name string) error {
	res := schema.GroupVersionResource{Group: "operator.kyma-project.io", Version: "v1alpha1", Resource: "apigateways"}
	err := c.k8sClient.Resource(res).Delete(context.Background(), name, v1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (c *scenario) disableKymaGateway() error {
	res := schema.GroupVersionResource{Group: "operator.kyma-project.io", Version: "v1alpha1", Resource: "apigateways"}
	apiGatewayCR, err := c.k8sClient.Resource(res).Get(context.Background(), hooks.ApiGatewayCRName, v1.GetOptions{})
	if err != nil {
		return err
	}
	apiGatewayCR.Object["spec"].(map[string]interface{})["enableKymaGateway"] = false
	_, err = c.k8sClient.Resource(res).Update(context.Background(), apiGatewayCR, v1.UpdateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (c *scenario) thereIsNoGateway(name, namespace string) error {
	return retry.Do(func() error {
		res := schema.GroupVersionResource{Group: "networking.istio.io", Version: "v1beta1", Resource: "gateways"}
		_, err := c.k8sClient.Resource(res).Namespace(namespace).Get(context.Background(), name, v1.GetOptions{})
		if err != nil {
			return nil
		}

		return fmt.Errorf("%s still exists", name)
	}, testcontext.GetRetryOpts()...)
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

func (c *scenario) resourceIsPresent(isPresent, kind, name string) error {
	const (
		is   = "is"
		isNo = "is no"
	)

	return retry.Do(func() error {
		gvr := resource.GetResourceGvr(kind, name)
		_, err := c.k8sClient.Resource(gvr).Get(context.Background(), name, v1.GetOptions{})
		if isPresent == is {
			if err != nil {
				if k8serrors.IsNotFound(err) {
					return fmt.Errorf("kind: %s, name: %s, should be present but is not", kind, name)
				}
				return err
			}

			return nil
		}

		if isPresent == isNo {
			if err == nil {
				return fmt.Errorf("kind: %s, name: %s, should not be present but is", kind, name)
			}
			if k8serrors.IsNotFound(err) {
				return nil
			}
			return err
		}
		return fmt.Errorf("choose between %s and %s", is, isNo)
	}, testcontext.GetRetryOpts()...)
}

func (c *scenario) namespacedResourceIsPresent(isPresent, kind, name, namespace string) error {
	const (
		is   = "is"
		isNo = "is no"
	)
	return retry.Do(func() error {
		gvr := resource.GetResourceGvr(kind, name)
		_, err := c.k8sClient.Resource(gvr).Namespace(namespace).Get(context.Background(), name, v1.GetOptions{})

		if isPresent == is {
			if err != nil {
				if k8serrors.IsNotFound(err) {
					return fmt.Errorf("kind: %s, name: %s, should be present but is not", kind, name)
				}
				return err
			}
			return nil
		}

		if isPresent == isNo {
			if err == nil {
				return fmt.Errorf("kind: %s, name: %s, should not be present but is", kind, name)
			}
			if k8serrors.IsNotFound(err) {
				return nil
			}
			return err
		}
		return fmt.Errorf("choose between %s and %s", is, isNo)

	}, testcontext.GetRetryOpts()...)
}

func (c *scenario) namespacedResourceHasStatusReady(kind, name, namespace string) error {
	return retry.Do(func() error {
		gvr := resource.GetResourceGvr(kind, name)
		unstr, err := c.k8sClient.Resource(gvr).Namespace(namespace).Get(context.Background(), name, v1.GetOptions{})
		if err != nil {
			return fmt.Errorf("could not get resource: %s, named: %s, in namespace %s", kind, name, namespace)
		}
		switch kind {
		case resource.Deployment.String():
			var dep v12.Deployment
			err = runtime.DefaultUnstructuredConverter.FromUnstructured(unstr.UnstructuredContent(), &dep)
			if err != nil {
				return fmt.Errorf("cannot convert unstructured to structured kind: %s, name: %s, namespace: %s", kind, name, namespace)
			}
			if dep.Status.UnavailableReplicas != 0 {
				return fmt.Errorf("kind: %s, name %s, namespace %s, is not Ready", kind, name, namespace)
			}
		case resource.HorizontalPodAutoscaler.String():
			var hpa v2.HorizontalPodAutoscaler
			err = runtime.DefaultUnstructuredConverter.FromUnstructured(unstr.UnstructuredContent(), &hpa)
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
	}, testcontext.GetRetryOpts()...)
}
