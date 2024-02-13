package ory

import (
	_ "embed"
	"github.com/cucumber/godog"
)

type servicePerPathScenario struct {
	*scenario
}

func initServicePerPath(ctx *godog.ScenarioContext, ts *testsuite) {
	s := ts.createScenario("service-per-path.yaml", "service-per-path")

	scenario := servicePerPathScenario{s}

	ctx.Step(`^Service per path: There is a httpbin service`, scenario.thereIsAHttpbinService)
	ctx.Step(`^Service per path: There is a helloworld service and an APIRule with two endpoints exposed with different services, one on spec level and one on rule level$`, scenario.theAPIRuleIsApplied)
	ctx.Step(`^Service per path: Calling the endpoint "([^"]*)" and "([^"]*)" with any token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointsWithInvalidTokenShouldResultInStatusBetween)
	ctx.Step(`^Service per path: Calling the endpoint "([^"]*)" and "([^"]*)" without a token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointsWithoutTokenShouldResultInStatusBetween)
	ctx.Step(`^Service per path: Teardown httpbin service$`, scenario.teardownHttpbinService)
}

func (s *servicePerPathScenario) callingTheEndpointsWithInvalidTokenShouldResultInStatusBetween(path1, path2 string, lower, higher int) error {
	err := s.callingTheEndpointWithInvalidTokenShouldResultInStatusBetween(path1, lower, higher)
	if err != nil {
		return err
	}
	return s.callingTheEndpointWithInvalidTokenShouldResultInStatusBetween(path2, lower, higher)
}

func (s *servicePerPathScenario) callingTheEndpointsWithoutTokenShouldResultInStatusBetween(path1, path2 string, lower, higher int) error {
	err := s.callingTheEndpointWithoutTokenShouldResultInStatusBetween(path1, lower, higher)
	if err != nil {
		return err
	}
	return s.callingTheEndpointWithoutTokenShouldResultInStatusBetween(path2, lower, higher)
}
