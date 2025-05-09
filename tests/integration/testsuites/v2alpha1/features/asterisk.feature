Feature: Exposing service using asterisks

  Scenario: Exposure of service using asterisks in paths
    Given The APIRule template file is set to "asterisk-paths.yaml"
    And There is a httpbin service
    When The APIRule is applied
    Then Calling the "/anything/random/one" endpoint with "GET" method should result in status between 200 and 200
    And Calling the "/anything/rand*m/one" endpoint with "GET" method should result in status between 200 and 200
    And Calling the "/anything/any/random/two" endpoint with "POST" method should result in status between 200 and 200
    And Calling the "/anything/any/random" endpoint with "PUT" method should result in status between 200 and 200
    And Calling the "/anything/a+b" endpoint with "PUT" method should result in status between 200 and 200
    And Calling the "/anything/a%20b" endpoint with "PUT" method should result in status between 200 and 200
    And Calling the "/anything/rand*m" endpoint with "PUT" method should result in status between 200 and 200
    And Calling the "/anything/random/one/any/random/two" endpoint with "DELETE" method should result in status between 200 and 200
    And Calling the "/anything/rand*m/one/any/rand*m/two" endpoint with "DELETE" method should result in status between 200 and 200
    And Calling the "/anything/random/one" endpoint with "POST" method should result in status between 404 and 404
    And Calling the "/anything/one" endpoint with "GET" method should result in status between 200 and 200
    And Calling the "/anything/one/two" endpoint with "GET" method should result in status between 200 and 200
    And Calling the "/anything/any/random/two" endpoint with "GET" method should result in status between 200 and 200
    And Calling the "/anything/" endpoint with "GET" method should result in status between 200 and 200
    And Teardown httpbin service
