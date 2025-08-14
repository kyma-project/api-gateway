package gateway

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	"github.com/kyma-project/api-gateway/internal/reconciliations"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	apiRuleCRDName string = "apirules.gateway.kyma-project.io"
)

//go:embed apirule_crd.yaml
var apiruleCRD []byte

func reconcileAPIRuleCRD(ctx context.Context, k8sClient client.Client, apiGatewayCR v1alpha1.APIGateway) error {
	ctrl.Log.Info("Reconciling APIRule CRD")

	if apiGatewayCR.IsInDeletion() {
		return deleteAPIRuleCRD(ctx, k8sClient)
	}

	return reconciliations.ApplyResource(ctx, k8sClient, apiruleCRD, map[string]string{})
}

func deleteAPIRuleCRD(ctx context.Context, k8sClient client.Client) error {
	ctrl.Log.Info("Deleting APIRule CRD if it exists", "name", apiRuleCRDName)
	c := v1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: apiRuleCRDName,
		},
	}
	err := k8sClient.Delete(ctx, &c)

	if err != nil && !k8serrors.IsNotFound(err) {
		return fmt.Errorf("failed to delete APIRule CRD %s: %v", apiRuleCRDName, err)
	}

	if k8serrors.IsNotFound(err) {
		ctrl.Log.Info("Skipped deletion of APIRule CRD as it wasn't present", "name", apiRuleCRDName)
	} else {
		ctrl.Log.Info("Successfully deleted APIRule CRD", "name", apiRuleCRDName)
	}

	return nil
}
