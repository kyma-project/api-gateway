Feature: CORS

  Scenario: No headers are returned when CORS is not specified in APIRule
    Given DefaultCORS: There is an httpbin service
    And DefaultCORS: The APIRule without CORS set up is applied
    Then DefaultCORS: Preflight calling the "/ip" endpoint with header Origin:"localhost" should result in status code 200 and no response header "Access-Control-Allow-Origin"
    And DefaultCORS: Preflight calling the "/ip" endpoint with header Origin:"localhost" should result in status code 200 and no response header "Access-Control-Allow-Methods"
    And DefaultCORS: Preflight calling the "/ip" endpoint with header Origin:"localhost" should result in status code 200 and no response header "Access-Control-Allow-Headers"
    And DefaultCORS: Preflight calling the "/ip" endpoint with header Origin:"localhost" should result in status code 200 and no response header "Access-Control-Expose-Headers"
    And DefaultCORS: Preflight calling the "/ip" endpoint with header Origin:"localhost" should result in status code 200 and no response header "Access-Control-Allow-Credentials"
    And DefaultCORS: Preflight calling the "/ip" endpoint with header Origin:"localhost" should result in status code 200 and no response header "Access-Control-Max-Age"
    And DefaultCORS: Teardown httpbin service

  Scenario: CORS is set up to custom values in APIRule
    Given CustomCORS: There is an httpbin service
    And CustomCORS: The APIRule with following CORS setup is applied AllowOrigins:'["regex": ".*local.kyma.dev"]', AllowMethods:'["GET", "POST"]', AllowHeaders:'["x-custom-allow-headers"]', AllowCredentials:"false", ExposeHeaders:'["x-custom-expose-headers"]', MaxAge:"300"
    Then CustomCORS: Preflight calling the "/ip" endpoint with header Origin:"test.local.kyma.dev" should result in status code 200 and response header "Access-Control-Allow-Origin" with value "test.local.kyma.dev"
    And CustomCORS: Preflight calling the "/ip" endpoint with header Origin:"localhost" should result in status code 200 and no response header "Access-Control-Allow-Origin"
    And CustomCORS: Preflight calling the "/ip" endpoint with header Origin:"a.local.kyma.dev" should result in status code 200 and response header "Access-Control-Allow-Methods" with value "GET,POST"
    And CustomCORS: Preflight calling the "/ip" endpoint with header Origin:"b.local.kyma.dev" should result in status code 200 and response header "Access-Control-Allow-Headers" with value "x-custom-allow-headers"
    And CustomCORS: Preflight calling the "/ip" endpoint with header Origin:"c.local.kyma.dev" should result in status code 200 and response header "Access-Control-Expose-Headers" with value "x-custom-expose-headers"
    And CustomCORS: Preflight calling the "/ip" endpoint with header Origin:"d.local.kyma.dev" should result in status code 200 and response header "Access-Control-Max-Age" with value "300"
    Then CustomCORS: Teardown httpbin service
