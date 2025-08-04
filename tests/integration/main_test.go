package api_gateway

import (
	"context"
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
	ratelimit "github.com/kyma-project/api-gateway/tests/integration/testsuites/rate-limit"
	"github.com/kyma-project/api-gateway/tests/integration/testsuites/upgrade"
	v2 "github.com/kyma-project/api-gateway/tests/integration/testsuites/v2"
	"github.com/kyma-project/api-gateway/tests/integration/testsuites/v2alpha1"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

func TestIstioJwt(t *testing.T) {
	ts, err := testcontext.New(istiojwt.NewTestsuite)
	if err != nil {
		t.Fatalf("Failed to create Istio JWT testsuite %s", err.Error())
	}
	originalJwtHandler, err := SwitchJwtHandler(ts, "istio")
	if err != nil {
		log.Print(err.Error())
		t.Fatalf("unable to switch to Istio jwtHandler")
	}
	defer cleanUp(t, ts, originalJwtHandler)
	defer tearDown(t, ts)
	runTestsuite(t, ts)
}

func TestCustomDomain(t *testing.T) {
	ts, err := testcontext.New(customdomain.NewTestsuite)
	if err != nil {
		t.Fatalf("Failed to create Custom domain testsuite %s", err.Error())
	}
	defer tearDown(t, ts)
	runTestsuite(t, ts)
}

func TestUpgrade(t *testing.T) {
	ts, err := testcontext.New(upgrade.NewTestsuite)

	if err != nil {
		t.Fatalf("Failed to create Upgrade testsuite %s", err.Error())
	}
	originalJwtHandler, err := SwitchJwtHandler(ts, "istio")
	if err != nil {
		log.Print(err.Error())
		t.Fatalf("unable to switch to Istio jwtHandler")
	}
	defer cleanUp(t, ts, originalJwtHandler)
	defer tearDown(t, ts)
	runTestsuite(t, ts)
}

func TestOryJwt(t *testing.T) {
	ts, err := testcontext.New(ory.NewTestsuite)
	if err != nil {
		t.Fatalf("Failed to create Ory testsuite %s", err.Error())
	}
	originalJwtHandler, err := SwitchJwtHandler(ts, "ory")
	if err != nil {
		log.Print(err.Error())
		t.Fatalf("unable to switch to Ory jwtHandler")
	}
	defer cleanUp(t, ts, originalJwtHandler)
	runTestsuite(t, ts)
}

func TestOryZeroDowntimeMigration(t *testing.T) {
	ts, err := testcontext.New(ory.NewZDTestsuite)
	if err != nil {
		t.Fatalf("Failed to create Ory Zero Downtime Migration testsuite %s", err.Error())
	}
	originalJwtHandler, err := SwitchJwtHandler(ts, "ory")
	if err != nil {
		log.Print(err.Error())
		t.Fatalf("unable to switch to Ory jwtHandler")
	}
	defer cleanUp(t, ts, originalJwtHandler)
	runTestsuite(t, ts)
}

func TestGateway(t *testing.T) {
	ts, err := testcontext.New(gateway.NewTestsuite)
	if err != nil {
		t.Fatalf("Failed to create Gateway testsuite %s", err.Error())
	}
	defer tearDown(t, ts)
	runTestsuite(t, ts)
}

func TestV2alpha1(t *testing.T) {
	ts, err := testcontext.New(v2alpha1.NewTestsuite)
	if err != nil {
		t.Fatalf("Failed to create v2alpha1 testsuite %s", err.Error())
	}
	originalJwtHandler, err := SwitchJwtHandler(ts, "ory")
	if err != nil {
		log.Print(err.Error())
		t.Fatalf("unable to switch to Ory jwtHandler")
	}
	defer cleanUp(t, ts, originalJwtHandler)
	runTestsuite(t, ts)
}

func TestRateLimit(t *testing.T) {
	ts, err := testcontext.New(ratelimit.NewTestsuite)
	if err != nil {
		t.Fatalf("Failed to create ratelimit testsuite %s", err.Error())
	}
	defer tearDown(t, ts)
	runTestsuite(t, ts)
}

func TestV2(t *testing.T) {
	ts, err := testcontext.New(v2.NewTestsuite)
	if err != nil {
		t.Fatalf("Failed to create v2 testsuite %s", err.Error())
	}
	originalJwtHandler, err := SwitchJwtHandler(ts, "ory")
	if err != nil {
		log.Print(err.Error())
		t.Fatalf("unable to switch to Ory jwtHandler")
	}
	defer cleanUp(t, ts, originalJwtHandler)
	runTestsuite(t, ts)
}

func runTestsuite(t *testing.T, testsuite testcontext.Testsuite) {
	opts := createGoDogOpts(t, testsuite.FeaturePath(), testsuite.TestConcurrency())
	suite := godog.TestSuite{
		Name:                testsuite.Name(),
		ScenarioInitializer: testsuite.InitScenarios,
		TestSuiteInitializer: func(ctx *godog.TestSuiteContext) {
			ctx.BeforeSuite(func() {
				log.Printf("Executing before suite hooks")
				for _, hook := range testsuite.BeforeSuiteHooks() {
					err := hook()
					if err != nil {
						t.Fatalf("Cannot run before suite hooks: %s", err.Error())
					}
				}
				log.Printf("Before suite hooks finished")
			})

			ctx.AfterSuite(func() {
				if t.Failed() {
					log.Printf("Test suite failed, skipping after suite on success hooks")
				} else {
					log.Printf("Executing after suite on success hooks")
					for _, hook := range testsuite.AfterSuiteHooks() {
						err := hook()
						if err != nil {
							t.Fatalf("Cannot run after suite hooks: %s", err.Error())
						}
					}
					log.Printf("After suite hooks executed")

					log.Printf("Tearing down test suite")
					testsuite.TearDown()
				}
			})
		},
		Options: &opts,
	}

	log.Printf("Starting godog test suite: %s", suite.Name)
	testExitCode := suite.Run()
	log.Printf("Godog test suite: %s has been finished with exit code: %d", suite.Name, testExitCode)

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
		goDogOpts.Format = "pretty,cucumber:cucumber-report.json"
	}

	return goDogOpts
}

func cleanUp(t *testing.T, c testcontext.Testsuite, orgJwtHandler string) {
	if t.Failed() {
		log.Printf("Tests failed, skipping jwt handler cleanup")
		return
	}
	_, err := SwitchJwtHandler(c, orgJwtHandler)
	if err != nil {
		log.Print(err.Error())
		panic("unable to switch back to original jwtHandler")
	}
}

func tearDown(t *testing.T, ts testcontext.Testsuite) {
	if t.Failed() {
		log.Printf("Tests failed, skipping teardown")
		return
	}
	ts.TearDown()
}

func shouldExportResults() bool {
	return os.Getenv("EXPORT_RESULT") == "true"
}

func createDefaultContext(t *testing.T) context.Context {
	ctx := testcontext.SetK8sClientInContext(context.Background(), client.GetK8sClient())
	return testcontext.SetTestingInContext(ctx, t)
}
