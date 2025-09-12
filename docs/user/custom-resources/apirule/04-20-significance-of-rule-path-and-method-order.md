## APIRule `v2` Rules Order Significance Explained
APIRule allows you to define a list of rules that specify how requests to your service should be handled. 

Operators `{*}` and `{**}` allow you to define a single APIRule `spec.rules` that matches multiple request paths. However, this also introduces the possibility of path conflicts when front parts of paths in different rusles with at least one common method are overlapping. A path conflict occurs when two or more APIRule `spec.rules` match the same path and share at least one common HTTP method. This is why rule order and well-thought-out methods for rule/path is important as knowing what outcome is expected will help in grouping/sorting path and methods per rule.

### Usage of `{*}`, `{**}` operators and `/*` wildcard in path

The path works with `{*}` or `{**}` path template operators or the `/*` wildcard. Regexp are not longer supported. These operators allow you to define a single APIRule `spec.rules` that matches multiple request paths. 

- `{*}`: Matches a single path component, up to the next path separator: `/`.
- `{**}` and `/*` are equivalent: Matches zero or more path segments. If present, must be the last operator.

To be a valid path template, the path must not contain `*`, `{`, or `}` outside of a supported operator or `/*` wildcard. No other characters are allowed in the path segment with the path template operator.


### Create Rules/ Specify Path and Methods for each Rule ?? 


First specify methods that you need for concrete path. When defining rules in an APIRule configuration, you can group multiple HTTP methods under a single rule for a specific path. This is useful when you want to apply the same access strategy to multiple methods for the same endpoint.
For example, if you want to allow both `GET` and `POST` requests to the `/{**}` path without authentication, you can define a single rule with both methods listed:

```yaml
...
rules:
  - methods:
      - GET
      - POST
    noAuth: true
    path: /{**}
```

The following exmaple will have same outcome as above but with separate rules for each method. This approach can be useful if you want to apply different access strategies or configurations to each method in the future:
```yaml
...
rules:
  - methods:
      - GET
    noAuth: true
    path: /{**}
  - methods:
      - POST
    noAuth: true
    path: /{**}
```
### Significanse of Rules Path Order 

 Next look for Overlapping paths this occurs when two or more rules in an APIRule configuration have paths that can match the same request URI and share at least one common method. This can lead to ambiguity about which rule should apply to a given request, especially if the rules also share common HTTP methods.

This is why rule order is important.

> [!NOTE]
> Rules defined earlier in the list have a higher priority than those defined later. Request will look for first match with path from top of APIRule `spec.rules` list. Therefore, we recommend defining rules from the most specific path to the most general looking.

The following list shows correct order of overlapping paths as each one encloses each other. They are listed from the most specific to the most general:

- `/anything/one`
- `/anything/one/two`
- `/anything/{*}/{*}/two`
- `/anything/{**}/two`
- `/anything/`
- `/anything/{**}`
- `/{**}`

If lets say we would specify it in incorrect order like:
- `/anything/{**}`
- `/anything/one`
- `/anything/one/two`
- `/anything/{*}/{*}/two`
- `/anything/{**}/two`
- `/anything/`
- `/{**}`

Of course for purpose of this example we assume every path has a common method with other path/rule. 

Then anything starting a path with `/anything/` would be matched by the first rule, and the subsequent rules with `/anything/` at the beginning would never be evaluated. This is because the first rule is a catch-all for any path that starts with `/anything/`, making the more specific rules redundant. The other rule that would work would be `/{**}` as it is the most general one so everything without `/anything/` at the beginnig would also work. 

### Significance of connection by Methods between Rules
Understanding the relationship between paths and methods in a rule is crucial to avoid unexpected behavior. 

If a rule shares at least one common method with a preceding rule then the path from preceding rule is being excluded from following rule. Take a look at example, the following APIRule configuration excludes the `POST` and `GET` methods for the path `/anything/one` with `noAuth` access strategy. This happens because the rule with the path `/anything/{**}` shares at least one common method (`GET`) with a preceding rule.

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
To use the `POST` method on the path `/anything/one` with `noAuth` access strategy, you must define separate rules for overlapping methods and paths. See the following example:
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

### Summary 

Ultimately try to follow this steps and find the best solution for your use case?? TODO 

