package gateway

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
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
)

//go:embed kyma_gateway.yaml
var kymaGatewayManifest []byte

func reconcileKymaGateway(ctx context.Context, k8sClient client.Client, apiGatewayCR v1alpha1.APIGateway, domain string) error {
	isEnabled := isKymaGatewayEnabled(apiGatewayCR)
	ctrl.Log.Info("Reconciling Kyma gateway", "KymaGatewayEnabled", isEnabled)

	if !isEnabled {
		return deleteKymaGateway(k8sClient)
	}

	return reconcileGateway(ctx, k8sClient, kymaGatewayName, kymaGatewayNamespace, domain)
}

func reconcileGateway(ctx context.Context, k8sClient client.Client, name, namespace, domain string) error {

	resourceTemplate, err := template.New("tmpl").Option("missingkey=error").Parse(string(kymaGatewayManifest))
	if err != nil {
		return fmt.Errorf("failed to parse template for gateway %s/%s: %v", namespace, name, err)
	}

	templateValues := make(map[string]string)
	templateValues["Name"] = name
	templateValues["Namespace"] = namespace
	templateValues["Domain"] = domain

	var resourceBuffer bytes.Buffer
	err = resourceTemplate.Execute(&resourceBuffer, templateValues)
	if err != nil {
		return fmt.Errorf("failed to apply parsed template for gateway %s/%s: %v", namespace, name, err)
	}

	var gateway unstructured.Unstructured
	err = yaml.Unmarshal(resourceBuffer.Bytes(), &gateway)
	if err != nil {
		return fmt.Errorf("failed to decode gateway yaml for %s/%s: %v", namespace, name, err)
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
		return fmt.Errorf("failed to create or update gateway %s/%s: %v", namespace, name, err)
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
