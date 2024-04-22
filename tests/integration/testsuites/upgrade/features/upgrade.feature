# Important: scenarios in this feature rely on the execution order
Feature: Upgrading API Gateway version
  Scenario: Fail when there is operator manifest is not applicable
    When Common: API Gateway is upgraded to current branch version with "failing" manifest and should "fail"

  Scenario: Upgrade from latest release to current version
    Given Common: There is a httpbin service
    And Common: There is an endpoint secured with JWT on path "/ip"
    And Common: The APIRule is applied
    And Common: Calling the "/ip" endpoint with a valid "JWT" token should result in status between 200 and 299
    When Common: API Gateway is upgraded to current branch version with "generated" manifest and should "succeed"
    And Common: A reconciliation happened in the last 60 seconds
    Then Common: Calling the "/ip" endpoint without a token should result in status between 400 and 403
    And Common: Calling the "/ip" endpoint with an invalid token should result in status between 400 and 403
    And Common: Calling the "/ip" endpoint with a valid "JWT" token should result in status between 200 and 299
    And Common: Teardown httpbin service

  Scenario: Create an APIRule with upgraded controller version
    Given Common: There is a httpbin service
    And Common: There is an endpoint secured with JWT on path "/headers"
    When Common: The APIRule is applied
    Then Common: Calling the "/headers" endpoint without a token should result in status between 400 and 403
    And Common: Calling the "/headers" endpoint with an invalid token should result in status between 400 and 403
    And Common: Calling the "/headers" endpoint with a valid "JWT" token should result in status between 200 and 299
    And Common: Teardown httpbin service
