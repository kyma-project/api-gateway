Feature: Exposing endpoints with ExtAuth

  Scenario: Calling a httpbin endpoint secured with ExtAuth
    Given The APIRule template file is set to "ext-auth-common.yaml"
    And There is a httpbin service
    And There is an endpoint secured with ExtAuth "sample-ext-authz-http" on path "/headers"
    When The APIRule is applied
    Then Calling the "/headers" endpoint with header "x-ext-authz" with value "deny" should result in status between 400 and 403
    And Calling the "/headers" endpoint with header "x-ext-authz" with value "allow" should result in status between 200 and 299

  Scenario: Calling a httpbin endpoint secured with ExtAuth with JWT restrictions
    Given The APIRule template file is set to "ext-auth-common.yaml"
    And There is a httpbin service
    And There is an endpoint secured with ExtAuth "sample-ext-authz-http" on path "/headers"
    And The endpoint has JWT restrictions
    When The APIRule is applied
    Then Calling the "/headers" endpoint with header "x-ext-authz" with value "allow" and no token should result in status between 400 and 403
    And Calling the "/headers" endpoint with header "x-ext-authz" with value "allow" and an invalid "JWT" token should result in status between 400 and 403
    And Calling the "/headers" endpoint with header "x-ext-authz" with value "deny" and a valid "JWT" token should result in status between 400 and 403
    And Calling the "/headers" endpoint with header "x-ext-authz" with value "allow" and a valid "JWT" token should result in status between 200 and 299
