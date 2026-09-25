# Rate Limiting in Kyma

In Kyma, you can use the [RateLimit](../custom-resources/ratelimit/04-10-ratelimit-custom-resource.md) custom resource (CR) to apply rate limiting to workloads and the Istio ingress gateway. Learn more about how rate limiting works and when to apply it.

## Local and Global Rate Limiting

There are two types of rate limiting:
- Local rate limiting is enforced independently by each Envoy proxy instance. Every Pod maintains its own token buckets in memory, with no coordination with other replicas.
- Global rate limiting uses a shared external store (such as Redis) so that all replicas count requests against the same pool of tokens. This gives a precise, consistent limit regardless of how many replicas are running — but it requires additional infrastructure.

The RateLimit CR only supports configuring local rate limits. You can either apply them per workload or per Istio ingress gateway. A single RateLimit CR can match multiple Pods, but each Pod must be matched by at most one RateLimit CR.

## Workload and Ingress Rate Limiting

You can apply the RateLimit CR to a workload's Envoy sidecar or to the Istio ingress gateway. The target is determined by the **selectorLabels** you configure.

Workload rate limiting is applied at the workload's Envoy sidecar, so it counts every request that reaches the Pod. This includes both external traffic that comes in through the ingress gateway and internal service-to-service traffic between workloads in the mesh. Each Pod's sidecar keeps its own token buckets, with no coordination between replicas.

To target a workload, set **selectorLabels** to match the workload's Pods and create the CR in the workload's namespace. See an example:

```yaml
apiVersion: gateway.kyma-project.io/v1alpha1
kind: RateLimit
metadata:
  name: httpbin-ratelimit
  namespace: test
spec:
  selectorLabels:
    app: httpbin
  local:
    defaultBucket:
      maxTokens: 100
      tokensPerFill: 50
      fillInterval: 30s
```

Ingress rate limiting is applied at the Istio ingress gateway. It counts every request that passes through the gateway and enforces the limit before the request is routed to a service. Because the gateway is the cluster's entry point, this is usually external traffic. Each ingress gateway replica maintains its own independent token buckets.

To target the ingress gateway, set **selectorLabels** to `app: istio-ingressgateway` and create the CR in the `istio-system` namespace. See an example:

```yaml
apiVersion: gateway.kyma-project.io/v1alpha1
kind: RateLimit
metadata:
  name: ingress-ratelimit
  namespace: istio-system
spec:
  selectorLabels:
    app: istio-ingressgateway
  local:
    defaultBucket:
      maxTokens: 100
      tokensPerFill: 50
      fillInterval: 30s
```

## Token Deduction Logic

Rate limiting in Kyma uses the token bucket algorithm. Each bucket is defined by three values:

| Field | Description |
|---|---|
| **maxTokens** | The bucket capacity and the number of tokens available at startup. |
| **tokensPerFill** |  The number of tokens replenished per fill interval, up to the **maxTokens** capacity. |
| **fillInterval** |  The duration over which **tokensPerFill** tokens are replenished. The minimum value is 50ms. |

Each incoming request consumes one token.

## Enforcement Mode

The **enforce** field controls whether rate limits block traffic. By default (`true`), requests that exceed a bucket's limit are rejected with HTTP `429 Too Many Requests`. When set to `false`, requests that exceed the limit are still counted and reflected in the **x-ratelimit-remaining** response header, but they are not blocked. See an example:

```yaml
...
spec:
  selectorLabels:
    app: my-app
  enforce: false
  enableResponseHeaders: true
  local:
    defaultBucket:
      maxTokens: 100
      tokensPerFill: 50
      fillInterval: 30s
```

To verify that your limits are sized correctly before enforcing them, use `enforce: false` together with `enableResponseHeaders: true`.

## Default and Additional Buckets

Every RateLimit CR requires one default bucket (**local.defaultBucket**). Optionally, you can define additional buckets that match requests by path or headers (**local.buckets**).

When only the default bucket is defined, it applies to every request:

```yaml
...
local:
  defaultBucket:
    maxTokens: 100
    tokensPerFill: 50
    fillInterval: 30s
```

When additional buckets are defined, matched requests are counted against the matching bucket only. The default bucket acts as a fallback for requests that do not match any additional bucket. The two are independent — a request counted against an additional bucket does not consume any token from the default bucket. You can match additional buckets by path, header, or both. See an example:

```yaml
...
local:
  defaultBucket:
    maxTokens: 100
    tokensPerFill: 50
    fillInterval: 30s
  buckets:
    - path: /ip
      bucket:
        maxTokens: 10
        tokensPerFill: 5
        fillInterval: 60s
```

In this example:
- A request to `/ip` consumes one token from the `/ip` bucket. The default bucket is not touched.
- A request to `/headers` does not match any additional bucket, so it consumes one token from the default bucket.

For any additional bucket, the **fillInterval** must be a multiple of the default bucket's **fillInterval**. Otherwise, the RateLimit CR enters the `Error` state and the rate limit is not applied.

## Behavior When a Workload Is Scaled

Because local rate limiting is applied per-instance, each replica enforces its own independent limit. If a workload has 10 replicas and a default bucket of `maxTokens: 10`, each replica allows up to 10 requests per fill interval. In theory, the cluster can accept up to 100 requests total — but only if traffic is spread evenly across all replicas. In practice, load balancing isn't perfectly uniform, so some replicas may exhaust their tokens and start returning `429` while others still have capacity. This means you can see rate limiting errors before the cluster-wide total reaches 100 requests.

When the workload is scaled to 15 replicas, the 5 new Pods start with full buckets and accept traffic freely, while the original 10 Pods may still be at their limit. A bucket reaching its limit does not directly trigger autoscaling.

## Limitations

Be aware of the following aspects before you apply rate limiting:

- **One RateLimit per Pod**: Each Pod can be targeted by at most one RateLimit CR. If you apply a second RateLimit CR to the same Pod, it enters the `Error` state and is not applied.
- **Istio sidecar injection is required**: The target Pod must have Istio sidecar injection enabled. A RateLimit CR that targets a Pod without the sidecar enters the `Error` state. This does not apply to the Istio ingress gateway, which runs the Envoy proxy natively.
- **Limits are enforced per replica**: Each Pod replica maintains its own token bucket and enforces limits independently.
- **State is not persisted**: The token bucket exists only in memory. When a Pod restarts, its bucket resets to full capacity and any previous state is lost.
- **Selector labels are namespace-scoped**: The **selectorLabels** field matches only Pods in the same namespace as the RateLimit CR.
