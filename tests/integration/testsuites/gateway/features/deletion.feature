Feature: Deleting API-Gateway CR
  Background:
    Given APIGateway CR is in "Ready" state with description ""

  Scenario: Deleting API-Gateway CR without blocking resources
    When APIGateway CR "test-gateway" is removed
    Then APIGateway CR "test-gateway" "is not" present

  Scenario: Deleting API-Gateway CR with APIRule present
    Given there is an "kyma-rule" APIRule with Gateway "kyma-system/some-other-gateway"
    When APIGateway CR "test-gateway" is removed
    Then APIGateway CR "test-gateway" "is" present
    And APIGateway CR is in "Warning" state with description "There are custom resources that block the deletion of API-Gateway CR. Please take a look at kyma-system/api-gateway-controller-manager logs to see more information about the warning"
    And APIRule "kyma-rule" is removed
    And APIGateway CR "test-gateway" "is not" present

  Scenario: Deleting API-Gateway CR with ORYRule present
    Given there is an "ory-rule" ORYRule
    When APIGateway CR "test-gateway" is removed
    Then APIGateway CR "test-gateway" "is" present
    And APIGateway CR is in "Warning" state with description "There are custom resources that block the deletion of API-Gateway CR. Please take a look at kyma-system/api-gateway-controller-manager logs to see more information about the warning"
    And ORYRule "ory-rule" is removed
    And APIGateway CR "test-gateway" "is not" present
