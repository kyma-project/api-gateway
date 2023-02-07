Feature: Exposing one endpoint with Istio JWT authorization strategy

  Scenario: Calling an httpbin endpoint secured
    Given Common: There is an endpoint secured with JWT on path "/ip"
    Then Common: Calling the "/ip" endpoint without a token should result in status between 400 and 403
    And Common: Calling the "/ip" endpoint with an invalid token should result in status between 400 and 403
    And Common: Calling the "/ip" endpoint with a valid "JWT" token should result in status between 200 and 299

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
