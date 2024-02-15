Feature: Exposing specific methods on paths

  Scenario: ExposeMethodsOnPathsAllowHandler: Expose GET, POST method for "/anything" and only PUT for "/anything/put" with allow access strategy
    Given ExposeMethodsOnPathsAllowHandler: There is a httpbin service
    When ExposeMethodsOnPathsAllowHandler: The APIRule is applied
    Then ExposeMethodsOnPathsAllowHandler: Calling the "/anything" endpoint with "GET" method with any token should result in status between 200 and 200
    Then ExposeMethodsOnPathsAllowHandler: Calling the "/anything" endpoint with "POST" method with any token should result in status between 200 and 200
    Then ExposeMethodsOnPathsAllowHandler: Calling the "/anything" endpoint with "DELETE" method with any token should result in status between 200 and 200
    Then ExposeMethodsOnPathsAllowHandler: Calling the "/anything" endpoint with "OPTIONS" method with any token should result in status between 200 and 200
    Then ExposeMethodsOnPathsAllowHandler: Calling the "/anything" endpoint with "PUT" method with any token should result in status between 200 and 200
    And ExposeMethodsOnPathsAllowHandler: Calling the "/anything/put" endpoint with "PUT" method with any token should result in status between 200 and 200
    And ExposeMethodsOnPathsAllowHandler: Calling the "/anything/put" endpoint with "POST" method with any token should result in status between 200 and 200
    And ExposeMethodsOnPathsAllowHandler: Teardown httpbin service

  Scenario: ExposeMethodsOnPathsNoAuthHandler: Expose GET, POST method for "/anything" and only PUT for "/anything/put" with no_auth access strategy
    Given ExposeMethodsOnPathsNoAuthHandler: There is a httpbin service
    When ExposeMethodsOnPathsNoAuthHandler: The APIRule is applied
    Then ExposeMethodsOnPathsNoAuthHandler: Calling the "/anything" endpoint with "GET" method with any token should result in status between 200 and 200
    Then ExposeMethodsOnPathsNoAuthHandler: Calling the "/anything" endpoint with "POST" method with any token should result in status between 200 and 200
    And ExposeMethodsOnPathsNoAuthHandler: Calling the "/anything" endpoint with "PUT" method with any token should result in status between 404 and 404
    And ExposeMethodsOnPathsNoAuthHandler: Calling the "/anything/put" endpoint with "PUT" method with any token should result in status between 200 and 200
    And ExposeMethodsOnPathsNoAuthHandler: Calling the "/anything/put" endpoint with "POST" method with any token should result in status between 404 and 404
    And ExposeMethodsOnPathsNoAuthHandler: Teardown httpbin service

  Scenario: ExposeMethodsOnPathsNoopHandler: Expose GET, POST method for "/anything" and only PUT for "/anything/put" with noop access strategy
    Given ExposeMethodsOnPathsNoopHandler: There is a httpbin service
    When ExposeMethodsOnPathsNoopHandler: The APIRule is applied
    Then ExposeMethodsOnPathsNoopHandler: Calling the "/anything" endpoint with "GET" method with any token should result in status between 200 and 200
    Then ExposeMethodsOnPathsNoopHandler: Calling the "/anything" endpoint with "POST" method with any token should result in status between 200 and 200
    And ExposeMethodsOnPathsNoopHandler: Calling the "/anything" endpoint with "PUT" method with any token should result in status between 404 and 404
    And ExposeMethodsOnPathsNoopHandler: Calling the "/anything/put" endpoint with "PUT" method with any token should result in status between 200 and 200
    And ExposeMethodsOnPathsNoopHandler: Calling the "/anything/put" endpoint with "POST" method with any token should result in status between 404 and 404
    And ExposeMethodsOnPathsNoopHandler: Teardown httpbin service

  Scenario: ExposeMethodsOnPathsJwtHandler: Expose GET, POST method for "/anything" and only PUT for "/anything/put" with jwt access strategy
    Given ExposeMethodsOnPathsJwtHandler: There is a httpbin service
    When ExposeMethodsOnPathsJwtHandler: The APIRule is applied
    Then ExposeMethodsOnPathsJwtHandler: Calling the "/anything" endpoint with "GET" method with a valid "JWT" token should result in status between 200 and 200
    Then ExposeMethodsOnPathsJwtHandler: Calling the "/anything" endpoint with "POST" method with a valid "JWT" token should result in status between 200 and 200
    And ExposeMethodsOnPathsJwtHandler: Calling the "/anything" endpoint with "PUT" method with a valid "JWT" token should result in status between 404 and 404
    And ExposeMethodsOnPathsJwtHandler: Calling the "/anything/put" endpoint with "PUT" method with a valid "JWT" token should result in status between 200 and 200
    And ExposeMethodsOnPathsJwtHandler: Calling the "/anything/put" endpoint with "POST" method with a valid "JWT" token should result in status between 404 and 404
    And ExposeMethodsOnPathsJwtHandler: Teardown httpbin service

  Scenario: ExposeMethodsOnPathsOAuth2Handler: Expose GET, POST method for "/anything" and only PUT for "/anything/put" with oauth2_introspection access strategy
    Given ExposeMethodsOnPathsOAuth2Handler: There is a httpbin service
    When ExposeMethodsOnPathsOAuth2Handler: The APIRule is applied
    Then ExposeMethodsOnPathsOAuth2Handler: Calling the "/anything" endpoint with "GET" method with a valid "Opaque" token should result in status between 200 and 200
    Then ExposeMethodsOnPathsOAuth2Handler: Calling the "/anything" endpoint with "POST" method with a valid "Opaque" token should result in status between 200 and 200
    And ExposeMethodsOnPathsOAuth2Handler: Calling the "/anything" endpoint with "PUT" method with a valid "Opaque" token should result in status between 404 and 404
    And ExposeMethodsOnPathsOAuth2Handler: Calling the "/anything/put" endpoint with "PUT" method with a valid "Opaque" token should result in status between 200 and 200
    And ExposeMethodsOnPathsOAuth2Handler: Calling the "/anything/put" endpoint with "POST" method with a valid "Opaque" token should result in status between 404 and 404
    And ExposeMethodsOnPathsOAuth2Handler: Teardown httpbin service