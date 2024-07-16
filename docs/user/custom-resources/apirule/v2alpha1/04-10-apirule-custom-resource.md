# APIRule v2alpha1 Custom Resource <!-- {docsify-ignore-all} -->

The `apirules.gateway.kyma-project.io` CustomResourceDefinition (CRD) describes the kind and the format of data the
APIGateway Controller listens for. To get the up-to-date CRD in the `yaml` format, run the following command:

```shell
kubectl get crd apirules.gateway.kyma-project.io -o yaml
```

## Specification of APIRule v2alpha1 Custom Resource

This table lists all parameters of APIRule `v2alpha1` CRD together with their descriptions:

**Spec:**

| Field                                            | Required  | Description                                                                                                                                                                                                                                                                           | Validation                                                                                            |
|:-------------------------------------------------|:---------:|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|:------------------------------------------------------------------------------------------------------|
| **gateway**                                      |  **YES**  | Specifies the Istio Gateway.                                                                                                                                                                                                                                                          | None                                                                                                  |
| **corsPolicy**                                   |  **NO**   | Allows configuring CORS headers sent with the response. If **corsPolicy** is not defined, the CORS headers are enforced to be empty.                                                                                                                                              | None                                                                                                  |
| **corsPolicy.allowHeaders**                      |  **NO**   | Specifies headers allowed with the **Access-Control-Allow-Headers** CORS header.                                                                                                                                                                                                      | None                                                                                                  |
| **corsPolicy.allowMethods**                      |  **NO**   | Specifies methods allowed with the **Access-Control-Allow-Methods** CORS header.                                                                                                                                                                                                      | None                                                                                                  |
| **corsPolicy.allowOrigins**                      |  **NO**   | Specifies origins allowed with the **Access-Control-Allow-Origins** CORS header.                                                                                                                                                                                                      | None                                                                                                  |
| **corsPolicy.allowCredentials**                  |  **NO**   | Specifies whether credentials are allowed in the **Access-Control-Allow-Credentials** CORS header.                                                                                                                                                                                    | None                                                                                                  |
| **corsPolicy.exposeHeaders**                     |  **NO**   | Specifies headers exposed with the **Access-Control-Expose-Headers** CORS header.                                                                                                                                                                                                     | None                                                                                                  |
| **corsPolicy.maxAge**                            |  **NO**   | Specifies the maximum age of CORS policy cache. The value is provided in the **Access-Control-Max-Age** CORS header.                                                                                                                                                                  | None                                                                                                  |
| **hosts**                                        |  **YES**  | Specifies the Service's communication address for inbound external traffic. If only the leftmost label is provided, the default domain name is used.                                                                                                                                  | The full domain name or the leftmost label cannot contain the wildcard character `*`.                 |
| **service.name**                                 |  **NO**   | Specifies the name of the exposed Service.                                                                                                                                                                                                                                            | None                                                                                                  |
| **service.namespace**                            |  **NO**   | Specifies the namespace of the exposed Service.                                                                                                                                                                                                                                       | None                                                                                                  |
| **service.port**                                 |  **NO**   | Specifies the communication port of the exposed Service.                                                                                                                                                                                                                              | None                                                                                                  |
| **timeout**                                      |  **NO**   | Specifies the timeout for HTTP requests in seconds for all Access Rules. The value can be overridden for each Access Rule. </br> If no timeout is specified, the default timeout of 180 seconds applies.                                                                              | The maximum timeout is limited to 3900 seconds (65 minutes).                                          |
| **rules**                                        |  **YES**  | Specifies the list of Access Rules.                                                                                                                                                                                                                                                   | None                                                                                                  |
| **rules.service**                                |  **NO**   | Services definitions at this level have higher precedence than the Service definition at the **spec.service** level.                                                                                                                                                                  | None                                                                                                  |
| **rules.path**                                   |  **YES**  | Specifies the path of the exposed Service. If the path specified in an Access Rule overlaps with the path of another Access Rule, both Access Rules are applied. This happens, for example, if one of the Access Rules' configurations contains `*`.                                  | The value can be either an exact path or a wildcard `*`.                                              |
| **rules.methods**                                |  **NO**   | Specifies the list of HTTP request methods available for **spec.rules.path**. The list of supported methods is defined in [RFC 9910: HTTP Semantics](https://www.rfc-editor.org/rfc/rfc9110.html) and [RFC 5789: PATCH Method for HTTP](https://www.rfc-editor.org/rfc/rfc5789.html). | None                                                                                                  |
| **rules.noAuth**                                 |  **NO**   | Setting `noAuth` to `true` disables authorization.                                                                                                                                                                                                                                    | When `noAuth` is set to true, it is not allowed to define the `jwt` access strategy on the same path. |
| **rules.jwt**                                    |  **NO**   | Specifies the Istio JWT access strategy.                                                                                                                                                                                                                                              | None                                                                                                  |
| **rules.jwt.authentications**                    |  **YES**  | Specifies the list of authentication objects.                                                                                                                                                                                                                                         | None                                                                                                  |
| **rules.jwt.authentications.issuer**             |  **YES**  | Identifies the issuer that issued the JWT. <br/>The value must be a URL. Although HTTP is allowed, it is recommended that you use only HTTPS endpoints.                                                                                                                               | None                                                                                                  |
| **rules.jwt.authentications.jwksUri**            |  **YES**  | Contains the URL of the provider’s public key set to validate the signature of the JWT. <br/>The value must be a URL. Although HTTP is allowed, it is recommended that you use only HTTPS endpoints.                                                                                  | None                                                                                                  |
| **rules.jwt.authentications.fromHeaders**        |  **NO**   | Specifies the list of headers from which the JWT token is extracted.                                                                                                                                                                                                                  | None                                                                                                  |
| **rules.jwt.authentications.fromHeaders.name**   |  **YES**  | Specifies the name of the header.                                                                                                                                                                                                                                                     | None                                                                                                  |
| **rules.jwt.authentications.fromHeaders.prefix** |  **NO**   | Specifies the prefix used before the JWT token. The default is `Bearer`.                                                                                                                                                                                                              | None                                                                                                  |
| **rules.jwt.authentications.fromParams**         |  **NO**   | Specifies the list of parameters from which the JWT token is extracted.                                                                                                                                                                                                               | None                                                                                                  |
| **rules.jwt.authorizations**                     |  **NO**   | Specifies the list of authorization objects.                                                                                                                                                                                                                                          | None                                                                                                  |
| **rules.jwt.authorizations.requiredScopes**      |  **NO**   | Specifies the list of required scope values for the JWT.                                                                                                                                                                                                                              | None                                                                                                  |
| **rules.jwt.authorizations.audiences**           |  **NO**   | Specifies the list of audiences required for the JWT.                                                                                                                                                                                                                                 | None                                                                                                  |
| **rules.timeout**                                |  **NO**   | Specifies the timeout, in seconds, for HTTP requests made to **spec.rules.path**. Timeout definitions set at this level take precedence over any timeout defined at the **spec.timeout** level.                                                                                       | The maximum timeout is limited to 3900 seconds (65 minutes).                                          |

> [!WARNING]
> When you use an unsupported `v1beta1` configuration in version `v2alpha1` of the APIRule CR, you get an empty **spec**. See [supported access strategies](04-15-api-rule-access-strategies.md).

> [!WARNING]
> The Ory handler is not supported in version `v2alpha1` of the APIRule. When **noAuth** is set to true, **jwt** cannot be defined on the same path.

> [!WARNING]
>  If a service is not defined at the **spec.service** level, all defined Access Rules must have it defined at the **spec.rules.service** level. Otherwise, the validation fails.

**Status:**

The following table lists the fields of the **status** section.

| Field                  | Description                                                                                                                       |
|:-----------------------|:----------------------------------------------------------------------------------------------------------------------------------|
| **status.state**       | Defines the reconciliation state of the APIRule. The possible states are `Ready`, `Warning`, `Error`, `Processing` or `Deleting`. |
| **status.description** | Detailed description of **status.state**.                                                                                         |

## Sample Custom Resource

See an exmplary APIRule custom resource:

```yaml
apiVersion: gateway.kyma-project.io/v2alpha1
kind: APIRule
metadata:
  name: service-exposed
spec:
  gateway: kyma-system/kyma-gateway
  hosts:
    - foo.bar
  service:
    name: foo-service
    namespace: foo-namespace
    port: 8080
  timeout: 360
  rules:
    - path: /.*
      methods: [ "GET" ]
      noAuth: true
```
