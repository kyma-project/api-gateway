Feature: APIRules v2alpha1 conversion

  Scenario: Migrate v1beta1 APIRule with allow handler that is unsupported in v2alpha1
    Given migrationAllowV1beta1: There is a httpbin service with Istio injection enabled
    And migrationAllowV1beta1: The APIRule is applied
    And migrationAllowV1beta1: APIRule has status "OK"
    And migrationAllowV1beta1: Calling the "/headers" endpoint without a token should result in status between 200 and 200
    When migrationAllowV1beta1: The APIRule is updated using manifest "migration-allow-v2alpha1.yaml"
    And migrationAllowV1beta1: Resource of Kind "AuthorizationPolicy" owned by APIRule exists
    Then migrationAllowV1beta1: APIRule has status "OK"

  Scenario: Migrate v1beta1 APIRule with no_auth handler that is supported in v2alpha1
    Given migrationNoAuthV1beta1: There is a httpbin service with Istio injection enabled
    And migrationNoAuthV1beta1: The APIRule is applied
    And migrationNoAuthV1beta1: APIRule has status "OK"
    And migrationNoAuthV1beta1: Calling the "/headers" endpoint without a token should result in status between 200 and 200
    When migrationNoAuthV1beta1: The APIRule is updated using manifest "migration-noauth-v2alpha1.yaml"
    And migrationNoAuthV1beta1: Resource of Kind "AuthorizationPolicy" owned by APIRule exists
    Then migrationNoAuthV1beta1: APIRule has status "OK"

  Scenario: Migrate v1beta1 APIRule with noop handler that is unsupported in v2alpha1
    Given migrationNoopV1beta1: There is a httpbin service with Istio injection enabled
    And migrationNoopV1beta1: The APIRule is applied
    And migrationNoopV1beta1: APIRule has status "OK"
    And migrationNoopV1beta1: Calling the "/headers" endpoint without a token should result in status between 200 and 200
    When migrationNoopV1beta1: The APIRule is updated using manifest "migration-noop-v2alpha1.yaml"
    And migrationNoopV1beta1: Resource of Kind "AuthorizationPolicy" owned by APIRule exists
    And migrationNoopV1beta1: Resource of Kind "Rule" owned by APIRule does not exist
    Then migrationNoopV1beta1: APIRule has status "OK"

  Scenario: Migrate v1beta1 APIRule with jwt handler that is supported in v2alpha1
    Given migrationJwtV1beta1: There is a httpbin service with Istio injection enabled
    And migrationJwtV1beta1: The APIRule is applied
    And migrationJwtV1beta1: APIRule has status "OK"
    And migrationJwtV1beta1: Calling the "/headers" endpoint with a valid "JWT" token should result in status between 200 and 200
    When migrationJwtV1beta1: The APIRule is updated using manifest "migration-jwt-v2alpha1.yaml"
    And migrationJwtV1beta1: Resource of Kind "RequestAuthentication" owned by APIRule exists
    And migrationJwtV1beta1: Resource of Kind "AuthorizationPolicy" owned by APIRule exists
    And migrationJwtV1beta1: VirtualService owned by APIRule has httpbin service as destination
    And migrationJwtV1beta1: Resource of Kind "Rule" owned by APIRule does not exist
    Then migrationJwtV1beta1: APIRule has status "OK"

  Scenario: Migrate v1beta1 APIRule with oauth2_introspection handler that is unsupported in v2alpha1
    Given migrationOAuth2IntrospectionJwtV1beta1: There is a httpbin service with Istio injection enabled
    And migrationOAuth2IntrospectionJwtV1beta1: The APIRule is applied
    And migrationOAuth2IntrospectionJwtV1beta1: APIRule has status "OK"
    And migrationOAuth2IntrospectionJwtV1beta1: Calling the "/headers" endpoint with a valid "JWT" token should result in status between 200 and 200
    When migrationOAuth2IntrospectionJwtV1beta1: The APIRule is updated using manifest "migration-oauth2-introspection-v2alpha1.yaml"
    And migrationOAuth2IntrospectionJwtV1beta1: Resource of Kind "RequestAuthentication" owned by APIRule exists
    And migrationOAuth2IntrospectionJwtV1beta1: Resource of Kind "AuthorizationPolicy" owned by APIRule exists
    And migrationOAuth2IntrospectionJwtV1beta1: VirtualService owned by APIRule has httpbin service as destination
    And migrationOAuth2IntrospectionJwtV1beta1: Resource of Kind "Rule" owned by APIRule does not exist
    Then migrationOAuth2IntrospectionJwtV1beta1: APIRule has status "OK"

  Scenario: Migrate v1beta1 APIRule with annotations with allow handler that is unsupported in v2alpha1
    Given migrationAllowV1beta1: There is a httpbin service with Istio injection enabled
    And migrationAllowV1beta1: The APIRule is applied
    And migrationAllowV1beta1: APIRule has status "OK"
    And migrationAllowV1beta1: Calling the "/headers" endpoint without a token should result in status between 200 and 200
    When migrationAllowV1beta1: The APIRule is updated using manifest "migration-allow-v2alpha1-with-annotations.yaml"
    And migrationAllowV1beta1: Resource of Kind "AuthorizationPolicy" owned by APIRule exists
    Then migrationAllowV1beta1: APIRule has status "OK"

  Scenario: Migrate v1beta1 APIRule with annotations with no_auth handler that is supported in v2alpha1
    Given migrationNoAuthV1beta1: There is a httpbin service with Istio injection enabled
    And migrationNoAuthV1beta1: The APIRule is applied
    And migrationNoAuthV1beta1: APIRule has status "OK"
    And migrationNoAuthV1beta1: Calling the "/headers" endpoint without a token should result in status between 200 and 200
    When migrationNoAuthV1beta1: The APIRule is updated using manifest "migration-noauth-v2alpha1-with-annotations.yaml"
    And migrationNoAuthV1beta1: Resource of Kind "AuthorizationPolicy" owned by APIRule exists
    Then migrationNoAuthV1beta1: APIRule has status "OK"

  Scenario: Migrate v1beta1 APIRule with annotations with noop handler that is unsupported in v2alpha1
    Given migrationNoopV1beta1: There is a httpbin service with Istio injection enabled
    And migrationNoopV1beta1: The APIRule is applied
    And migrationNoopV1beta1: APIRule has status "OK"
    And migrationNoopV1beta1: Calling the "/headers" endpoint without a token should result in status between 200 and 200
    When migrationNoopV1beta1: The APIRule is updated using manifest "migration-noop-v2alpha1-with-annotations.yaml"
    And migrationNoopV1beta1: Resource of Kind "AuthorizationPolicy" owned by APIRule exists
    And migrationNoopV1beta1: Resource of Kind "Rule" owned by APIRule does not exist
    Then migrationNoopV1beta1: APIRule has status "OK"

  Scenario: Migrate v1beta1 APIRule with annotations with jwt handler that is supported in v2alpha1
    Given migrationJwtV1beta1: There is a httpbin service with Istio injection enabled
    And migrationJwtV1beta1: The APIRule is applied
    And migrationJwtV1beta1: APIRule has status "OK"
    And migrationJwtV1beta1: Calling the "/headers" endpoint with a valid "JWT" token should result in status between 200 and 200
    When migrationJwtV1beta1: The APIRule is updated using manifest "migration-jwt-v2alpha1-with-annotations.yaml"
    And migrationJwtV1beta1: Resource of Kind "RequestAuthentication" owned by APIRule exists
    And migrationJwtV1beta1: Resource of Kind "AuthorizationPolicy" owned by APIRule exists
    And migrationJwtV1beta1: VirtualService owned by APIRule has httpbin service as destination
    And migrationJwtV1beta1: Resource of Kind "Rule" owned by APIRule does not exist
    Then migrationJwtV1beta1: APIRule has status "OK"

  Scenario: Migrate v1beta1 APIRule with annotations with oauth2_introspection handler that is unsupported in v2alpha1
    Given migrationOAuth2IntrospectionJwtV1beta1: There is a httpbin service with Istio injection enabled
    And migrationOAuth2IntrospectionJwtV1beta1: The APIRule is applied
    And migrationOAuth2IntrospectionJwtV1beta1: APIRule has status "OK"
    And migrationOAuth2IntrospectionJwtV1beta1: Calling the "/headers" endpoint with a valid "JWT" token should result in status between 200 and 200
    When migrationOAuth2IntrospectionJwtV1beta1: The APIRule is updated using manifest "migration-oauth2-introspection-v2alpha1-with-annotations.yaml"
    And migrationOAuth2IntrospectionJwtV1beta1: Resource of Kind "RequestAuthentication" owned by APIRule exists
    And migrationOAuth2IntrospectionJwtV1beta1: Resource of Kind "AuthorizationPolicy" owned by APIRule exists
    And migrationOAuth2IntrospectionJwtV1beta1: VirtualService owned by APIRule has httpbin service as destination
    And migrationOAuth2IntrospectionJwtV1beta1: Resource of Kind "Rule" owned by APIRule does not exist
    Then migrationOAuth2IntrospectionJwtV1beta1: APIRule has status "OK"

  Scenario: Migrate v1beta1 APIRule with short host with annotations with oauth2_introspection handler that is unsupported in v2alpha1
    Given migrationOAuth2IntrospectionShortHostV1beta1: There is a httpbin service with Istio injection enabled
    And migrationOAuth2IntrospectionShortHostV1beta1: The APIRule is applied
    And migrationOAuth2IntrospectionShortHostV1beta1: APIRule has status "OK"
    And migrationOAuth2IntrospectionShortHostV1beta1: Calling the "/headers" endpoint with a valid "JWT" token should result in status between 200 and 200
    When migrationOAuth2IntrospectionShortHostV1beta1: The APIRule is updated using manifest "migration-oauth2-introspection-v2alpha1-with-annotations-with-short-host.yaml"
    And migrationOAuth2IntrospectionShortHostV1beta1: Resource of Kind "RequestAuthentication" owned by APIRule exists
    And migrationOAuth2IntrospectionShortHostV1beta1: Resource of Kind "AuthorizationPolicy" owned by APIRule exists
    And migrationOAuth2IntrospectionShortHostV1beta1: VirtualService owned by APIRule has httpbin service as destination
    And migrationOAuth2IntrospectionShortHostV1beta1: Resource of Kind "Rule" owned by APIRule does not exist
    Then migrationOAuth2IntrospectionShortHostV1beta1: APIRule has status "OK"

  Scenario: Delete v1beta1 APIRule with handler that is unsupported in v2alpha1
    Given deleteAllowV1beta1: The APIRule is applied
    And deleteAllowV1beta1: APIRule has status "OK"
    When deleteAllowV1beta1: The APIRule is deleted using v2alpha1
    Then deleteAllowV1beta1: APIRule is not found

  Scenario: Delete v1beta1 APIRule with handler that is supported in v2alpha1
    Given deleteNoAuthV1beta1: There is a httpbin service with Istio injection enabled
    Given deleteNoAuthV1beta1: The APIRule is applied
    And deleteNoAuthV1beta1: APIRule has status "OK"
    When deleteNoAuthV1beta1: The APIRule is deleted using v2alpha1
    Then deleteNoAuthV1beta1: APIRule is not found

  Scenario: Reconciliation error stops migration from progressing
    # Does not deploy httpbin service to make sure that an Error will
    # happen on start of migration. v1beta1 does not check for the target
    # so it will be Ready, but v2/v2alpha1 does, and will Error.
    Given migrationError: The APIRule is applied
    And migrationError: APIRule has status "Ready"
    And migrationError: The APIRule contains original-version annotation set to "v1beta1"
    When migrationError: The APIRule is updated using manifest "migration-noop-v2.yaml"
    Then migrationError: APIRule has status "Error"
    And migrationError: Resource of Kind "AuthorizationPolicy" owned by APIRule does not exist
    And migrationError: Resource of Kind "Rule" owned by APIRule exists
    And migrationError: The APIRule contains original-version annotation set to "v2"
