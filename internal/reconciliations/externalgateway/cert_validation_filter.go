package externalgateway

import (
	"context"
	"fmt"
	"strings"

	externalv1alpha1 "github.com/kyma-project/api-gateway/apis/gateway/external/v1alpha1"
	"google.golang.org/protobuf/types/known/structpb"
	networkingv1alpha3 "istio.io/api/networking/v1alpha3"
	istiov1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// ReconcileCertValidationFilter creates or updates the EnvoyFilter that validates client certificate X509 fields
func ReconcileCertValidationFilter(ctx context.Context, k8sClient client.Client, external *externalv1alpha1.ExternalGateway, certSubjects []RegionCertSubject) error {
	filterName := fmt.Sprintf("%s-cert-validation", external.Spec.Gateway)

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
		envoyFilter.Labels = map[string]string{
			"app.kubernetes.io/managed-by":        "externalgateway-controller",
			"app.kubernetes.io/created-for":       fmt.Sprintf("%s-%s", external.Namespace, external.Name),
			"gateway.kyma-project.io/external-id": external.Name,
		}

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

// buildValidationLuaScript generates Lua script that validates X509 fields individually
func buildValidationLuaScript(certSubjects []RegionCertSubject) string {
	// Group certificate subjects by region, collecting unique CN/L combinations with their allowed OUs
	type regionConfig struct {
		CN  string
		L   string
		OUs map[string]bool // Using map for deduplication
	}

	regions := make(map[string]*regionConfig)

	for _, cert := range certSubjects {
		if _, exists := regions[cert.Region]; !exists {
			regions[cert.Region] = &regionConfig{
				CN:  cert.CN,
				L:   cert.L,
				OUs: make(map[string]bool),
			}
		}
		// Add all OUs for this region
		for _, ou := range cert.OU {
			regions[cert.Region].OUs[ou] = true
		}
	}

	// Build Lua regions table
	var luaRegions strings.Builder
	luaRegions.WriteString("local regions = {\n")

	for regionKey, config := range regions {
		luaRegions.WriteString(fmt.Sprintf("  [\"%s\"] = {\n", escapeString(regionKey)))
		luaRegions.WriteString(fmt.Sprintf("    CN = \"%s\",\n", escapeString(config.CN)))
		luaRegions.WriteString(fmt.Sprintf("    L = \"%s\",\n", escapeString(config.L)))
		luaRegions.WriteString("    OU = {")

		first := true
		for ou := range config.OUs {
			if !first {
				luaRegions.WriteString(", ")
			}
			luaRegions.WriteString(fmt.Sprintf("\"%s\"", escapeString(ou)))
			first = false
		}
		luaRegions.WriteString("},\n")
		luaRegions.WriteString("  },\n")
	}
	luaRegions.WriteString("}\n\n")

	// Complete Lua script with X509 field parsing and validation
	return luaRegions.String() + `function fail(request_handle, reason)
  request_handle:logErr(reason)
  request_handle:respond({[":status"] = "403"}, reason)
end

function envoy_on_request(request_handle)
  -- Check TLS connection
  local ssl = request_handle:streamInfo():downstreamSslConnection()
  if ssl == nil then
    fail(request_handle, "No TLS connection")
    return
  end

  -- Get certificate subject
  local subject = ssl:subjectPeerCertificate()
  if subject == nil then
    fail(request_handle, "No subject peer cert")
    return
  end

  request_handle:logInfo(string.format("Certificate subject: %s", subject))

  -- Parse X509 fields from subject string
  local cn = subject:match("CN=([^,]+)")
  local l = subject:match("L=([^,]+)")

  if not cn or not l then
    fail(request_handle, "Certificate missing CN or L field")
    return
  end

  -- Trim whitespace
  cn = cn:match("^%s*(.-)%s*$")
  l = l:match("^%s*(.-)%s*$")

  -- Validate against allowed regions
  for region_key, expected in pairs(regions) do
    if cn == expected.CN and l == expected.L then
      -- Check if any OU in the certificate matches allowed OUs
      for _, allowed_ou in ipairs(expected.OU) do
        -- Use plain string search for OU field
        local ou_pattern = "OU=" .. allowed_ou:gsub("([%^%$%(%)%%%.%[%]%*%+%-%?])", "%%%1")
        if subject:find(ou_pattern, 1, true) then
          request_handle:logInfo(string.format("Certificate validated for region: %s", region_key))
          return  -- Success - allow request
        end
      end
    end
  end

  -- No match found - reject request
  fail(request_handle, string.format("Certificate validation failed: CN=%s, L=%s", cn, l))
end
`
}

// escapeString escapes special characters for Lua string literals
func escapeString(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, "\n", `\n`)
	s = strings.ReplaceAll(s, "\r", `\r`)
	return s
}
