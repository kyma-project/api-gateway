Feature: Checking default kyma gateway
  Background:
    Given APIGateway CR is in "Ready" state with description ""

  Scenario: Kyma Gateway is not removed when there is an APIRule
    Given there is an "kyma-rule" APIRule with Gateway "kyma-system/kyma-gateway"
    When disabling Kyma gateway
    Then APIGateway CR is in "Warning" state with description "There are custom resources that block the deletion. Please take a look at kyma-system/api-gateway-controller-manager logs to see more information about the warning"
    And there is Istio Gateway "kyma-gateway" in "kyma-system" namespace
    And APIRule "kyma-rule" is removed
    And APIGateway CR is in "Ready" state with description ""

  Scenario: Kyma Gateway is not removed when there is a VirtualService
    Given there is an "kyma-vs" VirtualService with Gateway "kyma-system/kyma-gateway"
    When disabling Kyma gateway
    Then APIGateway CR is in "Warning" state with description "There are custom resources that block the deletion. Please take a look at kyma-system/api-gateway-controller-manager logs to see more information about the warning"
    And there is Istio Gateway "kyma-gateway" in "kyma-system" namespace
    And VirtualService "kyma-vs" is removed
    And APIGateway CR is in "Ready" state with description ""

  Scenario: Kyma Gateway is removed when there is no blocking resources
    When APIGateway CR "test-gateway" is removed
    Then APIGateway CR "test-gateway" "is not" present
    And gateway "kyma-gateway" in "kyma-system" namespace does not exist
