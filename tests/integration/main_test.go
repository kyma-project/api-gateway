package api_gateway

import (
	"context"
	"github.com/kyma-project/api-gateway/tests/integration/testsuites/v2alpha1"
	"log"
	"os"
	"testing"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/client"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/testcontext"
	customdomain "github.com/kyma-project/api-gateway/tests/integration/testsuites/custom-domain"
	"github.com/kyma-project/api-gateway/tests/integration/testsuites/gateway"
	istiojwt "github.com/kyma-project/api-gateway/tests/integration/testsuites/istio-jwt"
	"github.com/kyma-project/api-gateway/tests/integration/testsuites/ory"
	"github.com/kyma-project/api-gateway/tests/integration/testsuites/upgrade"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

func TestIstioJwt(t *testing.T) {
	config := testcontext.GetConfig()
	ts, err := testcontext.New(config, istiojwt.NewTestsuite)
	if err != nil {
		t.Fatalf("Failed to create Istio JWT testsuite %s", err.Error())
	}
	originalJwtHandler, err := SwitchJwtHandler(ts, "istio")
	if err != nil {
		log.Print(err.Error())
		t.Fatalf("unable to switch to Istio jwtHandler")
	}
	defer cleanUp(ts, originalJwtHandler)
	runTestsuite(t, ts, config)
}

func TestCustomDomain(t *testing.T) {
	config := testcontext.GetConfig()
	ts, err := testcontext.New(config, customdomain.NewTestsuite)
	if err != nil {
		t.Fatalf("Failed to create Custom domain testsuite %s", err.Error())
	}
	defer ts.TearDown()
	runTestsuite(t, ts, config)
}

func TestUpgrade(t *testing.T) {
	config := testcontext.GetConfig()
	config.TestConcurrency = 1
	ts, err := testcontext.New(config, upgrade.NewTestsuite)

	if err != nil {
		t.Fatalf("Failed to create Upgrade testsuite %s", err.Error())
	}
	originalJwtHandler, err := SwitchJwtHandler(ts, "istio")
	if err != nil {
		log.Print(err.Error())
		t.Fatalf("unable to switch to Istio jwtHandler")
	}
	defer cleanUp(ts, originalJwtHandler)
	defer ts.TearDown()
	runTestsuite(t, ts, config)
}

func TestOryJwt(t *testing.T) {
	config := testcontext.GetConfig()
	// We use a higher concurrency to counter the slower execution of the v2alpha1 JWT migration test
	config.TestConcurrency = 8
	ts, err := testcontext.New(config, ory.NewTestsuite)
	if err != nil {
		t.Fatalf("Failed to create Ory testsuite %s", err.Error())
	}
	originalJwtHandler, err := SwitchJwtHandler(ts, "ory")
	if err != nil {
		log.Print(err.Error())
		t.Fatalf("unable to switch to Ory jwtHandler")
	}
	defer cleanUp(ts, originalJwtHandler)
	runTestsuite(t, ts, config)
}

func TestGateway(t *testing.T) {
	config := testcontext.GetConfig()
	ts, err := testcontext.New(config, gateway.NewTestsuite)
	if err != nil {
		t.Fatalf("Failed to create Gateway testsuite %s", err.Error())
	}
	defer ts.TearDown()
	runTestsuite(t, ts, config)
}

func TestV2alpha1(t *testing.T) {
	config := testcontext.GetConfig()
	ts, err := testcontext.New(config, v2alpha1.NewTestsuite)
	if err != nil {
		t.Fatalf("Failed to create v2alpha1 testsuite %s", err.Error())
	}
	originalJwtHandler, err := SwitchJwtHandler(ts, "ory")
	if err != nil {
		log.Print(err.Error())
		t.Fatalf("unable to switch to Ory jwtHandler")
	}
	defer cleanUp(ts, originalJwtHandler)
	runTestsuite(t, ts, config)
}

func runTestsuite(t *testing.T, testsuite testcontext.Testsuite, config testcontext.Config) {
	opts := createGoDogOpts(t, testsuite.FeaturePath(), config.TestConcurrency)
	suite := godog.TestSuite{
		Name: testsuite.Name(),
		// We are not using ScenarioInitializer, as this function only needs to set up global resources
		TestSuiteInitializer: func(ctx *godog.TestSuiteContext) {
			ctx.BeforeSuite(func() {
				for _, hook := range testsuite.BeforeSuiteHooks() {
					err := hook()
					if err != nil {
						t.Fatalf("Cannot run before suite hooks: %s", err.Error())
					}
				}
			})

			testsuite.InitScenarios(ctx.ScenarioContext())

			ctx.AfterSuite(func() {
				for _, hook := range testsuite.AfterSuiteHooks() {
					err := hook()
					if err != nil {
						t.Fatalf("Cannot run after suite hooks: %s", err.Error())
					}
				}
			})
		},
		Options: &opts,
	}

	testExitCode := suite.Run()

	if shouldExportResults() {
		generateReport(testsuite)
	}

	if testExitCode != 0 {
		t.Fatalf("non-zero status returned, failed to run feature tests")
	}
}

func createGoDogOpts(t *testing.T, featuresPath []string, concurrency int) godog.Options {
	goDogOpts := godog.Options{
		Output:         colors.Colored(os.Stdout),
		Format:         "pretty",
		Paths:          featuresPath,
		Concurrency:    concurrency,
		TestingT:       t,
		Strict:         true,
		DefaultContext: createDefaultContext(t),
	}

	if shouldExportResults() {
		goDogOpts.Format = "pretty,junit:junit-report.xml,cucumber:cucumber-report.json"
	}

	return goDogOpts
}

func cleanUp(c testcontext.Testsuite, orgJwtHandler string) {

	c.TearDown()

	_, err := SwitchJwtHandler(c, orgJwtHandler)
	if err != nil {
		log.Print(err.Error())
		panic("unable to switch back to original jwtHandler")
	}
}

func shouldExportResults() bool {
	return os.Getenv("EXPORT_RESULT") == "true"
}

func createDefaultContext(t *testing.T) context.Context {
	ctx := testcontext.SetK8sClientInContext(context.Background(), client.GetK8sClient())
	return testcontext.SetTestingInContext(ctx, t)
}
