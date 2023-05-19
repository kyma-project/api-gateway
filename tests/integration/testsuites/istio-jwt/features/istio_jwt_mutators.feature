Feature: Configuring mutators for an APIRule secured with Istio JWT authorization strategy

  Scenario: Exposing an endpoint with header mutator configured
    Given JwtMutatorHeader: There is a httpbin service
    And JwtMutatorHeader: There is an endpoint on path "/headers" with a header mutator setting "x-mutators-test" header to "a-header-value"
    When JwtMutatorHeader: The APIRule is applied
    Then JwtMutatorHeader: Calling the "/headers" endpoint should return response with header "X-Mutators-Test" with value "a-header-value"
    Then JwtMutatorHeader: Teardown httpbin service

  Scenario: Exposing an endpoint with cookie mutator configured
    Given JwtMutatorCookie: There is a httpbin service
    And JwtMutatorCookie: There is an endpoint on path "/cookies" with a cookie mutator setting "x-mutators-test" cookie to "a-cookie-value"
    When JwtMutatorCookie: The APIRule is applied
    Then JwtMutatorCookie: Calling the "/cookies" endpoint should return response with cookie "x-mutators-test" with value "a-cookie-value"
    Then JwtMutatorCookie: Teardown httpbin service

  Scenario: Exposing an endpoint with a header mutator setting multiple headers and cookie mutator setting multiple cookies
    Given JwtMultipleMutators: There is a httpbin service
    And JwtMultipleMutators: There is an endpoint on path "/cookies" with a cookie mutator setting "x-cookie-mutators-test" cookie to "a-cookie-value" and "x-cookie-mutators-test2" cookie to "a-cookie-value2"
    And JwtMultipleMutators: There is an endpoint on path "/headers" with a header mutator setting "x-header-mutators-test" header to "a-header-value" and "x-header-mutators-test2" header to "a-header-value2"
    When JwtMultipleMutators: The APIRule is applied
    Then JwtMultipleMutators: Calling the "/headers" endpoint should return response with headers "X-Header-Mutators-Test" with value "a-header-value" and "X-Header-Mutators-Test2" with value "a-header-value2"
    Then JwtMultipleMutators: Calling the "/cookies" endpoint should return response with cookies "x-cookie-mutators-test" with value "a-cookie-value" and "x-cookie-mutators-test2" with value "a-cookie-value2"
    Then JwtMultipleMutators: Teardown httpbin service

  Scenario: Exposing an endpoint with header and cookie mutator configured should overwrite the original header and cookie
    Given JwtMutatorsOverwrite: There is a httpbin service
    And JwtMutatorsOverwrite: There is an endpoint on path "/headers" with a header mutator setting "x-mutators-test" header to "a-mutator-value"
    And JwtMutatorsOverwrite: There is an endpoint on path "/headers" with a cookie mutator setting "x-mutators-test" cookie to "a-mutator-cookie-value"
    When JwtMutatorsOverwrite: The APIRule is applied
    Then JwtMutatorsOverwrite: Calling the "/headers" endpoint with a request having cookie header with value "x-cookie-test=a-cookie-value" should return cookie header with value "x-mutators-test=a-mutator-cookie-value"
    Then JwtMutatorsOverwrite: Calling the "/headers" endpoint with a request having header "X-Mutators-Test" with value "a-header-value" should return same header with value "a-mutator-value"
    Then JwtMutatorsOverwrite: Teardown httpbin service
