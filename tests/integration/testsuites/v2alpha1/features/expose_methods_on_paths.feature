Feature: Exposing specific methods on paths

  Scenario: ExposeMethodsOnPathsNoAuthHandler: Expose GET, POST method for "/anything" and only PUT for "/anything/put" with noAuth
    Given ExposeMethodsOnPathsNoAuthHandler: There is a httpbin service
    When ExposeMethodsOnPathsNoAuthHandler: The APIRule is applied
    Then ExposeMethodsOnPathsNoAuthHandler: Calling the "/anything" endpoint with "GET" method with any token should result in status between 200 and 200
    Then ExposeMethodsOnPathsNoAuthHandler: Calling the "/anything" endpoint with "POST" method with any token should result in status between 200 and 200
    And ExposeMethodsOnPathsNoAuthHandler: Calling the "/anything" endpoint with "PUT" method with any token should result in status between 404 and 404
    And ExposeMethodsOnPathsNoAuthHandler: Calling the "/anything/put" endpoint with "PUT" method with any token should result in status between 200 and 200
    And ExposeMethodsOnPathsNoAuthHandler: Calling the "/anything/put" endpoint with "POST" method with any token should result in status between 404 and 404
    And ExposeMethodsOnPathsNoAuthHandler: Teardown httpbin service

  Scenario: ExposeMethodsOnPathsJwtHandler: Expose GET, POST method for "/anything" and only PUT for "/anything/put" secured by JWT
    Given ExposeMethodsOnPathsJwtHandler: There is a httpbin service
    When ExposeMethodsOnPathsJwtHandler: The APIRule is applied
    Then ExposeMethodsOnPathsJwtHandler: Calling the "/anything" endpoint with "GET" method with a valid "JWT" token should result in status between 200 and 200
    Then ExposeMethodsOnPathsJwtHandler: Calling the "/anything" endpoint with "POST" method with a valid "JWT" token should result in status between 200 and 200
    And ExposeMethodsOnPathsJwtHandler: Calling the "/anything" endpoint with "PUT" method with a valid "JWT" token should result in status between 404 and 404
    And ExposeMethodsOnPathsJwtHandler: Calling the "/anything/put" endpoint with "PUT" method with a valid "JWT" token should result in status between 200 and 200
    And ExposeMethodsOnPathsJwtHandler: Calling the "/anything/put" endpoint with "POST" method with a valid "JWT" token should result in status between 404 and 404
    And ExposeMethodsOnPathsJwtHandler: Teardown httpbin service
