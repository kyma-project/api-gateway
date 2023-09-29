Feature: Applying and deleting API-Gateway CR

  Background:
    Given Namespace "kyma-system" is "present"
    And API-Gateway CRD is installed
    And "API-Gateway CR" is not present on cluster
    And "Deployment" "api-gateway-controller-manager" in namespace "kyma-system" is ready

  Scenario: Applying of API-Gateway CR
    Given API-Gateway CR "api-gateway-sample" is applied in namespace "kyma-system"
    Then API-Gateway CR "api-gateway-sample" in namespace "kyma-system" has status "Ready"
    And "Gateway" "kyma-gateway" in namespace "kyma-system" is "present"

  Scenario: Deleting of API-Gateway CR without blocking resources
    Given API-Gateway CR "api-gateway-sample" is applied in namespace "kyma-system"
    And API-Gateway CR "api-gateway-sample" in namespace "kyma-system" has status "Ready"
    And "Gateway" "kyma-gateway" in namespace "kyma-system" is "present"
    When "API-Gateway CR" "api-gateway-sample" in namespace "kyma-system" is deleted
    Then "API-Gateway CR" is not present on cluster

  Scenario: Deleting of API-Gateway CR with blocking APIRule
    Given API-Gateway CR "api-gateway-sample" is applied in namespace "kyma-system"
    And API-Gateway CR "api-gateway-sample" in namespace "kyma-system" has status "Ready"
    And "Gateway" "kyma-gateway" in namespace "kyma-system" is "present"
    And Httpbin application "httpbin" is running in namespace "default"
    And "Deployment" "httpbin" in namespace "default" is ready
    And APIRule "test-api-rule" exposing service "httpbin.default.svc.cluster.local" by gateway "kyma-system/kyma-gateway" is configured in namespace "default"
    When "API-Gateway CR" "api-gateway-sample" in namespace "kyma-system" is deleted
    Then API-Gateway CR "api-gateway-sample" in namespace "kyma-system" has status "Warning" with description "There are custom resource(s) that block the deletion"
    And "API-Gateway CR" is present on cluster

  Scenario: Deleting of API-Gateway CR with blocking VirtualService
    Given API-Gateway CR "api-gateway-sample" is applied in namespace "kyma-system"
    And API-Gateway CR "api-gateway-sample" in namespace "kyma-system" has status "Ready"
    And "Gateway" "kyma-gateway" in namespace "kyma-system" is "present"
    And Httpbin application "httpbin" is running in namespace "default"
    And "Deployment" "httpbin" in namespace "default" is ready
    And Virtual service "test-vs" exposing service "httpbin.default.svc.cluster.local" by gateway "kyma-system/kyma-gateway" is configured in namespace "default"
    When "API-Gateway CR" "api-gateway-sample" in namespace "kyma-system" is deleted
    Then API-Gateway CR "api-gateway-sample" in namespace "kyma-system" has status "Warning" with description "There are custom resource(s) that block the deletion"
    And "API-Gateway CR" is present on cluster
