Feature: Checking default kyma gateway
  Background:
    Given APIGateway CR is in Ready state

  Scenario: Oathkeeper is installed and uninstalled depending on APIGateway presence
    Given there is APIGateway CR "kyma-gateway" in "kyma-system" namespace
    Then there "is" "Deployment" "ory-oathkeeper" in namespace "kyma-system"
    And "Deployment" "ory-oathkeeper" in namespace "kyma-system" has status "Ready"
    And there "is" "ConfigMap" "ory-oathkeeper-config" in namespace "kyma-system"
    And there "is" "CustomResourceDefinition" "rules.oathkeeper.ory.sh" in the cluster
    And there "is" "HorizontalPodAutoscaler" "ory-oathkeeper" in namespace "kyma-system"
    And "HorizontalPodAutoscaler" "ory-oathkeeper" in namespace "kyma-system" has status "Ready"
    And there "is" "Secret" "ory-oathkeeper-jwks-secret" in namespace "kyma-system"
    And there "is" "Service" "ory-oathkeeper-api" in namespace "kyma-system"
    And there "is" "Service" "ory-oathkeeper-proxy" in namespace "kyma-system"
    And there "is" "Service" "ory-oathkeeper-maester-metrics" in namespace "kyma-system"
    And there "is" "ServiceAccount" "ory-oathkeeper" in namespace "kyma-system"
    And there "is" "ServiceAccount" "oathkeeper-maester-account" in namespace "kyma-system"
    And there "is" "ClusterRole" "oathkeeper-maester-role" in the cluster
    And there "is" "ClusterRoleBinding" "oathkeeper-maester-role-binding" in the cluster
    And there "is" "PeerAuthentication" "ory-oathkeeper-maester-metrics" in namespace "kyma-system"
    Then APIGateway CR "kyma-gateway" is removed
    And there "is no" "Deployment" "ory-oathkeeper" in namespace "kyma-system"
    And there "is no" "ConfigMap" "ory-oathkeeper-config" in namespace "kyma-system"
    And there "is no" "CustomResourceDefinition" "rules.oathkeeper.ory.sh" in the cluster
    And there "is no" "HorizontalPodAutoscaler" "ory-oathkeeper" in namespace "kyma-system"
    And there "is no" "Secret" "ory-oathkeeper-jwks-secret" in namespace "kyma-system"
    And there "is no" "Service" "ory-oathkeeper-api" in namespace "kyma-system"
    And there "is no" "Service" "ory-oathkeeper-proxy" in namespace "kyma-system"
    And there "is no" "Service" "ory-oathkeeper-maester-metrics" in namespace "kyma-system"
    And there "is no" "ServiceAccount" "ory-oathkeeper" in namespace "kyma-system"
    And there "is no" "ServiceAccount" "oathkeeper-maester-account" in namespace "kyma-system"
    And there "is no" "ClusterRole" "oathkeeper-maester-role" in the cluster
    And there "is no" "ClusterRoleBinding" "oathkeeper-maester-role-binding" in the cluster
    And there "is no" "PeerAuthentication" "ory-oathkeeper-maester-metrics" in namespace "kyma-system"

  Scenario: Kyma gateway is not removed when there is an APIRule
    Given APIGateway CR is in Ready state
    Then there is an "kyma-rule" APIRule
    Then disabling kyma gateway will result in "Warning" state
    And there is APIGateway CR "kyma-gateway" in "kyma-system" namespace
    And APIRule "kyma-rule" is removed

  Scenario: Kyma gateway is removed when there is no APIRule
    Given APIGateway CR "kyma-gateway" is removed
    Then gateway "kyma-gateway" in "kyma-system" namespace does not exist
