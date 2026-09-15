# Configure Local Rate Limiting

The RateLimit custom resource (CR) allows you to apply local rate limit configuration for specific paths and headers of an exposed application.

> [!NOTE]
> Local rate limits apply to traffic directed toward the selected workload or Istio ingress gateway. If configured improperly, an attacker can exhaust all tokens and cause a Denial-of-Service attack, making the target service inaccessible.

## Prerequisites

* You have Istio and API Gateway modules in your cluster. See [Adding and Deleting a Kyma Module](https://help.sap.com/docs/btp/sap-business-technology-platform/enable-and-disable-kyma-module?locale=en-US&version=Cloud).
* To set up a custom Gateway, see [Configure a TLS Gateway in SAP BTP, Kyma Runtime](../istio-gateways/set-up-tls-gateway.md). Alternatively, you can use the default domain of your Kyma cluster and the default Gateway `kyma-system/kyma-gateway`.
  
  > [!NOTE]
  > Because the default Kyma domain is a wildcard domain, which uses a simple TLS Gateway, it is recommended that you set up your custom domain for use in a production environment. For more information, see [Istio Gateways](../istio-gateways/README.md).

  > [!TIP]
  > To learn what the default domain of your Kyma cluster is, run `kubectl get gateway -n kyma-system kyma-gateway -o jsonpath='{.spec.servers[0].hosts}'`.


## Deploy a Sample Service

1. Create a `test` namespace and enable Istio sidecar injection:
    ```bash
    kubectl create namespace test
    kubectl label namespace test istio-injection=enabled
    ```

2. Deploy and expose a sample HTTPBin Service:
    ```bash
    kubectl run httpbin --namespace test --image=kennethreitz/httpbin --labels app=httpbin
    kubectl expose --namespace test pod httpbin --port 80
    ```

3. Export the domain name under which you expose your HTTPBin Service:
    ```bash
    export WORKLOAD_DOMAIN={YOUR_WORKLOAD_DOMAIN}
    ```
    For example, `httpbin.my-domain.example.com`.

4. Create an APIRule to expose the HTTPBin Service:
    
    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: gateway.kyma-project.io/v2
    kind: APIRule
    metadata:
      name: httpbin
      namespace: test
    spec:
      hosts:
        - ${WORKLOAD_DOMAIN}
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

5. To verify the connection to the HTTPBin workload, run:
    ```bash
    curl -Lk https://${WORKLOAD_DOMAIN}/ip
    ```

    If successful, you get a response with the request's origin IP address:
    ```
    {
       "origin": "127.0.0.1"
    }
    ```

## Deploy Path-Based Rate Limit Configuration

The following example sets up a local rate limit for all endpoints exposed by the HTTPBin Service.
Additionally, it configures a separate rate limit for the `/ip` path.

Make sure that the **enableResponseHeaders** field is set to `true`. This enables the **x-ratelimit-limit** and **x-ratelimit-remaining** response headers, which can help confirm that the rate limits are working.

> [!NOTE]
> The **fillInterval** of each additional bucket must be a multiple of the default bucket's **fillInterval**.
> If the configuration is incorrect, the RateLimit CR is in the `Error` state, and the rate limit is not applied.

1. Create the RateLimit CR:

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

2. To verify the CR is applied, run:

    ```bash
    kubectl get ratelimits --namespace test ratelimit-path-sample
    ```

    If successful, you get the following response:
    ```
    NAME                    STATUS   AGE
    ratelimit-path-sample   Ready    1s
    ```

3. To verify the rate limit is working, run:

    ```bash
    curl -kLv https://${WORKLOAD_DOMAIN}/ip
    ```

    If successful, the response contains the **x-ratelimit-limit** and **x-ratelimit-remaining** headers:
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

4. To follow the next example, remove the RateLimit CR:

    ```bash
    kubectl delete ratelimits -n test ratelimit-path-sample
    ```

## Deploy Header-Based Rate Limit Configuration

The following example sets up a local rate limit for all endpoints exposed by the HTTPBin Service.
Additionally, it configures a separate rate limit for requests with the header **X-Rate-Limited** set to `true`.

Make sure that the **enableResponseHeaders** field is set to `true`. This enables the **x-ratelimit-limit** and **x-ratelimit-remaining** response headers, which can help confirm that the rate limits are working.

> [!NOTE]
> The **fillInterval** of each additional bucket must be a multiple of the default bucket's **fillInterval**.
> If the configuration is incorrect, the RateLimit CR is in the `Error` state, and the rate limit is not applied.

1. Create the RateLimit CR:

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

2. To verify the CR is applied, run:

    ```bash
    kubectl get ratelimits --namespace test ratelimit-header-sample
    ```

    If successful, you get the following response:
    ```
    NAME                      STATUS   AGE
    ratelimit-header-sample   Ready    1s
    ```

3. To verify the default bucket is working, send a request without the header:

    ```bash
    curl -kLv https://${WORKLOAD_DOMAIN}/headers
    ```

    If successful, the response contains the **x-ratelimit-limit** and **x-ratelimit-remaining** headers:
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

4. To verify the header-based bucket is working, send a request with the `X-Rate-Limited: true` header:

    ```bash
    curl -H "X-Rate-Limited: true" -kLv https://${WORKLOAD_DOMAIN}/headers
    ```

    If successful, the response shows a higher limit for the header-based bucket:
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

5. To follow the next example, remove the RateLimit CR:

    ```bash
    kubectl delete ratelimits -n test ratelimit-header-sample
    ```

## Deploy Path and Header-Based Rate Limit Configuration

The following example sets up a local rate limit for all endpoints exposed by the HTTPBin Service.
Additionally, it configures a separate rate limit for the `/headers` path that is applied only when the request contains the `X-Rate-Limited: true` header.

Make sure that the **enableResponseHeaders** field is set to `true`. This enables the **x-ratelimit-limit** and **x-ratelimit-remaining** response headers, which can help confirm that the rate limits are working.

> [!NOTE]
> The **fillInterval** of each additional bucket must be a multiple of the default bucket's **fillInterval**.
> If the configuration is incorrect, the RateLimit CR is in the `Error` state, and the rate limit is not applied.

1. Create the RateLimit CR:

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

2. To verify the CR is applied, run:

    ```bash
    kubectl get ratelimits --namespace test ratelimit-path-header-sample
    ```

    If successful, you get the following response:
    ```
    NAME                           STATUS   AGE
    ratelimit-path-header-sample   Ready    1s
    ```

3. To verify the default bucket is working, send a request without the header:

    ```bash
    curl -kLv https://${WORKLOAD_DOMAIN}/headers
    ```

    If successful, the response contains the **x-ratelimit-limit** and **x-ratelimit-remaining** headers:
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

4. To verify the header-based bucket is working, send a request to `/headers` with the `X-Rate-Limited: true` header:

    ```bash
    curl -H "X-Rate-Limited: true" -kLv https://${WORKLOAD_DOMAIN}/headers
    ```

    If successful, the response shows a higher limit for the header-based bucket:
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

5. To verify that the `/ip` endpoint uses the default bucket even with the header, run:

    ```bash
    curl -H "X-Rate-Limited: true" -kLv https://${WORKLOAD_DOMAIN}/ip
    ```

    You get the `HTTP/2 429` status code, which confirms that the default bucket limit has been exceeded:
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

6. To follow the next example, remove the RateLimit CR:

    ```bash
    kubectl delete ratelimits -n test ratelimit-path-header-sample
    ```

## Deploy Rate Limit to Istio Ingress Gateway

To rate limit requests to the Istio ingress gateway, you must create a RateLimit custom resource in the `istio-system` namespace and set the **selectorLabels** field to point to the Istio ingress gateway by including the label `app: istio-ingressgateway`.

> [!NOTE]
> The **fillInterval** of each additional bucket must be a multiple of the default bucket's **fillInterval**.
> If the configuration is incorrect, the RateLimit CR is in the `Error` state, and the rate limit is not applied.

1. Create the RateLimit CR:

    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: gateway.kyma-project.io/v1alpha1
    kind: RateLimit
    metadata:
      labels:
        app: istio-ingressgateway
      name: ratelimit-ingressgateway-path-header-sample
      namespace: istio-system
    spec:
      selectorLabels:
        app: istio-ingressgateway
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

2. To verify the CR is applied, run:

    ```bash
    kubectl get ratelimits --namespace istio-system ratelimit-ingressgateway-path-header-sample
    ```

    If successful, you get the following response:
    ```
    NAME                                          STATUS   AGE
    ratelimit-ingressgateway-path-header-sample   Ready    1s
    ```

3. To verify the default bucket is working, send a request to the `/ip` endpoint:

    ```bash
    curl -kLv https://${WORKLOAD_DOMAIN}/ip
    ```

    The first request succeeds with `HTTP/2 200`. The second request within the same fill interval is rejected:
    ```
    < HTTP/2 429
    < x-ratelimit-limit: 1
    < x-ratelimit-remaining: 0
    <
    local_rate_limited
    ```

4. To verify the header-based bucket is working, send a request to `/headers` with the `X-Rate-Limited: true` header:

    ```bash
    curl -H "X-Rate-Limited: true" -kLv https://${WORKLOAD_DOMAIN}/headers
    ```

    The response shows a higher limit for the header-based bucket:
    ```
    < HTTP/2 200
    < x-ratelimit-limit: 10
    < x-ratelimit-remaining: 9
    <
    ```

5. Remove the RateLimit CR:

    ```bash
    kubectl delete ratelimits -n istio-system ratelimit-ingressgateway-path-header-sample
    ```
