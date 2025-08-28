Feature: Checking default kyma gateway
  Background:
    Given APIGateway CR is in "Ready" state with description ""

  Scenario: Oathkeeper is installed and uninstalled depending on APIGateway presence
    Given there is Istio Gateway "kyma-gateway" in "kyma-system" namespace
    Then there "is" "Deployment" "ory-oathkeeper" in namespace "kyma-system"
    And "Deployment" "ory-oathkeeper" in namespace "kyma-system" has status "Ready"
    And there "is" "ConfigMap" "ory-oathkeeper-config" in namespace "kyma-system"
    And there "is" "CustomResourceDefinition" "rules.oathkeeper.ory.sh" in the cluster
    And there "is" "Secret" "ory-oathkeeper-jwks-secret" in namespace "kyma-system"
    And there "is" "Service" "ory-oathkeeper-api" in namespace "kyma-system"
    And there "is" "Service" "ory-oathkeeper-proxy" in namespace "kyma-system"
    And there "is" "Service" "ory-oathkeeper-maester-metrics" in namespace "kyma-system"
    And there "is" "ServiceAccount" "ory-oathkeeper" in namespace "kyma-system"
    And there "is" "ServiceAccount" "oathkeeper-maester-account" in namespace "kyma-system"
    And there "is" "ClusterRole" "oathkeeper-maester-role" in the cluster
    And there "is" "ClusterRoleBinding" "oathkeeper-maester-role-binding" in the cluster
    And there "is" "PeerAuthentication" "ory-oathkeeper-maester-metrics" in namespace "kyma-system"
    And there "is" "PodDisruptionBudget" "ory-oathkeeper" in namespace "kyma-system"
    Then APIGateway CR "default" is removed
    And there "is no" "Deployment" "ory-oathkeeper" in namespace "kyma-system"
    And there "is no" "ConfigMap" "ory-oathkeeper-config" in namespace "kyma-system"
    And there "is no" "CustomResourceDefinition" "rules.oathkeeper.ory.sh" in the cluster
    And there "is no" "Secret" "ory-oathkeeper-jwks-secret" in namespace "kyma-system"
    And there "is no" "Service" "ory-oathkeeper-api" in namespace "kyma-system"
    And there "is no" "Service" "ory-oathkeeper-proxy" in namespace "kyma-system"
    And there "is no" "Service" "ory-oathkeeper-maester-metrics" in namespace "kyma-system"
    And there "is no" "ServiceAccount" "ory-oathkeeper" in namespace "kyma-system"
    And there "is no" "ServiceAccount" "oathkeeper-maester-account" in namespace "kyma-system"
    And there "is no" "ClusterRole" "oathkeeper-maester-role" in the cluster
    And there "is no" "ClusterRoleBinding" "oathkeeper-maester-role-binding" in the cluster
    And there "is no" "PeerAuthentication" "ory-oathkeeper-maester-metrics" in namespace "kyma-system"
    And there "is no" "PodDisruptionBudget" "ory-oathkeeper" in namespace "kyma-system"

  Scenario: Oathkeeper is not installed when there is no api-gateway-config ConfigMap
    Given APIGateway CR "default" is removed
    Then deprecated v1beta1 configmap "api-gateway-config.operator.kyma-project.io" in namespace "kyma-system" is removed
    When APIGateway CR "default" is applied
    Then APIGateway CR is in "Ready" state with description ""
    And there "is no" "Deployment" "ory-oathkeeper" in namespace "kyma-system"
    And there "is no" "ConfigMap" "ory-oathkeeper-config" in namespace "kyma-system"
    And there "is no" "CustomResourceDefinition" "rules.oathkeeper.ory.sh" in the cluster
    And there "is no" "Secret" "ory-oathkeeper-jwks-secret" in namespace "kyma-system"
    And there "is no" "Service" "ory-oathkeeper-api" in namespace "kyma-system"
    And there "is no" "Service" "ory-oathkeeper-proxy" in namespace "kyma-system"
    And there "is no" "Service" "ory-oathkeeper-maester-metrics" in namespace "kyma-system"
    And there "is no" "ServiceAccount" "ory-oathkeeper" in namespace "kyma-system"
    And there "is no" "ServiceAccount" "oathkeeper-maester-account" in namespace "kyma-system"
    And there "is no" "ClusterRole" "oathkeeper-maester-role" in the cluster
    And there "is no" "ClusterRoleBinding" "oathkeeper-maester-role-binding" in the cluster
    And there "is no" "PeerAuthentication" "ory-oathkeeper-maester-metrics" in namespace "kyma-system"
    And there "is no" "PodDisruptionBudget" "ory-oathkeeper" in namespace "kyma-system"

  Scenario: Kyma Gateway is not removed when there is a VirtualService
    Given there is an "kyma-vs" VirtualService with Gateway "kyma-system/kyma-gateway"
    When disabling Kyma gateway
    Then APIGateway CR is in "Warning" state with description "There are custom resources that block the deletion of Kyma Gateway. Please take a look at kyma-system/api-gateway-controller-manager logs to see more information about the warning"
    And there is Istio Gateway "kyma-gateway" in "kyma-system" namespace
    And VirtualService "kyma-vs" is removed
    And APIGateway CR is in "Ready" state with description ""

  Scenario: Kyma Gateway is removed when there is no blocking resources
    When APIGateway CR "default" is removed
    Then APIGateway CR "default" "is not" present
    And gateway "kyma-gateway" in "kyma-system" namespace does not exist

  Scenario: Second APIGateway CR is applied to the cluster
    When APIGateway CR "second-api-gateway-cr" is applied
    Then Custom APIGateway CR "second-api-gateway-cr" is in "Warning" state with description "stopped APIGateway CR reconciliation: only APIGateway CR default reconciles the module"
    And APIGateway CR is in "Ready" state with description ""
    And APIGateway CR "second-api-gateway-cr" is removed