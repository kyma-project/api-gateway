Feature: Validation of APIRule

  Scenario: Validation errors on misconfigured APIRule
    Given The APIRule template file is set to "validation-error.yaml"
    When The misconfigured APIRule is applied
    Then APIRule has status "Error" with description containing "multiple jwt configurations"
