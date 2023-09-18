package gateway

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	dnsv1alpha1 "github.com/gardener/external-dns-management/pkg/apis/dns/v1alpha1"
	"github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/yaml"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"text/template"
)

const (
	kymaGatewayDnsEntryName      = "kyma-gateway"
	kymaGatewayDnsEntryNamespace = "kyma-system"
)

//go:embed dns_entry.yaml
var dnsEntryManifest []byte

func reconcileKymaGatewayDnsEntry(ctx context.Context, k8sClient client.Client, apiGatewayCR v1alpha1.APIGateway, domain string) error {
	name := kymaGatewayDnsEntryName
	namespace := kymaGatewayDnsEntryNamespace

	isEnabled := isKymaGatewayEnabled(apiGatewayCR)
	ctrl.Log.Info("Reconciling DNS entry", "KymaGatewayEnabled", isEnabled, "Name", name, "Namespace", namespace)

	if !isEnabled {
		return deleteDnsEntry(k8sClient, name, namespace)
	}

	istioIngressIp, err := fetchIstioIngressGatewayIp(ctx, k8sClient)
	if err != nil {
		return fmt.Errorf("failed to fetch Istio ingress gateway IP: %v", err)
	}

	return reconcileDnsEntry(ctx, k8sClient, name, namespace, domain, istioIngressIp)
}

func reconcileDnsEntry(ctx context.Context, k8sClient client.Client, name, namespace, domain, ingressGatewayIp string) error {

	resourceTemplate, err := template.New("tmpl").Option("missingkey=error").Parse(string(dnsEntryManifest))
	if err != nil {
		return fmt.Errorf("failed to parse template for DNSEntry yaml: %v", err)
	}

	templateValues := make(map[string]string)
	templateValues["Name"] = name
	templateValues["Namespace"] = namespace
	templateValues["Domain"] = domain
	templateValues["IngressGatewayServiceIp"] = ingressGatewayIp

	var resourceBuffer bytes.Buffer
	err = resourceTemplate.Execute(&resourceBuffer, templateValues)
	if err != nil {
		return fmt.Errorf("failed to apply parsed template for DNSEntry yaml: %v", err)
	}

	var dnsEntry unstructured.Unstructured
	err = yaml.Unmarshal(resourceBuffer.Bytes(), &dnsEntry)
	if err != nil {
		return fmt.Errorf("failed to decode DNSEntry yaml: %v", err)
	}

	spec := dnsEntry.Object["spec"]
	_, err = controllerutil.CreateOrUpdate(ctx, k8sClient, &dnsEntry, func() error {
		annotations := map[string]string{
			disclaimerKey: disclaimerValue,
		}
		dnsEntry.SetAnnotations(annotations)
		dnsEntry.Object["spec"] = spec
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to create or update DNSEntry: %v", err)
	}

	return nil

}

func deleteDnsEntry(k8sClient client.Client, name, namespace string) error {
	ctrl.Log.Info("Deleting DNSEntry if it exists", "Name", name, "Namespace", namespace)
	gw := dnsv1alpha1.DNSEntry{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	err := k8sClient.Delete(context.TODO(), &gw)

	if err != nil && !k8serrors.IsNotFound(err) {
		return fmt.Errorf("failed to delete DNSEntry %s/%s: %v", namespace, name, err)
	}

	if err == nil {
		ctrl.Log.Info("Successfully deleted DNSEntry", "Name", name, "Namespace", namespace)
	}

	return nil
}

func fetchIstioIngressGatewayIp(ctx context.Context, k8sClient client.Client) (string, error) {

	istioIngressGatewayNamespaceName := types.NamespacedName{
		Name:      "istio-ingressgateway",
		Namespace: "istio-system",
	}

	svc := corev1.Service{}
	if err := k8sClient.Get(ctx, istioIngressGatewayNamespaceName, &svc); err != nil {
		return "", err
	}

	if len(svc.Status.LoadBalancer.Ingress) == 0 {
		return "", fmt.Errorf("no ingress ip set for %s", istioIngressGatewayNamespaceName.String())
	}

	return svc.Status.LoadBalancer.Ingress[0].IP, nil
}
