Feature: Exposing endpoints with ExtAuth

  Scenario: Calling a httpbin endpoint secured with ExtAuth
    Given ExtAuthCommon: There is a httpbin service
    And ExtAuthCommon: There is an endpoint secured with ExtAuth "sample-ext-authz-http" on path "/headers"
    When ExtAuthCommon: The APIRule is applied
    Then ExtAuthCommon: Calling the "/headers" endpoint with header "x-ext-authz" with value "deny" should result in status between 400 and 403
    And ExtAuthCommon: Calling the "/headers" endpoint with header "x-ext-authz" with value "allow" should result in status between 200 and 299

  Scenario: Calling a httpbin endpoint secured with ExtAuth with JWT restrictions
    Given ExtAuthJwt: There is a httpbin service
    And ExtAuthJwt: There is an endpoint secured with ExtAuth "sample-ext-authz-http" on path "/headers"
    And ExtAuthJwt: The endpoint has JWT restrictions
    When ExtAuthJwt: The APIRule is applied
    Then ExtAuthJwt: Calling the "/headers" endpoint with header "x-ext-authz" with value "allow" and no token should result in status between 400 and 403
    And ExtAuthJwt: Calling the "/headers" endpoint with header "x-ext-authz" with value "allow" and an invalid "JWT" token should result in status between 400 and 403
    And ExtAuthJwt: Calling the "/headers" endpoint with header "x-ext-authz" with value "deny" and a valid "JWT" token should result in status between 400 and 403
    And ExtAuthJwt: Calling the "/headers" endpoint with header "x-ext-authz" with value "allow" and a valid "JWT" token should result in status between 200 and 299
