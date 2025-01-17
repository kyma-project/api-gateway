Feature: Rate limiting path and header based

  Scenario: Connectivity to pod rate limited by path and header based configuration
    Given there is a httpbin service
    When RateLimit path and header based configuration is applied
    Then calling the "/ip" endpoint with "GET" method with header 2 times should result in status code 429 for requests