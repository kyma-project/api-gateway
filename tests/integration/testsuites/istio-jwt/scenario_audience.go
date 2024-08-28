package istiojwt

import (
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

func initAudience(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("istio-jwt-audiences.yaml", "istio-jwt-audiences")

	ctx.Step(`^Audiences: There is a httpbin service$`, scenario.thereIsAHttpbinService)
	ctx.Step(`^Audiences: There is an endpoint secured with JWT on path "([^"]*)" requiring audiences '(\[.*\])'$`, scenario.thereIsAnEndpointWithAudiences)
	ctx.Step(`^Audiences: There is an endpoint secured with JWT on path "([^"]*)" requiring audience '(\[.*\])' or '(\[.*\])'$`, scenario.emptyStep)
	ctx.Step(`^Audiences: The APIRule is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(`^Audiences: Calling the "([^"]*)" endpoint with a valid "([^"]*)" token with "([^"]*)" "([^"]*)" and "([^"]*)" should result in status between (\d+) and (\d+)`, scenario.callingTheEndpointWithAValidToken)
	ctx.Step(`^Audiences: Teardown httpbin service$`, scenario.teardownHttpbinService)
}

func (s *scenario) thereIsAnEndpointWithAudiences(path string, audiences string) error {
	s.ManifestTemplate[fmt.Sprintf("%s%s", strings.TrimPrefix(path, "/"), "Audiences")] = audiences
	return nil
}
