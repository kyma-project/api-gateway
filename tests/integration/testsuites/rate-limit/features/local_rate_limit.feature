Feature: Local rate limiting

  Scenario: Connectivity to pod rate limited by header-based configuration
    Given there is a httpbin service
    When RateLimit header-based configuration is applied
    Then calling the "/ip" endpoint with header 2 times should result in status code 429 for requests

  Scenario: Connectivity to pod rate limited by path and header based configuration
    Given there is a httpbin service
    When RateLimit path and header based configuration is applied
    Then calling the "/headers" endpoint with header 2 times should result in status code 429 for requests

  Scenario: Connectivity to pod rate limited by path-based configuration
    Given there is a httpbin service
    When RateLimit path-based configuration is applied
    Then calling the "/ip" endpoint 2 times should result in status code 429 for requests
