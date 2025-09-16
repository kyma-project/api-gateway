# Ordering Rules in APIRule `v2`

APIRule allows you to define a list of rules that specify how requests to your service are handled. You can define a list of rules in the **spec.rules** field of the APIRule custom resource (CR). Each rule starts with a hyphen and must contain the following details: the path's name, allowed access methods, authentication method for the specified path. 

## Using the Operators `{*}`, `{**}`, and `/*`

The operators `{*}` and `{**}` allow you to define a single APIRule **spec.rules** that matches multiple request paths. However, using the operators also introduces the possibility of path conflicts. A path conflict occurs when two or more APIRule `spec.rules` match the same path and share at least one common HTTP method. This is why it is important to consider the order of rules and the access methods you want to allow for each path. Knowing the expected outcome helps in organizing and sorting the rules.

To define paths, you can use one of the following approaches:
- Specify the exact path name.
- Use the operator `{*}`. It matches a single path component, up to the next path separator: `/`.
- Use the operator `{**}` or `/*`. Both of these operators are equivalent. They match zero or more path segments. If present, `{**}` or `/*` must be the last operator.

To be a valid path template, the path must not contain `*`, `{`, or `}` outside of a supported operator or `/*` wildcard. No other characters are allowed in the path segment with the path template operator.

## Creating and Ordering the Rules
If your APIRule includes multiple rules, their order matters. Follow these steps when ordering the rules to make sure you avoid path conflicts:

1. Group paths based on common methods and access strategies.

  Specify methods that you want to allow for each path. If you want to allow more then one method for a path, you can use one of these approaches:
    - Group multiple HTTP methods in a single rule. This is useful when you want to apply the same access strategy to multiple methods for the same endpoint. For example, if you want to allow both `GET` and `POST` requests to the `/{**}` path without authentication, you can define a single rule with both methods:

      ```yaml
      ...
      rules:
        - methods:
            - GET
            - POST
          noAuth: true
          path: /{**}
      ```
   - Create a separate rule for each method. The outcome of this approach is the same as of grouping multiple methods in one rule. Use this approach if you want to apply different access strategies or configurations to each method in the future:

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

2. Order the rules.

    Look for the paths that overlap. Overlapping occurs when two or more rules in an APIRule configuration have paths that can match the same request URI and share at least one common method. This can lead to ambiguity about which rule should apply to a given request, especially if the rules also share common HTTP methods. 

    When defining the order of rules, you must take into account that the request is checked against each path, from top to bottom of your APIRule. If it matches the first path, the first rule applies. This is why rule order is important and you must define the paths starting from the most specific one, and leave the most general as the last one.

    > [!NOTE]
    > Rules defined earlier in the list have a higher priority than those defined later. The request searches for the first matching path starting from the top of the APIRule **spec.rules** list. Therefore, we recommend ordering rules starting with the most specific path and ending with the most general.

    The following list shows correct order of overlapping paths as each one encloses each other. They are listed from the most specific to the most general:
       - `/anything/one`
       - `/anything/one/two`
       - `/anything/{*}/{*}/two`
       - `/anything/{**}/two`
       - `/anything/`
       - `/anything/{**}`
       - `/{**}`

    See an example of listing paths in incorrect order like:
      - `/anything/{**}`
      - `/anything/one`
      - `/anything/one/two`
      - `/anything/{*}/{*}/two`
      - `/anything/{**}/two`
      - `/anything/`
      - `/{**}`

    For the purpose of this example, we assume that every path shares a common method with another rule. In this scenario, any path that starts with `/anything/` is matched with the first rule. As a result, subsequent rules that also begin with `/anything/` are never be evaluated. This is because the first rule acts as a catch-all for any path starting with `/anything/`, making the more specific rules redundant. The other rule that works is `/{**}` as it is the most general one that matches any path that does not start with `/anything/`. 

3. Check for excluding rules that share common methods.

    Understanding the relationship between paths and methods in a rule is crucial to avoid unexpected behavior. If a rule shares at least one common method with a preceding rule, then the path from preceding rule is excluded from this rule. 
    
    For example, the following APIRule configuration excludes the `POST` and `GET` methods for the path `/anything/one` with the `noAuth` access strategy. This happens because the rule with the path `/anything/{**}` shares at least one common method (`GET`) with the preceding rule.

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
    To use the `POST` method on the path `/anything/one` with the `noAuth` access strategy, you must define separate rules for overlapping methods and paths. See the following example:

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
