Feature: Exposing an unsecured endpoint

  Scenario: Unsecured: Calling an unsecured API endpoint
    Given Unsecured: There is a httpbin service
    When Unsecured: The APIRule is applied
    Then Unsecured: Calling the "/headers" endpoint with any token should result in status between 200 and 299
    And Unsecured: Calling the "/headers" endpoint without a token should result in status between 200 and 299
    And Unsecured: Teardown httpbin service
