package api_gateway

import (
	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/testcontext"
	"github.com/kyma-project/api-gateway/tests/integration/testsuites/custom-domain"
	"github.com/kyma-project/api-gateway/tests/integration/testsuites/istio-jwt"
	"log"
	"os"
	"testing"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

func TestIstioJwt(t *testing.T) {
	config := testcontext.GetConfig()
	testsuite := testcontext.New(config, istiojwt.NewTestsuite)
	orgJwtHandler, err := SwitchJwtHandler(testsuite, "istio")
	if err != nil {
		log.Print(err.Error())
		t.Fatalf("unable to switch to Istio jwtHandler")
	}
	defer cleanUp(testsuite, orgJwtHandler)

	opts := createGoDogOpts(t, "testsuites/istio-jwt/features/", config.TestConcurrency)
	suite := godog.TestSuite{
		Name: testsuite.Name(),
		// We are not using ScenarioInitializer, as this function only needs to set up global resources
		TestSuiteInitializer: func(ctx *godog.TestSuiteContext) {
			err := istiojwt.Init(ctx.ScenarioContext(), testsuite)
			if err != nil {
				t.Fatalf("Failed to initialize Istio JWT testsuite %s", err.Error())
			}
		},
		Options: &opts,
	}

	testExitCode := suite.Run()
	if testExitCode != 0 {
		t.Fatalf("non-zero status returned, failed to run feature tests")
	}
}

func TestCustomDomain(t *testing.T) {
	config := testcontext.GetConfig()
	testsuite := testcontext.New(config, customdomain.NewTestsuite)
	defer testsuite.TearDown()
	opts := createGoDogOpts(t, "testsuites/custom-domain/features/", config.TestConcurrency)

	customDomainSuite := godog.TestSuite{
		Name: testsuite.Name(),
		TestSuiteInitializer: func(ctx *godog.TestSuiteContext) {
			err := customdomain.Init(ctx.ScenarioContext(), testsuite)
			if err != nil {
				t.Fatalf("Failed to initialize Custom domain testsuite %s", err.Error())
			}
		},
		Options: &opts,
	}

	testExitCode := customDomainSuite.Run()

	if shouldExportResults() {
		generateReport()
	}

	if testExitCode != 0 {
		t.Fatalf("non-zero status returned, failed to run feature tests")
	}
}

func createGoDogOpts(t *testing.T, featuresPath string, concurrency int) godog.Options {
	goDogOpts := godog.Options{
		Output:      colors.Colored(os.Stdout),
		Format:      "pretty",
		Paths:       []string{featuresPath},
		Concurrency: concurrency,
		TestingT:    t,
	}

	if shouldExportResults() {
		goDogOpts.Format = "pretty,junit:junit-report.xml,cucumber:cucumber-report.json"
	}

	return goDogOpts
}

func cleanUp(c testcontext.Testsuite, orgJwtHandler string) {

	c.TearDown()

	if shouldExportResults() {
		generateReport()
	}

	_, err := SwitchJwtHandler(c, orgJwtHandler)
	if err != nil {
		log.Print(err.Error())
		panic("unable to switch back to original jwtHandler")
	}
}

func shouldExportResults() bool {
	return os.Getenv("EXPORT_RESULT") == "true"
}
