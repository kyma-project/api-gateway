# Rate Limiting in Kyma

In Kyma, you can use the [RateLimit](./04-10-ratelimit-custom-resource.md) custom resource (CR) to streamline the process of applying rate limiting to workloads and the Istio ingress gateway. Learn more about how rate limiting works and when to apply it.

## Local and Global Rate Limiting

There are two types of rate limiting:
- Local rate limiting that is enforced independently by each Envoy proxy instance. Every Pod maintains its own token buckets in memory, with no coordination with other replicas.
- Global rate limiting uses a shared external store (such as Redis) so that all replicas count requests against the same pool of tokens. This gives a precise, consistent limit regardless of how many replicas are running — but it requires additional infrastructure.

The RateLimit CR only supports configuring local rate limits. You can either apply them per workload or per Istio Ingress Gateway. You can create many RateLimit CRs but each of the must match at most one Pod.

## Workload and Ingress Rate Limiting

You can apply the RateLimit CR to a workload's Envoy sidecar or to the Istio ingress gateway. The target is determined automatically by the `selectorLabels` you configure.

Workload rate limiting is applied after the request passes through the ingress gateway and is routed to the destination service. Each Pod's sidecar maintains its own independent token buckets — there is no coordination between replicas.

Ingress rate limiting is applied at the cluster entry point, before requests are routed to any service. It protects the cluster as a whole and only counts inbound external traffic. Each ingress gateway replica maintains its own independent token buckets.

## Token Deduction Logic

Rate limiting in Kyma uses the token bucket algorithm. Each bucket is defined by three values:

| Field | Description |
|---|---|
| **maxTokens** | The bucket capacity and the number of tokens available at startup. |
| **tokensPerFill** | The number of tokens added at each fill interval. |
| **fillInterval** | How often tokens are added. The minimum value is 50ms. |

Each incoming request consumes one token. When a bucket runs out of tokens, further requests are rejected with HTTP `429 Too Many Requests`.

## Default and Additional Buckets (VERIFY THIS)

Every `RateLimit` CR requires one default bucket (`local.defaultBucket`). Optionally, you can define additional buckets that match requests by path or headers (`local.buckets`).

When only the default bucket is defined, it applies to every request:

```yaml
local:
  defaultBucket:
    maxTokens: 100
    tokensPerFill: 50
    fillInterval: 30s
```

When additional buckets are defined, matched requests are counted against the matching bucket only. The default bucket acts as a fallback for requests that do not match any additional bucket. The two are independent — a request counted against an additional bucket does not consume any token from the default bucket.

```yaml
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

## Behavior When a Workload Is Scaled

Because local rate limiting is applied per-instance, each replica enforces its own independent limit. If a workload has 10 replicas and a default bucket of `maxTokens: 10`, each replica allows up to 10 requests per fill interval. In theory, the cluster can accept up to 100 requests total — but only if traffic is spread evenly across all replicas. In practice, load balancing isn't perfectly uniform, so some replicas may exhaust their tokens and start returning `429` while others still have capacity. This means you can see rate limiting errors before the cluster-wide total reaches 100 requests.

When the workload is scaled to 15 replicas, the 5 new Pods start with full buckets and accept traffic freely, while the original 10 Pods may still be at their limit. A bucket reaching its limit does not trigger autoscaling. HPA scales workloads based on the number of requests a workload receives, not on whether a rate limit has been hit. The same applies to the ingress gateway — reaching a rate limit does not cause HPA to add more ingress gateway replicas.

## Limitations

Be aware of the following aspects and consider them before applying rate limiting:
- ...
