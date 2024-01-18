Feature: Checking default kyma gateway on gardener
  Background:
    Given APIGateway CR is in "Ready" state with description ""

  Scenario: API Gateway is completely deployed
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
    And there "is" "DNSEntry" "kyma-gateway" in namespace "kyma-system"
    And there is a "kyma-gateway-certs" secret in "istio-system" namespace
    And there is a "kyma-tls-cert" Gardener Certificate CR in "istio-system" namespace
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
