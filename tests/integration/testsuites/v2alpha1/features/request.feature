Feature: Support for request headers and cookies

  Scenario: Exposing an endpoint with request header configured
    Given RequestHeader: There is a httpbin service
    And RequestHeader: There is an endpoint on path "/headers" with a header mutator setting "x-mutators-test" header to "a-header-value"
    When RequestHeader: The APIRule is applied
    Then RequestHeader: Calling the "/headers" endpoint should return response with header "X-Mutators-Test" with value "a-header-value"
    Then RequestHeader: Teardown httpbin service

  Scenario: Exposing an endpoint with request cookie configured
    Given RequestCookie: There is a httpbin service
    And RequestCookie: There is an endpoint on path "/cookies" with a cookie mutator setting "x-mutators-test" cookie to "a-cookie-value"
    When RequestCookie: The APIRule is applied
    Then RequestCookie: Calling the "/cookies" endpoint should return response with cookie "x-mutators-test" with value "a-cookie-value"
    Then RequestCookie: Teardown httpbin service

