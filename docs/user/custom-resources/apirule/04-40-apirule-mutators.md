# APIRule Mutators

You can use mutators to enrich an incoming request with information. Different types of mutators are supported depending on the access strategy you use:

| Access Strategy      | Mutator support                                                           |
|:---------------------|:--------------------------------------------------------------------------|
| `jwt`                  | Istio `cookie` and `header` mutators                                           |
| `oauth2_introspection` | [Oathkeeper](https://www.ory.sh/docs/oathkeeper/pipeline/mutator) mutator |
| `noop`                 | [Oathkeeper](https://www.ory.sh/docs/oathkeeper/pipeline/mutator) mutator |
| `allow`                | No mutators supported                                                     |

This document explains and provides examples of Istio mutators that are compatible with JWT access strategy. Additionally, it explores the possibility of using Oathkeeper mutators with Istio, and provides guidance on how to configure them.

## Istio mutators
The `cookie` and `header` mutators are supported in combination with the JWT access strategy. You are allowed to configure multiple mutators for one APIRule, but only one mutator of each type is allowed.

### Header mutator
The headers are defined in the **headers** field of the header mutator configuration. The keys represent the names of the headers, and each value is a string. You can use [Envoy command operators](https://www.envoyproxy.io/docs/envoy/latest/configuration/observability/access_log/usage#command-operators) in the header value to perform operations such as copying an incoming header to a new header. The configured headers are applied to the request. They overwrite any existing headers with the same name.

<div tabs name="api-rule" group="sample-cr">
  <details>
  <summary label="Example">
  Example
  </summary>

In the following example, two different headers are configured: **X-Custom-Auth**, which uses the incoming Authorization header as a value, and **X-Some-Data** with the value `some-data`.

```yaml
apiVersion: gateway.kyma-project.io/v1beta1
kind: APIRule
metadata:
  name: service-secured
  namespace: $NAMESPACE
spec:
  gateway: kyma-system/kyma-gateway
  host: httpbin.$DOMAIN_TO_EXPOSE_WORKLOADS
  service:
    name: httpbin
    port: 8000
  rules:
    - path: /headers
      methods: ["GET"]
      mutators:
        - handler: header
          config:
            headers:
              X-Custom-Auth: "%REQ(Authorization)%"
              X-Some-Data: "some-data"
      accessStrategies:
        - handler: jwt
          config:
            authentications:
              - issuer: $ISSUER
                jwksUri: $JWKS_URI
```

  </details>
</div>

### Cookie mutator
To configure cookies, use the **cookies** mutator configuration field. The keys represent the names of the cookies, and each value is a string. As the cookie value, you can use [Envoy command operators](https://www.envoyproxy.io/docs/envoy/latest/configuration/observability/access_log/usage#command-operators). The configured cookies are set as `cookie`-header in the request and overwrite any existing cookies.

<div tabs name="api-rule" group="sample-cr">
  <details>
  <summary label="Example">
  Example
  </summary>

The following APIRule example has a new cookie added with the name `some-data` and the value `data`.

```yaml
apiVersion: gateway.kyma-project.io/v1beta1
kind: APIRule
metadata:
  name: service-secured
  namespace: $NAMESPACE
spec:
  gateway: kyma-system/kyma-gateway
  host: httpbin.$DOMAIN_TO_EXPOSE_WORKLOADS
  service:
    name: httpbin
    port: 8000
  rules:
    - path: /headers
      methods: ["GET"]
      mutators:
        - handler: cookie
          config:
            cookies:
              some-data: "data"
      accessStrategies:
        - handler: jwt
          config:
            authentications:
              - issuer: $ISSUER
                jwksUri: $JWKS_URI
```

  </details>
</div>

## Support for Ory Oathkeeper mutators with Istio

### Templating support

Ory Rules have support for templating in mutators. Simple cases can be implemented with the help of envoy filter, such as the one presented in [id-token-envoyfilter](id-token-envoyfilter) directory.

### Header mutator

`Headers` type mutator can be handled with Istio with the help of Virtual Service [HeaderOperations](https://istio.io/latest/docs/reference/config/networking/virtual-service/#Headers-HeaderOperations). With HeaderOperations it is only possible to add static data, but in the case of [Ory Headers Mutator](https://www.ory.sh/docs/oathkeeper/pipeline/mutator#headers) templating is supported which receives the current Authentication Session. To support similar capabilities in Istio an [EnvoyFilter](https://istio.io/latest/docs/reference/config/networking/envoy-filter/) must be used.

Ory configuration:

```yaml
...
mutators:
  - config:
      headers:
        X-Some-Arbitrary-Data: "test"
    handler: header
...
```

Coresponding Istio Virtual Service configuration:

```yaml
spec:
  http:
    - headers:
        request:
          set:
            X-Some-Arbitrary-Data: "test"
```

### Cookie mutator

The mutator of type `cookie` can be handled the same as the `header` mutator with Istio using Virtual Service HeaderOperations. Here applies the same limitiations for Istio as only static data can be added, but [Ory Cookie Mutator](https://www.ory.sh/docs/oathkeeper/pipeline/mutator#cookie) supports templating.

Ory configuration:

```yaml
...
mutators:
  - config:
      cookies:
        user: "test"
    handler: cookie
...
```

Coresponding Istio Virtual Service configuration:

```yaml
spec:
  http:
    - headers:
        request:
          set:
            Cookie: "user=test"
```

### Id_token mutator

It seems not to be possible to support the same functionality as `id_token` mutator as this requires a mechanism for encoding and signing the response from OAuth2 server (e.g. Ory Hydra) into a JWT. 
The other issue is that the JWKS used for signing this JWT is deployed as a secret `ory-oathkeeper-jwks-secret` that would have to be fetched in context of the implementation or mounted into the component doing the encoding.

### Hydrator mutator

Support for Hydrator token would require to call external APIs in context of Istio proxy. This mutator also influences other mutators by running before others and supplying them with the outcome of running.
