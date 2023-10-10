Feature: Checking default kyma gateway
  Background:
    # TODO
    Given APIGateway CR is applied
    Then APIGateway CR is in "Ready" state

  Scenario: Kyma gateway is not removed when there is an APIRule
    Given there is an "kyma-rule" APIRule
    Then disabling kyma gateway will result in "Warning" state
    And there is a "kyma-gateway" gateway in "kyma-system" namespace
    And APIRule "kyma-rule" is removed

  Scenario: Kyma gateway is removed when there is no APIRule
    Given gateway "kyma-gateway" is removed
    Then gateway "kyma-gateway" in "kyma-system" namespace does not exist
