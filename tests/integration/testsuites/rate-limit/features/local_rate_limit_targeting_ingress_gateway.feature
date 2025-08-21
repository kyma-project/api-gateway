Feature: Local rate limiting targeting Ingress Gateway

  Scenario: Connectivity to ingress rate limited by header-based configuration
    Given there is a httpbin service
    When RateLimit targeting Istio ingress gateway with header-based configuration is applied
    Then calling the "/ip" endpoint with header should result in status code 429 for requests
    And Teardown RateLimit
    And Teardown httpbin service

  Scenario: Connectivity to ingress rate limited by path and header based configuration
    Given there is a httpbin service
    When RateLimit targeting Istio ingress gateway with path and header based configuration is applied
    Then calling the "/headers" endpoint with header should result in status code 429 for requests
    And Teardown RateLimit
    And Teardown httpbin service

  Scenario: Connectivity to ingress rate limited by path-based configuration
    Given there is a httpbin service
    When RateLimit targeting Istio ingress gateway with path-based configuration is applied
    Then calling the "/ip" endpoint should result in status code 429 for requests
    And Teardown RateLimit
    And Teardown httpbin service
