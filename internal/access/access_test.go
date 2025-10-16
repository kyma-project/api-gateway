package access

import (
	"context"
	_ "embed"
	operatorv1alpha1 "github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/yaml"
	"testing"
)

// Shoot info with domain: test-shoot.com
//
//go:embed shoot_info_cm.yaml
var shootInfoCMYAML []byte

// Access config map with signature for domain: test-shoot.com
//
//go:embed access_cm.yaml
var accessCMYAML []byte

// Access config map with signature for wildcard domain: *.test-shoot.com
//
//go:embed access_cm_wildcard.yaml
var accessCMWildcardYAML []byte

// Access config map with signature for domain: test-shoot.wrong
//
//go:embed access_cm_wrong_domain.yaml
var accessCMWrongDomainYAML []byte

func TestShouldAllowAccessToV1Beta1(t *testing.T) {
	t.Run("should allow access to v1beta1 for wildcard domain", func(t *testing.T) {
		var shootInfoCM, accessCM unstructured.Unstructured
		shootInfoCM.SetGroupVersionKind(schema.GroupVersionKind{
			Group:   "",
			Version: "v1",
			Kind:    "ConfigMap",
		})
		accessCM.SetGroupVersionKind(schema.GroupVersionKind{
			Group:   "",
			Version: "v1",
			Kind:    "ConfigMap",
		})

		err := yaml.Unmarshal(shootInfoCMYAML, &shootInfoCM.Object)
		require.NoError(t, err)
		err = yaml.Unmarshal(accessCMWildcardYAML, &accessCM.Object)
		require.NoError(t, err)

		k8sClient := createFakeClient(t, &shootInfoCM, &accessCM)
		allowed, err := ShouldAllowAccessToV1Beta1(context.Background(), k8sClient)
		assert.NoError(t, err)
		assert.True(t, allowed)

		// Update the shoot info to another domain that matches the wildcard
		shootInfoCM.Object["data"].(map[string]interface{})["domain"] = "sub.test-shoot.com"
		k8sClient = createFakeClient(t, &shootInfoCM, &accessCM)
		allowed, err = ShouldAllowAccessToV1Beta1(context.Background(), k8sClient)
		assert.NoError(t, err)
		assert.True(t, allowed)
	})

	t.Run("should allow access to v1beta1 when signature is valid, and not to a wildcard domain", func(t *testing.T) {
		var shootInfoCM, accessCM unstructured.Unstructured
		shootInfoCM.SetGroupVersionKind(schema.GroupVersionKind{
			Group:   "",
			Version: "v1",
			Kind:    "ConfigMap",
		})
		accessCM.SetGroupVersionKind(schema.GroupVersionKind{
			Group:   "",
			Version: "v1",
			Kind:    "ConfigMap",
		})

		err := yaml.Unmarshal(shootInfoCMYAML, &shootInfoCM.Object)
		require.NoError(t, err)
		err = yaml.Unmarshal(accessCMYAML, &accessCM.Object)
		require.NoError(t, err)

		k8sClient := createFakeClient(t, &shootInfoCM, &accessCM)
		allowed, err := ShouldAllowAccessToV1Beta1(context.Background(), k8sClient)
		assert.NoError(t, err)
		assert.True(t, allowed)

		// Update the shoot info to a subdomain that does not match the exact domain
		shootInfoCM.Object["data"].(map[string]interface{})["domain"] = "sub.test-shoot.com"
		k8sClient = createFakeClient(t, &shootInfoCM, &accessCM)
		allowed, err = ShouldAllowAccessToV1Beta1(context.Background(), k8sClient)
		assert.NoError(t, err)
		assert.False(t, allowed)
	})

	t.Run("should disallow access to v1beta1 when domain does not match", func(t *testing.T) {
		var shootInfoCM, accessCM unstructured.Unstructured
		shootInfoCM.SetGroupVersionKind(schema.GroupVersionKind{
			Group:   "",
			Version: "v1",
			Kind:    "ConfigMap",
		})
		accessCM.SetGroupVersionKind(schema.GroupVersionKind{
			Group:   "",
			Version: "v1",
			Kind:    "ConfigMap",
		})

		err := yaml.Unmarshal(shootInfoCMYAML, &shootInfoCM.Object)
		require.NoError(t, err)
		err = yaml.Unmarshal(accessCMWrongDomainYAML, &accessCM.Object)
		require.NoError(t, err)

		k8sClient := createFakeClient(t, &shootInfoCM, &accessCM)
		allowed, err := ShouldAllowAccessToV1Beta1(context.Background(), k8sClient)
		assert.NoError(t, err)
		assert.False(t, allowed)
	})
}
func createFakeClient(t *testing.T, objects ...client.Object) client.Client {
	err := operatorv1alpha1.AddToScheme(scheme.Scheme)
	require.NoError(t, err)
	err = corev1.AddToScheme(scheme.Scheme)
	require.NoError(t, err)

	return fake.NewClientBuilder().WithScheme(scheme.Scheme).WithObjects(objects...).Build()
}
