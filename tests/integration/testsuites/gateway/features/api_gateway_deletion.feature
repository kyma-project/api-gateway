Feature: Deleting API-Gateway CR
  Background:
    Given APIGateway CR is applied

  Scenario: Deleting API-Gateway CR without blocking resources
    Given APIGateway CR is in "Ready" state with description ""
    When APIGateway CR is deleted
    Then APIGateway CR is "not present" on cluster

  Scenario: Deleting API-Gateway CR with APIRule present
    Given APIGateway CR is in "Ready" state with description ""
    And APIRule "test-api-rule" is applied
    When gateway "kyma-gateway" is removed
    Then APIGateway CR is "present" on cluster
    And APIGateway CR is in "Warning" state with description "There are APIRule(s) that block the deletion. Please take a look at kyma-system/api-gateway-controller-manager logs to see more information about the warning"
