Feature: Exposing endpoints with JWT

  Scenario: Calling a httpbin endpoint secured
    Given Common: There is a httpbin service
    And Common: There is an endpoint secured with JWT on path "/ip"
    When Common: The APIRule is applied
    Then Common: Calling the "/ip" endpoint without a token should result in status between 400 and 403
    And Common: Calling the "/ip" endpoint with an invalid token should result in status between 400 and 403
    And Common: Calling the "/ip" endpoint with a valid "JWT" token should result in status between 200 and 299
    And Common: Teardown httpbin service

  Scenario: Calling a httpbin endpoint secured on all paths
    Given Wildcard: There is a httpbin service
    And Wildcard: There is an endpoint secured with JWT on path "/*"
    When Wildcard: The APIRule is applied
    Then Wildcard: Calling the "/ip" endpoint without a token should result in status between 400 and 403
    And Wildcard: Calling the "/ip" endpoint with an invalid token should result in status between 400 and 403
    And Wildcard: Calling the "/ip" endpoint with a valid "JWT" token should result in status between 200 and 299
    Then Wildcard: Calling the "/headers" endpoint without a token should result in status between 400 and 403
    And Wildcard: Calling the "/headers" endpoint with an invalid token should result in status between 400 and 403
    And Wildcard: Calling the "/headers" endpoint with a valid "JWT" token should result in status between 200 and 299
    And Wildcard: Teardown httpbin service

  Scenario: Calling httpbin that has an endpoint secured by JWT and unrestricted endpoint
    Given JwtAndUnrestricted: There is a httpbin service
    And JwtAndUnrestricted: There is an endpoint secured with JWT on path "/ip" and /headers endpoint exposed with noAuth
    When JwtAndUnrestricted: The APIRule is applied
    Then JwtAndUnrestricted: Calling the "/ip" endpoint with a valid "JWT" token should result in status between 200 and 299
    And JwtAndUnrestricted: Calling the "/headers" endpoint without token should result in status between 200 and 299
    And JwtAndUnrestricted: Teardown httpbin service

  Scenario: Calling a httpbin endpoint secured with JWT that requires scopes claims
    Given JwtScopes: There is a httpbin service
    And JwtScopes: There is an endpoint secured with JWT on path "/ip" requiring scopes '["read", "write"]'
    And JwtScopes: There is an endpoint secured with JWT on path "/get" requiring scopes '["test", "write"]'
    And JwtScopes: There is an endpoint secured with JWT on path "/headers" requiring scopes '["read"]'
    When JwtScopes: The APIRule is applied
    Then JwtScopes: Calling the "/ip" endpoint with a valid "JWT" token with "scopes" "read" and "write" should result in status between 200 and 299
    And JwtScopes: Calling the "/get" endpoint with a valid "JWT" token with "scopes" "read" and "write" should result in status between 400 and 403
    And JwtScopes: Calling the "/headers" endpoint with a valid "JWT" token with "scopes" "read" and "write" should result in status between 200 and 299
    And JwtScopes: Teardown httpbin service

  Scenario: Calling a httpbin endpoint secured with JWT that requires aud claim
    Given JwtAudiences: There is a httpbin service
    And JwtAudiences: There is an endpoint secured with JWT on path "/get" requiring audiences '["https://example.com"]'
    And JwtAudiences: There is an endpoint secured with JWT on path "/ip" requiring audiences '["https://example.com", "https://example.com/user"]'
    And JwtAudiences: There is an endpoint secured with JWT on path "/cache" requiring audience '["https://example.com"]' or '["audienceNotInJWT"]'
    And JwtAudiences: There is an endpoint secured with JWT on path "/headers" requiring audiences '["https://example.com", "https://example.com/admin"]'
    When JwtAudiences: The APIRule is applied
    Then JwtAudiences: Calling the "/get" endpoint with a valid "JWT" token with "audiences" "https://example.com" and "https://example.com/user" should result in status between 200 and 299
    And JwtAudiences: Calling the "/ip" endpoint with a valid "JWT" token with "audiences" "https://example.com" and "https://example.com/user" should result in status between 200 and 299
    And JwtAudiences: Calling the "/cache" endpoint with a valid "JWT" token with "audiences" "https://example.com" and "https://example.com/user" should result in status between 200 and 299
    And JwtAudiences: Calling the "/headers" endpoint with a valid "JWT" token with "audiences" "https://example.com" and "https://example.com/user" should result in status between 400 and 403
    And JwtAudiences: Teardown httpbin service

  Scenario: Exposing a JWT secured endpoint with unavailable issuer and jwks URL
    Given JwtUnavailableIssuer: There is a httpbin service
    Given JwtUnavailableIssuer: There is an endpoint secured with JWT on path "/ip" with invalid issuer and jwks
    When JwtUnavailableIssuer: The APIRule is applied
    And JwtUnavailableIssuer: Calling the "/ip" endpoint with a valid "JWT" token should result in body containing "Jwt issuer is not configured"
    And JwtUnavailableIssuer: Teardown httpbin service

  Scenario: Exposing a JWT secured endpoint where issuer URL doesn't belong to jwks URL
    Given JwtIssuerJwksNotMatch: There is a httpbin service
    And JwtIssuerJwksNotMatch: There is an endpoint secured with JWT on path "/ip" with invalid issuer and jwks
    When JwtIssuerJwksNotMatch: The APIRule is applied
    And JwtIssuerJwksNotMatch: Calling the "/ip" endpoint with a valid "JWT" token should result in body containing "Jwt verification fails"
    And JwtIssuerJwksNotMatch: Teardown httpbin service

  Scenario: Exposing an endpoint secured with different JWT token from headers
    Given JwtTokenFromHeader: There is a httpbin service
    When JwtTokenFromHeader: The APIRule is applied
    Then JwtTokenFromHeader: Calling the "/headers" endpoint without a token should result in status between 400 and 403
    And JwtTokenFromHeader: Calling the "/headers" endpoint with a valid "JWT" token from default header should result in status between 400 and 403
    And JwtTokenFromHeader: Calling the "/headers" endpoint with a valid "JWT" token from header "x-jwt-token" and prefix "JwtToken" should result in status between 200 and 299
    And JwtTokenFromHeader: Teardown httpbin service

  Scenario: Calling a httpbin endpoint secured with different JWT token from params
    Given JwtTokenFromParam: There is a httpbin service
    When JwtTokenFromParam: The APIRule is applied
    And JwtTokenFromParam: Calling the "/ip" endpoint without a token should result in status between 400 and 403
    And JwtTokenFromParam: Calling the "/ip" endpoint with a valid "JWT" token from default header should result in status between 400 and 403
    And JwtTokenFromParam: Calling the "/ip" endpoint with a valid "JWT" token from parameter "jwt_token" should result in status between 200 and 299
    And JwtTokenFromParam: Teardown httpbin service

