package oathkeeper

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"time"

	"github.com/avast/retry-go/v4"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	"github.com/kyma-project/api-gateway/internal/reconciliations"
)

//go:embed crd.yaml
var crd []byte

const crdName = "rules.oathkeeper.ory.sh"

func reconcileOryOathkeeperRuleCRD(ctx context.Context, k8sClient client.Client, apiGatewayCR v1alpha1.APIGateway) error {
	ctrl.Log.Info("Reconciling Oathkeeper Rule CRD", "name", crdName)

	if apiGatewayCR.IsInDeletion() {
		return deleteCRD(ctx, k8sClient, crdName)
	}

	templateValues := make(map[string]string)
	templateValues["Name"] = crdName

	return reconciliations.ApplyResource(ctx, k8sClient, crd, templateValues)
}

func deleteCRD(ctx context.Context, k8sClient client.Client, name string) error {
	ctrl.Log.Info("Deleting Oathkeeper Rule CRD if it exists", "name", name)
	s := apiextensionsv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	err := k8sClient.Delete(ctx, &s)

	if err != nil && !k8serrors.IsNotFound(err) {
		return fmt.Errorf("failed to delete Oathkeeper Rule CRD %s: %w", name, err)
	}

	if k8serrors.IsNotFound(err) {
		ctrl.Log.Info("Skipped deletion of Oathkeeper Rule CRD as it wasn't present", "name", name)
	} else {
		ctrl.Log.Info("Successfully deleted Oathkeeper Rule CRD", "name", name)
	}

	// Wait for CRD to be successfully deleted, giving time for Oathkeeper maester to clean up Rule CRs
	return retry.Do(func() error {
		var crd apiextensionsv1.CustomResourceDefinition
		err := k8sClient.Get(ctx, types.NamespacedName{
			Name: crdName,
		}, &crd)

		if !k8serrors.IsNotFound(err) {
			return errors.New("rule crd is still present")
		}
		return nil
	}, retry.Attempts(10), retry.Delay(2*time.Second), retry.DelayType(retry.FixedDelay))
}
