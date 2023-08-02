---
title: APIRule
---

The `apirules.gateway.kyma-project.io` CustomResourceDefinition (CRD) describes the kind and the format of data the API Gateway Controller listens for. To get the up-to-date CRD in the `yaml` format, run the following command:

```shell
kubectl get crd apirules.gateway.kyma-project.io -o yaml
```

## Sample custom resource

This is a sample custom resource (CR) that the API Gateway Controller listens for to expose a service. The following example has the **rules** section specified which makes API Gateway Controller create an Oathkeeper Access Rule for the service.

<div tabs name="api-rule" group="sample-cr">
  <details>
  <summary label="v1beta1">
  v1beta1
  </summary>

```yaml
apiVersion: gateway.kyma-project.io/v1beta1
kind: APIRule
metadata:
  name: service-secured
spec:
  gateway: kyma-system/kyma-gateway
  host: foo.bar
  service:
    name: foo-service
    namespace: foo-namespace
    port: 8080
  timeout: 360  
  rules:
    - path: /.*
      methods: ["GET"]
      mutators: []
      accessStrategies:
        - handler: oauth2_introspection
          config:
            required_scope: ["read"]
```

  </details>
  <details>
  <summary label="v1alpha1">
  v1alpha1
  </summary>

>**NOTE:** Since Kyma 2.5 the `v1alpha1` resource has been deprecated. However, you can still create it. It is stored as `v1beta1`.

```yaml
apiVersion: gateway.kyma-project.io/v1alpha1
kind: APIRule
metadata:
  name: service-secured
spec:
  gateway: kyma-system/kyma-gateway
  service:
    name: foo-service
    port: 8080
    host: foo.bar
  rules:
    - path: /.*
      methods: ["GET"]
      mutators: []
      accessStrategies:
        - handler: oauth2_introspection
          config:
            required_scope: ["read"]
```

  </details>
</div>

## Specification

This table lists all the possible parameters of a given resource together with their descriptions:

| Field                            | Mandatory | Description                                                                                                                                                                                                                                                                                            |
|----------------------------------|:---------:|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **metadata.name**                |  **YES**  | Specifies the name of the exposed API.                                                                                                                                                                                                                                                                 |
| **spec.gateway**                 |  **YES**  | Specifies the Istio Gateway.                                                                                                                                                                                                                                                                           |
| **spec.host**                    |  **YES**  | Specifies the service's communication address for inbound external traffic. If only the leftmost label is provided, the default domain name will be used.                                                                                                                                              |
| **spec.service.name**            |  **NO**   | Specifies the name of the exposed service.                                                                                                                                                                                                                                                             |
| **spec.service.namespace**       |  **NO**   | Specifies the Namespace of the exposed service.                                                                                                                                                                                                                                                        |
| **spec.service.port**            |  **NO**   | Specifies the communication port of the exposed service.                                                                                                                                                                                                                                               |
| **spec.timeout**                 |  **NO**   | Specifies the timeout for HTTP requests in seconds for all Oathkeeper access rules, but can be overridden for each rule. The maximum timeout is limited to 3900 seconds (65 minutes). </br> If no timeout is specified, the default timeout of 180 seconds applies.                                    |
| **spec.rules**                   |  **YES**  | Specifies the list of Oathkeeper access rules.                                                                                                                                                                                                                                                         |
| **spec.rules.service**           |  **NO**   | Services definitions at this level have higher precedence than the service definition at the **spec.service** level.                                                                                                                                                                                   |
| **spec.rules.service.name**      |  **NO**   | Specifies the name of the exposed service.                                                                                                                                                                                                                                                             |
| **spec.rules.service.namespace** |  **NO**   | Specifies the Namespace of the exposed service.                                                                                                                                                                                                                                                        |
| **spec.rules.service.port**      |  **NO**   | Specifies the communication port of the exposed service.                                                                                                                                                                                                                                               |
| **spec.rules.path**              |  **YES**  | Specifies the path of the exposed service.                                                                                                                                                                                                                                                             |
| **spec.rules.methods**           |  **NO**   | Specifies the list of HTTP request methods available for **spec.rules.path**.                                                                                                                                                                                                                          |
| **spec.rules.mutators**          |  **NO**   | Specifies the list of [Oathkeeper](https://www.ory.sh/docs/next/oathkeeper/pipeline/mutator) or Istio mutators.                                                                                                                                                                                        |
| **spec.rules.accessStrategies**  |  **YES**  | Specifies the list of access strategies. Supported are [Oathkeeper](https://www.ory.sh/docs/next/oathkeeper/pipeline/authn) `oauth2_introspection`, `jwt`, `noop` and `allow`. We also support `jwt` as [Istio](https://istio.io/latest/docs/tasks/security/authorization/authz-jwt/) access strategy. |
| **spec.rules.timeout**           |  **NO**   | Specifies the timeout for HTTP requests for the rule in seconds. The maximum timeout is limited to 3900 seconds (65 minutes). Timeout definitions at this level have a higher precedence than the timeout definition at the **spec.timeout** level.                                                    |

>**CAUTION:** If `service` is not defined at **spec.service** level, all defined rules must have `service` defined at **spec.rules.service** level. Otherwise, the validation fails.

>**CAUTION:** We do not support having both Oathkeeper and Istio `jwt` access strategies defined. Access strategies `noop` or `allow` **cannot** be used with any other access strategy on the same **spec.rules.path**.

### JWT access strategy

#### Enabling Istio JWT

To enable Istio JWT, run the following command:

``` sh
kubectl patch configmap/api-gateway-config -n kyma-system --type merge -p '{"data":{"api-gateway-config":"jwtHandler: istio"}}'
```

#### Switching back to Oathkeeper JWT

To enable Oathkeeper JWT, run the following command:

``` sh
kubectl patch configmap/api-gateway-config -n kyma-system --type merge -p '{"data":{"api-gateway-config":"jwtHandler: ory"}}'
```

#### Istio JWT configuration

>**CAUTION:** Istio JWT is **not** a production-ready feature, and API might change.

This table lists all the possible parameters of the Istio JWT access strategy together with their descriptions:

| Field                                                                     | Mandatory | Description                                                                                                                                                               |
|:--------------------------------------------------------------------------|:----------|:--------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **spec.rules.accessStrategies.config**                                    | **YES**   | Access strategy configuration, must contain at least authentication or authorization.                                                                                     |
| **spec.rules.accessStrategies.config.authentications**                    | **YES**   | List of authentication objects.                                                                                                                                           |
| **spec.rules.accessStrategies.config.authentications.issuer**             | **YES**   | Identifies the issuer that issued the JWT. <br/>The value must be an URL. Although HTTP is allowed, it is recommended that you use only HTTPS endpoints.                              |
| **spec.rules.accessStrategies.config.authentications.jwksUri**            | **YES**   | URL of the provider’s public key set to validate the signature of the JWT. <br/>The value must be an URL. Although HTTP is allowed, it is recommended that you use only HTTPS endpoints.    |
| **spec.rules.accessStrategies.config.authentications.fromHeaders**        | **NO**    | List of headers from which the JWT token is taken.                                                                                                                        |
| **spec.rules.accessStrategies.config.authentications.fromHeaders.name**   | **YES**   | Name of the header.                                                                                                                                                       |
| **spec.rules.accessStrategies.config.authentications.fromHeaders.prefix** | **NO**    | Prefix used before the JWT token. The default is `Bearer `.                                                                                                                |
| **spec.rules.accessStrategies.config.authentications.fromParams**         | **NO**    | List of parameters from which the JWT token is taken.                                                                                                                     |
| **spec.rules.accessStrategies.config.authorizations**                     | **NO**    | List of authorization objects.                                                                                                                                            |
| **spec.rules.accessStrategies.config.authorizations.requiredScopes**      | **NO**    | List of required scope values for the JWT.                                                                                                                                |
| **spec.rules.accessStrategies.config.authorizations.audiences**           | **NO**    | List of audiences required for the JWT.                                                                                                                                   |

>**CAUTION:** You can define multiple JWT issuers, but each of them must be unique.

>**CAUTION:** Currently, we support only a single `fromHeader` **or** a single `fromParameter`. Specifying both of these fields for a JWT issuer is not supported.

<div tabs name="api-rule" group="sample-cr">
  <details>
  <summary label="Example">
  Example
  </summary>

In the following example, the APIRule has two defined Issuers. `ISSUER` uses a JWT token taken from the HTTP header, which is called `X-JWT-Assertion` and has a `Kyma ` prefix. `ISSUER2` uses a JWT token taken from the URL parameter, which is called `jwt-token`. The **requiredScopes** defined in the **authorizations** field allow only JWTs with the claims `scp`, `scope` or `scopes`, the value `test`, and the audiences `example.com` and `example.org` or JWTs with the claims `scp`, `scope` or `scopes` and the values `read` and `write`.

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
      mutators: []
      accessStrategies:
        - handler: jwt
          config:
            authentications:
            - issuer: $ISSUER
              jwksUri: $JWKS_URI
              fromHeaders:
              - name: X-JWT-Assertion
                prefix: "Kyma "
            - issuer: $ISSUER2
              jwksUri: $JWKS_URI2
              fromParameters:
              - "jwt_token"
            authorizations:
            - requiredScopes: ["test"]
              audiences: ["example.com", "example.org"]
            - requiredScopes: ["read", "write"]
```

  </details>
</div>

##### Authentications
Under the hood, an authentications array creates a corresponding [requestPrincipals](https://istio.io/latest/docs/reference/config/security/authorization-policy/#Source) array in the Istio's [Authorization Policy](https://istio.io/latest/docs/reference/config/security/authorization-policy/) resource. Every `requestPrincipals` string is formatted as `<ISSUSER>/*`.

##### Authorizations
The authorizations field is optional. When not defined, the authorization is satisfied if the JWT is valid. You can define multiple authorizations for an access strategy. When multiple authorizations are defined, the request is allowed if at least one of them is satisfied.

The **requiredScopes** and **audiences** fields are optional. If **requiredScopes** are defined, the JWT has to contain all the scopes in the `scp`, `scope`, or `scopes` claims to be authorized. If **audiences** are defined, the JWT has to contain all the audiences in the `aud` claim to be authorized.

### Mutators
Different types of mutators are supported depending on the access strategy.

| Access Strategy      | Mutator support                                                           |
|:---------------------|:--------------------------------------------------------------------------|
| jwt                  | Istio cookie and header mutator                                           |
| oauth2_introspection | [Oathkeeper](https://www.ory.sh/docs/oathkeeper/pipeline/mutator) mutator |
| noop                 | [Oathkeeper](https://www.ory.sh/docs/oathkeeper/pipeline/mutator) mutator |
| allow                | No mutators supported                                                     |

### Istio mutators
Mutators can be used to enrich an incoming request with information. The following mutators are supported in combination with the `jwt` access strategy and can be defined for each rule in an `ApiRule`: `header`,`cookie`. It's possible to configure multiple mutators for one rule, but only one mutator of each type is allowed.

#### Header mutator
The headers are specified in the **headers** field of the header mutator configuration field. The keys are the names of the headers, and each value is a string. In the header value, it is possible to use [Envoy command operators](https://www.envoyproxy.io/docs/envoy/latest/configuration/observability/access_log/usage#command-operators), for example, to write an incoming header into a new header. The configured headers are set to the request and overwrite all existing headers with the same name.

<div tabs name="api-rule" group="sample-cr">
  <details>
  <summary label="Example">
  Example
  </summary>

In the following example, two different headers are specified: **X-Custom-Auth**, which uses the incoming Authorization header as a value, and **X-Some-Data** with the value `some-data`.

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

#### Cookie mutator
The cookies are specified in the **cookies** field of the cookie mutator configuration field. The keys are the names of the cookies, and each value is a string. In the cookie value, it is possible to use [Envoy command operators](https://www.envoyproxy.io/docs/envoy/latest/configuration/observability/access_log/usage#command-operators). The configured cookies are set as `Cookie`-header in the request and overwrite an existing one.

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

## Additional information

When you fetch an existing APIRule CR, the system adds the **status** section which describes the status of the VirtualService and the Oathkeeper Access Rule created for this CR. The following table lists the fields of the **status** section.

| Field   |  Description |
|:---|:---|
| **status.apiRuleStatus** | Status code describing the APIRule CR. |
| **status.virtualServiceStatus.code** | Status code describing the VirtualService. |
| **status.virtualService.desc** | Current state of the VirtualService. |
| **status.accessRuleStatus.code** | Status code describing the Oathkeeper Rule. |
| **status.accessRuleStatus.desc** | Current state of the Oathkeeper Rule. |

### Status codes

These are the status codes used to describe VirtualServices and Oathkeeper Access Rules:

| Code   |  Description |
|---|---|
| **OK** | Resource created. |
| **SKIPPED** | Skipped creating a resource. |
| **ERROR** | Resource not created. |
