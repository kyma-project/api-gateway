package modules

import (
	_ "embed"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/client"
	"testing"

	"github.com/kyma-project/api-gateway/tests/e2e/pkg/setup"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	apiRuleConfigMapName        = "api-gateway-config.operator.kyma-project.io"
	apiRuleConfigMapNamespace   = "kyma-system"
	enableAPIRuleV1ConfigMapKey = "enableDeprecatedV1beta1APIRule"
)

func CreateDeprecatedV1configMap(t *testing.T) error {
	t.Helper()
	t.Log("Creating deprecated v1beta1 configmap")

	r, err := client.ResourcesClient(t)
	if err != nil {
		t.Logf("Failed to get resources client: %v", err)
		return err
	}

	cm := getDeprecatedV1beta1ConfigMap()
	err = r.Create(t.Context(), cm)
	if err != nil {
		t.Logf("Failed to create deprecated v1beta1 configmap: %v", err)
		return err
	}

	setup.DeclareCleanup(t, func() {
		t.Log("Cleaning up deprecated v1beta1 configmap")
		err := DeleteDeprecatedV1beta1ConfigMap(t)
		if err != nil {
			t.Logf("Failed to clean up deprecated v1beta1 configmap: %v", err)
		} else {
			t.Log("deprecated v1beta1 configmap cleaned up successfully")
		}
	})

	t.Log("deprecated v1beta1 configmap created successfully")
	return nil
}

func DeleteDeprecatedV1beta1ConfigMap(t *testing.T) error {
	t.Helper()
	t.Log("Beginning cleanup deprecated v1beta1 configmap")

	r, err := client.ResourcesClient(t)
	if err != nil {
		t.Logf("Failed to get resources client: %v", err)
		return err
	}

	cm := getDeprecatedV1beta1ConfigMap()
	err = r.Delete(setup.GetCleanupContext(), cm)
	if err != nil {
		t.Logf("Failed to delete deprecated v1beta1 configmap: %v", err)
		if k8serrors.IsNotFound(err) {
			t.Log("deprecated v1beta1 configmap not found, nothing to delete")
			return nil
		}
		return err
	}

	return nil
}

func getDeprecatedV1beta1ConfigMap() *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      apiRuleConfigMapName,
			Namespace: apiRuleConfigMapNamespace,
		},
		Data: map[string]string{
			enableAPIRuleV1ConfigMapKey: "true",
		},
	}
}
