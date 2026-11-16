# APIGateway CR Conditions

## Status

Accepted

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
| `CertificateReady` | Kyma Gateway TLS certificate (Gardener-issued or self-signed) |
| `DNSReady` | Kyma Gateway DNS (Gardener DNS entry or local domain) |
| `NetworkPolicyReady` | NetworkPolicy reconciliation |
| `DependenciesReady` | Module dependency check |

Oathkeeper does not receive a dedicated condition type because it is slated for deprecation. Its outcome is reported as reasons on the top-level `Ready` condition: success and disabled state are implied by `Ready=True` (the final `ReconcileSucceeded` write overwrites them), so only `OathkeeperReconcileFailed` is observable, as it causes an early exit before `ReconcileSucceeded` is written.

A condition exists only when its subsystem is actually in play. When a subsystem is disabled — spec.enableKymaGateway=false or spec.networkPoliciesEnabled=false — no condition is written for it at all: there is no disabled-specific reason and no placeholder. The overall APIGateway reconcile can still succeed and write Ready=True/ReconcileSucceeded while the corresponding subsystem condition is simply absent. This keeps the meaning of each condition name literal: ...Ready=True is reserved for the case where the subsystem is enabled and reconciled successfully, and the absence of a subsystem condition means that subsystem is not active.

If a subsystem was enabled on a previous pass and is later disabled, its condition is removed from the status.

The Kyma Gateway certificate and DNS entry behave differently by cluster type. On a Gardener cluster the gateway relies on a Gardener-issued `Certificate` and `DNSEntry` CRs, whose issuance is async. The `CertificateReady`/`DNSReady` conditions carry `False/...ReconcilePending` only during initial install — while the first reconcile is waiting for Gardener to provision them and no prior condition value exists. Once provisioning completes they become `True/...ReconcileSucceeded`; `False/...ReconcileFailed` is written whenever Gardener reports an error state. On steady-state reconciles after the first success, a transient pending state does not flip the condition back to `False/...ReconcilePending`: the condition keeps its last value. On a non-Gardener cluster no Gardener resources are created: the gateway uses the local `local.kyma.dev` domain and a bundled self-signed certificate, both immediately functional. Both conditions are `True` with distinct reasons — `CertificateSelfSigned` and `DNSLocalDomain` — so the local fallback is not mistaken for a Gardener-provisioned one. The Gardener/non-Gardener distinction is made inside the condition evaluation, not as an upfront gate.

The upfront `Processing` write remains conditional, as established by ADR [0011-module-status](0011-module-status.md): it is only set when the spec has changed or no `Ready` condition exists yet (initial install). When it fires, the top-level `Ready` condition is set to `False` with reason `ReconcileProcessing`. Subsystem-specific conditions are preserved from the last completed reconcile pass until the corresponding subsystem is reconciled again and its condition is updated. This signals that overall reconciliation is in progress without clearing the last known state of individual subsystems. This write is independent of the main reconciliation path and is not part of the accumulated condition slice.

The `Ready` condition type serves a dual role. For subsystem failures without a dedicated condition type (Oathkeeper), it carries the specific failure reason. For non-subsystem exits it carries aggregate reasons: `ReconcileSucceeded` on the happy path, `ReconcileFailed` for hard failures not attributable to a specific reason, and dedicated reasons for early-exit paths (`OlderCRExists`, `DeletionBlockedExistingResources`). Early-exit paths write only the `Ready` condition immediately and return, since there are no subsystem conditions to accumulate. `DeletionBlockedExistingResources` and `KymaGatewayDeletionBlocked` are therefore mutually exclusive: the former fires in `reconcileFinalizer` and causes the reconciler to return before `ReconcileKymaGateway` is ever called; the latter can only be set in a pass where `reconcileFinalizer` returned ready. Because only one condition per type is stored (upsert-by-type), the last write to `Ready` in a reconcile pass is the one that persists — on the happy path this is always `ReconcileSucceeded`.

The status writer merges new conditions into the existing set using upsert-by-type semantics rather than replacing the full conditions list. This allows per-subsystem conditions to accumulate across a reconcile pass and preserves conditions of subsystems that did not run. The writer is extended to also remove owned condition types that are no longer in play, since upsert alone can never clear a stale condition.

The `controller.Status` interface will be changed to carry a slice of conditions rather than a single pointer. This serves two purposes: subsystem conditions can be accumulated across the main reconciliation path, and a single subsystem reconciler can return more than one condition — for example `ReconcileKymaGateway` will produce both a `KymaGatewayReady` condition and a `CertificateReady` condition, which the current single-pointer interface cannot express. The existing `Condition() *metav1.Condition` method will be removed. Only the APIGateway controller calls it today, so the update is confined to that package and is a mechanical swap from a single pointer to a slice.

### Condition table

The `DNSReady` condition type uses `DNSEntryReconcile*` reasons on the Gardener path because the async DNS resource is a Gardener `DNSEntry` CR; the type is broader to accommodate the non-Gardener path (`DNSLocalDomain`), where no `DNSEntry` CR exists.

| Condition type | Status | Reason | Trigger |
|---|---|---|---|
| Ready | True | ReconcileSucceeded | All subsystems reconciled successfully |
| Ready | False | ReconcileProcessing | Spec changed or initial install |
| Ready | False | ReconcileFailed | Hard failure with no more specific reason |
| Ready | False | OlderCRExists | This CR is not the oldest in the cluster |
| Ready | False | DeletionBlockedExistingResources | APIRules/ORY Rules/RateLimits block deletion |
| Ready | False | OathkeeperReconcileFailed | Ory Oathkeeper reconciliation failed |
| KymaGatewayReady | True | KymaGatewayReconcileSucceeded | Gateway reconciled successfully |
| KymaGatewayReady | False | KymaGatewayReconcileFailed | Gateway reconciliation failed |
| KymaGatewayReady | False | KymaGatewayDeletionBlocked | Custom resources block gateway deletion |
| CertificateReady | True | CertificateReconcileSucceeded | Gardener certificate issued |
| CertificateReady | True | CertificateSelfSigned | Non-Gardener cluster; self-signed certificate secret applied successfully |
| CertificateReady | False | CertificateReconcilePending | Certificate applied, not yet issued |
| CertificateReady | False | CertificateReconcileFailed | Certificate reconciliation failed |
| DNSReady | True | DNSEntryReconcileSucceeded | Gardener DNS entry provisioned |
| DNSReady | True | DNSLocalDomain | Non-Gardener cluster; local `local.kyma.dev` domain used |
| DNSReady | False | DNSEntryReconcilePending | DNS entry applied, not yet provisioned |
| DNSReady | False | DNSEntryReconcileFailed | DNS entry reconciliation failed |
| NetworkPolicyReady | True | NetworkPolicyReconcileSucceeded | NetworkPolicy reconciled successfully |
| NetworkPolicyReady | False | NetworkPolicyReconcileFailed | NetworkPolicy reconciliation failed |
| DependenciesReady | True | DependenciesReconcileSucceeded | All required CRDs present |
| DependenciesReady | False | DependenciesMissing | Required CRDs not present |
| DependenciesReady | False | DependenciesReconcileFailed | Error during dependency discovery |

### Gardener provisioning requeue

`ReconcileKymaGateway` does not block on Gardener certificate or DNS entry issuance — those are async dependencies managed outside the reconciler, so the full reconcile loop continues through all remaining subsystems (Oathkeeper, VPA). A pending certificate or DNS entry is not treated as an error: it sets `False/...ReconcilePending` on the corresponding condition, but does not cause the reconciler to return an error status. On initial install, while the certificate or DNS entry is still pending, the subsystem conditions carry `False/...ReconcilePending` and the top-level `Ready` condition stays `False/ReconcileProcessing`: the CR is reported as processing, not ready, until the first provisioning completes. Once both reach `True`, the aggregate `Ready` flips to `True/ReconcileSucceeded`. This applies only on Gardener clusters; on non-Gardener clusters the certificate and DNS entry are immediately ready (see Decision above), so no pending state ever arises.

When the reconciler detects a `CertificateReady=False/CertificateReconcilePending` or `DNSReady=False/DNSEntryReconcilePending` condition, it requeues after 15 seconds instead of the normal 1-hour interval. Once Gardener finishes, the next reconcile writes the condition as `True` and falls through to the normal interval.

This keeps the requeue decision at the controller level, based on observable condition state, without any special-casing inside `ReconcileKymaGateway` or new watches on Gardener CRs.

## Consequences

Consumers can observe the state of individual subsystems without parsing `Message` strings. `kubectl wait --for=condition=KymaGatewayReady` and similar per-subsystem waits work out of the box. When Kyma Gateway or NetworkPolicy is intentionally disabled, `Ready=True` continues to show that the overall APIGateway reconcile succeeded while the corresponding subsystem condition is simply absent. A subsystem condition left over from a pass where the subsystem was enabled is removed once the subsystem is disabled, so the status never carries a stale entry.

The `controller.Status` interface is used by both the APIGateway and APIRule reconcilers, but only the APIGateway path calls `Condition()`. APIRule callers use `V1beta1Status()`, `V2alpha1Status()`, and `IsReady()`, which are unaffected by this change. Conditions for APIRule and RateLimit are out of scope here and will be addressed in a follow-up ADR. The CR status surface grows with the new condition types; existing consumers reading only `Ready` are unaffected, but CRD documentation must be updated. ADR [0004](0004-api-gateway-cr-status-improvements.md) is superseded by this decision.
