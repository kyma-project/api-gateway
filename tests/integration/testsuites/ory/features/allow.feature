Feature: Exposing endpoints using allow

  Scenario: Allow: Only allowing GET method for endpoints
    Given Allow: There is a httpbin service
    When Allow: The APIRule is applied
    Then Allow: Calling the "/anything" endpoint with "GET" method with any token should result in status between 200 and 299
    And Allow: Calling the "/anything" endpoint with "POST" method with any token should result in status between 404 and 404
    And Allow: Teardown httpbin service
