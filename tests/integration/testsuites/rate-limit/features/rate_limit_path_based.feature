Feature: Rate limiting path based

  Scenario: Connectivity to pod rate limited by path-based configuration
    Given there is a httpbin service
    When RateLimit path-based configuration is applied
    Then calling the "/ip" endpoint with "GET" method 2 times should result in status code 429 for requests
