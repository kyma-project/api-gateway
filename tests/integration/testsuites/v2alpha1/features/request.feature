Feature: Support for request mutators

  Scenario: Exposing an endpoint with header mutator configured
    Given MutatorHeader: There is a httpbin service
    And MutatorHeader: There is an endpoint on path "/headers" with a header mutator setting "x-mutators-test" header to "a-header-value"
    When MutatorHeader: The APIRule is applied
    Then MutatorHeader: Calling the "/headers" endpoint should return response with header "X-Mutators-Test" with value "a-header-value"
    Then MutatorHeader: Teardown httpbin service

  Scenario: Exposing an endpoint with cookie mutator configured
    Given MutatorCookie: There is a httpbin service
    And MutatorCookie: There is an endpoint on path "/cookies" with a cookie mutator setting "x-mutators-test" cookie to "a-cookie-value"
    When MutatorCookie: The APIRule is applied
    Then MutatorCookie: Calling the "/cookies" endpoint should return response with cookie "x-mutators-test" with value "a-cookie-value"
    Then MutatorCookie: Teardown httpbin service

