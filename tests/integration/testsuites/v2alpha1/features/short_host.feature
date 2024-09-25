Feature: Exposing endpoints with NoAuth when specifying short host name in APIRule

  Scenario: Calling a httpbin endpoint unsecured
    Given ShortHost: There is a httpbin service
    When ShortHost: The APIRule is applied
    Then ShortHost: Calling short host "httpbin" with path "/ip" without a token should result in status between 200 and 200
    And ShortHost: Teardown httpbin service

  Scenario: Calling a httpbin endpoint unsecured (error scenario)
    Given ShortHostError: There is a httpbin service
    When ShortHostError: Specifies custom Gateway "kyma-system"/"another-gateway"
    And ShortHostError: The APIRule is applied and contains error status with "Unable to get specified Gateway" message
    And ShortHostError: Teardown httpbin service
