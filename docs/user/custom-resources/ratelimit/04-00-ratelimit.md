# RateLimit Custom Resource

The `ratelimits.gateway.kyma-project.io` CustomResourceDefinition (CRD) describes the kind and the format of data that
RateLimit Controller uses to configure the request rate limit for applications.

To get the up-to-date CRD in the `yaml` format, run the following command:

```shell
kubectl get crd ratelimits.gateway.kyma-project.io -o yaml
```

## Specification

| Field                                 | Required | Description                                                                                                                                                                                                                                                                                                                                                                                        |
|---------------------------------------|----------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **selectorLabels**                    | **YES**  | Labels that specify the set of Pods to which the configuration applies.<br/> Each Pod can match only one RateLimit CR.<br/> The label scope is limited to the namespace where the resource is located.                                                                                                                                                                                             |
| **local**                             | **YES**  | Local rate limit configuration.                                                                                                                                                                                                                                                                                                                                                                    |
| **local.defaultBucket**               | **YES**  | The default token bucket for rate limiting requests. <br/>If additional local buckets are configured in the same RateLimit CR, this bucket serves as a fallback for requests that don't match any other bucket's criteria. <br/>Each request consumes a single token. If a token is available, the request is allowed. If no tokens are available, the request is rejected with status code `429`. |
| **local.defaultBucket.maxTokens**     | **YES**  | The maximum tokens that the bucket can hold. This is also the number of tokens that the bucket initially contains.                                                                                                                                                                                                                                                                                 |
| **local.defaultBucket.tokensPerFill** | **YES**  | The number of tokens added to the bucket during each fill interval.                                                                                                                                                                                                                                                                                                                                |
| **local.defaultBucket.fillInterval**  | **YES**  | The fill interval that tokens are added to the bucket. <br/>During each fill interval, `tokensPerFill` are added to the bucket. The bucket will never contain more than `maxTokens` tokens. The `fillInterval` must be greater than or equal to 50ms to avoid excessive refills.                                                                                                                   |
| **local.buckets**                     | **NO**   | List of additional rate limit buckets for requests. <br/>Each bucket must specify either a `path` or `headers`. <br/>For each request matching the bucket's criteria, a single token is consumed. If a token is available, the request is allowed. If no tokens are available, the request is rejected with status code `429`.                                                                     |
| **local.buckets.path**                | **NO**   | Specifies the path to be rate limited starting with `/`. <br/>For example, `/foo`.                                                                                                                                                                                                                                                                                                                 |
| **local.buckets.headers**             | **NO**   | Specifies the request headers to be rate limited. The key is the header name, and the value is the header value. All specified headers must be present in the request for this configuration to match. For example, `x-api-usage: BASIC`.                                                                                                                                                          |                                                                                                                                                                                    
| **local.buckets.maxTokens**           | **YES**  | The maximum tokens that the bucket can hold. This is also the number of tokens that the bucket initially contains.                                                                                                                                                                                                                                                                                 |
| **local.buckets.tokensPerFill**       | **YES**  | The number of tokens added to the bucket during each fill interval.                                                                                                                                                                                                                                                                                                                                |
| **local.buckets.fillInterval**        | **YES**  | The fill interval that tokens are added to the bucket. <br/>During each fill interval, `tokensPerFill` are added to the bucket. The bucket will never contain more than `maxTokens` tokens. The `fillInterval` must be greater than or equal to 50ms to avoid excessive refills.                                                                                                                   |
| **enableResponseHeaders**             | **NO**   | Enables **x-rate-limit** response headers. The default value is `false`.                                                                                                                                                                                                                                                                                                                           |
| **enforce**                           | **NO**   | Specifies whether the rate limit should be enforced. The default value is `true`.                                                                                                                                                                                                                                                                                                                  |


## Sample Custom Resource
   
The following example illustrates a RateLimit CR that limits the rate of requests to the `httpbin` application in the `default` namespace. 
The CR defines two local buckets: one for the `/headers` path and one for the `/ip` path. The `/headers` bucket limits only requests with the `x-api-version: v1` header. 
The default bucket is used for requests that don't match any other bucket's criteria. 
```yaml
apiVersion: gateway.kyma-project.io/v1alpha1
kind: RateLimit
metadata:
  name: httpbin-local-rate-limit
  namespace: default
spec:
  selectorLabels:
    app: httpbin
  local:
    defaultBucket:
      maxTokens: 10
      tokensPerFill: 5
      fillInterval: 30s
    buckets:
      - path: /headers
        headers:
          x-api-version: v1
        bucket:
          maxTokens: 2
          tokensPerFill: 2
          fillInterval: 30s
      - path: /ip
        bucket:
          maxTokens: 20
          tokensPerFill: 10
          fillInterval: 30s
```