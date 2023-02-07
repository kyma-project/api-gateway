Feature: Exposing one endpoint with Istio JWT authorization strategy

  Scenario: Calling an endpoint secured with JWT with a valid token
    Given Common: There is a deployment secured with JWT on path "/ip"
    Then Common: Calling the "/ip" endpoint without a token should result in status between 400 and 403
    And Common: Calling the "/ip" endpoint with an invalid token should result in status between 400 and 403
    And Common: Calling the "/ip" endpoint with a valid "JWT" token should result in status between 200 and 299

  Scenario: Calling an endpoint secured with JWT that requires JWT scope claims "read" and "write" with a valid token
    Given ScopesHappy: There is a deployment secured with JWT on path "/ip"
    Then ScopesHappy: Calling the "/ip" endpoint with a valid "JWT" token with scopes read and write should result in status between 200 and 299

  Scenario: Calling an endpoint secured with JWT that requires JWT scope claims "test" and "write" with a valid token
    Given ScopesUnhappy: There is a deployment secured with JWT on path "/ip"
    Then ScopesUnhappy: Calling the "/ip" endpoint with a valid "JWT" token with scopes read and write should result in status between 400 and 403

  Scenario: Calling an httpbin endpoint secured with JWT that requires aud claim
    Given Audiences: There is a endpoint secured with JWT on path "/get" requiring audiences '["https://example.com"]'
    And Audiences: There is a endpoint secured with JWT on path "/ip" requiring audiences '["https://example.com", "https://example.com/user"]'
    And Audiences: There is a endpoint secured with JWT on path "/headers" requiring audiences '["https://example.com", "https://example.com/admin"]'
    When Audiences: The APIRule is applied
    Then Audiences: Calling the "/get" endpoint with a valid "JWT" token with audiences "https://example.com" and "https://example.com/user" should result in status between 200 and 299
    And Audiences: Calling the "/ip" endpoint with a valid "JWT" token with audiences "https://example.com" and "https://example.com/user" should result in status between 200 and 299
    And Audiences: Calling the "/headers" endpoint with a valid "JWT" token with audiences "https://example.com" and "https://example.com/user" should result in status between 400 and 403
