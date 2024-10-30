Feature: Exposing service using asterisks

  Scenario: ExposeAsterisk: Exposure of service using asterisks in paths
    Given ExposeAsterisk: There is a httpbin service
    When ExposeAsterisk: The APIRule is applied
    Then ExposeAsterisk: Calling the "/anything/random/one" endpoint with "GET" method should result in status between 200 and 200
    And ExposeAsterisk: Calling the "/anything/any/random/two" endpoint with "POST" method should result in status between 200 and 200
    And ExposeAsterisk: Calling the "/anything/any/random" endpoint with "PUT" method should result in status between 200 and 200
    And ExposeAsterisk: Calling the "/anything/random/one/any/random/two" endpoint with "DELETE" method should result in status between 200 and 200
    And ExposeAsterisk: Teardown httpbin service
