Feature: Rate limiting header based

  Scenario: Connectivity to pod rate limited by header-based configuration
    Given there is a httpbin service
    When RateLimit header-based configuration is applied
    Then calling the "/ip" endpoint with "GET" method with header 2 times should result in status code 429 for requests