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
	ctx.Step(`^Service per path: There is a helloworld service`, scenario.thereIsAHelloWorldService)
	ctx.Step(`^Service per path: The APIRule .* is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(
		`^Service per path: Calling the endpoint "([^"]*)" and "([^"]*)" with any token should result in status between (\d+) and (\d+)$`,
		scenario.callingTheEndpointsWithInvalidTokenShouldResultInStatusBetween,
	)
	ctx.Step(
		`^Service per path: Calling the endpoint "([^"]*)" and "([^"]*)" without a token should result in status between (\d+) and (\d+)$`,
		scenario.callingTheEndpointsWithoutTokenShouldResultInStatusBetween,
	)
	ctx.Step(`^Service per path: Teardown httpbin service$`, scenario.teardownHttpbinService)
	ctx.Step(`^Service per path: Teardown helloworld service$`, scenario.teardownHelloWorldService)
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
