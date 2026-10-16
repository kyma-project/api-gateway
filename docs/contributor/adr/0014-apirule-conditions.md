# APIRule CR Conditions

## Status

Proposed

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
| `SubresourcesConfigured` | All sub-resource processors completed without error |
| `DependenciesReady` | Required CRDs are present |
| `Ready` | Aggregated: all of the above are `True` |

### Condition table

| Condition type | Status | Reason | Trigger |
|---|---|---|---|
| `VirtualServiceConfigured` | `True` | `VirtualServiceReconcileSucceeded` | VirtualService reconciled successfully |
| `VirtualServiceConfigured` | `False` | `VirtualServiceReconcileFailed` | VirtualService create/update/delete failed |
| `VirtualServiceConfigured` | `Unknown` | `VirtualServiceNotYetReady` | Not yet evaluated this pass (early exit from prior processor) |
| `RequestAuthenticationConfigured` | `True` | `RequestAuthenticationReconcileSucceeded` | RequestAuthentication reconciled successfully |
| `RequestAuthenticationConfigured` | `False` | `RequestAuthenticationReconcileFailed` | RequestAuthentication create/update/delete failed |
| `RequestAuthenticationConfigured` | `Unknown` | `RequestAuthenticationNotYetReady` | Not yet evaluated this pass |
| `AuthorizationPolicyConfigured` | `True` | `AuthorizationPolicyReconcileSucceeded` | AuthorizationPolicy reconciled successfully |
| `AuthorizationPolicyConfigured` | `False` | `AuthorizationPolicyReconcileFailed` | AuthorizationPolicy create/update/delete failed |
| `AuthorizationPolicyConfigured` | `Unknown` | `AuthorizationPolicyNotYetReady` | Not yet evaluated this pass |
| `SubresourcesConfigured` | `True` | `SubresourcesReconcileSucceeded` | All sub-resource processors succeeded |
| `SubresourcesConfigured` | `False` | `SubresourcesReconcileFailed` | One or more sub-resource processors failed |
| `SubresourcesConfigured` | `Unknown` | `SubresourcesNotYetReady` | Not yet evaluated this pass |
| `DependenciesReady` | `True` | `DependenciesReconcileSucceeded` | All required CRDs are present |
| `DependenciesReady` | `False` | `DependenciesMissing` | A required CRD is absent |
| `DependenciesReady` | `False` | `DependenciesError` | Error while checking CRD availability |
| `Ready` | `True` | `ReconcileSucceeded` | All sub-resources reconciled successfully |
| `Ready` | `False` | `ReconcileFailed` | Safety net for errors outside the processor pipeline |
| `Ready` | `False` | `ValidationFailed` | APIRule spec failed validation |
| `Ready` | `False` | `GatewayNotFound` | Referenced Gateway or ExternalGateway not found |
| `Ready` | `False` | `DependenciesMissing` | Required CRDs not present |

The `ResourceSelector` enum also includes `OnAccessRule` (ORY Rule), but that resource is reconciled only on the v1beta1 path, which is out of scope; it has no condition type in the v2 model.

`Warning` is part of the `State` enum but is not currently produced by the v2 reconciler; no condition reason is defined for it.

`Processing` and `Deleting` are transient `State` values. Conditions are not written during finalizer-driven deletion to avoid confusing status updates on a resource being removed.

### Stale-condition handling

The v2 processor pipeline is fail-fast: if VirtualService fails, RequestAuthentication and AuthorizationPolicy do not run. Without mitigation, their conditions would carry stale `True` values from the previous reconcile pass.

This is handled by following the same upfront reset pattern used by `ExternalGateway`: at the start of each reconcile, all per-resource conditions are written to `Unknown` before any processor runs. If a processor is skipped due to an early exit, its condition remains `Unknown` for that pass rather than carrying a stale value.

### Writing conditions

Conditions are written at every reconcile exit point using `apimeta.SetStatusCondition`, which only updates `LastTransitionTime` when `Status` changes. The generation of the v2alpha1 (hub) APIRule is passed as `ObservedGeneration` on every condition.

The `Message` field of each condition carries the same error text as `Description`. For `Ready=True`, the message is `"Reconciled successfully"`.

`GatewayNotFound` is set by the existing `discoverGateway` pre-flight check, which runs before the processor pipeline and already writes `State=Error` to the CR when the referenced Gateway or ExternalGateway cannot be found. The new condition is written at those same exit points.

`State` and `Description` continue to be written unchanged. Conditions are written in addition, in the same status update call.

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

Consumers can `kubectl wait --for=condition=Ready=True apirule/<name>` without polling the `State` field. Per-resource condition types (`VirtualServiceConfigured`, etc.) allow consumers to identify which sub-resource failed without parsing `Description`. The `State` and `Description` fields remain and existing consumers are unaffected.

The `APIRuleStatus` struct grows a `Conditions` field; the CRD schema and documentation must be updated. The CRD documentation must describe all condition types, statuses, reasons, and the `+listType=map` merge key. The v2alpha1 hub struct also gains the field; no change to conversion code is required because the JSON round-trip carries the field automatically.
