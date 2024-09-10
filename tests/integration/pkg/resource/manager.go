package resource

import (
	"context"
	"fmt"
	"github.com/avast/retry-go/v4"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/client"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/restmapper"
)

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
	case Role:
		return "Role"
	case RoleBinding:
		return "RoleBinding"
	case ClusterRole:
		return "ClusterRole"
	case ClusterRoleBinding:
		return "ClusterRoleBinding"
	case PeerAuthentication:
		return "PeerAuthentication"
	case PriorityClass:
		return "PriorityClass"
	case VirtualService:
		return "VirtualService"
	case Certificate:
		return "Certificate"
	case DNSEntry:
		return "DNSEntry"
	case PodDisruptionBudget:
		return "PodDisruptionBudget"
	case OryRule:
		return "Rule"
	case RequestAuthentication:
		return "RequestAuthentication"
	case AuthorizationPolicy:
		return "AuthorizationPolicy"
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
	Role
	RoleBinding
	ClusterRole
	ClusterRoleBinding
	PeerAuthentication
	PriorityClass
	VirtualService
	Certificate
	DNSEntry
	PodDisruptionBudget
	OryRule
	RequestAuthentication
	AuthorizationPolicy
)

type Manager struct {
	retryOptions []retry.Option
	mapper       *restmapper.DeferredDiscoveryRESTMapper
}

func NewManager(retryOpts []retry.Option) *Manager {

	mapper, err := client.GetDiscoveryMapper()
	if err != nil {
		panic(err)
	}

	return &Manager{
		retryOptions: retryOpts,
		mapper:       mapper,
	}
}

const errorCreatingResource = "error creating resource: %w"
const errorGettingResource = "error getting resource: %w"
const errorUpdatingResource = "error updating resource: %w"

func (m *Manager) CreateResources(k8sClient dynamic.Interface, resources ...unstructured.Unstructured) (*unstructured.Unstructured, error) {
	gotRes := &unstructured.Unstructured{}
	for _, res := range resources {
		resourceSchema, ns, _ := m.GetResourceSchemaAndNamespace(res)
		err := m.CreateResource(k8sClient, resourceSchema, ns, res)
		if err != nil {
			return nil, fmt.Errorf(errorCreatingResource, err)
		}
		gotRes, err = m.GetResource(k8sClient, resourceSchema, ns, res.GetName())
		if err != nil {
			return nil, fmt.Errorf(errorGettingResource, err)
		}
	}
	return gotRes, nil
}

func (m *Manager) CreateResourcesWithoutNS(k8sClient dynamic.Interface, resources ...unstructured.Unstructured) (*unstructured.Unstructured, error) {
	gotRes := &unstructured.Unstructured{}

	for _, res := range resources {
		resourceSchema, _, _ := m.GetResourceSchemaAndNamespace(res)
		err := m.CreateResourceWithoutNS(k8sClient, resourceSchema, res)
		if err != nil {
			return nil, fmt.Errorf(errorCreatingResource, err)
		}
		gotRes, err = m.GetResourceWithoutNS(k8sClient, resourceSchema, res.GetName())
		if err != nil {
			return nil, fmt.Errorf(errorGettingResource, err)
		}
	}

	return gotRes, nil
}

func (m *Manager) UpdateResources(k8sClient dynamic.Interface, resources ...unstructured.Unstructured) (*unstructured.Unstructured, error) {
	gotRes := &unstructured.Unstructured{}
	for _, res := range resources {
		resourceSchema, ns, _ := m.GetResourceSchemaAndNamespace(res)
		err := m.UpdateResource(k8sClient, resourceSchema, ns, res.GetName(), res)
		if err != nil {
			return nil, fmt.Errorf(errorUpdatingResource, err)
		}
		gotRes, err = m.GetResource(k8sClient, resourceSchema, ns, res.GetName())
		if err != nil {
			return nil, fmt.Errorf(errorGettingResource, err)
		}
	}
	return gotRes, nil
}

func (m *Manager) UpdateResourcesWithoutNS(k8sClient dynamic.Interface, resources ...unstructured.Unstructured) (*unstructured.Unstructured, error) {
	gotRes := &unstructured.Unstructured{}
	for _, res := range resources {
		resourceSchema, _, _ := m.GetResourceSchemaAndNamespace(res)
		err := m.UpdateResourceWithoutNS(k8sClient, resourceSchema, res.GetName(), res)
		if err != nil {
			return nil, fmt.Errorf(errorUpdatingResource, err)
		}
		gotRes, err = m.GetResourceWithoutNS(k8sClient, resourceSchema, res.GetName())
		if err != nil {
			return nil, fmt.Errorf(errorGettingResource, err)
		}
	}
	return gotRes, nil
}

func (m *Manager) CreateOrUpdateResources(k8sClient dynamic.Interface, resources ...unstructured.Unstructured) (*unstructured.Unstructured, error) {
	gotRes := &unstructured.Unstructured{}
	for _, res := range resources {
		resourceSchema, ns, _ := m.GetResourceSchemaAndNamespace(res)
		_, err := m.GetResource(k8sClient, resourceSchema, ns, res.GetName(), retry.Attempts(2), retry.Delay(1))

		if err != nil {
			if apierrors.IsNotFound(retry.Error{err}.Unwrap()) {
				err := m.CreateResource(k8sClient, resourceSchema, ns, res)
				if err != nil {
					return nil, fmt.Errorf(errorCreatingResource, err)
				}
			} else {
				return nil, fmt.Errorf(errorGettingResource, err)
			}
		} else {
			err = m.UpdateResource(k8sClient, resourceSchema, ns, res.GetName(), res)
			if err != nil {
				return nil, fmt.Errorf(errorUpdatingResource, err)
			}
		}
	}
	return gotRes, nil
}

func (m *Manager) CreateOrUpdateResourcesGVR(client dynamic.Interface, resources ...unstructured.Unstructured) (*unstructured.Unstructured, error) {
	gotRes := &unstructured.Unstructured{}
	for _, res := range resources {
		gvr, err := GetGvrFromUnstructured(m, res)
		if err != nil {
			return nil, err
		}
		_, err = client.Resource(*gvr).Namespace(res.GetNamespace()).Get(context.Background(), res.GetName(), metav1.GetOptions{})

		if err != nil {
			if apierrors.IsNotFound(err) {
				_, err := client.Resource(*gvr).Namespace(res.GetNamespace()).Create(context.Background(), &res, metav1.CreateOptions{})
				if err != nil {
					return nil, fmt.Errorf(errorCreatingResource, err)
				}
			} else {
				return nil, fmt.Errorf(errorGettingResource, err)
			}
		} else {
			err = m.UpdateResource(client, *gvr, res.GetNamespace(), res.GetName(), res)
			if err != nil {
				return nil, fmt.Errorf(errorUpdatingResource, err)
			}
		}
	}
	return gotRes, nil
}

func (m *Manager) CreateOrUpdateResourcesWithoutNS(k8sClient dynamic.Interface, resources ...unstructured.Unstructured) (*unstructured.Unstructured, error) {
	gotRes := &unstructured.Unstructured{}
	for _, res := range resources {
		resourceSchema, _, _ := m.GetResourceSchemaAndNamespace(res)
		_, err := m.GetResourceWithoutNS(k8sClient, resourceSchema, res.GetName(), retry.Attempts(2), retry.Delay(1))

		if err != nil {
			if apierrors.IsNotFound(retry.Error{err}.Unwrap()) {
				err := m.CreateResourceWithoutNS(k8sClient, resourceSchema, res)
				if err != nil {
					return nil, fmt.Errorf(errorCreatingResource, err)
				}
			} else {
				return nil, fmt.Errorf(errorGettingResource, err)
			}
		} else {
			err = m.UpdateResourceWithoutNS(k8sClient, resourceSchema, res.GetName(), res)
			if err != nil {
				return nil, fmt.Errorf(errorUpdatingResource, err)
			}
		}
	}
	return gotRes, nil
}

const errorDeletingResource = "error deleting resource: %w"

func (m *Manager) DeleteResources(k8sClient dynamic.Interface, resources ...unstructured.Unstructured) error {
	for _, res := range resources {
		resourceSchema, ns, name := m.GetResourceSchemaAndNamespace(res)
		err := m.DeleteResource(k8sClient, resourceSchema, ns, name)
		if err != nil {
			return fmt.Errorf(errorDeletingResource, err)
		}
	}
	return nil
}

func (m *Manager) DeleteResourcesWithoutNS(k8sClient dynamic.Interface, resources ...unstructured.Unstructured) error {
	for _, res := range resources {
		resourceSchema, _, name := m.GetResourceSchemaAndNamespace(res)
		err := m.DeleteResourceWithoutNS(k8sClient, resourceSchema, name)
		if err != nil {
			return fmt.Errorf(errorDeletingResource, err)
		}
	}
	return nil
}

func (m *Manager) GetResourceSchemaAndNamespace(manifest unstructured.Unstructured) (schema.GroupVersionResource, string, string) {
	namespace := manifest.GetNamespace()
	if namespace == "" {
		namespace = "default"
	}
	resourceName := manifest.GetName()

	if manifest.GroupVersionKind().Kind == "Namespace" {
		namespace = ""
	}

	mapping, err := m.mapper.RESTMapping(manifest.GroupVersionKind().GroupKind(), manifest.GroupVersionKind().Version)
	if err != nil {
		log.Fatal(err)
	}

	return mapping.Resource, namespace, resourceName
}

// CreateResource creates a given k8s resource
func (m *Manager) CreateResource(client dynamic.Interface, resourceSchema schema.GroupVersionResource, namespace string, manifest unstructured.Unstructured) error {
	return retry.Do(func() error {
		if _, err := client.Resource(resourceSchema).Namespace(namespace).Create(context.Background(), &manifest, metav1.CreateOptions{}); err != nil {
			return err
		}
		return nil
	}, m.retryOptions...)
}

// CreateResourceWithoutNS creates a given k8s resource without namespace
func (m *Manager) CreateResourceWithoutNS(client dynamic.Interface, resourceSchema schema.GroupVersionResource, manifest unstructured.Unstructured) error {
	return retry.Do(func() error {
		if _, err := client.Resource(resourceSchema).Create(context.Background(), &manifest, metav1.CreateOptions{}); err != nil {
			return err
		}
		return nil
	}, m.retryOptions...)
}

// UpdateResource updates a given k8s resource
func (m *Manager) UpdateResource(client dynamic.Interface, resourceSchema schema.GroupVersionResource, namespace string, name string, updateTo unstructured.Unstructured) error {
	return retry.Do(func() error {
		toUpdate, err := client.Resource(resourceSchema).Namespace(namespace).Get(context.Background(), name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		updateTo.SetResourceVersion(toUpdate.GetResourceVersion())
		_, err = client.Resource(resourceSchema).Namespace(namespace).Update(context.Background(), &updateTo, metav1.UpdateOptions{})
		if err != nil {
			return err
		}

		return nil
	}, m.retryOptions...)
}

// UpdateResourceWithoutNS updates a given k8s resource without namespace
func (m *Manager) UpdateResourceWithoutNS(client dynamic.Interface, resourceSchema schema.GroupVersionResource, name string, updateTo unstructured.Unstructured) error {
	return retry.Do(func() error {
		toUpdate, err := client.Resource(resourceSchema).Get(context.Background(), name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		updateTo.SetResourceVersion(toUpdate.GetResourceVersion())
		_, err = client.Resource(resourceSchema).Update(context.Background(), &updateTo, metav1.UpdateOptions{})
		if err != nil {
			return err
		}

		return nil
	}, m.retryOptions...)
}

// DeleteResource deletes a given k8s resource
func (m *Manager) DeleteResource(client dynamic.Interface, resourceSchema schema.GroupVersionResource, namespace string, resourceName string) error {
	return retry.Do(func() error {
		deletePolicy := metav1.DeletePropagationForeground
		deleteOptions := &metav1.DeleteOptions{
			PropagationPolicy: &deletePolicy,
		}
		if err := client.Resource(resourceSchema).Namespace(namespace).Delete(context.Background(), resourceName, *deleteOptions); err != nil {
			if !apierrors.IsNotFound(err) {
				return err
			}
		}
		return nil
	}, m.retryOptions...)
}

// DeleteResourceWithoutNS deletes a given k8s resource without namespace
func (m *Manager) DeleteResourceWithoutNS(client dynamic.Interface, resourceSchema schema.GroupVersionResource, resourceName string) error {
	return retry.Do(func() error {
		deletePolicy := metav1.DeletePropagationForeground
		deleteOptions := &metav1.DeleteOptions{
			PropagationPolicy: &deletePolicy,
		}
		if err := client.Resource(resourceSchema).Delete(context.Background(), resourceName, *deleteOptions); err != nil {
			if !apierrors.IsNotFound(err) {
				return err
			}
		}
		return nil
	}, m.retryOptions...)
}

// GetResource returns chosen k8s object
func (m *Manager) GetResource(client dynamic.Interface, resourceSchema schema.GroupVersionResource, namespace string, resourceName string, opts ...retry.Option) (*unstructured.Unstructured, error) {
	var res *unstructured.Unstructured
	if len(opts) == 0 {
		err := retry.Do(
			func() error {
				var err error
				res, err = client.Resource(resourceSchema).Namespace(namespace).Get(context.Background(), resourceName, metav1.GetOptions{})
				if err != nil {
					return err
				}
				return nil
			}, m.retryOptions...)
		if err != nil {
			return nil, err
		}
	} else {
		err := retry.Do(
			func() error {
				var err error
				res, err = client.Resource(resourceSchema).Namespace(namespace).Get(context.Background(), resourceName, metav1.GetOptions{})
				if err != nil {
					return err
				}
				return nil
			}, opts...)
		if err != nil {
			return nil, err
		}
	}

	return res, nil
}

// GetResourceWithoutNS returns chosen k8s object without namespace
func (m *Manager) GetResourceWithoutNS(client dynamic.Interface, resourceSchema schema.GroupVersionResource, resourceName string, opts ...retry.Option) (*unstructured.Unstructured, error) {
	var res *unstructured.Unstructured
	if len(opts) == 0 {
		err := retry.Do(
			func() error {
				var err error
				res, err = client.Resource(resourceSchema).Get(context.Background(), resourceName, metav1.GetOptions{})
				if err != nil {
					return err
				}
				return nil
			}, m.retryOptions...)
		if err != nil {
			return nil, err
		}
	} else {
		err := retry.Do(
			func() error {
				var err error
				res, err = client.Resource(resourceSchema).Get(context.Background(), resourceName, metav1.GetOptions{})
				if err != nil {
					return err
				}
				return nil
			}, opts...)
		if err != nil {
			return nil, err
		}
	}

	return res, nil
}

func (m *Manager) List(client dynamic.Interface, resourceSchema schema.GroupVersionResource, namespace string, listOptions metav1.ListOptions) (*unstructured.UnstructuredList, error) {
	var res *unstructured.UnstructuredList

	err := retry.Do(
		func() error {
			var err error
			res, err = client.Resource(resourceSchema).Namespace(namespace).List(context.Background(), listOptions)
			if err != nil {
				return err
			}
			return nil
		}, m.retryOptions...)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// GetStatus do a GetResource and extract status field
func (m *Manager) GetStatus(client dynamic.Interface, resourceSchema schema.GroupVersionResource, namespace string, resourceName string) (map[string]interface{}, error) {
	obj, err := m.GetResource(client, resourceSchema, namespace, resourceName)
	if err != nil {
		return nil, err
	}
	status, found, err := unstructured.NestedMap(obj.Object, "status")
	if err != nil || !found {
		return nil, fmt.Errorf("could not retrive status, or status not found:\n %+v", err)
	}
	return status, nil
}

func GetGvrFromUnstructured(m *Manager, resource unstructured.Unstructured) (*schema.GroupVersionResource, error) {
	gvk := resource.GroupVersionKind()
	mapping, err := m.mapper.RESTMapping(schema.GroupKind{
		Group: gvk.Group,
		Kind:  gvk.Kind,
	})
	if err != nil {
		return nil, err
	}
	gvr := &schema.GroupVersionResource{
		Group:    gvk.Group,
		Version:  gvk.Version,
		Resource: mapping.Resource.Resource,
	}
	return gvr, nil
}

func GetResourceGvr(kind string) schema.GroupVersionResource {
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
	case Role.String():
		gvr = schema.GroupVersionResource{
			Group:    "rbac.authorization.k8s.io",
			Version:  "v1",
			Resource: "roles",
		}
	case ClusterRole.String():
		gvr = schema.GroupVersionResource{
			Group:    "rbac.authorization.k8s.io",
			Version:  "v1",
			Resource: "clusterroles",
		}
	case RoleBinding.String():
		gvr = schema.GroupVersionResource{
			Group:    "rbac.authorization.k8s.io",
			Version:  "v1",
			Resource: "rolebindings",
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
	case PriorityClass.String():
		gvr = schema.GroupVersionResource{
			Group:    "scheduling.k8s.io",
			Version:  "v1",
			Resource: "priorityclasses",
		}
	case VirtualService.String():
		gvr = schema.GroupVersionResource{
			Group:    "networking.istio.io",
			Version:  "v1beta1",
			Resource: "virtualservices",
		}
	case Certificate.String():
		gvr = schema.GroupVersionResource{
			Group:    "cert.gardener.cloud",
			Version:  "v1alpha1",
			Resource: "certificates",
		}
	case DNSEntry.String():
		gvr = schema.GroupVersionResource{
			Group:    "dns.gardener.cloud",
			Version:  "v1alpha1",
			Resource: "dnsentries",
		}
	case PodDisruptionBudget.String():
		gvr = schema.GroupVersionResource{
			Group:    "policy",
			Version:  "v1",
			Resource: "poddisruptionbudgets",
		}
	case OryRule.String():
		gvr = schema.GroupVersionResource{
			Group:    "oathkeeper.ory.sh",
			Version:  "v1alpha1",
			Resource: "rules",
		}
	case RequestAuthentication.String():
		gvr = schema.GroupVersionResource{
			Group:    "security.istio.io",
			Version:  "v1",
			Resource: "requestauthentications",
		}
	case AuthorizationPolicy.String():
		gvr = schema.GroupVersionResource{
			Group:    "security.istio.io",
			Version:  "v1",
			Resource: "authorizationpolicies",
		}
	default:
		panic(fmt.Errorf("cannot get gvr for kind: %s", kind))
	}
	return gvr
}
