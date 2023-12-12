Feature: Exposing endpoints using noop

  Scenario: Noop: Only allowing GET method for endpoints
    Given Noop: There is a httpbin service
    When Noop: The APIRule is applied
    Then Noop: Calling the "/anything" endpoint with "GET" method with any token should result in status between 200 and 299
    And Noop: Calling the "/anything" endpoint with "POST" method with any token should result in status between 404 and 404
    And Noop: Teardown httpbin service
