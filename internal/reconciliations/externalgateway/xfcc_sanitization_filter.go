package externalgateway

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/types/known/structpb"
	networkingv1alpha3 "istio.io/api/networking/v1alpha3"
	istiov1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	externalv1alpha1 "github.com/kyma-project/api-gateway/apis/gateway/external/v1alpha1"
)

// ReconcileXFCCSanitizationFilter creates or updates the EnvoyFilter that configures native Envoy XFCC handling
// Uses NETWORK_FILTER to configure forward_client_cert_details: FORWARD_ONLY
func ReconcileXFCCSanitizationFilter(ctx context.Context, k8sClient client.Client, external *externalv1alpha1.ExternalGateway) error {
	filterName := external.XFCCFilterName()

	ctrl.Log.Info("Reconciling XFCC sanitization EnvoyFilter", "name", filterName, "namespace", istioSystemNamespace)

	envoyFilter := &istiov1alpha3.EnvoyFilter{
		ObjectMeta: metav1.ObjectMeta{
			Name:      filterName,
			Namespace: istioSystemNamespace,
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, k8sClient, envoyFilter, func() error {
		// Set labels
		envoyFilter.Labels = GetStandardLabels(external)

		// Build patch for NETWORK_FILTER with forward_client_cert_details: FORWARD_ONLY
		patch := &networkingv1alpha3.EnvoyFilter_Patch{
			Operation: networkingv1alpha3.EnvoyFilter_Patch_MERGE,
			Value: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"typed_config": structpb.NewStructValue(&structpb.Struct{
						Fields: map[string]*structpb.Value{
							"@type": structpb.NewStringValue(
								"type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager",
							),
							"forward_client_cert_details": structpb.NewStringValue("FORWARD_ONLY"),
						},
					}),
				},
			},
		}

		// Set desired spec with NETWORK_FILTER
		envoyFilter.Spec = networkingv1alpha3.EnvoyFilter{
			WorkloadSelector: &networkingv1alpha3.WorkloadSelector{
				Labels: map[string]string{
					"app":   "istio-ingressgateway",
					"istio": "ingressgateway",
				},
			},
			ConfigPatches: []*networkingv1alpha3.EnvoyFilter_EnvoyConfigObjectPatch{
				{
					ApplyTo: networkingv1alpha3.EnvoyFilter_NETWORK_FILTER,
					Match: &networkingv1alpha3.EnvoyFilter_EnvoyConfigObjectMatch{
						Context: networkingv1alpha3.EnvoyFilter_GATEWAY,
						ObjectTypes: &networkingv1alpha3.EnvoyFilter_EnvoyConfigObjectMatch_Listener{
							Listener: &networkingv1alpha3.EnvoyFilter_ListenerMatch{
								FilterChain: &networkingv1alpha3.EnvoyFilter_ListenerMatch_FilterChainMatch{
									Sni: external.Spec.ExternalDomain,
									Filter: &networkingv1alpha3.EnvoyFilter_ListenerMatch_FilterMatch{
										Name: "envoy.filters.network.http_connection_manager",
									},
								},
							},
						},
					},
					Patch: patch,
				},
			},
		}

		ctrl.Log.Info("Configured XFCC sanitization EnvoyFilter with native Envoy forward_client_cert_details: FORWARD_ONLY", "name", filterName)
		return nil
	})

	return err
}

// DeleteXFCCSanitizationFilter deletes the XFCC sanitization EnvoyFilter
func DeleteXFCCSanitizationFilter(ctx context.Context, k8sClient client.Client, filterName string) error {

	ctrl.Log.Info("Deleting XFCC sanitization EnvoyFilter if it exists", "name", filterName, "namespace", istioSystemNamespace)

	envoyFilter := &istiov1alpha3.EnvoyFilter{
		ObjectMeta: metav1.ObjectMeta{
			Name:      filterName,
			Namespace: istioSystemNamespace,
		},
	}

	err := k8sClient.Delete(ctx, envoyFilter)
	if err != nil && !k8serrors.IsNotFound(err) {
		return fmt.Errorf("failed to delete XFCC sanitization EnvoyFilter %s/%s: %w", istioSystemNamespace, filterName, err)
	}

	if k8serrors.IsNotFound(err) {
		ctrl.Log.Info("Skipped deletion of XFCC sanitization EnvoyFilter as it wasn't present", "name", filterName)
	} else {
		ctrl.Log.Info("Successfully deleted XFCC sanitization EnvoyFilter", "name", filterName)
	}

	return nil
}
