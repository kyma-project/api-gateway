# Ordering Rules in APIRule `v2`
To ensure your service requests are handled as intended, learn how to create and order rules in your APIRule custom resource (CR).

## Context
APIRule allows you to define a list of rules that specify how requests to your service are handled. You can define a list of rules in the **spec.rules** field of the APIRule CR. Each rule starts with a hyphen and must contain the following details: 
- The path on which the service is exposed
- The list of HTTP methods available for this path
- The specified access strategy

If your APIRule includes multiple rules, their order matters. Follow these steps when creating and ordering the rules to make sure you avoid path conflicts.

## Procedure
1. Specify the paths.

   To define the paths for each rule, use one of the following approaches. You can specify exact path names or, to match multiple paths, use the operators `{*}`, `{**}`, or the `/*` wildcard.

   - Specify the exact path name.

     For example, `/example/one` specifies the exact path `/example/one`, and `/` specifies the root path.

   - Use the operators `{*}`, `{**}`, and the `/*` wildcard.
        Expression | Description | Examples
        ---------|----------|---------
        `{*}` | It matches a single path component, up to the next path separator: `/`. | <ul><li>`/example/{*}/one` - matches requests with the path prefix `example`, exactly one additional segment in the middle, and the path suffix `one`. For example, possible matches include `/example/anything/one`. </li><li>`/example/{*}` - matches requests with path prefix `example` and exactly one other segment. For example, possible matches include: `/example/anything`. The paths `/example/` and `/example/anything/` don't match.</li></ul> 
        `{**}` | If it is the last element of the path, it matches zero or more path segments. <br> If it is not the last element of a path, it matches one or more path segments. <br> If multiple operators are present, `{**}` must be the last one. | <ul><li>`/example/{**}/one` - matches requests with the path prefix `example`, one or more additional segments in the middle, and the path suffix `one`. For example, possible matches include `/example/anything/two/one` and `/example/anything/one`. The paths `/example//one` and `/example/one` don't match. </li><li>`/example/{**}` - matches requests with the path prefix `example` followed by zero, or more additional segments. For example, possible matches include `/example/anything`, `/example/anything/more/`, and `/example/`. </li><li>`/{*}/example/{*}/{**}` - matches with not empty segment at the beginning, the path `example`, and at least one additional path segment. For example, possible matches include `/anything/example/anything/` and `/anything/example/anything/more`.</li></ul> 
        `/*` | It matches all paths. It is equivalent to the path `/{**}`. If defined, `/*` must be the only element in the entire path. You cannot combine the wildcard path with any other operators or path segments. It was introduced for backward compatibility. | <ul><li>`/*` - matches `/` and any other path, such as `/example/anything/more/` or `/example/`.</li></ul> 

   > [!NOTE]
   > To be a valid path template, the path must not contain `*`, `{`, or `}` outside of the supported operators or the `/*` wildcard. No other characters are allowed in the path segment with the path template operator or the wildcard.
   >

   Using the wildcard and the operators introduces the possibility of path conflicts. A path conflict occurs when two or more APIRule **spec.rules** match the same path and share at least one common HTTP method. This is why it is important to consider the order of rules and to understand the connection between rules based on the path prefix and shared HTTP methods. Knowing the expected outcome of a configured rule helps organize and sort the rules.

2. Group paths into rules based on common HTTP methods and access strategies.

   Specify HTTP methods that you want to allow for each path. If you want to allow more than one method for a path, you can use one of these approaches:
   - Group multiple HTTP methods in a single rule. This is useful when you want to apply the same access strategy to multiple HTTP methods for the same endpoint. 
    
     For example, to allow both GET and POST requests to all service endpoints, you specify the `/{**}` operator in the path without authentication. In such a case, you can define a single rule with both methods:

     ```yaml
     ...
       rules:
         - path: /{**}
           methods: 
             - GET
             - POST
           noAuth: true
     ```
   - Create a separate rule for each HTTP method. Use this approach if you want to apply different access strategies or configurations to each method.
     
     In this example, the outcome is the same as grouping multiple HTTP methods in one rule. You still allow both GET and POST requests to all service endpoints without authentication.
     
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
    > You can group multiple HTTP methods in a single rule if they share the same access strategy, or create separate rules for each method to allow for applying different configurations. Both approaches are valid.

3. Order the rules.

    Search for the paths that overlap. Overlapping occurs when two or more rules in an APIRule configuration contain paths that can match the same request and share at least one common HTTP method. This might lead to ambiguity about which rule applies to a given request. 

    When defining the order of rules, remember that each request is evaluated against the list of paths from top to bottom in your APIRule, and the first matching rule is applied.

    If a request matches the first path, the first rule applies. This is why you must define the rules starting from the most specific path, and leave the most general path as the last one.

    > [!NOTE]
    > Rules defined earlier in the list have a higher priority than those defined later. The request searches for the first matching rule starting from the top of the APIRule **spec.rules** list. Therefore, it's recommended to order rules from the most specific path to the most general.

    See the following example of an incorrect rules' order:
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
   Applying this configuration causes an APIRule to be in `Error` state with the description: 
    ```bash 
    Validation errors: Attribute '.spec.rules': Path /anything/{*}/one with method POST conflicts with at least one of the previous rule paths
    ```
   
   In the following APIRule, the first rule specifically matches requests to the path `/anything/{*}/one` with the POST method, requiring JWT authentication. The second rule acts as a catch-all for any other paths that start with `/anything/`, allowing unauthenticated POST and GET requests. By placing the more specific rule first, you ensure that requests to `/anything/{*}/one` are handled as intended, while all other matching paths are covered by the more general rule. This approach prevents the more general rule from overshadowing the specific one and ensures the correct access strategy is applied to each path. 
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
    You can test the above APIRule configuration using the following requests:

   | Request                                                                                                   | Rule Matched                | Access Strategy         | Expected Outcome                                              | HTTP Status Code         |
   |-----------------------------------------------------------------------------------------------------------|-----------------------------|-------------------------|---------------------------------------------------------------|--------------------------|
   | `curl -ik -X GET https://{SUBDOMAIN}.{DOMAIN_NAME}/anything/more`                                         | Second rule (`/anything/{**}`) | noAuth                  | Unauthenticated GET to `/anything/more` allowed               | `200 OK`                 |
   | `curl -ik -X POST https://{SUBDOMAIN}.{DOMAIN_NAME}/anything/more`                                        | Second rule (`/anything/{**}`) | noAuth                  | Unauthenticated POST to `/anything/more` allowed              | `200 OK`                 |
   | `curl -ik -X POST https://{SUBDOMAIN}.{DOMAIN_NAME}/anything/more/one --header "Authorization: Bearer <valid_jwt_token>"` | First rule (`/anything/{*}/one`) | JWT required            | Authenticated POST to `/anything/more/one` allowed            | `200 OK`                 |
   | `curl -ik -X POST https://{SUBDOMAIN}.{DOMAIN_NAME}/anything/more/one`                                    | First rule (`/anything/{*}/one`) | JWT required            | Unauthenticated POST to `/anything/more/one` denied           | `403 RBAC: access denied`|
   | `curl -ik -X GET https://{SUBDOMAIN}.{DOMAIN_NAME}/anything/more/one`                                     | No matching rule            | -                       | GET to `/anything/more/one` not allowed (no rule matches)     | `403 RBAC: access denied`|
4. Check for excluding rules that share common methods.

   > [!NOTE] 
   > Understanding the relationship between paths and methods in a rule is crucial to avoid unexpected behavior. If a rule shares at least one common method with a preceding rule, then the path from preceding rule is excluded from this rule. 
    
    For example, the following APIRule configuration excludes the POST and GET methods for the path `/anything/{*}/one` with the **noAuth** access strategy. This happens because the rule with the path `/anything/{**}` shares one common HTTP method POST with the preceding rule with the path `/anything/{*}/one`.

    ```yaml
    ...
      rules:
        - methods:
          - POST
          jwt:
            authentications:
              - issuer: https://example.com
                jwksUri: https://example.com/.well-known/jwks.json
          path: /anything/{*}/one
        - methods:
          - GET
          - POST
          noAuth: true
          path: /anything/{**}
    ```

   This outcome might be unexpected if you intended to allow GET requests to `/anything/{*}/one` without authentication. To achieve that, you must specifically define separate rules for overlapping methods and paths. 

    The following APIRule configuration allows POST requests to `/anything/{*}/one` with JWT authentication, while permitting unauthenticated POST requests to all paths starting with `/anything/` except for the `/anything/{*}/one` endpoint. Additionally, it allows unauthenticated GET requests to all paths prefixed with `/anything/`. See the following example:

    ```yaml
    ...
      rules:
        - methods:
            - POST
          jwt:
            authentications:
              - issuer: https://example.com
                jwksUri: https://example.com/.well-known/jwks.json
          path: /anything/{*}/one
        - methods:
            - POST
          noAuth: true
          path: /anything/{**}
        - methods:
            - GET
          noAuth: true
          path: /anything/{**}
    ```
   
    You can test the above APIRule configuration using the following requests:

| Request                                                                                                   | Rule Matched                     | Access Strategy | Expected Outcome                                     | HTTP Status Code         |
|-----------------------------------------------------------------------------------------------------------|----------------------------------|-----------------|------------------------------------------------------|--------------------------|
| `curl -ik -X GET https://{SUBDOMAIN}.{DOMAIN_NAME}/anything/more`                                         | Third rule (`/anything/{**}`)    | noAuth          | Unauthenticated GET to `/anything/more` allowed      | `200 OK`                 |
| `curl -ik -X POST https://{SUBDOMAIN}.{DOMAIN_NAME}/anything/more`                                        | Second rule (`/anything/{**}`)   | noAuth          | Unauthenticated POST to `/anything/more` allowed     | `200 OK`                 |
| `curl -ik -X POST https://{SUBDOMAIN}.{DOMAIN_NAME}/anything/more/one --header "Authorization: Bearer <valid_jwt_token>"` | First rule (`/anything/{*}/one`) | JWT required    | Authenticated POST to `/anything/more/one` allowed   | `200 OK`                 |
| `curl -ik -X POST https://{SUBDOMAIN}.{DOMAIN_NAME}/anything/more/one`                                    | First rule (`/anything/{*}/one`) | JWT required    | Unauthenticated POST to `/anything/more/one` denied  | `403 RBAC: access denied`|
| `curl -ik -X GET https://{SUBDOMAIN}.{DOMAIN_NAME}/anything/more/one`                                     | Third rule (`/anything/{**}`)    | noAuth          | Unauthenticated GET to `/anything/more/one` allowed  | `200 OK`                 |