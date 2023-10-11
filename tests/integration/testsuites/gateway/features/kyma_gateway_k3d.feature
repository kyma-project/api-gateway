Feature: Checking default kyma gateway on k3d
  Background:
    Given APIGateway CR is in "Ready" state

  Scenario: Kyma gateway is deployed
    Given there is a "kyma-gateway" gateway in "kyma-system" namespace
    Then there is a "kyma-gateway-certs" secret in "istio-system" namespace
