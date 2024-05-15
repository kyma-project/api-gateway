Feature: Exposing endpoints with Istio JWT and NoAuth with v1beta2 APIRule

  Scenario: Calling a httpbin endpoint secured
    Given v1beta2IstioJWT: There is a httpbin service
    And v1beta2IstioJWT: There is an endpoint secured with JWT on path "/ip"
    When v1beta2IstioJWT: The APIRule is applied
    Then v1beta2IstioJWT: Calling the "/ip" endpoint without a token should result in status between 400 and 403
    And v1beta2IstioJWT: Calling the "/ip" endpoint with an invalid token should result in status between 400 and 403
    And v1beta2IstioJWT: Calling the "/ip" endpoint with a valid "JWT" token should result in status between 200 and 299
    And v1beta2IstioJWT: Teardown httpbin service

  Scenario: Expose GET, POST method for "/anything" and only PUT for "/anything/put" with noAuth
    Given v1beta2NoAuthHandler: There is a httpbin service
    When v1beta2NoAuthHandler: The APIRule is applied
    Then v1beta2NoAuthHandler: Calling the "/anything" endpoint with "GET" method with any token should result in status between 200 and 200
    Then v1beta2NoAuthHandler: Calling the "/anything" endpoint with "POST" method with any token should result in status between 200 and 200
    And v1beta2NoAuthHandler: Calling the "/anything" endpoint with "PUT" method with any token should result in status between 404 and 404
    And v1beta2NoAuthHandler: Calling the "/anything/put" endpoint with "PUT" method with any token should result in status between 200 and 200
    And v1beta2NoAuthHandler: Calling the "/anything/put" endpoint with "POST" method with any token should result in status between 404 and 404
    And v1beta2NoAuthHandler: Teardown httpbin service

  Scenario: Expose GET method for "/anything" with noAuth and recover if conversion webhook certificate secret is rotated
    Given v1beta2NoAuthHandlerRecover: There is a httpbin service
    And v1beta2NoAuthHandlerRecover: The APIRule is applied
    And v1beta2NoAuthHandlerRecover: Calling the "/anything" endpoint with "GET" method with any token should result in status between 200 and 200
    When v1beta2NoAuthHandlerRecover: Certificate secret is reset
    Then v1beta2NoAuthHandlerRecover: Certificate secret is rotated
    And v1beta2NoAuthHandlerRecover: Calling the "/anything" endpoint with "GET" method with any token should result in status between 200 and 200
    And v1beta2NoAuthHandlerRecover: Teardown httpbin service
