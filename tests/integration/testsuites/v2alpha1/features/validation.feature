Feature: Validation of APIRule

  Scenario: Validation errors on misconfigured APIRule
    Then ValidationError: APIRule is applied and contains error status with "multiple jwt configurations" message
