Feature: Exposing endpoints with NoAuth

  Scenario: Calling a httpbin endpoint unsecured on all paths
    Given Wildcard: There is a httpbin service
    When Wildcard: The APIRule is applied
    Then Wildcard: Calling the "/ip" endpoint without a token should result in status between 200 and 200
    And Wildcard: Calling the "/status/200" endpoint without a token should result in status between 200 and 200
    And Wildcard: Calling the "/headers" endpoint without a token should result in status between 200 and 200
    And Wildcard: Teardown httpbin service
