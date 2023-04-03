Feature: Exposing endpoints with Istio JWT authorization strategy

  Scenario: Calling a httpbin endpoint secured
    Given Common: There is a httpbin service
    When Common: Common: The APIRule is applied
    Then Common: Calling the "/ip" endpoint without a token should result in status between 400 and 403
    And Common: Calling the "/ip" endpoint with an invalid token should result in status between 400 and 403
    And Common: Calling the "/ip" endpoint with a valid "JWT" token should result in status between 200 and 299

  Scenario: Calling httpbin that has an endpoint secured by JWT and unrestricted endpoints
    Given JwtAndUnrestricted: There is a httpbin service
    And JwtAndUnrestricted: There is an endpoint secured with JWT on path "/ip"
    And JwtAndUnrestricted: There is an endpoint with handler "allow" on path "/headers"
    And JwtAndUnrestricted: There is an endpoint with handler "noop" on path "/json"
    When JwtAndUnrestricted: The APIRule is applied
    Then JwtAndUnrestricted: Calling the "/ip" endpoint with a valid "JWT" token should result in status between 200 and 299
    And JwtAndUnrestricted: Calling the "/headers" endpoint without token should result in status between 200 and 299
    And JwtAndUnrestricted: Calling the "/json" endpoint without token should result in status between 200 and 299

  Scenario: Calling a httpbin endpoint secured with JWT that requires scopes claims
    Given Scopes: There is a httpbin service
    And Scopes: There is an endpoint secured with JWT on path "/ip" requiring scopes '["read", "write"]'
    And Scopes: There is an endpoint secured with JWT on path "/get" requiring scopes '["test", "write"]'
    And Scopes: There is an endpoint secured with JWT on path "/headers" requiring scopes '["read"]'
    When Scopes: The APIRule is applied
    Then Scopes: Calling the "/ip" endpoint with a valid "JWT" token with scope claims "read" and "write" should result in status between 200 and 299
    And Scopes: Calling the "/get" endpoint with a valid "JWT" token with scope claims "read" and "write" should result in status between 400 and 403
    And Scopes: Calling the "/headers" endpoint with a valid "JWT" token with scope claims "read" and "write" should result in status between 200 and 299

  Scenario: Calling a httpbin endpoint secured with JWT that requires aud claim
    Given Audiences: There is a httpbin service
    And Audiences: There is an endpoint secured with JWT on path "/get" requiring audiences '["https://example.com"]'
    And Audiences: There is an endpoint secured with JWT on path "/ip" requiring audiences '["https://example.com", "https://example.com/user"]'
    And Audiences: There is an endpoint secured with JWT on path "/cache" requiring audience '["https://example.com"]' or '["audienceNotInJWT"]'
    And Audiences: There is an endpoint secured with JWT on path "/headers" requiring audiences '["https://example.com", "https://example.com/admin"]'
    When Audiences: The APIRule is applied
    Then Audiences: Calling the "/get" endpoint with a valid "JWT" token with audiences "https://example.com" and "https://example.com/user" should result in status between 200 and 299
    And Audiences: Calling the "/ip" endpoint with a valid "JWT" token with audiences "https://example.com" and "https://example.com/user" should result in status between 200 and 299
    And Audiences: Calling the "/cache" endpoint with a valid "JWT" token with audiences "https://example.com" and "https://example.com/user" should result in status between 200 and 299
    And Audiences: Calling the "/headers" endpoint with a valid "JWT" token with audiences "https://example.com" and "https://example.com/user" should result in status between 400 and 403

  Scenario: Endpoints secured by JWT should fallback to service defined on root level when there is no service defined on rule level
    Given ServiceFallback: There is a httpbin service
    And ServiceFallback: There is an endpoint secured with JWT on path "/headers" with service definition
    And ServiceFallback: There is an endpoint secured with JWT on path "/ip"
    When ServiceFallback: The APIRule with service on root level is applied
    Then ServiceFallback: Calling the "/headers" endpoint with a valid "JWT" token should result in status between 200 and 299
    And ServiceFallback: Calling the "/ip" endpoint with a valid "JWT" token should result in status between 200 and 299

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

  Scenario: Exposing different services with same methods
    Given DiffSvcSameMethods: There is a httpbin service
    And DiffSvcSameMethods: There is a workload and service for httpbin and helloworld
    And DiffSvcSameMethods: There is an endpoint secured with JWT on path "/headers" for httpbin service with methods '["GET", "POST"]'
    And DiffSvcSameMethods: There is an endpoint secured with JWT on path "/hello" for helloworld service with methods '["GET", "POST"]'
    When DiffSvcSameMethods: The APIRule is applied
    Then DiffSvcSameMethods: Calling the "/headers" endpoint with a valid "JWT" token should result in status between 200 and 299
    And DiffSvcSameMethods: Calling the "/hello" endpoint with a valid "JWT" token should result in status between 200 and 299

  Scenario: Exposing a JWT secured endpoint with unavailable issuer and jwks URL
    Given JwtIssuerUnavailable: There is a httpbin service
    Given JwtIssuerUnavailable: There is an endpoint secured with JWT on path "/ip" with invalid issuer and jwks
    When JwtIssuerUnavailable: The APIRule is applied
    And JwtIssuerUnavailable: Calling the "/ip" endpoint with a valid "JWT" token should result in body containing "Jwt issuer is not configured"

  Scenario: Exposing a JWT secured endpoint where issuer URL doesn't belong to jwks URL
    Given JwtIssuerJwksNotMatch: There is a httpbin service
    And JwtIssuerJwksNotMatch: There is an endpoint secured with JWT on path "/ip" with invalid issuer and jwks
    When JwtIssuerJwksNotMatch: The APIRule is applied
    And JwtIssuerJwksNotMatch: Calling the "/ip" endpoint with a valid "JWT" token should result in body containing "Jwks doesn't have key to match kid or alg from Jwt"

  Scenario: Calling a httpbin endpoint secured with different JWT token from options
    Given JwtTokenFrom: There is a httpbin service
    When JwtTokenFrom: Common: The APIRule is applied
    Then JwtTokenFrom: Calling the "/headers" endpoint without a token should result in status between 400 and 403
    And JwtTokenFrom: Calling the "/headers" endpoint with a valid "JWT" token from header "x-jwt-token" and prefix "JwtToken" should result in status between 200 and 299
    And JwtTokenFrom: Calling the "/ip" endpoint without a token should result in status between 400 and 403
    And JwtTokenFrom: Calling the "/ip" endpoint with a valid "JWT" token from parameter "jwt_token" should result in status between 200 and 299
