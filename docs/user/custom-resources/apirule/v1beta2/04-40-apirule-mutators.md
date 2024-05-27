# APIRule Mutators

You can use mutators to enrich an incoming request with information. For the `no_auth` access strategy mutators are not supported. APIRule in version v1beta2 supports Istio `cookie` and `header` mutators for the Istio `jwt` access strategy. For the `no_auth` access strategy mutators are not supported. You are allowed to configure both mutators for one APIRule, but only one mutator of each type is allowed.

## Header Mutator
The headers are defined in the **headers** field of the `header` mutator. The keys represent the names of the headers, and each value is a string. You can use [Envoy command operators](https://www.envoyproxy.io/docs/envoy/latest/configuration/observability/access_log/usage#command-operators) in the header value to perform operations such as copying an incoming header to a new one. The configured headers are applied to the request. They overwrite any existing headers with the same name.

### Example

In the following example, two different headers are configured: **X-Custom-Auth**, which uses the incoming Authorization header as a value, and **X-Some-Data** with the value `some-data`.

```yaml
apiVersion: gateway.kyma-project.io/v1beta2
kind: APIRule
metadata:
  name: service-config
spec:
  gateway: kyma-system/kyma-gateway
  hosts:
    - app1.example.com
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
      jwt:
        authentications:
          - issuer: https://example.com
            jwksUri: https://example.com/.well-known/jwks.json
        authorizations:
          - audiences: ["app1"]
```

### Cookie Mutator
To configure cookies, use the **cookies** mutator configuration field. The keys represent the names of the cookies, and each value is a string. As the cookie value, you can use [Envoy command operators](https://www.envoyproxy.io/docs/envoy/latest/configuration/observability/access_log/usage#command-operators). The configured cookies are set as the `cookie` header in the request and overwrite any existing cookies.

#### Example

The following APIRule example has a new cookie added with the name **some-data** and the value `data`.

```yaml
apiVersion: gateway.kyma-project.io/v1beta2
kind: APIRule
metadata:
  name: service-config
spec:
  gateway: kyma-system/kyma-gateway
  hosts:
    - app1.example.com
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
      jwt:
        authentications:
          - issuer: https://example.com
            jwksUri: https://example.com/.well-known/jwks.json
        authorizations:
          - audiences: ["app1"]
```