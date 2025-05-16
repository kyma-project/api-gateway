# Important: scenarios in this feature rely on the execution order
Feature: Upgrading API Gateway version
  Scenario: Upgrade from latest release to current version
    Given Upgrade: There is a httpbin service
    And Upgrade: There is an endpoint secured with JWT on path "/ip"
    And Upgrade: The APIRule is applied
    And Upgrade: Calling the "/ip" endpoint with a valid "JWT" token should result in status between 200 and 299
    And Upgrade: Fetch APIRule last processed time
    And Upgrade: APIRules have correct version annotation
    When Upgrade: API Gateway is upgraded to current branch version manifest
    And Upgrade: APIRule was reconciled again
    Then Upgrade: APIRules have correct version annotation
    And Upgrade: Calling the "/ip" endpoint without a token should result in status between 400 and 403
    And Upgrade: Calling the "/ip" endpoint with an invalid token should result in status between 400 and 403
    And Upgrade: Calling the "/ip" endpoint with a valid "JWT" token should result in status between 200 and 299
    And Upgrade: Teardown httpbin service

  Scenario: Create an APIRule with upgraded controller version
    Given Upgrade: There is a httpbin service
    And Upgrade: There is an endpoint secured with JWT on path "/headers"
    When Upgrade: The APIRule is applied
    Then Upgrade: Calling the "/headers" endpoint without a token should result in status between 400 and 403
    And Upgrade: Calling the "/headers" endpoint with an invalid token should result in status between 400 and 403
    And Upgrade: Calling the "/headers" endpoint with a valid "JWT" token should result in status between 200 and 299
    And Upgrade: Teardown httpbin service
