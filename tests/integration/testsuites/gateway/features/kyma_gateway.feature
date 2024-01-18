Feature: Checking default kyma gateway
  Background:
    Given APIGateway CR is in "Ready" state with description ""

  Scenario: APIGateway is completely installed
    Given there "is" "CustomResourceDefinition" "apigateways.operator.kyma-project.io" in the cluster
    And there "is" "CustomResourceDefinition" "apirules.gateway.kyma-project.io" in the cluster
    And there "is" "Deployment" "api-gateway-controller-manager" in namespace "kyma-system"
    And there "is" "ServiceAccount" "api-gateway-controller-manager" in namespace "kyma-system"
    And there "is" "Role" "api-gateway-leader-election-role" in namespace "kyma-system"
    And there "is" "RoleBinding" "api-gateway-leader-election-rolebinding" in namespace "kyma-system"
    And there "is" "ClusterRole" "api-gateway-manager-role" in the cluster
    And there "is" "ClusterRoleBinding" "api-gateway-manager-rolebinding" in the cluster
    And there "is" "ConfigMap" "api-gateway-apirule-ui.operator.kyma-project.io" in namespace "kyma-system"
    And there "is" "ConfigMap" "api-gateway-ui.operator.kyma-project.io" in namespace "kyma-system"
    And there "is" "Service" "api-gateway-operator-metrics" in namespace "kyma-system"
    And there "is" "PriorityClass" "api-gateway-priority-class" in the cluster
    And there is Istio Gateway "kyma-gateway" in "kyma-system" namespace
    And there "is" "Certificate" "kyma-tls-cert" in namespace "istio-system"
    And there "is" "DNSEntry" "kyma-gateway" in namespace "kyma-system"
    And there "is" "VirtualService" "istio-healthz" in namespace "istio-system"
    Then there "is" "Deployment" "ory-oathkeeper" in namespace "kyma-system"
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
    Then APIGateway CR "test-gateway" is removed
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

  Scenario: Kyma Gateway is not removed when there is a VirtualService
    Given there is an "kyma-vs" VirtualService with Gateway "kyma-system/kyma-gateway"
    When disabling Kyma gateway
    Then APIGateway CR is in "Warning" state with description "There are custom resources that block the deletion of Kyma Gateway. Please take a look at kyma-system/api-gateway-controller-manager logs to see more information about the warning"
    And there is Istio Gateway "kyma-gateway" in "kyma-system" namespace
    And VirtualService "kyma-vs" is removed
    And APIGateway CR is in "Ready" state with description ""

  Scenario: Kyma Gateway is removed when there is no blocking resources
    When APIGateway CR "test-gateway" is removed
    Then APIGateway CR "test-gateway" "is not" present
    And gateway "kyma-gateway" in "kyma-system" namespace does not exist

  Scenario: Second APIGateway CR is applied to the cluster
    When APIGateway CR "second-api-gateway-cr" is applied
    Then Custom APIGateway CR "second-api-gateway-cr" is in "Error" state with description "stopped APIGateway CR reconciliation: only APIGateway CR test-gateway reconciles the module"
    And APIGateway CR is in "Ready" state with description ""
    And APIGateway CR "second-api-gateway-cr" is removed