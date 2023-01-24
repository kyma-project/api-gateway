Feature: Exposing one endpoint with Istio JWT authorization strategy

  Scenario: IstioJWT: Calling an endpoint secured with JWT with a valid token
    Given IstioJWT: There is a deployment secured with JWT on path "/ip"
    Then IstioJWT: Calling the "/ip" endpoint without a token should result in status between 400 and 403
    And IstioJWT: Calling the "/ip" endpoint with a invalid token should result in status between 400 and 403
    And IstioJWT: Calling the "/ip" endpoint with a valid "JWT" token should result in status between 200 and 299
