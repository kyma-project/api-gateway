

# RateLimit Custom Resource
The `ratelimits.gateway.kyma-project.io` CustomResourceDefinition (CRD) describes the kind 
and the format of data that RateLimit Controller uses to configure the request rate 
limit for applications.

To get the up-to-date CRD in the yaml format, run the following command:
```bash
kubectl get crd ratelimits.gateway.kyma-project.io -o yaml
```

## Sample Custom Resource
This is a sample RateLimit CR:

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

BucketConfig represents a rate limit bucket configuration.

Appears in:
- [LocalConfig](#localconfig)

| Field | Description | Validation |
| --- | --- | --- |
| **path** <br /> string |  | Optional |
| **headers** <br /> object (keys:string, values:string) |  | Optional |
| **bucket** <br /> [BucketSpec](#bucketspec) |  | Required <br /> |

### BucketSpec

BucketSpec defines the token bucket specification.

Appears in:
- [BucketConfig](#bucketconfig)
- [LocalConfig](#localconfig)

| Field | Description | Validation |
| --- | --- | --- |
| **maxTokens** <br /> integer |  | Required <br /> |
| **tokensPerFill** <br /> integer |  | Required <br /> |
| **fillInterval** <br /> [Duration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.33/#duration-v1-meta) |  | Format: duration <br />Required <br /> |

### LocalConfig

LocalConfig represents the local rate limit configuration.

Appears in:
- [RateLimitSpec](#ratelimitspec)

| Field | Description | Validation |
| --- | --- | --- |
| **defaultBucket** <br /> [BucketSpec](#bucketspec) |  | Required <br /> |
| **buckets** <br /> [BucketConfig](#bucketconfig) array |  | Optional |

### RateLimit

RateLimit is the Schema for the ratelimits API

| Field | Description | Validation |
| --- | --- | --- |
| **apiVersion** <br /> string | `gateway.kyma-project.io/v1alpha1` | None |
| **kind** <br /> string | `RateLimit` | None |
| **metadata** <br /> [ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.33/#objectmeta-v1-meta) | For more information on the metadata fields, see Kubernetes API documentation. | Optional |
| **spec** <br /> [RateLimitSpec](#ratelimitspec) |  | Optional |
| **status** <br /> [RateLimitStatus](#ratelimitstatus) |  | Optional |

### RateLimitSpec

RateLimitSpec defines the desired state of RateLimit

Appears in:
- [RateLimit](#ratelimit)

| Field | Description | Validation |
| --- | --- | --- |
| **selectorLabels** <br /> object (keys:string, values:string) |  | MinProperties: 1 <br />Required <br /> |
| **local** <br /> [LocalConfig](#localconfig) |  | Required <br /> |
| **enableResponseHeaders** <br /> boolean | EnableResponseHeaders enables x-rate-limit response headers. The default value is false. | Optional |
| **enforce** <br /> boolean | Enforce specifies whether the rate limit should be enforced. The default value is `true`. | Optional |

### RateLimitStatus

RateLimitStatus defines the observed state of RateLimit

Appears in:
- [RateLimit](#ratelimit)

| Field | Description | Validation |
| --- | --- | --- |
| **description** <br /> string | Description defines the description of current State of RateLimit. | Optional |
| **state** <br /> string | State describes the overall status of RateLimit. Values are `Ready`, `Processing`, `Warning` and `Error` | Optional |

