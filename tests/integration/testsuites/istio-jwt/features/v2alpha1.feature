Feature: Exposing endpoints with Istio JWT and NoAuth with v2alpha1 APIRule

  Scenario: Calling a httpbin endpoint secured
    Given v2alpha1IstioJWT: There is a httpbin service
    And v2alpha1IstioJWT: There is an endpoint secured with JWT on path "/ip"
    When v2alpha1IstioJWT: The APIRule is applied
    Then v2alpha1IstioJWT: Calling the "/ip" endpoint without a token should result in status between 400 and 403
    And v2alpha1IstioJWT: Calling the "/ip" endpoint with an invalid token should result in status between 400 and 403
    And v2alpha1IstioJWT: Calling the "/ip" endpoint with a valid "JWT" token should result in status between 200 and 299
    And v2alpha1IstioJWT: Teardown httpbin service

  Scenario: Expose GET, POST method for "/anything" and only PUT for "/anything/put" with noAuth
    Given v2alpha1NoAuthHandler: There is a httpbin service
    When v2alpha1NoAuthHandler: The APIRule is applied
    Then v2alpha1NoAuthHandler: Calling the "/anything" endpoint with "GET" method with any token should result in status between 200 and 200
    Then v2alpha1NoAuthHandler: Calling the "/anything" endpoint with "POST" method with any token should result in status between 200 and 200
    And v2alpha1NoAuthHandler: Calling the "/anything" endpoint with "PUT" method with any token should result in status between 404 and 404
    And v2alpha1NoAuthHandler: Calling the "/anything/put" endpoint with "PUT" method with any token should result in status between 200 and 200
    And v2alpha1NoAuthHandler: Calling the "/anything/put" endpoint with "POST" method with any token should result in status between 404 and 404
    And v2alpha1NoAuthHandler: Teardown httpbin service

  Scenario: Expose GET method for "/anything" with noAuth and recover if conversion webhook certificate secret is rotated
    Given v2alpha1NoAuthHandlerRecover: There is a httpbin service
    And v2alpha1NoAuthHandlerRecover: The APIRule is applied
    And v2alpha1NoAuthHandlerRecover: Calling the "/anything" endpoint with "GET" method with any token should result in status between 200 and 200
    When v2alpha1NoAuthHandlerRecover: Certificate secret is reset
    Then v2alpha1NoAuthHandlerRecover: Certificate secret is rotated
    And v2alpha1NoAuthHandlerRecover: Calling the "/anything" endpoint with "GET" method with any token should result in status between 200 and 200
    And v2alpha1NoAuthHandlerRecover: Teardown httpbin service
