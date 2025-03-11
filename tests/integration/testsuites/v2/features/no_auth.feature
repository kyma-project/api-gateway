Feature: Exposing endpoints with NoAuth

  Scenario: Calling a httpbin endpoint unsecured on all paths
    Given The APIRule template file is set to "no-auth-wildcard.yaml"
    And There is a httpbin service
    When The APIRule is applied
    Then Calling the "/ip" endpoint without a token should result in status between 200 and 200
    And Calling the "/status/200" endpoint without a token should result in status between 200 and 200
    And Calling the "/headers" endpoint without a token should result in status between 200 and 200
    And Teardown httpbin service

  Scenario: In-cluster calling a httpbin endpoint unsecured
    Given The APIRule template file is set to "no-auth-wildcard.yaml"
    And There is a httpbin service
    When The APIRule is applied
    Then In-cluster calling the "/status/200" endpoint without a token should succeed
    And In-cluster calling the "/headers" endpoint without a token should succeed
