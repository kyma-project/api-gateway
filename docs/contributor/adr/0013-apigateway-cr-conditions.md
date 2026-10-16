# APIGateway CR Conditions

## Status

Proposed

## Context

ADR [0004-api-gateway-cr-status-improvements](0004-api-gateway-cr-status-improvements.md) introduced a single `Ready` condition type on the APIGateway CR, using multiple reasons to describe what happened. This is insufficient for users to understand the state of individual subsystems (Kyma Gateway, Oathkeeper, NetworkPolicy, dependencies).

When a consumer observes `Ready=False`, they cannot tell which subsystem failed without inspecting `Reason` and parsing `Message`. There is no way to `kubectl wait` on a specific subsystem's readiness. The `controller.Status` interface carries a single condition pointer, so per-subsystem detail is lost by the time the main reconcile loop writes to the CR: each subsystem reconciler can return at most one condition, and only the last one written survives on the CR.

Additionally, `ErrCertificatePending` uses `KymaGatewayReconcileFailed` as its reason, making a normal async wait indistinguishable from a hard failure for consumers keying on `Reason`.

## Decision

Introduce dedicated condition types alongside the existing top-level `Ready`:

| Condition type | Subsystem              |
|---|------------------------|
| `Ready` | Top-level condition    |
| `KymaGatewayReady` | Kyma Gateway reconciliation |
| `CertificateReady` | Gardener certificate issuance |
| `DNSEntryReady` | Gardener DNS entry provisioning |
| `NetworkPolicyReady` | NetworkPolicy reconciliation |
| `DependenciesReady` | Module dependency check |

Oathkeeper does not receive a dedicated condition type because it is slated for deprecation; investing in a new condition surface for a subsystem on its way out is not worthwhile. Its reconciliation outcome continues to be reported as reasons on the top-level `Ready` condition. `OathkeeperReconcileSucceeded` and `OathkeeperReconcileDisabled` both return a ready status, so the reconcile loop continues and the final `Ready=True/ReconcileSucceeded` write always follows and overwrites them — the constants remain in the codebase but their values are never observable on the CR. Oathkeeper success and disabled state are implied by `Ready=True`; only `OathkeeperReconcileFailed` is observable, as it causes an early exit before `ReconcileSucceeded` is written.

Disabled subsystems are represented explicitly rather than reported as ready. When spec.enableKymaGateway=false or spec.networkPoliciesEnabled=false, the overall APIGateway reconcile can still succeed and write Ready=True/ReconcileSucceeded, but the corresponding subsystem condition must be written as non-ready with a disabled-specific reason and message. Concretely, KymaGatewayReady uses False/KymaGatewayReconcileDisabled, and NetworkPolicyReady uses False/NetworkPolicyReconcileDisabled. This keeps the meaning of each condition name literal: ...Ready=True is reserved for the case where the subsystem is enabled and reconciled successfully.

The upfront `Processing` write remains conditional, as established by ADR [0011-module-status](0011-module-status.md): it is only set when the spec has changed or no `Ready` condition exists yet. When it fires, the top-level `Ready` condition is set to `Unknown` with reason `ReconcileProcessing`. Subsystem-specific conditions are preserved from the last completed reconcile pass until the corresponding subsystem is reconciled again and its condition is updated. This signals that overall reconciliation is in progress without clearing the last known state of individual subsystems. This write is independent of the main reconciliation path and is not part of the accumulated condition slice. This differs from `ExternalGateway`, which writes its full `ProcessingConditions` set — every subsystem condition reset to `Unknown` — unconditionally at the start of every reconcile; APIGateway instead keeps each subsystem's last known state and uses only the aggregate `Ready` condition to represent in-progress reconciliation.

The `Ready` condition type serves a dual role. For subsystem failures without a dedicated condition type (Oathkeeper), it carries the specific failure reason. For non-subsystem exits it carries aggregate reasons: `ReconcileSucceeded` on the happy path, `ReconcileFailed` for hard failures not attributable to a specific reason, and dedicated reasons for early-exit paths (`OlderCRExists`, `DeletionBlockedExistingResources`). Early-exit paths write only the `Ready` condition immediately and return, since there are no subsystem conditions to accumulate. `DeletionBlockedExistingResources` and `KymaGatewayDeletionBlocked` are therefore mutually exclusive: the former fires in `reconcileFinalizer` and causes the reconciler to return before `ReconcileKymaGateway` is ever called; the latter can only be set in a pass where `reconcileFinalizer` returned ready. Because only one condition per type is stored (upsert-by-type), the last write to `Ready` in a reconcile pass is the one that persists — on the happy path this is always `ReconcileSucceeded`.

`UpdateApiGatewayStatus` already merges new conditions into the existing set using upsert-by-type semantics (`meta.SetStatusCondition` from `k8s.io/apimachinery`) rather than replacing `Status.Conditions` wholesale. This decision relies on that behavior: accumulating per-subsystem conditions across a reconcile pass and preserving the conditions of subsystems that did not run only works because the write upserts rather than replaces.

The `controller.Status` interface will be changed to carry a slice of conditions rather than a single pointer. This serves two purposes: subsystem conditions can be accumulated across the main reconciliation path, and a single subsystem reconciler can return more than one condition — for example `ReconcileKymaGateway` will produce both a `KymaGatewayReady` condition and a `CertificateReady` condition, which the current single-pointer interface cannot express. The existing `Condition() *metav1.Condition` method will be removed. Only the APIGateway controller calls it today, so the update is confined to that package and is a mechanical swap from a single pointer to a slice.

### Condition table

| Condition type | Status | Reason | Trigger |
|---|---|---|---|
| Ready | True | ReconcileSucceeded | All subsystems reconciled successfully |
| Ready | Unknown | ReconcileProcessing | Spec changed or initial install |
| Ready | False | ReconcileFailed | Hard failure with no more specific reason |
| Ready | False | OlderCRExists | This CR is not the oldest in the cluster |
| Ready | False | DeletionBlockedExistingResources | APIRules/ORY Rules/RateLimits block deletion |
| Ready | False | OathkeeperReconcileFailed | Ory Oathkeeper reconciliation failed |
| KymaGatewayReady | True | KymaGatewayReconcileSucceeded | Gateway reconciled successfully |
| KymaGatewayReady | False | KymaGatewayReconcileDisabled | Gateway reconciliation skipped because Kyma Gateway is disabled |
| KymaGatewayReady | Unknown | KymaGatewayReconcileProcessing | Certificate still being issued on initial install (no prior `KymaGatewayReady=True` exists) |
| KymaGatewayReady | False | KymaGatewayReconcileFailed | Gateway reconciliation failed; or certificate pending during renewal (prior `KymaGatewayReady=True` existed — degradation must be visible) |
| KymaGatewayReady | False | KymaGatewayDeletionBlocked | Custom resources block gateway deletion |
| CertificateReady | True | CertificateReconcileSucceeded | Certificate issued |
| CertificateReady | Unknown | CertificateReconcilePending | Certificate applied, not yet issued |
| CertificateReady | False | CertificateReconcileFailed | Certificate reconciliation failed |
| DNSEntryReady | True | DNSEntryReconcileSucceeded | DNS entry provisioned |
| DNSEntryReady | Unknown | DNSEntryReconcilePending | DNS entry applied, not yet provisioned |
| DNSEntryReady | False | DNSEntryReconcileFailed | DNS entry reconciliation failed |
| NetworkPolicyReady | True | NetworkPolicyReconcileSucceeded | NetworkPolicy reconciled successfully |
| NetworkPolicyReady | False | NetworkPolicyReconcileDisabled | NetworkPolicy reconciliation skipped because the subsystem is disabled |
| NetworkPolicyReady | False | NetworkPolicyReconcileFailed | NetworkPolicy reconciliation failed |
| DependenciesReady | True | DependenciesReconcileSucceeded | All required CRDs present |
| DependenciesReady | False | DependenciesMissing | Required CRDs not present |
| DependenciesReady | False | DependenciesReconcileFailed | Error during dependency discovery |

`CertificateReady` and `DNSEntryReady` are only written on Gardener clusters where certificate and DNS entry provisioning is managed. On non-Gardener clusters these condition types are absent from the CR. This follows the same pattern as `ExternalGateway`, which also omits these conditions when Gardener is unavailable.

## Consequences

Consumers can observe the state of individual subsystems without parsing `Message` strings. `kubectl wait --for=condition=KymaGatewayReady` and similar per-subsystem waits work out of the box. Certificate-pending on initial install is distinguishable from a hard failure by reason alone (`KymaGatewayReconcileProcessing` vs `KymaGatewayReconcileFailed`). Certificate-pending during renewal when the gateway was previously Ready is surfaced as `KymaGatewayReconcileFailed` to make the degradation visible rather than hiding it as a normal processing state. When Kyma Gateway or NetworkPolicy is intentionally disabled, `Ready=True` continues to show that the overall APIGateway reconcile succeeded while `KymaGatewayReady=False/KymaGatewayReconcileDisabled` and `NetworkPolicyReady=False/NetworkPolicyReconcileDisabled` make it explicit that those subsystems are not active by design.

The `controller.Status` interface is used by both the APIGateway and APIRule reconcilers, but only the APIGateway path calls `Condition()`. APIRule callers use `V1beta1Status()`, `V2alpha1Status()`, and `IsReady()`, which are unaffected by this change. Conditions for APIRule and RateLimit are out of scope here and will be addressed in a follow-up ADR. The CR status surface grows with the new condition types; existing consumers reading only `Ready` are unaffected, but CRD documentation must be updated. ADR [0004](0004-api-gateway-cr-status-improvements.md) is superseded by this decision.
