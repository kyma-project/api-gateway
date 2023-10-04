package oathkeeper

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	"github.com/kyma-project/api-gateway/internal/reconciliations"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//go:embed crd.yaml
var crd []byte

const crdName = "rules.oathkeeper.ory.sh"

func reconcileOryOathkeeperRuleCRD(ctx context.Context, k8sClient client.Client, apiGatewayCR v1alpha1.APIGateway) error {
	ctrl.Log.Info("Reconciling Ory Config PeerAuthentication", "name", crdName)

	if apiGatewayCR.IsInDeletion() {
		return deleteCRD(k8sClient, crdName)
	}

	templateValues := make(map[string]string)
	templateValues["Name"] = crdName

	return reconciliations.ApplyResource(ctx, k8sClient, crd, templateValues)
}

func deleteCRD(k8sClient client.Client, name string) error {
	ctrl.Log.Info("Deleting Oathkeeper Rule CRD if it exists", "name", name)
	s := apiextensionsv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	err := k8sClient.Delete(context.Background(), &s)

	if err != nil && !k8serrors.IsNotFound(err) {
		return fmt.Errorf("failed to delete Oathkeeper Rule CRD %s: %v", name, err)
	}

	ctrl.Log.Info("Successfully deleted Oathkeeper Rule CRD", "name", name)

	return nil
}
