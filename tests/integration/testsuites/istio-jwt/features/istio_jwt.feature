Feature: Exposing endpoints with Istio JWT authorization strategy

  Scenario: Calling a httpbin endpoint secured
    Given Common: There is a httpbin service
    And Common: There is an endpoint secured with JWT on path "/ip"
    When Common: The APIRule is applied
    Then Common: Calling the "/ip" endpoint without a token should result in status between 400 and 403
    And Common: Calling the "/ip" endpoint with an invalid token should result in status between 400 and 403
    And Common: Calling the "/ip" endpoint with a valid "JWT" token should result in status between 200 and 299
    And Common: Teardown httpbin service

  Scenario: Calling a httpbin endpoint secured on wildcard `/.*` path
    Given Regex: There is a httpbin service
    And Regex: There is an endpoint secured with JWT on path "/.*"
    When Regex: The APIRule is applied
    Then Regex: Calling the "/ip" endpoint without a token should result in status between 400 and 403
    And Regex: Calling the "/ip" endpoint with an invalid token should result in status between 400 and 403
    And Regex: Calling the "/ip" endpoint with a valid "JWT" token should result in status between 200 and 299
    Then Regex: Calling the "/headers" endpoint without a token should result in status between 400 and 403
    And Regex: Calling the "/headers" endpoint with an invalid token should result in status between 400 and 403
    And Regex: Calling the "/headers" endpoint with a valid "JWT" token should result in status between 200 and 299
    And Regex: Teardown httpbin service

  Scenario: Calling a httpbin endpoint secured on wildcard `/*` path
    Given Prefix: There is a httpbin service
    And Prefix: There is an endpoint secured with JWT on path "/*"
    When Prefix: The APIRule is applied
    Then Prefix: Calling the "/ip" endpoint without a token should result in status between 400 and 403
    And Prefix: Calling the "/ip" endpoint with an invalid token should result in status between 400 and 403
    And Prefix: Calling the "/ip" endpoint with a valid "JWT" token should result in status between 200 and 299
    Then Prefix: Calling the "/headers" endpoint without a token should result in status between 400 and 403
    And Prefix: Calling the "/headers" endpoint with an invalid token should result in status between 400 and 403
    And Prefix: Calling the "/headers" endpoint with a valid "JWT" token should result in status between 200 and 299
    And Prefix: Teardown httpbin service

  Scenario: Calling httpbin that has an endpoint secured by JWT and unrestricted endpoints
    Given JwtAndUnrestricted: There is a httpbin service
    When JwtAndUnrestricted: The APIRule is applied
    Then JwtAndUnrestricted: Calling the "/ip" endpoint with a valid "JWT" token should result in status between 200 and 299
    And JwtAndUnrestricted: Calling the "/headers" endpoint without token should result in status between 200 and 299
    And JwtAndUnrestricted: Calling the "/json" endpoint without token should result in status between 200 and 299
    And JwtAndUnrestricted: Teardown httpbin service

  Scenario: Calling a httpbin endpoint secured with JWT that requires scopes claims
    Given Scopes: There is a httpbin service
    When Scopes: The APIRule is applied
    Then Scopes: Calling the "/ip" endpoint with a valid "JWT" token with "scopes" "read" and "write" should result in status between 200 and 299
    And Scopes: Calling the "/get" endpoint with a valid "JWT" token with "scopes" "read" and "write" should result in status between 400 and 403
    And Scopes: Calling the "/headers" endpoint with a valid "JWT" token with "scopes" "read" and "write" should result in status between 200 and 299
    And Scopes: Teardown httpbin service

  Scenario: Calling a httpbin endpoint secured with JWT that requires aud claim
    Given Audiences: There is a httpbin service
    When Audiences: The APIRule is applied
    Then Audiences: Calling the "/get" endpoint with a valid "JWT" token with "audiences" "https://example.com" and "https://example.com/user" should result in status between 200 and 299
    And Audiences: Calling the "/ip" endpoint with a valid "JWT" token with "audiences" "https://example.com" and "https://example.com/user" should result in status between 200 and 299
    And Audiences: Calling the "/cache" endpoint with a valid "JWT" token with "audiences" "https://example.com" and "https://example.com/user" should result in status between 200 and 299
    And Audiences: Calling the "/headers" endpoint with a valid "JWT" token with "audiences" "https://example.com" and "https://example.com/user" should result in status between 400 and 403
    And Audiences: Teardown httpbin service

  Scenario: Endpoints secured by JWT should fallback to service defined on root level when there is no service defined on rule level
    Given ServiceFallback: There is a httpbin service
    When ServiceFallback: The APIRule with service on root level is applied
    Then ServiceFallback: Calling the "/headers" endpoint with a valid "JWT" token should result in status between 200 and 299
    And ServiceFallback: Calling the "/ip" endpoint with a valid "JWT" token should result in status between 200 and 299
    And ServiceFallback: Teardown httpbin service

  Scenario: Calling a httpbin endpoint secured with JWT in two namespaces
    Given TwoNamespaces: There is a httpbin service
    And TwoNamespaces: There are two namespaces with workload
    And TwoNamespaces: There is an endpoint secured with JWT on path "/get" in APIRule Namespace
    And TwoNamespaces: There is an endpoint secured with JWT on path "/hello" in different namespace
    When TwoNamespaces: The APIRule is applied
    Then TwoNamespaces: Calling the "/get" endpoint with a valid "JWT" token should result in status between 200 and 299
    And TwoNamespaces: Calling the "/hello" endpoint with a valid "JWT" token should result in status between 200 and 299
    And TwoNamespaces: Calling the "/get" endpoint without token should result in status between 400 and 403
    And TwoNamespaces: Calling the "/hello" endpoint without token should result in status between 400 and 403
    And TwoNamespaces: Teardown httpbin service

  Scenario: Exposing different services with same methods
    Given DiffSvcSameMethods: There is a httpbin service
    And DiffSvcSameMethods: There is a workload and service for httpbin and helloworld
    When DiffSvcSameMethods: The APIRule is applied
    Then DiffSvcSameMethods: Calling the "/headers" endpoint with a valid "JWT" token should result in status between 200 and 299
    And DiffSvcSameMethods: Calling the "/hello" endpoint with a valid "JWT" token should result in status between 200 and 299
    And DiffSvcSameMethods: Teardown httpbin service

  Scenario: Exposing a JWT secured endpoint with unavailable issuer and jwks URL
    Given JwtIssuerUnavailable: There is a httpbin service
    Given JwtIssuerUnavailable: There is an endpoint secured with JWT on path "/ip" with invalid issuer and jwks
    When JwtIssuerUnavailable: The APIRule is applied
    And JwtIssuerUnavailable: Calling the "/ip" endpoint with a valid "JWT" token should result in body containing "Jwt issuer is not configured"
    And JwtIssuerUnavailable: Teardown httpbin service

  Scenario: Exposing a JWT secured endpoint where issuer URL doesn't belong to jwks URL
    Given JwtIssuerJwksNotMatch: There is a httpbin service
    And JwtIssuerJwksNotMatch: There is an endpoint secured with JWT on path "/ip" with invalid issuer and jwks
    When JwtIssuerJwksNotMatch: The APIRule is applied
    And JwtIssuerJwksNotMatch: Calling the "/ip" endpoint with a valid "JWT" token should result in body containing "Jwt verification fails"
    And JwtIssuerJwksNotMatch: Teardown httpbin service

  Scenario: Calling a httpbin endpoint secured with different JWT token from headers
    Given JwtTokenFromHeaders: There is a httpbin service
    When JwtTokenFromHeaders: The APIRule is applied
    Then JwtTokenFromHeaders: Calling the "/headers" endpoint without a token should result in status between 400 and 403
    And JwtTokenFromHeaders: Calling the "/headers" endpoint with a valid "JWT" token from default header should result in status between 400 and 403
    And JwtTokenFromHeaders: Calling the "/headers" endpoint with a valid "JWT" token from header "x-jwt-token" and prefix "JwtToken" should result in status between 200 and 299
    And JwtTokenFromHeaders: Teardown httpbin service

  Scenario: Calling a httpbin endpoint secured with different JWT token from params
    Given JwtTokenFromParams: There is a httpbin service
    When JwtTokenFromParams: The APIRule is applied
    And JwtTokenFromParams: Calling the "/ip" endpoint without a token should result in status between 400 and 403
    And JwtTokenFromParams: Calling the "/ip" endpoint with a valid "JWT" token from default header should result in status between 400 and 403
    And JwtTokenFromParams: Calling the "/ip" endpoint with a valid "JWT" token from parameter "jwt_token" should result in status between 200 and 299
    And JwtTokenFromParams: Teardown httpbin service

  Scenario: Calling a helloworld endpoint with custom label selector service
    Given CustomLabelSelector: There is a helloworld service with custom label selector name "custom-name"
    And CustomLabelSelector: There is an endpoint secured with JWT on path "/hello"
    When CustomLabelSelector: The APIRule is applied
    Then CustomLabelSelector: Calling the "/hello" endpoint without a token should result in status between 400 and 403
    And CustomLabelSelector: Calling the "/hello" endpoint with a valid "JWT" token should result in status between 200 and 299
    And CustomLabelSelector: Teardown helloworld service
