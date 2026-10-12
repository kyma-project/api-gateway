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

Add a `Conditions []metav1.Condition` field to `APIRuleStatus` in `apis/gateway/v2/apirule_types.go`, alongside the existing `State` and `Description` fields. The existing fields are preserved for backwards compatibility.

Introduce a single top-level condition type `Ready` with a fixed set of reasons. The v2 APIRule reconciliation manages three Istio sub-resource types — VirtualService, RequestAuthentication, and AuthorizationPolicy. Because sub-resource processors do not run independently and cannot be individually waited on, per-sub-resource condition types are not introduced. Instead, the failing sub-resource type is expressed as a distinct `Reason` on the single `Ready` condition. This gives consumers machine-readable failure attribution without the overhead of a multi-condition model.

The failing sub-resource type is not currently available at the point status is written. `processing.Reconcile` builds a `map[ResourceSelector][]error` keyed by sub-resource, but `GetStatusForErrorMap` immediately collapses it into a flattened `Description` string, and `ReconciliationV2alpha1Status` carries only `State` and `Description` back to the controller. Deriving the sub-resource reasons therefore requires threading the `ResourceSelector` keys through the status pipeline. This is a change to the status-pipeline signature, not only to the type definitions.

The alternative of introducing per-sub-resource condition types (one per processor) was considered and rejected. Sub-resources are reconciled sequentially in a single pipeline pass; if VirtualService fails, RequestAuthentication and AuthorizationPolicy do not run. Exposing per-resource condition types would leave those conditions stale from a previous reconcile pass, which is misleading. A single `Ready` condition with a discriminating `Reason` avoids this problem.

### Condition table

| Condition type | Status | Reason | Trigger |
|---|---|---|---|
| Ready | True | ReconcileSucceeded | All sub-resources reconciled successfully |
| Ready | False | VirtualServiceReconcileFailed | VirtualService create/update/delete failed |
| Ready | False | RequestAuthenticationReconcileFailed | RequestAuthentication create/update/delete failed |
| Ready | False | AuthorizationPolicyReconcileFailed | AuthorizationPolicy create/update/delete failed |
| Ready | False | ReconcileFailed | Safety net for errors outside the processor pipeline |
| Ready | False | ValidationFailed | APIRule spec failed validation |
| Ready | False | GatewayNotFound | Referenced Gateway or ExternalGateway not found |
| Ready | False | DependenciesMissing | Required CRDs not present |

The reconciler does not write an upfront in-progress condition before starting work. Before the first reconcile completes, the `Conditions` slice is empty; consumers should treat the absence of a `Ready` condition the same as `Ready=Unknown`.

The `ResourceSelector` enum also includes `OnAccessRule` (ORY Rule), but that resource is reconciled only on the v1beta1 path, which is out of scope; it has no reason on the v2 `Ready` condition.

`Warning` is part of the `State` enum but is not currently produced by the v2 reconciler; no condition reason is defined for it.

`Processing` and `Deleting` are transient `State` values. Conditions are not written during finalizer-driven deletion to avoid confusing status updates on a resource being removed.

### Writing the condition

The `Ready` condition is written at every reconcile exit point using `apimeta.SetStatusCondition`, which only updates `LastTransitionTime` when `Status` changes. Each call passes the generation of the v2alpha1 (hub) APIRule being reconciled as `ObservedGeneration`, since that is the object the conditions are written to.

The `Message` field of each condition carries the same error text as `Description`. For `Ready=True`, the message is `"Reconciled successfully"`.

`GatewayNotFound` is set by the existing `discoverGateway` pre-flight check, which runs before the processor pipeline and already writes `State=Error` to the CR when the referenced Gateway or ExternalGateway cannot be found. The new condition is written at those same exit points.

The `State` and `Description` fields continue to be written as they are today. The condition is written in addition, in the same status update call.

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

Consumers can `kubectl wait --for=condition=Ready=True apirule/<name>` without polling the `State` field. The reason on a `Ready=False` condition pinpoints the failure class (validation, gateway discovery, sub-resource reconciliation, missing dependencies) without parsing `Description`. The `State` and `Description` fields remain and existing consumers are unaffected.

The `APIRuleStatus` struct grows a `Conditions` field; the CRD schema and documentation must be updated. The CRD documentation must describe all condition types, statuses, reasons, and the `+listType=map` merge key. The v2alpha1 hub struct also gains the field; no change to conversion code is required because the JSON round-trip carries the field automatically.
