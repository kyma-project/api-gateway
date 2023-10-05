Feature: Checking default kyma gateway
  Scenario: Kyma gateway is deployed
    Given there is an APIGateway operator in "Ready" state
    Then there is a "kyma-gateway" gateway in "kyma-system" namespace
    And there is a "kyma-gateway-certs" secret in "istio-system" namespace

  Scenario: Kyma gateway cannot be disabled when there is an APIRule
    Given there is an "kyma-rule" APIRule
    Then disabling kyma gateway will result in "Warning" due to existing APIRule
    And enabling kyma gateway back will result in "Ready" state
    And removing an APIRule and APIGateway operator will also remove "kyma-gateway" in "kyma-system" namespace
    And enabling kyma gateway back will result in "Ready" state
