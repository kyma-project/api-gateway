Feature: Exposing endpoints with NoAuth
# apply v2 (noauth)
  Scenario: Calling a httpbin endpoint unsecured on all paths
    Given The APIRule template file is set to "no-auth-wildcard.yaml"
    And There is a httpbin service
    When The APIRule is applied
    Then Calling the "/ip" endpoint without a token should result in status between 200 and 200
    And Calling the "/status/200" endpoint without a token should result in status between 200 and 200
    And Calling the "/headers" endpoint without a token should result in status between 200 and 200
    And The APIRule contains original-version annotation set to "v2"
    And Resource of Kind "AuthorizationPolicy" owned by APIRule exists
    And Teardown httpbin service

  Scenario: In-cluster calling a httpbin endpoint unsecured
    Given The APIRule template file is set to "no-auth-wildcard.yaml"
    And There is a httpbin service
    When The APIRule is applied
    Then In-cluster calling the "/status/200" endpoint without a token should fail
    And In-cluster calling the "/headers" endpoint without a token should fail

# update v2 (noAuth)
  Scenario: Updating an APIRule and calling a httpbin endpoint unsecured on all paths
    Given The APIRule template file is set to "no-auth-wildcard.yaml"
    And There is a httpbin service
    And The APIRule is applied
    And The APIRule contains original-version annotation set to "v2"
    When The APIRule is updated using manifest "no-auth-wildcard-updated.yaml"
    Then Calling the "/ip" endpoint without a token should result in status between 200 and 200
    And Calling the "/status/200" endpoint without a token should result in status between 200 and 200
    And Calling the "/headers" endpoint without a token should result in status between 200 and 200
    And The APIRule contains original-version annotation set to "v2"
    And Resource of Kind "AuthorizationPolicy" owned by APIRule exists
    And Teardown httpbin service
