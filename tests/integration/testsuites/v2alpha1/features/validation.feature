Feature: Validation of APIRule

  Scenario: Validation errors on misconfigured APIRule
    When The misconfigured APIRule is applied
    Then APIRule has status "Error" with description containing "multiple jwt configurations"
