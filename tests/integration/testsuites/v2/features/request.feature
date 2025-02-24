Feature: Support for request headers and cookies

  Scenario: Exposing an endpoint with request header configured
    Given The APIRule template file is set to "jwt-request-header.yaml"
    And There is a httpbin service
    And There is an endpoint on path "/headers" with a header mutator setting "X-Request-Test" header to "a-header-value"
    When The APIRule is applied
    Then Calling the "/headers" endpoint should return response with header "X-Request-Test" with value "a-header-value"
    Then Teardown httpbin service

  Scenario: Exposing an endpoint with request cookie configured
    Given The APIRule template file is set to "jwt-request-cookie.yaml"
    And There is a httpbin service
    And There is an endpoint on path "/headers" with a cookie mutator setting "x-request-test" cookie to "a-cookie-value"
    When The APIRule is applied
    Then Calling the "/headers" endpoint should return response with header "Cookie" with value "x-request-test=a-cookie-value"
    Then Teardown httpbin service

