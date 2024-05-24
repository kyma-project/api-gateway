Feature: APIRules v1beta2 conversion

  Scenario: Delete v1beta1 APIRule with handler that is unsupported in v1beta2
    Given v1beta1AllowDelete: The APIRule is applied
    When v1beta1AllowDelete: The APIRule is deleted using v1beta2
    Then v1beta1AllowDelete: APIRule is not found

  Scenario: Delete v1beta1 APIRule with handler that is supported in v1beta2
    Given v1beta1NoAuthDelete: The APIRule is applied
    When v1beta1NoAuthDelete: The APIRule is deleted using v1beta2
    Then v1beta1NoAuthDelete: APIRule is not found