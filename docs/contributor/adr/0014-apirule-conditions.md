# APIRule CR Conditions

## Status

Proposed

## Context

The APIRule CR (v2) currently reports its reconciliation outcome through a flat `State` enum field and a free-text `Description` field. There are no structured Kubernetes conditions on the CR.

This has several practical drawbacks:

- Consumers cannot `kubectl wait --for=condition=...` on any APIRule state; the only machine-readable field is the `State` enum, which requires polling and string comparison.
- When `State=Error` or `State=Warning`, the cause is buried in `Description`, which is a prose string aggregating errors from multiple sub-resources (VirtualService, RequestAuthentication, AuthorizationPolicy). Consumers must parse human-readable text to determine what failed.
- The reconciliation pipeline (`processing.Reconcile`) processes sub-resource types sequentially. The first type that produces an error terminates the loop and reports a combined error string. There is no way to distinguish a VirtualService reconciliation failure from an AuthorizationPolicy failure without reading the message.

This ADR addresses that gap for the v2 APIRule only; v1beta1 is deprecated and not in scope.

## Decision

Add a `Conditions []metav1.Condition` field to `APIRuleStatus` in `apis/gateway/v2/apirule_types.go`, alongside the existing `State` and `Description` fields. The existing fields are preserved for backwards compatibility.

Introduce a single top-level condition type `Ready` with a fixed set of reasons. The v2 APIRule reconciliation manages three Istio sub-resource types — VirtualService, RequestAuthentication, and AuthorizationPolicy. Because sub-resource processors do not run independently and cannot be individually waited on, per-sub-resource condition types are not introduced. Instead, the failing sub-resource type is expressed as a distinct `Reason` on the single `Ready` condition. This gives consumers machine-readable failure attribution without the overhead of a multi-condition model.

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

### Writing the condition

The `Ready` condition is written at every reconcile exit point and only updates `LastTransitionTime` when `Status` changes.

`GatewayNotFound` is set by the existing `discoverGateway` pre-flight check, which runs before the processor pipeline and already writes `State=Error` to the CR when the referenced Gateway or ExternalGateway cannot be found. The new condition is written at those same exit points.

The `State` and `Description` fields continue to be written as they are.

### ObservedGeneration

Each condition's `ObservedGeneration` is set to `apiRule.Generation` on every reconcile exit point.

### Scope

This decision covers the v2 APIRule only. v1beta1 is deprecated; its `APIRuleResourceStatus` is not changed.

Both the v2 and v2alpha1 (hub) `APIRuleStatus` types receive the `Conditions` field. The `ConvertTo`/`ConvertFrom` conversion functions between the two versions copy the slice so conditions are not lost across version boundaries.

## Consequences

Consumers can `kubectl wait --for=condition=Ready=True apirule/<name>` without polling the `State` field. The reason on a `Ready=False` condition pinpoints the failure class (validation, gateway discovery, sub-resource reconciliation, missing dependencies) without parsing `Description`. The `State` and `Description` fields remain and existing consumers are unaffected.

The `APIRuleStatus` struct grows a `Conditions` field; the CRD schema and documentation must be updated. The v2alpha1 conversion path must propagate conditions across hub-version boundaries so that conditions are not lost on version conversion. Documentation for the APIRule CRD must reflect the new condition types, statuses, and reasons.
