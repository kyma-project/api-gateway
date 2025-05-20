package v2

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/cucumber/godog"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/helpers"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/manifestprocessor"
)

func initScenario(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario()

	ctx.Step(`^APIRule has status "([^"]*)" with description containing "([^"]*)"$`, scenario.theAPIRuleHasStatusWithDesc)
	ctx.Step(
		`^Calling short host "([^"]*)" with path "([^"]*)" without a token should result in status between (\d+) and (\d+)$`,
		scenario.callingShortHostWithoutTokenShouldResultInStatusBetween,
	)
	ctx.Step(`^Calling the "([^"]*)" endpoint should return response with header "([^"]*)" with value "([^"]*)"$`, scenario.callingTheEndpointShouldResultInBodyContaining)
	ctx.Step(
		`^Calling the "([^"]*)" endpoint with "([^"]*)" method should result in status between (\d+) and (\d+)$`,
		scenario.callingTheEndpointWithMethodShouldResultInStatusBetween,
	)
	ctx.Step(
		`^Calling the "([^"]*)" endpoint with "([^"]*)" method with a valid "([^"]*)" token should result in status between (\d+) and (\d+)$`,
		scenario.callingTheEndpointWithMethodWithValidTokenShouldResultInStatusBetween,
	)
	ctx.Step(
		`^Calling the "([^"]*)" endpoint with "([^"]*)" method with any token should result in status between (\d+) and (\d+)$`,
		scenario.callingTheEndpointWithMethodWithInvalidTokenShouldResultInStatusBetween,
	)
	ctx.Step(
		`^Calling the "([^"]*)" endpoint with a valid "([^"]*)" token from default header should result in status between (\d+) and (\d+)$`,
		scenario.callingTheEndpointWithValidTokenShouldResultInStatusBetween,
	)
	ctx.Step(
		`^Calling the "([^"]*)" endpoint with a valid "([^"]*)" token from header "([^"]*)" and prefix "([^"]*)" should result in status between (\d+) and (\d+)$`,
		scenario.callingTheEndpointWithValidTokenFromHeaderShouldResultInStatusBetween,
	)
	ctx.Step(
		`^Calling the "([^"]*)" endpoint with a valid "([^"]*)" token from parameter "([^"]*)" should result in status between (\d+) and (\d+)$`,
		scenario.callingTheEndpointWithValidTokenFromParameterShouldResultInStatusBetween,
	)
	ctx.Step(
		`^Calling the "([^"]*)" endpoint with a valid "([^"]*)" token should result in body containing "([^"]*)"$`,
		scenario.callingTheEndpointWithValidTokenShouldResultInBodyContaining,
	)
	ctx.Step(
		`^Calling the "([^"]*)" endpoint with a valid "([^"]*)" token should result in status between (\d+) and (\d+)$`,
		scenario.callingTheEndpointWithValidTokenShouldResultInStatusBetween,
	)
	ctx.Step(
		`^Calling the "([^"]*)" endpoint with a valid "([^"]*)" token with "([^"]*)" "([^"]*)" and "([^"]*)" should result in status between (\d+) and (\d+)`,
		scenario.callingTheEndpointWithAValidToken,
	)
	ctx.Step(
		`^Calling the "([^"]*)" endpoint with an invalid token should result in status between (\d+) and (\d+)$`,
		scenario.callingTheEndpointWithInvalidTokenShouldResultInStatusBetween,
	)
	ctx.Step(
		`^Calling the "([^"]*)" endpoint with header "([^"]*)" with value "([^"]*)" and a valid "([^"]*)" token should result in status between (\d+) and (\d+)$`,
		scenario.callingTheEndpointWithHeaderAndValidJwt,
	)
	ctx.Step(
		`^Calling the "([^"]*)" endpoint with header "([^"]*)" with value "([^"]*)" and an invalid "([^"]*)" token should result in status between (\d+) and (\d+)$`,
		scenario.callingTheEndpointWithHeaderAndInvalidJwt,
	)
	ctx.Step(
		`^Calling the "([^"]*)" endpoint with header "([^"]*)" with value "([^"]*)" and no token should result in status between (\d+) and (\d+)$`,
		scenario.callingTheEndpointWithHeader,
	)
	ctx.Step(`^Calling the "([^"]*)" endpoint with header "([^"]*)" with value "([^"]*)" should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithHeader)
	ctx.Step(`^Calling the "([^"]*)" endpoint without a token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithoutTokenShouldResultInStatusBetween)
	ctx.Step(`^In-cluster calling the "([^"]*)" endpoint without a token should fail$`, scenario.inClusterCallingTheEndpointWithoutTokenShouldFail)
	ctx.Step(`^In-cluster calling the "([^"]*)" endpoint with a valid "([^"]*)" token should fail$`, scenario.inClusterCallingTheEndpointWithTokenShouldFail)
	ctx.Step(
		`^Preflight calling the "([^"]*)" endpoint with header Origin:"([^"]*)" should result in status code (\d+) and no response header "([^"]*)"$`,
		scenario.preflightEndpointCallNoResponseHeader,
	)
	ctx.Step(
		`^Preflight calling the "([^"]*)" endpoint with header Origin:"([^"]*)" should result in status code (\d+) and response header "([^"]*)" with value "([^"]*)"$`,
		scenario.preflightEndpointCallResponseHeaders,
	)
	ctx.Step(`^Specifies custom Gateway "([^"]*)"/"([^"]*)"`, scenario.specifiesCustomGateway)
	ctx.Step(`^Teardown helloworld service$`, scenario.teardownHelloworldCustomLabelService)
	ctx.Step(`^Teardown httpbin service$`, scenario.teardownHttpbinService)
	ctx.Step(`^Template value "([^"]*)" is set to "([^"]*)"$`, scenario.templateValueIsSetTo)
	ctx.Step(`^The APIRule is applied and contains error status with "([^"]*)" message$`, scenario.theAPIRulev2IsAppliedExpectError)
	ctx.Step(`^The APIRule is applied$`, scenario.theAPIRulev2IsApplied)
	ctx.Step(`^The APIRule template file is set to "([^"]*)"$`, scenario.theAPIRuleTemplateFileIsSetTo)
	ctx.Step(
		`^The APIRule with following CORS setup is applied AllowOrigins:'(\[.*\])', AllowMethods:'(\[.*\])', AllowHeaders:'(\[.*\])', AllowCredentials:"([^"]*)", ExposeHeaders:'(\[.*\])', MaxAge:"([^"]*)"$`,
		scenario.applyApiRuleWithCustomCORS,
	)
	ctx.Step(`^The APIRule with service on root level is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(`^The APIRule without CORS set up is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(`^The endpoint has JWT restrictions$`, scenario.theEndpointHasJwtRestrictionsWithScope)
	ctx.Step(`^The misconfigured APIRule is applied$`, scenario.theMisconfiguredAPIRuleIsApplied)
	ctx.Step(`^There is a helloworld service with custom label selector name "([^"]*)"$`, scenario.thereIsHelloworldCustomLabelService)
	ctx.Step(`^There is a httpbin service$`, scenario.thereIsAHttpbinService)
	ctx.Step(`^There is a service with workload in a second namespace`, scenario.thereIsServiceInSecondNamespace)
	ctx.Step(`^There is a workload and service for httpbin and helloworld$`, scenario.thereAreTwoServices)
	ctx.Step(`^There is an endpoint on path "([^"]*)" with a cookie mutator setting "([^"]*)" cookie to "([^"]*)"$`, scenario.thereIsAnEndpointWithCookie)
	ctx.Step(`^There is an endpoint on path "([^"]*)" with a header mutator setting "([^"]*)" header to "([^"]*)"$`, scenario.thereIsAnEndpointWithHeader)
	ctx.Step(`^There is an endpoint secured with ExtAuth "([^"]*)" on path "([^"]*)"$`, scenario.thereIsAnEndpointWithExtAuth)
	ctx.Step(`^There is an endpoint secured with JWT on path "([^"]*)" and /headers endpoint exposed with noAuth$`, scenario.thereIsAnJwtSecuredPath)
	ctx.Step(`^There is an endpoint secured with JWT on path "([^"]*)" for helloworld service with methods '(\[.*\])'$`, scenario.thereIsAJwtSecuredPathWithMethods)
	ctx.Step(`^There is an endpoint secured with JWT on path "([^"]*)" for httpbin service with methods '(\[.*\])'$`, scenario.thereIsAJwtSecuredPathWithMethods)
	ctx.Step(`^There is an endpoint secured with JWT on path "([^"]*)" in APIRule Namespace$`, scenario.thereIsAnJwtSecuredPath)
	ctx.Step(`^There is an endpoint secured with JWT on path "([^"]*)" in different namespace$`, scenario.thereIsAnJwtSecuredPathInDifferentNamespace)
	ctx.Step(`^There is an endpoint secured with JWT on path "([^"]*)" requiring audience '(\[.*\])' or '(\[.*\])'$`, scenario.emptyStep)
	ctx.Step(`^There is an endpoint secured with JWT on path "([^"]*)" requiring audiences '(\[.*\])'$`, scenario.thereIsAnEndpointWithAudiences)
	ctx.Step(`^There is an endpoint secured with JWT on path "([^"]*)" requiring scopes '(\[.*\])'$`, scenario.thereIsAnEndpointWithRequiredScopes)
	ctx.Step(`^There is an endpoint secured with JWT on path "([^"]*)" with invalid issuer and jwks$`, scenario.emptyStep)
	ctx.Step(`^There is an endpoint secured with JWT on path "([^"]*)" with service definition$`, scenario.thereIsAnEndpointWithServiceDefinition)
	ctx.Step(`^There is an endpoint secured with JWT on path "([^"]*)"$`, scenario.thereIsAnJwtSecuredPath)
	ctx.Step(`^There is an httpbin service$`, scenario.thereIsAHttpbinService)
}

func (s *scenario) applyApiRuleWithCustomCORS(allowOrigins, allowMethods, allowHeaders, allowCredentials, exposeHeaders, maxAge string) error {
	s.ManifestTemplate["AllowOrigins"] = allowOrigins
	s.ManifestTemplate["AllowMethods"] = allowMethods
	s.ManifestTemplate["AllowHeaders"] = allowHeaders
	s.ManifestTemplate["AllowCredentials"] = allowCredentials
	s.ManifestTemplate["ExposeHeaders"] = exposeHeaders
	s.ManifestTemplate["MaxAge"] = maxAge
	r, err := manifestprocessor.ParseFromFileWithTemplate(s.ApiResourceManifestPath, s.ApiResourceDirectory, s.ManifestTemplate)
	if err != nil {
		return err
	}

	_, err = s.resourceManager.CreateOrUpdateResources(s.k8sClient, r...)
	if err != nil {
		return err
	}

	return nil
}

func (s *scenario) thereIsAnEndpointWithAudiences(path string, audiences string) error {
	s.ManifestTemplate[fmt.Sprintf("%s%s", strings.TrimPrefix(path, "/"), "Audiences")] = audiences
	return nil
}

func (s *scenario) thereIsAnEndpointWithRequiredScopes(path string, scopes string) error {
	s.ManifestTemplate[fmt.Sprintf("%s%s", strings.TrimPrefix(path, "/"), "RequiredScopes")] = scopes
	return nil
}

func (s *scenario) callingTheEndpointWithValidTokenFromHeaderShouldResultInStatusBetween(endpoint, tokenType string, fromHeader string, prefix string, lower, higher int) error {
	asserter := &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher}
	tokenFrom := tokenFrom{
		From:     fromHeader,
		Prefix:   prefix,
		AsHeader: true,
	}
	return s.callingEndpointWithMethodAndHeaders(fmt.Sprintf("%s/%s", s.Url, strings.TrimLeft(endpoint, "/")), http.MethodGet, tokenType, asserter, nil, &tokenFrom)
}

func (s *scenario) callingTheEndpointWithValidTokenFromParameterShouldResultInStatusBetween(endpoint, tokenType string, fromParameter string, lower, higher int) error {
	asserter := &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher}
	tokenFrom := tokenFrom{
		From:     fromParameter,
		AsHeader: false,
	}
	return s.callingEndpointWithMethodAndHeaders(fmt.Sprintf("%s/%s", s.Url, strings.TrimLeft(endpoint, "/")), http.MethodGet, tokenType, asserter, nil, &tokenFrom)
}

func (s *scenario) thereIsAnEndpointWithHeader(_, header, headerValue string) error {
	s.ManifestTemplate["header"] = header
	s.ManifestTemplate["headerValue"] = headerValue
	return nil
}

func (s *scenario) thereIsAnEndpointWithCookie(_, cookie, cookieValue string) error {
	s.ManifestTemplate["cookie"] = cookie
	s.ManifestTemplate["cookieValue"] = cookieValue
	return nil
}

func (s *scenario) thereIsHelloworldCustomLabelService(labelName string) error {
	s.ManifestTemplate["CustomLabelName"] = labelName
	resources, err := manifestprocessor.ParseFromFileWithTemplate("testing-helloworld-custom-label-app.yaml", s.ApiResourceDirectory, s.ManifestTemplate)
	if err != nil {
		return err
	}
	_, err = s.resourceManager.CreateResources(s.k8sClient, resources...)

	s.Url = fmt.Sprintf("https://helloworld-%s.%s", s.TestID, s.Domain)

	return err
}

func (s *scenario) teardownHelloworldCustomLabelService() error {
	resources, err := manifestprocessor.ParseFromFileWithTemplate("testing-helloworld-custom-label-app.yaml", s.ApiResourceDirectory, s.ManifestTemplate)
	if err != nil {
		return err
	}
	err = s.resourceManager.DeleteResources(s.k8sClient, resources...)
	if err != nil {
		return err
	}

	s.Url = ""

	return nil
}

func (s *scenario) thereAreTwoServices() error {
	resources, err := manifestprocessor.ParseFromFileWithTemplate("testing-helloworld-app.yaml", s.ApiResourceDirectory, s.ManifestTemplate)
	if err != nil {
		return err
	}
	_, err = s.resourceManager.CreateResources(s.k8sClient, resources...)
	return err
}

func (s *scenario) thereIsAJwtSecuredPathWithMethods(path string, methods string) {
	pathName := strings.TrimPrefix(path, "/")
	s.ManifestTemplate[fmt.Sprintf("%s%s", pathName, "Methods")] = methods
	s.ManifestTemplate[fmt.Sprintf("%sJwtSecuredPath", pathName)] = path
}

func (s *scenario) thereIsAnEndpointWithServiceDefinition(path string) {
	s.ManifestTemplate["jwtSecuredPathWithService"] = path
}

func (s *scenario) thereIsAnJwtSecuredPathInDifferentNamespace(path string) {
	s.ManifestTemplate["otherNamespacePath"] = path
}

func (s *scenario) thereIsServiceInSecondNamespace() error {
	resources, err := manifestprocessor.ParseFromFileWithTemplate("second-service.yaml", s.ApiResourceDirectory, s.ManifestTemplate)
	if err != nil {
		return err
	}
	_, err = s.resourceManager.CreateResources(s.k8sClient, resources...)
	return err
}
