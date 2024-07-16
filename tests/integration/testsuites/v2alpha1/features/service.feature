Feature: Exposing services in APIRule

  Scenario: Endpoints exposed in APIRule should fallback to service defined on root level when there is no service defined on rule level
    Given ServiceFallback: There is a httpbin service
    And ServiceFallback: There is an endpoint secured with JWT on path "/headers" with service definition
    And ServiceFallback: There is an endpoint secured with JWT on path "/ip"
    When ServiceFallback: The APIRule with service on root level is applied
    Then ServiceFallback: Calling the "/headers" endpoint with a valid "JWT" token should result in status between 200 and 299
    And ServiceFallback: Calling the "/ip" endpoint with a valid "JWT" token should result in status between 200 and 299
    And ServiceFallback: Teardown httpbin service

  Scenario: Exposing endpoints in two namespaces
    Given ServiceTwoNamespaces: There is a httpbin service
    And ServiceTwoNamespaces: There is a service with workload in a second namespace
    And ServiceTwoNamespaces: There is an endpoint secured with JWT on path "/get" in APIRule Namespace
    And ServiceTwoNamespaces: There is an endpoint secured with JWT on path "/hello" in different namespace
    When ServiceTwoNamespaces: The APIRule is applied
    Then ServiceTwoNamespaces: Calling the "/get" endpoint with a valid "JWT" token should result in status between 200 and 299
    And ServiceTwoNamespaces: Calling the "/hello" endpoint with a valid "JWT" token should result in status between 200 and 299
    And ServiceTwoNamespaces: Calling the "/get" endpoint without token should result in status between 400 and 403
    And ServiceTwoNamespaces: Calling the "/hello" endpoint without token should result in status between 400 and 403
    And ServiceTwoNamespaces: Teardown httpbin service

  Scenario: Exposing different services with same methods
    Given ServiceDiffSvcSameMethods: There is a httpbin service
    And ServiceDiffSvcSameMethods: There is a workload and service for httpbin and helloworld
    And ServiceDiffSvcSameMethods: There is an endpoint secured with JWT on path "/headers" for httpbin service with methods '["GET", "POST"]'
    And ServiceDiffSvcSameMethods: There is an endpoint secured with JWT on path "/hello" for helloworld service with methods '["GET", "POST"]'
    When ServiceDiffSvcSameMethods: The APIRule is applied
    Then ServiceDiffSvcSameMethods: Calling the "/headers" endpoint with a valid "JWT" token should result in status between 200 and 299
    And ServiceDiffSvcSameMethods: Calling the "/hello" endpoint with a valid "JWT" token should result in status between 200 and 299
    And ServiceDiffSvcSameMethods: Teardown httpbin service

  Scenario: Calling a helloworld endpoint with custom label selector service
    Given ServiceCustomLabelSelector: There is a helloworld service with custom label selector name "custom-name"
    And ServiceCustomLabelSelector: There is an endpoint secured with JWT on path "/hello"
    When ServiceCustomLabelSelector: The APIRule is applied
    Then ServiceCustomLabelSelector: Calling the "/hello" endpoint without a token should result in status between 400 and 403
    And ServiceCustomLabelSelector: Calling the "/hello" endpoint with a valid "JWT" token should result in status between 200 and 299
    And ServiceCustomLabelSelector: Teardown helloworld service
