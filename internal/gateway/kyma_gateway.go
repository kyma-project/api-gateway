package gateway

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	"github.com/kyma-project/api-gateway/controllers"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	"k8s.io/apimachinery/pkg/util/yaml"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"text/template"
)

const (
	kymaGatewayName      = "kyma-gateway"
	kymaGatewayNamespace = "kyma-system"
	disclaimerKey        = "apigateways.operator.kyma-project.io/managed-by-disclaimer"
	disclaimerValue      = "DO NOT EDIT - This resource is managed by Kyma.\nAny modifications are discarded and the resource is reverted to the original state."
)

//go:embed kyma_gateway.yaml
var manifest []byte

// Reconcile returns a status reflecting
func Reconcile(ctx context.Context, k8sClient client.Client, apiGatewayCR v1alpha1.APIGateway) controllers.Status {

	if !isKymaGatewayEnabled(apiGatewayCR) {

		apiRuleExists, err := anyApiRuleExists(ctx, k8sClient)
		if err != nil {
			return controllers.NewErrorStatus(err, "Error during evaluation of Kyma Gateway reconciliation")
		}

		// In the future, we want to be more selective and block the deletion of the Kyma gateway only if it is actually
		// used by an APIRule, since currently an APIRule and no Kyma gateway always result in an error status.
		if apiRuleExists {
			return controllers.NewWarningStatus(fmt.Errorf("kyma gateway deletion blocked by APIRules"), "Kyma Gateway cannot be disabled because APIRules exist.")
		}
	}

	if err := reconcileKymaGateway(ctx, k8sClient, apiGatewayCR); err != nil {
		return controllers.NewErrorStatus(err, "Error during Kyma Gateway reconciliation")
	}

	return controllers.NewSuccessfulStatus()
}

func reconcileKymaGateway(ctx context.Context, k8sClient client.Client, apiGatewayCR v1alpha1.APIGateway) error {
	isEnabled := isKymaGatewayEnabled(apiGatewayCR)
	ctrl.Log.Info("Reconciling Kyma gateway", "KymaGatewayEnabled", isEnabled)

	if !isEnabled {
		return deleteKymaGateway(k8sClient)
	}

	resourceTemplate, err := template.New("tmpl").Option("missingkey=error").Parse(string(manifest))
	if err != nil {
		return fmt.Errorf("failed to parse template for Kyma gateway yaml: %v", err)
	}

	templateValues := make(map[string]string)
	domain, err := getKymaGatewayDomain(ctx, k8sClient)
	if err != nil {
		return fmt.Errorf("failed to get Kyma gateway domain: %v", err)
	}
	templateValues["Name"] = kymaGatewayName
	templateValues["Namespace"] = kymaGatewayNamespace
	templateValues["DomainName"] = domain

	var resourceBuffer bytes.Buffer
	err = resourceTemplate.Execute(&resourceBuffer, templateValues)
	if err != nil {
		return fmt.Errorf("failed to apply parsed template for Kyma gateway yaml: %v", err)
	}

	var gateway unstructured.Unstructured
	err = yaml.Unmarshal(resourceBuffer.Bytes(), &gateway)
	if err != nil {
		return fmt.Errorf("failed to decode Kyma gateway yaml: %v", err)
	}

	spec := gateway.Object["spec"]
	_, err = controllerutil.CreateOrUpdate(ctx, k8sClient, &gateway, func() error {
		annotations := map[string]string{
			disclaimerKey: disclaimerValue,
		}
		gateway.SetAnnotations(annotations)
		gateway.Object["spec"] = spec
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to create or update Kyma gateway: %v", err)
	}

	return nil
}

func isKymaGatewayEnabled(cr v1alpha1.APIGateway) bool {
	return cr.Spec.EnableKymaGateway != nil && *cr.Spec.EnableKymaGateway == true
}

func deleteKymaGateway(k8sClient client.Client) error {
	ctrl.Log.Info("Deleting Kyma gateway if it exists")
	gw := v1alpha3.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kymaGatewayName,
			Namespace: kymaGatewayNamespace,
		},
	}
	err := k8sClient.Delete(context.TODO(), &gw)

	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to delete Kyma gateway")
	}

	if err == nil {
		ctrl.Log.Info("Successfully deleted Kyma gateway")
	}

	return nil
}

func anyApiRuleExists(ctx context.Context, k8sClient client.Client) (bool, error) {
	apiRuleList := v1beta1.APIRuleList{}
	err := k8sClient.List(ctx, &apiRuleList)
	if err != nil {
		return false, err
	}

	return len(apiRuleList.Items) > 0, nil
}
