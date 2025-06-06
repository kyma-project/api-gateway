Feature: Exposing different endpoints with different handlers

  Scenario: Exposing httpbin endpoints with JWT and OAuth2
    Given NoAuth_JWT: There is a httpbin service
    When NoAuth_JWT: The APIRules are applied
    Then NoAuth_JWT: Calling the "/ip" endpoint without a token should result in status between 400 and 403
    And NoAuth_JWT: Calling the "/ip" endpoint with an invalid token should result in status between 400 and 403
    And NoAuth_JWT: Calling the "/ip" endpoint with a valid "JWT" token with "scopes" "read" and "write" should result in status between 200 and 299
    And NoAuth_JWT: Calling the "/get" endpoint on prefix "httpbin2" without a token should result in status between 200 and 299
    # The other host should not be accessible on "/get" path
    And NoAuth_JWT: Calling the "/get" endpoint without a token should result in status between 404 and 404
    And NoAuth_JWT: Teardown httpbin service

  Scenario: Exposing httpbin endpoints with JWT and OAuth2
    Given OAuth2: There is a httpbin service
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