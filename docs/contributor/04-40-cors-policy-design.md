# Proposal for APIRule API for configuration of CORS headers in Istio VirtualService

## API Proposal

In Istio VirtualService, user can allow CORS configuration with the following parameters:
- AllowHeaders - List of HTTP headers that can be used when requesting the resource. Serialized to Access-Control-Allow-Headers header.
- AllowMethods - List of HTTP methods allowed to access the resource. The content will be serialized into the Access-Control-Allow-Methods header.
- AllowOrigins (with type [StringMatch](https://istio.io/latest/docs/reference/config/networking/virtual-service/#StringMatch)) - String patterns that match allowed origins. An origin is allowed if any of the string matchers match. If a match is found, then the outgoing Access-Control-Allow-Origin would be set to the origin as provided by the client.
- AllowCredentials (true or false) - Indicates whether the caller is allowed to send the actual request (not the preflight) using credentials. Translates to Access-Control-Allow-Credentials header.
- ExposeHeaders - A list of HTTP headers that the browsers are allowed to access. Serialized into Access-Control-Expose-Headers header.
- MaxAge - Specifies how long the results of a preflight request can be cached. Translates to the Access-Control-Max-Age header.

We should allow exposure of all of those parameters. The structure that would hold this configuration would look like following:
```go
type CorsPolicy struct {
	AllowHeaders     []string               `json:"allowHeaders,omitempty"`
	AllowMethods     []string               `json:"allowMethods,omitempty"`
	AllowOrigins     []*v1beta1.StringMatch `json:"allowOrigins,omitempty"`
	AllowCredentials bool                   `json:"allowCredentials"`
	ExposeHeaders    []string               `json:"exposeHeaders,omitempty"`
	MaxAge           *time.Duration         `json:"maxAge,omitempty"`
}
```

## Security considerations

### Default values

It should be noted, that in the most secure scenario, CORS should be configured to not respond with any of the `Access-Control` headers. However since we need to consider backwards compatibility with current implementation, we should take notice that the current configuration, for all APIRules, is as follows:
```yaml
CorsAllowOrigins: "regex:.*"
CorsAllowMethods: "GET,POST,PUT,DELETE,PATCH"
CorsAllowHeaders: "Authorization,Content-Type,*"
```

**Decision**
\
We have decided that the go to default values should be empty, meaning secure-by-default configuration. The transition to that default should be gradual, with staying with current CORS configuration for now.

### CORS headers sanitization

Another thing is that if the workload will provide CORS headers by it's own, Istio Ingress Gateway will NOT sanitize/change the CORS headers unless the request origin matches any of those that are set up in `AllowOrigins` VirtualService configuration. This is a possible security risk, because it might not be expected that the server response will contain different headers than those defined in the APIRule.

```yaml
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: oauth2-test-ckxrj
  namespace: default
spec:
  gateways:
  - kyma-system/kyma-gateway
  hosts:
  - hello.local.kyma.dev
  http:
  - corsPolicy:
      allowHeaders:
      - Authorization
      - Content-Type
      - '*'
      allowMethods:
      - GET
      - POST
      - PUT
      - DELETE
      - PATCH
      allowOrigins:
      - exact: https://test.com
    headers:
      request:
        set:
          x-forwarded-host: hello.local.kyma.dev
    match:
    - uri:
        regex: /.*
    route:
    - destination:
        host: helloworld.default.svc.cluster.local
        port:
          number: 5000
      weight: 100
    timeout: 180s

```

**Decision**
\
APIRule will become the singular source of truth, ignoring upstream response headers. If the configuration for CORS is empty, we should enforce the default configuration mentioned in [Default values](#default-values).
