Feature: Exposing two endpoints with Istio JWT authorization strategy with scopes

  Scenario: IstioJWTScopes: Calling an endpoint secured with JWT with a valid token
    Given IstioJWTScopes: There is a deployment secured with JWT on path "/ip"
    Then IstioJWTScopes: Calling the "/ip" endpoint with a valid "JWT" token and valid scopes should result in status between 200 and 299
    And IstioJWTScopes: Calling the second "/ip" endpoint with a valid "JWT" token and invalid scopes should result in status between 400 and 403
