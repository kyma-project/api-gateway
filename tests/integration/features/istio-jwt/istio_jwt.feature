Feature: Exposing one endpoint with Istio JWT authorization strategy

  Scenario: Calling an httpbin endpoint secured with JWT in two namespaces
    Given TwoNamespaces: There are two namespaces with workload
    And TwoNamespaces: There is an endpoint secured with JWT on path "/get" in APIRule Namespace
    And TwoNamespaces: There is an endpoint secured with JWT on path "/hello" in different namespace
    When TwoNamespaces: The APIRule is applied
    Then TwoNamespaces: Calling the "/get" endpoint with a valid "JWT" token should result in status between 200 and 299
    And TwoNamespaces: Calling the "/hello" endpoint with a valid "JWT" token should result in status between 200 and 299
