Feature: Checking default kyma gateway
  Background:
    Given APIGateway CR is applied
    Then APIGateway CR is in "Ready" state

  Scenario: Kyma gateway is deployed
    Then there is a "kyma-gateway" gateway in "kyma-system" namespace
    And there is a "kyma-gateway-certs" secret in "istio-system" namespace

  Scenario: Kyma gateway is not removed when there is an APIRule
    Given there is an "kyma-rule" APIRule
    Then disabling kyma gateway will result in "Warning" state
    And there is a "kyma-gateway" gateway in "kyma-system" namespace
    And APIRule "kyma-rule" is removed

  Scenario: Kyma gateway is removed when there is no APIRule
    And disabling kyma gateway will result in "Ready" state
    And gateway "kyma-gateway" in "kyma-system" namespace does not exist
