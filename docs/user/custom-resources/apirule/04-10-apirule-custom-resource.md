# APIRule Custom Resource

The apirules.gateway.kyma-project.io CustomResourceDefinition (CRD) describes the kind and the format of data the APIGateway Controller listens for. To get the up-to-date CRD in the yaml format, run the following command:

```shell
kubectl get crd istios.operator.kyma-project.io -o yaml
```

## APIVersions
- [gateway.kyma-project.io/v2](#gatewaykyma-projectiov2)



## Resource Types
- [APIRule](#apirule)



### APIRule



APIRule is the Schema for ApiRule APIs.





| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `gateway.kyma-project.io/v2` | | |
| `kind` _string_ | `APIRule` | | |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.33/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. | None | None |
| `spec` _[APIRuleSpec](#apirulespec)_ |  | None | Required <br /> |
| `status` _[APIRuleStatus](#apirulestatus)_ |  | None | None |


### APIRuleSpec



APIRuleSpec defines the desired state of ApiRule.



_Appears in:_
- [APIRule](#apirule)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `hosts` _[Host](#host) array_ | Specifies the URLs of the exposed service. | None | MaxItems: 1 <br />MaxLength: 255 <br />MinItems: 1 <br /> |
| `service` _[Service](#service)_ | Specifies the Service’s communication address for inbound external traffic.<br />It must be a RFC 1123 label or a valid, fully qualified domain name (FQDN) in the following<br />format: at least two domain labels with characters, numbers, or hyphens. The host must<br />start and end with an alphanumeric character. If you use a short host name at the spec.hosts<br />level, the referenced Gateway must provide the same single host for all Server definitions<br />and it must be prefixed with *.. Otherwise, the validation fails. | None | None |
| `gateway` _string_ | Specifies the Istio Gateway. The field must reference an existing Gateway in the cluster.<br />Provide the Gateway in the format namespace/gateway.<br />Both the namespace and the Gateway name cannot be longer than 63 characters each. | None | MaxLength: 127 <br /> |
| `corsPolicy` _[CorsPolicy](#corspolicy)_ | Allows configuring CORS headers sent with the response. If corsPolicy is not defined,<br />the CORS headers are empty. | None | None |
| `rules` _[Rule](#rule) array_ | Defines an ordered list of access rules. Each rule is an atomic access configuration that<br />defines how to access a specific HTTP path. A rule consists of a path<br />pattern, one or more allowed HTTP methods, exactly one access strategy (`jwt`, `extAuth`,<br />or `noAuth`), and other optional configuration fields. | None | MinItems: 1 <br /> |
| `timeout` _[Timeout](#timeout)_ | Specifies the timeout for HTTP requests in seconds for all Access Rules.<br />The value can be overridden for each Access Rule. If no timeout is specified,<br />the default timeout of 180 seconds applies. | None | Maximum: 3900 <br />Minimum: 1 <br /> |


### APIRuleStatus



Describes the observed state of the APIRule.



_Appears in:_
- [APIRule](#apirule)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `lastProcessedTime` _[Time](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.33/#time-v1-meta)_ |  | None | None |
| `state` _[State](#state)_ | Signifies the current state of the APIRule.<br />Value can be one of ("Ready", "Processing", "Error", "Deleting", "Warning"). | None | Enum: [Processing Deleting Ready Error Warning] <br />Required <br /> |
| `description` _string_ | Description of the APIRule's status. | None | None |


### CorsPolicy



CorsPolicy allows configuration of CORS headers received downstream. If this is not defined, the default values are applied.
If CorsPolicy is configured, CORS headers received downstream will be only those defined on the APIRule



_Appears in:_
- [APIRuleSpec](#apirulespec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `allowHeaders` _string array_ | Indicates whether credentials are allowed in the Access-Control-Allow-Credentials CORS header. | None | None |
| `allowMethods` _string array_ | Lists headers allowed with the Access-Control-Allow-Headers CORS header. | None | None |
| `allowOrigins` _[StringMatch](#stringmatch)_ | Lists headers allowed with the Access-Control-Allow-Methods CORS header. | None | None |
| `allowCredentials` _boolean_ | Lists origins allowed with the Access-Control-Allow-Origins CORS header. | None | None |
| `exposeHeaders` _string array_ | Lists headers allowed with the Access-Control-Expose-Headers CORS header. | None | None |
| `maxAge` _integer_ | Specifies the maximum age of CORS policy cache. The value is provided in the Access-Control-Max-Age CORS header. | None | Minimum: 1 <br /> |


### ExtAuth



ExtAuth contains configuration for paths that use external authorization.



_Appears in:_
- [Rule](#rule)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `authorizers` _string array_ | Specifies the name of the external authorization handler. | None | MinItems: 1 <br /> |
| `restrictions` _[JwtConfig](#jwtconfig)_ | Specifies JWT configuration for the external authorization handler. | None | None |


### Host

_Underlying type:_ _string_

The host is the URL of the exposed service. Lowercase RFC 1123 labels and FQDN are supported.

_Validation:_
- MaxLength: 255

_Appears in:_
- [APIRuleSpec](#apirulespec)



### HttpMethod

_Underlying type:_ _string_

HttpMethod specifies the HTTP request method. The list of supported methods is defined in RFC 9910: HTTP Semantics and RFC 5789: PATCH Method for HTTP.

_Validation:_
- Enum: [GET HEAD POST PUT DELETE CONNECT OPTIONS TRACE PATCH]

_Appears in:_
- [Rule](#rule)



### JwtAuthentication



Specifies the list of Istio JWT authentication objects.



_Appears in:_
- [JwtConfig](#jwtconfig)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `issuer` _string_ | Identifies the issuer that issued the JWT. The value must be a URL.<br />Although HTTP is allowed, it is recommended that you use only HTTPS endpoints. | None | None |
| `jwksUri` _string_ | Contains the URL of the provider’s public key set to validate the signature of the JWT.<br />The value must be a URL. Although HTTP is allowed, it is recommended that you use only HTTPS endpoints. | None | None |
| `fromHeaders` _[JwtHeader](#jwtheader) array_ | Specifies the list of headers from which the JWT token is extracted. | None | None |
| `fromParams` _string array_ | Specifies the list of parameters from which the JWT token is extracted. | None | None |


### JwtAuthorization



Specifies the list of Istio JWT authorization objects.



_Appears in:_
- [JwtConfig](#jwtconfig)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `requiredScopes` _string array_ | Specifies the list of required scope values for the JWT. | None | None |
| `audiences` _string array_ | Specifies the list of audiences required for the JWT. | None | None |


### JwtConfig



Configures Istio JWT authentication and authorization.



_Appears in:_
- [ExtAuth](#extauth)
- [Rule](#rule)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `authentications` _[JwtAuthentication](#jwtauthentication) array_ | Specifies the list of authentication objects. | None | None |
| `authorizations` _[JwtAuthorization](#jwtauthorization) array_ | Specifies the list of authorization objects. | None | None |


### JwtHeader

_Underlying type:_ _[struct{Name string "json:\"name\""; Prefix string "json:\"prefix,omitempty\""}](#struct{name-string-"json:\"name\"";-prefix-string-"json:\"prefix,omitempty\""})_

Specifies the list of parameters from which the JWT token is extracted.



_Appears in:_
- [JwtAuthentication](#jwtauthentication)





### Request







_Appears in:_
- [Rule](#rule)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `cookies` _object (keys:string, values:string)_ | Specifies a list of cookie key-value pairs, that are forwarded inside the Cookie header. | None | None |
| `headers` _object (keys:string, values:string)_ | Specifies a list of header key-value pairs that are forwarded as header=value to the target workload. | None | None |


### Rule



Defines an ordered list of access rules. Each rule is an atomic access configuration that
defines how to access a specific HTTP path. A rule consists of a path pattern, one or more
allowed HTTP methods, exactly one access strategy (`jwt`, `extAuth`, or `noAuth`),
and other optional configuration fields.



_Appears in:_
- [APIRuleSpec](#apirulespec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `path` _string_ | Specifies the path on which the service is exposed. The supported configurations are:<br /> - Exact path (e.g. /abc) - matches the specified path exactly.<br /> - The `\{*\}` operator (e.g. `/foo/\{*\}` or `/foo/\{*\}/bar`) -<br /> matches any request that matches the pattern with exactly one path segment in the operator's place.<br /> - The `\{**\}` operator (e.g. `/foo/\{**\}` or `/foo/\{**\}/bar`) -<br /> matches any request that matches the pattern with zero or more path segments in the operator's place.<br /> The `\{**\}` operator must be the last operator in the path.<br /> - The wildcard path `/*` - matches all paths. Equivalent to `/\{**\}` path. | None | Pattern: `^((\/([A-Za-z0-9-._~!$&'()+,;=:@]\|%[0-9a-fA-F]\{2\})*)\|(\/\\{\*\{1,2\}\\}))+$\|^\/\*$` <br /> |
| `service` _[Service](#service)_ | Services definitions at this level have higher precedence than the Service definition at the spec.service level. | None | None |
| `methods` _[HttpMethod](#httpmethod) array_ | Specifies the list of HTTP request methods available for spec.rules.path.<br />The list of supported methods is defined in [RFC 9910: HTTP Semantics](https://www.rfc-editor.org/rfc/rfc9110.html)<br />and [RFC 5789: PATCH Method for HTTP](https://www.rfc-editor.org/rfc/rfc5789.html). | None | Enum: [GET HEAD POST PUT DELETE CONNECT OPTIONS TRACE PATCH] <br />MinItems: 1 <br /> |
| `noAuth` _boolean_ | Disables authorization when set to true. | None | None |
| `jwt` _[JwtConfig](#jwtconfig)_ | Specifies the Istio JWT configuration. | None | None |
| `extAuth` _[ExtAuth](#extauth)_ | Specifies the external authorization configuration. | None | None |
| `timeout` _[Timeout](#timeout)_ | Specifies the timeout, in seconds, for HTTP requests made to spec.rules.path.<br />Timeout definitions set at this level take precedence over any timeout defined<br />at the spec.timeout level. The maximum timeout is limited to 3900 seconds (65 minutes). | None | Maximum: 3900 <br />Minimum: 1 <br /> |
| `request` _[Request](#request)_ | Defines request modification rules, which are applied before forwarding the request to the target workload. | None | None |


### Service



Specifies the Istio Gateway. The field must reference an existing Gateway in the cluster.
Provide the Gateway in the format namespace/gateway.
Both the namespace and the Gateway name cannot be longer than 63 characters each.



_Appears in:_
- [APIRuleSpec](#apirulespec)
- [Rule](#rule)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _string_ | Specifies the name of the exposed Service. If you don't specify a service is defined<br />at the spec.service level, each defined access rule must specify a service at the spec.rules.service<br />level. Otherwise, the validation fails. | None | None |
| `namespace` _string_ | Specifies the namespace of the exposed Service. | None | Pattern: `^[a-z0-9]([-a-z0-9]*[a-z0-9])?$` <br /> |
| `port` _integer_ | Specifies the communication port of the exposed Service. | None | Maximum: 65535 <br />Minimum: 1 <br /> |
| `external` _boolean_ | Specifies if the service is internal (deployed in the cluster) or external. | None | None |


### State

_Underlying type:_ _string_





_Appears in:_
- [APIRuleStatus](#apirulestatus)

| Field | Description |
| --- | --- |
| `Ready` |  |
| `Processing` |  |
| `Error` |  |
| `Deleting` |  |
| `Warning` |  |


### StringMatch

_Underlying type:_ _map[string]string array_





_Appears in:_
- [CorsPolicy](#corspolicy)



### Timeout

_Underlying type:_ _integer_

Timeout for HTTP requests in seconds. The timeout can be configured up to 3900 seconds (65 minutes).

_Validation:_
- Maximum: 3900
- Minimum: 1

_Appears in:_
- [APIRuleSpec](#apirulespec)
- [Rule](#rule)



