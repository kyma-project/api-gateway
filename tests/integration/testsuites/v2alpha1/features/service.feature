Feature: Exposing services in APIRule

  Scenario: Endpoints exposed in APIRule should fallback to service defined on root level when there is no service defined on rule level
    Given The APIRule template file is set to "service-fallback.yaml"
    And There is a httpbin service
    And There is an endpoint secured with JWT on path "/headers" with service definition
    And There is an endpoint secured with JWT on path "/ip"
    When The APIRule with service on root level is applied
    Then Calling the "/headers" endpoint with a valid "JWT" token should result in status between 200 and 299
    And Calling the "/ip" endpoint with a valid "JWT" token should result in status between 200 and 299
    And Teardown httpbin service

  Scenario: Exposing endpoints in two namespaces
    Given The APIRule template file is set to "service-two-namespaces.yaml"
    And There is a httpbin service
    And There is a service with workload in a second namespace
    And There is an endpoint secured with JWT on path "/get" in APIRule Namespace
    And There is an endpoint secured with JWT on path "/hello" in different namespace
    When The APIRule is applied
    Then Calling the "/get" endpoint with a valid "JWT" token should result in status between 200 and 299
    And Calling the "/hello" endpoint with a valid "JWT" token should result in status between 200 and 299
    And Calling the "/get" endpoint without token should result in status between 400 and 403
    And Calling the "/hello" endpoint without token should result in status between 400 and 403
    And Teardown httpbin service

  Scenario: Exposing different services with same methods
    Given The APIRule template file is set to "service-diff-same-methods.yaml"
    And There is a httpbin service
    And There is a workload and service for httpbin and helloworld
    And There is an endpoint secured with JWT on path "/headers" for httpbin service with methods '["GET", "POST"]'
    And There is an endpoint secured with JWT on path "/hello" for helloworld service with methods '["GET", "POST"]'
    When The APIRule is applied
    Then Calling the "/headers" endpoint with a valid "JWT" token should result in status between 200 and 299
    And Calling the "/hello" endpoint with a valid "JWT" token should result in status between 200 and 299
    And Teardown httpbin service

  Scenario: Calling a helloworld endpoint with custom label selector service
    Given The APIRule template file is set to "service-custom-label-selector.yaml"
    And There is a httpbin service
    And There is an endpoint secured with JWT on path "/hello"
    When The APIRule is applied
    Then Calling the "/hello" endpoint without a token should result in status between 400 and 403
    And Calling the "/hello" endpoint with a valid "JWT" token should result in status between 200 and 299
    And Teardown helloworld service
