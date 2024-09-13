Feature: Exposing endpoints with NoAuth

  Scenario: Calling a httpbin endpoint unsecured on all paths
    Given NoAuthWildcard: There is a httpbin service
    When NoAuthWildcard: The APIRule is applied
    Then NoAuthWildcard: Calling the "/ip" endpoint without a token should result in status between 200 and 200
    And NoAuthWildcard: Calling the "/status/200" endpoint without a token should result in status between 200 and 200
    And NoAuthWildcard: Calling the "/headers" endpoint without a token should result in status between 200 and 200
    And NoAuthWildcard: Teardown httpbin service
