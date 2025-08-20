Feature: APIRules v2alpha1 conversion zero downtime

# v1beta1 (allow) -> v2alpha1 (allow)
  Scenario: Zero downtime migration v1beta1 APIRule with allow handler that is unsupported in v2alpha1
    Given migrationAllowV1beta1: There is a httpbin service with Istio injection enabled
    And migrationAllowV1beta1: The APIRule is applied
    And migrationAllowV1beta1: APIRule has status "Ready"
    And migrationAllowV1beta1: The APIRule contains original-version annotation set to "v1beta1"
    And migrationAllowV1beta1: Calling the "/headers" endpoint without a token should result in status between 200 and 200
    And migrationAllowV1beta1: There are continuous requests to path "/headers"
    When migrationAllowV1beta1: The APIRule is updated using manifest "migration-allow-v2alpha1.yaml"
    And migrationAllowV1beta1: Resource of Kind "AuthorizationPolicy" owned by APIRule exists
    Then migrationAllowV1beta1: APIRule has status "Ready"
    And migrationAllowV1beta1: The APIRule contains original-version annotation set to "v2alpha1"
    And migrationAllowV1beta1: All continuous requests should succeed

# v1beta1 (no_auth) -> v2alpha1 (no_auth)
  Scenario: Zero downtime migration v1beta1 APIRule with no_auth handler that is supported in v2alpha1
    Given migrationNoAuthV1beta1: There is a httpbin service with Istio injection enabled
    And migrationNoAuthV1beta1: The APIRule is applied
    And migrationNoAuthV1beta1: APIRule has status "Ready"
    And migrationNoAuthV1beta1: The APIRule contains original-version annotation set to "v1beta1"
    And migrationNoAuthV1beta1: Calling the "/headers" endpoint without a token should result in status between 200 and 200
    And migrationNoAuthV1beta1: There are continuous requests to path "/headers"
    When migrationNoAuthV1beta1: The APIRule is updated using manifest "migration-noauth-v2alpha1.yaml"
    And migrationNoAuthV1beta1: Resource of Kind "AuthorizationPolicy" owned by APIRule exists
    Then migrationNoAuthV1beta1: APIRule has status "Ready"
    And migrationNoAuthV1beta1: The APIRule contains original-version annotation set to "v2alpha1"
    And migrationNoAuthV1beta1: All continuous requests should succeed

# v1beta1 (noop) -> v2alpha1 (noop)
  Scenario: Zero downtime migration v1beta1 APIRule with noop handler that is unsupported in v2alpha1
    Given migrationNoopV1beta1: There is a httpbin service with Istio injection enabled
    And migrationNoopV1beta1: The APIRule is applied
    And migrationNoopV1beta1: APIRule has status "Ready"
    And migrationNoopV1beta1: The APIRule contains original-version annotation set to "v1beta1"
    And migrationNoopV1beta1: Calling the "/headers" endpoint without a token should result in status between 200 and 200
    And migrationNoopV1beta1: There are continuous requests to path "/headers"
    When migrationNoopV1beta1: The APIRule is updated using manifest "migration-noop-v2alpha1.yaml"
    And migrationNoopV1beta1: Resource of Kind "AuthorizationPolicy" owned by APIRule exists
    And migrationNoopV1beta1: Resource of Kind "Rule" owned by APIRule does not exist
    Then migrationNoopV1beta1: APIRule has status "Ready"
    And migrationNoopV1beta1: The APIRule contains original-version annotation set to "v2alpha1"
    And migrationNoopV1beta1: All continuous requests should succeed

# v1beta1 (jwt) -> v2alpha1 (jwt)
  Scenario: Zero downtime migration v1beta1 APIRule with jwt handler that is supported in v2alpha1
    Given migrationJwtV1beta1: There is a httpbin service with Istio injection enabled
    And migrationJwtV1beta1: The APIRule is applied
    And migrationJwtV1beta1: APIRule has status "Ready"
    And migrationJwtV1beta1: The APIRule contains original-version annotation set to "v1beta1"
    And migrationJwtV1beta1: Calling the "/headers" endpoint with a valid "JWT" token should result in status between 200 and 200
    And migrationJwtV1beta1: There are continuous requests to path "/headers"
    When migrationJwtV1beta1: The APIRule is updated using manifest "migration-jwt-v2alpha1.yaml"
    And migrationJwtV1beta1: Resource of Kind "RequestAuthentication" owned by APIRule exists
    And migrationJwtV1beta1: Resource of Kind "AuthorizationPolicy" owned by APIRule exists
    And migrationJwtV1beta1: VirtualService owned by APIRule has httpbin service as destination
    And migrationJwtV1beta1: Resource of Kind "Rule" owned by APIRule does not exist
    Then migrationJwtV1beta1: APIRule has status "Ready"
    And migrationJwtV1beta1: The APIRule contains original-version annotation set to "v2alpha1"
    And migrationJwtV1beta1: All continuous requests should succeed

# v1beta1 (extauth) -> v2alpha1 (extauth)
  Scenario: Zero downtime migration v1beta1 APIRule with oauth2_introspection handler that is unsupported in v2alpha1
    Given migrationOAuth2IntrospectionJwtV1beta1: There is a httpbin service with Istio injection enabled
    And migrationOAuth2IntrospectionJwtV1beta1: The APIRule is applied
    And migrationOAuth2IntrospectionJwtV1beta1: APIRule has status "Ready"
    And migrationOAuth2IntrospectionJwtV1beta1: The APIRule contains original-version annotation set to "v1beta1"
    And migrationOAuth2IntrospectionJwtV1beta1: Calling the "/headers" endpoint with a valid "JWT" token should result in status between 200 and 200
    And migrationOAuth2IntrospectionJwtV1beta1: There are continuous requests to path "/headers"
    When migrationOAuth2IntrospectionJwtV1beta1: The APIRule is updated using manifest "migration-oauth2-introspection-v2alpha1.yaml"
    And migrationOAuth2IntrospectionJwtV1beta1: Resource of Kind "RequestAuthentication" owned by APIRule exists
    And migrationOAuth2IntrospectionJwtV1beta1: Resource of Kind "AuthorizationPolicy" owned by APIRule exists
    And migrationOAuth2IntrospectionJwtV1beta1: VirtualService owned by APIRule has httpbin service as destination
    And migrationOAuth2IntrospectionJwtV1beta1: Resource of Kind "Rule" owned by APIRule does not exist
    Then migrationOAuth2IntrospectionJwtV1beta1: APIRule has status "Ready"
    And migrationOAuth2IntrospectionJwtV1beta1: The APIRule contains original-version annotation set to "v2alpha1"
    And migrationOAuth2IntrospectionJwtV1beta1: All continuous requests should succeed
