package gateway

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	"github.com/kyma-project/api-gateway/controllers"
	"github.com/kyma-project/api-gateway/internal/reconciliations"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	nonGardenerDomainName = "local.kyma.dev"
)

// ReconcileKymaGateway reconciles the kyma-gateway and creates all required resources for the Gateway to fully work. It also adds a finalizer to
// APIGateway CR and handles the deletion of the resources if the APIGateway CR is deleted.
// Returns a Status object with the result of the reconciliation and an error if the reconciliation failed.
func ReconcileKymaGateway(ctx context.Context, k8sClient client.Client, apiGatewayCR *v1alpha1.APIGateway) controllers.Status {

	if isKymaGatewayEnabled(*apiGatewayCR) {
		if !hasKymaGatewayFinalizer(*apiGatewayCR) {
			if err := addKymaGatewayFinalizer(ctx, k8sClient, apiGatewayCR); err != nil {
				return controllers.ErrorStatus(err, "Failed to add finalizer during Kyma Gateway reconciliation")
			}
		}
	} else {
		apiRuleExists, err := anyApiRuleExists(ctx, k8sClient)
		if err != nil {
			return controllers.ErrorStatus(err, "Error during evaluation of Kyma Gateway reconciliation")
		}

		// In the future, we want to be more selective and block the deletion of the Kyma gateway only if it is actually
		// used by an APIRule, since currently an APIRule and no Kyma gateway always result in an error status.
		if apiRuleExists {
			return controllers.WarningStatus(fmt.Errorf("kyma gateway deletion blocked by APIRules"), "Kyma Gateway cannot be disabled because APIRules exist.")
		}

		if !hasKymaGatewayFinalizer(*apiGatewayCR) {
			ctrl.Log.Info("Kyma Gateway is disabled and no finalizer exists, reconciliation is skipped.")
			return controllers.ReadyStatus()
		}
	}

	isGardenerCluster, err := reconciliations.RunsOnGardenerCluster(ctx, k8sClient)
	if err != nil {
		return controllers.ErrorStatus(err, "Error during Kyma Gateway reconciliation")
	}

	var reconcileErr error
	if isGardenerCluster {
		reconcileErr = reconcileGardenerKymaGateway(ctx, k8sClient, *apiGatewayCR)
	} else {
		reconcileErr = reconcileNonGardenerKymaGateway(ctx, k8sClient, *apiGatewayCR)
	}

	if reconcileErr != nil {
		return controllers.ErrorStatus(reconcileErr, "Error during Kyma Gateway reconciliation")
	}

	// Besides on disabling the Kyma gateway, we also need to remove the finalizer on APIGateway deletion to make sure we are not blocking the deletion of the CR.
	if !isKymaGatewayEnabled(*apiGatewayCR) || apiGatewayCR.IsInDeletion() {
		if err := removeKymaGatewayFinalizer(ctx, k8sClient, apiGatewayCR); err != nil {
			return controllers.ErrorStatus(err, "Failed to remove finalizer during Kyma Gateway reconciliation")
		}
	}

	return controllers.ReadyStatus()
}

func anyApiRuleExists(ctx context.Context, k8sClient client.Client) (bool, error) {
	apiRuleList := v1beta1.APIRuleList{}
	err := k8sClient.List(ctx, &apiRuleList)
	if err != nil {
		return false, err
	}

	return len(apiRuleList.Items) > 0, nil
}

func reconcileGardenerKymaGateway(ctx context.Context, k8sClient client.Client, apiGatewayCR v1alpha1.APIGateway) error {
	domain, err := reconciliations.GetGardenerDomain(ctx, k8sClient)
	if err != nil {
		return fmt.Errorf("failed to get Kyma gateway domain: %v", err)
	}

	if err := reconcileKymaGatewayDnsEntry(ctx, k8sClient, apiGatewayCR, domain); err != nil {
		return err
	}

	if err := reconcileKymaGatewayCertificate(ctx, k8sClient, apiGatewayCR, domain); err != nil {
		return err
	}

	if err := reconcileKymaGatewayVirtualService(ctx, k8sClient, apiGatewayCR, domain); err != nil {
		return err
	}

	return reconcileKymaGateway(ctx, k8sClient, apiGatewayCR, domain)
}

func reconcileNonGardenerKymaGateway(ctx context.Context, k8sClient client.Client, apiGatewayCR v1alpha1.APIGateway) error {
	if err := reconcileNonGardenerCertificateSecret(ctx, k8sClient, apiGatewayCR); err != nil {
		return err
	}

	return reconcileKymaGateway(ctx, k8sClient, apiGatewayCR, nonGardenerDomainName)
}
