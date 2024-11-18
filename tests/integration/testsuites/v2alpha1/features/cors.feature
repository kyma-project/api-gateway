Feature: CORS

  Scenario: No headers are returned when CORS is not specified in APIRule
    Given There is an httpbin service
    And The APIRule without CORS set up is applied
    Then Preflight calling the "/ip" endpoint with header Origin:"localhost" should result in status code 200 and no response header "Access-Control-Allow-Origin"
    And Preflight calling the "/ip" endpoint with header Origin:"localhost" should result in status code 200 and no response header "Access-Control-Allow-Methods"
    And Preflight calling the "/ip" endpoint with header Origin:"localhost" should result in status code 200 and no response header "Access-Control-Allow-Headers"
    And Preflight calling the "/ip" endpoint with header Origin:"localhost" should result in status code 200 and no response header "Access-Control-Expose-Headers"
    And Preflight calling the "/ip" endpoint with header Origin:"localhost" should result in status code 200 and no response header "Access-Control-Allow-Credentials"
    And Preflight calling the "/ip" endpoint with header Origin:"localhost" should result in status code 200 and no response header "Access-Control-Max-Age"
    And Teardown httpbin service

  Scenario: CORS is set up to custom values in APIRule
    Given There is an httpbin service
    And The APIRule with following CORS setup is applied AllowOrigins:'["regex": ".*local.kyma.dev"]', AllowMethods:'["GET", "POST"]', AllowHeaders:'["x-custom-allow-headers"]', AllowCredentials:"false", ExposeHeaders:'["x-custom-expose-headers"]', MaxAge:"300"
    Then Preflight calling the "/ip" endpoint with header Origin:"test.local.kyma.dev" should result in status code 200 and response header "Access-Control-Allow-Origin" with value "test.local.kyma.dev"
    And Preflight calling the "/ip" endpoint with header Origin:"localhost" should result in status code 200 and no response header "Access-Control-Allow-Origin"
    And Preflight calling the "/ip" endpoint with header Origin:"a.local.kyma.dev" should result in status code 200 and response header "Access-Control-Allow-Methods" with value "GET,POST"
    And Preflight calling the "/ip" endpoint with header Origin:"b.local.kyma.dev" should result in status code 200 and response header "Access-Control-Allow-Headers" with value "x-custom-allow-headers"
    And Preflight calling the "/ip" endpoint with header Origin:"c.local.kyma.dev" should result in status code 200 and response header "Access-Control-Expose-Headers" with value "x-custom-expose-headers"
    And Preflight calling the "/ip" endpoint with header Origin:"d.local.kyma.dev" should result in status code 200 and response header "Access-Control-Max-Age" with value "300"
    Then Teardown httpbin service
