Feature: Exposing different endpoints with different handlers

  Scenario: Exposing httpbin endpoints with JWT and OAuth2
    Given OAuth2: There is a httpbin service
    And OAuth2: There is an endpoint secured with JWT on path "/ip" requiring scopes '["read", "write"]'
    And OAuth2: There is an endpoint secured with OAuth2 on path "/get" requiring scopes '["read", "write"]'
    When OAuth2: The APIRule is applied
    Then OAuth2: Calling the "/ip" endpoint without a token should result in status between 400 and 403
    And OAuth2: Calling the "/ip" endpoint with an invalid token should result in status between 400 and 403
    And OAuth2: Calling the "/ip" endpoint with a valid "Opaque" token with "scopes" "read" and "write" should result in status between 400 and 403
    And OAuth2: Calling the "/ip" endpoint with a valid "JWT" token with "scopes" "read" and "write" should result in status between 200 and 299
    And OAuth2: Calling the "/get" endpoint without a token should result in status between 400 and 403
    And OAuth2: Calling the "/get" endpoint with an invalid token should result in status between 400 and 403
    And OAuth2: Calling the "/get" endpoint with a valid "Opaque" token with "scopes" "read" and "write" should result in status between 200 and 299
    And OAuth2: Calling the "/get" endpoint with a valid "JWT" token with "scopes" "read" and "write" should result in status between 400 and 403
    And OAuth2: Teardown httpbin service
