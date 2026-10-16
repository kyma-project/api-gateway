# RateLimit CR Conditions

## Status

Proposed

## Context

The RateLimit CR currently reports its reconciliation outcome through a flat `State` string field (`Ready`, `Warning`, `Error`) and a free-text `Description` field on `RateLimitStatus`. There are no structured Kubernetes conditions on the CR.

This has the same class of drawbacks as the APIRule CR before ADR [0014-apirule-conditions](0014-apirule-conditions.md):

- Consumers cannot `kubectl wait --for=condition=...` on any RateLimit state; the only machine-readable field is the `State` string, which requires polling and string comparison.
- When `State=Error` or `State=Warning`, the cause is in `Description` as a prose string. Consumers must parse human-readable text to distinguish a missing dependency from a validation failure from an EnvoyFilter apply error.
- There is no in-progress signal, and conditions do not add one. The `State` string has no `Processing` value; the reconciler has no upfront status write, so the first observable condition is the terminal outcome of a reconcile.

The RateLimit reconciler (`ratelimit_controller.go`) has a linear path with four distinct failure classes:

1. **Dependency check** — required CRDs not present (`dependencies.RateLimit().AreAvailable`). Sets `State=Error`.
2. **APIGateway CR check** — no APIGateway CR exists in the cluster, or the existing one is not in `Ready` state. Sets `State=Warning`.
3. **Validation** — `ratelimit.Validate` rejects the spec. Sets `State=Error`.
4. **EnvoyFilter reconciliation** — create or update of the single managed EnvoyFilter fails. Sets `State=Error`.

All four currently call the same `rl.Status.Error` or `rl.Status.Warning` helper, making them indistinguishable to a machine consumer.

## Decision

Add a `Conditions []metav1.Condition` field to `RateLimitStatus` in `apis/gateway/ratelimit/v1alpha1/ratelimit_types.go`, alongside the existing `State` and `Description` fields.

Introduce a single top-level condition type `Ready` with a fixed set of reasons covering the four failure classes above. The RateLimit reconciler manages a single sub-resource (EnvoyFilter); there are no independently observable subsystems, so a single condition type with distinct reasons is sufficient.

### Condition table

| Condition type | Status | Reason | Trigger |
|---|---|---|---|
| Ready | True | ReconcileSucceeded | EnvoyFilter reconciled successfully |
| Ready | False | DependenciesMissing | Required CRDs not present |
| Ready | False | APIGatewayNotReady | No APIGateway CR in the cluster, or it is not in Ready state |
| Ready | False | ValidationFailed | RateLimit spec failed validation |
| Ready | False | EnvoyFilterReconcileFailed | EnvoyFilter create or update failed |
| Ready | False | ReconcileFailed | Catch-all for unexpected errors not covered by the named reasons above (for example, status update failure) |

The reconciler does not write an upfront in-progress condition before starting work. Before the first reconcile completes, the `Conditions` slice is empty; consumers should treat the absence of a `Ready` condition the same as `Ready=Unknown`.

The existing `State=Warning` is set for the `APIGatewayNotReady` paths. The condition maps this to `Ready=False` rather than preserving the `Warning` semantic. A missing or not-ready APIGateway means the RateLimit cannot be reconciled at all, which is a hard not-ready state from the consumer's perspective regardless of the transient cause. This divergence is intentional; `Ready=False` gives consumers a consistent binary signal without introducing a `Ready=Warning` status that Kubernetes conditions do not define.

### Writing the condition

The `Ready` condition is written at every reconcile exit point using `apimeta.SetStatusCondition`, which only updates `LastTransitionTime` when `Status` changes. Each call passes the current `rateLimit.Generation` as `ObservedGeneration`.

The `Message` field of each condition carries the same error text as `Description`. For `Ready=True`, the message is `"Reconciled successfully"`.

The `State` and `Description` fields continue to be written as they are today via the existing `Error`, `Warning`, and `Ready` helpers on `RateLimitStatus`. The condition is written in addition to those helpers in the same status update call. No existing consumers are broken.

### Kubebuilder markers

The `Conditions` field requires the following markers for server-side apply and strategic merge patch correctness:

```go
// +listType=map
// +listMapKey=type
Conditions []metav1.Condition `json:"conditions,omitempty"`
```

### Condition helpers

Condition construction follows the pattern already established for the `ExternalGateway` CR (`apis/gateway/external/v1alpha1/conditions.go`): typed reason constants and factory helpers, so condition values are not hand-built at each reconcile exit point.

## Consequences

Consumers can `kubectl wait --for=condition=Ready=True ratelimit/<name>` without polling `State`. The reason on a `Ready=False` condition unambiguously identifies whether the problem is a missing dependency, an unready APIGateway, a spec validation error, or an EnvoyFilter apply failure — without parsing `Description`. The `State` and `Description` fields remain and existing consumers are unaffected.

The `RateLimitStatus` struct grows a `Conditions` field; the CRD schema and documentation must be updated to reflect the new condition type, statuses, and reasons, including the `+listType=map` merge key.
