### APIRule `v2` Significance of Rules Order
APIRule allows you to define a list of rules that specify how requests to your service should be handled. Each rule consists of fields that define:
- `service`
- `path`
- `methods` 
- and an authentication mechanism.

Also other fields are available, but they are not relevant to this discussion. But we will focus only on above as they are the most relevant to understanding the significance of rules order.

WHow should I specify paths and methods in rules?

### Significance of Rules Path Order
`path` field in a rule defines the request path that the rule applies to. The path can include static segments (e.g., `/anything/one`) and dynamic segments using wildcards Operators `{*}` and `{**}` and `/*` which is equivalent to `/{**}`.
They allow you to define a single APIRule that matches multiple request paths.
However, this also introduces the possibility of path conflicts.
A path conflict occurs when two or more APIRule resources match the same path and share at least one common HTTP method. This is why rule order is important.

Rules defined earlier in the list have a higher priority than those defined later. Therefore, we recommend defining rules from the most specific path to the most general.

See an example of a valid **rules.path** order, listed from the most specific to the most general:
- `/anything/one`
- `/anything/one/two`
- `/anything/{*}/one`
- `/anything/{*}/one/{**}/two`
- `/anything/{*}/{*}/two`
- `/anything/{**}/two`
- `/anything/`
- `/anything/{**}`
- `/{**}`

### Significance of connection by Methods between Rules
Understanding the relationship between paths and methods in a rule is crucial to avoid unexpected behavior. 
If a rule shares at least one common method with a preceding rule then the path from preceding rule is being excluded from following rule. Take a look at example, the following APIRule configuration excludes the `POST` and `GET` methods for the path `/anything/one` with `noAuth`. This happens because the rule with the path `/anything/{**}` shares at least one common method (`GET`) with a preceding rule.

```yaml
...
rules:
  - methods:
    - GET
    jwt:
      authentications:
        - issuer: https://example.com
          jwksUri: https://example.com/.well-known/jwks.json
    path: /anything/one
  - methods:
    - GET
    - POST
    noAuth: true
    path: /anything/{**}
```
To use the `POST` method on the path `/anything/one`, you must define separate rules for overlapping methods and paths. See the following example:
```yaml
...
rules:
  - methods:
      - GET
    jwt:
      authentications:
        - issuer: https://example.com
          jwksUri: https://example.com/.well-known/jwks.json
    path: /anything/one
  - methods:
      - GET
    noAuth: true
    path: /anything/{**}
  - methods:
      - POST
    noAuth: true
    path: /anything/{**}
```


### Significance of Rules Order

End user can specify how she/he wants to order the rules. The order of rules is significant because it determines which rule takes precedence when multiple rules could apply to a given request.


```yaml
...
rules:
  - methods:
    - GET
    jwt:
      authentications:
        - issuer: https://example.com
          jwksUri: https://example.com/.well-known/jwks.json
    path: /anything/one
  - methods:
    - GET
    - POST
    noAuth: true
    path: /anything/{**}
```