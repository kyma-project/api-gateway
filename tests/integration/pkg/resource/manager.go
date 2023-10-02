package resource

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/client"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/restmapper"
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

func (m *Manager) CreateResources(k8sClient dynamic.Interface, resources ...unstructured.Unstructured) (*unstructured.Unstructured, error) {
	gotRes := &unstructured.Unstructured{}
	for _, res := range resources {
		resourceSchema, ns, _ := m.GetResourceSchemaAndNamespace(res)
		err := m.CreateResource(k8sClient, resourceSchema, ns, res)
		if err != nil {
			return nil, err
		}
		gotRes, err = m.GetResource(k8sClient, resourceSchema, ns, res.GetName())
		if err != nil {
			return nil, err
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
			return nil, err
		}
		gotRes, err = m.GetResource(k8sClient, resourceSchema, ns, res.GetName())
		if err != nil {
			return nil, err
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
					return nil, err
				}
			} else {
				return nil, err
			}
		} else {
			err = m.UpdateResource(k8sClient, resourceSchema, ns, res.GetName(), res)
			if err != nil {
				return nil, err
			}
		}
	}
	return gotRes, nil
}

func (m *Manager) DeleteResources(k8sClient dynamic.Interface, resources ...unstructured.Unstructured) error {
	for _, res := range resources {
		resourceSchema, ns, name := m.GetResourceSchemaAndNamespace(res)
		err := m.DeleteResource(k8sClient, resourceSchema, ns, name)
		if err != nil {
			return err
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

// UpdateResource updates a given k8s resource
func (m *Manager) UpdateResource(client dynamic.Interface, resourceSchema schema.GroupVersionResource, namespace string, name string, updateTo unstructured.Unstructured) error {
	return retry.Do(func() error {
		time.Sleep(5 * time.Second) //TODO: delete after waiting for resource creation is implemented
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

// GetResource returns chosed k8s object
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

func (m *Manager) MergeAndUpdateOrCreateResources(client dynamic.Interface, resources []unstructured.Unstructured) error {
	for _, resource := range resources {
		gvr, err := getGvrFromUnstructured(m, resource)
		if err != nil {
			return err
		}
		r, err := client.Resource(*gvr).Namespace(resource.GetNamespace()).Get(context.TODO(), resource.GetName(), metav1.GetOptions{})

		if err != nil {
			if apierrors.IsNotFound(err) {
				_, err := client.Resource(*gvr).Namespace(resource.GetNamespace()).Create(context.TODO(), &resource, metav1.CreateOptions{})
				if err != nil {
					return err
				}
			} else {
				return err
			}
		} else {
			updatedResource := mergeUnstructured(r, &resource)
			gvr, err := getGvrFromUnstructured(m, *updatedResource)
			if err != nil {
				return err
			}
			_, err = client.Resource(*gvr).Namespace(updatedResource.GetNamespace()).Update(context.TODO(), updatedResource, metav1.UpdateOptions{})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func getGvrFromUnstructured(m *Manager, resource unstructured.Unstructured) (*schema.GroupVersionResource, error) {
	gvk := resource.GroupVersionKind()
	mapping, err := m.mapper.RESTMapping(schema.GroupKind{
		Group: gvk.Group,
		Kind:  gvk.Kind,
	})
	if err != nil {
		return nil, err
	}
	res := mapping.Resource.Resource
	gvr := &schema.GroupVersionResource{
		Group:    gvk.Group,
		Version:  gvk.Version,
		Resource: res,
	}

	return gvr, nil
}

func mergeUnstructured(old, new *unstructured.Unstructured) *unstructured.Unstructured {
	oldMap := old.Object
	newMap := new.Object
	mergeMaps(oldMap, newMap)
	return &unstructured.Unstructured{Object: oldMap}
}

func mergeMaps(o, n map[string]any) {
	for k, nv := range n {
		if ov, ok := o[k]; ok {
			ovm, oldIsMap := ov.(map[string]any)
			nvm, newIsMap := nv.(map[string]any)
			if oldIsMap && newIsMap {
				mergeMaps(ovm, nvm)
			} else {
				o[k] = nv
			}
		} else {
			o[k] = nv
		}
	}
}
