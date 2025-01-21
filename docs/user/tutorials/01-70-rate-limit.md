# Tutorials - Rate Limit

RateLimit CR allows creating local rate limit configuration for specific paths for the exposed application.

## Prerequisites

This tutorial works on kubernetes cluster with istio-operator and api-gateway-operator installed.
Follow the installation of each module before moving forward.

## Deploying a service

Create test namespace and label it to enable istio injection:
```bash
kubectl create namespace test
kubectl label namespace test istio-injection=enabled
```

Deploy and expose simple httpbin service:
```bash
kubectl run httpbin --namespace test --image=kennethreitz/httpbin --labels app=httpbin
kubectl expose --namespace test pod httpbin --port 80
```

Create APIRule to expose previously created workload:

> [NOTE]
> `httpbin.local.kyma.dev` domain will always resolve to `127.0.0.1`.
> Make sure that istio-ingressgateway is accessible under that IP.

```bash
cat <<EOF | kubectl apply -f -
apiVersion: gateway.kyma-project.io/v2alpha1
kind: APIRule
metadata:
  name: httpbin
  namespace: test
spec:
  hosts:
    - httpbin.local.kyma.dev
  gateway: kyma-system/kyma-gateway
  rules:
    - path: /*
      service:
        name: httpbin
        port: 80
      methods: ["GET","POST"]
      noAuth: true
EOF
```

Make sure that the connection to httpbin workload is working:
```bash
curl -Lk https://httpbin.local.kyma.dev/ip
```
```
{
   "origin": "127.0.0.1"
}
```

## Deploying path-based rate limit configuration

The following example sets up local rate limit for all endpoints exposed by httpbin service.
Additionally, it configures separate rate limit configuration for `/ip` path.

Make sure that `enableResponseHeaders` field is set to `true`. This enables `X-RateLimit` headers in response.
It will help identify if rate limits are working.

> [NOTE]
> Token limit *must* be a multiplication of token bucket fill timer.
> If the configuration is incorrect, RateLimit will be in Error state, and the rate limit will not be applied.

Apply the following RateLimit CR into the cluster:
```bash
cat <<EOF | kubectl apply -f -
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
EOF
```

Check if the rate limit configuration is applied:
```bash
kubectl get ratelimits --namespace test ratelimit-path-sample
```
```
NAME                    STATUS   AGE
ratelimit-path-sample   Ready    1s
```

Check if the rate limit configuration is working, and the response headers contain `X-RateLimit` headers in response:
```bash
curl -kLv https://httpbin.local.kyma.dev/ip
```
```
(...)
* Request completely sent off
< HTTP/2 200 
< server: istio-envoy
< date: ***
< content-type: application/json
< content-length: 29
< x-envoy-upstream-service-time: 1
< x-ratelimit-limit: 10
< x-ratelimit-remaining: 9
< 
{
  "origin": "127.0.0.1"
}
* Connection #0 to host httpbin.local.kyma.dev left intact
```

Path-based rate limit is configured. Remove the RateLimit CR from a cluster.
```bash
kubectl delete ratelimits -n test ratelimit-path-sample
```

## Deploying header-based rate limit configuration

The following example sets up local rate limit for all endpoints exposed by httpbin service.
Additionally, it configures separate rate limit configuration for requests with header `X-Rate-Limited` set to `true`.

Make sure that `enableResponseHeaders` field is set to `true`. This enables `X-RateLimit` headers in response.
It will help identify if rate limits are working.

> [NOTE]
> Token limit *must* be a multiplication of token bucket fill timer.
> If the configuration is incorrect, RateLimit will be in Error state, and the rate limit will not be applied.

Apply the following RateLimit CR into the cluster:
```bash
cat <<EOF | kubectl apply -f -
apiVersion: gateway.kyma-project.io/v1alpha1
kind: RateLimit
metadata:
  labels:
    app: httpbin
  name: ratelimit-header-sample
  namespace: test
spec:
  selectorLabels:
    app: httpbin
  enableResponseHeaders: true
  local:
    defaultBucket:
      maxTokens: 1
      tokensPerFill: 1
      fillInterval: 30s
    buckets:
      - headers:
          X-Rate-Limited: "true"
        bucket:
          maxTokens: 10
          tokensPerFill: 5
          fillInterval: 30s
EOF
```

Check if the rate limit configuration is applied:
```bash
kubectl get ratelimits --namespace test ratelimit-header-sample
```
```
NAME                      STATUS   AGE
ratelimit-header-sample   Ready    1s
```

Now the workload is rate limited.
Check if the rate limit configuration is working, and the response headers contain `X-RateLimit` headers in response:
```bash
curl -kLv https://httpbin.local.kyma.dev/headers
```
```
(...)
* Request completely sent off
< HTTP/2 200 
< server: istio-envoy
< date: ***
< content-type: application/json
< content-length: 529
< x-envoy-upstream-service-time: 17
< x-ratelimit-limit: 1
< x-ratelimit-remaining: 0
< 
{
  "headers": {
    "Accept": "*/*", 
    "Host": "httpbin.local.kyma.dev", 
    "User-Agent": "curl/8.7.1", 
    "X-Envoy-Attempt-Count": "1", 
    "X-Envoy-Expected-Rq-Timeout-Ms": "180000", 
    "X-Envoy-Internal": "true", 
    "X-Forwarded-Host": "httpbin.local.kyma.dev"
  }
}
* Connection #0 to host httpbin.local.kyma.dev left intact
```

If user provides a `X-Rate-Limited: "true"` header in a request, it will be projected with different rate limits.
Check if the header-based rate limit is configured:
```bash
curl -H "X-Rate-Limited: true" -kLv https://httpbin.local.kyma.dev/headers
```
```
(...)
* Request completely sent off
< HTTP/2 200 
< server: istio-envoy
< date: ***
< content-type: application/json
< content-length: 560
< x-envoy-upstream-service-time: 2
< x-ratelimit-limit: 10
< x-ratelimit-remaining: 9
< 
{
  "headers": {
    "Accept": "*/*", 
    "Host": "httpbin.local.kyma.dev", 
    "User-Agent": "curl/8.7.1", 
    "X-Envoy-Attempt-Count": "1", 
    "X-Envoy-Expected-Rq-Timeout-Ms": "180000", 
    "X-Envoy-Internal": "true", 
    "X-Forwarded-Host": "httpbin.local.kyma.dev", 
    "X-Rate-Limited": "true"
  }
}
* Connection #0 to host httpbin.local.kyma.dev left intact
```

Header-based rate limit is configured. Remove the RateLimit CR from a cluster.
```bash
kubectl delete ratelimits -n test ratelimit-header-sample
```

## Path and header-based rate limit configuration

Rate limit configuration also can be configured to rate limit connection per path.
That means both `path` and `headers` fields can be used freely.

The following example sets up local rate limit for all endpoints exposed by httpbin service.
Additionally, it configures separate rate limit configuration for `/headers` path only if the request contains `X-Rate-Limited: true` header.

Make sure that `enableResponseHeaders` field is set to `true`. This enables `X-RateLimit` headers in response.
It will help identify if rate limits are working.

> [NOTE]
> Token limit *must* be a multiplication of token bucket fill timer.
> If the configuration is incorrect, RateLimit will be in Error state, and the rate limit will not be applied.

Apply the following RateLimit CR into the cluster:
```bash
cat <<EOF | kubectl apply -f -
apiVersion: gateway.kyma-project.io/v1alpha1
kind: RateLimit
metadata:
  labels:
    app: httpbin
  name: ratelimit-path-header-sample
  namespace: test
spec:
  selectorLabels:
    app: httpbin
  enableResponseHeaders: true
  local:
    defaultBucket:
      maxTokens: 1
      tokensPerFill: 1
      fillInterval: 30s
    buckets:
      - headers:
          X-Rate-Limited: "true"
        path: /headers
        bucket:
          maxTokens: 10
          tokensPerFill: 5
          fillInterval: 30s
EOF
```

Check if the rate limit configuration is applied:
```bash
kubectl get ratelimits --namespace test ratelimit-path-header-sample
```
```
NAME                           STATUS   AGE
ratelimit-path-header-sample   Ready    1s
```

Now the workload is rate limited.
Check if the rate limit configuration is working, and the response headers contain `X-RateLimit` headers in response.
This call will use tokens from default bucket:
```bash
curl -kLv https://httpbin.local.kyma.dev/headers
```
```
(...)
* Request completely sent off
< HTTP/2 200 
< server: istio-envoy
< date: ***
< content-type: application/json
< content-length: 529
< x-envoy-upstream-service-time: 17
< x-ratelimit-limit: 1
< x-ratelimit-remaining: 0
< 
{
  "headers": {
    "Accept": "*/*", 
    "Host": "httpbin.local.kyma.dev", 
    "User-Agent": "curl/8.7.1", 
    "X-Envoy-Attempt-Count": "1", 
    "X-Envoy-Expected-Rq-Timeout-Ms": "180000", 
    "X-Envoy-Internal": "true", 
    "X-Forwarded-Host": "httpbin.local.kyma.dev"
  }
}
* Connection #0 to host httpbin.local.kyma.dev left intact
```

If user provides a `X-Rate-Limited: "true"` header in a request to the `/headers` endpoint, it will be projected with different rate limits.
Check if the header-based rate limit is configured:
```bash
curl -H "X-Rate-Limited: true" -kLv https://httpbin.local.kyma.dev/headers
```
```
(...)
* Request completely sent off
< HTTP/2 200 
< server: istio-envoy
< date: ***
< content-type: application/json
< content-length: 560
< x-envoy-upstream-service-time: 2
< x-ratelimit-limit: 10
< x-ratelimit-remaining: 9
< 
{
  "headers": {
    "Accept": "*/*", 
    "Host": "httpbin.local.kyma.dev", 
    "User-Agent": "curl/8.7.1", 
    "X-Envoy-Attempt-Count": "1", 
    "X-Envoy-Expected-Rq-Timeout-Ms": "180000", 
    "X-Envoy-Internal": "true", 
    "X-Forwarded-Host": "httpbin.local.kyma.dev", 
    "X-Rate-Limited": "true"
  }
}
* Connection #0 to host httpbin.local.kyma.dev left intact
```

If the request with the header is sent to different endpoint, the token will be used from the default bucket.
That means access to `/ip` endpoint should be rate-limited.

```bash
curl -H "X-Rate-Limited: true" -kLv https://httpbin.local.kyma.dev/ip
```
```
(...)
> X-Rate-Limited: true
> 
* Request completely sent off
< HTTP/2 429 
< content-length: 18
< content-type: text/plain
< x-ratelimit-limit: 1
< x-ratelimit-remaining: 0
< date: Wed, 22 Jan 2025 14:07:10 GMT
< server: istio-envoy
< x-envoy-upstream-service-time: 2
< 
* Connection #0 to host httpbin.local.kyma.dev left intact
local_rate_limited
```

Path and header-based rate limit is configured. Remove the RateLimit CR from a cluster.
```bash
kubectl delete ratelimits -n test ratelimit-path-header-sample
```