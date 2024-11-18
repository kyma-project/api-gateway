Feature: Exposing endpoints with NoAuth when specifying short host name in APIRule

  Scenario: Calling a httpbin endpoint unsecured
    Given There is a httpbin service
    When The APIRule is applied
    Then Calling short host "httpbin" with path "/ip" without a token should result in status between 200 and 200
    And Teardown httpbin service

  Scenario: Calling a httpbin endpoint unsecured (error scenario)
    Given There is a httpbin service
    When Specifies custom Gateway "kyma-system"/"another-gateway"
    And The APIRule is applied and contains error status with "Could not get specified Gateway" message
    And Teardown httpbin service
