Feature: Configuring mutators for an APIRule secured with Istio JWT authorization strategy

  Scenario: Exposing an endpoint with header mutator configured
    Given JwtMutatorHeader: There is a httpbin service
    And JwtMutatorHeader: There is an endpoint on path "/headers" with a header mutator setting "x-mutators-test" header to "a-header-value"
    When JwtMutatorHeader: The APIRule is applied
    Then JwtMutatorHeader: Calling the "/headers" endpoint should return response with header "X-Mutators-Test" with value "a-header-value"

  Scenario: Exposing an endpoint with cookie mutator configured
    Given JwtMutatorCookie: There is a httpbin service
    And JwtMutatorCookie: There is an endpoint on path "/cookies" with a cookie mutator setting "x-mutators-test" cookie to "a-cookie-value"
    When JwtMutatorCookie: The APIRule is applied
    Then JwtMutatorCookie: Calling the "/cookies" endpoint should return response with cookie "x-mutators-test" with value "a-cookie-value"

    Scenario: Exposing an endpoint with a header mutator setting multiple headers and cookie mutator setting multiple cookies
      # TODO: Implement this scenario