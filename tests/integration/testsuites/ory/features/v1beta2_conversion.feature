Feature: APIRules v1beta2 conversion

  Scenario: Migrate v1beta1 APIRule with allow handler that is unsupported in v1beta2
    Given migrationAllowV1beta1: The APIRule is applied
    And migrationAllowV1beta1: APIRule has status "OK"
    When migrationAllowV1beta1: The APIRule is updated using manifest "migration-allow-v1beta2.yaml"
    Then migrationAllowV1beta1: APIRule has status "OK"

  Scenario: Migrate v1beta1 APIRule with no_auth handler that is supported in v1beta2
    Given migrationNoAuthV1beta1: The APIRule is applied
    And migrationNoAuthV1beta1: APIRule has status "OK"
    When migrationNoAuthV1beta1: The APIRule is updated using manifest "migration-noauth-v1beta2.yaml"
    Then migrationNoAuthV1beta1: APIRule has status "OK"

  Scenario: Migrate v1beta1 APIRule with noop handler that is unsupported in v1beta2
    Given migrationNoopV1beta1: The APIRule is applied
    And migrationNoopV1beta1: APIRule has status "OK"
    When migrationNoopV1beta1: The APIRule is updated using manifest "migration-noop-v1beta2.yaml"
    Then migrationNoopV1beta1: APIRule has status "OK"

  Scenario: Migrate v1beta1 APIRule with jwt handler that is supported in v1beta2
    Given migrationJwtV1beta1: The APIRule is applied
    And migrationJwtV1beta1: APIRule has status "OK"
    When migrationJwtV1beta1: The APIRule is updated using manifest "migration-jwt-v1beta2.yaml"
    Then migrationJwtV1beta1: APIRule has status "OK"

  Scenario: Delete v1beta1 APIRule with handler that is unsupported in v1beta2
    Given deleteAllowV1beta1: The APIRule is applied
    When deleteAllowV1beta1: The APIRule is deleted using v1beta2
    Then deleteAllowV1beta1: APIRule is not found

  Scenario: Delete v1beta1 APIRule with handler that is supported in v1beta2
    Given deleteNoAuthV1beta1: The APIRule is applied
    When deleteNoAuthV1beta1: The APIRule is deleted using v1beta2
    Then deleteNoAuthV1beta1: APIRule is not found
