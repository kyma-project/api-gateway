Feature: Checking default kyma gateway on gardener
  Scenario: Kyma gateway is deployed
    Given there is an APIGateway operator in "Ready" state
    Then there is a "kyma-gateway" gateway in "kyma-system" namespace
    And there is a "kyma-cert" Certificate CR
    And there is a "kyma-dns" DNSEntry CR
    And there is a "kyma-gateway-certs" secret in "istio-system" namespace
