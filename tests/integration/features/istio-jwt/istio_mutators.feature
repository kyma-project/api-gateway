Feature: Configuring mutators for an APIRule using Istio resources

  Scenario: Exposing an endpoint with header mutator configured
    Given Mutator-Header: There is a httpbin service
    And Mutator-Header: There is an endpoint on path "/headers" with a header mutator setting "x-mutators-test" header to "a-header-value"
    When Mutator-Header: The APIRule is applied
    Then Mutator-Header: Calling the "/headers" endpoint should return response with header "x-mutators-test" with value "a-header-value"

  Scenario: Exposing an endpoint with cookies mutator configured
    Given Mutator-Cookie: There is a httpbin service
    And Mutator-Cookie: There is an on path "/cookies" with a cookies mutator setting "x-mutators-test" cookie to "a-cookie-value"
    When Mutator-Cookie: The APIRule is applied
    And Mutator-Cookie: Calling the "/cookies" endpoint should return response with cookie "x-mutators-test" with value "a-cookie-value"