Feature: CORS

  Scenario: CORS is set up to default when it is not specified in APIRule
    Given DefaultCORS: There is an httpbin service
    And DefaultCORS: The APIRule without CORS set up is applied
    Then DefaultCORS: Preflight calling the "/ip" endpoint with header Origin:"localhost" should result in status code 200 and response header "Access-Control-Allow-Origin" with value "localhost"
    And DefaultCORS: Preflight calling the "/ip" endpoint with header Origin:"localhost" should result in status code 200 and response header "Access-Control-Allow-Methods" with value "GET,POST,PUT,DELETE,PATCH"
    And DefaultCORS: Preflight calling the "/ip" endpoint with header Origin:"localhost" should result in status code 200 and response header "Access-Control-Allow-Headers" with value "Authorization,Content-Type,*"
    Then DefaultCORS: Teardown httpbin service

  Scenario: CORS is set up to custom values in APIRule
    Given CustomCORS: There is an httpbin service
    And CustomCORS: The APIRule with custom CORS setup is applied
    Then CustomCORS: Preflight calling the "/ip" endpoint with header Origin:"localhost" should result in status code 200 and response header "Access-Control-Allow-Origin" with value "localhost"
    And CustomCORS: Preflight calling the "/ip" endpoint with header Origin:"test" should result in status code 200 and no response header "Access-Control-Allow-Origin"
    And CustomCORS: Preflight calling the "/ip" endpoint with header Origin:"localhost" should result in status code 200 and response header "Access-Control-Allow-Methods" with value "GET,POST"
    And CustomCORS: Preflight calling the "/ip" endpoint with header Origin:"localhost" should result in status code 200 and response header "Access-Control-Allow-Headers" with value "x-custom-allow-headers"
    And CustomCORS: Preflight calling the "/ip" endpoint with header Origin:"localhost" should result in status code 200 and response header "Access-Control-Expose-Headers" with value "x-custom-expose-headers"
    And CustomCORS: Preflight calling the "/ip" endpoint with header Origin:"localhost" should result in status code 200 and response header "Access-Control-Max-Age" with value "300"
    Then CustomCORS: Teardown httpbin service
