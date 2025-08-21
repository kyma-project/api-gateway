Feature: Local rate limiting targeting pod

  Scenario: Connectivity to pod rate limited by header-based configuration
    Given there is a httpbin service
    When RateLimit header-based configuration is applied
    Then calling the "/ip" endpoint with header should result in status code 429 for requests
    And Teardown RateLimit
    And Teardown httpbin service

  Scenario: Connectivity to pod rate limited by path and header based configuration
    Given there is a httpbin service
    When RateLimit path and header based configuration is applied
    Then calling the "/headers" endpoint with header should result in status code 429 for requests
    And Teardown RateLimit
    And Teardown httpbin service

  Scenario: Connectivity to pod rate limited by path-based configuration
    Given there is a httpbin service
    When RateLimit path-based configuration is applied
    Then calling the "/ip" endpoint should result in status code 429 for requests
    And Teardown RateLimit
    And Teardown httpbin service
