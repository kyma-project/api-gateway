Feature: Exposing endpoints with Istio JWT and NoAuth with v2 APIRule

  Scenario: Calling a httpbin endpoint secured
    Given v2IstioJWT: There is a httpbin service
    And v2IstioJWT: There is an endpoint secured with JWT on path "/ip"
    When v2IstioJWT: The APIRule is applied
    Then v2IstioJWT: Calling the "/ip" endpoint without a token should result in status between 400 and 403
    And v2IstioJWT: Calling the "/ip" endpoint with an invalid token should result in status between 400 and 403
    And v2IstioJWT: Calling the "/ip" endpoint with a valid "JWT" token should result in status between 200 and 299
    And v2IstioJWT: Teardown httpbin service

  Scenario: Expose GET, POST method for "/anything" and only PUT for "/anything/put" with noAuth
    Given v2NoAuthHandler: There is a httpbin service
    When v2NoAuthHandler: The APIRule is applied
    Then v2NoAuthHandler: Calling the "/anything" endpoint with "GET" method with any token should result in status between 200 and 200
    Then v2NoAuthHandler: Calling the "/anything" endpoint with "POST" method with any token should result in status between 200 and 200
    And v2NoAuthHandler: Calling the "/anything" endpoint with "PUT" method with any token should result in status between 404 and 404
    And v2NoAuthHandler: Calling the "/anything/put" endpoint with "PUT" method with any token should result in status between 200 and 200
    And v2NoAuthHandler: Calling the "/anything/put" endpoint with "POST" method with any token should result in status between 404 and 404
    And v2NoAuthHandler: Teardown httpbin service

  Scenario: Expose GET method for "/anything" with noAuth and recover if conversion webhook certificate secret is rotated
    Given v2NoAuthHandlerRecover: There is a httpbin service
    And v2NoAuthHandlerRecover: The APIRule is applied
    And v2NoAuthHandlerRecover: Calling the "/anything" endpoint with "GET" method with any token should result in status between 200 and 200
    When v2NoAuthHandlerRecover: Certificate secret is reset
    Then v2NoAuthHandlerRecover: Certificate secret is rotated
    And v2NoAuthHandlerRecover: Calling the "/anything" endpoint with "GET" method with any token should result in status between 200 and 200
    And v2NoAuthHandlerRecover: Teardown httpbin service
