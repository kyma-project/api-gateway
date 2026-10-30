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

Introduce per-resource condition types for each Istio sub-resource managed by the v2 reconciler, plus a top-level `Ready` condition that aggregates them.

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
| `RequestAuthenticationConfigured` | `True` | `RequestAuthenticationReconcileSucceeded` | RequestAuthentication reconciled successfully |
| `RequestAuthenticationConfigured` | `False` | `RequestAuthenticationReconcileFailed` | RequestAuthentication create/update/delete failed |
| `AuthorizationPolicyConfigured` | `True` | `AuthorizationPolicyReconcileSucceeded` | AuthorizationPolicy reconciled successfully |
| `AuthorizationPolicyConfigured` | `False` | `AuthorizationPolicyReconcileFailed` | AuthorizationPolicy create/update/delete failed |
| `DependenciesReady` | `True` | `DependenciesReconcileSucceeded` | All required CRDs are present |
| `DependenciesReady` | `False` | `DependenciesMissing` | A required CRD is absent |
| `DependenciesReady` | `False` | `DependenciesReconcileFailed` | Error while checking CRD availability |
| `Ready` | `True` | `ReconcileSucceeded` | All sub-resources reconciled successfully |
| `Ready` | `False` | `ReconcileFailed` | Safety net for errors outside the processor pipeline |
| `Ready` | `False` | `ValidationFailed` | APIRule spec failed validation |
| `Ready` | `False` | `GatewayNotFound` | Referenced Gateway or ExternalGateway not found |

The `ResourceSelector` enum also includes `OnAccessRule` (ORY Rule), but that resource is reconciled only on the v1beta1 path, which is out of scope; it has no condition type in the v2 model.

`Warning` is part of the `State` enum but is not currently produced by the v2 reconciler; no condition reason is defined for it.

`Processing` and `Deleting` are transient `State` values. Conditions are not written during finalizer-driven deletion to avoid confusing status updates on a resource being removed.

### Stale-condition handling

The v2 processor pipeline is fail-fast: if VirtualService fails, RequestAuthentication and AuthorizationPolicy do not run. A sub-resource may also not be in play at all for a given APIRule — for example, an APIRule with no authentication produces no RequestAuthentication. In both cases a per-resource condition could otherwise carry a stale value from a previous pass.

This is handled by writing conditions only for what actually applies, rather than resetting everything upfront:

- A per-resource condition is written only when that sub-resource is actually managed for the APIRule on this pass (present in the effective spec and produced by the processor). A sub-resource that is not in play gets no condition at all — no placeholder, no `Processing` reason.
- A condition that is no longer applicable (present from a previous pass, stale this pass) is removed at the terminal status write. Upsert-by-type alone can never clear a condition, so removal is explicit.
- If the fail-fast pipeline exits early, sub-resources that never ran simply have no condition written this pass, and any stale condition from a prior pass is removed.

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

Consumers can `kubectl wait --for=condition=Ready=True apirule/<name>` without polling the `State` field. Per-resource condition types (`VirtualServiceConfigured`, etc.) allow consumers to identify which sub-resource failed without parsing `Description`. Per-resource conditions appear only for sub-resources in use; sub-resources not in play carry no condition. The `State` and `Description` fields remain and existing consumers are unaffected. Conditions are written in addition to the existing fields in the same status update call.

The `APIRuleStatus` struct grows a `Conditions` field; the CRD schema and documentation must be updated. The v2alpha1 hub struct also gains the field; no change to conversion code is required because the JSON round-trip carries the field automatically.
