package externalgateway

import (
	"context"
	"fmt"
	"strings"

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

// ReconcileCertValidationFilter creates or updates the EnvoyFilter that validates client certificate X509 fields
func ReconcileCertValidationFilter(ctx context.Context, k8sClient client.Client, external *externalv1alpha1.ExternalGateway, certSubjects []string) error {
	filterName := fmt.Sprintf("%s-cert-validation", external.GatewayName())

	ctrl.Log.Info("Reconciling certificate validation EnvoyFilter", "name", filterName, "namespace", istioSystemNamespace, "regions", len(certSubjects))

	// Generate Lua script with X509 field-based validation
	luaScript := buildValidationLuaScript(certSubjects)

	envoyFilter := &istiov1alpha3.EnvoyFilter{
		ObjectMeta: metav1.ObjectMeta{
			Name:      filterName,
			Namespace: istioSystemNamespace,
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, k8sClient, envoyFilter, func() error {
		// Set labels
		envoyFilter.Labels = GetStandardLabels(external)

		// Build patch for Lua filter
		patch := &networkingv1alpha3.EnvoyFilter_Patch{
			Operation: networkingv1alpha3.EnvoyFilter_Patch_INSERT_BEFORE,
			Value: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"name": structpb.NewStringValue("envoy.filters.http.lua"),
					"typed_config": structpb.NewStructValue(&structpb.Struct{
						Fields: map[string]*structpb.Value{
							"@type": structpb.NewStringValue("type.googleapis.com/envoy.extensions.filters.http.lua.v3.Lua"),
							"defaultSourceCode": structpb.NewStructValue(&structpb.Struct{
								Fields: map[string]*structpb.Value{
									"inlineString": structpb.NewStringValue(luaScript),
								},
							}),
						},
					}),
				},
			},
		}

		// Set desired spec
		envoyFilter.Spec = networkingv1alpha3.EnvoyFilter{
			WorkloadSelector: &networkingv1alpha3.WorkloadSelector{
				Labels: map[string]string{
					"app":   "istio-ingressgateway",
					"istio": "ingressgateway",
				},
			},
			ConfigPatches: []*networkingv1alpha3.EnvoyFilter_EnvoyConfigObjectPatch{
				{
					ApplyTo: networkingv1alpha3.EnvoyFilter_HTTP_FILTER,
					Match: &networkingv1alpha3.EnvoyFilter_EnvoyConfigObjectMatch{
						Context: networkingv1alpha3.EnvoyFilter_GATEWAY,
						ObjectTypes: &networkingv1alpha3.EnvoyFilter_EnvoyConfigObjectMatch_Listener{
							Listener: &networkingv1alpha3.EnvoyFilter_ListenerMatch{
								FilterChain: &networkingv1alpha3.EnvoyFilter_ListenerMatch_FilterChainMatch{
									Sni: external.Spec.ExternalDomain,
									Filter: &networkingv1alpha3.EnvoyFilter_ListenerMatch_FilterMatch{
										Name: "envoy.filters.network.http_connection_manager",
										SubFilter: &networkingv1alpha3.EnvoyFilter_ListenerMatch_SubFilterMatch{
											Name: "envoy.filters.http.router",
										},
									},
								},
							},
						},
					},
					Patch: patch,
				},
			},
		}

		ctrl.Log.Info("Configured certificate validation EnvoyFilter with X509 field parsing", "name", filterName)
		return nil
	})

	return err
}

// DeleteCertValidationFilter deletes the certificate validation EnvoyFilter
func DeleteCertValidationFilter(ctx context.Context, k8sClient client.Client, gatewayName string) error {
	filterName := fmt.Sprintf("%s-cert-validation", gatewayName)

	ctrl.Log.Info("Deleting certificate validation EnvoyFilter if it exists", "name", filterName, "namespace", istioSystemNamespace)

	envoyFilter := &istiov1alpha3.EnvoyFilter{
		ObjectMeta: metav1.ObjectMeta{
			Name:      filterName,
			Namespace: istioSystemNamespace,
		},
	}

	err := k8sClient.Delete(ctx, envoyFilter)
	if err != nil && !k8serrors.IsNotFound(err) {
		return fmt.Errorf("failed to delete certificate validation EnvoyFilter %s/%s: %w", istioSystemNamespace, filterName, err)
	}

	if k8serrors.IsNotFound(err) {
		ctrl.Log.Info("Skipped deletion of certificate validation EnvoyFilter as it wasn't present", "name", filterName)
	} else {
		ctrl.Log.Info("Successfully deleted certificate validation EnvoyFilter", "name", filterName)
	}

	return nil
}

// buildValidationLuaScript generates Lua script that performs strict X509 certificate validation
// by comparing full certificate subject strings (already reversed to match Envoy's format)
func buildValidationLuaScript(certSubjects []string) string {
	// Build Lua array of expected certificate subject strings
	var luaVars strings.Builder
	luaVars.WriteString("local expectedSubjects = {\n")

	for i, subject := range certSubjects {
		if i > 0 {
			luaVars.WriteString(",\n")
		}
		fmt.Fprintf(&luaVars, "  \"%s\"", escapeString(subject))
	}
	luaVars.WriteString("\n}\n\n")

	return luaVars.String() + generateStrictValidationLua()
}

// escapeString escapes special characters for Lua string literals
func escapeString(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, "\n", `\n`)
	s = strings.ReplaceAll(s, "\r", `\r`)
	return s
}

// generateStrictValidationLua returns the Lua script for strict X.509 subject validation
// Compares full certificate subject strings for exact match
func generateStrictValidationLua() string {
	return `function fail(request_handle, reason)
  request_handle:logErr(reason)
  request_handle:respond({[":status"] = "403"}, "Forbidden")
end

function envoy_on_request(request_handle)
  local ssl = request_handle:streamInfo():downstreamSslConnection()
  if ssl == nil then
    fail(request_handle, "No TLS connection")
    return
  end

  local subjectPeerCert = ssl:subjectPeerCertificate()
  if subjectPeerCert == nil then
    fail(request_handle, "No subject peer certificate")
    return
  end

  request_handle:logInfo(string.format("Certificate subject: %s", subjectPeerCert))

  -- Try to match against any expected subject (exact string comparison)
  local matchFound = false

  for _, expectedSubject in ipairs(expectedSubjects) do
    if subjectPeerCert == expectedSubject then
      matchFound = true
      request_handle:logInfo("Certificate subject validated successfully")
      break
    end
  end

  if not matchFound then
    fail(request_handle, string.format("Certificate subject validation failed: subject (%s) not in allowed list", subjectPeerCert))
    return
  end
end
`
}
