# APIRule CR Conditions

## Status

Accepted

## Context

The APIRule CR (v2) currently reports its reconciliation outcome through a flat `State` enum field and a free-text `Description` field on `APIRuleStatus`. There are no structured Kubernetes conditions on the CR.

This has several practical drawbacks:

- Consumers cannot `kubectl wait --for=condition=...` on any APIRule state; the only machine-readable field is the `State` enum, which requires polling and string comparison.
- When `State=Error` or `State=Warning`, the cause is buried in `Description`, which is a prose string aggregating errors from multiple sub-resources (VirtualService, RequestAuthentication, AuthorizationPolicy). Consumers must parse human-readable text to determine what failed.
- The reconciliation pipeline (`processing.Reconcile`) processes sub-resource types sequentially. The first type that produces an error terminates the loop and reports a combined error string. There is no way to distinguish a VirtualService reconciliation failure from an AuthorizationPolicy failure without reading the message.

This ADR addresses that gap for the v2 APIRule only; v1beta1 is deprecated and not in scope.

## Decision

Add a `Conditions []metav1.Condition` field to `APIRuleStatus` in both `apis/gateway/v2/apirule_types.go` and `apis/gateway/v2alpha1/apirule_types.go`, alongside the existing `State` and `Description` fields. The existing fields are preserved for backwards compatibility.

### Condition model

Introduce per-resource condition types for each Istio sub-resource managed by the v2 reconciler, plus a top-level `Ready` condition that aggregates them. This follows the pattern established for the `ExternalGateway` CR.

| Condition type | Meaning |
|---|---|
| `VirtualServiceConfigured` | VirtualService create/update/delete succeeded |
| `RequestAuthenticationConfigured` | RequestAuthentication create/update/delete succeeded |
| `AuthorizationPolicyConfigured` | AuthorizationPolicy create/update/delete succeeded |
| `DependenciesReady` | Required CRDs are present |
| `Ready` | Aggregated: all of the above are `True` |

`GatewayNotFound` is a reason on `Ready` rather than a dedicated condition type because the gateway reference is a spec field — its absence is treated as a spec validation failure rather than an independent environmental dependency.

### Condition table

| Condition type | Status | Reason | Trigger |
|---|---|---|---|
| `VirtualServiceConfigured` | `True` | `VirtualServiceReconcileSucceeded` | VirtualService reconciled successfully |
| `VirtualServiceConfigured` | `False` | `VirtualServiceReconcileFailed` | VirtualService create/update/delete failed |
| `VirtualServiceConfigured` | `False` | `VirtualServiceReconcileProcessing` | Not yet evaluated this pass |
| `RequestAuthenticationConfigured` | `True` | `RequestAuthenticationReconcileSucceeded` | RequestAuthentication reconciled successfully |
| `RequestAuthenticationConfigured` | `False` | `RequestAuthenticationReconcileFailed` | RequestAuthentication create/update/delete failed |
| `RequestAuthenticationConfigured` | `False` | `RequestAuthenticationReconcileProcessing` | Not yet evaluated this pass |
| `AuthorizationPolicyConfigured` | `True` | `AuthorizationPolicyReconcileSucceeded` | AuthorizationPolicy reconciled successfully |
| `AuthorizationPolicyConfigured` | `False` | `AuthorizationPolicyReconcileFailed` | AuthorizationPolicy create/update/delete failed |
| `AuthorizationPolicyConfigured` | `False` | `AuthorizationPolicyReconcileProcessing` | Not yet evaluated this pass |
| `DependenciesReady` | `True` | `DependenciesReconcileSucceeded` | All required CRDs are present |
| `DependenciesReady` | `False` | `DependenciesMissing` | A required CRD is absent |
| `DependenciesReady` | `False` | `DependenciesReconcileFailed` | Error while checking CRD availability |
| `Ready` | `True` | `ReconcileSucceeded` | All sub-resources reconciled successfully |
| `Ready` | `False` | `ReconcileProcessing` | Reconcile in progress; sub-resources not yet evaluated this pass |
| `Ready` | `False` | `ReconcileFailed` | Safety net for errors outside the processor pipeline |
| `Ready` | `False` | `ValidationFailed` | APIRule spec failed validation |
| `Ready` | `False` | `GatewayNotFound` | Referenced Gateway or ExternalGateway not found |

The `ResourceSelector` enum also includes `OnAccessRule` (ORY Rule), but that resource is reconciled only on the v1beta1 path, which is out of scope; it has no condition type in the v2 model.

`Warning` is part of the `State` enum but is not currently produced by the v2 reconciler; no condition reason is defined for it.

`Processing` and `Deleting` are transient `State` values. Conditions are not written during finalizer-driven deletion to avoid confusing status updates on a resource being removed.

### Stale-condition handling

The v2 processor pipeline is fail-fast: if VirtualService fails, RequestAuthentication and AuthorizationPolicy do not run. Without mitigation, their conditions would carry stale `True` values from the previous reconcile pass.

This is handled by following the same upfront reset pattern used by `ExternalGateway`: at the start of each reconcile, every per-resource condition, along with the aggregate `Ready` condition, is reset to `False/...ReconcileProcessing` before any processor runs. If a processor is skipped due to an early exit, its condition remains `False/...ReconcileProcessing` for that pass rather than carrying a stale value.

This reset is an in-memory stale-guard: status is persisted only once per reconcile at the terminal write, so `False/...ReconcileProcessing` is observed only for sub-resources not evaluated on that pass, not as a transient in-progress signal.

`GatewayNotFound` is set by the existing `discoverGateway` pre-flight check, which runs before the processor pipeline and already writes `State=Error` to the CR when the referenced Gateway or ExternalGateway cannot be found. The new condition is written at those same exit points.

### Kubebuilder markers

The `Conditions` field requires the following markers for server-side apply and strategic merge patch correctness:

```go
// +listType=map
// +listMapKey=type
Conditions []metav1.Condition `json:"conditions,omitempty"`
```

### Condition helpers

Condition construction follows the pattern already established for the `ExternalGateway` CR (`apis/gateway/external/v1alpha1/conditions.go`): typed reason constants and factory helpers, so condition values are not hand-built at each reconcile exit point.

### Scope

This decision covers the v2 APIRule only. v1beta1 is deprecated; its `APIRuleResourceStatus` is not changed.

Both the v2 and v2alpha1 (hub) `APIRuleStatus` types receive the `Conditions` field with the same JSON tag (`json:"conditions,omitempty"`). The existing `ConvertTo`/`ConvertFrom` conversion between the two versions uses a JSON round-trip (`convertOverJson`); because the field name and JSON tag are identical in both structs, conditions are preserved across version boundaries without any change to the conversion functions.

## Consequences

Consumers can `kubectl wait --for=condition=Ready=True apirule/<name>` without polling the `State` field. Per-resource condition types (`VirtualServiceConfigured`, etc.) allow consumers to identify which sub-resource failed without parsing `Description`. The `State` and `Description` fields remain and existing consumers are unaffected. Conditions are written in addition to the existing fields in the same status update call.

The `APIRuleStatus` struct grows a `Conditions` field; the CRD schema and documentation must be updated. The CRD documentation must describe all condition types, statuses, reasons, and the `+listType=map` merge key. The v2alpha1 hub struct also gains the field; no change to conversion code is required because the JSON round-trip carries the field automatically.
