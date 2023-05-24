Feature: Exposing unsecure API and then securing it with JWT and OAuth2 on two paths

  Scenario: UnsecureToSecure: Securing an unsecured API with OAuth2 and JWT
    Given UnsecureToSecure: There is an unsecured API with all paths available without authorization
    And UnsecureToSecure: Calling the "/headers" endpoint without a token should result in status between 200 and 299
    And UnsecureToSecure: Calling the "/image" endpoint without a token should result in status between 200 and 299
    When UnsecureToSecure: API is secured with OAuth2 on path /headers and JWT on path /image
    Then UnsecureToSecure: Calling the "/headers" endpoint without a token should result in status between 400 and 403
    And UnsecureToSecure: Calling the "/image" endpoint without a token should result in status between 400 and 403
    And UnsecureToSecure: Calling the "/headers" endpoint with an invalid token should result in status between 400 and 403
    And UnsecureToSecure: Calling the "/image" endpoint with an invalid token should result in status between 400 and 403
    And UnsecureToSecure: Calling the "/headers" endpoint with a valid "OAuth2" token should result in status between 200 and 299
    And UnsecureToSecure: Calling the "/image" endpoint with a valid "JWT" token should result in status between 200 and 299
    And UnsecureToSecure: Teardown httpbin service
