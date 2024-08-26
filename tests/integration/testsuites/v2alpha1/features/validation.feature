Feature: Validation of APIRule

  Scenario: Validation errors on misconfigured APIRule
    When ValidationError: The misconfigured APIRule is applied
    Then ValidationError: APIRule has status "Error" with description containing "multiple jwt configurations"
