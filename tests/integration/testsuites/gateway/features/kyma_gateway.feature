Feature: Checking default kyma gateway
  Background:
    Given APIGateway CR is applied
    Then APIGateway CR is in "Ready" state

  Scenario: Oathkeeper is installed when there is APIGateway CR present
    Given there is a "kyma-gateway" gateway in "kyma-system" namespace
    Then there is a "Deployment" "ory-oathkeeper" in namespace "kyma-system"
    And "Deployment" "ory-oathkeeper" in namespace "kyma-system" has status "Ready"
    And there is a "ConfigMap" "ory-oathkeeper-config" in namespace "kyma-system"
    And there is a "CustomResourceDefinition" "rules.oathkeeper.ory.sh" in the cluster
    And there is a "HorizontalPodAutoscaler" "ory-oathkeeper" in namespace "kyma-system"
    And "HorizontalPodAutoscaler" "ory-oathkeeper" in namespace "kyma-system" has status "Ready"
    And there is a "Secret" "ory-oathkeeper-jwks-secret" in namespace "kyma-system"
    And there is a "Service" "ory-oathkeeper-api" in namespace "kyma-system"
    And there is a "Service" "ory-oathkeeper-proxy" in namespace "kyma-system"
    And there is a "Service" "ory-oathkeeper-maester-metrics" in namespace "kyma-system"
    And there is a "ServiceAccount" "ory-oathkeeper" in namespace "kyma-system"
    And there is a "ServiceAccount" "oathkeeper-maester-account" in namespace "kyma-system"
    And there is a "ClusterRole" "oathkeeper-maester-role" in the cluster
    And there is a "ClusterRoleBinding" "oathkeeper-maester-role-binding" in the cluster
    And there is a "PeerAuthentication" "ory-oathkeeper-maester-metrics" in namespace "kyma-system"

  Scenario: Kyma gateway is not removed when there is an APIRule
    Given there is an "kyma-rule" APIRule
    Then disabling kyma gateway will result in "Warning" state
    And there is a "kyma-gateway" gateway in "kyma-system" namespace
    And APIRule "kyma-rule" is removed

  Scenario: Kyma gateway is removed when there is no APIRule
    Given gateway "kyma-gateway" is removed
    Then gateway "kyma-gateway" in "kyma-system" namespace does not exist
