Feature: Exposing endpoints with JWT

  Scenario: Calling a httpbin endpoint secured
    Given The APIRule template file is set to "jwt-common.yaml"
    And There is a httpbin service
    And There is an endpoint secured with JWT on path "/ip"
    When The APIRule is applied
    Then Calling the "/ip" endpoint without a token should result in status between 400 and 403
    And Calling the "/ip" endpoint with an invalid token should result in status between 400 and 403
    And Calling the "/ip" endpoint with a valid "JWT" token should result in status between 200 and 299
    And Teardown httpbin service

  Scenario: Calling a httpbin endpoint secured on all paths
    Given The APIRule template file is set to "jwt-common.yaml"
    And There is a httpbin service
    And There is an endpoint secured with JWT on path "/*"
    When The APIRule is applied
    Then Calling the "/ip" endpoint without a token should result in status between 400 and 403
    And Calling the "/ip" endpoint with an invalid token should result in status between 400 and 403
    And Calling the "/ip" endpoint with a valid "JWT" token should result in status between 200 and 299
    Then Calling the "/headers" endpoint without a token should result in status between 400 and 403
    And Calling the "/headers" endpoint with an invalid token should result in status between 400 and 403
    And Calling the "/headers" endpoint with a valid "JWT" token should result in status between 200 and 299
    And Teardown httpbin service

  Scenario: Calling httpbin that has an endpoint secured by JWT and unrestricted endpoint
    Given The APIRule template file is set to "jwt-and-unrestricted.yaml"
    And There is a httpbin service
    And There is an endpoint secured with JWT on path "/ip" and /headers endpoint exposed with noAuth
    When The APIRule is applied
    Then Calling the "/ip" endpoint with a valid "JWT" token should result in status between 200 and 299
    And Calling the "/headers" endpoint without token should result in status between 200 and 299
    And Teardown httpbin service

  Scenario: Calling a httpbin endpoint secured with JWT that requires scopes claims
    Given The APIRule template file is set to "jwt-scopes.yaml"
    And There is a httpbin service
    And There is an endpoint secured with JWT on path "/ip" requiring scopes '["read", "write"]'
    And There is an endpoint secured with JWT on path "/get" requiring scopes '["test", "write"]'
    And There is an endpoint secured with JWT on path "/headers" requiring scopes '["read"]'
    When The APIRule is applied
    Then Calling the "/ip" endpoint with a valid "JWT" token with "scopes" "read" and "write" should result in status between 200 and 299
    And Calling the "/get" endpoint with a valid "JWT" token with "scopes" "read" and "write" should result in status between 400 and 403
    And Calling the "/headers" endpoint with a valid "JWT" token with "scopes" "read" and "write" should result in status between 200 and 299
    And Teardown httpbin service

  Scenario: Calling a httpbin endpoint secured with JWT that requires aud claim
    Given The APIRule template file is set to "jwt-audiences.yaml"
    And There is a httpbin service
    And There is an endpoint secured with JWT on path "/get" requiring audiences '["https://example.com"]'
    And There is an endpoint secured with JWT on path "/ip" requiring audiences '["https://example.com", "https://example.com/user"]'
    And There is an endpoint secured with JWT on path "/cache" requiring audience '["https://example.com"]' or '["audienceNotInJWT"]'
    And There is an endpoint secured with JWT on path "/headers" requiring audiences '["https://example.com", "https://example.com/admin"]'
    When The APIRule is applied
    Then Calling the "/get" endpoint with a valid "JWT" token with "audiences" "https://example.com" and "https://example.com/user" should result in status between 200 and 299
    And Calling the "/ip" endpoint with a valid "JWT" token with "audiences" "https://example.com" and "https://example.com/user" should result in status between 200 and 299
    And Calling the "/cache" endpoint with a valid "JWT" token with "audiences" "https://example.com" and "https://example.com/user" should result in status between 200 and 299
    And Calling the "/headers" endpoint with a valid "JWT" token with "audiences" "https://example.com" and "https://example.com/user" should result in status between 400 and 403
    And Teardown httpbin service

  Scenario: Exposing a JWT secured endpoint with unavailable issuer and jwks URL
    Given The APIRule template file is set to "jwt-unavailable-issuer.yaml"
    And There is a httpbin service
    Given There is an endpoint secured with JWT on path "/ip" with invalid issuer and jwks
    When The APIRule is applied
    And Calling the "/ip" endpoint with a valid "JWT" token should result in body containing "Jwt issuer is not configured"
    And Teardown httpbin service

  Scenario: Exposing a JWT secured endpoint where issuer URL doesn't belong to jwks URL
    Given The APIRule template file is set to "jwt-issuer-jwks-not-match.yaml"
    And There is a httpbin service
    And There is an endpoint secured with JWT on path "/ip" with invalid issuer and jwks
    When The APIRule is applied
    And Calling the "/ip" endpoint with a valid "JWT" token should result in body containing "Jwt verification fails"
    And Teardown httpbin service

  Scenario: Exposing an endpoint secured with different JWT token from headers
    Given The APIRule template file is set to "jwt-token-from-header.yaml"
    And Template value "JWTHeaderName" is set to "x-jwt-token"
    And There is a httpbin service
    When The APIRule is applied
    Then Calling the "/headers" endpoint without a token should result in status between 400 and 403
    And Calling the "/headers" endpoint with a valid "JWT" token from default header should result in status between 400 and 403
    And Calling the "/headers" endpoint with a valid "JWT" token from header "x-jwt-token" and prefix "JwtToken" should result in status between 200 and 299
    And Teardown httpbin service

  Scenario: Calling a httpbin endpoint secured with different JWT token from params
    Given The APIRule template file is set to "jwt-token-from-param.yaml"
    And Template value "FromParamName" is set to "jwt_token"
    And There is a httpbin service
    When The APIRule is applied
    And Calling the "/ip" endpoint without a token should result in status between 400 and 403
    And Calling the "/ip" endpoint with a valid "JWT" token from default header should result in status between 400 and 403
    And Calling the "/ip" endpoint with a valid "JWT" token from parameter "jwt_token" should result in status between 200 and 299
    And Teardown httpbin service
