Feature: Exposing one endpoint with Istio JWT authorization strategy

  Scenario: Calling an httpbin endpoint secured with JWT that requires aud claim
    Given Audiences: There is a endpoint secured with JWT on path "/get" requiring audiences '["https://example.com"]'
    And Audiences: There is a endpoint secured with JWT on path "/ip" requiring audiences '["https://example.com", "https://example.com/user"]'
    And Audiences: There is a endpoint secured with JWT on path "/headers" requiring audiences '["https://example.com", "https://example.com/admin"]'
    When Audiences: The APIRule is applied
    Then Audiences: Calling the "/get" endpoint with a valid "JWT" token with audiences "https://example.com" and "https://example.com/user" should result in status between 200 and 299
    And Audiences: Calling the "/ip" endpoint with a valid "JWT" token with audiences "https://example.com" and "https://example.com/user" should result in status between 200 and 299
    And Audiences: Calling the "/headers" endpoint with a valid "JWT" token with audiences "https://example.com" and "https://example.com/user" should result in status between 400 and 403
