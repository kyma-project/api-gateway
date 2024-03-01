Feature: Exposing unsecure API and then securing it with OAuth2

  Scenario: SecureToUnsecure: Securing an unsecured endpoint with OAuth2
    Given SecureToUnsecure: There is an httpbin application secured with OAuth2
    And SecureToUnsecure: Calling the "/headers" endpoint with a valid "OAuth2" token should result in status between 200 and 299
    When SecureToUnsecure: Update APIRule to expose the endpoint with noop strategy
    Then SecureToUnsecure: Calling the "/headers" endpoint without a token should result in status beetween 200 and 299
    And SecureToUnsecure: Calling the "/headers" endpoint with any token should result in status beetween 200 and 299
    And SecureToUnsecure: Teardown httpbin service
