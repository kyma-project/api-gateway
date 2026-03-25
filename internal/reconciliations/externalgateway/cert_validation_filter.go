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
func ReconcileCertValidationFilter(ctx context.Context, k8sClient client.Client, external *externalv1alpha1.ExternalGateway, certSubjects []RegionCertSubject) error {
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

// buildLuaTable generates a Lua table representation of a RegionCertSubject
// Example output: {C="US", O="Org", OU={"value1", "value2"}, L="gateway", CN="name"}
func buildLuaTable(cert RegionCertSubject) string {
	var parts []string

	if cert.C != "" {
		parts = append(parts, fmt.Sprintf("C=\"%s\"", escapeString(cert.C)))
	}
	if cert.O != "" {
		parts = append(parts, fmt.Sprintf("O=\"%s\"", escapeString(cert.O)))
	}
	if len(cert.OU) > 0 {
		var ouValues []string
		for _, ou := range cert.OU {
			ouValues = append(ouValues, fmt.Sprintf("\"%s\"", escapeString(ou)))
		}
		parts = append(parts, fmt.Sprintf("OU={%s}", strings.Join(ouValues, ", ")))
	}
	if cert.L != "" {
		parts = append(parts, fmt.Sprintf("L=\"%s\"", escapeString(cert.L)))
	}
	if cert.CN != "" {
		parts = append(parts, fmt.Sprintf("CN=\"%s\"", escapeString(cert.CN)))
	}

	return "{" + strings.Join(parts, ", ") + "}"
}

// buildValidationLuaScript generates Lua script that performs strict X509 certificate validation
// All parsing and structuring is done in Go; Lua only validates pre-structured data
func buildValidationLuaScript(certSubjects []RegionCertSubject) string {
	// Build Lua array of expected certificate structures
	var luaVars strings.Builder
	luaVars.WriteString("local expectedCerts = {\n")

	for i, cert := range certSubjects {
		if i > 0 {
			luaVars.WriteString(",\n")
		}
		fmt.Fprintf(&luaVars, "  %s", buildLuaTable(cert))
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
// Expects structured certificate data generated by Go (expectedCerts table)
func generateStrictValidationLua() string {
	return `-- Extract field value from certificate subject string
-- Example: extractField("CN=test, OU=org", "CN") returns "test"
function extractField(subject, field)
  local pattern = field .. "=([^,]+)"
  local value = subject:match(pattern)
  if value then
    return value:match("^%s*(.-)%s*$")  -- trim whitespace
  end
  return nil
end

-- Extract all occurrences of a field from certificate subject string
-- Example: extractAllFields("OU=org1, OU=org2", "OU") returns {"org1", "org2"}
function extractAllFields(subject, field)
  local values = {}
  local pattern = field .. "=([^,]+)"
  for value in subject:gmatch(pattern) do
    value = value:match("^%s*(.-)%s*$")  -- trim whitespace
    table.insert(values, value)
  end
  return values
end

-- Compare two arrays for exact equality (same values in same order)
function arraysEqual(arr1, arr2)
  if #arr1 ~= #arr2 then
    return false
  end
  for i = 1, #arr1 do
    if arr1[i] ~= arr2[i] then
      return false
    end
  end
  return true
end

-- Check if actual certificate subject matches expected certificate structure
function certMatches(subjectPeerCert, expectedCert)
  -- Extract fields from actual certificate
  local actualC = extractField(subjectPeerCert, "C")
  local actualO = extractField(subjectPeerCert, "O")
  local actualL = extractField(subjectPeerCert, "L")
  local actualCN = extractField(subjectPeerCert, "CN")
  local actualOU = extractAllFields(subjectPeerCert, "OU")

  -- Validate C (Country)
  if expectedCert.C and actualC ~= expectedCert.C then
    return false, string.format("C mismatch: expected '%s', got '%s'", expectedCert.C, actualC or "nil")
  end

  -- Validate O (Organization)
  if expectedCert.O and actualO ~= expectedCert.O then
    return false, string.format("O mismatch: expected '%s', got '%s'", expectedCert.O, actualO or "nil")
  end

  -- Validate L (Locality)
  if expectedCert.L and actualL ~= expectedCert.L then
    return false, string.format("L mismatch: expected '%s', got '%s'", expectedCert.L, actualL or "nil")
  end

  -- Validate CN (Common Name)
  if expectedCert.CN and actualCN ~= expectedCert.CN then
    return false, string.format("CN mismatch: expected '%s', got '%s'", expectedCert.CN, actualCN or "nil")
  end

  -- Validate OU (Organizational Units) - must match exactly in order
  if expectedCert.OU then
    if not arraysEqual(actualOU, expectedCert.OU) then
      local expectedOUStr = table.concat(expectedCert.OU, ", ")
      local actualOUStr = table.concat(actualOU, ", ")
      return false, string.format("OU mismatch: expected [%s], got [%s]", expectedOUStr, actualOUStr)
    end
  end

  return true, "Match"
end

function fail(request_handle, reason)
  request_handle:logErr(reason)

  -- Return 403 Forbidden without reason in body (to avoid leaking certificate details) and log the reason for debugging
  request_handle:respond({[":status"] = "403"}, "Forbidden:  ")
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

  -- Try to match against any expected certificate (OR logic at certificate level)
  local matchFound = false
  local lastError = ""

  for _, expectedCert in ipairs(expectedCerts) do
    local matched, reason = certMatches(subjectPeerCert, expectedCert)

    if matched then
      matchFound = true
      request_handle:logInfo("Certificate subject validated successfully")
      break
    else
      lastError = reason
    end
  end

  if not matchFound then
    fail(request_handle, string.format("Certificate subject validation failed: %s", lastError))
    return
  end
end
`
}
