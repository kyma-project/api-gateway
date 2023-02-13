Feature: Exposing endpoints with Istio JWT authorization strategy

  Scenario: Calling an httpbin endpoint secured
    Given Common: There is an endpoint secured with JWT on path "/ip"
    Then Common: Calling the "/ip" endpoint without a token should result in status between 400 and 403
    And Common: Calling the "/ip" endpoint with an invalid token should result in status between 400 and 403
    And Common: Calling the "/ip" endpoint with a valid "JWT" token should result in status between 200 and 299

  Scenario: Calling httpbin that has an endpoint secured by JWT and unrestricted endpoints
    Given JwtAndUnrestricted: There is an endpoint secured with JWT on path "/ip"
    And JwtAndUnrestricted: There is an endpoint with handler "allow" on path "/headers"
    And JwtAndUnrestricted: There is an endpoint with handler "noop" on path "/json"
    When JwtAndUnrestricted: The APIRule is applied
    Then JwtAndUnrestricted: Calling the "/ip" endpoint with a valid "JWT" token should result in status between 200 and 299
    And JwtAndUnrestricted: Calling the "/headers" endpoint without token should result in status between 200 and 299
    And JwtAndUnrestricted: Calling the "/json" endpoint without token should result in status between 200 and 299

  Scenario: Calling an httpbin endpoint secured with JWT that requires scopes claims
    Given Scopes: There is an endpoint secured with JWT on path "/ip" requiring scopes '["read", "write"]'
    And Scopes: There is an endpoint secured with JWT on path "/get" requiring scopes '["test", "write"]'
    And Scopes: There is an endpoint secured with JWT on path "/headers" requiring scopes '["read"]'
    When Scopes: The APIRule is applied
    Then Scopes: Calling the "/ip" endpoint with a valid "JWT" token with scope claims "read" and "write" should result in status between 200 and 299
    And Scopes: Calling the "/get" endpoint with a valid "JWT" token with scope claims "read" and "write" should result in status between 400 and 403
    And Scopes: Calling the "/headers" endpoint with a valid "JWT" token with scope claims "read" and "write" should result in status between 200 and 299

  Scenario: Calling an httpbin endpoint secured with JWT that requires aud claim
    Given Audiences: There is an endpoint secured with JWT on path "/get" requiring audiences '["https://example.com"]'
    And Audiences: There is an endpoint secured with JWT on path "/ip" requiring audiences '["https://example.com", "https://example.com/user"]'
    And Audiences: There is an endpoint secured with JWT on path "/headers" requiring audiences '["https://example.com", "https://example.com/admin"]'
    When Audiences: The APIRule is applied
    Then Audiences: Calling the "/get" endpoint with a valid "JWT" token with audiences "https://example.com" and "https://example.com/user" should result in status between 200 and 299
    And Audiences: Calling the "/ip" endpoint with a valid "JWT" token with audiences "https://example.com" and "https://example.com/user" should result in status between 200 and 299
    And Audiences: Calling the "/headers" endpoint with a valid "JWT" token with audiences "https://example.com" and "https://example.com/user" should result in status between 400 and 403


  Scenario: Endpoints secured by JWT should fallback to service defined on root level when there is no service defined on rule level
    Given ServiceFallback: There is an endpoint secured with JWT on path "/headers" with service definition
    And ServiceFallback: There is an endpoint secured with JWT on path "/ip"
    When ServiceFallback: The APIRule with service on root level is applied
    Then ServiceFallback: Calling the "/headers" endpoint with a valid "JWT" token should result in status between 200 and 299
    And ServiceFallback: Calling the "/ip" endpoint with a valid "JWT" token should result in status between 200 and 299
