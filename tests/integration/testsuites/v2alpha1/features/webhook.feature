Feature: Webhook

  Scenario: Recover when conversion webhook certificate secret is rotated and verify by exposing GET method for "/anything" with noAuth
    Given WebhookCertificateRecovery: There is a httpbin service
    And WebhookCertificateRecovery: The APIRule is applied
    And WebhookCertificateRecovery: Calling the "/anything" endpoint with "GET" method with any token should result in status between 200 and 200
    When WebhookCertificateRecovery: Certificate secret is reset
    Then WebhookCertificateRecovery: Certificate secret is rotated
    And WebhookCertificateRecovery: Calling the "/anything" endpoint with "GET" method with any token should result in status between 200 and 200
    And WebhookCertificateRecovery: Teardown httpbin service
