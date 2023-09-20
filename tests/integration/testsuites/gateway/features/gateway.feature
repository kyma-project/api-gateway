Feature: Checking default kyma gateway

  @k3d
  Scenario: Kyma gateway is deployed on k3d cluster
    Given there is a "kyma-gateway" Gateway in "kyma-system" namespace on k3d cluster

  @gardener
  Scenario: Kyma gateway is deployed on Gardener cluster
    Given there is a "kyma-gateway" Gateway in "kyma-system" namespace on gardener cluster
