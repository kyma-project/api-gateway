Feature: Apply and update APIRules v1beta1
# apply v1beta1 (noAuth)
  Scenario: Creating APIRule v1beta1 (noAuth)
    Given applyUpdateNoAuthV1beta1: There is a httpbin service with Istio injection enabled
    When applyUpdateNoAuthV1beta1: The APIRule is applied
    And applyUpdateNoAuthV1beta1: APIRule has status "OK"
    Then applyUpdateNoAuthV1beta1: Calling the "/anything" endpoint without a token should result in status between 200 and 200
    And applyUpdateNoAuthV1beta1: The APIRule contains original-version annotation set to "v1beta1"
    And applyUpdateNoAuthV1beta1: APIRule has status "OK"
    And applyUpdateNoAuthV1beta1: Resource of Kind "Rule" owned by APIRule exists
    And applyUpdateNoAuthV1beta1: Teardown httpbin service

# update v1beta1 (noAuth)
  Scenario: Updating an APIRule and calling a httpbin endpoint "/ip" (noAuth)
    Given applyUpdateNoAuthV1beta1: There is a httpbin service with Istio injection enabled
    And applyUpdateNoAuthV1beta1: The APIRule is applied
    And applyUpdateNoAuthV1beta1: APIRule has status "OK"
    And applyUpdateNoAuthV1beta1: The APIRule contains original-version annotation set to "v1beta1"
    When applyUpdateNoAuthV1beta1: The APIRule is updated using manifest "v1beta1-noop-updated.yaml"
    Then applyUpdateNoAuthV1beta1: Calling the "/ip" endpoint without a token should result in status between 200 and 200
    And applyUpdateNoAuthV1beta1: The APIRule contains original-version annotation set to "v1beta1"
    And applyUpdateNoAuthV1beta1: APIRule has status "OK"
    And applyUpdateNoAuthV1beta1: Resource of Kind "Rule" owned by APIRule exists
    And applyUpdateNoAuthV1beta1: Teardown httpbin service
