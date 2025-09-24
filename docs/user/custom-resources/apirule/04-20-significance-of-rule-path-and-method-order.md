# Ordering Rules in APIRule `v2`

APIRule allows you to define a list of rules that specify how requests to your service are handled. You can define a list of rules in the **spec.rules** field of the APIRule custom resource (CR). Each rule starts with a hyphen and should contain the following details: 
- the path on which the service is exposed, 
- the list of HTTP methods available for `spec.rules.path`
- specified access strategy.

## Using the Operators `{*}`, `{**}`, and `/*` wildcard

The operators `{*}`, `{**}`, and `/*` wildcard allow you to define a single APIRule **spec.rules** that matches multiple request paths. 

To define paths, you can use one or more of the following approaches:
- **Specify the exact path name.** 
  
   Samples:
  - `/example/one` 
     
     Specifies the exact path `/example/one`.
  - `/` 

    Specifies the root path.
- **Use the operator `{*}`. It matches a single path component, up to the next path separator: `/`.**

  Samples:
    - `/example/{*}/one`

      Matches requests with the path prefix `example`, exactly one additional segment in the middle, and the path suffix `one`. For example, possible match include `/example/anything/one`.

  - `/example/{*}` 
  
     Matches requests with path prefix `example` and containing exactly one other segment,  for example possible match:  `/example/anything`. 
  
    Paths `/example/` and `/example/anything/` won't match .
- **Use the operator `{**}`.** 

    **It matches zero or more path segments if it is the last element of a path.** 

    **It matches one or more path segments if it is not the last element of a path. If present, `{**}` must be the last operator.**

  Samples:
  - `/example/{**}/one` 
    
    Matches `/example/anything/two/one`, `/example/anything/one`. Paths `/example//one` and `/example/one` won't match. 
  - `/example/{**}` 

    Matches `/example/anything`, `/example/anything/more/`, and `/example/`.
  - `/{*}/example/{*}/{**}` 
  
     Matches `/anything/example/anything/`, `/anything/example/anything/more`.
- **Use only the wildcard `/*`. It matches zero or more path segments. It is equivalent to path specified like `/{**}`. It cannot be used like operators above it needs to be specified as the only thing in a whole path. It was introduced to be backward compatible.**

  Samples:
  - `/*` 

    Matches `/`, `/example/anything/more/`, and `/example/`.
  
 > [!NOTE]
 > To be a valid path template, the path must not contain `*`, `{`, or `}` outside of a supported operator or `/*` wildcard. No other characters are allowed in the path segment with the path template operator and the wildcard.
 >
However, using the wildcard and operators also introduces the possibility of path conflicts. A path conflict occurs when two or more APIRule `spec.rules` match the same path and share at least one common HTTP method. This is why it is important to consider the order of rules and to understand connection between rules based on the path prefix and shared HTTP methods. Knowing the expected outcome of a configured rule helps in organizing and sorting them.


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
    > [!NOTE]
    > You can group multiple HTTP methods in a single rule if they share the same access strategy, or create separate rules for each method to allow for different configurations in the future. Both approaches are valid and result in the same access for the example shown above.

2. **Order the rules.**

    Look for the paths that overlap. Overlapping occurs when two or more rules in an APIRule configuration have paths that can match the same request and those rules share at least one common HTTP method. This can lead to ambiguity about which rule should apply to a given request, especially if the rules also share common HTTP methods. 

    When defining the order of rules, remember that each request is evaluated against the list of paths from top to bottom in your APIRule, and the first matching rule is applied.

    If it matches the first path, the first rule applies. This is why rule order is important, and you must define them starting from the most specific paths, and leave the most general path as the last one.

    > [!NOTE]
    > Rules defined earlier in the list have a higher priority than those defined later. The request searches for the first matching rule starting from the top of the APIRule **spec.rules** list. Therefore, we recommend ordering rules starting with the most specific path and ending with the most general.

    Example of incorrect rule order that causes all matches to be captured by the first rule: this configuration allows also unauthenticated `POST` access to `/anything/{*}/one` as this path prefixes with `anything` segment so the fir st rule applies, whereas the intended behaviour is to require JWT authentication for the `/anything/{*}/one` endpoint.

   ```yaml
   ...
     rules:
       - path: /anything/{**}
         methods:
           - POST
           - GET
         noAuth: true
       - path: /anything/{*}/one
         methods:
           - POST
         jwt:
           authentications:
             - issuer: https://example.com
               jwksUri: https://example.com/.well-known/jwks.json
   ```
   In the APIRule below, the first rule specifically matches requests to `/anything/{*}/one` with the POST method, requiring JWT authentication. The second rule acts as a catch-all for any other paths that start with `/anything/`, allowing unauthenticated POST and GET requests. By placing the more specific rule first, you ensure that requests to `/anything/{*}/one` are handled as intended, while all other matching paths are covered by the more general rule. This approach prevents the more general rule from overshadowing the specific one and ensures the correct access strategy is applied to each path. 
   ```yaml
   ...
     rules:
       - path: /anything/{*}/one
         methods:
           - POST
         jwt:
           authentications:
             - issuer: https://example.com
               jwksUri: https://example.com/.well-known/jwks.json
       - path: /anything/{**}
         methods:
           - POST
           - GET
         noAuth: true
   ```

3. **Check for excluding rules that share common methods.**

   > [!NOTE] 
   > Understanding the relationship between paths and methods in a rule is crucial to avoid unexpected behavior. If a rule shares at least one common method with a preceding rule, then the path from preceding rule is excluded from this rule. 
    
    For example, the following APIRule configuration excludes the `POST` and `GET` methods for the path `/anything/one` with the `noAuth` access strategy. This happens because the rule with the path `/anything/{**}` shares one common HTTP method `POST` with the preceding rule with path `/anything/one`.

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
        - methods:
          - GET
          - POST
          noAuth: true
          path: /anything/{**}
    ```

   This outcome might be unexpected if you intended to allow `GET` requests to `/anything/one` without authentication. To achieve that, you must specifically define separate rules for overlapping methods and paths. Below APIRule configuration allows `POST` requests to `/anything/one` with JWT authentication, while permitting unauthenticated `POST` requests to all paths starting with `/anything/` except for the `/anything/one` endpoint. Additionally, it allows unauthenticated `GET` requests to all paths prefixed with `/anything/`.   See the following example:

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
        - methods:
            - POST
          noAuth: true
          path: /anything/{**}
        - methods:
            - GET
          noAuth: true
          path: /anything/{**}
    ```
