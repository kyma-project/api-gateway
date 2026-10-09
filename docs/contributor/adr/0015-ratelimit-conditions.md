# RateLimit CR Conditions

## Status

Proposed

## Context

The RateLimit CR currently reports its reconciliation outcome through a flat `State` string field (`Ready`, `Warning`, `Error`) and a free-text `Description` field on `RateLimitStatus`. There are no structured Kubernetes conditions on the CR.

This has the same class of drawbacks as the APIRule CR before ADR [0013-apirule-conditions](0014-apirule-conditions.md):

- Consumers cannot `kubectl wait --for=condition=...` on any RateLimit state; the only machine-readable field is the `State` string, which requires polling and string comparison.
- When `State=Error` or `State=Warning`, the cause is in `Description` as a prose string. Consumers must parse human-readable text to distinguish a missing dependency from a validation failure from an EnvoyFilter apply error.
- There is no in-progress signal. The `State` enum has no `Processing` value; conditions intentionally omit an in-progress entry for the same reason.

The RateLimit reconciler (`ratelimit_controller.go`) has a linear path with four distinct failure classes:

1. **Dependency check** — required CRDs not present (`dependencies.RateLimit().AreAvailable`).
2. **APIGateway CR check** — no APIGateway CR exists in the cluster, or the existing one is not in `Ready` state.
3. **Validation** — `ratelimit.Validate` rejects the spec.
4. **EnvoyFilter reconciliation** — create or update of the single managed EnvoyFilter fails.

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

### Writing the condition

The `Ready` condition is written at every reconcile exit point and only updates `LastTransitionTime` when `Status` changes.

The `State` and `Description` fields continue to be written as they are today via the existing `Error`, `Warning`, and `Ready` helpers on `RateLimitStatus`. The condition is written in addition to those helpers in the same status update call. No existing consumers are broken.

### ObservedGeneration

Each condition's `ObservedGeneration` is set to `rateLimit.Generation` on every reconcile exit point.

## Consequences

Consumers can `kubectl wait --for=condition=Ready=True ratelimit/<name>` without polling `State`. The reason on a `Ready=False` condition unambiguously identifies whether the problem is a missing dependency, an unready APIGateway, a spec validation error, or an EnvoyFilter apply failure — without parsing `Description`. The `State` and `Description` fields remain and existing consumers are unaffected.

The `RateLimitStatus` struct grows a `Conditions` field; the CRD schema and documentation must be updated to reflect the new condition type, statuses, and reasons.
