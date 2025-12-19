

# RateLimit Custom Resource
The `ratelimits.gateway.kyma-project.io` CustomResourceDefinition (CRD) describes the kind 
and the format of data that RateLimit Controller uses to configure the request rate 
limit for applications.

To get the up-to-date CRD in the yaml format, run the following command:
```bash
kubectl get crd ratelimits.gateway.kyma-project.io -o yaml
```

## Sample Custom Resource
This is a sample RateLimit custom resource (CR):

```yaml
apiVersion: gateway.kyma-project.io/v1alpha1
kind: RateLimit
metadata:
  labels:
    app: httpbin
  name: ratelimit-path-sample
  namespace: test
spec:
  selectorLabels:
    app: httpbin
  enableResponseHeaders: true
  local:
    defaultBucket:
      maxTokens: 5
      tokensPerFill: 5
      fillInterval: 60s
    buckets:
      - path: /ip
        bucket:
          maxTokens: 10
          tokensPerFill: 5
          fillInterval: 60m
```

## Custom Resource Parameters
The following tables list all the possible parameters of a given resource together with their descriptions.

### APIVersions
- gateway.kyma-project.io/v1alpha1

### Resource Types
- [RateLimit](#ratelimit)

### BucketConfig

Contains rate limit bucket configuration.

Appears in:
- [LocalConfig](#localconfig)

| Field | Description | Validation |
| --- | --- | --- |
| **path** <br> string | Specifies the path for which rate limiting is applied. The path must start with `/`. For example, `/foo`. | Optional |
| **headers** <br> object (keys:string, values:string) | Specifies the headers for which rate limiting is applied. The key is the header's name, and the value is the header's value.All specified headers must be present in the request for this configuration to match. For example, `x-api-usage: BASIC`. | Optional |
| **bucket** <br> [BucketSpec](#bucketspec) | Defines the token bucket specification. | Required  |

### BucketSpec

Defines the token bucket specification.

Appears in:
- [BucketConfig](#bucketconfig)
- [LocalConfig](#localconfig)

| Field | Description | Validation |
| --- | --- | --- |
| **maxTokens** <br> integer | The maximum number of tokens that the bucket can hold.This is also the number of tokens that the bucket initially contains. | Required  |
| **tokensPerFill** <br> integer | The number of tokens added to the bucket during each fill interval. | Required  |
| **fillInterval** <br> [Duration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.33/#duration-v1-meta) | Specifies the fill interval. During each fill interval, the number of tokens specified in the**tokensPerFill** field is added to the bucket.The bucket cannot contain more than maxTokens tokens.The fillInterval must be greater than or equal to 50ms to avoid excessive refills. | Format: duration Required  |

### LocalConfig

Defines the local rate limit configuration.

Appears in:
- [RateLimitSpec](#ratelimitspec)

| Field | Description | Validation |
| --- | --- | --- |
| **defaultBucket** <br> [BucketSpec](#bucketspec) | The default token bucket for rate limiting requests.If additional local buckets are configured in the same RateLimit CR, this bucket serves as a fallback for requests that don't match any other bucket's criteria.Each request consumes a single token. If a token is available, the request is allowed. If no tokens are available, the request is rejected with status code `429`. | Required  |
| **buckets** <br> [BucketConfig](#bucketconfig) array | Specifies a list of additional rate limit buckets for requests. Each bucket must specify either a path or headers.For each request matching the bucket's criteria, a single token is consumed. If a token is available, the request is allowed.If no tokens are available, the request is rejected with status code `429`. | Optional |

### RateLimit

RateLimit is the Schema for reate limits API.

| Field | Description | Validation |
| --- | --- | --- |
| **apiVersion** <br> string | `gateway.kyma-project.io/v1alpha1` | Optional |
| **kind** <br> string | `RateLimit` | Optional |
| **metadata** <br> [ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.33/#objectmeta-v1-meta) | For more information on the metadata fields, see Kubernetes API documentation. | Optional |
| **spec** <br> [RateLimitSpec](#ratelimitspec) | Defines the desired state of the RateLimit CR. | Optional |
| **status** <br> [RateLimitStatus](#ratelimitstatus) | Defines the current state of the RateLimit CR. | Optional |

### RateLimitSpec

Defines the desired state of the RateLimit CR.

Appears in:
- [RateLimit](#ratelimit)

| Field | Description | Validation |
| --- | --- | --- |
| **selectorLabels** <br> object (keys:string, values:string) | Contains labels that specify the set of Pods or `istio-ingressgateway` to which the configuration applies.Each Pod must match only one RateLimit CR.The label scope is limited to the namespace where the resource is located. | MinProperties: 1 Required  |
| **local** <br> [LocalConfig](#localconfig) | Defines the local rate limit configuration. | Required  |
| **enableResponseHeaders** <br> boolean | Enables **x-rate-limit** response headers. The default value is `false`. | Optional |
| **enforce** <br> boolean | Controls whether rate limiting is enforced. If true, requests exceeding limits are rejected.If false, request limits are monitored but requests that exceed limits are not blocked.The default value is `true`. | Optional |

### RateLimitStatus

RateLimitStatus defines the observed state of RateLimit

Appears in:
- [RateLimit](#ratelimit)

| Field | Description | Validation |
| --- | --- | --- |
| **description** <br> string | Description defines the description of current State of RateLimit. | Optional |
| **state** <br> string | State describes the overall status of RateLimit. The possible values are `Ready`, `Warning`, and `Error`. | Optional |

