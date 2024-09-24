Feature: Exposing endpoints with NoAuth when specifying short host only in APIRule

  Scenario: Calling a httpbin endpoint unsecured
    Given ShortHost: There is a httpbin service
    When ShortHost: The APIRule is applied
    Then ShortHost: Calling the "/ip" endpoint without a token should result in status between 200 and 200
    And ShortHost: Teardown httpbin service
