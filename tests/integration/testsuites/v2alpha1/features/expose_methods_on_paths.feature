Feature: Exposing specific methods on paths

  Scenario: ExposeMethodsOnPathsNoAuthHandler: Expose GET, POST method for "/anything" and only PUT for "/anything/put" with noAuth
    Given There is a httpbin service
    When The APIRule is applied
    Then Calling the "/anything" endpoint with "GET" method with any token should result in status between 200 and 200
    Then Calling the "/anything" endpoint with "POST" method with any token should result in status between 200 and 200
    And Calling the "/anything" endpoint with "PUT" method with any token should result in status between 404 and 404
    And Calling the "/anything/put" endpoint with "PUT" method with any token should result in status between 200 and 200
    And Calling the "/anything/put" endpoint with "POST" method with any token should result in status between 404 and 404
    And Teardown httpbin service

  Scenario: ExposeMethodsOnPathsJwtHandler: Expose GET, POST method for "/anything" and only PUT for "/anything/put" secured by JWT
    Given There is a httpbin service
    When The APIRule is applied
    Then Calling the "/anything" endpoint with "GET" method with a valid "JWT" token should result in status between 200 and 200
    Then Calling the "/anything" endpoint with "POST" method with a valid "JWT" token should result in status between 200 and 200
    And Calling the "/anything" endpoint with "PUT" method with a valid "JWT" token should result in status between 404 and 404
    And Calling the "/anything/put" endpoint with "PUT" method with a valid "JWT" token should result in status between 200 and 200
    And Calling the "/anything/put" endpoint with "POST" method with a valid "JWT" token should result in status between 404 and 404
    And Teardown httpbin service
