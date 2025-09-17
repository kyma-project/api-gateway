# Ordering Rules in APIRule `v2`

APIRule allows you to define a list of rules that specify how requests to your service are handled. You can define a list of rules in the **spec.rules** field of the APIRule custom resource (CR). Each rule starts with a hyphen and should contain the following details: 
- the path on which the service is exposed, 
- the list of HTTP methods available for `spec.rules.path`
- specified access strategy.

## Using the Operators `{*}`, `{**}`, and `/*`

The operators `{*}`, `{**}`, and `/*` allow you to define a single APIRule **spec.rules** that matches multiple request paths. 

To define paths, you can use one or more of the following approaches:
- **Specify the exact path name.** 
  
   Samples:
  - `/example/one` 
     
     Specifies the exact path `/example/one`.
  - `/` 

    Specifies the root path.
- **Use the operator `{*}`. It matches a single path component, up to the next path separator: `/`.**

  Samples:
  - `/example/{*}` M
  
     Matches requests with path prefix `example` and containing exactly one other segment,  for example possible matches:  `/example/anything`, `/example/`.
  - `/example/{*}/one` 

    Matches requests with the path prefix `example`, exactly one additional segment in the middle, and the path suffix `one`. For example, possible matches include `/example/anything/one` and `/example/two/one`.
- **Use the operator `{**}` or `/*`. Both of these operators are equivalent. They match zero or more path segments. If present, `{**}` or `/*` must be the last operator.**

  Samples:
  - `/example/{**}` or `example/*` 
  
    Matches `/example/anything`, `/example/anything/more`, and `/example/`.
  - `/example/{**}/one` or `example/*/one` 
    
    Matches `/example/anything/two/one`, `/example/anything/one`.
  - `/{*}/example/{*}/{**}` or `/{*}/example/{*}/*` 
  
     Matches `/anything/example/anything`, `/anything/example/anything/more`.
TODO: potestowac i lepiej opisac
 > [!NOTE]
 > To be a valid path template, the path must not contain `*`, `{`, or `}` outside of a supported operator or `/*` wildcard. No other characters are allowed in the path segment with the path template operator and the wildcard.
 >
However, using the operators also introduces the possibility of path conflicts. A path conflict occurs when two or more APIRule `spec.rules` match the same path and share at least one common HTTP method. This is why it is important to consider the order of rules and to understand connection between rules based on the path prefix and shared HTTP methods. Knowing the expected outcome of a configured rule helps in organizing and sorting them.


## Creating and Ordering the Rules
If your APIRule includes multiple rules, their order matters. Follow these steps when ordering the rules to make sure you avoid path conflicts:

1. **Group paths into rules based on common HTTP methods and access strategies.**

   Specify HTTP methods that you want to allow for each path. If you want to allow more than one method for a path, you can use one of these approaches:
   - Group multiple HTTP methods in a single rule. This is useful when you want to apply the same access strategy to multiple HTTP methods for the same endpoint. 
    
     For example, if you want to allow both `GET` and `POST` requests to all service endpoints so the `/{**}` operator is specified in path without authentication, you can define a single rule with both methods:

     ```yaml
     ...
       rules:
       - path: /{**}
         methods: 
           - GET
           - POST
         noAuth: true
     ```
   - Create a separate rule for each HTTP method. The outcome of this approach in below example is the same as of grouping multiple HTTP methods in one rule. You still allow both `GET` and `POST` requests to all service endpoints without authentication.
   
     Use this approach if you want to apply different access strategies or configurations to each method in the future:
     ```yaml
     ...
       rules:
       - path: /{**}
         methods: 
           - GET
         noAuth: true
       - path: /{**}
         methods: 
           - POST
         noAuth: true     
     ```
todo: krotkie podsumowanie w note tutaj?
2. **Order the rules.**

    Look for the paths that overlap. Overlapping occurs when two or more rules in an APIRule configuration have paths that can match the same request URI and those rules share at least one common HTTP method. This can lead to ambiguity about which rule should apply to a given request, especially if the rules also share common HTTP methods. 

    When defining the order of rules, remember that each request is evaluated against the list of paths from top to bottom in your APIRule, and the first matching rule is applied.

    If it matches the first path, the first rule applies. This is why rule order is important and you must define the paths starting from the most specific one, and leave the most general as the last one.

    > [!NOTE]
    > Rules defined earlier in the list have a higher priority than those defined later. The request searches for the first matching path starting from the top of the APIRule **spec.rules** list. Therefore, we recommend ordering rules starting with the most specific path and ending with the most general.

TODO: poprawic yaml
Sample of incorrect order pochałaniajcym wszystkie matche na pierwszej ruli: i wsm to otwieramy sobie dostep do /anything/one na post z noauth a chcialibysmy raczej zeby na one byl jwt 

```yaml
     ...
       rules:
          - path: /anything/{**}
         methods: 
           - POST
   - GET
         noAuth: true
        - methods:
          - POST
          jwt:
            authentications:
              - issuer: https://example.com
                jwksUri: https://example.com/.well-known/jwks.json
          path: /anything/one
       - path: /anything/{**}/one
         methods: 
           - GET
         noAuth: true     
  ```
   todo: opis po zmianie cos w tym stylu czyli od most specific to most genereal suffix???
```yaml
   ...
   rules:
   
        - methods:
            - POST
              jwt:
              authentications:
                - issuer: https://example.com
                  jwksUri: https://example.com/.well-known/jwks.json
                  path: /anything/one
        - path: /anything/{**}/one
          methods:
            - GET
              noAuth: true
       - path: /anything/{**}
   methods:
   - POST
    - GET
      noAuth: true       
   ```
 todo: poprawic ten opis
    In this scenario, any path that starts with `/anything/` is matched with the first rule. As a result, subsequent rules that also begin with `/anything/` are never evaluated as match happens in first rule for example for `/anything/one/two` will happen in first rule. This is because the first rule acts as a catch-all for any path starting with `/anything/`, making the more specific rules redundant. The other rule that works is `/{**}` as it is the most general one that matches any path that does not start with `/anything/`. 

3. **Check for excluding rules that share common methods.**

   > [!NOTE] 
   > Understanding the relationship between paths and methods in a rule is crucial to avoid unexpected behavior. If a rule shares at least one common method with a preceding rule, then the path from preceding rule is excluded from this rule. 
    
    For example, the following APIRule configuration excludes the `POST` and `GET` methods for the path `/anything/one` with the `noAuth` access strategy. This happens because the rule with the path `/anything/{**}` shares one common HTTP method `GET` with the preceding rule with path `/anything/one`.

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

   This outcome might be unexpected if you intended to allow `POST` requests to `/anything/one` without authentication. To achieve that, you must specifically define separate rules for overlapping methods and paths. Below APIRule configuration allows `GET` requests to `/anything/one` with JWT authentication, while permitting unauthenticated `GET` requests to all paths starting with `/anything/` except for the `/anything/one` endpoint. Additionally, it allows unauthenticated `POST` requests to all paths prefixed with `/anything/`.   See the following example:

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
