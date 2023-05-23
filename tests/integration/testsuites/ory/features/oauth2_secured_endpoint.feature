Feature: Exposing an endpoint with OAuth2

  Scenario: OAuth2: Exposing an endpoint with OAuth2
    Given OAuth2: There is an endpoint secured with OAuth2 introspection
    Then OAuth2: Calling the "/image" endpoint without a token should result in status between 400 and 403
    And OAuth2: Calling the "/image" endpoint with a invalid token should result in status between 400 and 403
    And OAuth2: Calling the "/image" endpoint with a valid "OAuth2" token should result in status between 200 and 299
    And OAuth2: Teardown httpbin service