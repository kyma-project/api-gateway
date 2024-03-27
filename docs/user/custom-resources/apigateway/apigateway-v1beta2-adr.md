
# APIRule v1beta2 API proposal

Date: 2024-03-22

## Status

- Proposed: 22.03.2024

## Context

Due to deprecation of Ory and new features in api-gateway, next version of APIRule resource needs to be defined.

**Spec:**

| Field                           | Mandatory | Description                                                                                                                                                                                                                                                                                                                                  |
|---------------------------------|:---------:|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------| 
| **gateway**                     |  **YES**  | Specifies the Istio Gateway.                                                                                                                                                                                                                                                                                                                 |
| **corsPolicy**                  |  **NO**   | Allows configuring CORS headers sent with the response. If **corsPolicy** is not defined, the default values are applied. If **corsPolicy** is configured, only the CORS headers defined in the APIRule are sent with the response. For more information, see the [decision record](https://github.com/kyma-project/api-gateway/issues/752). |
| **corsPolicy.allowHeaders**     |  **NO**   | Specifies headers allowed with the **Access-Control-Allow-Headers** CORS header.                                                                                                                                                                                                                                                             |
| **corsPolicy.allowMethods**     |  **NO**   | Specifies methods allowed with the **Access-Control-Allow-Methods** CORS header.                                                                                                                                                                                                                                                             |
| **corsPolicy.allowOrigins**     |  **NO**   | Specifies origins allowed with the **Access-Control-Allow-Origins** CORS header.                                                                                                                                                                                                                                                             |
| **corsPolicy.allowCredentials** |  **NO**   | Specifies whether credentials are allowed in the **Access-Control-Allow-Credentials** CORS header.                                                                                                                                                                                                                                           |
| **corsPolicy.exposeHeaders**    |  **NO**   | Specifies headers exposed with the **Access-Control-Expose-Headers** CORS header.                                                                                                                                                                                                                                                            |
| **corsPolicy.maxAge**           |  **NO**   | Specifies the maximum age of CORS policy cache. The value is provided in the **Access-Control-Max-Age** CORS header.                                                                                                                                                                                                                         |
| **hosts**                       |  **YES**  | Specifies the Service's communication address for inbound external traffic. If only the leftmost label is provided, the default domain name will be used.                                                                                                                                                                                    |
| **service.name**                |  **NO**   | Specifies the name of the exposed Service.                                                                                                                                                                                                                                                                                                   |
| **service.namespace**           |  **NO**   | Specifies the namespace of the exposed Service.                                                                                                                                                                                                                                                                                              |
| **service.port**                |  **NO**   | Specifies the communication port of the exposed Service.                                                                                                                                                                                                                                                                                     |
| **timeout**                     |  **NO**   | Specifies the timeout for HTTP requests in seconds for all Access Rules. The value can be overridden for each Access Rule. The maximum timeout is limited to 3900 seconds (65 minutes). </br> If no timeout is specified, the default timeout of 180 seconds applies.                                                                        |
| **rules**                       |  **YES**  | Specifies the list of Access Rules.                                                                                                                                                                                                                                                                                                          |
| **rules.service**               |  **NO**   | Services definitions at this level have higher precedence than the Service definition at the **spec.service** level.                                                                                                                                                                                                                         |
| **rules.path**                  |  **YES**  | Specifies the path of the exposed Service.                                                                                                                                                                                                                                                                                                   |
| **rules.methods**               |  **NO**   | Specifies the list of HTTP request methods available for **spec.rules.path**. The list of supported methods is defined in [RFC 9910: HTTP Semantics](https://www.rfc-editor.org/rfc/rfc9110.html) and [RFC 5789: PATCH Method for HTTP](https://www.rfc-editor.org/rfc/rfc5789.html).                                                        |
| **rules.mutators**              |  **NO**   | Specifies the list of the Istio mutators. Currently, the `Headers` and `Cookie` mutators are supported.                                                                                                                                                                                                                                      |
| **rules.accessStrategies**      |  **YES**  | Specifies the list of access strategies. Supported are `no_auth`, `extAuth` and `jwt` as [Istio](https://istio.io/latest/docs/tasks/security/authorization/authz-jwt/) access strategy.                                                                                                                                                      |
| **rules.timeout**               |  **NO**   | Specifies the timeout, in seconds, for HTTP requests made to **spec.rules.path**. The maximum timeout is limited to 3900 seconds (65 minutes). Timeout definitions set at this level take precedence over any timeout defined at the **spec.timeout** level.                                                                                 |

### Validation

Hosts:
- The **hosts** field elements must contain either full domain names or the leftmost label.
- The **hosts** field elements cannot contain the wildcard character `*`.

### Examples

- Multiple hosts with external authorizer:
```yaml
apiVersion: gateway.kyma-project.io/v1beta2
kind: APIRule
metadata:
  name: service-config
spec:
  gateway: kyma-system/kyma-gateway
  hosts:
    - api1.example.com
    - api2.example.com
  service:
    name: httpbin
    port: 8000
  rules:
    - path: /headers
      methods: ["GET"]
      accessStrategy:
        extAuth:
          name: oauth2-proxy
          restrictions:
            authentications:
              - issuer: https://example.com
                jwksUri: https://example.com/.well-known/jwks.json            
            authorizations:
              - audiences: ["app1"]
```

- One host with jwt:
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
      accessStrategy:
        jwt:
          authentications:
            - issuer: https://example.com
              jwksUri: https://example.com/.well-known/jwks.json            
          authorizations:
            - audiences: ["app1"]
```

- One host with noAuth:
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
      noAuth: true
```

- Istio mutators:
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
        - handler: cookie
          config:
            cookies:
              user: "test"
      accessStrategy:
        jwt:
          authentications:
            - issuer: https://example.com
              jwksUri: https://example.com/.well-known/jwks.json            
          authorizations:
            - audiences: ["app1"]
```
