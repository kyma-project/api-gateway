package externalgateway

import (
	"context"
	"fmt"

	networkingv1beta1 "istio.io/api/networking/v1beta1"
	istiov1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	externalv1alpha1 "github.com/kyma-project/api-gateway/apis/gateway/external/v1alpha1"
)

// ReconcileGateway creates or updates the Istio Gateway resource with mTLS
func ReconcileGateway(ctx context.Context, k8sClient client.Client, scheme *runtime.Scheme, external *externalv1alpha1.ExternalGateway, internalDomain string) error {
	gatewayName := external.GatewayName()
	namespace := external.Namespace

	ctrl.Log.Info("Reconciling Gateway with mTLS", "name", gatewayName, "namespace", namespace)

	gateway := &istiov1beta1.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Name:      gatewayName,
			Namespace: namespace,
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, k8sClient, gateway, func() error {
		// Set labels
		gateway.Labels = GetStandardLabels(external)

		// Set owner reference
		if err := controllerutil.SetControllerReference(external, gateway, scheme); err != nil {
			return fmt.Errorf("failed to set owner reference: %w", err)
		}

		// Set desired spec
		gateway.Spec = networkingv1beta1.Gateway{
			Selector: map[string]string{
				"istio": "ingressgateway",
			},
			Servers: []*networkingv1beta1.Server{
				{
					Port: &networkingv1beta1.Port{
						Number:   443,
						Name:     "https-external",
						Protocol: "HTTPS",
					},
					Hosts: []string{
						external.Spec.ExternalDomain,
						internalDomain,
					},
					Tls: &networkingv1beta1.ServerTLSSettings{
						Mode:           networkingv1beta1.ServerTLSSettings_MUTUAL,
						CredentialName: fmt.Sprintf("%s-tls", gatewayName),
					},
				},
			},
		}

		ctrl.Log.Info("Configured Gateway", "name", gatewayName, "mode", "MUTUAL")
		return nil
	})

	return err
}

// DeleteGateway deletes the Gateway resource
func DeleteGateway(ctx context.Context, k8sClient client.Client, namespace, gatewayName string) error {
	ctrl.Log.Info("Deleting Gateway if it exists", "name", gatewayName, "namespace", namespace)

	gateway := &istiov1beta1.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Name:      gatewayName,
			Namespace: namespace,
		},
	}

	err := k8sClient.Delete(ctx, gateway)
	if err != nil && !k8serrors.IsNotFound(err) {
		return fmt.Errorf("failed to delete Gateway %s/%s: %w", namespace, gatewayName, err)
	}

	if k8serrors.IsNotFound(err) {
		ctrl.Log.Info("Skipped deletion of Gateway as it wasn't present", "name", gatewayName)
	} else {
		ctrl.Log.Info("Successfully deleted Gateway", "name", gatewayName)
	}

	return nil
}
