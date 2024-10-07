# APIRule v2alpha1 Custom Resource <!-- {docsify-ignore-all} -->

> [!WARNING]
> APIRule in version `v1beta1` will become deprecated on October 28, 2024. To prepare for the introduction of the stable APIRule in version `v2`, you can start testing the API and the migration procedure using version `v2alpha1`. APIRule `v2alpha1` was introduced for testing purposes only and is not meant for use in a production environment. For more information, see the [APIRule migration blog post](https://community.sap.com/t5/technology-blogs-by-sap/sap-btp-kyma-runtime-apirule-migration-noauth-and-jwt-handlers/ba-p/13882833).

The `apirules.gateway.kyma-project.io` CustomResourceDefinition (CRD) describes the kind and the format of data the
APIGateway Controller listens for. To get the up-to-date CRD in the `yaml` format, run the following command:

```shell
kubectl get crd apirules.gateway.kyma-project.io -o yaml
```

## Specification of APIRule v2alpha1 Custom Resource

This table lists all parameters of APIRule `v2alpha1` CRD together with their descriptions:

**Spec:**

| Field                                            | Required | Description                                                                                                                                                                                                                                                                           | Validation                                                                                                            |
|:-------------------------------------------------|:--------:|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|:----------------------------------------------------------------------------------------------------------------------|
| **gateway**                                      | **YES**  | Specifies the Istio Gateway. The value must reference an actual Gateway in the cluster.                                                                                                                                                                                               | It must be in the `namespace/gateway` format. The namespace and the Gateway cannot be longer than 63 characters each. |
| **corsPolicy**                                   |  **NO**  | Allows configuring CORS headers sent with the response. If **corsPolicy** is not defined, the CORS headers are enforced to be empty.                                                                                                                                                  | None                                                                                                                  |
| **corsPolicy.allowHeaders**                      |  **NO**  | Specifies headers allowed with the **Access-Control-Allow-Headers** CORS header.                                                                                                                                                                                                      | None                                                                                                                  |
| **corsPolicy.allowMethods**                      |  **NO**  | Specifies methods allowed with the **Access-Control-Allow-Methods** CORS header.                                                                                                                                                                                                      | None                                                                                                                  |
| **corsPolicy.allowOrigins**                      |  **NO**  | Specifies origins allowed with the **Access-Control-Allow-Origins** CORS header.                                                                                                                                                                                                      | None                                                                                                                  |
| **corsPolicy.allowCredentials**                  |  **NO**  | Specifies whether credentials are allowed in the **Access-Control-Allow-Credentials** CORS header.                                                                                                                                                                                    | None                                                                                                                  |
| **corsPolicy.exposeHeaders**                     |  **NO**  | Specifies headers exposed with the **Access-Control-Expose-Headers** CORS header.                                                                                                                                                                                                     | None                                                                                                                  |
| **corsPolicy.maxAge**                            |  **NO**  | Specifies the maximum age of CORS policy cache. The value is provided in the **Access-Control-Max-Age** CORS header.                                                                                                                                                                  | None                                                                                                                  |
| **hosts**                                        | **YES**  | Specifies the Service's communication address for inbound external traffic. It must be a RFC 1123 label or a valid, fully qualified domain name (FQDN) in the following format: at least two domain labels with characters, numbers, or hyphens.                                                          | Lowercase RFC 1123 label or FQDN format.                                                                                                |
| **service.name**                                 |  **NO**  | Specifies the name of the exposed Service.                                                                                                                                                                                                                                            | None                                                                                                                  |
| **service.namespace**                            |  **NO**  | Specifies the namespace of the exposed Service.                                                                                                                                                                                                                                       | None                                                                                                                  |
| **service.port**                                 |  **NO**  | Specifies the communication port of the exposed Service.                                                                                                                                                                                                                              | None                                                                                                                  |
| **timeout**                                      |  **NO**  | Specifies the timeout for HTTP requests in seconds for all Access Rules. The value can be overridden for each Access Rule. </br> If no timeout is specified, the default timeout of 180 seconds applies.                                                                              | The maximum timeout is limited to 3900 seconds (65 minutes).                                                          |
| **rules**                                        | **YES**  | Specifies the list of Access Rules.                                                                                                                                                                                                                                                   | None                                                                                                                  |
| **rules.service**                                |  **NO**  | Services definitions at this level have higher precedence than the Service definition at the **spec.service** level.                                                                                                                                                                  | None                                                                                                                  |
| **rules.path**                                   | **YES**  | Specifies the path of the exposed Service. If the path specified in an Access Rule overlaps with the path of another Access Rule, both Access Rules are applied. This happens, for example, if one of the Access Rules' configurations contains `/*`.                                 | The value can be either an exact path or the path wildcard `/*`.                                                      |
| **rules.methods**                                |  **NO**  | Specifies the list of HTTP request methods available for **spec.rules.path**. The list of supported methods is defined in [RFC 9910: HTTP Semantics](https://www.rfc-editor.org/rfc/rfc9110.html) and [RFC 5789: PATCH Method for HTTP](https://www.rfc-editor.org/rfc/rfc5789.html). | None                                                                                                                  |
| **rules.noAuth**                                 |  **NO**  | Setting `noAuth` to `true` disables authorization.                                                                                                                                                                                                                                    | Must be set to true if jwt and extAuth are not specified.                                                             |
| **rules.request**                                |  **NO**  | Defines request modification rules, which are applied before forwarding the request to the target workload.                                                                                                                                                                           | None                                                                                                                  |
| **rules.request.cookies**                        |  **NO**  | Specifies a list of cookie key-value pairs, that are forwarded inside the **Cookie** header.                                                                                                                                                                                          | None                                                                                                                  |
| **rules.request.headers**                        |  **NO**  | Specifies a list of header key-value pairs that are forwarded as header=value to the target workload.                                                                                                                                                                                 | None                                                                                                                  |
| **rules.jwt**                                    |  **NO**  | Specifies the Istio JWT access strategy.                                                                                                                                                                                                                                              | Must exists if noAuth and extAuth are not specified.                                                                  |
| **rules.jwt.authentications**                    | **YES**  | Specifies the list of authentication objects.                                                                                                                                                                                                                                         | None                                                                                                                  |
| **rules.jwt.authentications.issuer**             | **YES**  | Identifies the issuer that issued the JWT. <br/>The value must be a URL. Although HTTP is allowed, it is recommended that you use only HTTPS endpoints.                                                                                                                               | None                                                                                                                  |
| **rules.jwt.authentications.jwksUri**            | **YES**  | Contains the URL of the provider’s public key set to validate the signature of the JWT. <br/>The value must be a URL. Although HTTP is allowed, it is recommended that you use only HTTPS endpoints.                                                                                  | None                                                                                                                  |
| **rules.jwt.authentications.fromHeaders**        |  **NO**  | Specifies the list of headers from which the JWT token is extracted.                                                                                                                                                                                                                  | None                                                                                                                  |
| **rules.jwt.authentications.fromHeaders.name**   | **YES**  | Specifies the name of the header.                                                                                                                                                                                                                                                     | None                                                                                                                  |
| **rules.jwt.authentications.fromHeaders.prefix** |  **NO**  | Specifies the prefix used before the JWT token. The default is `Bearer`.                                                                                                                                                                                                              | None                                                                                                                  |
| **rules.jwt.authentications.fromParams**         |  **NO**  | Specifies the list of parameters from which the JWT token is extracted.                                                                                                                                                                                                               | None                                                                                                                  |
| **rules.jwt.authorizations**                     |  **NO**  | Specifies the list of authorization objects.                                                                                                                                                                                                                                          | None                                                                                                                  |
| **rules.jwt.authorizations.requiredScopes**      |  **NO**  | Specifies the list of required scope values for the JWT.                                                                                                                                                                                                                              | None                                                                                                                  |
| **rules.jwt.authorizations.audiences**           |  **NO**  | Specifies the list of audiences required for the JWT.                                                                                                                                                                                                                                 | None                                                                                                                  |
| **rules.extAuth**                                |  **NO**  | Specifies the Istio External Authorization access strategy.                                                                                                                                                                                                                           | Must exists if noAuth and jwt are not specified.                                                                      |
| **rules.extAuth.authorizers**                    | **YES**  | Specifies the Istio External Authorization authorizers. In case extAuth is configured, at least one must be present.                                                                                                                                                                  | Validated that the provider exists in Istio external authorization providers.                                         |
| **rules.extAuth.restrictions**                   |  **NO**  | Specifies the Istio External Authorization JWT restrictions. Field configuration is the same as for `rules.jwt`.                                                                                                                                                                      | None                                                                                                                  |
| **rules.timeout**                                |  **NO**  | Specifies the timeout, in seconds, for HTTP requests made to **spec.rules.path**. Timeout definitions set at this level take precedence over any timeout defined at the **spec.timeout** level.                                                                                       | The maximum timeout is limited to 3900 seconds (65 minutes).                                                          |

> [!WARNING]
> When you use an unsupported `v1beta1` configuration in version `v2alpha1` of the APIRule CR, you get an empty **spec**. See [supported access strategies](04-15-api-rule-access-strategies.md).

> [!WARNING]
> The Ory handler is not supported in version `v2alpha1` of the APIRule. When **noAuth** is set to true, **jwt** cannot be defined on the same path.

> [!WARNING]
>  If a service is not defined at the **spec.service** level, all defined Access Rules must have it defined at the **spec.rules.service** level. Otherwise, the validation fails.

> [!WARNING]
>  If a short host name is defined at the **spec.hosts** level, the referenced Gateway must provide the same single host for all [Server](https://istio.io/latest/docs/reference/config/networking/gateway/#Server) definitions and it must be prefixed with `*.`. Otherwise, the validation fails.

**Status:**

The following table lists the fields of the **status** section.

| Field                  | Description                                                                                                                       |
|:-----------------------|:----------------------------------------------------------------------------------------------------------------------------------|
| **status.state**       | Defines the reconciliation state of the APIRule. The possible states are `Ready`, `Warning`, `Error`, `Processing` or `Deleting`. |
| **status.description** | Detailed description of **status.state**.                                                                                         |

## Sample Custom Resource

See an exemplary APIRule custom resource:

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
    - path: /*
      methods: [ "GET" ]
      noAuth: true
```

This sample APIRule illustrates the usage of a short host name. It uses the domain from the referenced Gateway `kyma-system/kyma-gateway`:

```yaml
apiVersion: gateway.kyma-project.io/v2alpha1
kind: APIRule
metadata:
  name: service-exposed
spec:
  gateway: kyma-system/kyma-gateway
  hosts:
    - foo
  service:
    name: foo-service
    namespace: foo-namespace
    port: 8080
  timeout: 360
  rules:
    - path: /*
      methods: [ "GET" ]
      noAuth: true
```
