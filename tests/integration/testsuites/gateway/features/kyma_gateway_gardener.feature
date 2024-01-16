Feature: Checking default kyma gateway on gardener
  Background:
    Given APIGateway CR is in "Ready" state with description ""

  Scenario: Kyma gateway is deployed
    Given there is Istio Gateway "kyma-gateway" in "kyma-system" namespace
    And there is a "kyma-tls-cert" Gardener Certificate CR in "istio-system" namespace
    And there is a "kyma-gateway" Gardener DNSEntry CR in "kyma-system" namespace
    And there is a "kyma-gateway-certs" secret in "istio-system" namespace
    And there "is" "VirtualService" "istio-healthz" in namespace "istio-system"
