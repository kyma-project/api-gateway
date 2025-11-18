# APIRule Custom Resource

The `apirules.gateway.kyma-project.io` CustomResourceDefinition (CRD) describes the kind and the format of data used to configure an APIRule custom resource (CR). To get the up-to-date CRD in the yaml format, run the following command:

```bash
kubectl get crd apirules.gateway.kyma-project.io -o yaml
```

## APIVersions
- [gateway.kyma-project.io/v2](#gatewaykyma-projectiov2)



## Resource Types
- [APIRule](#apirule)



### APIRule



APIRule is the schema for APIRule APIs.





| Field | Description | Validation |
| --- | --- | --- |
| **apiVersion** <br /> string | `gateway.kyma-project.io/v2` | None |
| **kind** <br /> string | `APIRule` | None |
| **metadata** <br /> [ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.33/#objectmeta-v1-meta) | For more information on the metadata fields, see Kubernetes API documentation. | None |
| **spec** <br /> [APIRuleSpec](#apirulespec) | Defines the desired state of the APIRule. | Required <br /> |
| **status** <br /> [APIRuleStatus](#apirulestatus) | Describes the observed status of the APIRule. | None |


### APIRuleSpec



**APIRuleSpec** defines the desired state of the APIRule.



Appears in:
- [APIRule](#apirule)

| Field | Description | Validation |
| --- | --- | --- |
| **hosts** <br /> [Host](#host) array | Specifies the Service’s communication address for inbound external traffic.<br />The following formats are supported:<br />- A fully qualified domain name (FQDN) with at least two domain labels separated by dots. Each label must consist of lowercase alphanumeric characters or '-',<br />and must start and end with a lowercase alphanumeric character. For example, `my-example.domain.com`, or `example.com`.<br />- One lowercase RFC 1123 label (referred to as short host name) that must consist of lowercase alphanumeric characters or '-', and must start and end with a lowercase alphanumeric character. For example, `my-host`.<br />If you define a single label, the domain name is taken from the Gateway referenced in the APIRule. In this case, the Gateway must provide the same single host for all Server definitions<br />and it must be prefixed with `*.`. Otherwise, the validation fails. | MaxItems: 1 <br />MaxLength: 255 <br />MinItems: 1 <br /> |
| **service** <br /> [Service](#service) | Specifies the backend Service that receives traffic. The Service can be deployed inside the cluster.<br />If you don't define a Service at the **spec.service** level, each defined rule must<br />specify a Service at the **spec.rules.service** level. Otherwise, the validation fails. | None |
| **gateway** <br /> string | Specifies the Istio Gateway. The field must reference an existing Gateway in the cluster.<br />Provide the Gateway in the format `namespace/gateway`.<br />Both the namespace and the Gateway name cannot be longer than 63 characters each. | MaxLength: 127 <br /> |
| **corsPolicy** <br /> [CorsPolicy](#corspolicy) | Allows configuring CORS headers sent with the response. If **corsPolicy** is not defined, the CORS headers are removed from the response. | None |
| **rules** <br /> [Rule](#rule) array | Defines an ordered list of access rules. Each rule is an atomic configuration that<br />defines how to access a specific HTTP path. A rule consists of a path<br />pattern, one or more allowed HTTP methods, exactly one access strategy (**jwt**, **extAuth**,<br />or **noAuth**), and other optional configuration fields. | MinItems: 1 <br /> |
| **timeout** <br /> [Timeout](#timeout) | Specifies the timeout for HTTP requests in seconds for all rules.<br />You can override the value for each rule. If no timeout is specified, the default timeout of 180 seconds applies. | Maximum: 3900 <br />Minimum: 1 <br /> |


### APIRuleStatus



Describes the observed status of the APIRule.



Appears in:
- [APIRule](#apirule)

| Field | Description | Validation |
| --- | --- | --- |
| **lastProcessedTime** <br /> [Time](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.33/#time-v1-meta) | Represents the last time the APIRule status was processed. | None |
| **state** <br /> [State](#state) | Defines the reconciliation state of the APIRule.<br />The possible states are `Ready`, `Warning`, or `Error`. | Enum: [Processing Deleting Ready Error Warning] <br />Required <br /> |
| **description** <br /> string | Contains the description of the APIRule's status. | None |


### CorsPolicy



Allows configuring CORS headers sent with the response. If **corsPolicy** is not defined,
the CORS headers are removed from the response.



Appears in:
- [APIRuleSpec](#apirulespec)

| Field | Description | Validation |
| --- | --- | --- |
| **allowHeaders** <br /> string array | Indicates whether credentials are allowed in the **Access-Control-Allow-Credentials** CORS header. | None |
| **allowMethods** <br /> string array | Lists headers allowed with the **Access-Control-Allow-Headers** CORS header. | None |
| **allowOrigins** <br /> [StringMatch](#stringmatch) | Lists headers allowed with the **Access-Control-Allow-Methods** CORS header. | None |
| **allowCredentials** <br /> boolean | Lists origins allowed with the **Access-Control-Allow-Origins** CORS header. | None |
| **exposeHeaders** <br /> string array | Lists headers allowed with the **Access-Control-Expose-Headers** CORS header. | None |
| **maxAge** <br /> integer | Specifies the maximum age of CORS policy cache. The value is provided in the **Access-Control-Max-Age** CORS header. | Minimum: 1 <br /> |


### ExtAuth



**ExtAuth** contains configuration for paths that use external authorization.



Appears in:
- [Rule](#rule)

| Field | Description | Validation |
| --- | --- | --- |
| **authorizers** <br /> string array | Specifies the name of the external authorization handler. | MinItems: 1 <br /> |
| **restrictions** <br /> [JwtConfig](#jwtconfig) | Specifies JWT configuration for the external authorization handler. | None |


### Host

Underlying type: string

The host is the URL of the exposed Service. Lowercase RFC 1123 labels and FQDN are supported.

Validation:
- MaxLength: 255

Appears in:
- [APIRuleSpec](#apirulespec)



### HttpMethod

Underlying type: string

HttpMethod specifies the HTTP request method. The list of supported methods is defined in in
[RFC 9910: HTTP Semantics](https://www.rfc-editor.org/rfc/rfc9110.html) and [RFC 5789: PATCH Method for HTTP](https://www.rfc-editor.org/rfc/rfc5789.html).

Validation:
- Enum: [GET HEAD POST PUT DELETE CONNECT OPTIONS TRACE PATCH]

Appears in:
- [Rule](#rule)



### JwtAuthentication



Specifies the list of Istio JWT authentication objects.



Appears in:
- [JwtConfig](#jwtconfig)

| Field | Description | Validation |
| --- | --- | --- |
| **issuer** <br /> string | Identifies the issuer that issued the JWT. The value must be a URL.<br />Although HTTP is allowed, it is recommended that you use only HTTPS endpoints. | None |
| **jwksUri** <br /> string | Contains the URL of the provider’s public key set to validate the signature of the JWT.<br />The value must be a URL. Although HTTP is allowed, it is recommended that you use only HTTPS endpoints. | None |
| **fromHeaders** <br /> [JwtHeader](#jwtheader) array | Specifies the list of headers from which the JWT token is extracted. | None |
| **fromParams** <br /> string array | Specifies the list of parameters from which the JWT token is extracted. | None |


### JwtAuthorization



Specifies the list of Istio JWT authorization objects.



Appears in:
- [JwtConfig](#jwtconfig)

| Field | Description | Validation |
| --- | --- | --- |
| **requiredScopes** <br /> string array | Specifies the list of required scope values for the JWT. | None |
| **audiences** <br /> string array | Specifies the list of audiences required for the JWT. | None |


### JwtConfig



Configures Istio JWT authentication and authorization.



Appears in:
- [ExtAuth](#extauth)
- [Rule](#rule)

| Field | Description | Validation |
| --- | --- | --- |
| **authentications** <br /> [JwtAuthentication](#jwtauthentication) array | Specifies the list of authentication objects. | None |
| **authorizations** <br /> [JwtAuthorization](#jwtauthorization) array | Specifies the list of authorization objects. | None |


### JwtHeader

Underlying type: [struct{Name string "json:\"name\""; Prefix string "json:\"prefix,omitempty\""}](#struct{name-string-"json:\"name\"";-prefix-string-"json:\"prefix,omitempty\""})

Specifies the header from which the JWT token is extracted.



Appears in:
- [JwtAuthentication](#jwtauthentication)





### Request







Appears in:
- [Rule](#rule)

| Field | Description | Validation |
| --- | --- | --- |
| **cookies** <br /> object (keys:string, values:string) | Specifies a list of cookie key-value pairs, that are forwarded inside the Cookie header. | None |
| **headers** <br /> object (keys:string, values:string) | Specifies a list of header key-value pairs that are forwarded as header=value to the target workload. | None |


### Rule



Defines an ordered list of access rules. Each rule is an atomic access configuration that
defines how to access a specific HTTP path. A rule consists of a path pattern, one or more
allowed HTTP methods, exactly one access strategy (`jwt`, `extAuth`, or `noAuth`),
and other optional configuration fields. The order of rules in the APIRule CR is important.
Rules defined earlier in the list have a higher priority than those defined later.



Appears in:
- [APIRuleSpec](#apirulespec)

| Field | Description | Validation |
| --- | --- | --- |
| **path** <br /> string | Specifies the path on which the Service is exposed. The supported configurations are:<br /> - Exact path (e.g. /abc) - matches the specified path exactly.<br /> - The `\{*\}` operator (for example, `/foo/\{*\}` or `/foo/\{*\}/bar`) - matches<br />any request that matches the pattern with exactly one path segment in the operator's place.<br /> - The `\{**\}` operator (for example, `/foo/\{**\}` or `/foo/\{**\}/bar`) -<br /> matches any request that matches the pattern with zero or more path segments in the operator's place.<br /> The `\{**\}` operator must be the last operator in the path.<br /> - The wildcard path `/*` - matches all paths. Equivalent to the `/\{**\}` path.<br />The value might contain the operators `\{*\}` and/or `\{**\}`. It can also be a wildcard match `/*`.<br />For more information, see [Ordering Rules in APIRule v2](https://kyma-project.io/external-content/api-gateway/docs/user/custom-resources/apirule/04-20-significance-of-rule-path-and-method-order.html). | Pattern: `^((\/([A-Za-z0-9-._~!$&'()+,;=:@]\|%[0-9a-fA-F]\{2\})*)\|(\/\\{\*\{1,2\}\\}))+$\|^\/\*$` <br /> |
| **service** <br /> [Service](#service) | Specifies the backend Service that receives traffic. The Service must be deployed inside the cluster.<br />If you don't define a Service at the **spec.service** level, each defined rule must<br />specify a Service at the **spec.rules.service** level. Otherwise, the validation fails. | None |
| **methods** <br /> [HttpMethod](#httpmethod) array | Specifies the list of HTTP request methods available for spec.rules.path.<br />The list of supported methods is defined in [RFC 9910: HTTP Semantics](https://www.rfc-editor.org/rfc/rfc9110.html)<br />and [RFC 5789: PATCH Method for HTTP](https://www.rfc-editor.org/rfc/rfc5789.html). | Enum: [GET HEAD POST PUT DELETE CONNECT OPTIONS TRACE PATCH] <br />MinItems: 1 <br /> |
| **noAuth** <br /> boolean | Disables authorization when set to `true`. | None |
| **jwt** <br /> [JwtConfig](#jwtconfig) | Specifies the Istio JWT configuration. | None |
| **extAuth** <br /> [ExtAuth](#extauth) | Specifies the external authorization configuration. | None |
| **timeout** <br /> [Timeout](#timeout) | Specifies the timeout, in seconds, for HTTP requests made to spec.rules.path.<br />Timeout definitions set at this level take precedence over any timeout defined<br />at the spec.timeout level. The maximum timeout is limited to 3900 seconds (65 minutes). | Maximum: 3900 <br />Minimum: 1 <br /> |
| **request** <br /> [Request](#request) | Defines request modification rules, which are applied before forwarding the request to the target workload. | None |


### Service



Specifies the backend Service that receives traffic. The Service must be deployed inside the cluster.
If you don't define a Service at the **spec.service** level, each defined rule must
specify a Service at the **spec.rules.service** level. Otherwise, the validation fails.



Appears in:
- [APIRuleSpec](#apirulespec)
- [Rule](#rule)

| Field | Description | Validation |
| --- | --- | --- |
| **name** <br /> string | Specifies the name of the exposed Service. | None |
| **namespace** <br /> string | Specifies the namespace of the exposed Service. | Pattern: `^[a-z0-9]([-a-z0-9]*[a-z0-9])?$` <br /> |
| **port** <br /> integer | Specifies the communication port of the exposed Service. | Maximum: 65535 <br />Minimum: 1 <br /> |


### State

Underlying type: string





Appears in:
- [APIRuleStatus](#apirulestatus)

| Field | Description |
| --- | --- |
| **Ready** | The APIRule's reconciliation is finished.<br /> |
| **Processing** | The APIRule is being created or updated.<br /> |
| **Error** | An error occurred during reconciliation.<br /> |
| **Deleting** | The APIRule is being deleted.<br /> |
| **Warning** | The APIRule is misconfigured.<br /> |


### StringMatch

Underlying type: map[string]string array





Appears in:
- [CorsPolicy](#corspolicy)



### Timeout

Underlying type: integer

Specifies the timeout for HTTP requests in seconds for all rules.
You can override the value for each rule. If no timeout is specified, the default timeout of 180 seconds applies.

Validation:
- Maximum: 3900
- Minimum: 1

Appears in:
- [APIRuleSpec](#apirulespec)
- [Rule](#rule)



