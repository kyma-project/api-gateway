
# APIRule `v1beta2` API Proposal

Date: 2024-03-22

## Status

- Proposed: 22.03.2024

## Context

Due to the deprecation of Ory and the introduction of new features in API Gateway, the next version of APIRule resource needs to be defined.

**Spec:**

| Field                            | Mandatory | Description                                                                                                                                                                                                                                                                                                                                  | Validation                                                                                                                                                                                                                       |
|:---------------------------------|:---------:|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|:---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------| 
| **gateway**                      |  **YES**  | Specifies the Istio Gateway.                                                                                                                                                                                                                                                                                                                 |                                                                                                                                                                                                                                  |
| **corsPolicy**                   |  **NO**   | Allows configuring CORS headers sent with the response. If **corsPolicy** is not defined, the default values are applied. If **corsPolicy** is configured, only the CORS headers defined in the APIRule are sent with the response. For more information, see the [decision record](https://github.com/kyma-project/api-gateway/issues/752). |                                                                                                                                                                                                                                  |
| **corsPolicy.allowHeaders**      |  **NO**   | Specifies headers allowed with the **Access-Control-Allow-Headers** CORS header.                                                                                                                                                                                                                                                             |                                                                                                                                                                                                                                  |
| **corsPolicy.allowMethods**      |  **NO**   | Specifies methods allowed with the **Access-Control-Allow-Methods** CORS header.                                                                                                                                                                                                                                                             |                                                                                                                                                                                                                                  |
| **corsPolicy.allowOrigins**      |  **NO**   | Specifies origins allowed with the **Access-Control-Allow-Origins** CORS header.                                                                                                                                                                                                                                                             |                                                                                                                                                                                                                                  |
| **corsPolicy.allowCredentials**  |  **NO**   | Specifies whether credentials are allowed in the **Access-Control-Allow-Credentials** CORS header.                                                                                                                                                                                                                                           |                                                                                                                                                                                                                                  |
| **corsPolicy.exposeHeaders**     |  **NO**   | Specifies headers exposed with the **Access-Control-Expose-Headers** CORS header.                                                                                                                                                                                                                                                            |                                                                                                                                                                                                                                  |
| **corsPolicy.maxAge**            |  **NO**   | Specifies the maximum age of CORS policy cache. The value is provided in the **Access-Control-Max-Age** CORS header.                                                                                                                                                                                                                         |                                                                                                                                                                                                                                  |
| **hosts**                        |  **YES**  | Specifies the Service's communication address for inbound external traffic. If only the leftmost label is provided, the default domain name is used.                                                                                                                                                                                         | The full domain name or the leftmost label cannot contain the wildcard character `*`.                                                                                                                                            |
| **service.name**                 |  **NO**   | Specifies the name of the exposed Service.                                                                                                                                                                                                                                                                                                   |                                                                                                                                                                                                                                  |
| **service.namespace**            |  **NO**   | Specifies the namespace of the exposed Service.                                                                                                                                                                                                                                                                                              |                                                                                                                                                                                                                                  |
| **service.port**                 |  **NO**   | Specifies the communication port of the exposed Service.                                                                                                                                                                                                                                                                                     |                                                                                                                                                                                                                                  |
| **timeout**                      |  **NO**   | Specifies the timeout for HTTP requests in seconds for all Access Rules. The value can be overridden for each Access Rule. </br> If no timeout is specified, the default timeout of 180 seconds applies.                                                                                                                                     | The maximum timeout is limited to 3900 seconds (65 minutes).                                                                                                                                                                     | 
| **rules**                        |  **YES**  | Specifies the list of Access Rules.                                                                                                                                                                                                                                                                                                          |                                                                                                                                                                                                                                  |
| **rules.service**                |  **NO**   | Services definitions at this level have higher precedence than the Service definition at the **spec.service** level.                                                                                                                                                                                                                         |                                                                                                                                                                                                                                  |
| **rules.path**                   |  **YES**  | Specifies the path of the exposed Service.                                                                                                                                                                                                                                                                                                   |                                                                                                                                                                                                                                  |
| **rules.methods**                |  **NO**   | Specifies the list of HTTP request methods available for **spec.rules.path**. The list of supported methods is defined in [RFC 9910: HTTP Semantics](https://www.rfc-editor.org/rfc/rfc9110.html) and [RFC 5789: PATCH Method for HTTP](https://www.rfc-editor.org/rfc/rfc5789.html).                                                        |                                                                                                                                                                                                                                  |
| **rules.mutators**               |  **NO**   | Specifies the list of the request mutators.                                                                                                                                                                                                                                                                                                  | Currently, the `Headers` and `Cookie` mutators are supported. For more information, see the [documentation](https://github.com/kyma-project/api-gateway/blob/main/docs/user/custom-resources/apirule/04-40-apirule-mutators.md). |
| **rules.noAuth**                 |  **NO**   | Setting `noAuth` to `true` disables authorization.                                                                                                                                                                                                                                                                                           | When `noAuth` is set to true, it is not allowed to define any other access strategy.                                                                                                                                             |
| **rules.accessStrategy**         |  **NO**   | Specifies the access strategy.                                                                                                                                                                                                                                                                                                               | Supported are either `extAuth` or `jwt` as [Istio](https://istio.io/latest/docs/tasks/security/authorization/authz-jwt/) access strategy.                                                                                        |
| **rules.accessStrategy.extAuth** |  **NO**   | Specifies the list of external authorizers. For more information see below and the [External Authorizer ADR](https://github.com/kyma-project/api-gateway/issues/938).                                                                                                                                                                        |                                                                                                                                                                                                                                  |
| **rules.accessStrategy.jwt**     |  **NO**   | Specifies the Istio JWT access strategy. For more information see [JWT Access Strategy](https://github.com/kyma-project/api-gateway/blob/main/docs/user/custom-resources/apirule/04-20-apirule-istio-jwt-access-strategy.md) and the bellow table.                                                                                           |                                                                                                                                                                                                                                  |
| **rules.timeout**                |  **NO**   | Specifies the timeout, in seconds, for HTTP requests made to **spec.rules.path**. Timeout definitions set at this level take precedence over any timeout defined at the **spec.timeout** level.                                                                                                                                              | The maximum timeout is limited to 3900 seconds (65 minutes).                                                                                                                                                                     |

**External Authorizer Access Strategy:**

| Field                        | Description                                                                                                                                                                                          |
|:-----------------------------|:-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **name**                     | Specifies the name of the external authorizer.                                                                                                                                                       |
| **authentications**          | Specifies the list of authentication objects.                                                                                                                                                        |
| **authentications.issuer**   | Identifies the issuer that issued the JWT. <br/>The value must be an URL. Although HTTP is allowed, it is recommended that you use only HTTPS endpoints.                                             |
| **authentications.jwksUri**  | Contains the URL of the provider’s public key set to validate the signature of the JWT. <br/>The value must be a URL. Although HTTP is allowed, it is recommended that you use only HTTPS endpoints. |
| **authorizations**           | Specifies the list of authorization objects.                                                                                                                                                         |
| **authorizations.audiences** | Specifies the list of audiences required for the JWT.                                                                                                                                                |

**JWT Access Strategy:**

| Field                                  | Mandatory | Description                                                                                                                                                                                          |
|:---------------------------------------|:----------|:-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **authentications**                    | **YES**   | Specifies the list of authentication objects.                                                                                                                                                        |
| **authentications.issuer**             | **YES**   | Identifies the issuer that issued the JWT. <br/>The value must be a URL. Although HTTP is allowed, it is recommended that you use only HTTPS endpoints.                                              |
| **authentications.jwksUri**            | **YES**   | Contains the URL of the provider’s public key set to validate the signature of the JWT. <br/>The value must be a URL. Although HTTP is allowed, it is recommended that you use only HTTPS endpoints. |
| **authentications.fromHeaders**        | **NO**    | Specifies the list of headers from which the JWT token is extracted.                                                                                                                                 |
| **authentications.fromHeaders.name**   | **YES**   | Specifies the name of the header.                                                                                                                                                                    |
| **authentications.fromHeaders.prefix** | **NO**    | Specifies the prefix used before the JWT token. The default is `Bearer `.                                                                                                                            |
| **authentications.fromParams**         | **NO**    | Specifies the list of parameters from which the JWT token is extracted.                                                                                                                              |
| **authorizations**                     | **NO**    | Specifies the list of authorization objects.                                                                                                                                                         |
| **authorizations.requiredScopes**      | **NO**    | Specifies the list of required scope values for the JWT.                                                                                                                                             |
| **authorizations.audiences**           | **NO**    | Specifies the list of audiences required for the JWT.                                                                                                                                                |

### Examples

- Multiple hosts with external authorizers:
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
          - name: oauth2-proxy
            authentications:
              - issuer: https://example.com
                jwksUri: https://example.com/.well-known/jwks.json            
            authorizations:
              - audiences: ["app1"]
          - name: geo-blocker
```

- One host with JWT:
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

- One host with `noAuth`:
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

- External authorizer with `noAuth` set to `false`:
```yaml
apiVersion: gateway.kyma-project.io/v1beta2
kind: APIRule
metadata:
  name: service-config
spec:
  gateway: kyma-system/kyma-gateway
  hosts:
    - api1.example.com
  service:
    name: httpbin
    port: 8000
  rules:
    - path: /headers
      methods: ["GET"]
      noAuth: false
      accessStrategy:
        extAuth:
          - name: geo-blocker
```
