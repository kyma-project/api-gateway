package api_gateway

import "github.com/cucumber/godog"

func initCustomDomainTestsuite(ctx *godog.ScenarioContext) {
	InitializeScenarioCustomDomain(ctx)
}
