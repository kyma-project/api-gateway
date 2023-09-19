package gateway

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	"github.com/kyma-project/api-gateway/controllers"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	nonGardenerDomainName = "local.kyma.dev"
	disclaimerKey         = "apigateways.operator.kyma-project.io/managed-by-disclaimer"
	disclaimerValue       = "DO NOT EDIT - This resource is managed by Kyma.\nAny modifications are discarded and the resource is reverted to the original state."
)

// Reconcile returns a status reflecting
func Reconcile(ctx context.Context, k8sClient client.Client, apiGatewayCR v1alpha1.APIGateway) controllers.Status {

	if !isKymaGatewayEnabled(apiGatewayCR) {

		apiRuleExists, err := anyApiRuleExists(ctx, k8sClient)
		if err != nil {
			return controllers.ErrorStatus(err, "Error during evaluation of Kyma Gateway reconciliation")
		}

		// In the future, we want to be more selective and block the deletion of the Kyma gateway only if it is actually
		// used by an APIRule, since currently an APIRule and no Kyma gateway always result in an error status.
		if apiRuleExists {
			return controllers.WarningStatus(fmt.Errorf("kyma gateway deletion blocked by APIRules"), "Kyma Gateway cannot be disabled because APIRules exist.")
		}
	}

	isGardenerCluster, err := runsOnGardnerCluster(ctx, k8sClient)
	if err != nil {
		return controllers.ErrorStatus(err, "Error during Kyma Gateway reconciliation")
	}

	if isGardenerCluster {

		domain, err := getGardenerDomain(ctx, k8sClient)
		if err != nil {
			err = fmt.Errorf("failed to get Kyma gateway domain: %v", err)
			return controllers.ErrorStatus(err, "Error during Kyma Gateway reconciliation")
		}

		if err := reconcileKymaGatewayDnsEntry(ctx, k8sClient, apiGatewayCR, domain); err != nil {
			return controllers.ErrorStatus(err, "Error during Kyma Gateway DNSEntry reconciliation")
		}

		if err := reconcileKymaGatewayCertificate(ctx, k8sClient, apiGatewayCR, domain); err != nil {
			return controllers.ErrorStatus(err, "Error during Kyma Gateway Certificate reconciliation")
		}

		if err := reconcileKymaGateway(ctx, k8sClient, apiGatewayCR, domain); err != nil {
			return controllers.ErrorStatus(err, "Error during Kyma Gateway reconciliation")
		}

	} else {
		if err := reconcileKymaGateway(ctx, k8sClient, apiGatewayCR, nonGardenerDomainName); err != nil {
			return controllers.ErrorStatus(err, "Error during Kyma Gateway reconciliation")
		}

		if err := reconcileNonGardenerCertificateSecret(ctx, k8sClient); err != nil {
			return controllers.ErrorStatus(err, "Error during Kyma Gateway Certificate Secret reconciliation")
		}
	}

	return controllers.SuccessfulStatus()
}

func anyApiRuleExists(ctx context.Context, k8sClient client.Client) (bool, error) {
	apiRuleList := v1beta1.APIRuleList{}
	err := k8sClient.List(ctx, &apiRuleList)
	if err != nil {
		return false, err
	}

	return len(apiRuleList.Items) > 0, nil
}
