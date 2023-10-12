Feature: Checking default kyma gateway
  Background:
    Given APIGateway CR is in Ready state

  Scenario: Kyma gateway is not removed when there is an APIRule
    Given APIGateway CR is in Ready state
    Then there is an "kyma-rule" APIRule
    Then disabling kyma gateway will result in "Warning" state
    And there is APIGateway CR "kyma-gateway" in "kyma-system" namespace
    And APIRule "kyma-rule" is removed

  Scenario: Kyma gateway is removed when there is no APIRule
    Given APIGateway CR "kyma-gateway" is removed
    Then gateway "kyma-gateway" in "kyma-system" namespace does not exist
