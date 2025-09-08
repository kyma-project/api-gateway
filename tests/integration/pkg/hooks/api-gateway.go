package hooks

import (
	"bytes"
	"context"
	"fmt"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"log"
	"os"

	"github.com/avast/retry-go/v4"
	"github.com/cucumber/godog"
	ratelimit "github.com/kyma-project/api-gateway/apis/gateway/ratelimit/v1alpha1"
	v2 "github.com/kyma-project/api-gateway/apis/gateway/v2"
	"github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	oryv1alpha1 "github.com/kyma-project/api-gateway/internal/types/ory/oathkeeper-maester/api/v1alpha1"
	k8sclient "github.com/kyma-project/api-gateway/tests/integration/pkg/client"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/testcontext"
	"github.com/kyma-project/api-gateway/tests/integration/testsuites/patch"
	"github.com/pkg/errors"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

const templateFileName string = "pkg/hooks/manifests/apigateway.yaml"
const ApiGatewayCRName string = "default"

const (
	kymaDNSName                 = "kyma-gateway"
	kymaDNSNamespace            = "kyma-system"
	kymaGatewayName             = "kyma-gateway"
	kymaGatewayNamespace        = "kyma-system"
	kymaCertName                = "kyma-tls-cert"
	kymaCertNamespace           = "istio-system"
	shootInfoConfigMapName      = "shoot-info"
	shootInfoConfigMapNamespace = "kube-system"
	apiRuleConfigMapName        = "api-gateway-config.operator.kyma-project.io"
	apiRuleConfigMapNamespace   = "kyma-system"
	enableAPIRuleV1ConfigMapKey = "enableDeprecatedV1beta1APIRule"
)

var dnsKind = schema.GroupVersionKind{Group: "dns.gardener.cloud", Version: "v1alpha1", Kind: "DNSEntry"}
var gatewayKind = schema.GroupVersionKind{Group: "networking.istio.io", Version: "v1alpha3", Kind: "Gateway"}
var certKind = schema.GroupVersionKind{Group: "cert.gardener.cloud", Version: "v1alpha1", Kind: "Certificate"}

var ApplyApiGatewayCrScenarioHook = func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
	k8sClient, err := testcontext.GetK8sClientFromContext(ctx)
	if err != nil {
		return ctx, err
	}

	if err := createDeprecatedV1ConfigMap(context.Background(), k8sClient); err != nil {
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

var DeleteBlockingResourcesScenarioHook = func(ctx context.Context, sc *godog.Scenario, _ error) (context.Context, error) {
	return ctx, deleteBlockingResources(ctx)
}

var Removev2alpha1VersionRequiredFieldsHook = func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
	k8sClient, err := testcontext.GetK8sClientFromContext(ctx)
	if err != nil {
		return ctx, err
	}

	return ctx, patch.Removev2alpha1VersionRequiredFields(k8sClient)
}

var ApiGatewayCrTearDownScenarioHook = func(ctx context.Context, sc *godog.Scenario, _ error) (context.Context, error) {
	if apiGateways, ok := testcontext.GetApiGatewayCRsFromContext(ctx); ok {
		// We can ignore a failed removal of the ApiGateway CR, because we need to run force remove in any case to make sure no resource is left before the next scenario
		for _, apiGateway := range apiGateways {
			_ = retry.Do(func() error {
				err := removeObjectFromCluster(ctx, apiGateway)
				if err != nil {
					return fmt.Errorf("failed to delete ApiGateway CR %s", apiGateway.GetName())
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

func applyAndVerifyApiGateway(scaleDownOathkeeper bool) error {
	log.Printf("Creating APIGateway CR %s", ApiGatewayCRName)
	k8sClient := k8sclient.GetK8sClient()

	if err := createDeprecatedV1ConfigMap(context.Background(), k8sClient); err != nil {
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

var ApplyAndVerifyApiGatewayWithoutOathkeeperCrSuiteHook = func() error {
	return applyAndVerifyApiGateway(true)
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

func waitUntilObjectIsRemoved(ctx context.Context, gvk schema.GroupVersionKind, objectName string, namespace string) error {
	c, err := testcontext.GetK8sClientFromContext(ctx)
	if err != nil {
		return err
	}

	unstructuredObj := unstructured.Unstructured{}
	unstructuredObj.SetGroupVersionKind(gvk)

	err = c.Get(ctx, client.ObjectKey{Namespace: namespace, Name: objectName}, &unstructuredObj)

	if k8serrors.IsNotFound(err) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error while getting object [kind: %s, name: %s, namespace: %s]: %w", gvk, objectName, namespace, err)
	}

	if unstructuredObj.GetDeletionTimestamp() == nil {
		return fmt.Errorf("object [kind: %s, name: %s, namespace: %s] has no deletion timestamp", gvk, objectName, namespace)
	}

	return retry.Do(func() error {
		err = c.Get(ctx, client.ObjectKey{Namespace: namespace, Name: objectName}, &unstructuredObj)

		if k8serrors.IsNotFound(err) {
			return nil
		}

		if err != nil {
			return fmt.Errorf("error while getting object [kind: %s, name: %s, namespace: %s]: %w", gvk, objectName, namespace, err)
		}

		return fmt.Errorf("object [kind: %s, name: %s, namespace: %s] still exists", gvk, objectName, namespace)

	}, testcontext.GetRetryOpts()...)
}

var WaitUntilApiGatewayDepsAreRemovedHook = func(ctx context.Context, sc *godog.Scenario, _ error) (context.Context, error) {
	err := waitUntilObjectIsRemoved(ctx, dnsKind, kymaDNSName, kymaDNSNamespace)
	if err != nil {
		return ctx, err
	}

	err = waitUntilObjectIsRemoved(ctx, gatewayKind, kymaGatewayName, kymaGatewayNamespace)
	if err != nil {
		return ctx, err
	}

	err = waitUntilObjectIsRemoved(ctx, certKind, kymaCertName, kymaCertNamespace)
	if err != nil {
		return ctx, err
	}

	return ctx, nil
}

func createDeprecatedV1ConfigMap(ctx context.Context, c client.Client) error {
	log.Printf("Creating APIGateway V1 ConfigMap")
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      shootInfoConfigMapName,
			Namespace: shootInfoConfigMapNamespace,
		},
		Data: map[string]string{
			"domain": "test-shoot.com",
		},
	}

	if err := c.Create(ctx, cm); err != nil {
		if !k8serrors.IsAlreadyExists(err) {
			return err
		}
		existing := &corev1.ConfigMap{}
		if err := c.Get(ctx, client.ObjectKey{Name: shootInfoConfigMapName, Namespace: shootInfoConfigMapNamespace}, existing); err != nil {
			return err
		}
		if existing.Data == nil {
			existing.Data = map[string]string{}
		}
		for k, v := range cm.Data {
			existing.Data[k] = v
		}
		return c.Update(ctx, existing)
	}

	cm2 := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "apirule-access",
			Namespace: "kyma-system",
		},
		BinaryData: map[string][]byte{
			"access.sig": []byte("owGbwMvMwCXG+Pmv5SmepjrGNRJJzCn5yRk7je+XpBaX6BZn5OeX6CXn53J1lLIwiHExyIopsmgF3dY/0e5yW1vcaS1MJysTSA8DF6cATES7gpFh5aZXlgkd4QqTPputkT2ge2jN/Zar1kv9lry7+FS+fZVsDcN/t7PbX35LClFsTD53Zu+rC71HgleFCmydI9LXLf6KN1mWFQA="),
		},
	}
	if err := c.Create(ctx, cm2); err != nil {
		if !k8serrors.IsAlreadyExists(err) {
			return err
		}
		existing := &corev1.ConfigMap{}
		if err := c.Get(ctx, client.ObjectKey{Name: "apirule-access", Namespace: "kyma-system"}, existing); err != nil {
			existing.BinaryData = map[string][]byte{}
		}
		for k, v := range cm2.BinaryData {
			existing.BinaryData[k] = v
		}
		return c.Update(ctx, existing)
	}
	return nil
}
