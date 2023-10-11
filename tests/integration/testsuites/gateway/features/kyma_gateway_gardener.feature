Feature: Checking default kyma gateway on gardener
  Scenario: Kyma gateway is deployed
    Given APIGateway CR is in "Ready" state
    And there is a "kyma-gateway" gateway in "kyma-system" namespace
    And there is a "kyma-gateway" Gardener Certificate CR in "istio-system" namespace
    And there is a "kyma-gateway" Gardener DNSEntry CR in "kyma-system" namespace
    And there is a "kyma-gateway-certs" secret in "istio-system" namespace
