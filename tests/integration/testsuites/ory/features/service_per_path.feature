Feature: Exposing two services on path level

  Scenario: Service per path
    Given Service per path: There is a httpbin service
    When Service per path: There is a helloworld service and an APIRule with two endpoints exposed with different services, one on spec level and one on rule level
    Then Service per path: Calling the endpoint "/headers" and "/hello" with any token should result in status between 200 and 299
    And Service per path: Calling the endpoint "/headers" and "/hello" without a token should result in status between 200 and 299
    And Service per path: Teardown httpbin service
