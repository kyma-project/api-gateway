package v2alpha1

import (
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

func initJwtAudience(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("jwt-audiences.yaml", "jwt-audiences")

	ctx.Step(`^JwtAudiences: There is a httpbin service$`, scenario.thereIsAHttpbinService)
	ctx.Step(`^JwtAudiences: There is an endpoint secured with JWT on path "([^"]*)" requiring audiences '(\[.*\])'$`, scenario.thereIsAnEndpointWithAudiences)
	ctx.Step(`^JwtAudiences: There is an endpoint secured with JWT on path "([^"]*)" requiring audience '(\[.*\])' or '(\[.*\])'$`, scenario.emptyStep)
	ctx.Step(`^JwtAudiences: The APIRule is applied$`, scenario.theAPIRuleV2Alpha1IsApplied)
	ctx.Step(`^JwtAudiences: Calling the "([^"]*)" endpoint with a valid "([^"]*)" token with "([^"]*)" "([^"]*)" and "([^"]*)" should result in status between (\d+) and (\d+)`, scenario.callingTheEndpointWithAValidToken)
	ctx.Step(`^JwtAudiences: Teardown httpbin service$`, scenario.teardownHttpbinService)
}

func (s *scenario) thereIsAnEndpointWithAudiences(path string, audiences string) error {
	s.ManifestTemplate[fmt.Sprintf("%s%s", strings.TrimPrefix(path, "/"), "Audiences")] = audiences
	return nil
}
