Feature: APIRules v2alpha1 conversion

  Scenario: Migrate v1beta1 APIRule with allow handler that is unsupported in v2alpha1
    Given migrationAllowV1beta1: There is a httpbin service with Istio injection enabled
    And migrationAllowV1beta1: The APIRule is applied
    And migrationAllowV1beta1: APIRule has status "OK"
    And migrationAllowV1beta1: Wait for "50" seconds
    When migrationAllowV1beta1: The APIRule is updated using manifest "migration-allow-v2alpha1.yaml"
    And migrationNoopV1beta1: Wait for "10" seconds
    Then migrationAllowV1beta1: APIRule has status "OK"

  Scenario: Migrate v1beta1 APIRule with no_auth handler that is supported in v2alpha1
    Given migrationNoAuthV1beta1: There is a httpbin service with Istio injection enabled
    And migrationNoAuthV1beta1: The APIRule is applied
    And migrationNoAuthV1beta1: APIRule has status "OK"
    And migrationNoAuthV1beta1: Wait for "50" seconds
    When migrationNoAuthV1beta1: The APIRule is updated using manifest "migration-noauth-v2alpha1.yaml"
    And migrationNoopV1beta1: Wait for "10" seconds
    Then migrationNoAuthV1beta1: APIRule has status "OK"

  Scenario: Migrate v1beta1 APIRule with noop handler that is unsupported in v2alpha1
    Given migrationNoopV1beta1: There is a httpbin service with Istio injection enabled
    And migrationNoopV1beta1: The APIRule is applied
    And migrationNoopV1beta1: APIRule has status "OK"
    And migrationNoopV1beta1: Wait for "50" seconds
    When migrationNoopV1beta1: The APIRule is updated using manifest "migration-noop-v2alpha1.yaml"
    And migrationNoopV1beta1: Resource of Kind "Rule" owned by APIRule does not exist
    Then migrationNoopV1beta1: APIRule has status "OK"

  Scenario: Migrate v1beta1 APIRule with jwt handler that is supported in v2alpha1
    Given migrationJwtV1beta1: There is a httpbin service with Istio injection enabled
    And migrationJwtV1beta1: The APIRule is applied
    And migrationJwtV1beta1: APIRule has status "OK"
    And migrationJwtV1beta1: Wait for "50" seconds
    When migrationJwtV1beta1: The APIRule is updated using manifest "migration-jwt-v2alpha1.yaml"
    And migrationJwtV1beta1: Resource of Kind "RequestAuthentication" owned by APIRule exists
    And migrationJwtV1beta1: Resource of Kind "AuthorizationPolicy" owned by APIRule exists
    And migrationJwtV1beta1: VirtualService owned by APIRule has httpbin service as destination
    And migrationJwtV1beta1: Resource of Kind "Rule" owned by APIRule does not exist
    Then migrationJwtV1beta1: APIRule has status "OK"

  Scenario: Delete v1beta1 APIRule with handler that is unsupported in v2alpha1
    Given deleteAllowV1beta1: The APIRule is applied
    And deleteAllowV1beta1: APIRule has status "OK"
    When deleteAllowV1beta1: The APIRule is deleted using v2alpha1
    Then deleteAllowV1beta1: APIRule is not found

  Scenario: Delete v1beta1 APIRule with handler that is supported in v2alpha1
    Given deleteNoAuthV1beta1: The APIRule is applied
    And deleteNoAuthV1beta1: APIRule has status "OK"
    When deleteNoAuthV1beta1: The APIRule is deleted using v2alpha1
    Then deleteNoAuthV1beta1: APIRule is not found
